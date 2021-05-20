package splash

import (
	"context"
	"errors"

	"github.com/amupxm/go-video-concat/models"
	"gorm.io/gorm"

	postgres "github.com/amupxm/go-video-concat/packages/database"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/minio/minio-go/v7"
)

type (
	Splash               struct{ *models.Splash }
	SplashMessageChannel struct {
		Message string
		Status  bool
	}
	SplashOperation interface {
		AddSplash() bool
	}
)

func (splash *Splash) AddSplash(uuid string) (bool, string) {
	var Database = &postgres.PostgresConnection
	var ObjectStorage = &s3.ObjectStorage
	if splash.BaseFile == "" || splash.BaseAudio == "" {
		return false, "object name should not be empty"
	}
	_, err := ObjectStorage.Client.StatObject(context.Background(), "splash-base", splash.BaseFile, minio.StatObjectOptions{})
	if err != nil {
		return false, "base file dont exits in storage"
	}
	_, err = ObjectStorage.Client.StatObject(context.Background(), "splash-audio", splash.BaseAudio, minio.StatObjectOptions{})
	if err != nil {
		return false, "audio file dont exits in storage"
	}
	splash.FileCode = uuid
	splash.Status = false
	result := Database.DBCli.Create(splash)
	return result.Error == nil, "done"
}

func (splash *Splash) FindSplash(code string) (Splash, bool) {
	var resultFrame Splash
	var Database = &postgres.PostgresConnection
	result := Database.DBCli.First(&resultFrame, "file_code = ?", code)
	return resultFrame, !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func (splash *Splash) GetFile(splashCode string) (*minio.Object, bool) {
	var tmp *minio.Object
	_, status := splash.FindSplash(splashCode)
	if !status {
		return tmp, status
	}

	var ObjectStorage = &s3.ObjectStorage
	reader, err := ObjectStorage.Client.GetObject(context.Background(), "splash", splashCode, minio.GetObjectOptions{})
	if err != nil {
		return tmp, false
	}
	return reader, true
}
