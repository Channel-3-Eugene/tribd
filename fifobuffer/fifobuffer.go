package fifobuffer

import (
	"sync"
)

// T is a type parameter for the FIFOBuffer, allowing it to store any type.
type FIFOBuffer[T any] struct {
	sync.Mutex
	queue []T
}

// NewFIFOBuffer creates a new, empty FIFOBuffer for items of type T.
func NewFIFOBuffer[T any]() *FIFOBuffer[T] {
	return &FIFOBuffer[T]{
		queue: make([]T, 0),
	}
}

// Push adds an item of type T to the buffer.
func (b *FIFOBuffer[T]) Push(item T) {
	b.Lock()
	defer b.Unlock()
	b.queue = append(b.queue, item)
}

// Pop removes and returns the first item from the buffer.
// It returns the item and true if the buffer is not empty, or the zero value and false if it is.
func (b *FIFOBuffer[T]) Pop() (T, bool) {
	b.Lock()
	defer b.Unlock()
	if len(b.queue) == 0 {
		var zero T // Get the zero value of T
		return zero, false
	}
	item := b.queue[0]
	b.queue = b.queue[1:]
	return item, true
}
