.PHONY: cli server

cli: 
	go run cmd/cli/main.go

server: 
	go run cmd/server/main.go

build:
	docker build -t myfitnesspal .

compose:
	docker compose up --build -d

