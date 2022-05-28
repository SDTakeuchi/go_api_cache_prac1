package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type item struct {
	value   string
	expires int64
}

func (i *item) IsExpired(time int64) bool {
	if i.expires == 0 {
		return true
	}
	return time > i.expires
}

type Cache struct {
	items map[string]*item
	mu    sync.RWMutex
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	var s string
	if v, ok := c.items[key]; ok {
		s = v.value
	}
	c.mu.RUnlock()
	return s
}

func (c *Cache) Put(key, value string, expires int64) {
	c.mu.Lock()
	if _, found := c.items[key]; !found {
		c.items[key] = &item{
			value:   value,
			expires: expires,
		}
	}
	c.mu.Unlock()
}

func NewCache() *Cache {
	c := &Cache{items: make(map[string]*item)}
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			<-t.C
			c.mu.RLock()
			for k, v := range c.items {
				if v.IsExpired(time.Now().UnixNano()) {
					log.Printf("%v expired at %d", c.items, time.Now())
					delete(c.items, k)
				}
			}
			c.mu.RUnlock()
		}
	}()
	return c
}

func main() {
	http.HandleFunc("/", CacheTestView)
	http.ListenAndServe(":8080", nil)
}

func CacheTestView(w http.ResponseWriter, r *http.Request) {
	fk := "first-key"
	sk := "second-key"

	cache := NewCache()

	cache.Put(fk, "first-value", time.Now().Add(2*time.Second).UnixNano())
	fmt.Println(cache.Get(fk))

	time.Sleep(4 * time.Second)
	if len(cache.Get(fk)) == 0 {
		cache.Put(sk, "second-value", time.Now().Add(30*time.Second).UnixNano())
	}
	fmt.Println(cache.Get(sk))
}
