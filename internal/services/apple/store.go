package apple

import (
	"sync"
	"time"
)

type timedValue struct {
	value  interface{}
	expiry time.Time
	mutex  sync.Mutex
}

var data = make(map[string]*timedValue)

func StoreValue(key string, value interface{}) {
	expiry := time.Now().Add(10 * time.Minute)

	tv := &timedValue{
		value:  value,
		expiry: expiry,
	}

	data[key] = tv
}

func GetValue(key string) (interface{}, bool) {
	tv, ok := data[key]
	if !ok {
		return nil, false
	}

	tv.mutex.Lock()
	defer tv.mutex.Unlock()

	if time.Now().After(tv.expiry) {
		delete(data, key)
		return nil, false
	}

	return tv.value, true
}
