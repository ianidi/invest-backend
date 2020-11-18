package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/ianidi/exchange-server/internal/auth"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/mail"
	"github.com/ianidi/exchange-server/internal/otp"
	"github.com/ianidi/exchange-server/internal/utils"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// ProfileHandler struct
type ProfileHandler struct {
	RD AuthInterface
	TK TokenInterface
}

func NewProfile(rd AuthInterface, tk TokenInterface) *ProfileHandler {
	return &ProfileHandler{rd, tk}
}

func (h *ProfileHandler) Signin(c *gin.Context) {

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    auth.ACTION_SIGNIN,
		Method:    auth.METHOD_EMAIL,
	}

	var query struct {
		Email    string `json:"Email" binding:"required"`
		Password string `json:"Password" binding:"required"`
		Admin    bool   `json:"Admin"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if utils.Error(c, identity.ParseEmail(query.Email)) {
		return
	}

	if utils.Error(c, identity.CheckPassword(query.Password, false)) {
		return
	}

	if utils.Error(c, identity.QueryMemberByEmail()) {
		return
	}

	if utils.Error(c, identity.ValidatePassword()) {
		return
	}

	if query.Admin && identity.Member.Role < 2 {
		utils.Error(c, errors.New("INVALID_ROLE"))
		return
	}

	if utils.Error(c, h.RespondAuthorizationHeader(c, identity.Member.MemberID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}

func (h *ProfileHandler) Logout(c *gin.Context) {
	//If metadata is passed and the tokens valid, delete them from the redis store
	// metadata, _ := h.TK.ExtractTokenMetadata(c.Request)
	// if metadata != nil {
	// 	deleteErr := h.RD.DeleteTokens(metadata)
	// 	if deleteErr != nil {
	// 		c.JSON(http.StatusBadRequest, deleteErr.Error())
	// 		return
	// 	}
	// }
	c.JSON(http.StatusOK, "Successfully logged out")
}

func (h *ProfileHandler) Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]

	//verify the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("jwt_secret")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err)
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		userId, roleOk := claims["user_id"].(string)
		if roleOk == false {
			c.JSON(http.StatusUnprocessableEntity, "unauthorized")
			return
		}
		//Delete the previous Refresh Token
		delErr := h.RD.DeleteRefresh(refreshUuid)
		if delErr != nil { //if any goes wrong
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}

		if utils.Error(c, h.RespondAuthorizationHeader(c, cast.ToInt64(userId))) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": true})
	} else {

		if utils.Error(c, errors.New("REFRESH_EXPIRED")) {
			return
		}
	}
}

func (h *ProfileHandler) RespondAuthorizationHeader(c *gin.Context, MemberID int64) error {
	var err error

	//Create new pairs of refresh and access tokens
	ts, err := h.TK.CreateToken(cast.ToString(MemberID))
	if err != nil {
		return err
	}

	//Save the tokens metadata to redis
	// err = h.RD.CreateAuth(cast.ToString(MemberID), ts)
	// if err != nil {
	// 	return err
	// }

	c.Header(AuthorizationHeaderKey, ts.AccessToken)
	c.Header(RefreshHeaderKey, ts.RefreshToken)

	return nil
}

// AuthVerify auth verify email
// @Summary auth verify email
// @Description Auth verify
// @Tags Auth, Verify
// @Accept  json
// @Produce  json
// @ID Auth-Verify
// @Param   Action		query		string		true		"Action (signup/reset)"
// @Param   Password	query		string		true		"New password (reset)"
// @Param   Hash			query		string		true		"Email md5 hash"
// @Param   Code			query		string		true		"OTP code"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /auth/verify [post]
func (h *ProfileHandler) AuthVerify(c *gin.Context) {
	db := db.GetDB()

	var err error

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Method:    auth.METHOD_EMAIL,
	}

	var query struct {
		Action   string `json:"Action" binding:"required"`
		Password string `json:"Password"`
		Hash     string `json:"Hash" binding:"required"`
		Code     string `json:"Code" binding:"required"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if query.Action != auth.ACTION_SIGNUP && query.Action != auth.ACTION_RESET {
		utils.Error(c, errors.New("INVALID_ACTION"))
		return
	}

	identity.Action = query.Action

	if utils.Error(c, identity.QueryVerifyByHash(query.Hash)) {
		return
	}

	if utils.Error(c, identity.ValidateOTPCode(query.Code)) {
		return
	}

	//Reset member password
	if identity.Action == auth.ACTION_RESET {
		if utils.Error(c, identity.CheckPassword(query.Password, true)) {
			return
		}

		identity.PasswordHash, err = identity.HashBcrypt(identity.Password)
		if utils.Error(c, err) {
			return
		}

		tx := db.MustBegin()
		tx.MustExec("UPDATE Member SET PasswordHash=$1 WHERE MemberID=$2", identity.PasswordHash, identity.Verify.MemberID)
		tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", auth.STATUS_SUCCESS, identity.Verify.VerifyID)
		tx.Commit()

		if utils.Error(c, h.RespondAuthorizationHeader(c, identity.Verify.MemberID)) {
			return
		}
	}

	//Confirm member email after visiting sign up confirmation link
	if identity.Action == auth.ACTION_SIGNUP {

		// if utils.Error(c, identity.QueryOnboardingRecord(identity.Verify.MemberID)) {
		// 	return
		// }

		// if identity.Onboarding.Email != true {
		// 	tx := db.MustBegin()
		// 	tx.MustExec("UPDATE Onboarding SET Email=$1 WHERE MemberID=$2", true, identity.Verify.MemberID)
		// 	tx.Commit()
		// }

		//TODO: verify record being marked as success on full onboarding completion

		if utils.Error(c, h.RespondAuthorizationHeader(c, identity.Verify.MemberID)) {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// AuthSignup member signup
// @Summary member signup
// @Description Signup
// @Tags Auth, Signup
// @Accept  json
// @Produce  json
// @ID Auth-Signup
// @Param   Email			query		string		true		"Email"
// @Param   Password	query		string		true		"Password"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /auth/signup [post]
func (h *ProfileHandler) AuthSignup(c *gin.Context) {
	db := db.GetDB()

	var err error

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    auth.ACTION_SIGNUP,
		Method:    auth.METHOD_EMAIL,
	}

	var query struct {
		Email    string `json:"Email" binding:"required"`
		Password string `json:"Password" binding:"required"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if utils.Error(c, identity.ParseEmail(query.Email)) {
		return
	}

	if utils.Error(c, identity.CheckPassword(query.Password, true)) {
		return
	}

	err = identity.CheckEmailInUse()
	if utils.Error(c, err) {
		return
	}

	identity.PasswordHash, err = identity.HashBcrypt(identity.Password)
	if utils.Error(c, err) {
		return
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Member (Email, PasswordHash, IP, Created) VALUES ($1, $2, $3, $4)", identity.Email, identity.PasswordHash, c.ClientIP(), identity.Timestamp)
	tx.Commit()

	//Get new member record
	if utils.Error(c, identity.QueryMemberByEmail()) {
		return
	}

	if utils.Error(c, identity.GenerateOTP()) {
		return
	}

	identity.CodeHash, err = identity.HashBcrypt(identity.Code)
	if utils.Error(c, err) {
		return
	}

	identity.EmailHash, err = identity.HashMD5(identity.Email)
	if utils.Error(c, err) {
		return
	}

	tx = db.MustBegin()
	// TODO: tx.MustExec("INSERT INTO Onboarding (MemberID, Email, Contract, Phone, Password) VALUES ($1, $2, $3, $4, $5)", identity.Member.MemberID, false, false, false, true)
	tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Email, EmailHash, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", identity.Member.MemberID, identity.Action, identity.Method, identity.CodeHash, identity.Email, identity.EmailHash, auth.STATUS_PENDING, c.ClientIP(), identity.Timestamp)
	tx.Commit()

	if utils.Error(c, identity.GenerateConfirmationLink()) {
		return
	}

	mail.SendMail(identity.Email, "Account activation", "Dear customer,", "Please activate your account by clicking on the button below.", "Activate account", identity.ConfirmationLink, "Thank you for joining our platform.")

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// AuthSignup member signup resend confirmation email
// @Summary member signup resend confirmation email
// @Description Signup
// @Tags Auth, Signup
// @Accept  json
// @Produce  json
// @ID Auth-Signup-Resend
// @Param   Email			query		string		true		"Email"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /auth/signup/resend [post]
func (h *ProfileHandler) AuthSignupResend(c *gin.Context) {
	db := db.GetDB()

	var err error

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    auth.ACTION_SIGNUP,
		Method:    auth.METHOD_EMAIL,
	}

	var query struct {
		Email string `json:"Email" binding:"required"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if utils.Error(c, identity.ParseEmail(query.Email)) {
		return
	}

	if utils.Error(c, identity.QueryMemberByEmail()) {
		return
	}

	timeout, err := identity.QueryVerifyTimeout()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   err.Error(),
			"timeout": timeout,
		})
		return
	}

	if utils.Error(c, identity.QueryOnboardingRecord(identity.Member.MemberID)) {
		return
	}

	//Check it's a new member
	if identity.Onboarding.Complete == false {
		utils.Error(c, errors.New("ALREADY_VERIFIED"))
		return
	}

	if utils.Error(c, identity.GenerateOTP()) {
		return
	}

	identity.CodeHash, err = identity.HashBcrypt(identity.Code)
	if utils.Error(c, err) {
		return
	}

	identity.EmailHash, err = identity.HashMD5(identity.Email)
	if utils.Error(c, err) {
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Action=$3 AND Method=$4 AND Status=$5", auth.STATUS_CANCELLED, identity.Member.MemberID, identity.Action, identity.Method, auth.STATUS_PENDING)
	tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Email, EmailHash, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", identity.Member.MemberID, identity.Action, identity.Method, identity.CodeHash, identity.Email, identity.EmailHash, auth.STATUS_PENDING, c.ClientIP(), identity.Timestamp)
	tx.Commit()

	if utils.Error(c, identity.GenerateConfirmationLink()) {
		return
	}

	mail.SendMail(identity.Email, "Account activation", "Dear customer,", "Please activate your account by clicking on the button below.", "Activate account", identity.ConfirmationLink, "Thank you for joining our platform.")

	c.JSON(http.StatusOK, gin.H{"status": true})
}

// AuthPhoneRequest request OTP SMS code for 2fa auth
// @Summary request OTP SMS code for auth
// @Description PhoneRequestCode
// @Tags Public, Auth
// @Accept  json
// @Produce  json
// @ID Phone-Auth
// @Param   Phone		query		string		true		"Phone number"
// @Success 200 {object} Settings
// @Failure 400 {object} Error
// @Router /auth/phone/request [post]
func (h *ProfileHandler) AuthPhoneRequest(c *gin.Context) {
	db := db.GetDB()

	var err error

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    auth.ACTION_SIGNIN,
		Method:    auth.METHOD_PHONE,
	}

	var query struct {
		Phone string `json:"Phone" binding:"required"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if utils.Error(c, identity.ParsePhone(query.Phone)) {
		return
	}

	if utils.Error(c, identity.QueryMemberByPhone()) {
		return
	}

	timeout, err := identity.QueryVerifyTimeout()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   err.Error(),
			"timeout": timeout,
		})
		return
	}

	if utils.Error(c, identity.GenerateOTP()) {
		return
	}

	identity.CodeHash, err = identity.HashBcrypt(identity.Code)
	if utils.Error(c, err) {
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Action=$3 AND Method=$4 AND Status=$5", auth.STATUS_CANCELLED, identity.Member.MemberID, identity.Action, identity.Method, auth.STATUS_PENDING)
	tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Phone, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", identity.Member.MemberID, identity.Action, identity.Method, identity.CodeHash, identity.Phone, auth.STATUS_PENDING, c.ClientIP(), identity.Timestamp)
	tx.Commit()

	err = otp.SendSMS(identity.Phone, fmt.Sprintf("Your code is %s", identity.Code))
	if utils.Error(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"status":  true,
		"timeout": identity.Timeout,
	})
}

// AuthReset request password reset
// @Summary request password reset
// @Description AuthReset
// @Tags Public, Auth
// @Accept  json
// @Produce  json
// @ID Auth-Reset
// @Param   Email		query		string		true		"Email"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /auth/reset [post]
func (h *ProfileHandler) AuthReset(c *gin.Context) {
	db := db.GetDB()

	var err error

	identity := auth.Identity{
		Timestamp: time.Now().Unix(),
		Timeout:   60,
		Action:    auth.ACTION_RESET,
		Method:    auth.METHOD_EMAIL,
	}

	var query struct {
		Email string `json:"Email" binding:"required"`
	}

	if utils.Error(c, utils.ShouldBindJSON(c, &query)) {
		return
	}

	if utils.Error(c, identity.ParseEmail(query.Email)) {
		return
	}

	if utils.Error(c, identity.QueryMemberByEmail()) {
		return
	}

	//Suggest resending email confirmation link if member has not yet set a password
	if utils.Error(c, identity.QueryOnboardingRecord(identity.Member.MemberID)) {
		return
	}

	if identity.Onboarding.Password == false {
		utils.Error(c, errors.New("RESEND_EMAIL_CONFIRMATION"))
		return
	}

	timeout, err := identity.QueryVerifyTimeout()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   err.Error(),
			"timeout": timeout,
		})
		return
	}

	if utils.Error(c, identity.GenerateOTP()) {
		return
	}

	identity.CodeHash, err = identity.HashBcrypt(identity.Code)
	if utils.Error(c, err) {
		return
	}

	identity.EmailHash, err = identity.HashMD5(identity.Email)
	if utils.Error(c, err) {
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE Verify SET Status=$1 WHERE MemberID=$2 AND Action=$3 AND Method=$4 AND Status=$5", auth.STATUS_CANCELLED, identity.Member.MemberID, auth.ACTION_RESET, auth.METHOD_EMAIL, auth.STATUS_PENDING)
	tx.MustExec("INSERT INTO Verify (MemberID, Action, Method, CodeHash, Email, EmailHash, Status, IP, Created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", identity.Member.MemberID, auth.ACTION_RESET, auth.METHOD_EMAIL, identity.CodeHash, identity.Email, identity.EmailHash, auth.STATUS_PENDING, c.ClientIP(), identity.Timestamp)
	tx.Commit()

	if utils.Error(c, identity.GenerateConfirmationLink()) {
		return
	}

	mail.SendMail(identity.Email, "Password reset", "Dear customer,", "Please reset your password by clicking on the button below.", "Reset password", identity.ConfirmationLink, "If you have received a password reset email without requesting one, we recommend that you take steps to secure your account.")

	c.JSON(http.StatusOK, gin.H{"status": true})
}
