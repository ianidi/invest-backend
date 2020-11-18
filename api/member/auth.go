package member

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/bcrypt"
)

// EmailUpdate запрос на обновление email пользователя
// @Summary запрос на обновление email пользователя
// @Description EmailUpdate
// @Tags Member, Email
// @Accept  json
// @Produce  json
// @ID Member-Email-Update
// @Param   Email	query		string		true		"Новый email адрес"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /profile/email/update [post]
func EmailUpdate(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		Email string `json:"Email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if err := validation.Validate(query.Email,
		validation.Length(0, 250),
		validation.Required,
		is.Email,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_EMAIL"})
		return
	}

	query.Email = strings.ToLower(query.Email)

	var count int

	//Существует ли пользователь с данным Email
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

	//Предотвратить запрос новых кодов подтверждения чаще, чем раз в 1 минуту
	if err := db.Get(&count, "SELECT count(*) FROM Verify WHERE MemberID=$1 AND Type=$2 AND Created>$3", sender.MemberID, "email_verification", time.Now().Unix()-60); err != nil {
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
			"error":  "TOO_OFTEN_EMAIL_UPDATE_REQUESTS",
		})
		return
	}

	//Сгенерировать код подтверждения
	Code, err := password.Generate(6, 6, 0, true, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	//Зашифровать код подтверждения в bcrypt
	CodeHash, err := bcrypt.GenerateFromPassword([]byte(Code), 12)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Type=$3 AND Status=$4", "cancelled", sender.MemberID, "email_verification", "pending")
	tx.MustExec("INSERT INTO Verify (MemberID, Code, Type, Email, Status, Created) VALUES ($1, $2, $3, $4, $5, $6)", sender.MemberID, CodeHash, "email_verification", query.Email, "pending", time.Now().Unix())
	tx.Commit()

	//Отправить код подтверждения на новый email
	//mail.SendMail(query.Email, "Email change request", "Please use code "+Code+" to update your email")

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// EmailVerify подтверждение обновления email пользователя с помощью кода подтверждения
// @Summary подтверждение обновления email пользователя с помощью кода подтверждения
// @Description EmailVerify
// @Tags Member, Email
// @Accept  json
// @Produce  json
// @ID Member-Email-Verify
// @Param   Code	query		string		true		"Код для подтверждения email"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /profile/email/verify [post]
func EmailVerify(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		Code string `json:"Code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	if err := validation.Validate(query.Code,
		validation.Length(6, 6),
		validation.Required,
		is.Digit,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	var verify models.Verify

	//Найти запросы на подтверждение от пользователя и проверить, что они не истекли по времени (сутки)
	if err := db.Get(&verify, "SELECT * FROM Verify WHERE MemberID=$1 AND Action=$2 AND Method=$3 AND Status=$4 AND Created>$5", sender.MemberID, "email", "signup", "pending", time.Now().Unix()-86400); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_VERIFY_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	//Проверка кода
	if err := bcrypt.CompareHashAndPassword([]byte(verify.CodeHash), []byte(query.Code)); err != nil {

		var status = "pending"
		var error = "WRONG_CODE"

		// Если число попыток ввода кода превышено
		if verify.Attempts == 2 {
			status = "fail"
			error = "TOO_MANY_ATTEMPTS"
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  error,
		})

		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Attempts=Attempts+$1, Status=$2 WHERE VerifyID=$3", 1, status, verify.VerifyID)
		tx.Commit()

		return
	}

	//Проверка пройдена, сменить email пользователя
	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET Email=$1 WHERE MemberID=$2", verify.Email, sender.MemberID)
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", "success", verify.VerifyID)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// ProfilePassword change member password
// @Summary change member password
// @Description ProfilePassword
// @Tags Member, Password
// @Accept  json
// @Produce  json
// @ID Member-Profile-Password
// @Param   PasswordCurrent	query		string		true		"Current password"
// @Param   PasswordNew			query		string		true		"New password"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /profile/password/update [post]
func ProfilePassword(c *gin.Context) {
	db := db.GetDB()

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	var query struct {
		PasswordCurrent string `json:"PasswordCurrent" binding:"required"`
		PasswordNew     string `json:"PasswordNew" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "REQUIRED", "type": "validation"})
		return
	}

	if err := validation.Validate(query.PasswordCurrent,
		validation.Length(6, 40),
		validation.Required,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_PASSWORD_CURRENT"})
		return
	}

	if err := validation.Validate(query.PasswordNew,
		validation.Length(6, 40),
		validation.Required,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_PASSWORD_NEW"})
		return
	}

	if query.PasswordCurrent == query.PasswordNew {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "PASSWORD_THE_SAME"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(sender.PasswordHash), []byte(query.PasswordCurrent)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "INVALID_PASSWORD_CURRENT"})
		return
	}

	PasswordNewHash, err := bcrypt.GenerateFromPassword([]byte(query.PasswordNew), 12)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Member SET PasswordHash=$1 WHERE MemberID=$2", PasswordNewHash, sender.MemberID)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"status": true})
}
