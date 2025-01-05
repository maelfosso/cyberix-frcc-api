POSTGRESQL_URL='postgres://frcc:123@localhost:5432/frcc?sslmode=disable'

.PHONY: build cover start test test-integration

compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/main cmd/server/*

cover:
	go tool cover -html=cover.out

start:
	go run cmd/server/*.go

migrate-up:
	migrate -database ${POSTGRESQL_URL} -path storage/migrations up

test:
	go test -coverprofile=cover.out -short ./...

test-integration:
	go test -coverprofile=cover.out -p 1 ./...

ngrok:
	ngrok http --region=us --hostname=api.stockinos.ngrok.io 8000

