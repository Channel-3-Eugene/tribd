package pll

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPIDController(t *testing.T) {
	t.Run("A positive delta should decrease the delay", func(t *testing.T) {
		period := 10 * time.Millisecond
		delay := 10 * time.Millisecond
		pll := &PLL{
			kp:     1,
			ki:     1,
			kd:     1,
			period: period,
			delay:  delay,
		}

		delta := 5 * time.Millisecond
		pll.pidController(delta)
		assert.Less(t, pll.delay, delay)
	})

	t.Run("A negative delta should increase the delay", func(t *testing.T) {
		period := 10 * time.Millisecond
		delay := 8 * time.Millisecond
		pll := &PLL{
			kp:     1,
			ki:     1,
			kd:     1,
			period: period,
			delay:  delay,
		}

		delta := -5 * time.Millisecond
		pll.pidController(delta)
		assert.Greater(t, pll.delay, delay)
	})
}
