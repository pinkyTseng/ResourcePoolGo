package resPool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"

	"github.com/pinkyTseng/ResourcePoolGo/collections"
)

var defaultMaxSize int = 10

type Pool[T any] interface {
	// This creates or returns a ready-to-use item from the resource pool
	Acquire(context.Context) (*PoolResource[T], error)
	// This releases an active resource back to the resource pool
	Release(*PoolResource[T])
	// This returns the number of idle items
	NumIdle() int
}

// creator is a function called by the pool to create a resource.
// maxIdleSize is the number of maximum idle items kept in the pool
// maxIdleTime is the maximum idle time for an idle item to be swept from the pool
func New[T any](
	creator func(context.Context) (*PoolResource[T], error),
	maxIdleSize int,
	maxIdleTime time.Duration,
) Pool[T] {
	// please implement
	thePool := GenericPool[T]{
		idleObjects:   collections.NewSyncMap(),
		activeObjects: collections.NewSyncMap(),
		PoolIdArr:     collections.NewSyncArr(),
		creator:       creator,
		maxIdleSize:   maxIdleSize,
		maxIdleTime:   maxIdleTime,
		MaxSize:       defaultMaxSize,
		globalMtx:     &sync.RWMutex{},
	}
	return thePool
}

type PoolResource[T any] struct {
	value T
	timer *time.Timer
	id    uintptr
}
type GenericPool[T any] struct {
	idleObjects   *collections.SyncIdentityMap
	activeObjects *collections.SyncIdentityMap
	PoolIdArr     *collections.SyncArr
	creator       func(context.Context) (*PoolResource[T], error)
	maxIdleSize   int
	maxIdleTime   time.Duration
	MaxSize       int
	globalMtx     *sync.RWMutex
}

func (p GenericPool[T]) Acquire(ctx context.Context) (*PoolResource[T], error) {

	p.globalMtx.Lock()
	nowSize := p.activeObjects.Size() + p.idleObjects.Size()

	if nowSize+1 > p.MaxSize {
		p.globalMtx.Unlock()
		return nil, errors.New("over Max resouces limitation now")
	}

	if p.idleObjects.Size() > 0 {
		id := p.PoolIdArr.GetFirst()
		res := p.idleObjects.GetByUintptr(id).(PoolResource[T])
		res.timer.Stop()
		p.PoolIdArr.RemoveFirst()
		p.idleObjects.RemoveByUintptr(id)

		p.activeObjects.PutByUintptr(id, res)

		go func() {

			nowSize = p.activeObjects.Size() + p.idleObjects.Size()
			if nowSize >= p.MaxSize {
				return
			}

			newResAddr, newErr := p.creator(ctx)
			if newErr != nil {
				log.Fatalf("creator error: %v", newErr)
			}

			timer := time.NewTimer(p.maxIdleTime)

			p.globalMtx.Lock()
			theintptr := p.getPoolResourceUintptr(newResAddr)
			p.PoolIdArr.AddByAddr(theintptr)

			newResAddr.id = theintptr
			newResAddr.timer = timer

			p.idleObjects.Put(newResAddr, *newResAddr)
			p.globalMtx.Unlock()

			<-timer.C
			fmt.Println("Acquire maxIdleTime achieved")
			timer.Stop()

			p.globalMtx.Lock()
			p.idleObjects.RemoveByUintptr(theintptr)
			p.PoolIdArr.Remove(theintptr)
			p.globalMtx.Unlock()
		}()
		p.globalMtx.Unlock()
		return &res, nil

	} else {
		newResAddr, newErr := p.creator(ctx)
		if newErr != nil {
			log.Fatalf("creator error: %v", newErr)
		}
		theintptr := p.getPoolResourceUintptr(newResAddr)
		newResAddr.id = theintptr
		p.activeObjects.Put(newResAddr, *newResAddr)
		p.globalMtx.Unlock()
		return newResAddr, nil
	}

}

func (p GenericPool[T]) Release(res *PoolResource[T]) {
	p.globalMtx.Lock()
	if p.activeObjects.GetByUintptr(res.id) == nil {
		p.globalMtx.Unlock()
		return
	} else {
		p.activeObjects.RemoveByUintptr(res.id)
	}

	nowSize := p.activeObjects.Size() + p.idleObjects.Size()

	if p.idleObjects.Size() < p.maxIdleSize && nowSize+1 <= p.MaxSize {

		timer := time.NewTimer(p.maxIdleTime)
		theintptr := p.getPoolResourceUintptr(res)
		p.PoolIdArr.AddByAddr(theintptr)

		res.id = theintptr
		res.timer = timer
		p.idleObjects.Put(res, *res)
		go func() {
			<-timer.C
			fmt.Println("Release maxIdleTime achieved")
			timer.Stop()
			p.globalMtx.Lock()
			p.idleObjects.RemoveByUintptr(theintptr)
			p.PoolIdArr.Remove(theintptr)
			p.globalMtx.Unlock()
		}()
	}
	p.globalMtx.Unlock()
}

func (p GenericPool[T]) NumIdle() int {
	p.globalMtx.RLock()
	num := p.idleObjects.Size()
	p.globalMtx.RUnlock()
	return num
}

func (p GenericPool[T]) NumActive() int {
	p.globalMtx.RLock()
	num := p.activeObjects.Size()
	p.globalMtx.RUnlock()
	return num
}

func (p GenericPool[T]) getPoolResourceUintptr(addr *PoolResource[T]) uintptr {
	return uintptr(unsafe.Pointer(addr))
}

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}
