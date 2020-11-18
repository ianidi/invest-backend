package member

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/jwt"
	"github.com/ianidi/exchange-server/internal/models"
)

func QueryMember(c *gin.Context) (models.Member, error) {
	db := db.GetDB()

	MemberID := c.MustGet(jwt.MemberKey).(string)

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1", MemberID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_MEMBER_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		c.Abort()
		return member, err
	}

	return member, nil
}
