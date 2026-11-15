package app_test

// **Der Haushalt am stillen Doppel** — SEC-08 bis SEC-10 (#2149–#2151).
// Eine eigene Datei, damit `faktorvorrichtung_test.go` unter der
// Zeilengrenze aus §2 bleibt.

import (
	"context"
	"time"
)

func (s *stilleFaktoren) SchrittVerbrauchen(context.Context, int64, int64) (bool, error) {
	return false, s.fehler
}

func (s *stilleFaktoren) FaktorversuchBeginnen(context.Context, int64, time.Time) (time.Time, error) {
	return time.Time{}, s.fehler
}

func (s *stilleFaktoren) FaktorversuchGelungen(context.Context, int64) error { return s.fehler }
