package main

import (
	"log"
	"net"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"
	"github.com/ethandilley/myfitnesspal/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	foodv1.RegisterFoodServiceServer(grpcServer, service.NewFoodService())
	logv1.RegisterLogServiceServer(grpcServer, service.NewLogService())

	log.Println("server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
