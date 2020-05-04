package grpc

import (
	"context"
	pb "github.com/stoex/go-lock/internal/generated"
)

// LockService represents a grpc service hanlder
type LockService struct{}

// GetLock is responsible for aquiring a resource lock
func (s *LockService) GetLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

// RefreshLock is responsible for refreshing / extending a resource lock
func (s *LockService) RefreshLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

// DeleteLock is responsible for deleting / removing a resource lock
func (s *LockService) DeleteLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}

// CheckLock returns information about a lock
func (s *LockService) CheckLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {

}
