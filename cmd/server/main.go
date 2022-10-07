package main

import (
	"log"
	"net/http"

	"github.com/salvovitale/go-s3-file-server-example/internal/s3"
	"github.com/salvovitale/go-s3-file-server-example/internal/store/postgres"
	"github.com/salvovitale/go-s3-file-server-example/internal/web"
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

	dsn := "postgres://postgres:secret@localhost/postgres?sslmode=disable"

	store, err := postgres.NewStore(dsn)
	if err != nil {
		log.Fatal(err)
	}

	s3Client, err := s3.New(s3url, accessKey, secretKey, false)
	if err != nil {
		log.Fatal(err)
	}

	csrfKey := []byte("01234567890123456789012345678901") //32 bytes long
	h := web.NewHandler(store, s3Client, bucketName, csrfKey)

	http.ListenAndServe(":3000", h)
}
