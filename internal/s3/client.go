package s3

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type S3Client struct {
	*minio.Client
}

func New(url string, accessKey string, secretKey string, secure bool) (*S3Client, error) {
	minioClient, err := CreateClient(url, accessKey, secretKey, secure)
	if err != nil {
		return nil, err
	}
	return &S3Client{minioClient}, nil
}

func (s *S3Client) PutObject(bucketName, objectName string, file io.Reader, size int64) error {
	n, err := s.Client.PutObject(context.Background(), bucketName, objectName, file, size, minio.PutObjectOptions{})
	if err != nil {
		log.Error().Err(err).Msg("error putting object")
		return err
	}
	log.Info().Msgf("object %s uploaded successfully, size: %d", objectName, n.Size)
	return nil
}

func CreateClient(url string, accessKey string, secretKey string, secure bool) (*minio.Client, error) {
	s3Client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}
