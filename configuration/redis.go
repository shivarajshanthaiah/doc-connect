package configuration

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// Client variable can used to save key value pairs in redis
var Client *redis.Client

// InitRedis function initializes redis server
func InitRedis() {
	var err error
	MaxRetries := 5
	RetryDelay := time.Second * 5
	for i := 0; i < MaxRetries; i++ {
		Client = redis.NewClient(&redis.Options{
			Network:  "tcp",
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		_, err = Client.Ping(ctx).Result()
		if err == nil {
			break
		}

		fmt.Printf("Failed to connect to Redis (Attempt %d/%d): %s\n", i+1, MaxRetries, err.Error())
		time.Sleep(RetryDelay)
	}
	if err != nil {
		panic("Failed to connect to Redis after multiple attempts: " + err.Error())
	}
}

// SetRedis willset a key value in redis server
func SetRedis(key string, value any, expirationTime time.Duration) error {
	if err := Client.Set(context.Background(), key, value, expirationTime).Err(); err != nil {
		return err
	}
	return nil
}

// GetRedis will get the value from redis server using key
func GetRedis(key string) (string, error) {
	jsonData, err := Client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return jsonData, nil
}
