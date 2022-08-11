package resPool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// I think the test case contains all unit test function
func TestAcquireReleaseNoTimeout(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	src1, err := StringPool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}
	if src1.value != "I am string res" {
		t.Errorf("Acquire() string fail %v", src1.value)
	}

	genericPool := StringPool.(GenericPool[string])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	StringPool.Release(src1)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("after maxIdleTime")

	idleCount = StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

}

type ConnectResource struct {
	connAddr string
	port     int
}

func TestStructAcquireRelease(t *testing.T) {

	f := func(context.Context) (*PoolResource[ConnectResource], error) {
		connectResource := &ConnectResource{
			connAddr: "mysql://localhost",
			port:     3306,
		}
		//testStr := "I am string res"
		poolResource := &PoolResource[ConnectResource]{
			value: *connectResource,
		}
		return poolResource, nil
	}

	StringPool := New[ConnectResource](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	src1, err := StringPool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}
	if src1.value.connAddr != "mysql://localhost" {
		t.Errorf("Acquire() connAddr fail %v", src1.value.connAddr)
	}

	genericPool := StringPool.(GenericPool[ConnectResource])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	StringPool.Release(src1)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("after maxIdleTime")

	idleCount = StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)
}

func TestAcquireReleaseTimeout(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	src1, err := StringPool.Acquire(ctx)
	if err != nil {
		t.Errorf("Acquire() fail %v", err)
	}
	if src1.value != "I am string res" {
		t.Errorf("Acquire() string fail %v", src1.value)
	}

	genericPool := StringPool.(GenericPool[string])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	StringPool.Release(src1)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

	time.Sleep(15 * time.Second)

	fmt.Println("after maxIdleTime")

	idleCount = StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount = genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

}

func TestRepeatRelease(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var srcs []*PoolResource[string]

	for i := 0; i < 3; i++ {

		nowsrc, err := StringPool.Acquire(ctx)
		if err != nil {
			t.Errorf("Acquire() fail %v", err)
		}
		if nowsrc.value != "I am string res" {
			t.Errorf("Acquire() string fail %v", nowsrc.value)
		}
		srcs = append(srcs, nowsrc)

	}

	genericPool := StringPool.(GenericPool[string])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	src1 := srcs[0]

	releaseAndShow(genericPool, src1)
	releaseAndShow(genericPool, src1)
	releaseAndShow(genericPool, src1)
	releaseAndShow(genericPool, src1)

	idleCount := StringPool.NumIdle()
	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

}

func TestRepeatReleaseMG(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var srcs []*PoolResource[string]

	wg := &sync.WaitGroup{}
	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func() {
			nowsrc, err := StringPool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}
			if nowsrc.value != "I am string res" {
				t.Errorf("Acquire() string fail %v", nowsrc.value)
			}
			srcs = append(srcs, nowsrc)
			wg.Done()
		}()
	}

	wg.Wait()

	genericPool := StringPool.(GenericPool[string])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	src1 := srcs[0]

	wg.Add(4)

	for i := 0; i < 4; i++ {
		go func() {
			releaseAndShow(genericPool, src1)
			wg.Done()
		}()
	}

	wg.Wait()

	idleCount := StringPool.NumIdle()
	if idleCount != 1 {
		t.Errorf("idleCount is %v, expected be 1\n", idleCount)
	}

}

func TestReleaseOverLimitMG(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var srcs []*PoolResource[string]

	wg := &sync.WaitGroup{}
	wg.Add(6)

	mtx := &sync.Mutex{}

	for i := 0; i < 6; i++ {
		go func() {
			nowsrc, err := StringPool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}
			if nowsrc.value != "I am string res" {
				t.Errorf("Acquire() string fail %v", nowsrc.value)
			}
			mtx.Lock()
			srcs = append(srcs, nowsrc)
			mtx.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	genericPool := StringPool.(GenericPool[string])

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	wg.Add(6)

	for i := 0; i < 6; i++ {
		go func(idx int) {
			releaseAndShow(genericPool, srcs[idx])
			wg.Done()
		}(i)
	}

	wg.Wait()

	idleCount := StringPool.NumIdle()
	if idleCount != 3 {
		t.Errorf("idleCount is %v, expected be 3\n", idleCount)
	}

}

func TestAcquireOverLimitMG(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	var srcs []*PoolResource[string]

	wg := &sync.WaitGroup{}
	wg.Add(15)

	mtx := &sync.Mutex{}

	for i := 0; i < 15; i++ {
		go func() {
			nowsrc, err := StringPool.Acquire(ctx)
			if err != nil {
				fmt.Printf("Acquire() fail %v\n", err)
			} else {
				if nowsrc.value != "I am string res" {
					t.Errorf("Acquire() string fail %v", nowsrc.value)
				}
				mtx.Lock()
				srcs = append(srcs, nowsrc)
				mtx.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	fmt.Printf("after Acquire\n")

	genericPool := StringPool.(GenericPool[string])

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			releaseAndShow(genericPool, srcs[idx])
			wg.Done()
		}(i)
	}

	wg.Wait()

	idleCount = StringPool.NumIdle()
	if idleCount != 3 {
		t.Errorf("idleCount is %v, expected be 3\n", idleCount)
	}

}

func TestAcquireThenReleaseMG(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(5)

	genericPool := StringPool.(GenericPool[string])

	for i := 0; i < 5; i++ {
		go func() {
			nowsrc, err := StringPool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}
			if nowsrc.value != "I am string res" {
				t.Errorf("Acquire() string fail %v", nowsrc.value)
			}
			// time.Sleep(1 * time.Second)
			releaseAndShow(genericPool, nowsrc)
			wg.Done()
		}()
	}

	wg.Wait()

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	if activeCount != 0 {
		t.Errorf("idleCount is %v, expected be 0\n", idleCount)
	}
}

func TestAcquire2Release1MG(t *testing.T) {
	f := func(context.Context) (*PoolResource[string], error) {
		testStr := "I am string res"
		poolResource := &PoolResource[string]{
			value: testStr,
		}
		return poolResource, nil
	}

	StringPool := New[string](
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(5)

	genericPool := StringPool.(GenericPool[string])

	for i := 0; i < 5; i++ {
		go func() {
			var srcs []*PoolResource[string]
			for i := 0; i < 2; i++ {
				nowsrc, err := StringPool.Acquire(ctx)
				if err != nil {
					t.Errorf("Acquire() fail %v", err)
				}
				if nowsrc.value != "I am string res" {
					t.Errorf("Acquire() string fail %v", nowsrc.value)
				}
				srcs = append(srcs, nowsrc)
			}
			// time.Sleep(1 * time.Second)
			releaseAndShow(genericPool, srcs[0])
			wg.Done()
		}()
	}

	wg.Wait()

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	if idleCount > 3 {
		t.Errorf("idleCount is %v, expected be < 3\n", idleCount)
	}

}

func TestStructAcquireThenReleaseMG(t *testing.T) {
	f := func(context.Context) (*PoolResource[ConnectResource], error) {
		connectResource := &ConnectResource{
			connAddr: "mysql://localhost",
			port:     3306,
		}
		//testStr := "I am string res"
		poolResource := &PoolResource[ConnectResource]{
			value: *connectResource,
		}
		return poolResource, nil
	}

	StringPool := New(
		f,
		3,
		time.Second*10,
	)

	if StringPool.NumIdle() != 0 {
		t.Errorf("NumIdle() != 0 when init")
	}

	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(5)

	genericPool := StringPool.(GenericPool[ConnectResource])

	for i := 0; i < 5; i++ {
		go func() {
			nowsrc, err := StringPool.Acquire(ctx)
			if err != nil {
				t.Errorf("Acquire() fail %v", err)
			}
			if nowsrc.value.connAddr != "mysql://localhost" {
				t.Errorf("Acquire() connAddr fail %v", nowsrc.value.connAddr)
			}
			// time.Sleep(1 * time.Second)
			// releaseAndShow(genericPool, nowsrc)
			releaseAndShowStruct(genericPool, nowsrc)
			wg.Done()
		}()
	}

	wg.Wait()

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)

	idleCount := StringPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	if activeCount != 0 {
		t.Errorf("idleCount is %v, expected be 0\n", idleCount)
	}
}

func releaseAndShow(genericPool GenericPool[string], src *PoolResource[string]) {
	genericPool.Release(src)

	idleCount := genericPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)
}

func releaseAndShowStruct(genericPool GenericPool[ConnectResource], src *PoolResource[ConnectResource]) {
	genericPool.Release(src)

	idleCount := genericPool.NumIdle()
	fmt.Printf("now idleCount %v\n", idleCount)

	activeCount := genericPool.NumActive()
	fmt.Printf("now activeCount %v\n", activeCount)
}
