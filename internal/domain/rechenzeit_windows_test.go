package domain_test

import (
	"testing"
	"time"
)

// uhrAnfang ist der Nullpunkt der Uhr unter Windows.
var uhrAnfang = time.Now()

// rechenzeit misst unter Windows die Wanduhr — Windows ist beschrieben und
// nicht gebaut (AGENTS.md §12); gemessen wird auf darwin.
func rechenzeit(t *testing.T) time.Duration {
	t.Helper()

	return time.Since(uhrAnfang)
}
