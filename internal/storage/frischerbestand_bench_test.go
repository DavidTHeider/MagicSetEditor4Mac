package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was ein frischer, migrierter Bestand kostet** — J02
// (#2266).
// Fast jeder Test in `storage`, `service` und `app` legt einen an; die
// Teststufe der Pipeline fährt ihn unter `-race` tausendfach.
//
//	go test -run '^$' -bench BenchmarkFrischerBestand ./internal/storage/
//	go test -race -run '^$' -bench BenchmarkFrischerBestand ./internal/storage/
func BenchmarkFrischerBestand(b *testing.B) {
	for b.Loop() {
		db, fehler := storage.Oeffnen(b.Context(), storage.ImArbeitsspeicher)
		if fehler != nil {
			b.Fatal(fehler)
		}

		if fehler := storage.Migriere(b.Context(), db, nil); fehler != nil {
			b.Fatal(fehler)
		}

		_ = db.Close() //nolint:errcheck // gemessen wird das Anlegen
	}
}
