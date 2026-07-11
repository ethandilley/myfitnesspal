package client

import (
	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewFoodClient(addr string) (foodv1.FoodServiceClient, func() error, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return foodv1.NewFoodServiceClient(conn), conn.Close, nil
}

func NewLogClient(addr string) (logv1.LogServiceClient, func() error, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return logv1.NewLogServiceClient(conn), conn.Close, nil
}
