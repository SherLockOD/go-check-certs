package config

import (
	"go.uber.org/zap"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
)

var (
	// Logger : for zap.Logger
	Logger *zap.Logger
	err    error

	// MessageAppID : message app id
	MessageAppID = ""
	// MessageAppKey : message app key
	MessageAppKey = ""

	// MongoAddr : mongo addr
	MongoAddr = ""
	// MongoDatabase : mongo database
	MongoDatabase = ""
	// MongoUsername : mongo username
	MongoUsername = ""
	// MongoPassword : mongo password
	MongoPassword = ""
	// MongoSession : for mongo session
	MongoSession *mgo.Session
)

func init() {
	zapLog := zap.NewProductionConfig()
	zapLog.DisableStacktrace = true
	Logger, err = zapLog.Build()
	if err != nil {
		panic(err)
	}

	MessageAppID = os.Getenv("MESSAGEAPPID")
	if MessageAppID == "" {
		panic("MESSAGEAPPID is null")
	}
	MessageAppKey = os.Getenv("MESSAGEAPPKEY")
	if MessageAppKey == "" {
		panic("MESSAGEAPPKEY is null")
	}

	MongoAddr = os.Getenv("MONGOADDR")
	if MongoAddr == "" {
		panic("MONGOADDR is null")
	}
	MongoDatabase = os.Getenv("MONGODATABASE")
	if MongoDatabase == "" {
		panic("MONGODATABASE is null")
	}
	MongoUsername = os.Getenv("MONGOUSERNAME")
	if MongoUsername == "" {
		panic("MONGOUSERNAME is null")
	}
	MongoPassword = os.Getenv("MONGOPASSWORD")
	if MongoPassword == "" {
		panic("MONGOPASSWORD is null")
	}
	dailInfo := &mgo.DialInfo{
		Addrs:     strings.Split(MongoAddr, ","),
		Direct:    false,
		Timeout:   time.Second * 1,
		Database:  MongoDatabase,
		Source:    "admin",
		Username:  MongoUsername,
		Password:  MongoPassword,
		PoolLimit: 1024,
	}
	MongoSession, err = mgo.DialWithInfo(dailInfo)
	if err != nil {
		panic(err)
	}

	// mgo.Strong
	// session 的读写一直向主服务器发起并使用一个唯一的连接，因此所有的读写操作完全的一致。
	// mgo.Monotonic
	// session 的读操作开始是向某个 secondary 服务器发起（且通过一个唯一的连接），只要出现了一次写操作，session 的连接就会切换至 primary 服务器。
	// mgo.Eventual
	// session 的读操作会向任意的其他服务器发起，多次读操作并不一定使用相同的连接，也就是读操作不一定有序。session 的写操作总是向主服务器发起，但是可能使用不同的连接，也就是写操作也不一定有序。
	MongoSession.SetMode(mgo.Eventual, true)

	Logger.Info("hello world")
}
