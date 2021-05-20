package frame

import (
	"context"
	"errors"
	"log"

	"github.com/amupxm/go-video-concat/models"
	postgres "github.com/amupxm/go-video-concat/packages/database"
	"github.com/amupxm/go-video-concat/packages/s3"
	"github.com/minio/minio-go/v7"

	"gorm.io/gorm"
)

type (
	Frame          struct{ *models.Frame }
	FrameOperation interface {
		AddFrame() bool
		ListFrame() []Frame
		FindFrame(code string) (Frame, bool)
		GetFile(frameCode string) (Frame, bool)
	}
)

func (frame *Frame) AddFrame() bool {
	var Database = &postgres.PostgresConnection
	var ObjectStorage = &s3.ObjectStorage
	// check file code exists in storage
	_, err := ObjectStorage.Client.StatObject(context.Background(), "frame", frame.FileCode, minio.StatObjectOptions{})
	if err != nil {
		log.Fatal(err)
		return false

	}
	result := Database.DBCli.Create(frame)
	return result.Error == nil
}

func (frame *Frame) ListFrame() []Frame {
	var Database = &postgres.PostgresConnection
	var respone []Frame
	Database.DBCli.Find(&respone)
	return respone
}

func (frame *Frame) FindFrame(code string) (Frame, bool) {
	var resultFrame Frame
	var Database = &postgres.PostgresConnection
	result := Database.DBCli.First(&resultFrame, "file_code = ?", code)
	return resultFrame, !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func (frame *Frame) GetFile(frameCode string) (*minio.Object, bool) {
	var tmp *minio.Object
	_, status := frame.FindFrame(frameCode)
	if !status {
		return tmp, status
	}

	var ObjectStorage = &s3.ObjectStorage
	reader, err := ObjectStorage.Client.GetObject(context.Background(), "frame", frameCode, minio.GetObjectOptions{})
	if err != nil {
		return tmp, false
	}
	return reader, true
}
