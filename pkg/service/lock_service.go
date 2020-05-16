package service

import (
	"context"
	pb "github.com/stoex/go-lock/internal/generated"
	"github.com/stoex/go-lock/pkg/redlock"
)

// LockService represents a grpc service handler
type LockService struct {
	redlock *redlock.Redlock
}

// NewLockService returns a pointer to a LockService instance.
// The errors that could be returned from this come from the redis clients.
func NewLockService(addr []string) (*LockService, error) {
	service := LockService{}
	clients, err := redlock.NewRedisClientPool(addr)

	if err != nil {
		return nil, err
	}

	service.redlock = redlock.NewRedlock()

	for _, c := range clients {
		if err := service.redlock.AddRedisClient(c); err != nil {
			return nil, err
		}
	}

	return &service, nil
}

// GetLock is responsible for aquiring a resource lock
func (s *LockService) GetLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	ttl, err := s.redlock.Lock(req.ResourceId, req.LockId, int(req.Ttl))

	if err != nil {
		return nil, err
	}

	return &pb.LockResponse{
		Status:     1,
		ResourceId: req.ResourceId,
		LockId:     req.LockId,
		Ttl:        uint32(ttl),
	}, nil
}

// RefreshLock is responsible for refreshing / extending a resource lock
func (s *LockService) RefreshLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	ttl, err := s.redlock.Refresh(req.ResourceId, req.LockId, int(req.Ttl))

	if err != nil {
		return nil, err
	}

	return &pb.LockResponse{
		Status:     1,
		ResourceId: req.ResourceId,
		LockId:     req.LockId,
		Ttl:        uint32(ttl),
	}, nil
}

// DeleteLock is responsible for deleting / removing a resource lock
func (s *LockService) DeleteLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	err := s.redlock.Unlock(req.ResourceId, req.LockId)

	if err != nil {
		return nil, err
	}

	return &pb.LockResponse{
		Status: 1,
	}, nil
}

// CheckLock returns information about a lock
func (s *LockService) CheckLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	l, err := s.redlock.Check(req.ResourceId)

	if err != nil {
		return nil, err
	}

	return &pb.LockResponse{
		Status:     1,
		ResourceId: l.Resource,
		LockId:     l.ID,
		Ttl:        uint32(l.TTL),
	}, nil
}
