package fifobuffer

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestFIFOBufferInitialization(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	assert.Equal(t, 0, len(buf.queue), "Buffer should be initialized empty")
}

func TestPush(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	buf.Push(1)
	buf.Push(2)
	assert.Equal(t, 2, len(buf.queue), "Buffer should contain two items after two pushes")
}

func TestPop(t *testing.T) {
	buf := NewFIFOBuffer[int]()
	buf.Push(1)
	buf.Push(2)
	item, ok := buf.Pop()
	assert.True(t, ok, "Pop should succeed")
	assert.Equal(t, 1, item, "First item should be 1")
	item, ok = buf.Pop()
	assert.True(t, ok, "Pop should succeed")
	assert.Equal(t, 2, item, "Second item should be 2")
	_, ok = buf.Pop()
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
