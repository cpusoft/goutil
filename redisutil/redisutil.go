package redisutil
import (

	"context"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/go-redis/redis"
)

//func redisOptions() *redis.Options {
//	return &redis.Options{
//		Addr: redisAddr,
//		DB:   15,
//
//		DialTimeout:  10 * time.Second,
//		ReadTimeout:  30 * time.Second,
//		WriteTimeout: 30 * time.Second,
//
//		MaxRetries: -1,
//
//		PoolSize:           10,
//		PoolTimeout:        30 * time.Second,
//		IdleTimeout:        time.Minute,
//		IdleCheckFrequency: 100 * time.Millisecond,
//	}
//}


//
//
//rdb := redis.NewClient(&redis.Options{
//Addr:     "localhost:6379", // use default Addr
//Password: "",               // no password set
//DB:       0,                // use default DB
//})
//
//pong, err := rdb.Ping(ctx).Result()
//fmt.Println(pong, err)
//// Output: PONG <nil>


var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func InitRedis() (err error) {
	addr := conf.String("redis::addr")
	password := conf.String("redis::password")
	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr, // use default Addr
		Password: password,               // no password set
	})
	return nil
}

