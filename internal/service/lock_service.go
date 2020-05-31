package service

import (
	"context"
	"fmt"
	pb "github.com/stoex/go-lock/internal/generated"
	"github.com/stoex/go-lock/internal/logger"
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
	logger.Info(ctx, fmt.Sprintf("<- get :: resource %s :: lock-id %s :: ttl %d", req.ResourceId, req.LockId, req.Ttl))

	ttl, err := s.redlock.Lock(req.ResourceId, req.LockId, int(req.Ttl))

	if err != nil {
		logger.Error(ctx, "-> get fail")
		return nil, err
	}

	logger.Info(ctx, fmt.Sprintf("-> get ok, ttl: %d", ttl))

	return &pb.LockResponse{
		Status:     1,
		ResourceId: req.ResourceId,
		LockId:     req.LockId,
		Ttl:        uint32(ttl),
	}, nil
}

// RefreshLock is responsible for refreshing / extending a resource lock
func (s *LockService) RefreshLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	logger.Info(ctx, fmt.Sprintf("<- refresh :: resource %s :: lock-id %s :: ttl %d", req.ResourceId, req.LockId, req.Ttl))

	ttl, err := s.redlock.Refresh(req.ResourceId, req.LockId, int(req.Ttl))

	if err != nil {
		logger.Error(ctx, "-> refresh fail")
		return nil, err
	}

	logger.Info(ctx, fmt.Sprintf("-> refresh ok, ttl: %d", ttl))

	return &pb.LockResponse{
		Status:     1,
		ResourceId: req.ResourceId,
		LockId:     req.LockId,
		Ttl:        uint32(ttl),
	}, nil
}

// DeleteLock is responsible for deleting / removing a resource lock
func (s *LockService) DeleteLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	logger.Info(ctx, fmt.Sprintf("<- delete :: resource %s :: lock-id %s", req.ResourceId, req.LockId))

	err := s.redlock.Unlock(req.ResourceId, req.LockId)

	if err != nil {
		logger.Error(ctx, "-> delete fail")
		return nil, err
	}

	logger.Info(ctx, "-> delete ok")

	return &pb.LockResponse{
		Status: 1,
	}, nil
}

// CheckLock returns information about a lock
func (s *LockService) CheckLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	logger.Info(ctx, fmt.Sprintf("<- check :: resource: %s", req.ResourceId))

	l, err := s.redlock.Check(req.ResourceId)

	if err != nil {
		logger.Error(ctx, "-> check fail")
		return nil, err
	}

	logger.Info(ctx, fmt.Sprintf("-> check ok :: resource %s :: lock-id %s :: ttl %d", l.Resource, l.ID, l.TTL))

	return &pb.LockResponse{
		Status:     1,
		ResourceId: l.Resource,
		LockId:     l.ID,
		Ttl:        uint32(l.TTL),
	}, nil
}
