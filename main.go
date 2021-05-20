package main

import (
	"github.com/amupxm/go-video-concat/packages/cache"
	postgres "github.com/amupxm/go-video-concat/packages/database"

	ApiController "github.com/amupxm/go-video-concat/controller"

	s3 "github.com/amupxm/go-video-concat/packages/s3"

	"github.com/amupxm/go-video-concat/config"
	"github.com/gin-gonic/gin"
)

var AppConfig = &config.ConfigStruct{}

func main() {
	router := gin.Default()
	AppConfig.LoadConfigs()

	// =======  Database and Storage =========
	postgres.PostgresConnection.ConnectDatabase(AppConfig)
	postgres.AutoMigration()
	buckets := []string{"frame", "thumbnails", "splash", "upload", "splash-base", "splash-audio", "outputs"}
	s3.ObjectStorage.Connect(AppConfig)
	s3.InitBuckets(buckets)
	cache.Init(AppConfig)
	// =======================================

	// ============  Controllers  ============
	// 1 - frame
	router.POST("/frame/upload", ApiController.Frame_Upload)
	router.POST("/frame", ApiController.Frame_Add)
	router.GET("/frame", ApiController.Frame_list)
	router.GET("/frame/:code/file", ApiController.Frame_File)
	router.GET("/frame/:code", ApiController.Frame_Single)

	// 2 - splash video (splash)
	router.POST("/splash/base", ApiController.Splash_Base)
	router.POST("/splash/audio", ApiController.Splash_Audio)
	router.POST("/splash", ApiController.Splash_Add)
	router.GET("/splash/:code", ApiController.Splash_file)

	// 3 - generator
	router.POST("/generator/upload", ApiController.Generator_Upload)
	router.POST("/generator", ApiController.Generator_Generate)
	router.GET("/generator/:code", ApiController.Generator_Status)
	router.Run()
}
