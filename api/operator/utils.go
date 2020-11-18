package operator

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/jwt"
)

//Middleware to check member permission
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sender, err := QueryMember(c)
		if err != nil {
			c.Abort()
			return
		}

		//2 - operator, 3 - admin
		if sender.Role != 2 && sender.Role != 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
			})
			c.Abort()
			return
		}
	}
}

func QueryMember(c *gin.Context) (Member, error) {
	db := db.GetDB()

	MemberID := c.MustGet(jwt.MemberKey).(string)

	var member Member

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
