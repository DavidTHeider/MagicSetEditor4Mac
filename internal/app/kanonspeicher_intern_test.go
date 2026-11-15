//go:build fremdpakete

package app

// **Was die Zwischenspeicher nach dem Kanonlauf belegen** — PERF03
// (#2353):
// die Bytegrenzen für Skriptbäume und Paketbilder sind Vorgaben des Issues,
// und ob sie den Arbeitssatz tragen, sagt nur ein Lauf über echte Sets. **Ein
// Zwischenspeicher, der öfter räumt als er trifft, ist keiner** (#1288).

import (
	"testing"
)

// nenneSpeicher schreibt Belegung, Treffer und Fehlgriffe als Kennzahlen.
func nenneSpeicher(t *testing.T, anwendung *App) {
	t.Helper()

	stand := anwendung.Zwischenspeicher(t.Context())
	if stand.Fehler != nil {
		t.Errorf("Zwischenspeicher: %+v", stand.Fehler)

		return
	}

	for _, speicher := range stand.Speicher {
		if speicher.Name != SpeicherSkripte && speicher.Name != SpeicherPaketbilder {
			continue
		}

		t.Logf("kennzahl speicher.%s.bytes=%d", speicher.Name, speicher.Bytes)
		t.Logf("kennzahl speicher.%s.eintraege=%d", speicher.Name, speicher.Eintraege)
		t.Logf("kennzahl speicher.%s.treffer=%d", speicher.Name, speicher.Treffer)
		t.Logf("kennzahl speicher.%s.fehlgriffe=%d", speicher.Name, speicher.Fehlgriffe)

		// **Öfter frisch geholt als getroffen** heißt: die Grenze trägt den
		// Arbeitssatz nicht.
		if speicher.Fehlgriffe > speicher.Treffer {
			t.Errorf("%s: %d Fehlgriffe bei %d Treffern — die Grenze trägt den Arbeitssatz nicht",
				speicher.Name, speicher.Fehlgriffe, speicher.Treffer)
		}
	}
}
