package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/shopspring/decimal"
)

// BalanceDeposit
// @Summary
// @Description BalanceDeposit
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Balance-Deposit
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /balance/deposit [get]
func BalanceDeposit(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		Amount float64 `json:"Amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if query.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_AMOUNT"})
		return
	}

	Amount := sender.USD.Decimal.Add(decimal.NewFromFloat(query.Amount))

	if Amount.GreaterThanOrEqual(decimal.NewFromInt(50000)) {
		Amount = decimal.NewFromInt(50000)
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Deposit (MemberID, Amount) VALUES ($1, $2)", sender.MemberID, query.Amount)
	tx.Commit()

	c.JSON(200, gin.H{
		"status":  true,
		"balance": Amount,
	})
}

// BalanceWithdraw
// @Summary
// @Description BalanceWithdraw
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Balance-Withdraw
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /balance/withdraw [get]
func BalanceWithdraw(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		Amount float64 `json:"Amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if query.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_AMOUNT"})
		return
	}

	//TODO: select all withdrawal requests and count total sum. it should be bigger or equal to balance, esle throw INSUFFICIENT

	if decimal.NewFromFloat(query.Amount).GreaterThan(sender.USD.Decimal) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "WITHDRAW_INSUFFICIENT"})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Withdrawal (MemberID, Amount) VALUES ($1, $2)", sender.MemberID, query.Amount)
	tx.Commit()

	c.JSON(200, gin.H{
		"status":  true,
		"balance": sender.USD.Decimal.Sub(decimal.NewFromFloat(query.Amount)),
	})
}
