package main

import "sync"

type LockedMap struct {
	l   sync.RWMutex
	Map map[string]string
}

func NewLockedMap() LockedMap {
	return LockedMap{
		Map: make(map[string]string),
	}
}

func (L *LockedMap) Get(key string) (val string, ok bool) {
	L.l.RLock()
	defer L.l.RUnlock()
	val, ok = L.Map[key]
	return
}

func (L *LockedMap) Put(key, val string) {
	L.l.Lock()
	defer L.l.Unlock()
	L.Map[key] = val
}

func (L *LockedMap) Delete(key string) {
	L.l.Lock()
	defer L.l.Unlock()
	delete(L.Map, key)
}
