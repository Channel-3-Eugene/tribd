package pll

import (
	"sync"
	"time"
)

// PLL represents a Phase-Locked Loop
type PLL struct {
	EventCh   chan bool // Channel for event signal
	TriggerCh chan bool // Channel for PLL output signal (adjusted by PID controller)

	period time.Duration // Period of reference clock signal
	ticker *time.Ticker  // Ticker for reference clock signal
	delay  time.Duration // number of nanoseconds to wait after receiving a ticker signal

	kp int // Proportional gain (over 100)
	ki int // Integral gain (over 100)
	kd int // Derivative gain (over 100)

	integral  int       // Integral term for PID controller
	lastTick  time.Time // last tick time
	lastDelta int       // Last delta (error) for PID controller

	mu sync.Mutex
}

// NewPLL creates a new instance of PLL and initializes it with the specified parameters
func NewPLL(mbps float64, kp, ki, kd int) *PLL {
	const packetSize = 188 * 8                                            // packet size in bits
	bitrate := mbps * 1e6                                                 // convert Mbps to bps
	period := time.Duration(float64(packetSize) / float64(bitrate) * 1e9) // period in nanoseconds
	return &PLL{
		EventCh:   make(chan bool),
		TriggerCh: make(chan bool),
		period:    period,
		delay:     period, // start with the period as the delay
		ticker:    time.NewTicker(period),
		kp:        kp,
		ki:        ki,
		kd:        kd,
	}
}

// Start begins the operation of the PLL
func (pll *PLL) Start() {
	go func() {
		pll.mu.Lock()
		defer pll.mu.Unlock()

		for {
			select {
			case <-pll.ticker.C:
				pll.lastTick = time.Now()
				go func() {
					time.Sleep(time.Duration(pll.delay))
					pll.TriggerCh <- true
				}()
			case <-pll.EventCh:
				// delta should be the difference between now and the next tick
				delta := time.Since(pll.lastTick.Add(pll.period))
				pll.pidController(delta)
			}
		}
	}()
}

// Stop stops the operation of the PLL
func (pll *PLL) Stop() {
	pll.mu.Lock()
	defer pll.mu.Unlock()

	pll.ticker.Stop()
	close(pll.EventCh)
	close(pll.TriggerCh)
}

// pidController implements a PID controller for adjusting the PLL output signal
// all math is integer math here
// we are adjusting the delay to get event to happen near the next tick
func (pll *PLL) pidController(delta time.Duration) {
	porportional := int(delta) * pll.kp / 100
	pll.integral += int(delta) * pll.ki / 100
	derivative := (int(delta) - pll.lastDelta) * pll.kd / 100
	pll.lastDelta = int(delta)

	delay := pll.delay - time.Duration(porportional+pll.integral+derivative)
	if delay > pll.period {
		delay = pll.period
	}
	if delay < 0 {
		delay = 0
	}
	pll.delay = delay
}
