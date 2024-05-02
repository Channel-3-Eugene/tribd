package main

import (
	"math"
	"time"
)

// PLL represents a Phase-Locked Loop.
type PLL struct {
	refFrequency float64 // Reference frequency (Hz)
	kp           float64 // Proportional gain
	ki           float64 // Integral gain
	kd           float64 // Derivative gain
	setPoint     float64 // Target phase difference
	prevError    float64 // Previous error
	integral     float64 // Integral of error
}

// NewPLL creates a new instance of PLL with specified parameters.
func NewPLL(refFrequency, kp, ki, kd float64) *PLL {
	return &PLL{
		refFrequency: refFrequency,
		kp:           kp,
		ki:           ki,
		kd:           kd,
	}
}

// Update adjusts the PLL's set point based on the current phase difference.
func (pll *PLL) Update(phaseDifference float64) float64 {
	error := pll.setPoint - phaseDifference

	// Proportional term
	proportional := pll.kp * error

	// Integral term
	pll.integral += error
	integral := pll.ki * pll.integral

	// Derivative term
	derivative := pll.kd * (error - pll.prevError)

	// Calculate output frequency
	outputFrequency := pll.refFrequency + proportional + integral + derivative

	// Store current error for next iteration
	pll.prevError = error

	return outputFrequency
}
