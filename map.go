package skiplist

import (
	"math"
	"math/rand"
	"sync"
)

// Map is the struct to hold the details of a map
type Map struct {
	comp      func(a, b interface{}) bool
	head      []*mapElement
	mutex     sync.RWMutex
	length    int
	maxLevels int
	r         *rand.Rand
}

// mapElement is the struct to hold elements of the map
type mapElement struct {
	key  interface{}
	val  interface{}
	next []*mapElement
}

// NewMap creates a new empty map, it takes a
// comparison function that should implement Less
func NewMap(less func(a, b interface{}) bool) *Map {
	return &Map{
		comp:      less,
		head:      make([]*mapElement, 32),
		maxLevels: 32,
		r:         rand.New(rand.NewSource(123123))}
}

func newElement(k interface{}, v interface{}, levels int) *mapElement {
	return &mapElement{k, v, make([]*mapElement, levels)}
}

func randomLevels(m *Map) int {
	level := int(math.Log(1.0-m.r.Float64()) / math.Log(1.0-0.5))
	if level >= m.maxLevels {
		level = m.maxLevels
	}
	if level == 0 {
		level++
	}
	return level
}

// Put takes a key and value, and puts the value
// in the map for the key, replacing an existing value.
// returns true if it overwrites, false if it inserts a new key/value pair
func (m *Map) Put(k interface{}, v interface{}) bool {
	m.mutex.Lock()
	backPointer := make([]*mapElement, m.maxLevels)
	for level := m.maxLevels - 1; level >= 0; level-- {
		var e *mapElement = nil
		if level+1 == m.maxLevels || backPointer[level+1] == nil {
			e = m.head[level]
		} else {
			e = backPointer[level+1]
		}
		for e != nil {
			// if they are equal, overwrite
			if m.comp(k, e.key) == m.comp(e.key, k) {
				e.val = v
				m.mutex.Unlock()
				return true
			}
			// if inspected val is greater than k, go back and down a level
			if m.comp(k, e.key) {
				break
			}
			backPointer[level] = e
			e = e.next[level]
		}
	}
	// create new element
	e := newElement(k, v, randomLevels(m))

	// connect new element up with backPointer
	for level := 0; level < len(e.next); level++ {
		if backPointer[level] == nil {
			e.next[level] = m.head[level]
			m.head[level] = e
		} else {
			e.next[level] = backPointer[level].next[level]
			backPointer[level].next[level] = e
		}
	}

	m.length++
	m.mutex.Unlock()
	return false
}

// Len returns the length of a Map
func (m *Map) Len() int {
	m.mutex.RLock()
	e := m.head[0]
	ret := 0
	for e != nil {
		ret++
		e = e.next[0]
	}
	m.mutex.RUnlock()
	return ret
}

// Get returns the value for a key, and true if it finds the key,
// false otherwise
func (m *Map) Get(k interface{}) (interface{}, bool) {
	m.mutex.RLock()
	backPointer := make([]*mapElement, m.maxLevels)
	for level := m.maxLevels - 1; level >= 0; level-- {
		var e *mapElement = nil
		if level+1 == m.maxLevels || backPointer[level+1] == nil {
			e = m.head[level]
		} else {
			e = backPointer[level+1]
		}
		for e != nil {
			// if they are equal, return val
			if m.comp(k, e.key) == m.comp(e.key, k) {
				return e.val, true
			}
			// if inspected val is greater than k, go back and down a level
			if m.comp(k, e.key) {
				break
			}
			backPointer[level] = e
			e = e.next[level]
		}
	}
	m.mutex.RUnlock()
	return nil, false
}

// Remove removes the element (k/v pair) for a key,
// returns true if it found and removed, false otherwise
func (m *Map) Remove(k interface{}) bool {
	m.mutex.Lock()
	backPointer := make([]*mapElement, m.maxLevels)
	for level := m.maxLevels - 1; level >= 0; level-- {
		var e *mapElement = nil
		if level+1 == m.maxLevels || backPointer[level+1] == nil {
			e = m.head[level]
		} else {
			e = backPointer[level+1]
		}
		for e != nil {
			// if they are equal, return val
			if m.comp(k, e.key) == m.comp(e.key, k) {
				for level := 0; level < len(e.next); level++ {
					if backPointer[level] == nil {
						m.head[level] = e.next[level]
					} else {
						backPointer[level].next[level] = e.next[level]
					}
				}

				m.length--
				m.mutex.Unlock()
				return true
			}
			// if inspected val is greater than k, go back and down a level
			if m.comp(k, e.key) {
				break
			}
			backPointer[level] = e
			e = e.next[level]
		}
	}
	m.mutex.Unlock()
	return false
}
