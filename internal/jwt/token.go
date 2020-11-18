package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

type tokenservice struct{}

func NewToken() *tokenservice {
	return &tokenservice{}
}

type TokenInterface interface {
	CreateToken(MemberID string) (*TokenDetails, error)
	ExtractTokenMetadata(*http.Request) (*AccessDetails, error)
	TokenMetadata(string) (*AccessDetails, error)
}

//Token implements the TokenInterface
var _ TokenInterface = &tokenservice{}

func (t *tokenservice) CreateToken(MemberID string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 86400).Unix() //expires after 30 min TODO: 2 months
	// td.TokenUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	// td.RefreshUuid = td.TokenUuid + "++" + MemberID

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	// atClaims["access_uuid"] = td.TokenUuid
	atClaims["user_id"] = MemberID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	// td.RefreshUuid = td.TokenUuid + "++" + MemberID

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = MemberID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	td.RefreshToken, err = rt.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func TokenValid(r *http.Request) error {
	tokenString := ExtractToken(r)

	token, err := verifyToken(tokenString)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func TokenValidNoExtraction(tokenString string) error {
	token, err := verifyToken(tokenString)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("jwt_secret")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

//get the token from the request body
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func extract(token *jwt.Token) (*AccessDetails, error) {

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		// accessUuid, ok := claims["access_uuid"].(string)
		MemberID, userOk := claims["user_id"].(string)
		if userOk == false { //ok == false ||
			return nil, errors.New("unauthorized")
		} else {
			return &AccessDetails{
				// TokenUuid: accessUuid,
				UserId: MemberID,
			}, nil
		}
	}
	return nil, errors.New("something went wrong")
}

func (t *tokenservice) ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	tokenString := ExtractToken(r)

	token, err := verifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	acc, err := extract(token)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (t *tokenservice) TokenMetadata(tokenString string) (*AccessDetails, error) {
	token, err := verifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	acc, err := extract(token)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
