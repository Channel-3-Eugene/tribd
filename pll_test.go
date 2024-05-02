package main

import (
	"math"
	"testing"
	"time"
)

func TestPLL(t *testing.T) {
	// Example usage
	refFrequency := 100.0                       // Reference frequency (Hz)
	kp := 0.1                                   // Proportional gain
	ki := 0.01                                  // Integral gain
	kd := 0.05                                  // Derivative gain
	setPoint := 0.0                             // Target phase difference
	samplingFrequency := 1000.0                 // Sampling frequency (Hz)
	samplingInterval := 1.0 / samplingFrequency // Sampling interval (seconds)

	// Create a new PLL
	pll := NewPLL(refFrequency, kp, ki, kd)

	// Simulate PLL operation for 1 second
	for t := 0.0; t < 1.0; t += samplingInterval {
		// Simulate phase difference (example: sine wave)
		phaseDifference := math.Sin(2 * math.Pi * t)

		// Update PLL and get output frequency
		outputFrequency := pll.Update(phaseDifference)

		// Simulate some processing time
		time.Sleep(time.Duration(samplingInterval * 1000 * 1000)) // Sleep in microseconds

		// Print output frequency
		println("Output frequency:", outputFrequency)
	}
}
