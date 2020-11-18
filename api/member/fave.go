package member

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
)

// FaveUpdate add / remove from favorites (watcheed assets)
// @Summary add / remove from favorites (watcheed assets)
// @Description FaveUpdate
// @Tags Member
// @Accept  json
// @Produce  json
// @ID Member-Fave-Update
// @Param   AssetID			query		int				true		"AssetID"
// @Param   Fave				query		bool			false		"Status"
// @Success 200 {object} Fave
// @Failure 400 {object} Error
// @Router /fave [post]
func FaveUpdate(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		AssetID int  `json:"AssetID" binding:"required"`
		Fave    bool `json:"Fave"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Asset WHERE AssetID=$1", query.AssetID); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ASSET",
		})
		return
	}

	if query.Fave == true {
		tx := db.MustBegin()
		tx.MustExec("DELETE FROM Fave WHERE MemberID=$1 AND AssetID=$2", sender.MemberID, query.AssetID)
		tx.MustExec("INSERT INTO Fave (MemberID, AssetID) VALUES ($1, $2)", sender.MemberID, query.AssetID)
		tx.Commit()
	}

	if query.Fave == false {
		tx := db.MustBegin()
		tx.MustExec("DELETE FROM Fave WHERE AssetID=$1 AND MemberID=$2", query.AssetID, sender.MemberID)
		tx.Commit()
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

	c.JSON(200, gin.H{
		"status": true,
		"fave":   fave,
	})
}
