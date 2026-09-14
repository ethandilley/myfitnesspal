package main

import (
	"context"
	foodv1connect "github.com/ethandilley/myfitnesspal/gen/proto/food/v1/foodv1connect"
	logv1connect "github.com/ethandilley/myfitnesspal/gen/proto/log/v1/logv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"os"

	connectcors "connectrpc.com/cors"
	"github.com/ethandilley/myfitnesspal/internal/db"
	"github.com/ethandilley/myfitnesspal/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/rs/cors"
)

func withCORS(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "https://my.dilleystone.com"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return c.Handler(h)
}

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL not set")
	}
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer conn.Close(context.Background())

	queries := db.New(conn)

	mux := http.NewServeMux()
	mux.Handle(foodv1connect.NewFoodServiceHandler(service.NewFoodService(queries)))
	mux.Handle(logv1connect.NewLogServiceHandler(service.NewLogService(queries)))

	log.Println("server listening on :50051")
	if err := http.ListenAndServe(":50051", h2c.NewHandler(withCORS(mux), &http2.Server{})); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
