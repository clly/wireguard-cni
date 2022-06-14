package server

import (
	"expvar"
	"sync"
)

type Server struct {
	wgKey     *mapDB
	expvarMap *expvar.Map
}

type mapDB struct {
	db map[string]string
	m  *sync.RWMutex
}

func newMapDB() *mapDB {
	return &mapDB{
		db: map[string]string{
			"": "",
		},
		m: &sync.RWMutex{},
	}
}

func (m *mapDB) Set(k string, v string) {
	m.m.Lock()
	m.db[k] = v
	m.m.Unlock()
}

func (m *mapDB) Get(k string) (val string, ok bool) {
	m.m.RLock()
	val, ok = m.db[k]
	m.m.RUnlock()
	return val, ok
}