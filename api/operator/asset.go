package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/shopspring/decimal"
)

// AssetGetByID
// @Summary
// @Description AssetGetByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Asset-Get-By-ID
// @Param   id	path		int		true		"v ID"
// @Success 200 {object} Asset
// @Failure 400 {object} Error
// @Router /operator/asset/{id} [get]
func AssetGetByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		AssetID int `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var asset Asset

	if err := db.Get(&asset, "SELECT * FROM Asset WHERE AssetID=$1", query.AssetID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_ASSET_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": asset,
	})
}

// AssetGet
// @Summary
// @Description AssetGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Asset-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} Asset
// @Failure 400 {object} Error
// @Router /operator/asset [get]
func AssetGet(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		Offset int `form:"offset"`
		Limit  int `form:"limit"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	if query.Limit == 0 {
		query.Limit = 1000
	}

	var asset []*Asset

	if err := db.Select(&asset, "SELECT * FROM Asset ORDER BY MarketID ASC OFFSET $1 LIMIT $2", query.Offset, query.Limit); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_ASSET_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": asset,
	})
}

// AssetUpdateByID
// @Summary
// @Description AssetUpdateByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Asset-Update-By-ID
// @Param   AssetID					query		int				true		"ID"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/asset/update [post]
func AssetUpdateByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		AssetID         int     `json:"AssetID" binding:"required"`
		Title           string  `json:"Title" binding:"required"`
		Description     string  `json:"Description" binding:"required"`
		BuySpread       float64 `json:"BuySpread" binding:"required"`
		SellSpread      float64 `json:"SellSpread" binding:"required"`
		DecimalScale    int     `json:"DecimalScale" binding:"required"`
		Priority        int     `json:"Priority" binding:"required"`
		Sentiment       int     `json:"Sentiment" binding:"required"`
		SentimentType   string  `json:"SentimentType" binding:"required"`
		LeverageAllowed int     `json:"LeverageAllowed"`
		Tradable        bool    `json:"Tradable"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var asset Asset

	if err := db.Get(&asset, "SELECT * FROM Asset WHERE AssetID=$1", query.AssetID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var rateBuy decimal.Decimal
	var rateSell decimal.Decimal

	var BuySpread = decimal.NewFromFloat(query.BuySpread)
	var SellSpread = decimal.NewFromFloat(query.SellSpread)

	//Forex spread has different formula (counted in pips)
	if asset.MarketID == 3 {
		rateBuy = asset.Rate.Decimal.Add(BuySpread)
		rateSell = asset.Rate.Decimal.Sub(SellSpread)
	} else {
		rateBuy = asset.Rate.Decimal.Add(asset.Rate.Decimal.Mul(BuySpread.Div(decimal.NewFromInt(100))))
		rateSell = asset.Rate.Decimal.Sub(asset.Rate.Decimal.Mul(SellSpread.Div(decimal.NewFromInt(100))))
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Asset SET Title=$1, Description=$2, BuySpread=$3, SellSpread=$4, DecimalScale=$5, Priority=$6, Sentiment=$7, SentimentType=$8, Tradable=$9, RateBuy=$10, RateSell=$11, LeverageAllowed=$12 WHERE AssetID=$13", query.Title, query.Description, query.BuySpread, query.SellSpread, query.DecimalScale, query.Priority, query.Sentiment, query.SentimentType, query.Tradable, rateBuy, rateSell, query.LeverageAllowed, query.AssetID)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
