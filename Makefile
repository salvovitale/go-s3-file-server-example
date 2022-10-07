.PHONY: postgres adminer migrate migrate-down s3-server

postgres:
	docker run --rm -ti -p 5432:5432 -e POSTGRES_PASSWORD=secret postgres

migrate:
	./ci/db/migrations/migrate -source file://ci/db/migrations \
											 -database postgres://postgres:secret@localhost/postgres?sslmode=disable up

migrate-down:
	./ci/db/migrations/migrate -source file://ci/db/migrations \
											 -database postgres://postgres:secret@localhost/postgres?sslmode=disable down

run:
	go run ./cmd/server/main.go

run-w-reflex:
	reflex -s go run ./cmd/server/main.go

s3-server:
	docker run --rm -ti \
	   -p 9000:9000 \
	   -p 9090:9090 \
	   --name minio \
	   -e "MINIO_ROOT_USER=admin" \
	   -e "MINIO_ROOT_PASSWORD=changemeplease" \
	   minio/minio server /data --console-address ":9090"

s3-create-bucket:
	go run ./cmd/s3/main.go