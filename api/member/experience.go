package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
)

// ExperienceUpdate
// @Summary
// @Description ExperienceUpdate
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Experience-New
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /experience [post]
func ExperienceUpdate(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		Question int    `json:"Question" binding:"required"`
		Answer   string `json:"Answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if query.Question <= 0 || query.Question >= 8 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_QUESTION"})
		return
	}

	//TODO: validate answer

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Experience WHERE MemberID=$1 AND Question=$2", sender.MemberID, query.Question)
	tx.MustExec("INSERT INTO Experience (MemberID, Question, Answer) VALUES ($1, $2, $3)", sender.MemberID, query.Question, query.Answer)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
