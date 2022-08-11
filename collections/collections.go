package collections

import (
	"reflect"
	"sync"
)

type Iterator interface {
	HasNext() bool
	Next() interface{}
	Remove()
}

type SyncIdentityMap struct {
	sync.RWMutex
	m map[uintptr]interface{}
}

func NewSyncMap() *SyncIdentityMap {
	return &SyncIdentityMap{m: make(map[uintptr]interface{})}
}

func (m *SyncIdentityMap) Get(key interface{}) interface{} {
	m.RLock()
	keyPtr := genKey(key)
	value := m.m[keyPtr]
	m.RUnlock()
	return value
}

func (m *SyncIdentityMap) GetByUintptr(keyPtr uintptr) interface{} {
	m.RLock()
	value := m.m[keyPtr]
	m.RUnlock()
	return value
}

func genKey(key interface{}) uintptr {
	keyValue := reflect.ValueOf(key)
	return keyValue.Pointer()
}

func (m *SyncIdentityMap) Put(key interface{}, value interface{}) {
	m.Lock()
	keyPtr := genKey(key)
	m.m[keyPtr] = value
	m.Unlock()
}

func (m *SyncIdentityMap) Remove(key interface{}) {
	m.Lock()
	keyPtr := genKey(key)
	delete(m.m, keyPtr)
	m.Unlock()
}

func (m *SyncIdentityMap) RemoveByUintptr(keyPtr uintptr) {
	m.Lock()
	// keyPtr := genKey(key)
	delete(m.m, keyPtr)
	m.Unlock()
}

func (m *SyncIdentityMap) Size() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func (m *SyncIdentityMap) Values() []interface{} {
	m.RLock()
	defer m.RUnlock()
	list := make([]interface{}, len(m.m))
	i := 0
	for _, v := range m.m {
		list[i] = v
		i++
	}
	return list
}

type SyncArr struct {
	sync.RWMutex
	arr []uintptr
}

func NewSyncArr() *SyncArr {
	return &SyncArr{arr: make([]uintptr, 0)}
}

func (m *SyncArr) Remove(item uintptr) {
	m.RLock()
	var delIndex int
	for i, other := range m.arr {
		if other == item {
			delIndex = i
		}
	}
	m.RUnlock()
	m.Lock()
	m.arr = append(m.arr[:delIndex], m.arr[delIndex+1:]...)
	m.Unlock()
}

func (m *SyncArr) RemoveFirst() {
	m.Lock()
	m.arr = append(m.arr[1:])
	m.Unlock()
}

func (m *SyncArr) GetFirst() uintptr {
	m.RLock()
	value := m.arr[0]
	m.RUnlock()
	return value
}

func (m *SyncArr) AddByAddr(val uintptr) {
	m.Lock()
	m.arr = append(m.arr, val)
	m.Unlock()
}
