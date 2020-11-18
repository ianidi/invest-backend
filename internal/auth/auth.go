package auth

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"hash"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/muesli/crunchy"
	"github.com/nyaruka/phonenumbers"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cast"
	"golang.org/x/crypto/bcrypt"
)

const (
	METHOD_PHONE     = "phone"
	METHOD_EMAIL     = "email"
	ACTION_SIGNIN    = "signin"
	ACTION_SIGNUP    = "signup"
	ACTION_CHANGE    = "change"
	ACTION_RESET     = "reset"
	STATUS_PENDING   = "pending"
	STATUS_CANCELLED = "cancelled"
	STATUS_FAIL      = "fail"
	STATUS_SUCCESS   = "success"
)

type Identity struct {
	Member           models.Member
	Verify           models.Verify
	Onboarding       models.Onboarding
	Action           string
	Method           string
	Phone            string //Phone number
	Email            string //Email address
	EmailHash        string //Email address md5 hash
	Password         string //Current (auth) or new (reset) member password
	PasswordHash     string //Password hash
	Code             string //OTP code
	CodeHash         string //OTP code hash
	ConfirmationLink string //Email confirmation link
	Timestamp        int64  //Current UNIX timestamp
	Timeout          int64  //How frequently in seconds OTP codes can be requested
}

//Generate confirmation code
func (identity *Identity) GenerateOTP() error {
	code, err := password.Generate(6, 6, 0, true, false)
	if err != nil {
		return err
	}

	identity.Code = code

	return nil
}

//Hash using bcrypt
func (identity Identity) HashBcrypt(str string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), 12)
	if err != nil {
		return "", err
	}

	return cast.ToString(hash), nil
}

//Hash using md5
func (identity Identity) HashMD5(str string) (string, error) {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:]), nil
}

//Parse phone number
func (identity *Identity) ParsePhone(phone string) error {

	// phone = strings.Replace(phone, "+", "", -1)

	phoneNumber, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return err
	}

	if phonenumbers.IsValidNumber(phoneNumber) == false {
		return errors.New("INVALID_PHONE")
	}

	identity.Phone = cast.ToString(phoneNumber.GetCountryCode()) + cast.ToString(phoneNumber.GetNationalNumber())

	return nil
}

//Parse email
func (identity *Identity) ParseEmail(email string) error {

	if err := validation.Validate(email,
		validation.Length(0, 250),
		validation.Required,
		is.Email,
	); err != nil {
		return err
	}

	identity.Email = strings.ToLower(email)

	return nil
}

//Query member by email
func (identity *Identity) QueryMemberByEmail() error {
	db := db.GetDB()

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE Email=$1", identity.Email); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_MEMBER_RECORD")
		}
		return err
	}

	identity.Member = member

	return nil
}

//Query member by phone
func (identity *Identity) QueryMemberByPhone() error {
	db := db.GetDB()

	var member models.Member

	if err := db.Get(&member, "SELECT * FROM Member WHERE Phone=$1", identity.Email); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("NO_MEMBER_RECORD")
		}
		return err
	}

	identity.Member = member

	return nil
}

//Query member count by phone
func (identity Identity) CheckPhoneInUse() error {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Phone=$1", identity.Phone); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count > 0 {
		return errors.New("PHONE_ALREADY_IN_USE")
	}

	return nil
}

//Query member count by email
func (identity Identity) CheckEmailInUse() error {
	db := db.GetDB()

	var count int

	if err := db.Get(&count, "SELECT count(*) FROM Member WHERE Email=$1", identity.Email); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if count > 0 {
		return errors.New("EMAIL_ALREADY_IN_USE")
	}

	return nil
}

//Find verification code sent in the last timeout value, in case found return time left until new code request
func (identity Identity) QueryVerifyTimeout() (int64, error) {
	db := db.GetDB()

	var verify models.Verify

	if err := db.Get(&verify, "SELECT * FROM Verify WHERE MemberID=$1 AND Action=$2 AND Method=$3 AND Status=$4", identity.Member.MemberID, identity.Action, identity.Method, STATUS_PENDING); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	//Not enough time passed after the last code request
	if verify.Created > identity.Timestamp-identity.Timeout {
		return identity.Timeout - identity.Timestamp + verify.Created, errors.New("CODE_TIMEOUT")
	}

	return 0, nil

}

//Validate password
func (identity *Identity) CheckPassword(password string, validateStrength bool) error {

	if err := validation.Validate(password,
		validation.Required,
	); err != nil {
		return errors.New("INVALID_PASSWORD")
	}

	//TODO: remove validateStrength
	if validateStrength {
		validator := crunchy.NewValidatorWithOpts(crunchy.Options{
			// MinLength is the minimum length required for a valid password
			// (must be >= 1, default is 8)
			MinLength: 6,

			// MinDiff is the minimum amount of unique characters required for a valid password
			// (must be >= 1, default is 5)
			MinDiff: 5,

			// Hashers will be used to find hashed passwords in dictionaries
			Hashers: []hash.Hash{md5.New(), sha1.New(), sha256.New(), sha512.New()},

			// Check haveibeenpwned.com database
			CheckHIBP: true,
		})

		if err := validator.Check(password); err != nil {
			return errors.New("WEAK_PASSWORD")
		}
	}

	identity.Password = password

	return nil
}

//Query verify by md5 hash (email)
func (identity *Identity) QueryVerifyByHash(hash string) error {
	db := db.GetDB()

	var verify models.Verify

	//Find verify request
	if err := db.Get(&verify, "SELECT * FROM Verify WHERE EmailHash=$1 AND Action=$2 AND Method=$3 AND Status=$4", hash, identity.Action, identity.Method, STATUS_PENDING); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("INVALID_LINK")
		}
		return err
	}

	//Request is expired (day)
	if verify.Created < identity.Timestamp-86400 {
		return errors.New("EXPIRED_LINK")
	}

	identity.Verify = verify
	return nil
}

//Validate OTP code
func (identity Identity) ValidateOTPCode(code string) error {
	db := db.GetDB()

	if err := bcrypt.CompareHashAndPassword([]byte(identity.Verify.CodeHash), []byte(code)); err != nil {

		tx := db.MustBegin()
		tx.MustExec("UPDATE Verify SET Attempts=Attempts+$1 WHERE VerifyID=$2", 1, identity.Verify.VerifyID)
		tx.Commit()

		//Verify attempts exceeded
		if identity.Verify.Attempts >= 3 {

			tx := db.MustBegin()
			tx.MustExec("UPDATE Verify SET Status=$1 WHERE VerifyID=$2", STATUS_FAIL, identity.Verify.VerifyID)
			tx.Commit()

			return errors.New("TOO_MANY_ATTEMPTS")
		}

		return errors.New("INVALID_CODE")
	}

	return nil
}

//Validate password
func (identity Identity) ValidatePassword() error {

	if err := bcrypt.CompareHashAndPassword([]byte(identity.Member.PasswordHash), []byte(identity.Password)); err != nil {

		//TODO: bruteforce protection

		return errors.New("INVALID_PASSWORD")
	}

	return nil
}

//Query setttings
func (identity Identity) QuerySettings() (models.Settings, error) {
	db := db.GetDB()

	var settings models.Settings

	err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)
	if err != nil {
		return settings, err
	}

	return settings, nil
}

//Query onboarding
func (identity *Identity) QueryOnboardingRecord(MemberID int64) error {
	db := db.GetDB()

	var onboarding models.Onboarding

	if err := db.Get(&onboarding, "SELECT * FROM Onboarding WHERE MemberID=$1", MemberID); err != nil {
		return err
	}

	identity.Onboarding = onboarding

	return nil
}

//Generate email confirmation link
func (identity *Identity) GenerateConfirmationLink() error {

	settings, err := identity.QuerySettings()
	if err != nil {
		return err
	}

	identity.ConfirmationLink = settings.PlatformURL + "verify/" + identity.Action + "/" + identity.EmailHash + "/" + identity.Code

	return nil

}
