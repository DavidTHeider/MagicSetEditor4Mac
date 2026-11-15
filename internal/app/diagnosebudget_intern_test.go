package app

// **Die Diagnose liest nur das Ende des Protokolls** — Befund 3 des Nachlaufs
// über Phase 6.1.
//
// Der Wächter über ganzes Lesen begründete das Lesen des Protokolls mit
// seiner Grenze von 4 MiB. Die gilt aber nur beim Start (`Protokoll` schneidet
// dort ab); während einer Sitzung wächst die Datei ohne Grenze, und die
// Diagnose las sie ganz, um danach zweihundert Zeilen zu behalten.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// **Nicht parallel**, und das ist Absicht: gemessen wird die Zuteilung des
// ganzen Prozesses, und parallele Tests warten, bis die folgenden fertig sind
// (`internal/werkzeuge/zuteilungsmessung_test.go`).
func TestDieDiagnoseLiestNurDasEndeDesProtokolls(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "protokoll.log")

	// **Ein Protokoll mit Loch** — 512 MiB, ohne die Platte zu füllen, und am
	// Ende die Zeilen, auf die es ankommt.
	datei, err := os.Create(pfad) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := datei.WriteAt([]byte("\nletzte Zeile\n"), 512<<20); err != nil {
		t.Fatal(err)
	}

	if err := datei.Close(); err != nil {
		t.Fatal(err)
	}

	var vorher, nachher runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&vorher)

	ende := protokollende(pfad, nil, DialogtexteFuer("de"))

	runtime.ReadMemStats(&nachher)

	if !strings.Contains(ende, "letzte Zeile") {
		t.Errorf("die letzte Zeile fehlt: %q", ende[max(0, len(ende)-80):])
	}

	// **Unter 64 MiB bei einem Protokoll von 512 MiB** — gelesen wird das
	// Fenster von 16 MiB, dazu Zeichenkette und Zeilen daraus.
	if belegt := nachher.TotalAlloc - vorher.TotalAlloc; belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — das Protokoll wurde ganz gelesen", belegt>>20)
	}
}

// **Ein Protokoll, das sich nicht lesen lässt, wird eine Auskunft** — ein
// Verzeichnis an seiner Stelle lässt sich öffnen und vermessen, aber nicht
// lesen.
func TestEinUnlesbaresProtokollWirdEineAuskunft(t *testing.T) {
	t.Parallel()

	texte := DialogtexteFuer("de")

	ende := protokollende(t.TempDir(), nil, texte)
	if !strings.HasPrefix(ende, texte.DiagnoseKeinProtokoll+": ") {
		t.Errorf("ein unlesbares Protokoll ergab %q", ende)
	}
}

// **Die Zeilen werden vom Ende her gezählt** — mehr als die Grenze, genau so
// viele und weniger.
func TestDieLetztenZeilenKommenVomEnde(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		"a\nb\nc\n\n": "b\nc",
		"b\nc":        "b\nc",
		"c\n":         "c",
		"":            "",
	} {
		if bekommen := string(letzteZeilen([]byte(roh), 2)); bekommen != erwartet {
			t.Errorf("%q ergab %q, erwartet %q", roh, bekommen, erwartet)
		}
	}
}
