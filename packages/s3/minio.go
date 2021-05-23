package s3

import (
	"context"
	"log"

	"github.com/amupxm/go-video-concat/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorageStruct struct {
	Client *minio.Client
}

var ObjectStorage = ObjectStorageStruct{}

func (object *ObjectStorageStruct) Connect(cfg *config.ConfigStruct) {

	minioClient, err := minio.New(cfg.Storage.Minio_host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.Minio_user, cfg.Storage.Minio_password, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	ObjectStorage.Client = minioClient
}

func InitBuckets(buckets []string) {
	for _, bucketName := range buckets {
		status, err := ObjectStorage.Client.BucketExists(context.Background(), bucketName)
		if err != nil {
			log.Fatal(err)
		}
		if !status {
			if err := ObjectStorage.Client.MakeBucket(context.Background(),
				bucketName,
				minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: false}); err != nil {
				log.Fatal(err)
			}
		}
	}
}
