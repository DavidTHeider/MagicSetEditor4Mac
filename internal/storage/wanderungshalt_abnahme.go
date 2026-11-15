//go:build abnahme

package storage

// **Ein Halt vor dem Festschreiben, nur im Abnahmebau** — M02
// (#2346).
//
// `MSE_ABNAHME_WANDERUNGSHALT` nennt die Dauer, `MSE_ABNAHME_WANDERUNGSMARKE`
// die Datei, die beim Halt entsteht. Die Hüllenprobe `wanderung09` wartet auf
// die Marke und beendet den Prozess dann hart — genau in der offenen
// Transaktion der Wanderung. Im Freigabebau gibt es diese Datei nicht.

import (
	"os"
	"time"
)

func init() {
	frist, err := time.ParseDuration(os.Getenv("MSE_ABNAHME_WANDERUNGSHALT"))
	if err != nil {
		return
	}

	vorDemFestschreiben = func() {
		if marke := os.Getenv("MSE_ABNAHME_WANDERUNGSMARKE"); marke != "" {
			_ = os.WriteFile(marke, []byte("halt\n"), 0o600) //nolint:errcheck,gosec // eine Probe ohne Marke wartet vergeblich und sagt es
		}

		time.Sleep(frist)
	}
}
