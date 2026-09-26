APP_NAME ?= myfitnesspal

.PHONY: cli server

cli: 
	go run cmd/cli/main.go

server: 
	go run cmd/server/main.go

compose:
	docker compose up --build -d

migrate:
	goose -dir db/migrations postgres "$(DB_URL)" up

build-frontend:
	docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t 192.168.1.201:30500/$(APP_NAME)-frontend:latest \
  --push -f frontend/Dockerfile frontend

build-server:
	docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t 192.168.1.201:30500/$(APP_NAME)-server:latest \
  --push  .

build-migrate:
	docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t 192.168.1.201:30500/$(APP_NAME)-migrate:latest \
  --push -f Dockerfile.migrate .
