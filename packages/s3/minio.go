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

	m, err := minio.New(cfg.Storage.Minio_host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.Minio_user, cfg.Storage.Minio_password, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	ObjectStorage.Client = m
}

func InitBuckets(buckets []string) {
	for _, bucket := range buckets {
		s, err := ObjectStorage.Client.BucketExists(context.Background(), bucket)
		if err != nil {
			log.Fatal(err)
		}
		if !s {
			if err := ObjectStorage.Client.MakeBucket(context.Background(),
				bucket,
				minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: false}); err != nil {
				log.Fatal(err)
			}
		}
	}
}
