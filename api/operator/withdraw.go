package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
)

// WithdrawGet
// @Summary
// @Description WithdrawGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Withdraw-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} Withdrawal
// @Failure 400 {object} Error
// @Router /operator/withdraw [get]
func WithdrawGet(c *gin.Context) {
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

	var withdrawal []*Withdrawal

	if err := db.Select(&withdrawal, "SELECT * FROM Withdrawal OFFSET $1 LIMIT $2", query.Offset, query.Limit); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_WITHDRAWAL_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	for _, withdrawalRow := range withdrawal {
		if err := db.Get(&withdrawalRow.MemberEmail, "SELECT Email FROM Member WHERE MemberID=$1", withdrawalRow.MemberID); err != nil {
			if err != sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": false,
					"error":  err.Error(),
				})
			}
			return
		}
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": withdrawal,
	})
}

// WithdrawUpdate
// @Summary
// @Description WithdrawUpdate
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Withdraw-Update
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/withdraw/update [post]
func WithdrawUpdate(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		WithdrawalID int  `json:"WithdrawalID" binding:"required"`
		Action       bool `json:"Action"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var withdrawal Withdrawal

	if err := db.Get(&withdrawal, "SELECT * FROM Withdrawal WHERE WithdrawalID=$1 AND Status=$2", query.WithdrawalID, "pending"); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_WITHDRAWAL_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	var status = "declined"

	if query.Action == true {
		status = "complete"
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Withdrawal SET Status=$1 WHERE WithdrawalID=$2", status, query.WithdrawalID)
	if status == "complete" {
		tx.MustExec("UPDATE Member SET USD=USD-$1 WHERE MemberID=$2", withdrawal.Amount.Decimal, withdrawal.MemberID)
		tx.MustExec("INSERT INTO History (MemberID, Type, Status, Currency, Profit, ProfitAbs, ProfitNegative) VALUES ($1, $2, $3, $4, $5, $6, $7)", withdrawal.MemberID, "balance", "withdrawal", "USD", withdrawal.Amount.Decimal, withdrawal.Amount.Decimal.Abs(), true)
	}
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
