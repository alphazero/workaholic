package workaholic

import (
	"testing"
)

// ============================================================================
// Test assumptions about consts, etc.
// ============================================================================

func TestConstantValueAssumptions_interrupt_code(t *testing.T) {
	for i, code := range Interrupts {
		if code <= 0 {
			t.Fatalf("Interrupts[%d]=%d is not an acceptable interrupt_code", i, code)
		}
	}
}

//func acceptableInterruptCode (c interrupt_code) bool {
//	return c > 0
//}

func TestConstantValueAssumptions_status_code(t *testing.T) {
	errmsg := [...]string{
		"StatusCodes[%d]=%d is not an acceptable critical status_code",
		"StatusCodes[%d]=%d is not an acceptable noncritical status_code",
	}
	var stype int
	for i, code := range StatusCodes {
		acceptable := true
		switch {
		case i < criticalStats:
			stype = 0
			acceptable = code < 0
		default:
			stype = 1
			acceptable = code >= 0
		}
		if !acceptable {
			t.Fatalf(errmsg[stype], i, code)
		}
	}
}
