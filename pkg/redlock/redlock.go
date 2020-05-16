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

// Lock describes a structure holding all relevant lock info.
// Resource is the identifier of what is locked, id is identifying the lock itself
// and ttl is the remaining time in seconds until the lock expires.
type Lock struct {
	// Resource is the identifier for the given resource
	Resource string
	// ID is the id of the lock (the value)
	ID string
	// TTL is the expiry time for this particular lock
	TTL int
}

// NewRedlock creates a Redlock
func NewRedlock() *Redlock {
	return &Redlock{
		retryCount:  DefaultRetryCount,
		retryDelay:  DefaultRetryDelay,
		driftFactor: ClockDriftFactor,
		quorum:      1, // int(math.Floor(float64(1/2)) + 1),
		clients:     nil,
	}
}

// AddRedisClient adds a client to the redlock manager
func (r *Redlock) AddRedisClient(client redis.Cmdable) error {
	if client == nil {
		return errors.New("client is nil")
	}

	r.clients = append(r.clients, client)
	r.quorum = int(math.Floor(float64(len(r.clients)/2)) + 1)

	return nil
}

// AddRedisClientPool adds a pool of redis clients to the redlock manager
func (r *Redlock) AddRedisClientPool(pool []*redis.Client) {
	for _, c := range pool {
		if c != nil {
			r.clients = append(r.clients, c)
		}
	}

	r.quorum = int(math.Floor(float64(len(r.clients)/2)) + 1)
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

// SetDriftFactor sets aquisition lock drift factor in milliseconds
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
	reply := client.Set(resource, val, time.Duration(ttl)*time.Second)
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
		reply := client.Set(resource, lockID, time.Duration(ttl)*time.Second)
		if reply.Val() != "OK" || reply.Err() != nil {
			c <- false
		}
		c <- true
	}
	c <- false
}

func checkLockInstance(client redis.Cmdable, resource string, c chan *Lock) {
	if client == nil {
		c <- nil
	}
	if check := client.Exists(resource); check.Val() == int64(0) {
		c <- nil
	}

	id := client.Get(resource).Val()
	ttl := client.TTL(resource).Val()
	c <- &Lock{resource, id, int(ttl.Seconds())}
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

// Check checks if the lock exists & returns the lock data
func (r *Redlock) Check(resource string) (*Lock, error) {
	for i := 0; i < r.retryCount; i++ {
		c := make(chan *Lock, len(r.clients))

		for _, cli := range r.clients {
			go checkLockInstance(cli, resource, c)
		}
		for j := 0; j < len(r.clients); j++ {
			l := <-c

			if l != nil {
				return l, nil
			}
		}

		// Wait a random delay before to retry
		time.Sleep(time.Duration(rand.Intn(r.retryDelay)) * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to check lock :: resource %s", resource)
}
