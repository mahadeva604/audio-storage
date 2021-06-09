build:
	docker-compose build audio-storage

run:
	docker-compose up audio-storage

test:
	go test -v ./...

test_cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

migrate_up:
	migrate -path ./schema -database 'postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable' up

migrate_down:
	migrate -path ./schema -database 'postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable' down

swag:
	swag init -g cmd/main.go
