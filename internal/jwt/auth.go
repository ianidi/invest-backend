package jwt

import (
	"github.com/ianidi/exchange-server/internal/redis"
	"github.com/mediocregopher/radix/v3"
)

const (
	// MemberKey is used to set a user into gin.Context.
	MemberKey              = "MemberID"
	AuthorizationHeaderKey = "Authorization"
	RefreshHeaderKey       = "Refresh"
)

var Service *ProfileHandler

type AuthInterface interface {
	// CreateAuth(string, *TokenDetails) error
	FetchAuth(string) (string, error)
	DeleteRefresh(string) error
	// DeleteTokens(*AccessDetails) error
}

type service struct {
	client *radix.Pool
}

func Init() *ProfileHandler {

	var rd = NewAuth()
	var tk = NewToken()
	Service = NewProfile(rd, tk)

	return Service
}

var _ AuthInterface = &service{}

func NewAuth() *service {
	client := redis.GetRedis()

	return &service{client: client}
}

type AccessDetails struct {
	// TokenUuid string
	UserId string
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	// TokenUuid    string
	RefreshUuid string
	AtExpires   int64
	RtExpires   int64
}

//Save token metadata to Redis
// func (tk *service) CreateAuth(userId string, td *TokenDetails) error {
// 	// at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
// 	// rt := time.Unix(td.RtExpires, 0)
// 	// now := time.Now()

// 	var err error

// 	now := time.Now().Unix()

// 	var atCreated string
// 	err = tk.client.Do(radix.Cmd(&atCreated, "SET", td.TokenUuid, userId, "EX", cast.ToString(td.AtExpires-now)))
// 	if err != nil {
// 		return err
// 	}

// 	var rtCreated string
// 	err = tk.client.Do(radix.Cmd(&rtCreated, "SET", td.RefreshUuid, userId, "EX", cast.ToString(td.RtExpires-now)))
// 	if err != nil {
// 		return err
// 	}

// 	if atCreated == "0" || rtCreated == "0" {
// 		return errors.New("no record inserted")
// 	}
// 	return nil
// }

//Check the metadata saved
func (tk *service) FetchAuth(tokenUuid string) (string, error) {
	var MemberID string
	err := tk.client.Do(radix.Cmd(&MemberID, "GET", tokenUuid))
	if err != nil {
		return "", err
	}

	return MemberID, nil
}

//Once a user row in the token table
// func (tk *service) DeleteTokens(authD *AccessDetails) error {
// 	//get the refresh uuid
// 	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, authD.UserId)
// 	//delete access token
// 	var deletedAt int64
// 	err := tk.client.Do(radix.Cmd(&deletedAt, "DEL", authD.TokenUuid))
// 	if err != nil {
// 		return err
// 	}

// 	//delete refresh token
// 	var deletedRt int64
// 	err = tk.client.Do(radix.Cmd(&deletedRt, "DEL", refreshUuid))
// 	if err != nil {
// 		return err
// 	}

// 	//When the record is deleted, the return value is 1
// 	if deletedAt != 1 || deletedRt != 1 {
// 		return errors.New("something went wrong")
// 	}
// 	return nil
// }

func (tk *service) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	var deleted int64
	err := tk.client.Do(radix.Cmd(&deleted, "DEL", refreshUuid))
	if err != nil || deleted == 0 {
		return err
	}

	return nil
}
