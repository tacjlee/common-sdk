package fxcaches

import (
	"fmt"
	"github.com/dgraph-io/ristretto"
)

var FxCache *ristretto.Cache

func InitializeCache(numberKeys int64, maxCost int64, keysPerBuffer int64) (*ristretto.Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numberKeys,    // number of keys to track frequency of (100.000)
		MaxCost:     maxCost,       // maximum cost of fxcaches (1GB)
		BufferItems: keysPerBuffer, // number of keys per Get buffer
	})
	if err != nil {
		return nil, err
	}
	FxCache = cache
	return cache, nil
}

func InitializeDefaultCache() (*ristretto.Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,     // number of keys to track frequency of (1.000.000)
		MaxCost:     1 << 30, // maximum cost of fxcaches (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, err
	}
	FxCache = cache
	return cache, nil
}

func Set(key string, value interface{}) {
	//currentTimestamp := time.Now().Unix()
	if !FxCache.Set(key, value, 1) {
		fmt.Println("Failed to set item in fxcaches")
	} else {
		FxCache.Wait() // Wait for the fxcaches to process the item
	}
}

func SetCacheWithCost(key string, value interface{}, cost int64) {
	if !FxCache.Set(key, value, cost) {
		fmt.Println("Failed to set item in fxcaches")
	} else {
		FxCache.Wait() // Wait for the fxcaches to process the item
	}
}

func GetIfPresent(key string) interface{} {
	value, found := FxCache.Get(key)
	if found {
		return value
	} else {
		return nil
	}
}

func Clear() {
	FxCache.Clear()
}

func Delete(key interface{}) {
	FxCache.Del(key)
}
