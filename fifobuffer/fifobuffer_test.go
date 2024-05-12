package fifobuffer

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFIFOBufferInitialization(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	assert.Equal(t, 0, len(buf.queue), "Buffer should be initialized empty")
}

// Generate a random integer using crypto/rand
func randomInt() int {
	var n int32
	binary.Read(rand.Reader, binary.BigEndian, &n) // Reads a 32-bit integer
	return int(n)
}

// Generate a slice of random integers with a length up to maxSize
func randomIntSlice(maxSize int) []int {
	size := randomInt() % maxSize // Ensure size is within the bounds
	if size < 0 {
		size = -size // Convert negative size to positive
	}
	if size == 0 {
		size = 1 // Ensure at least one element
	}

	nums := make([]int, size)
	for i := range nums {
		nums[i] = randomInt()
	}
	return nums
}

func TestPushRandom(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	randomNumbers := randomIntSlice(10) // Generate a random slice with up to 10 integers
	for _, num := range randomNumbers {
		buf.Push(num)
	}
	assert.Equal(t, len(randomNumbers), len(buf.queue), "Buffer should contain the same number of items as pushes")
}

func TestPopRandom(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	randomNumbers := randomIntSlice(10) // Generate a random slice with up to 10 integers
	for _, num := range randomNumbers {
		buf.Push(num)
	}

	for _, expectedNum := range randomNumbers {
		item, ok := buf.Pop()
		assert.True(t, ok, "Pop should succeed")
		assert.Equal(t, expectedNum, item, "Popped item should match the pushed order")
	}

	_, ok := buf.Pop()
	assert.False(t, ok, "Pop should fail on empty buffer")
}

func TestConcurrency(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			buf.Push(val)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, len(buf.queue), "Buffer should contain 100 items after concurrent pushes")

	poppedCount := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := buf.Pop()
			if ok {
				poppedCount++
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, 100, poppedCount, "Should pop 100 items concurrently")
}

func TestZeroValues(t *testing.T) {
	buf := NewFIFOBuffer[string]()
	item, ok := buf.Pop()
	assert.False(t, ok, "Pop should return false for empty buffer")
	assert.Equal(t, "", item, "Pop should return zero value for string when empty")
}
