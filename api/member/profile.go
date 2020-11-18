package member

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/ianidi/exchange-server/internal/db"
)

// ProfileGet вернуть данные профиля пользователя
// @Summary вернуть данные профиля пользователя
// @Description ProfileGet
// @Tags Member, Profile
// @Accept  json
// @Produce  json
// @ID Member-Profile-Get
// @Success 200 {object} Member
// @Failure 400 {object} Error
// @Router /profile [get]
func ProfileGet(c *gin.Context) {
	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": sender,
	})
}

// ProfileUpdate обновить профиль пользователя
// @Summary обновить профиль пользователя
// @Description ProfileUpdate
// @Tags Member, Profile
// @Accept  json
// @Produce  json
// @ID Member-Profile-Update
// @Param   Name				query		string		true		"Имя"
// @Param   Surname			query		string		true		"Фамилия"
// @Param   Lastname		query		string		false		"Отчество"
// @Param   Birthday		query		string		true		"Дата рождения в формате YYYY-MM-DD"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /profile/update [post]
func ProfileUpdate(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		FirstName   string `json:"FirstName" binding:"required"`
		LastName    string `json:"LastName" binding:"required"`
		Birthday    string `json:"Birthday" binding:"required"`
		Citizenship string `json:"Citizenship"`
		Gender      string `json:"Gender" binding:"required"`
		Country     string `json:"Country"`
		City        string `json:"City"`
		Zip         string `json:"Zip"`
		Address1    string `json:"Address1"`
		Address2    string `json:"Address2"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if err := validation.Validate(query.FirstName,
		validation.Required,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_FIRSTNAME",
		})
		return
	}

	if err := validation.Validate(query.LastName,
		validation.Required,
		validation.Length(2, 30),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_LASTNAME",
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
		//validation.Length(10, 10),
		validation.Date("2006-01-02").Min(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).Max(time.Date(time.Now().Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_BIRTHDAY",
		})
		return
	}

	//TODO: validate counntry code
	if err := validation.Validate(query.Citizenship,
		validation.Length(2, 2),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_CITIZENSHIP",
		})
		return
	}

	//TODO: validate counntry code
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

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET FirstName=$1, LastName=$2, Birthday=$3, Citizenship=$4, Country=$5, City=$6, Zip=$7, Address1=$8, Address2=$9, Gender=$10 WHERE MemberID=$11", query.FirstName, query.LastName, query.Birthday, query.Citizenship, query.Country, query.City, query.Zip, query.Address1, query.Address2, query.Gender, sender.MemberID)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// ProfileNotifications
// @Summary
// @Description ProfileNotifications
// @Tags Member, Profile
// @Accept  json
// @Produce  json
// @ID Profile-Notifications
// @Param   EmailNotifications	query		string		true		"Значение параметра"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /profile/notifications [post]
func ProfileNotifications(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		EmailNotifications bool `json:"EmailNotifications"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET EmailNotifications=$1 WHERE MemberID=$2", query.EmailNotifications, sender.MemberID)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"status": true})
}
