package resPoolOld2

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/pinkyTseng/ResourcePoolGo/collections"

// 	"github.com/google/uuid"
// )

// var defaultMaxSize int = 10

// type Pool[T any] interface {
// 	// This creates or returns a ready-to-use item from the resource pool
// 	Acquire(context.Context) (*T, error)
// 	// This releases an active resource back to the resource pool
// 	Release(*T)
// 	// This returns the number of idle items
// 	NumIdle() int
// }

// // creator is a function called by the pool to create a resource.
// // maxIdleSize is the number of maximum idle items kept in the pool
// // maxIdleTime is the maximum idle time for an idle item to be swept from the pool
// func New[T any](
// 	creator func(context.Context) (*T, error),
// 	maxIdleSize int,
// 	maxIdleTime time.Duration,
// ) Pool[T] {
// 	// please implement

// 	thePool := GenericPool[T]{
// 		// TotalSize: 0,
// 		// Resources:   make(chan T, defaultMaxSize),

// 		// Resources:   make(chan Resource[T], defaultMaxSize),
// 		idleObjects:   collections.NewSyncMap(),
// 		activeObjects: collections.NewSyncMap(),
// 		// ResPool:    map[uuid.UUID]Resource[T]{},
// 		PoolIdArr: make([]uuid.UUID, 0, defaultMaxSize),

// 		creator:     creator,
// 		maxIdleSize: maxIdleSize,
// 		maxIdleTime: maxIdleTime,
// 		MaxSize:     defaultMaxSize,
// 	}
// 	// thePool.TotalSize = defaultTotalSize
// 	// thePool.Resources = make(chan T, defaultTotalSize)
// 	// thePool.creator = creator
// 	// thePool.ma

// 	return thePool
// }

// // type AnyRes[T any] struct {
// // }

// type Resource[T any] struct {
// 	value T
// 	// AnyRes[T]
// 	// pool           *Pool[T]
// 	// creationTime time.Time

// 	// lastUsedNano int64
// 	timer *time.Timer
// 	// T

// 	// poolResetCount int
// 	// status         byte
// }

// type GenericPool[T any] struct {
// 	// Resources   chan T
// 	// Resources   chan Resource[T]
// 	// ResPool   map[uuid.UUID]Resource[T]
// 	idleObjects   *collections.SyncIdentityMap
// 	activeObjects *collections.SyncIdentityMap
// 	PoolIdArr     []uuid.UUID
// 	// TotalSize   int
// 	creator     func(context.Context) (*T, error)
// 	maxIdleSize int
// 	maxIdleTime time.Duration

// 	MaxSize int
// }

// func (p GenericPool[T]) Acquire(ctx context.Context) (*T, error) {

// 	nowSize := p.activeObjects.Size() + p.idleObjects.Size()

// 	if nowSize+1 > p.MaxSize {
// 		return nil, errors.New("No more resouce can get now!!")
// 	}

// 	// p.TotalSize++
// 	if len(p.ResPool) > 0 {
// 		// if len(p.Resources) > 0 {
// 		id := p.PoolIdArr[0]
// 		res := p.ResPool[id]
// 		res.timer.Stop()

// 		p.PoolIdArr = p.PoolIdArr[1:]
// 		delete(p.ResPool, id)
// 		// res := <-p.Resources
// 		// res.timer.Stop()
// 		go func() {
// 			// if p.TotalSize >= p.MaxSize {
// 			// 	return
// 			// }

// 			newRes, newErr := p.creator(ctx)
// 			// newResVal, newErr := p.creator(ctx)
// 			if newErr != nil {
// 				log.Fatalf("creator error: %v", newErr)
// 			}

// 			timer := time.NewTimer(p.maxIdleTime)

// 			uuidValue := uuid.New()
// 			p.PoolIdArr = append(p.PoolIdArr, uuidValue)
// 			p.ResPool[uuidValue] = Resource[T]{
// 				value: newRes,
// 				timer: timer,
// 			}
// 			// // p.Resources <- newRes
// 			// p.Resources <- Resource[T]{
// 			// 	value: newRes,
// 			// 	timer: timer,
// 			// 	// newResVal,
// 			// }

// 			select {
// 			case <-timer.C:
// 				fmt.Println("maxIdleTime achieved")
// 				timer.Stop()
// 				p.removeResource(uuidValue)
// 			}
// 			// timer.Stop()
// 		}()
// 		return res.value, nil
// 		// return res, nil
// 	} else {
// 		newRes, newErr := p.creator(ctx)
// 		if newErr != nil {
// 			log.Fatalf("creator error: %v", newErr)
// 		}
// 		return newRes, nil
// 	}
// 	// res := new(T)
// 	// return *res, nil
// }

// func (p GenericPool[T]) removeResource(id uuid.UUID) {
// 	p.PoolIdArr = remove(p.PoolIdArr, id)
// 	delete(p.ResPool, id)
// }

// func remove[T comparable](l []T, item T) []T {
// 	for i, other := range l {
// 		if other == item {
// 			return append(l[:i], l[i+1:]...)
// 		}
// 	}
// 	return l
// }

// func (p GenericPool[T]) Release(res *T) {
// 	if len(p.ResPool) < p.maxIdleSize {
// 		timer := time.NewTimer(p.maxIdleTime)
// 		uuidValue := uuid.New()
// 		p.PoolIdArr = append(p.PoolIdArr, uuidValue)
// 		p.ResPool[uuidValue] = Resource[T]{
// 			value: res,
// 			timer: timer,
// 		}
// 	}
// }

// func (p GenericPool[T]) NumIdle() int {
// 	return len(p.ResPool)
// }

// func init() {
// 	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
// }
