package redis

import (
	"fmt"
	"time"

	"github.com/mediocregopher/radix/v3"
	"github.com/spf13/viper"
)

var redisClient *radix.Pool
var psclient radix.PubSubConn

func GetRedis() *radix.Pool {
	return redisClient
}

func InitRedis() {
	// init redis connection pool
	initPool()

	//init info channel subscription
	initSubscription()
}

func initSubscription() {
	var err error

	customConnFunc := radix.PersistentPubSubConnFunc(func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(10*time.Second),
			radix.DialAuthPass(viper.GetString("redis_password")),
		)
	})

	psclient, err = radix.PersistentPubSubWithOpts("tcp", viper.GetString("redis_host"), customConnFunc)
	if err != nil {
		fmt.Println("radix subscription error", err)
	}
}

func initPool() {
	var err error

	DefaultConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(10*time.Second),
			radix.DialAuthPass(viper.GetString("redis_password")),
		)
	}

	redisClient, err = radix.NewPool("tcp", viper.GetString("redis_host"), 10, radix.PoolConnFunc(DefaultConnFunc))
	if err != nil {
		fmt.Println("radix error", err)
	}
}

type Channel struct {
	Name    string
	Message string
	MsgCh   chan radix.PubSubMessage
}

func (channel Channel) GetInfoChannel() (chan radix.PubSubMessage, error) {

	// Subscribe to a channel called "myChannel". All publishes to "myChannel"
	// will get sent to msgCh after this
	infoChannel := make(chan radix.PubSubMessage)
	if err := psclient.Subscribe(infoChannel, "info"); err != nil {
		return infoChannel, err
	}

	return infoChannel, nil
}

func (channel Channel) PubToChannel() error {
	var err error

	// This example retrieves the current integer value of `key` and sets its
	// new value to be the increment of that, all using the same connection
	// instance. NOTE that it does not do this atomically like the INCR command
	// would.
	key := ""
	err = redisClient.Do(radix.WithConn(key, func(conn radix.Conn) error {
		return conn.Do(radix.Cmd(nil, "PUBLISH", channel.Name, channel.Message))
	}))
	if err != nil {
		return err
	}
	return nil
}
