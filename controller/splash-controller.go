package controller

import (
	AppContext "context"
	"fmt"
	"log"
	"net/http"

	"github.com/amupxm/go-video-concat/interfaces/splash"
	"github.com/amupxm/go-video-concat/packages/ffmpeg"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/minio/minio-go/v7"
)

func Splash_Audio(context *gin.Context) {
	file, err := context.FormFile("upload")
	// if no file contained
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "invalid file input",
		})
		return
	}

	uuid, _ := uuid.NewV4()

	fileIOReader, err := file.Open()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "file error",
		})
	}
	var storage = &s3.ObjectStorage

	uploadInformation, err := storage.Client.PutObject(
		AppContext.Background(),
		"splash-audio",
		uuid.String(),
		fileIOReader,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": err,
		})
	}
	log.Print(uploadInformation.LastModified.Second())
	context.JSON(200, gin.H{
		"ok":        true,
		"message":   "file uploaded successfully",
		"file_name": uuid.String(),
	})
}

func Splash_Base(context *gin.Context) {
	file, err := context.FormFile("upload")
	// if no file contained
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "invalid file input",
		})
		return
	}

	uuid, _ := uuid.NewV4()

	fileIOReader, err := file.Open()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "file error",
		})
	}
	var storage = &s3.ObjectStorage

	uploadInformation, err := storage.Client.PutObject(
		AppContext.Background(),
		"splash-base",
		uuid.String(),
		fileIOReader,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": err,
		})
	}
	log.Print(uploadInformation.LastModified.Second())
	context.JSON(200, gin.H{
		"ok":        true,
		"message":   "file uploaded successfully",
		"file_name": uuid.String(),
	})
}

func Splash_Add(context *gin.Context) {
	var Splash splash.Splash

	if err := context.ShouldBindJSON(&Splash); err != nil {
		context.JSON(http.StatusNotAcceptable, gin.H{
			"ok":      false,
			"message": "invalid requires structure",
		})
		return
	}

	uuid, _ := uuid.NewV4()

	status, message := Splash.AddSplash(uuid.String())
	context.JSON(http.StatusAccepted, gin.H{
		"ok":      status,
		"message": message,
		"code":    uuid.String(),
	})
	// now should start proccessing
	res := ffmpeg.GenerateSplash(&Splash)
	fmt.Print(res)

}

func Splash_file(context *gin.Context) {
	splashCode := context.Param("code")
	var splash splash.Splash
	file, status := splash.GetFile(splashCode)
	stat, err := file.Stat()
	if err != nil || !status {
		context.JSON(http.StatusAccepted, gin.H{
			"ok": false,
		})
	}
	extraHeaders := map[string]string{}
	context.DataFromReader(http.StatusOK, stat.Size, stat.ContentType, file, extraHeaders)
}
