package poolNoLock

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var showDetail = false

type ConnectResource struct {
	connAddr string
	port     int
}

func fString(context.Context) (string, error) {
	testStr := "I am string res"
	return testStr, nil
}

func fStruct(context.Context) (ConnectResource, error) {
	connectResource := ConnectResource{
		connAddr: "mysql://localhost",
		port:     3306,
	}
	return connectResource, nil
}

func onlyShowIdleAndActiveCounts(idle, active int) {
	if showDetail {
		fmt.Printf("now idleCount %v\n", idle)
		fmt.Printf("now activeCount %v\n", active)
	}
}

func validateIdleAndActiveCounts(idle, idleExp, active, activeExp int, t *testing.T) {
	if showDetail {
		fmt.Printf("now idleCount %v\n", idle)
		fmt.Printf("now activeCount %v\n", active)
	}

	if idle != idleExp {
		t.Errorf("idleCount is %v, expected be %v\n", idle, idleExp)
	}
	if active != activeExp {
		t.Errorf("activeCount is %v, expected be %v\n", active, activeExp)
	}
}

// I think the test case contains all unit test function
// func TestAcquireReleaseNoTimeout(t *testing.T) {
func TestAcquireRelease(t *testing.T) {

	thePool := New(
		fString,
		3,
		time.Second*2,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	src1, err := thePool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}
	if src1 != "I am string res" {
		t.Errorf("Acquire() string fail %v", src1)
	}

	activeCount := thePool.Getcnt()
	idleCount := thePool.NumIdle()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)

	thePool.Release(&src1)

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 1, int(activeCount), 0, t)

	time.Sleep(2 * time.Second)

	if showDetail {
		fmt.Println("after maxIdleTime")
	}

	thePool.Acquire(ctx)

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)
}

func TestStructAcquireRelease(t *testing.T) {
	thePool := New(
		fStruct,
		3,
		time.Second*2,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	src1, err := thePool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}

	if src1.connAddr != "mysql://localhost" {
		t.Errorf("Acquire() connAddr fail %v", src1.connAddr)
	}

	activeCount := thePool.Getcnt()
	idleCount := thePool.NumIdle()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)

	thePool.Release(&src1)

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 1, int(activeCount), 0, t)

	time.Sleep(2 * time.Second)

	if showDetail {
		fmt.Println("after maxIdleTime")
	}

	thePool.Acquire(ctx)

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)
}

func TestBatchAcquireAndRelease(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var reslist []*ConnectResource

	for i := 0; i < 3; i++ {
		v, err := thePool.Acquire(ctx)
		if err != nil {
			t.Errorf("Acquire() fail %v", err)
		}
		reslist = append(reslist, &v)

		idleCount := thePool.NumIdle()
		activeCount := thePool.Getcnt()
		validateIdleAndActiveCounts(idleCount, 0, int(activeCount), i+1, t)
	}

	for i := 0; i < 3; i++ {
		thePool.Release(reslist[i])

		idleCount := thePool.NumIdle()
		activeCount := thePool.Getcnt()
		validateIdleAndActiveCounts(idleCount, i+1, int(activeCount), 3-(i+1), t)
	}

	thePool.Acquire(ctx)

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 2, int(activeCount), 1, t)
}

func TestRepeatRelease(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	validateIdleAndActiveCounts(thePool.NumIdle(), 0, int(thePool.Getcnt()), 0, t)

	ctx := context.Background()

	src1, err := thePool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}

	activeCount := thePool.Getcnt()
	idleCount := thePool.NumIdle()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)

	for i := 0; i < 3; i++ {
		thePool.Release(&src1)

		idleCount = thePool.NumIdle()
		activeCount = thePool.Getcnt()
		validateIdleAndActiveCounts(idleCount, 1, int(activeCount), 0, t)
	}

	thePool.Acquire(ctx)

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 1, t)
}

func TestRepeatReleaseMG(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var reslist []*ConnectResource

	wg := &sync.WaitGroup{}
	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func() {
			v, err := thePool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}
			reslist = append(reslist, &v)
			wg.Done()
		}()
	}

	wg.Wait()

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 3, t)

	src1 := reslist[0]

	wg.Add(4)

	for i := 0; i < 4; i++ {
		go func() {
			thePool.Release(src1)

			idleCount := thePool.NumIdle()
			activeCount := thePool.Getcnt()
			validateIdleAndActiveCounts(idleCount, 1, int(activeCount), 2, t)

			wg.Done()
		}()
	}

	wg.Wait()

	idleCount = thePool.NumIdle()
	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

}

func TestReleaseOverLimitMG(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		10,
		10,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var reslist []*ConnectResource
	wg := &sync.WaitGroup{}
	wg.Add(6)

	mtx := &sync.Mutex{}

	for i := 0; i < 6; i++ {
		go func() {
			nowsrc, err := thePool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}

			mtx.Lock()
			reslist = append(reslist, &nowsrc)
			mtx.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 6, t)

	wg.Add(6)

	for i := 0; i < 6; i++ {
		go func(idx int) {
			thePool.Release(reslist[idx])
			wg.Done()
		}(i)
	}

	wg.Wait()

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 3, int(activeCount), 0, t)

}

func TestAcquireOverLimitMG(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	acquireResult := make(chan *ConnectResource, 6)

	for i := 0; i < 6; i++ {
		go func() {
			nowsrc, err := thePool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			} else {
				acquireResult <- &nowsrc
			}
		}()
	}

	time.Sleep(2 * time.Second)

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 3, t)

	for i := 0; i < 3; i++ {
		nowAddr := <-acquireResult
		thePool.Release(nowAddr)
		idleCount := thePool.NumIdle()
		activeCount := thePool.Getcnt()
		validateIdleAndActiveCounts(idleCount, 0, int(activeCount), 3, t)
	}

	for i := 0; i < 3; i++ {
		nowAddr := <-acquireResult
		thePool.Release(nowAddr)
		idleCount := thePool.NumIdle()
		activeCount := thePool.Getcnt()
		validateIdleAndActiveCounts(idleCount, i+1, int(activeCount), 3-(i+1), t)
	}

	idleCount = thePool.NumIdle()
	activeCount = thePool.Getcnt()
	validateIdleAndActiveCounts(idleCount, 3, int(activeCount), 0, t)

}

func TestStructAcquireThenReleaseMG(t *testing.T) {

	thePool := New(
		fStruct,
		10,
		time.Second*10,
		10,
		10,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	acquireResult := make(chan *ConnectResource, 6)

	for i := 0; i < 5; i++ {
		go func() {
			nowsrc, err := thePool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			} else {
				acquireResult <- &nowsrc
			}
		}()
	}

	execCount := int32(0)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	for i := 0; i < 5; i++ {
		nowAddr := <-acquireResult
		thePool.Release(nowAddr)
		fold := atomic.AddInt32(&execCount, 1)
		if fold == 5 {
			wg.Done()
		}
	}

	wg.Wait()

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	onlyShowIdleAndActiveCounts(idleCount, int(activeCount))
	if activeCount != 0 {
		t.Errorf("activeCount is %v, expected be %v\n", activeCount, 0)
	}

}

func TestAcquireOverLimitMGRandom(t *testing.T) {

	thePool := New(
		fStruct,
		3,
		time.Second*10,
		3,
		3,
	)

	if thePool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	acquireResult := make(chan *ConnectResource, 6)

	for i := 0; i < 15; i++ {
		go func() {
			nowsrc, err := thePool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			} else {
				acquireResult <- &nowsrc
			}
		}()
	}

	execCount := int32(0)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	for i := 0; i < 15; i++ {
		nowAddr := <-acquireResult
		thePool.Release(nowAddr)
		fold := atomic.AddInt32(&execCount, 1)
		if fold == 15 {
			wg.Done()
		}
	}

	wg.Wait()

	idleCount := thePool.NumIdle()
	activeCount := thePool.Getcnt()
	onlyShowIdleAndActiveCounts(idleCount, int(activeCount))
	if activeCount != 0 {
		t.Errorf("activeCount is %v, expected be %v\n", activeCount, 0)
	}

}
