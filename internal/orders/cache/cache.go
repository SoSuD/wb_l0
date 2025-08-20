package cache

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] struct {
	mu      sync.Mutex
	cap     int
	ll      *list.List // список от MRU (front) к LRU (back)
	cache   map[K]*list.Element
	onEvict func(k K, v V)
}

type entry[K comparable, V any] struct {
	key K
	val V
}

func New[K comparable, V any](capacity int, onEvict func(K, V)) *LRU[K, V] {
	if capacity <= 0 {
		panic("lru: capacity must be > 0")
	}
	return &LRU[K, V]{
		cap:     capacity,
		ll:      list.New(),
		cache:   make(map[K]*list.Element),
		onEvict: onEvict,
	}
}

func (l *LRU[K, V]) Get(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var zero V
	if ele, ok := l.cache[key]; ok {
		l.ll.MoveToFront(ele)
		return ele.Value.(*entry[K, V]).val, true
	}
	return zero, false
}

func (l *LRU[K, V]) Put(key K, val V) (evicted bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if ele, ok := l.cache[key]; ok {
		ent := ele.Value.(*entry[K, V])
		ent.val = val
		l.ll.MoveToFront(ele)
		return false
	}

	ent := &entry[K, V]{key: key, val: val}
	ele := l.ll.PushFront(ent)
	l.cache[key] = ele

	if l.ll.Len() > l.cap {
		l.evict(1)
		return true
	}
	return false
}

func (l *LRU[K, V]) Remove(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	ele, ok := l.cache[key]
	if !ok {
		return false
	}
	ent := ele.Value.(*entry[K, V])
	delete(l.cache, key)
	l.ll.Remove(ele)
	if l.onEvict != nil {
		l.onEvict(ent.key, ent.val)
	}
	return true
}

func (l *LRU[K, V]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.ll.Len()
}

func (l *LRU[K, V]) Purge() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.onEvict != nil {
		for k, ele := range l.cache {
			ent := ele.Value.(*entry[K, V])
			l.onEvict(k, ent.val)
		}
	}
	l.ll.Init()
	l.cache = make(map[K]*list.Element)
}

func (l *LRU[K, V]) Keys() []K {
	l.mu.Lock()
	defer l.mu.Unlock()
	keys := make([]K, 0, len(l.cache))
	for e := l.ll.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry[K, V]).key)
	}
	return keys
}

func (l *LRU[K, V]) evict(n int) {
	for i := 0; i < n; i++ {
		ele := l.ll.Back()
		if ele == nil {
			return
		}
		ent := ele.Value.(*entry[K, V])
		delete(l.cache, ent.key)
		l.ll.Remove(ele)
		if l.onEvict != nil {
			l.onEvict(ent.key, ent.val)
		}
	}
}
