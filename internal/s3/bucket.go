package s3

import (
	"context"

	"github.com/minio/minio-go/v7"
)

func CreateBucket(s3Client *minio.Client, bucketName string, region string) error {
	opts := minio.MakeBucketOptions{
		Region: region,
	}
	err := s3Client.MakeBucket(context.Background(), bucketName, opts)
	if err != nil {
		return err
	}
	return nil
}

func BucketExists(s3Client *minio.Client, bucketName string) (bool, error) {
	exists, err := s3Client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return false, err
	}
	return exists, nil
}
