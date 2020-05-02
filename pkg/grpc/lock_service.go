package grpc

import (
	"context"
	pb "go-lock/internal/generated"
)

type LockService struct{}

func (s *LockService) GetLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

func (s *LockService) RefreshLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

func (s *LockService) DeleteLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

func (s *LockService) CheckLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}
