package data_structure

import (
	"Nietzsche/internal/config"
	"log"
	"time"
)

type Obj struct {
	Value interface{}
}

type Dict struct {
	dictStore        map[string]*Obj
	expiredDictStore map[string]uint64
}

func CreateDict() *Dict {
	res := Dict{
		dictStore:        make(map[string]*Obj),
		expiredDictStore: make(map[string]uint64),
	}
	return &res
}

func (d *Dict) GetExpireDictStore() map[string]uint64 {
	return d.expiredDictStore
}

func (d *Dict) GetDictStore() map[string]*Obj {
	return d.dictStore
}

func (d *Dict) NewObj(key string, value interface{}, ttlMs int64) *Obj {
	obj := &Obj{
		Value: value,
	}
	if ttlMs > 0 {
		d.SetExpiry(key, ttlMs)
	}
	return obj
}

func (d *Dict) GetExpiry(key string) (uint64, bool) {
	exp, exist := d.expiredDictStore[key]
	return exp, exist
}

func (d *Dict) SetExpiry(key string, ttlMs int64) {
	d.expiredDictStore[key] = uint64(time.Now().UnixMilli()) + uint64(ttlMs)
}

func (d *Dict) HasExpired(key string) bool {
	exp, exist := d.expiredDictStore[key]
	if !exist {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func (d *Dict) Get(k string) *Obj {
	v := d.dictStore[k]
	if v != nil {
		if d.HasExpired(k) {
			d.Del(k)
			return nil
		}
	}
	return v
}

func (d *Dict) evictRandom() {
	evictCount := int64(config.EvictionRatio * float64(config.MaxKeyNumber))
	log.Print("trigger random eviction")
	for k := range d.dictStore {
		d.Del(k)
		evictCount--
		if evictCount == 0 {
			break
		}
	}
}

func (d *Dict) evict() {
	switch config.EvictionPolicy {
	case "allkeys-random":
		d.evictRandom()
	}
}

func (d *Dict) Set(k string, obj *Obj) {
	if len(d.dictStore) == config.MaxKeyNumber {
		d.evict()
	}
	v := d.dictStore[k]
	if v == nil {
		HashKeySpaceStat.Key++
	}
	d.dictStore[k] = obj
}

func (d *Dict) Del(k string) bool {
	log.Printf("Delete key %s", k)
	if _, exist := d.dictStore[k]; exist {
		delete(d.dictStore, k)
		delete(d.expiredDictStore, k)
		HashKeySpaceStat.Key--
		return true
	}
	return false
}
