package service

import (
	"context"

	pb "github.com/nibroos/s-erp-api/service/internal/proto"
)

type HealthService struct {
	pb.UnimplementedHealthServiceServer
}

func (s *HealthService) CheckHealth(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Message: "Service is running"}, nil
}
