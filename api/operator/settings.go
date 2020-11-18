package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
)

// SettingsGet
// @Summary
// @Description SettingsGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Settings-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} History
// @Failure 400 {object} Error
// @Router /operator/settings [get]
func SettingsGet(c *gin.Context) {
	db := db.GetDB()

	var settings models.Settings

	if err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_SETTINGS_RECORD",
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
		"result": settings,
	})
}

// SettingsUpdate
// @Summary
// @Description SettingsUpdate
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Settings-Update-By-ID
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/settings/update [post]
func SettingsUpdate(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		Title                      string `json:"Title" binding:"required"`
		PlatformURL                string `json:"PlatformURL" binding:"required"`
		NewsURL                    string `json:"NewsURL" binding:"required"`
		SMTPHost                   string `json:"SMTPHost" binding:"required"`
		SMTPUsername               string `json:"SMTPUsername" binding:"required"`
		SMTPPassword               string `json:"SMTPPassword" binding:"required"`
		SMTPPort                   int    `json:"SMTPPort" binding:"required"`
		SMTPFromEmail              string `json:"SMTPFromEmail" binding:"required"`
		SMTPFromName               string `json:"SMTPFromName" binding:"required"`
		LeverageAllowedCrypto      int    `json:"LeverageAllowedCrypto" binding:"required"`
		LeverageAllowedStock       int    `json:"LeverageAllowedStock" binding:"required"`
		LeverageAllowedForex       int    `json:"LeverageAllowedForex" binding:"required"`
		LeverageAllowedCommodities int    `json:"LeverageAllowedCommodities" binding:"required"`
		LeverageAllowedIndices     int    `json:"LeverageAllowedIndices" binding:"required"`
		StopLossProtection         int    `json:"StopLossProtection" binding:"required"`
		TakeProfitProtection       int    `json:"TakeProfitProtection" binding:"required"`
		StopLossAllowed            int    `json:"StopLossAllowed" binding:"required"`
		TakeProfitAllowed          int    `json:"TakeProfitAllowed" binding:"required"`
		APIKeyIEX                  string `json:"APIKeyIEX" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Settings SET Title=$1, PlatformURL=$2, NewsURL=$3, SMTPHost=$4, SMTPUsername=$5, SMTPPassword=$6, SMTPPort=$7, SMTPFromEmail=$8, SMTPFromName=$9, 	LeverageAllowedCrypto=$10, LeverageAllowedStock=$11, LeverageAllowedForex=$12, LeverageAllowedCommodities=$13, LeverageAllowedIndices=$14, StopLossProtection=$15, TakeProfitProtection=$16, StopLossAllowed=$17, TakeProfitAllowed=$18, APIKeyIEX=$19 WHERE SettingsID=$20", query.Title, query.PlatformURL, query.NewsURL, query.SMTPHost, query.SMTPUsername, query.SMTPPassword, query.SMTPPort, query.SMTPFromEmail, query.SMTPFromName, query.LeverageAllowedCrypto, query.LeverageAllowedStock, query.LeverageAllowedForex, query.LeverageAllowedCommodities, query.LeverageAllowedIndices, query.StopLossProtection, query.TakeProfitProtection, query.StopLossAllowed, query.TakeProfitAllowed, query.APIKeyIEX, 1)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
