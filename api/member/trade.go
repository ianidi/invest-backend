package member

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/ianidi/exchange-server/internal/trade"
	"github.com/shopspring/decimal"
)

// TradeNew
// @Summary
// @Description Trade
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Trade-New
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /trade/new [post]
func TradeNew(c *gin.Context) {
	var order trade.Order
	var err error

	order.Timestamp = time.Now().Unix()

	order.Member, err = QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	order.MemberID = order.Member.MemberID

	var query struct {
		AssetID    int64  `json:"AssetID" binding:"required"`
		Rate       string `json:"Rate"`
		Qty        string `json:"Qty" binding:"required"`
		Action     string `json:"Action" binding:"required"`
		Type       string `json:"Type" binding:"required"`
		StopLoss   int64  `json:"StopLoss"`
		TakeProfit int64  `json:"TakeProfit"`
		Leverage   int64  `json:"Leverage"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if query.Action != trade.ACTION_BUY && query.Action != trade.ACTION_SELL {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_ACTION"})
		return
	}

	order.Action = query.Action

	if query.Type != trade.ORDER_LIMIT && query.Type != trade.ORDER_MARKET {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_ORDER_TYPE"})
		return
	}

	order.Type = query.Type

	order.MemberRate.Decimal, err = decimal.NewFromString(query.Rate)
	if err != nil && order.Type == trade.ORDER_LIMIT {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "TRADE_INVALID_PRICE"})
		return
	}

	if order.Type == trade.ORDER_LIMIT && (order.MemberRate.Decimal.IsZero() || order.MemberRate.Decimal.IsNegative()) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "TRADE_INVALID_PRICE"})
		return
	}

	order.Qty.Decimal, err = decimal.NewFromString(query.Qty)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_QTY"})
		return
	}

	if order.Qty.Decimal.IsZero() || order.Qty.Decimal.IsNegative() {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_QTY"})
		return
	}

	//Get asset record
	order.Asset, err = order.QueryAsset(query.AssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//Get member balance for asset specified
	order.BalanceAsset.Decimal, err = order.QueryAssetBalance()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//TODO: If asset is stock, its' qty must not have decimal point. Check decimal points for different kinds of assets
	// if order.Asset.MarketID != 1 {
	//
	// }

	//Determine stop loss / take profit values for order
	stopLossAllowed, takeProfitAllowed, err := order.DetermineMaxAllowedSLTP()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	order.StopLoss.Decimal = decimal.NewFromInt(query.StopLoss)
	order.TakeProfit.Decimal = decimal.NewFromInt(query.TakeProfit)

	if order.StopLoss.Decimal.IsZero() != false && (order.StopLoss.Decimal.IsNegative() || order.StopLoss.Decimal.GreaterThan(stopLossAllowed)) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_STOP_LOSS"})
		return
	}

	if order.TakeProfit.Decimal.IsZero() != false && (order.TakeProfit.Decimal.IsNegative() || order.StopLoss.Decimal.GreaterThan(takeProfitAllowed)) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_TAKE_PROFIT"})
		return
	}

	//Query current asset market rate
	order.MarketRate.Decimal = order.DetermineMarketRate()

	//Determine current asset rate
	order.RateEntry.Decimal = order.DetermineRateEntry()

	//If member didn't pass a leverage value and current asset MarketID is not Forex, set default 1x value, otherwise determine default max allowed system leverage for asset by MarketID
	if query.Leverage == 0 && order.Asset.MarketID != 3 {
		order.Leverage.Decimal = decimal.NewFromInt(1)
	} else {
		order.Leverage.Decimal, err = order.DetermineDefaultLeverage()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
			return
		}
	}

	//If member passed a leverage value and current asset is not Forex (members can't set leverage for Forex lots)
	if query.Leverage > 0 && order.Asset.MarketID != 3 {

		leverage := decimal.NewFromInt(query.Leverage)

		//Determine max leverage value allowed for current member for this asset
		leverageAllowed, err := order.DetermineLeverageAllowed()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
			return
		}

		if leverage.GreaterThan(leverageAllowed) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_LEVERAGE"})
			return
		}

		//Assign member defined leverage value
		order.Leverage.Decimal = leverage
	}

	//Determine order total real market value (cost) (without leverage)
	order.TotalReal.Decimal, err = order.DetermineTotalReal()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//Total order value (cost) for member balance (leverage applied)
	order.Total.Decimal = order.TotalReal.Decimal.Div(order.Leverage.Decimal)

	//If it is a Forex asset, additional calculations are required
	if order.Asset.MarketID == 3 {

		//Determine one pip
		order.OnePip.Decimal, err = order.DetermineOnePip()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
			return
		}

		//How much pips forex asset rate had at order creation time
		order.PipsRateEntry.Decimal = order.RateEntry.Decimal.Div(order.OnePip.Decimal)

		//How much forex asset member bought / sold (for display in dashboard)
		order.ForexAmount.Decimal = order.TotalReal.Decimal.Div(order.RateEntry.Decimal)
	}

	//Get member account balance before order was placed
	order.BalanceEntry.Decimal, err = order.QueryCurrentBalance()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//Сheck that user has enough funds (leverage applied) on balance to buy this qty of assets
	if order.Total.Decimal.GreaterThan(order.BalanceEntry.Decimal) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "TRADE_INSUFFICIENT_WALLET"})
		return
	}

	//Determine order status
	order.Status, err = order.DetermineStatus()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//Prevent limit orders from losing / gaining too much profit
	if order.Type == trade.ORDER_LIMIT {
		difference := order.RateEntry.Decimal.Div(order.MarketRate.Decimal).Mul(decimal.NewFromInt(100)).Sub(decimal.NewFromInt(100))

		settings, err := order.QuerySettings()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
			return
		}

		//Buy more expensive than market price
		if order.Action == trade.ACTION_BUY && settings.StopLossProtection.Decimal.IsZero() == false && difference.IsPositive() && difference.Abs().GreaterThan(settings.StopLossProtection.Decimal) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "MAX_STOP_LOSS"})
			return
		}

		//Buy less expensive than market price
		if order.Action == trade.ACTION_BUY && settings.TakeProfitProtection.Decimal.IsZero() == false && difference.IsNegative() && difference.Abs().GreaterThan(settings.TakeProfitProtection.Decimal) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "MAX_TAKE_PROFIT"})
			return
		}

		//Sell less expensive than market price
		if order.Action == trade.ACTION_SELL && settings.StopLossProtection.Decimal.IsZero() == false && difference.IsNegative() && difference.Abs().GreaterThan(settings.StopLossProtection.Decimal) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "MAX_STOP_LOSS"})
			return
		}

		//Sell more expensive than market price
		if order.Action == trade.ACTION_SELL && settings.TakeProfitProtection.Decimal.IsZero() == false && difference.IsPositive() && difference.Abs().GreaterThan(settings.TakeProfitProtection.Decimal) {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "MAX_TAKE_PROFIT"})
			return
		}

	}

	//Open creates new order and returns its' TradeID
	order.TradeID, err = order.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//CalculateProfit returns order struct with profit calculation and calls a function to record it to database
	order.CalculateProfit()

	c.JSON(200, gin.H{
		"status": true,
	})
}

// TradeClose
// @Summary
// @Description Trade
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Trade-Close
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /trade/close [post]
func TradeClose(c *gin.Context) {
	db := db.GetDB()

	var order trade.Order
	var err error

	order.Member, err = QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	order.MemberID = order.Member.MemberID

	var query struct {
		TradeID int `json:"TradeID" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	err = db.Get(&order, "SELECT * FROM Trade WHERE TradeID=$1 AND MemberID=$2 AND (Status=$3 OR Status=$4)", query.TradeID, order.Member.MemberID, trade.STATUS_OPEN, trade.STATUS_PENDING)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "INVALID_TRADE",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	//Get asset record
	order.Asset, err = order.QueryAsset(order.AssetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	//Close order
	order.Close()

	var tradeUpdated []*models.Trade

	err = db.Select(&tradeUpdated, "SELECT * FROM Trade WHERE MemberID=$1 ORDER BY TradeID DESC", order.Member.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"trade":  tradeUpdated,
	})
}

// TradeCloseBulk
// @Summary
// @Description Trade
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Trade-Close-Bulk
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /trade/close/bulk [post]
func TradeCloseBulk(c *gin.Context) {
	db := db.GetDB()

	var order trade.Order
	var err error

	order.Member, err = QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	order.MemberID = order.Member.MemberID

	var query struct {
		Type string `json:"Type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if query.Type != trade.TYPE_LOSE && query.Type != trade.TYPE_PROFIT && query.Type != trade.TYPE_ALL {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_ACTION"})
		return
	}

	var orders []trade.Order

	if query.Type == trade.TYPE_LOSE {
		err = db.Select(&orders, "SELECT * FROM Trade WHERE Profit<$1 AND MemberID=$2 AND Status=$3", 0, order.Member.MemberID, trade.STATUS_OPEN)
	}

	if query.Type == trade.TYPE_PROFIT {
		err = db.Select(&orders, "SELECT * FROM Trade WHERE Profit>$1 AND MemberID=$2 AND Status=$3", 0, order.Member.MemberID, trade.STATUS_OPEN)
	}

	if query.Type == trade.TYPE_ALL {
		err = db.Select(&orders, "SELECT * FROM Trade WHERE MemberID=$1 AND (Status=$2 OR Status=$3)", order.Member.MemberID, trade.STATUS_OPEN, trade.STATUS_PENDING)
	}

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	for _, orderRow := range orders {

		order := orderRow

		//Get asset record
		order.Asset, err = order.QueryAsset(order.AssetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
			return
		}

		//Close order
		order.Close()

	}

	var tradeUpdated []*models.Trade

	err = db.Select(&tradeUpdated, "SELECT * FROM Trade WHERE MemberID=$1 ORDER BY TradeID DESC", order.Member.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"trade":  tradeUpdated,
	})
}
