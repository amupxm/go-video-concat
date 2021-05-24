package cache

import (
	"context"
	"log"

	"github.com/amupxm/go-video-concat/config"
	"github.com/go-redis/redis/v8"
)

type CacheRedisContext struct {
	Cli *redis.Client
}

var ctx = context.Background()
var CacheRedis = CacheRedisContext{}

func Init(config *config.ConfigStruct) {
	db := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       int(config.Redis.DB),
	})
	CacheRedis.Cli = db
}

func NewProccess(uuid string, sectionsCount int) bool {
	var r = &CacheRedis
	errors := make([]error, 4)
	errors[1] = r.Cli.Set(ctx, uuid+":status", "started", 0).Err()
	for _, e := range errors {
		if e != nil {
			return false
		}
	}
	return true
}

func UpdateStatus(uuid, message string, status bool) bool {
	var r = &CacheRedis
	errors := make([]error, 2)
	errors[1] = r.Cli.Set(ctx, uuid+":status", status, 0).Err()
	errors[0] = r.Cli.Set(ctx, uuid+":message", message, 0).Err()

	for _, e := range errors {
		if e == nil {
			return false
		}
	}
	return true
}

func GetStatus(code string) (bool, string) {
	var r = &CacheRedis
	log.Println(code + ":status")
	respStatus := r.Cli.Get(ctx, code+":status")
	respMessage := r.Cli.Get(ctx, code+":message")
	message := respMessage.Val()
	status, err := respStatus.Bool()
	if err == nil {
		return status, message
	}
	return false, "invalid code"
}
