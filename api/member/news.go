package member

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
)

// NewsGet
// @Summary
// @Description NewsGet
// @Tags Public, News
// @Accept  json
// @Produce  json
// @ID Public-Info-Get
// @Success 200 {object} News
// @Failure 400 {object} Error
// @Router /news [get]
func NewsGet(c *gin.Context) {
	db := db.GetDB()

	var news []*models.News

	err := db.Select(&news, "SELECT * FROM News ORDER BY NewsID DESC LIMIT 3")

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": true,
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

	var settings models.Settings

	err = db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": true,
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

	c.JSON(200, gin.H{
		"status":  true,
		"result":  news,
		"NewsURL": settings.NewsURL,
	})
}
