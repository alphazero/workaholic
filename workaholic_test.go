package workaholic

import (
	"testing"
)

// ============================================================================
// Test assumptions about consts, etc.
// ============================================================================

func TestConstantValueAssumptions_interrupt_code (t *testing.T) {
	for i, code := range Interrupts {
		if !acceptableInterruptCode (code) {
			t.Fatalf("Interrupts[%d]=%d is not an acceptable interrupt_code",i, code)
		}
	}
}
func acceptableInterruptCode (c interrupt_code) bool {
	return c > 0
}
func TestConstantValueAssumptions_status_code (t *testing.T) {
	for i, code := range StatusCodes {
		if !acceptableInterruptCode (code) {
			t.Fatalf("Interrupts[%d]=%d is not an acceptable interrupt_code",i, code)
		}
	}
}
func acceptableStatusCode (c interrupt_code) bool {
	return c > 0
}
