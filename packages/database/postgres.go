package postgres

import (
	"fmt"
	"log"

	"github.com/amupxm/go-video-concat/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	DBCli *gorm.DB
}

var PostgresConnection = Postgres{}

func (d *Postgres) ConnectDatabase(config *config.ConfigStruct) {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v",
		config.Database.Postgres_host,
		config.Database.Postgres_username,
		config.Database.Postgres_password,
		config.Database.Postgres_databasename,
		config.Database.Postgres_port,
		"disable",
		"Asia/Tehran",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	d.DBCli = db
}
