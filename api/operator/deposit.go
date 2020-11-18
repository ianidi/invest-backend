package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
)

// DepositGet
// @Summary
// @Description DepositGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Deposit-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} Deposit
// @Failure 400 {object} Error
// @Router /operator/deposit [post]
func DepositGet(c *gin.Context) {
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

	var deposit []*Deposit

	if err := db.Select(&deposit, "SELECT * FROM Deposit OFFSET $1 LIMIT $2", query.Offset, query.Limit); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_DEPOSIT_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	for _, depositRow := range deposit {
		if err := db.Get(&depositRow.MemberEmail, "SELECT Email FROM Member WHERE MemberID=$1", depositRow.MemberID); err != nil {
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
		"result": deposit,
	})
}

// DepositUpdate
// @Summary
// @Description DepositUpdate
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Deposit-Update
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/deposit/update [post]
func DepositUpdate(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		DepositID int  `json:"DepositID" binding:"required"`
		Action    bool `json:"Action"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var deposit Deposit

	if err := db.Get(&deposit, "SELECT * FROM Deposit WHERE DepositID=$1 AND Status=$2", query.DepositID, "pending"); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_DEPOSIT_RECORD",
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
	tx.MustExec("UPDATE Deposit SET Status=$1 WHERE DepositID=$2", status, query.DepositID)
	if status == "complete" {
		tx.MustExec("UPDATE Member SET USD=$1 WHERE MemberID=$2", deposit.Amount.Decimal, deposit.MemberID)
		tx.MustExec("INSERT INTO History (MemberID, Type, Status, Currency, Profit, ProfitAbs) VALUES ($1, $2, $3, $4, $5, $6)", deposit.MemberID, "balance", "deposit", "USD", deposit.Amount.Decimal, deposit.Amount.Decimal.Abs())
	}
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
