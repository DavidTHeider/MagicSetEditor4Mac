//go:build unix

package domain_test

import (
	"syscall"
	"testing"
	"time"
)

// rechenzeit gibt die Rechenzeit dieses Prozesses — Nutzer und System.
//
// **Dieselbe Messung wie in `internal/msescript`** und aus demselben Grund
// (ENG-03, #1728):
// ein Verhältnis zweier Wanduhrzeiten misst unter Last die Last. Hier ist der
// Wächter schon am 2026-09-22 so rot geworden — 8,3 bei einer Schranke von 8,
// allein gemessen 3,2. Die Rechenzeit zählt nur, was dieser Prozess tut.
func rechenzeit(t *testing.T) time.Duration {
	t.Helper()

	var verbrauch syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &verbrauch); err != nil {
		t.Fatalf("Rechenzeit lesen: %v", err)
	}

	return time.Duration(verbrauch.Utime.Nano() + verbrauch.Stime.Nano())
}
