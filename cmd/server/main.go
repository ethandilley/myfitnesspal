package main

import (
	"context"
	"log"
	"net"
	"os"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"
	"github.com/ethandilley/myfitnesspal/internal/db"
	"github.com/ethandilley/myfitnesspal/internal/service"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer conn.Close(context.Background())

	queries := db.New(conn)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	foodv1.RegisterFoodServiceServer(grpcServer, service.NewFoodService(queries))
	logv1.RegisterLogServiceServer(grpcServer, service.NewLogService(queries))

	log.Println("server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
