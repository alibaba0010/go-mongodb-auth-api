package database

import (
	"context"

	"gin-mongo-aws/internal/logger"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(addr, password string, db int) error {
       RedisClient = redis.NewClient(&redis.Options{
	       Addr:     addr,
	       Password: password,
	       DB:       db,
       })

       _, err := RedisClient.Ping(context.Background()).Result()
       if err != nil {
	       return err
       }

       logger.Log.Info("Connected to Redis")
       return nil
}

func CloseRedis() {
       if RedisClient != nil {
	       if err := RedisClient.Close(); err != nil {
		       logger.Log.Error("Error closing Redis connection")
	       } else {
		       logger.Log.Info("Closed Redis connection")
	       }
       }
}
