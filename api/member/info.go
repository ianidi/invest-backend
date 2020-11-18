package member

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
)

// InfoGet
// @Summary
// @Description InfoGet
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Info-Get
// @Success 200 {object} Asset
// @Failure 400 {object} Error
// @Router /info/init [get]
func InfoGetInit(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var asset []*models.Asset

	err = db.Select(&asset, "SELECT * FROM Asset WHERE Active=$1 ORDER BY Priority DESC", true)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"result": nil,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	now := time.Now().Unix()

	for _, assetRow := range asset {
		if err := db.Select(&assetRow.Performance, "SELECT * FROM Rate WHERE AssetID=$1 AND Timestamp>$2 ORDER BY RateID ASC LIMIT 8", assetRow.AssetID, now-86400); err != nil {
			if err != sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": false,
					"error":  err.Error(),
				})
			}
			return
		}
	}

	var wallet []*models.Wallet

	err = db.Select(&wallet, "SELECT * FROM Wallet WHERE MemberID=$1 AND Balance>$2", sender.MemberID, 0)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var fave []*models.Fave

	err = db.Select(&fave, "SELECT * FROM Fave WHERE MemberID=$1", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var history []*models.History

	err = db.Select(&history, "SELECT * FROM History WHERE MemberID=$1 ORDER BY HistoryID DESC", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var trade []*models.Trade

	err = db.Select(&trade, "SELECT * FROM Trade WHERE MemberID=$1 ORDER BY TradeID DESC", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var alert []*models.Alert

	err = db.Select(&alert, "SELECT * FROM Alert WHERE MemberID=$1 AND Status=$2 ORDER BY AlertID DESC", sender.MemberID, true)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var settings models.Settings

	err = db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)

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
		"status":                     true,
		"email":                      sender.Email,
		"balanceUSD":                 sender.USD,
		"balanceEUR":                 sender.EUR,
		"MemberStopLossAllowed":      sender.StopLossAllowed,
		"MemberTakeProfitAllowed":    sender.TakeProfitAllowed,
		"MemberLeverageAllowed":      sender.LeverageAllowed,
		"StopLossAllowed":            settings.StopLossAllowed,
		"TakeProfitAllowed":          settings.TakeProfitAllowed,
		"LeverageAllowedCrypto":      settings.LeverageAllowedCrypto,
		"LeverageAllowedStock":       settings.LeverageAllowedStock,
		"LeverageAllowedForex":       settings.LeverageAllowedForex,
		"LeverageAllowedCommodities": settings.LeverageAllowedCommodities,
		"LeverageAllowedIndices":     settings.LeverageAllowedIndices,

		"asset":   asset,
		"alerts":  alert,
		"fave":    fave,
		"wallet":  wallet,
		"history": history,
		"trade":   trade,
	})
}

// InfoGet
// @Summary
// @Description InfoGet
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Info-Get
// @Success 200 {object} Asset
// @Failure 400 {object} Error
// @Router /info [get]
func InfoGet(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var wallet []*models.Wallet

	err = db.Select(&wallet, "SELECT * FROM Wallet WHERE MemberID=$1 AND Balance>$2", sender.MemberID, 0)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var fave []*models.Fave

	err = db.Select(&fave, "SELECT * FROM Fave WHERE MemberID=$1", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var history []*models.History

	err = db.Select(&history, "SELECT * FROM History WHERE MemberID=$1 ORDER BY HistoryID DESC", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var trade []*models.Trade

	err = db.Select(&trade, "SELECT * FROM Trade WHERE MemberID=$1 ORDER BY TradeID DESC", sender.MemberID)

	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var settings models.Settings

	err = db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)

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
		"status":                     true,
		"balanceUSD":                 sender.USD,
		"balanceEUR":                 sender.EUR,
		"MemberStopLossAllowed":      sender.StopLossAllowed,
		"MemberTakeProfitAllowed":    sender.TakeProfitAllowed,
		"MemberLeverageAllowed":      sender.LeverageAllowed,
		"StopLossAllowed":            settings.StopLossAllowed,
		"TakeProfitAllowed":          settings.TakeProfitAllowed,
		"LeverageAllowedCrypto":      settings.LeverageAllowedCrypto,
		"LeverageAllowedStock":       settings.LeverageAllowedStock,
		"LeverageAllowedForex":       settings.LeverageAllowedForex,
		"LeverageAllowedCommodities": settings.LeverageAllowedCommodities,
		"LeverageAllowedIndices":     settings.LeverageAllowedIndices,

		"fave":    fave,
		"wallet":  wallet,
		"history": history,
		"trade":   trade,
	})
}
