package member

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
	"github.com/shopspring/decimal"
)

type PriceAlert struct {
	Price shopspring.Numeric
	Type  string
}

// AlertNew create price alert
// @Summary create price alert
// @Description AlertNew
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Alert-New
// @Success 200 {object} Alert
// @Failure 400 {object} Error
// @Router /alert [get]
func AlertNew(c *gin.Context) {
	db := db.GetDB()

	var alert PriceAlert

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		AssetID int    `json:"AssetID" binding:"required"`
		Price   string `json:"Price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	alert.Price.Decimal, err = decimal.NewFromString(query.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "ALERT_INVALID_PRICE"})
		return
	}

	if alert.Price.Decimal.IsZero() || alert.Price.Decimal.IsNegative() {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "ALERT_INVALID_PRICE"})
		return
	}

	var asset models.Asset

	if err := db.Get(&asset, "SELECT * FROM Asset WHERE AssetID=$1", query.AssetID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "INVALID_ASSET",
			})
			return
		}
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if alert.Price.Decimal.GreaterThan(asset.Rate.Decimal) {
		alert.Type = "h"
	} else {
		alert.Type = "l"
	}

	if alert.Price.Decimal.Equal(asset.Rate.Decimal) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "ALERT_RATE_THE_SAME",
		})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Alert (MemberID, AssetID, Type, Price, Currency) VALUES ($1, $2, $3, $4, $5)", sender.MemberID, asset.AssetID, alert.Type, alert.Price.Decimal, asset.Currency)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}

// AlertNew open price alert
// @Summary open price alert
// @Description AlertOpen
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Alert-New
// @Success 200 {object} Alert
// @Failure 400 {object} Error
// @Router /alert/open [get]
func AlertOpen(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Alert SET TimestampOpen=$1 WHERE MemberID=$2 AND Status=$3", time.Now().Unix(), sender.MemberID, true)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
