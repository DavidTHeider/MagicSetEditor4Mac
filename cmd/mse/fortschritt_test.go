package main

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Stand steht in einer Zeile, die sich überschreibt** — FMT-03 (#1692):
// gedrosselt, die erste und die letzte immer, und danach ein Umbruch, damit
// die Ausgabe des Befehls in einer eigenen Zeile beginnt.
func TestDerZeilenmelderUeberschreibtSeineZeile(t *testing.T) {
	var ziel bytes.Buffer

	jetzt := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	melder := zeilenmelder(&ziel, func() time.Time { return jetzt })

	melder(domain.Fortschritt{Erledigt: 0, Gesamt: 3, Einheit: domain.EinheitKarte})
	melder(domain.Fortschritt{Erledigt: 1, Gesamt: 3, Einheit: domain.EinheitKarte}) // zu früh
	jetzt = jetzt.Add(time.Second)
	melder(domain.Fortschritt{Erledigt: 2, Gesamt: 3, Einheit: domain.EinheitKarte})
	melder(domain.Fortschritt{Erledigt: 2, Gesamt: 3, Einheit: domain.EinheitKarte}) // nicht gestiegen
	melder(domain.Fortschritt{Erledigt: 3, Gesamt: 3, Einheit: domain.EinheitKarte}) // die letzte

	erwartet := "\r0 von 3 Karten\x1b[K\r2 von 3 Karten\x1b[K\r3 von 3 Karten\x1b[K\n"
	if ziel.String() != erwartet {
		t.Errorf("%q\nerwartet %q", ziel.String(), erwartet)
	}
}

// **Nur ein Terminal bekommt ihn** — in einer Pipe oder einer Datei wäre er
// Rauschen, das ein Skript wieder herausfiltern müsste.
func TestNurEinTerminalBekommtDenStand(t *testing.T) {
	if _, da := terminalmelder(&bytes.Buffer{}); da {
		t.Error("ein Puffer bekam einen Melder")
	}

	datei, fehler := os.CreateTemp(t.TempDir(), "strom")
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	if _, da := terminalmelder(datei); da {
		t.Error("eine Datei bekam einen Melder")
	}

	// **`/dev/null` ist ein Zeichengerät wie ein Terminal** — der einzige,
	// den ein Lauf ohne Terminal öffnen kann.
	geraet, fehler := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = geraet.Close() }() //nolint:errcheck // nur geschrieben

	melder, da := terminalmelder(geraet)
	if !da {
		t.Fatal("ein Zeichengerät bekam keinen Melder")
	}

	melder(domain.Fortschritt{Erledigt: 1, Gesamt: 2})
}
