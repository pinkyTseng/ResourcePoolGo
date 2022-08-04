package resPool

import (
	"context"
	"log"
	"time"
)

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (T, error)
	// This releases an active resource back to the resource pool
	Release(T)
	// This returns the number of idle items
	NumIdle() int
}

// creator is a function called by the pool to create a resource.
// maxIdleSize is the number of maximum idle items kept in the pool
// maxIdleTime is the maximum idle time for an idle item to be swept from the pool
func New[T any](
	creator func(context.Context) (T, error),
	maxIdleSize int,
	maxIdleTime time.Duration,
) Pool[T] {
	// please implement
}

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	//resourceFactory.GinToRun()
}
