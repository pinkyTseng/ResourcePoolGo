package resPool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pinkyTseng/ResourcePoolGo/collections"
	// "collections"

	"unsafe"
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
		// TotalSize: 0,
		// Resources:   make(chan T, defaultMaxSize),

		// Resources:   make(chan Resource[T], defaultMaxSize),
		idleObjects:   collections.NewSyncMap(),
		activeObjects: collections.NewSyncMap(),
		// ResPool:    map[uuid.UUID]Resource[T]{},
		// PoolIdArr: make([]uuid.UUID, 0, defaultMaxSize),
		// PoolIdArr: make([]uintptr, 0, defaultMaxSize),
		PoolIdArr: collections.NewSyncArr(),

		creator:     creator,
		maxIdleSize: maxIdleSize,
		maxIdleTime: maxIdleTime,
		MaxSize:     defaultMaxSize,
		globalMtx:   &sync.RWMutex{},
	}
	// thePool.TotalSize = defaultTotalSize
	// thePool.Resources = make(chan T, defaultTotalSize)
	// thePool.creator = creator
	// thePool.ma

	return thePool
}

// type AnyRes[T any] struct {
// }

type PoolResource[T any] struct {
	value T
	// AnyRes[T]
	// pool           *Pool[T]
	// creationTime time.Time

	// lastUsedNano int64
	timer *time.Timer
	id    uintptr
	// T

	// poolResetCount int
	// status         byte
}

type GenericPool[T any] struct {
	// Resources   chan T
	// Resources   chan Resource[T]
	// ResPool   map[uuid.UUID]Resource[T]
	idleObjects   *collections.SyncIdentityMap
	activeObjects *collections.SyncIdentityMap
	PoolIdArr     *collections.SyncArr
	// PoolIdArr     []uintptr
	// TotalSize   int
	creator     func(context.Context) (*PoolResource[T], error)
	maxIdleSize int
	maxIdleTime time.Duration

	MaxSize   int
	globalMtx *sync.RWMutex
}

func (p GenericPool[T]) Acquire(ctx context.Context) (*PoolResource[T], error) {

	p.globalMtx.Lock()
	nowSize := p.activeObjects.Size() + p.idleObjects.Size()

	if nowSize+1 > p.MaxSize {
		p.globalMtx.Unlock()
		return nil, errors.New("Over Max resouces limitation now!!")
	}

	// p.TotalSize++
	if p.idleObjects.Size() > 0 {
		// if len(p.Resources) > 0 {
		// id := p.PoolIdArr[0]
		id := p.PoolIdArr.GetFirst()
		// res := p.idleObjects.Get(id).(PoolResource[T])
		res := p.idleObjects.GetByUintptr(id).(PoolResource[T])
		// res := p.ResPool[id]
		res.timer.Stop()
		// p.PoolIdArr = p.PoolIdArr[1:]
		p.PoolIdArr.RemoveFirst()

		// delete(p.ResPool, id)
		// p.idleObjects.Remove(id)
		p.idleObjects.RemoveByUintptr(id)
		// res := <-p.Resources
		// res.timer.Stop()
		go func() {

			nowSize = p.activeObjects.Size() + p.idleObjects.Size()
			if nowSize >= p.MaxSize {
				return
			}

			// p.globalMtx.Lock()
			newResAddr, newErr := p.creator(ctx)
			// p.globalMtx.Unlock()

			// newResVal, newErr := p.creator(ctx)
			if newErr != nil {
				log.Fatalf("creator error: %v", newErr)
			}

			timer := time.NewTimer(p.maxIdleTime)

			p.globalMtx.Lock()
			// uuidValue := uuid.New()
			// p.PoolIdArr = append(p.PoolIdArr, newResAddr)
			theintptr := p.getPoolResourceUintptr(newResAddr)
			p.PoolIdArr.AddByAddr(theintptr)
			// idleObjects

			*&newResAddr.id = theintptr
			*&newResAddr.timer = timer

			// p.idleObjects.Put(newResAddr, *&newResAddr)
			p.idleObjects.Put(newResAddr, *newResAddr)
			p.globalMtx.Unlock()
			// newPoolResource := PoolResource[T]{
			// 	value: newRes,
			// 	timer: timer,
			// 	id: theintptr,
			// }

			// // p.Resources <- newRes
			// p.Resources <- Resource[T]{
			// 	value: newRes,
			// 	timer: timer,
			// 	// newResVal,
			// }

			select {
			case <-timer.C:
				fmt.Println("Acquire maxIdleTime achieved")
				timer.Stop()

				// p.removeResource(uuidValue)

				p.globalMtx.Lock()
				// p.idleObjects.Remove(theintptr)
				p.idleObjects.Remove(newResAddr)
				p.PoolIdArr.Remove(theintptr)
				p.globalMtx.Unlock()

			}
			// timer.Stop()
		}()
		// return res.value, nil
		p.globalMtx.Unlock()
		return &res, nil
	} else {
		newResAddr, newErr := p.creator(ctx)
		if newErr != nil {
			log.Fatalf("creator error: %v", newErr)
		}

		theintptr := p.getPoolResourceUintptr(newResAddr)
		// idleObjects
		*&newResAddr.id = theintptr

		p.activeObjects.Put(newResAddr, *&newResAddr)
		p.globalMtx.Unlock()
		return newResAddr, nil
	}
	// res := new(T)
	// return *res, nil
}

// func (p GenericPool[T]) removeResource(id uuid.UUID) {
// 	p.PoolIdArr = remove(p.PoolIdArr, id)
// 	delete(p.ResPool, id)
// }

func (p GenericPool[T]) Release(res *PoolResource[T]) {
	p.globalMtx.Lock()
	if p.activeObjects.Get(res) == nil {
		p.globalMtx.Unlock()
		return
	} else {
		p.activeObjects.Remove(res)
	}

	// p.globalMtx.Lock()
	if p.idleObjects.Size() < p.maxIdleSize {
		// timer := time.NewTimer(p.maxIdleTime)
		// uuidValue := uuid.New()

		// go func() {

		timer := time.NewTimer(p.maxIdleTime)
		theintptr := p.getPoolResourceUintptr(res)
		p.PoolIdArr.AddByAddr(theintptr)
		// idleObjects

		*&res.id = theintptr
		*&res.timer = timer

		// p.idleObjects.Put(newResAddr, *&newResAddr)
		p.idleObjects.Put(res, *res)
		go func() {

			select {
			case <-timer.C:
				fmt.Println("Release maxIdleTime achieved")
				timer.Stop()

				// p.removeResource(uuidValue)

				// p.idleObjects.Remove(theintptr)
				p.globalMtx.Lock()
				p.idleObjects.Remove(res)
				p.PoolIdArr.Remove(theintptr)
				p.globalMtx.Unlock()

			}

		}()

		// p.PoolIdArr = append(p.PoolIdArr, uuidValue)
		// p.ResPool[uuidValue] = Resource[T]{
		// 	value: res,
		// 	timer: timer,
		// }
	}
	p.globalMtx.Unlock()

	// if len(p.ResPool) < p.maxIdleSize {
	// 	timer := time.NewTimer(p.maxIdleTime)
	// 	uuidValue := uuid.New()
	// 	p.PoolIdArr = append(p.PoolIdArr, uuidValue)
	// 	p.ResPool[uuidValue] = Resource[T]{
	// 		value: res,
	// 		timer: timer,
	// 	}
	// }
}

func (p GenericPool[T]) NumIdle() int {
	p.globalMtx.RLock()
	num := p.idleObjects.Size()
	// return p.idleObjects.Size()
	p.globalMtx.RUnlock()
	return num
}

func (p GenericPool[T]) NumActive() int {
	p.globalMtx.RLock()
	num := p.activeObjects.Size()
	// return p.activeObjects.Size()
	p.globalMtx.RUnlock()
	return num
}

func (p GenericPool[T]) getPoolResourceUintptr(addr *PoolResource[T]) uintptr {
	return uintptr(unsafe.Pointer(addr))
}

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}
