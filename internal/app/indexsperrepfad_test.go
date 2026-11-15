package app_test

// **Die Absprache über die Prozesse braucht einen Ort** — Werkstück A14 von
// Phase 5.3 (#1453).
//
// # Der Befund
//
// **`Indexsperre` ist an den Flächen ohne Fenster ein Nichts.** Sie hängt am
// Pfad des Bestands, und `RetrievalAnwendungUeber` baut die Anwendung über
// `NeueApp("")` — ohne Pfad gibt es keine Absprache, und `Nehmen` gibt
// „genommen" zurück, ohne etwas genommen zu haben.
//
// **Der Satz daneben sagt genau das Gegenteil** (`retrievalnaht.go`): „der
// Riegel im Prozess hält das Fenster auf, **nicht** `mse indexaufbau`" — die
// Absprache gibt es also gerade für diese Fläche, und dort wirkt sie nicht.
// Zwei `mse indexaufbau` bauen gleichzeitig, gegen dieselbe Datei.
//
// **Gefunden hat es der Flächenlauf zum laufenden Rebuild** (A14): der zweite
// Aufbau kam mit `0` zurück, wo eine Absage stehen sollte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Anwendung ohne Fenster kennt den Ort ihres Bestands.**
func TestDieRetrievalanwendungKenntIhrenBestandsort(t *testing.T) {
	t.Parallel()

	ort := t.TempDir() + "/bestand.db"

	anwendung := app.RetrievalAnwendungUeber(offeneProbedatenbank(t)).
		MitBestandsort(ort)

	if gelesen := anwendung.Bestandsort(); gelesen != ort {
		t.Errorf("der Ort ist %q statt %q", gelesen, ort)
	}
}

// **Und ohne Ort sagt sie es, statt eine Absprache vorzutäuschen.**
func TestOhneBestandsortGibtEsKeineAbsprache(t *testing.T) {
	t.Parallel()

	if gelesen := app.RetrievalAnwendungUeber(
		offeneProbedatenbank(t)).Bestandsort(); gelesen != "" {
		t.Errorf("der Ort ist %q statt leer", gelesen)
	}
}
