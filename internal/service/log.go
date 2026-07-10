package service

import (
	"context"
	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"
	"log"
)

type LogService struct {
	logv1.UnimplementedLogServiceServer
}

func NewLogService() *LogService {
	return &LogService{}
}

func (s *LogService) CreateLogEntry(ctx context.Context, req *logv1.CreateLogEntryRequest) (*logv1.CreateLogEntryResponse, error) {
	log.Printf("CreateLogEntry called: %+v", req)
	return &logv1.CreateLogEntryResponse{}, nil
}

func (s *LogService) DeleteLogEntry(ctx context.Context, req *logv1.DeleteLogEntryRequest) (*logv1.DeleteLogEntryResponse, error) {
	log.Printf("DeleteLogEntry called: %+v", req)
	return &logv1.DeleteLogEntryResponse{}, nil
}

func (s *LogService) ListLogEntries(ctx context.Context, req *logv1.ListLogEntriesRequest) (*logv1.ListLogEntriesResponse, error) {
	log.Printf("ListLogEntries called: %+v", req)
	return &logv1.ListLogEntriesResponse{}, nil
}

func (s *LogService) ListLogEntriesByDate(ctx context.Context, req *logv1.ListLogEntriesByDateRequest) (*logv1.ListLogEntriesByDateResponse, error) {
	log.Printf("ListLogEntriesByDate called: %+v", req)
	return &logv1.ListLogEntriesByDateResponse{}, nil
}
