// Package dwrr implements a Deficit Weighted Round Robin (DWRR) scheduler.
// It is a generic package that allows scheduling of any type of items (T).
package dwrr

import "sync"

// DWRR represents a Deficit Weighted Round Robin scheduler for any type.
type DWRR[T any] struct {
	quantums []uint     // Array of quantums, each representing the weight of a queue.
	queues   [][]T      // Array of queues, where each queue holds items of type T.
	maxTake  uint       // Maximum number of items allowed to take from each queue in one cycle.
	mu       sync.Mutex // Mutex to ensure that access to the queues and quantums is thread-safe.
}

// NewDWRR creates a new DWRR scheduler with a specified number of queues and a maxTake limit.
// `count` specifies the number of queues.
// `maxTake` is the maximum number of items that can be processed from each queue per operation cycle.
func NewDWRR[T any](count uint, maxTake uint) *DWRR[T] {
	quantums := make([]uint, count)
	for i := range quantums {
		quantums[i] = 1
	}
	return &DWRR[T]{
		quantums: quantums,
		queues:   make([][]T, count),
		maxTake:  maxTake,
	}
}

// AddQueue appends a new queue to the scheduler.
// Initially, this queue will be empty and have a quantum of 0.
func (dwrr *DWRR[T]) AddQueue() {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	dwrr.queues = append(dwrr.queues, nil)
	dwrr.quantums = append(dwrr.quantums, 0)
}

// RemoveQueue removes the last queue from the scheduler.
// If there are no queues, the function does nothing.
func (dwrr *DWRR[T]) RemoveQueue() {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	if len(dwrr.queues) == 0 {
		return
	}

	dwrr.queues = dwrr.queues[:len(dwrr.queues)-1]
	dwrr.quantums = dwrr.quantums[:len(dwrr.quantums)-1]
}

// Enqueue adds items to a specific queue.
// `queue` is the index of the queue to which items are added.
// `items` is a slice of items of type T to be added to the queue.
// The quantum for the queue is updated to reflect the new queue length.
func (dwrr *DWRR[T]) Enqueue(queue uint, items []T) {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	dwrr.queues[queue] = append(dwrr.queues[queue], items...)
	dwrr.quantums[queue] = uint(len(dwrr.queues[queue]))
}

// Dequeue removes and returns the first item from a specified queue.
// `queue` is the index of the queue from which the item is removed.
// If the queue is empty, it returns nil.
// The quantum for the queue is decremented by one.
func (dwrr *DWRR[T]) Dequeue(queue uint) *T {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	if len(dwrr.queues[queue]) == 0 {
		return nil
	}

	item := dwrr.queues[queue][0]
	dwrr.queues[queue] = dwrr.queues[queue][1:]
	dwrr.quantums[queue]--

	return &item
}

// DequeueAll removes and returns all items from a specified queue.
// `queue` is the index of the queue from which items are removed.
// This operation resets the queue and its corresponding quantum to zero.
func (dwrr *DWRR[T]) DequeueAll(queue uint) []T {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	items := dwrr.queues[queue]
	dwrr.queues[queue] = nil
	dwrr.quantums[queue] = 0

	return items
}

// Do processes each queue based on its quantum and the maxTake limit.
// It returns a slice of slices, each containing the items taken from the respective queue.
// This method ensures that no queue is allowed to take more than its quantum or the maxTake limit.
func (dwrr *DWRR[T]) Do() [][]T {
	dwrr.mu.Lock()
	defer dwrr.mu.Unlock()

	take := make([][]T, len(dwrr.queues))

	for i, queue := range dwrr.queues {
		if len(queue) == 0 {
			dwrr.quantums[i] = 1
			continue
		}

		split := dwrr.quantums[i]

		qlen := uint(len(queue))
		if split > qlen {
			split = qlen
		}

		if split > dwrr.maxTake {
			split = dwrr.maxTake
		}

		// Pre-allocate memory for take[i] slice
		take[i] = make([]T, split)
		copy(take[i], queue[:split])

		// Reuse queue slice by copying the remaining elements
		copy(queue, queue[split:])

		// Trim queue slice
		dwrr.queues[i] = queue[:qlen-split]

		dwrr.quantums[i] = qlen - split
	}

	return take
}
