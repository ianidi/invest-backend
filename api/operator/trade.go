package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
)

// TradeGet
// @Summary
// @Description TradeGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Trade-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} Trade
// @Failure 400 {object} Error
// @Router /operator/trade [get]
func TradeGet(c *gin.Context) {
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
		query.Limit = 100
	}

	var trade []*Trade

	if err := db.Select(&trade, "SELECT * FROM Trade ORDER BY TradeID DESC OFFSET $1 LIMIT $2", query.Offset, query.Limit); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_TRADE_RECORD",
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
		"result": trade,
	})
}
