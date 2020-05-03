package redlock

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
)

const (
	// DefaultRetryCount is the max retry times for lock acquire
	DefaultRetryCount = 10

	// DefaultRetryDelay is upper wait time in millisecond for lock acquire retry
	DefaultRetryDelay = 200

	// ClockDriftFactor is clock drift factor, more information refers to doc
	ClockDriftFactor = 0.01
)

// Redlock holds the redis lock
type Redlock struct {
	retryCount  int
	retryDelay  int
	driftFactor float64

	clients []redis.Cmdable
	quorum  int
}

// NewRedlock creates a Redlock
func NewRedlock(client redis.Cmdable) (*Redlock, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	return &Redlock{
		retryCount:  DefaultRetryCount,
		retryDelay:  DefaultRetryDelay,
		driftFactor: ClockDriftFactor,
		quorum:      1, // int(math.Floor(float64(1/2)) + 1),
		clients:     []redis.Cmdable{client},
	}, nil
}

// AddClient adds a client to the redlock manager
func (r *Redlock) AddClient(client redis.Cmdable) error {
	if client == nil {
		return errors.New("client is nil")
	}

	r.clients = append(r.clients, client)
	r.quorum = int(math.Floor(float64(len(r.clients)/2)) + 1)

	return nil
}

// SetRetryCount sets acquire lock retry count
func (r *Redlock) SetRetryCount(count int) {
	if count <= 0 {
		return
	}
	r.retryCount = count
}

// SetRetryDelay sets acquire lock retry max internal in millisecond
func (r *Redlock) SetRetryDelay(delay int) {
	if delay <= 0 {
		return
	}
	r.retryDelay = delay
}

// SetDriftFactor sets aquire lock drift factor in milliseconds
func (r *Redlock) SetDriftFactor(fac float64) {
	if fac <= 0 {
		return
	}
	r.driftFactor = fac
}

func lockInstance(client redis.Cmdable, resource string, val string, ttl int, c chan bool) {
	if client == nil {
		c <- false
		return
	}
	if check := client.Exists(resource); check.Val() == int64(1) {
		c <- false
	}
	reply := client.Set(resource, val, time.Duration(ttl)*time.Millisecond)
	if reply.Err() != nil || reply.Val() != "OK" {
		c <- false
		return
	}
	c <- true
}

func unlockInstance(client redis.Cmdable, resource string, lockID string, c chan bool) {
	if client == nil {
		c <- false
	}
	if check := client.Get(resource); check.Val() == lockID {
		reply := client.Del(resource)
		if reply.Val() != int64(1) || reply.Err() != nil {
			c <- false
		}
		c <- true
	}

	c <- false
}

func refreshInstance(client redis.Cmdable, resource string, lockID string, ttl int, c chan bool) {
	if client == nil {
		c <- false
	}
	if check := client.Get(resource); check.Val() == lockID {
		reply := client.Set(resource, lockID, time.Duration(ttl)*time.Millisecond)
		if reply.Val() != "OK" || reply.Err() != nil {
			c <- false
		}
		c <- true
	}
	c <- false
}

// Lock acquires a distribute lock
func (r *Redlock) Lock(resource string, lockID string, ttl int) (int64, error) {
	for i := 0; i < r.retryCount; i++ {
		c := make(chan bool, len(r.clients))
		success := 0
		start := time.Now()

		for _, cli := range r.clients {
			go lockInstance(cli, resource, lockID, ttl, c)
		}
		for j := 0; j < len(r.clients); j++ {
			if <-c {
				success++
			}
		}

		drift := int(float64(ttl)*r.driftFactor) + 2
		costTime := time.Since(start).Nanoseconds() / 1e6
		validityTime := int64(ttl) - costTime - int64(drift)
		if success >= r.quorum && validityTime > 0 {
			return validityTime, nil
		}
		// Wait a random delay before to retry
		time.Sleep(time.Duration(rand.Intn(r.retryDelay)) * time.Millisecond)
	}

	return 0, fmt.Errorf("failed to aquire lock :: resource %s :: lock id %s", resource, lockID)
}

// Unlock releases an acquired lock
func (r *Redlock) Unlock(resource string, lockID string) error {
	c := make(chan bool, len(r.clients))
	success := 0

	for _, cli := range r.clients {
		go unlockInstance(cli, resource, lockID, c)
	}
	for i := 0; i < len(r.clients); i++ {
		if <-c {
			success++
		}
	}

	if success >= r.quorum {
		return nil
	}

	return fmt.Errorf("failed to unlock :: resource %s :: lock id %s", resource, lockID)
}

// Refresh checks if the lock exists & refreshes the ttl
func (r *Redlock) Refresh(resource string, lockID string, ttl int) (int64, error) {
	for i := 0; i < r.retryCount; i++ {
		c := make(chan bool, len(r.clients))
		success := 0
		start := time.Now()

		for _, cli := range r.clients {
			go refreshInstance(cli, resource, lockID, ttl, c)
		}
		for j := 0; j < len(r.clients); j++ {
			if <-c {
				success++
			}
		}

		drift := int(float64(ttl)*r.driftFactor) + 2
		costTime := time.Since(start).Nanoseconds() / 1e6
		validityTime := int64(ttl) - costTime - int64(drift)
		if success >= r.quorum && validityTime > 0 {
			return validityTime, nil
		}
		// Wait a random delay before to retry
		time.Sleep(time.Duration(rand.Intn(r.retryDelay)) * time.Millisecond)
	}

	return 0, fmt.Errorf("failed to refresh lock :: resource %s :: lock id %s", resource, lockID)
}
