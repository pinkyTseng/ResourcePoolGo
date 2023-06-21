package poolNoLock

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (T, error)
	// This releases an active resource back to the resource pool
	Release(*T)
	// This returns the number of idle items
	NumIdle() int

	Getcnt() int32
}

type NoLockPool[T any] struct {
	idleQueue chan genericRes[T]

	reqQueue chan genericResReq[T]

	creator func(context.Context) (T, error)

	maxIdleTimeout time.Duration

	maxCnt int32
	cnt    int32

	idlePollAddrMap map[*T]bool
	mu              sync.Mutex
}

func (s *NoLockPool[T]) Getcnt() int32 {
	return s.cnt
}

func (s *NoLockPool[T]) GetWaitCount() int {
	return len(s.reqQueue)
}

// creator is a function called by the pool to create a resource.
// maxIdleSize is the number of maximum idle items kept in the pool
// maxIdleTime is the maximum idle time for an idle item to be swept from the pool
func New[T any](
	creator func(context.Context) (T, error),
	maxIdleSize int,
	maxIdleTime time.Duration,
	maxActiveSize int32,
	maxWaitSize int32,
) *NoLockPool[T] {
	thePool := &NoLockPool[T]{
		idleQueue:       make(chan genericRes[T], maxIdleSize),
		reqQueue:        make(chan genericResReq[T], maxWaitSize),
		maxIdleTimeout:  maxIdleTime,
		maxCnt:          maxActiveSize,
		cnt:             0,
		creator:         creator,
		idlePollAddrMap: make(map[*T]bool),
	}
	return thePool
}

func (s *NoLockPool[T]) Acquire(ctx context.Context) (rtv T, err error) {
	for {
		if ctx.Err() != nil {
			err = ctx.Err()
			return
		}
		select {
		case c, ok := <-s.idleQueue:
			if ok {
				s.mu.Lock()
				_, exist := s.idlePollAddrMap[c.objaddr]
				if exist {
					delete(s.idlePollAddrMap, c.objaddr)
				}
				s.mu.Unlock()
				if time.Now().Sub(c.lastActiveTime) > s.maxIdleTimeout {
					continue
				}
				atomic.AddInt32(&s.cnt, 1)
				return c.c, nil
			}
			err = errors.New("pool had closed")
			return

		default:

			cnt := atomic.AddInt32(&s.cnt, 1)
			if cnt <= s.maxCnt {
				rtv, err = s.creator(ctx)
				return
			}
			atomic.AddInt32(&s.cnt, -1)

			// Blocking it
			req := genericResReq[T]{
				ch: make(chan T, 1),
			}
			select {
			case s.reqQueue <- req:
				select {
				case c := <-req.ch:
					rtv = c
					return
				case <-ctx.Done():
					// should take it from the queue
					go func() {
						// got it, but no usage, it is timeout
						<-req.ch
					}()
					// into the queue, but no one give back the resource
					atomic.AddInt32(&s.cnt, -1)
					err = ctx.Err()
					return
				}
			case <-ctx.Done():
				// even no into the queue
				atomic.AddInt32(&s.cnt, -1)
				err = ctx.Err()
				return
			}
		}
	}
}

func (s *NoLockPool[T]) Release(res *T) {
	select {
	case req, ok := <-s.reqQueue:
		if !ok {
			log.Println("When Release pool had been closed!!!")
		}
		req.ch <- *res
	default:
		s.mu.Lock()
		_, exist := s.idlePollAddrMap[res]
		if exist {
			s.mu.Unlock()
			return
		}
		s.idlePollAddrMap[res] = true
		s.mu.Unlock()
		atomic.AddInt32(&s.cnt, -1)
		select {
		case s.idleQueue <- genericRes[T]{
			c:              *res,
			lastActiveTime: time.Now(),
			objaddr:        res,
		}:
		default:

		}
	}

}

func (s *NoLockPool[T]) NumIdle() int {
	return len(s.idleQueue)
}

type genericRes[T any] struct {
	c              T
	lastActiveTime time.Time
	objaddr        *T
}

type genericResReq[T any] struct {
	ch chan T
}
