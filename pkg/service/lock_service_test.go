package service

import (
	"context"
	"github.com/alicebob/miniredis"
	"github.com/elliotchance/redismock"
	"github.com/go-redis/redis"
	pb "github.com/stoex/go-lock/internal/generated"
	"github.com/stoex/go-lock/pkg/redlock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const (
	testResourceID = "resource"
	testLockID     = "iamalockid"
	testTTL        = 50 // seconds
	bufSize        = 1024 * 1024
)

var (
	rl  *redlock.Redlock
	lis *bufconn.Listener
)

func newTestRedisNode() *redismock.ClientMock {
	mr1, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr1.Addr()})
	return redismock.NewNiceMock(client)
}

// NewTestRedlock returns a mocked redlock instance for testing
func newTestRedlock() (*redlock.Redlock, error) {
	mocks := []*redismock.ClientMock{newTestRedisNode(), newTestRedisNode(), newTestRedisNode()}

	manager := redlock.NewRedlock()
	for i := 0; i < len(mocks); i++ {
		if err := manager.AddRedisClient(mocks[i]); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	rl, _ = newTestRedlock()

	pb.RegisterLockServer(s, &LockService{redlock: rl})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGetLock(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	defer rl.Unlock(testResourceID, testLockID)

	client := pb.NewLockClient(conn)
	res, err := client.GetLock(ctx, &pb.LockRequest{ResourceId: testResourceID, LockId: testLockID, Ttl: testTTL})

	if err != nil {
		t.Fatalf("GetLock failed: %v", err)
	}

	assert.Equal(t, res.Status, pb.ResponseStatus_OK)
	assert.Equal(t, res.LockId, testLockID)
	assert.Equal(t, res.ResourceId, testResourceID)
	assert.LessOrEqual(t, int(res.Ttl), testTTL)
}

func TestRefreshLock(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	defer rl.Unlock(testResourceID, testLockID)

	// set lock
	rl.Lock(testResourceID, testLockID, testTTL)

	client := pb.NewLockClient(conn)
	res, err := client.RefreshLock(ctx, &pb.LockRequest{ResourceId: testResourceID, LockId: testLockID, Ttl: testTTL})

	if err != nil {
		t.Fatalf("GetLock failed: %v", err)
	}

	assert.Equal(t, res.Status, pb.ResponseStatus_OK)
	assert.Equal(t, res.LockId, testLockID)
	assert.Equal(t, res.ResourceId, testResourceID)
	assert.LessOrEqual(t, int(res.Ttl), testTTL)
}

func TestDeleteLock(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	defer rl.Unlock(testResourceID, testLockID)

	// set lock
	rl.Lock(testResourceID, testLockID, testTTL)

	client := pb.NewLockClient(conn)
	res, err := client.DeleteLock(ctx, &pb.LockRequest{ResourceId: testResourceID, LockId: testLockID})

	if err != nil {
		t.Fatalf("GetLock failed: %v", err)
	}

	assert.Equal(t, res.Status, pb.ResponseStatus_OK)
}

func TestCheckLock(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	defer rl.Unlock(testResourceID, testLockID)

	// set lock
	rl.Lock(testResourceID, testLockID, testTTL)

	client := pb.NewLockClient(conn)
	res, err := client.CheckLock(ctx, &pb.LockRequest{ResourceId: testResourceID})

	if err != nil {
		t.Fatalf("GetLock failed: %v", err)
	}

	assert.Equal(t, res.Status, pb.ResponseStatus_OK)
	assert.Equal(t, res.LockId, testLockID)
	assert.Equal(t, res.ResourceId, testResourceID)
	assert.LessOrEqual(t, int(res.Ttl), testTTL)
}
