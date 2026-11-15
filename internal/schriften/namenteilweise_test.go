package schriften_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// **Für ihre Namen wird eine Schrift nicht ganz gelesen** — E03
// (#2240).
//
// Findet der Dateiname nichts, öffnet [schriften.Suche] jede Schrift ihrer
// Orte und liest ihren eigenen Namen. Bis E03 las sie dafür jede Datei ganz —
// die Systemschriften dieses Rechners, `Apple Color Emoji.ttc` mit 192 MB
// darunter: rund 860 MB Zuteilung je frischem Zeichenwerk, für Namen, die in
// einer Tabelle von ein paar hundert Byte stehen.
//
// **Gemessen an der Zuteilung über eine aufgeblähte Datei**: dieselbe Schrift
// mit 16 MiB angehängt — die Tabellen stehen über Versätze, und was dahinter
// liegt, liest keine Schrift. Sie steht hinter der gesuchten, damit die Suche
// danach die kleine liest und nicht sie.
func TestFuerIhreNamenWirdEineSchriftNichtGanzGelesen(t *testing.T) {
	quelle, name := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	const angehaengt = 16 << 20

	verzeichnis := t.TempDir()

	if fehler := schreibe(t, verzeichnis, "a-voellig-anders.ttf", roh); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := schreibe(t, verzeichnis, "z-aufgeblaeht.ttf", append(roh, make([]byte, angehaengt)...)); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	suche := schriften.Neue([]string{verzeichnis})

	var vorher, nachher runtime.MemStats

	runtime.ReadMemStats(&vorher)

	_, gesucht := suche.Suche(name, false)

	runtime.ReadMemStats(&nachher)

	if gesucht != nil {
		t.Fatalf("die Schrift %q wurde nicht gefunden: %v", name, gesucht)
	}

	if zugeteilt := nachher.TotalAlloc - vorher.TotalAlloc; zugeteilt >= angehaengt/2 {
		t.Errorf("die Suche teilte %d Byte zu — die aufgeblähte Datei wurde ganz gelesen (%d Byte angehängt)", zugeteilt, angehaengt)
	}
}
