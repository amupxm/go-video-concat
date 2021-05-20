package controller

import (
	AppContext "context"
	"fmt"
	"log"
	"net/http"

	"github.com/amupxm/go-video-concat/models"
	"github.com/amupxm/go-video-concat/packages/cache"
	"github.com/amupxm/go-video-concat/packages/ffmpeg"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/minio/minio-go/v7"
)

func Generator_Upload(context *gin.Context) {
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
		"upload",
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

func Generator_Generate(context *gin.Context) {
	var recipe models.VideoRecipe
	if err := context.ShouldBindJSON(&recipe); err != nil {
		context.JSON(http.StatusNotAcceptable, gin.H{
			"ok":      false,
			"message": "invalid structure",
		})
		fmt.Print(err)
		return
	}
	var f ffmpeg.FFmpeg_Generator
	f.Recipe = &recipe
	f.Error = &ffmpeg.FFmpeg_Message{}
	processId, _ := uuid.NewV4()
	f.UUID = processId.String()
	context.JSON(http.StatusNotAcceptable, gin.H{
		"ok":      true,
		"message": "please wait",
		"code":    f.UUID,
	})
	go func() {
		ffmpeg.Generator(&f)

	}()
}

func Generator_Status(context *gin.Context) {
	code := context.Param("code")
	status, message := cache.GetStatus(code)
	context.JSON(200, gin.H{
		"status":  status,
		"message": message,
	})
}
