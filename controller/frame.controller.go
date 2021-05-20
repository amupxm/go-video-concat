package controller

import (
	AppContext "context"
	"fmt"

	"log"
	"net/http"

	frame "github.com/amupxm/go-video-concat/interfaces/frame"
	"github.com/amupxm/go-video-concat/packages/s3"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/minio/minio-go/v7"
)

// upload add

func Frame_Upload(context *gin.Context) {
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
		"frame",
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

// add
func Frame_Add(context *gin.Context) {
	var frame frame.Frame
	if err := context.ShouldBindJSON(&frame); err != nil {
		context.JSON(http.StatusNotAcceptable, gin.H{
			"ok":      false,
			"message": "invalid json structure",
		})
	}
	status := frame.AddFrame()

	context.JSON(http.StatusAccepted, gin.H{
		"ok":   status,
		"data": frame,
	})
}

// get all

func Frame_list(context *gin.Context) {
	var frame frame.Frame
	frameList := frame.ListFrame()
	context.JSON(http.StatusAccepted, gin.H{
		"ok":   true,
		"data": frameList,
	})
}

// TODO single view
func Frame_Single(context *gin.Context) {
	frameCode := context.Param("code")
	var frame frame.Frame
	frameList, status := frame.FindFrame(frameCode)
	context.JSON(http.StatusAccepted, gin.H{
		"ok":   status,
		"data": frameList,
	})
}

// TODO single file view
func Frame_File(context *gin.Context) {
	frameCode := context.Param("code")
	var frame frame.Frame
	file, status := frame.GetFile(frameCode)
	stat, err := file.Stat()
	if err != nil || !status {
		context.JSON(http.StatusAccepted, gin.H{
			"ok": false,
		})
	}
	extraHeaders := map[string]string{}
	fmt.Print(stat)
	context.DataFromReader(http.StatusOK, stat.Size, stat.ContentType, file, extraHeaders)
}

// TODO remove
