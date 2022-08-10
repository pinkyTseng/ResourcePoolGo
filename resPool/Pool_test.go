package resPool

import (
	"context"
	"fmt"
	"testing"
	"time"
	// "unsafe"
)

// type CheckStruct[T any] struct {
// 	// Resources   chan T
// 	// Resources   chan Resource[T]
// 	v1   T
// 	size int
// }

// type MyInner struct {
// 	id      string
// 	thetype string
// }

// type MyOutter struct {
// 	MyInner
// 	name string
// 	size int
// }

// func passMyOutter(s1 MyOutter) {
// 	fmt.Printf("s %v\n", s1)
// 	//s1AddrCalle := &s
// 	fmt.Printf("s1AddrCalle %p\n", &s1)
// 	//fmt.Printf("s1AddrCalle %p\n", s1AddrCalle)

// }
// func TestInner(t *testing.T) {

// 	s1 := MyOutter{}
// 	s1.id = "123"
// 	s1.thetype = "t1"
// 	s1.name = "s1"
// 	s1.size = 10

// 	fmt.Printf("s1 %v\n", s1)

// 	// var s1Addr string =  &s1

// 	s1Addr := &s1
// 	fmt.Printf("s1Addr %p\n", s1Addr)
// 	s1up := unsafe.Pointer(&s1)
// 	fmt.Printf("s1up %v\n", s1up)
// 	s1uptr := uintptr(s1up)
// 	fmt.Printf("s1uptr %v\n", s1uptr)

// 	passMyOutter(s1)

// 	s2 := MyOutter{}
// 	s2.id = "123"
// 	s2.thetype = "t1"
// 	s2.name = "s1"
// 	s2.size = 10
// 	s2Addr := &s2
// 	fmt.Printf("s2Addr %p\n", s2Addr)

// 	// s2 := MyInner{}
// 	// s2.id = "aaa"
// 	// s2.thetype = "t2"
// 	//s2.extId = "eid"

// 	// useInner(s1.)

// }

// func useInner(s MyInner) {
// 	fmt.Printf("id %v", s.id)
// 	fmt.Printf("thetype %v", s.thetype)
// }

// func TestStruct(t *testing.T) {
// 	// ts := "sto"
// 	s1 := &CheckStruct[string]{
// 		v1:   "sto",
// 		size: 10,
// 	}
// 	fmt.Println(s1)
// 	ss1 := *s1
// 	fmt.Println(&s1)
// 	fmt.Println(&ss1)
// 	s2 := &CheckStruct[string]{
// 		v1:   "sto",
// 		size: 10,
// 	}
// 	fmt.Println(s2)
// 	fmt.Println(&s2)

// 	if s1 == s2 {
// 		fmt.Println("s1 == s2")
// 	}
// }

func TestStringPool(t *testing.T) {
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

// func TestCheckAge(t *testing.T) {
// 	vo1 := model.UserVO{Id: 100, Name: "adult", Age: 22}
// 	if !checkAge(vo1) {
// 		t.Errorf("age: %v should be Ok but show fail", vo1.Age)
// 	}

// 	vo2 := model.UserVO{Id: 99, Name: "child", Age: 15}
// 	if checkAge(vo2) {
// 		t.Errorf("age: %v is too young but show OK", vo2.Age)
// 	}
// }
