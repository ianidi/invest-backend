package operator

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/ianidi/exchange-server/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// MemberGetByID
// @Summary
// @Description MemberGetByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Member-Get-By-ID
// @Param   id	path		int		true		"Member ID"
// @Success 200 {object} Member
// @Failure 400 {object} Error
// @Router /operator/member/{id} [get]
func MemberGetByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		MemberID int `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	sender, err := QueryMember(c)
	if err != nil {
		return
	}

	var member Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1 AND Role<=$2", query.MemberID, sender.Role); err != nil {
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
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": member,
	})
}

// MemberGet
// @Summary
// @Description MemberGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Member-Get
// @Param   Offset	query		int		false		"Offset"
// @Param   Limit		query		int		false		"Limit"
// @Success 200 {object} MemberMerged
// @Failure 400 {object} Error
// @Router /operator/member [get]
func MemberGet(c *gin.Context) {
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

	sender, err := QueryMember(c)
	if err != nil {
		return
	}

	var member []*Member

	if err := db.Select(&member, "SELECT * FROM Member WHERE Role<=$1 ORDER BY MemberID DESC OFFSET $2 LIMIT $3", sender.Role, query.Offset, query.Limit); err != nil {

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
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": member,
	})
}

// MemberCreate
// @Summary
// @Description MemberCreate
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Member-Create
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/member/create [post]
func MemberCreate(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		return
	}

	var query struct {
		Password           string  `json:"Password" binding:"required"`
		Email              string  `json:"Email" binding:"required"`
		Role               int     `json:"Role"`
		Name               string  `json:"Name"`
		Surname            string  `json:"Surname"`
		Birthday           string  `json:"Birthday"`
		Citizenship        string  `json:"Citizenship"`
		Gender             string  `json:"Gender"`
		Country            string  `json:"Country"`
		City               string  `json:"City"`
		Zip                string  `json:"Zip"`
		Address1           string  `json:"Address1"`
		Address2           string  `json:"Address2"`
		USD                float64 `json:"USD"`
		EUR                float64 `json:"EUR"`
		StopLossAllowed    int     `json:"StopLossAllowed"`
		TakeProfitAllowed  int     `json:"TakeProfitAllowed"`
		LeverageAllowed    int     `json:"LeverageAllowed"`
		EmailNotifications bool    `json:"EmailNotifications"`
		Status             string  `json:"Status"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	//TODO: verify status

	if err := validation.Validate(query.Email,
		validation.Required,
		is.Email,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_EMAIL",
		})
		return
	}

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Email=$1", query.Email); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "EMAIL_ALREADY_IN_USE",
		})
		return
	}

	query.Email = strings.ToLower(query.Email)

	if err := validation.Validate(query.Name,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_NAME",
		})
		return
	}

	if err := validation.Validate(query.Surname,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_SURNAME",
		})
		return
	}

	if (query.Role != 1 && query.Role != 2 && query.Role != 3) || query.Role >= sender.Role {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ROLE",
		})
		return
	}

	if query.Gender != "" && query.Gender != "m" && query.Gender != "f" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_GENDER",
		})
		return
	}

	//Year not less 1900 & not bigger than current. Date format YYYY-MM-DD
	if err := validation.Validate(query.Birthday,
		validation.Date("2006-01-02").Min(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).Max(time.Date(time.Now().Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_BIRTHDAY",
		})
		return
	}

	if query.Birthday == "" {
		query.Birthday = "0001-01-01"
	}

	if err := validation.Validate(query.Citizenship,
		validation.Length(2, 2),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_CITIZENSHIP",
		})
		return
	}

	if err := validation.Validate(query.Country,
		validation.Length(2, 2),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_COUNTRY",
		})
		return
	}

	if err := validation.Validate(query.City,
		validation.Length(0, 50),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_CITY",
		})
		return
	}

	if err := validation.Validate(query.Zip,
		validation.Length(0, 20),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ZIP",
		})
		return
	}

	if err := validation.Validate(query.Address1,
		validation.Length(0, 100),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ADDRESS1",
		})
		return
	}

	if err := validation.Validate(query.Address2,
		validation.Length(0, 100),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ADDRESS2",
		})
		return
	}

	PasswordHash, err := bcrypt.GenerateFromPassword([]byte(query.Password), 12)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	var created = time.Now().Unix()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Member (Email, PasswordHash, Role, Name, Surname, Birthday, Citizenship, Country, City, Zip, Address1, Address2, Gender, EmailNotifications, USD, EUR, StopLossAllowed, TakeProfitAllowed, LeverageAllowed, Status, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)", query.Email, PasswordHash, query.Role, query.Name, query.Surname, query.Birthday, query.Citizenship, query.Country, query.City, query.Zip, query.Address1, query.Address2, query.Gender, query.EmailNotifications, query.USD, query.EUR, query.StopLossAllowed, query.TakeProfitAllowed, query.LeverageAllowed, query.Status, created)
	tx.Commit()

	var RecordID int

	if err := db.Get(&RecordID, "SELECT MemberID FROM Member WHERE Email=$1 AND Created=$2", query.Email, created); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"status":   true,
		"RecordID": RecordID,
	})
}

// MemberUpdateByID
// @Summary
// @Description MemberUpdateByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Member-Update-By-ID
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/member/update [post]
func MemberUpdateByID(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		return
	}

	var query struct {
		MemberID           int    `json:"MemberID" binding:"required"`
		Email              string `json:"Email"`
		Password           string `json:"Password"`
		Role               int    `json:"Role"`
		Name               string `json:"Name"`
		Surname            string `json:"Surname"`
		Birthday           string `json:"Birthday"`
		Citizenship        string `json:"Citizenship"`
		Gender             string `json:"Gender"`
		Country            string `json:"Country"`
		City               string `json:"City"`
		Zip                string `json:"Zip"`
		Address1           string `json:"Address1"`
		Address2           string `json:"Address2"`
		EmailNotifications bool   `json:"EmailNotifications"`
		StopLossAllowed    int    `json:"StopLossAllowed"`
		TakeProfitAllowed  int    `json:"TakeProfitAllowed"`
		LeverageAllowed    int    `json:"LeverageAllowed"`
		Active             bool   `json:"Active"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if err := validation.Validate(query.Email,
		validation.Required,
		is.Email,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_EMAIL",
		})
		return
	}

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Email=$1 AND MemberID!=$2", query.Email, query.MemberID); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "EMAIL_ALREADY_IN_USE",
		})
		return
	}

	query.Email = strings.ToLower(query.Email)

	if err := validation.Validate(query.Name,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_NAME",
		})
		return
	}

	if err := validation.Validate(query.Surname,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_SURNAME",
		})
		return
	}

	if (query.Role != 1 && query.Role != 2 && query.Role != 3) || query.Role >= sender.Role {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ROLE",
		})
		return
	}

	if query.Gender != "" && query.Gender != "m" && query.Gender != "f" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_GENDER",
		})
		return
	}

	//Year not less 1900 & not bigger than current. Date format YYYY-MM-DD
	if err := validation.Validate(query.Birthday,
		validation.Date("2006-01-02").Min(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).Max(time.Date(time.Now().Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_BIRTHDAY",
		})
		return
	}

	if err := validation.Validate(query.Citizenship,
		validation.Length(2, 2),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_CITIZENSHIP",
		})
		return
	}

	if err := validation.Validate(query.Country,
		validation.Length(2, 2),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_COUNTRY",
		})
		return
	}

	if err := validation.Validate(query.City,
		validation.Length(0, 50),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_CITY",
		})
		return
	}

	if err := validation.Validate(query.Zip,
		validation.Length(0, 20),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ZIP",
		})
		return
	}

	if err := validation.Validate(query.Address1,
		validation.Length(0, 100),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ADDRESS1",
		})
		return
	}

	if err := validation.Validate(query.Address2,
		validation.Length(0, 100),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_ADDRESS2",
		})
		return
	}

	if query.Password != "" {
		PasswordHash, err := bcrypt.GenerateFromPassword([]byte(query.Password), 12)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}

		tx := db.MustBegin()
		tx.MustExec("UPDATE Member SET PasswordHash=$1 WHERE MemberID=$2", PasswordHash, query.MemberID)
		tx.Commit()
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET Email=$1, Role=$2, Name=$3, Surname=$4, Birthday=$5, Citizenship=$6, Country=$7, City=$8, Zip=$9, Address1=$10, Address2=$11, Gender=$12, EmailNotifications=$13, StopLossAllowed=$14, TakeProfitAllowed=$15, LeverageAllowed=$16, Active=$17 WHERE MemberID=$18", query.Email, query.Role, query.Name, query.Surname, query.Birthday, query.Citizenship, query.Country, query.City, query.Zip, query.Address1, query.Address2, query.Gender, query.EmailNotifications, query.StopLossAllowed, query.TakeProfitAllowed, query.LeverageAllowed, query.Active, query.MemberID)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}

// MemberDeleteByID
// @Summary
// @Description MemberDeleteByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-Member-Delete-By-ID
// @Param   MemberID	query		int				true		"ID"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/member/delete [delete]
func MemberDeleteByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		MemberID int `json:"MemberID" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	sender, err := QueryMember(c)
	if err != nil {
		return
	}

	var member Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE MemberID=$1 AND Role<$2 AND Role!=$3", query.MemberID, sender.Role, 2); err != nil {
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
		return
	}

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Member WHERE MemberID=$1", query.MemberID)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
