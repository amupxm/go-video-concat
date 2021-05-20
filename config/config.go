package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type (
	ConfigStruct struct {
		Database DatabaseConfig
		Storage  StorageConfig
		Redis    RedisConfig
	}
	DatabaseConfig struct {
		Postgres_host         string
		Postgres_username     string
		Postgres_password     string
		Postgres_databasename string
		Postgres_port         string
	}
	StorageConfig struct {
		Minio_user     string
		Minio_password string
		Minio_host     string
	}

	RedisConfig struct {
		Host     string
		Password string
		DB       uint
	}
)

var config *ConfigStruct

func (cfg *ConfigStruct) LoadConfigs() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg.Database.Postgres_databasename = os.Getenv("dbname")
	cfg.Database.Postgres_host = os.Getenv("dbhost")
	cfg.Database.Postgres_password = os.Getenv("dbpass")
	cfg.Database.Postgres_port = os.Getenv("dbport")
	cfg.Database.Postgres_username = os.Getenv("dbusername")
	cfg.Storage.Minio_host = os.Getenv("Minio_host")
	cfg.Storage.Minio_password = os.Getenv("Minio_password")
	cfg.Storage.Minio_user = os.Getenv("Minio_user")

	cfg.Redis.Host = os.Getenv("Redis_host")
	cfg.Redis.Password = os.Getenv("Redis_password")
	cfg.Redis.DB = 0

	config = cfg
}
