package redlock

import (
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/elliotchance/redismock"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testResourceID = "resource"
	testLockID     = "iamalockid"
	testTTL        = 50000000
)

func newTestRedisNode() *redismock.ClientMock {
	mr1, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr1.Addr()})
	return redismock.NewNiceMock(client)
}

func newTestRedlock() (*Redlock, error) {
	mocks := []*redismock.ClientMock{newTestRedisNode(), newTestRedisNode(), newTestRedisNode()}

	manager := NewRedlock()
	for i := 0; i < len(mocks); i++ {
		if err := manager.AddRedisClient(mocks[i]); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func TestRedlock_Lock(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	ttl, err := redlock.Lock(testResourceID, testLockID, testTTL)
	assert.LessOrEqual(t, int(ttl), testTTL, "ttl should be less than or equal 1000")
}

func TestRedlock_LockFailQuorum(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	// Mocking responses for the second and third client
	redlock.clients[0].(*redismock.ClientMock).
		On("Set", testResourceID, testLockID, time.Duration(testTTL)*time.Millisecond).
		Return(redis.NewStatusResult("0", nil))
	redlock.clients[1].(*redismock.ClientMock).
		On("Set", testResourceID, testLockID, time.Duration(testTTL)*time.Millisecond).
		Return(redis.NewStatusResult("0", nil))

	ttl, err := redlock.Lock(testResourceID, testLockID, testTTL)

	assert.Equal(t, 0, int(ttl), "ttl should be 0")
	assert.Error(t, err, "redlock should return an error")
}

func TestRedlock_LockFailAlreadyExists(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	// Setting the lock values beforehand
	for _, client := range redlock.clients {
		client.Set(testResourceID, testLockID, testTTL)
	}
	ttl, err := redlock.Lock(testResourceID, testLockID, testTTL)
	assert.Equal(t, int(ttl), 0, "ttl should be 0")
	assert.Error(t, err, "redlock should return an error")
}

func TestRedlock_Unlock(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	for _, client := range redlock.clients {
		client.Set(testResourceID, testLockID, testTTL)
	}

	err = redlock.Unlock(testResourceID, testLockID)
	assert.NoError(t, err, "unlock should not return an error")
}

func TestRedlock_UnlockFailNotExists(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}

	err = redlock.Unlock(testResourceID, testLockID)
	assert.Error(t, err, "unlock should return an error")
}

func TestRedlock_UnlockFailQuorum(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	for _, client := range redlock.clients {
		client.Set(testResourceID, testLockID, testTTL)
	}
	// Mocking responses for the second and third client
	redlock.clients[1].(*redismock.ClientMock).
		On("Del", []string{testResourceID}).
		Return(redis.NewIntCmd("0", nil))
	redlock.clients[2].(*redismock.ClientMock).
		On("Del", []string{testResourceID}).
		Return(redis.NewIntCmd("0", nil))

	err = redlock.Unlock(testResourceID, testLockID)

	assert.Error(t, err, "unlock should return an error")
}

func TestRedlock_Refresh(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}
	for _, client := range redlock.clients {
		client.Set(testResourceID, testLockID, testTTL)
	}

	ttl, err := redlock.Refresh(testResourceID, testLockID, testTTL)
	assert.LessOrEqual(t, int(ttl), testTTL, "ttl should be less than or equal 1000")
}

func TestRedlock_RefreshFailNotExists(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}

	ttl, err := redlock.Refresh(testResourceID, testLockID, testTTL)
	assert.Equal(t, 0, int(ttl), "refresh ttl should be 0")
	assert.Error(t, err, "refresh should return an error")
}

func TestRedlock_RefreshFailQuorum(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}

	redlock.clients[0].Set(testResourceID, testLockID, time.Duration(testTTL)*time.Millisecond)

	// Mocking responses for the second and third client
	redlock.clients[1].(*redismock.ClientMock).
		On("Set", testResourceID, testLockID, time.Duration(testTTL)*time.Millisecond).
		Return(redis.NewStatusResult("0", nil))
	redlock.clients[2].(*redismock.ClientMock).
		On("Set", testResourceID, testLockID, time.Duration(testTTL)*time.Millisecond).
		Return(redis.NewStatusResult("0", nil))

	ttl, err := redlock.Refresh(testResourceID, testLockID, testTTL)
	assert.Equal(t, 0, int(ttl), "refresh ttl should be 0")
	assert.Error(t, err, "refresh should return an error")
}

func TestRedlock_CheckLock(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}

	_, err = redlock.Lock(testResourceID, testLockID, testTTL)

	if err != nil {
		t.Fatal(fmt.Sprintf("could not create a lock: %s", err.Error()))
	}

	l, err := redlock.Check(testResourceID)

	if err != nil {
		t.Fatal(fmt.Sprintf("check lock failed: %s", err.Error()))
	}

	assert.NotNil(t, l, "lock details should not be nil")
	assert.Equal(t, testResourceID, l.Resource, "returned resource should match")
	assert.Equal(t, testLockID, l.ID, "returned lock value should match")
	assert.LessOrEqual(t, testTTL, l.TTL, fmt.Sprintf("returned lock ttl (%d) should be less or equal to testTTL (%d)", l.TTL, testTTL))
}

func TestRedlock_CheckLockError(t *testing.T) {
	redlock, err := newTestRedlock()
	if err != nil {
		t.Fatal(fmt.Sprintf("could not create redlock instance: %s", err.Error()))
	}

	_, err = redlock.Check(testResourceID)
	assert.Error(t, err, "lock does not exist - check should return an error")
}
