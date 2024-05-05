package dwrr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDWRR(t *testing.T) {
	dwrr := NewDWRR[int](10, 5)
	assert.Equal(t, 10, len(dwrr.queues))
	assert.Equal(t, 10, len(dwrr.quantums))
	assert.Equal(t, uint(5), dwrr.maxTake)
}

func TestAddQueue(t *testing.T) {
	dwrr := NewDWRR[int](0, 5)
	dwrr.AddQueue()
	assert.Equal(t, 1, len(dwrr.queues))
	assert.Equal(t, 1, len(dwrr.quantums))
}

func TestRemoveQueue(t *testing.T) {
	dwrr := NewDWRR[int](2, 5)
	dwrr.RemoveQueue()
	assert.Equal(t, 1, len(dwrr.queues))
	dwrr.RemoveQueue()
	assert.Equal(t, 0, len(dwrr.queues))
	dwrr.RemoveQueue() // Test removing from empty
	assert.Equal(t, 0, len(dwrr.queues))
}

func TestEnqueueDequeue(t *testing.T) {
	dwrr := NewDWRR[int](1, 5)
	dwrr.Enqueue(0, []int{1, 2, 3})
	assert.Equal(t, []int{1, 2, 3}, dwrr.queues[0])

	item := dwrr.Dequeue(0)
	assert.NotNil(t, item)
	assert.Equal(t, 1, *item)
	assert.Equal(t, []int{2, 3}, dwrr.queues[0])

	// Dequeue until empty
	dwrr.Dequeue(0)
	dwrr.Dequeue(0)
	item = dwrr.Dequeue(0)
	assert.Nil(t, item)
}

func TestDequeueAll(t *testing.T) {
	dwrr := NewDWRR[int](1, 5)
	dwrr.Enqueue(0, []int{1, 2, 3})
	items := dwrr.DequeueAll(0)
	assert.Equal(t, []int{1, 2, 3}, items)
	assert.Equal(t, 0, len(dwrr.queues[0]))
}

func TestDoMultipleRounds(t *testing.T) {
	// Initialize DWRR with 2 queues and a maxTake of 2
	dwrr := NewDWRR[int](2, 2)
	dwrr.Enqueue(0, []int{1, 2, 3, 4, 5}) // More items than maxTake to test multiple rounds
	dwrr.Enqueue(1, []int{5, 6, 7, 8, 9}) // Similarly for second queue

	// First Call to Do
	result := dwrr.Do()
	assert.Equal(t, [][]int{{1, 2}, {5, 6}}, result)
	assert.Equal(t, []int{3, 4, 5}, dwrr.queues[0])
	assert.Equal(t, []int{7, 8, 9}, dwrr.queues[1])

	// Second Call to Do
	result = dwrr.Do()
	assert.Equal(t, [][]int{{3, 4}, {7, 8}}, result)
	assert.Equal(t, []int{5}, dwrr.queues[0])
	assert.Equal(t, []int{9}, dwrr.queues[1])

	// Third Call to Do
	result = dwrr.Do()
	assert.Equal(t, [][]int{{5}, {9}}, result)
	assert.Empty(t, dwrr.queues[0])
	assert.Empty(t, dwrr.queues[1])

	// Fourth Call to Do (Should handle empty queues correctly)
	result = dwrr.Do()
	assert.Equal(t, [][]int{nil, nil}, result)
	assert.Empty(t, dwrr.queues[0])
	assert.Empty(t, dwrr.queues[1])
}
