package trade

import (
	"database/sql"
	"errors"

	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
	"github.com/shopspring/decimal"
)

const (
	CURRENCY_USD     = "USD"
	ACTION_SELL      = "s"
	ACTION_BUY       = "b"
	ORDER_LIMIT      = "l"
	ORDER_MARKET     = "m"
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
	STATUS_OPEN      = "open"
	STATUS_CLOSED    = "closed"
	TYPE_LOSE        = "lose"
	TYPE_PROFIT      = "profit"
	TYPE_ALL         = "all"
)

type Order struct {
	Member models.Member //Member information
	Asset  models.Asset  //Asset information
	models.Trade
}

type OrderInterface interface {
	QueryAsset() (models.Asset, error)
	DetermineMarketRate() (decimal.Decimal, error)
	QueryAssetBalance() (decimal.Decimal, error)
	QueryCurrentBalance() (decimal.Decimal, error)
	QueryMember() (decimal.Decimal, error)
	DetermineRateEntry() (decimal.Decimal, error)
	DetermineStatus() (string, error)
	DetermineOnePip() (decimal.Decimal, error)
	DetermineDefaultLeverage() (decimal.Decimal, error)
	QuerySettings() (models.Settings, error)
	DetermineLeverageAllowed() (decimal.Decimal, error)
	DetermineTotalReal() (decimal.Decimal, error)
	DetermineForexProfit() (decimal.Decimal, error)
	DetermineOrderSLTP() (decimal.Decimal, decimal.Decimal, error)
	DetermineMaxAllowedSLTP() (decimal.Decimal, decimal.Decimal, error)
	CalculateProfit() error
	UpdateProfit() error
	Open() error
	Close() error
	CloseOpen() error
	CloseSLTP() error
	CancelPending() error
}

func (order Order) QueryAsset(AssetID int64) (models.Asset, error) {
	db := db.GetDB()

	var asset models.Asset

	err := db.Get(&asset, "SELECT * FROM Asset WHERE AssetID=$1", AssetID)

	if err != nil {
		if err == sql.ErrNoRows {
			return asset, errors.New("INVALID_ASSET")
		}
		return asset, err
	}

	//Not tradable if asset is disabled or rate is 0
	if asset.Tradable == false || asset.Active == false || asset.Rate.Decimal.IsZero() {
		return asset, errors.New("TRADE_ASSET_NOT_TRADABLE")
	}

	return asset, nil
}

//Query current asset rate
func (order Order) DetermineMarketRate() decimal.Decimal {

	var rate decimal.Decimal

	if order.Action == ACTION_BUY {
		rate = order.Asset.RateBuy.Decimal
	}

	if order.Action == ACTION_SELL {
		rate = order.Asset.RateSell.Decimal
	}

	return rate
}

//Query asset balance in member wallet
func (order Order) QueryAssetBalance() (decimal.Decimal, error) {
	db := db.GetDB()

	walletRecord := true

	var wallet models.Wallet

	err := db.Get(&wallet, "SELECT * FROM Wallet WHERE MemberID=$1 AND AssetID=$2", order.MemberID, order.Asset.AssetID)

	if err != nil {
		if err == sql.ErrNoRows {
			walletRecord = false
		} else {
			return wallet.Balance.Decimal, err
		}
	}

	//No wallet record
	if !walletRecord {
		tx := db.MustBegin()
		tx.MustExec("INSERT INTO Wallet (MemberID, AssetID) VALUES ($1, $2)", order.MemberID, order.Asset.AssetID)
		tx.Commit()
	}

	return wallet.Balance.Decimal, nil
}

//Determine asset rate depending on order type
func (order Order) DetermineRateEntry() decimal.Decimal {

	var rate decimal.Decimal

	if order.Type == ORDER_MARKET {

		rate = order.MarketRate.Decimal

	}

	if order.Type == ORDER_LIMIT {

		rate = order.MemberRate.Decimal

	}

	return rate
}

//Query current member balance
func (order Order) QueryCurrentBalance() (decimal.Decimal, error) {
	db := db.GetDB()

	var balance shopspring.Numeric

	err := db.Get(&balance, "SELECT "+order.Asset.Currency+" FROM Member WHERE MemberID=$1", order.MemberID)

	if err != nil {
		if err == sql.ErrNoRows {
			return balance.Decimal, errors.New("INVALID_MEMBER")
		}
		return balance.Decimal, err
	}

	return balance.Decimal, nil
}

//Query current member
func (order Order) QueryMember() (models.Member, error) {
	db := db.GetDB()

	var member models.Member

	err := db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1", order.MemberID)

	if err != nil {
		if err == sql.ErrNoRows {
			return member, errors.New("INVALID_MEMBER")
		}
		return member, err
	}

	return member, nil
}

//Determine order status
func (order Order) DetermineStatus() (string, error) {

	var status string

	status = STATUS_OPEN

	// limit (rate determined by member). Check if buy limit order doesn't meet requirement at creation time
	if order.Type == ORDER_LIMIT {

		if order.Action == ACTION_BUY && order.RateEntry.Decimal.LessThanOrEqual(order.MarketRate.Decimal) {
			status = STATUS_PENDING
		}

		if order.Action == ACTION_SELL && order.RateEntry.Decimal.GreaterThanOrEqual(order.MarketRate.Decimal) {
			status = STATUS_PENDING
		}
	}

	return status, nil
}

//Determine one forex pip (0.0001 for all assets, 0.01 for JPY)
func (order Order) DetermineOnePip() (decimal.Decimal, error) {

	onePip := decimal.NewFromInt(1)

	for i := 0; i < order.Asset.PipDecimals; i++ {
		onePip = onePip.Div(decimal.NewFromInt(10))
	}

	return onePip, nil
}

//Query setttings
func (order Order) QuerySettings() (models.Settings, error) {
	db := db.GetDB()

	var settings models.Settings

	err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)
	if err != nil {
		return settings, err
	}

	return settings, nil
}

//Determine default max allowed system leverage for asset by MarketID
func (order Order) DetermineDefaultLeverage() (decimal.Decimal, error) {

	var err error
	var leverage decimal.Decimal

	settings, err := order.QuerySettings()
	if err != nil {
		return leverage, err
	}

	//Crypto
	if order.Asset.MarketID == 1 {
		leverage = settings.LeverageAllowedCrypto.Decimal
	}

	//Stock (nasdaq, it, cannabis)
	if order.Asset.MarketID == 2 || order.Asset.MarketID == 4 || order.Asset.MarketID == 6 {
		leverage = settings.LeverageAllowedStock.Decimal
	}

	//Forex
	if order.Asset.MarketID == 3 {
		leverage = settings.LeverageAllowedForex.Decimal
	}

	//Commodities
	if order.Asset.MarketID == 5 {
		leverage = settings.LeverageAllowedCommodities.Decimal
	}

	//Indices
	if order.Asset.MarketID == 7 {
		leverage = settings.LeverageAllowedIndices.Decimal
	}

	return leverage, nil
}

//Determine max leverage value allowed for current member for this asset
func (order Order) DetermineLeverageAllowed() (decimal.Decimal, error) {

	//Default system leverage from settings for this asset by MarketID
	leverageAllowed := order.Leverage.Decimal

	//Check current asset leverage param. If it is not default (0), use it to overwrite system default value
	if !order.Asset.LeverageAllowed.Decimal.IsZero() {
		leverageAllowed = order.Asset.LeverageAllowed.Decimal
	}

	//If the leverage is bigger than max member leverage, set leverage as max member leverage
	if order.Member.LeverageAllowed.Decimal.LessThan(leverageAllowed) {
		leverageAllowed = order.Member.LeverageAllowed.Decimal
	}

	return leverageAllowed, nil
}

//Determine total real market value
func (order Order) DetermineTotalReal() (decimal.Decimal, error) {

	var totalReal decimal.Decimal

	//A Forex lot (MarketID is 3)
	// 1 lot = 100 000 = need to have 1 000$ on balance
	// Mini Lot 0.1 lot = 10 000 = need to have 100$ on balance
	// Micro Lot 0.01 lot = 1000 = need to have 10$ on balance
	// Nano Lot 0.001 lot = 100 = need to have 1$ on balance
	if order.Asset.MarketID == 3 {
		totalReal = order.Qty.Decimal.Mul(decimal.NewFromInt(100000))
	}

	//Not a Forex lot
	//Total qty = order.Qty * rate
	if order.Asset.MarketID != 3 {
		totalReal = order.Qty.Decimal.Mul(order.RateEntry.Decimal)
	}

	return totalReal, nil
}

//Close order
func (order Order) Close() error {
	var err error

	//Order is pending
	if order.Status == STATUS_PENDING {
		order.CancelPending()
	}

	//Order is open
	if order.Status == STATUS_OPEN {
		order, err = order.CalculateProfit()
		if err != nil {
			return err
		}

		//Close order in database
		order.CloseOrder()
	}

	return nil
}

//Open creates new order and returns its' TradeID
func (order Order) Open() (int64, error) {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET "+order.Asset.Currency+"="+order.Asset.Currency+"-$1 WHERE MemberID=$2", order.Total.Decimal, order.MemberID)
	tx.MustExec("INSERT INTO Trade (MemberID, AssetID, Type, Action, MemberRate, MarketRate, RateEntry, Qty, TotalReal, Total, BalanceEntry, StopLoss, TakeProfit,  OnePip, PipsRateEntry, Leverage, Status, Timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)", order.MemberID, order.Asset.AssetID, order.Type, order.Action, order.MemberRate.Decimal, order.Asset.Rate.Decimal, order.RateEntry.Decimal, order.Qty.Decimal, order.TotalReal.Decimal, order.Total.Decimal, order.BalanceEntry.Decimal, order.StopLoss.Decimal, order.TakeProfit.Decimal, order.OnePip.Decimal, order.PipsRateEntry.Decimal, order.Leverage.Decimal, order.Status, order.Timestamp)

	//Add asset to wallet balance if the buy order was opened instantly
	if order.Action == ACTION_BUY && order.Status == STATUS_OPEN {
		tx.MustExec("UPDATE Wallet SET Balance=Balance+$1 WHERE MemberID=$2 AND AssetID=$3", order.Qty.Decimal, order.MemberID, order.Asset.AssetID)
	}
	tx.Commit()

	var RecordID int64

	if err := db.Get(&RecordID, "SELECT TradeID FROM Trade WHERE MemberID=$1 AND AssetID=$2 AND Type=$3 AND Action=$4 AND Status=$5 AND Timestamp=$6 ORDER BY TradeID DESC", order.MemberID, order.Asset.AssetID, order.Type, order.Action, order.Status, order.Timestamp); err != nil {
		return RecordID, err
	}

	tx = db.MustBegin()
	tx.MustExec("INSERT INTO History (MemberID, AssetID, TradeID, Type, Action, Status, Currency, Qty, Rate, Leverage, Profit, ProfitAbs, ProfitNegative, Timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", order.MemberID, order.Asset.AssetID, RecordID, order.Type, order.Action, order.Status, order.Asset.Currency, order.Qty.Decimal, order.RateEntry.Decimal, order.Leverage.Decimal, order.Total.Decimal.Mul(decimal.NewFromInt(-1)), order.Total.Decimal.Abs(), true, order.Timestamp)
	tx.Commit()

	return RecordID, nil
}

//CalculateProfit returns order struct with profit calculation and calls a function to record it to database
func (order Order) CalculateProfit() (Order, error) {
	var err error

	//Don't calculate the profit for pending orders
	if order.Status == STATUS_PENDING {
		return order, nil
	}

	if order.Action == ACTION_BUY {
		order.RateClosed = order.Asset.RateSell
	} else {
		order.RateClosed = order.Asset.RateBuy
	}

	//Calculate final order TotalReal
	var newTotalReal = order.RateClosed.Decimal.Mul(order.Qty.Decimal)

	var newTotal = newTotalReal

	//Apply leverage to TotalReal (if it's not a Forex asset)
	if order.Asset.MarketID != 3 {
		newTotal = newTotal.Div(order.Leverage.Decimal)
	}

	if order.Action == ACTION_BUY {
		order.Profit.Decimal = newTotal.Sub(order.Total.Decimal)
	} else {
		order.Profit.Decimal = order.Total.Decimal.Sub(newTotal)
	}

	//Forex
	if order.Asset.MarketID == 3 {
		order.Profit.Decimal, err = order.DetermineForexProfit()
		if err != nil {
			return order, err
		}
		if order.Action == ACTION_SELL {
			order.Profit.Decimal = order.Profit.Decimal.Mul(decimal.NewFromInt(-1))
		}

	}

	//bad gain formula: order.Gain.Decimal = order.Profit.Decimal.Div(order.MarketRate.Decimal).Mul(decimal.NewFromInt(100))

	//How much % profit asset gained (or lost) since order creation
	order.Gain.Decimal = order.Asset.Rate.Decimal.Mul(decimal.NewFromInt(100)).Div(order.MarketRate.Decimal).Sub(decimal.NewFromInt(100))

	if order.Action == ACTION_SELL {
		order.Gain.Decimal = order.Gain.Decimal.Mul(decimal.NewFromInt(-1))
	}

	//Forex gain leverage fix
	if order.Asset.MarketID == 3 {
		order.Gain.Decimal = order.Gain.Decimal.Mul(order.Leverage.Decimal)
	}

	//Leverage calculation (if it's not a Forex asset)
	if order.Asset.MarketID != 3 {

		//Restore profit by multiplying with leverage
		order.Profit.Decimal = order.Profit.Decimal.Mul(order.Leverage.Decimal)

		//Apply leverage to profit by multiplying with leverage once again
		order.Profit.Decimal = order.Profit.Decimal.Mul(order.Leverage.Decimal)

	}

	//Get member account balance
	order.BalanceClosed.Decimal, err = order.QueryCurrentBalance()
	if err != nil {
		return order, err
	}

	order.BalanceClosed.Decimal = order.BalanceClosed.Decimal.Add(order.Profit.Decimal)

	//Profit absolute value (no negative sign)
	order.ProfitAbs.Decimal = order.Profit.Decimal.Abs()

	//Order total profit is less than 0
	if order.Profit.Decimal.IsNegative() {
		order.ProfitNegative = true
	} else {
		order.ProfitNegative = false
	}

	//Update order profit
	order.UpdateProfit()

	return order, nil
}

//Update order profit
func (order Order) UpdateProfit() error {
	db := db.GetDB()

	tx := db.MustBegin()
	tx.MustExec("UPDATE Trade SET Profit=$1, ProfitAbs=$2, ProfitNegative=$3, Gain=$4 WHERE TradeID=$5", order.Profit.Decimal, order.ProfitAbs.Decimal, order.ProfitNegative, order.Gain.Decimal, order.TradeID)
	tx.Commit()

	return nil
}

//Close order in database
func (order Order) CloseOrder() error {
	db := db.GetDB()

	//Return money used to purchase the order and add profit to it
	Profit := order.Total.Decimal.Add(order.Profit.Decimal)

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET "+order.Asset.Currency+"="+order.Asset.Currency+"+$1 WHERE MemberID=$2", Profit, order.MemberID)
	tx.MustExec("UPDATE Trade SET Status=$1, BalanceClosed=$2, RateClosed=$3, ClosedBySystem=$4, DateClosed=current_timestamp WHERE TradeID=$5", STATUS_CLOSED, order.BalanceClosed.Decimal, order.RateClosed.Decimal, order.ClosedBySystem, order.TradeID)
	tx.MustExec("INSERT INTO History (MemberID, AssetID, TradeID, Type, Action, Status, Currency, Qty, Rate, Leverage, Profit, ProfitAbs, ProfitNegative, Timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", order.MemberID, order.Asset.AssetID, order.TradeID, order.Type, order.Action, STATUS_CLOSED, order.Asset.Currency, order.Qty.Decimal, order.RateClosed.Decimal, order.Leverage.Decimal, Profit, Profit.Abs(), Profit.IsNegative(), order.Timestamp)

	//Deduct member asset balance is case of long order
	if order.Action == ACTION_BUY {
		tx.MustExec("UPDATE Wallet SET Balance=Balance-$1 WHERE MemberID=$2 AND AssetID=$3", order.Qty.Decimal, order.MemberID, order.Asset.AssetID)
	}
	tx.Commit()

	return nil

}

//Determine forex profit
func (order Order) DetermineForexProfit() (decimal.Decimal, error) {

	// How much money to pay user per pips (pip value)
	if order.Asset.BaseCurrency == CURRENCY_USD {
		order.PipValue.Decimal = decimal.NewFromInt(10).Mul(order.Qty.Decimal)
	} else {
		order.PipValue.Decimal = order.OnePip.Decimal.Div(order.RateClosed.Decimal).Mul(order.TotalReal.Decimal)
	}

	//How much pips Forex lot rate had at order closure time
	order.PipsRateClosed.Decimal = order.RateClosed.Decimal.Div(order.OnePip.Decimal)

	//Difference in pips asset rate between order creation and order closure
	difference := order.PipsRateClosed.Decimal.Sub(order.PipsRateEntry.Decimal)

	//On order close pay amount of pips (difference) * forex.PipValue
	earnings := difference.Mul(order.PipValue.Decimal)

	return earnings, nil
}

//Cancel pending limit order
func (order Order) CancelPending() error {
	db := db.GetDB()
	var err error

	//Get member account balance
	order.BalanceClosed.Decimal, err = order.QueryCurrentBalance()
	if err != nil {
		return err
	}

	//Member receives money on order closure, add them to member account balance
	order.BalanceClosed.Decimal = order.BalanceClosed.Decimal.Add(order.Total.Decimal)

	tx := db.MustBegin()
	tx.MustExec("UPDATE Trade SET Status=$1, BalanceClosed=$2, Profit=$3, ProfitAbs=$4, DateClosed=current_timestamp WHERE TradeID=$5", STATUS_CANCELLED, order.BalanceClosed.Decimal, 0, 0, order.TradeID)
	tx.MustExec("UPDATE Member SET "+order.Asset.Currency+"="+order.Asset.Currency+"+$1 WHERE MemberID=$2", order.Total.Decimal, order.MemberID)
	tx.MustExec("INSERT INTO History (MemberID, AssetID, TradeID, Type, Action, Status, Currency, Qty, Rate, Leverage, Profit, ProfitAbs, ProfitNegative, Timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", order.MemberID, order.Asset.AssetID, order.TradeID, order.Type, order.Action, STATUS_CANCELLED, order.Asset.Currency, order.Qty.Decimal, order.RateClosed.Decimal, order.Leverage.Decimal, order.Total.Decimal, order.Total.Decimal.Abs(), false, order.Timestamp)
	tx.Commit()

	return nil
}

//Close order that meets stop loss / take profit requirements
func (order Order) CloseSLTP() error {
	var err error

	//Stop if order doesn't have profit
	if order.ProfitAbs.Decimal.IsZero() {
		return nil
	}

	//Determine stop loss / take profit values for order
	stopLoss, takeProfit, err := order.DetermineOrderSLTP()
	if err != nil {
		return err
	}

	//In case the order is closed by system, set the following flag
	order.ClosedBySystem = true

	//StopLoss requirement met
	if stopLoss.IsZero() == false && order.ProfitNegative && order.Gain.Decimal.Abs().GreaterThanOrEqual(stopLoss) {
		order.CloseOrder()
	}

	//TakeProfit requirement met
	if takeProfit.IsZero() == false && order.ProfitNegative == false && order.Gain.Decimal.Abs().GreaterThanOrEqual(takeProfit) {
		order.CloseOrder()
	}

	return nil
}

//Determine stop loss / take profit values for order
func (order Order) DetermineOrderSLTP() (decimal.Decimal, decimal.Decimal, error) {

	var err error
	stopLoss := decimal.NewFromInt(0)
	takeProfit := decimal.NewFromInt(0)

	settings, err := order.QuerySettings()
	if err != nil {
		return stopLoss, takeProfit, err
	}

	//Order has this setting
	if order.StopLoss.Decimal.IsZero() == false {
		stopLoss = order.StopLoss.Decimal

		//Default system setting (system max order profit protection)
	} else if settings.StopLossProtection.Decimal.IsZero() == false {
		stopLoss = settings.StopLossProtection.Decimal
	}

	//Order has this setting
	if order.TakeProfit.Decimal.IsZero() == false {
		takeProfit = order.TakeProfit.Decimal

		//Default system setting (system max order loss protection)
	} else if settings.TakeProfitProtection.Decimal.IsZero() == false {
		takeProfit = settings.TakeProfitProtection.Decimal
	}

	return stopLoss, takeProfit, nil
}

//Determine max allowed stop loss / take profit values for member
func (order Order) DetermineMaxAllowedSLTP() (decimal.Decimal, decimal.Decimal, error) {

	var err error
	stopLoss := decimal.NewFromInt(0)
	takeProfit := decimal.NewFromInt(0)

	settings, err := order.QuerySettings()
	if err != nil {
		return stopLoss, takeProfit, err
	}

	//Member has this setting set up by admin
	if order.Member.StopLossAllowed.Decimal.IsZero() == false {
		stopLoss = order.Member.StopLossAllowed.Decimal

		//Default system setting (system max order profit default allowance)
	} else if settings.StopLossAllowed.Decimal.IsZero() == false {
		stopLoss = settings.StopLossAllowed.Decimal
	}

	//Member has this setting set up by admin
	if order.Member.TakeProfitAllowed.Decimal.IsZero() == false {
		takeProfit = order.Member.TakeProfitAllowed.Decimal

		//Default system setting (system max order loss default allowance)
	} else if settings.TakeProfitAllowed.Decimal.IsZero() == false {
		takeProfit = settings.TakeProfitAllowed.Decimal
	}

	return stopLoss, takeProfit, nil
}
