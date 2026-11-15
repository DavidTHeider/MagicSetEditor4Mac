package msescript

import (
	"testing"
	"time"
)

// uhrAnfang ist der Nullpunkt der Uhr unter Windows.
var uhrAnfang = time.Now()

// rechenzeit misst unter Windows die Wanduhr.
//
// **Windows ist beschrieben und nicht gebaut** (AGENTS.md §12): diese Datei
// hält nur die Übersetzung am Leben. Gemessen wird auf darwin, und dort gilt
// `rechenzeit_unix_test.go`.
func rechenzeit(t *testing.T) time.Duration {
	t.Helper()

	return time.Since(uhrAnfang)
}
