package postgres

import (
	"log"

	"github.com/amupxm/go-video-concat/models"
)

func AutoMigration() {
	if err := PostgresConnection.DBCli.AutoMigrate(
		&models.Frame{},
		&models.Splash{},
	); err != nil {
		log.Panicln(err)
	}
}
