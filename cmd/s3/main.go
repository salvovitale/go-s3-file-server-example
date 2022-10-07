package main

import (
	"github.com/rs/zerolog/log"
	"github.com/salvovitale/go-s3-file-server-example/internal/s3"
)

// TODO pass this value as command line arguments
const (
	s3url      = "localhost:9000"
	accessKey  = "Nnml9tuG5lnV4r2n"
	secretKey  = "CRt3rgP9hiPr6JcQldtSIeRVzsJ7or4o"
	region     = "eu-west-1"
	bucketName = "test-bucket"
)

func main() {
	s3Client, err := s3.CreateClient(s3url, accessKey, secretKey, false)
	if err != nil {
		log.Error().Err(err).Msg("error creating s3 client")
	}

	found, err := s3.BucketExists(s3Client, bucketName)
	if err != nil {
		log.Error().Err(err).Msg("error checking if bucket exists")
	}

	if !found {
		err = s3.CreateBucket(s3Client, bucketName, region)
		if err != nil {
			log.Error().Err(err).Msg("error creating bucket")
		}
		log.Info().Msg("bucket created successfully")
	} else {
		log.Info().Msg("bucket already exists")
	}
}
