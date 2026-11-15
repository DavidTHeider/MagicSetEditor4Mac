package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einfuhrAus fährt einen Setinhalt ein und gibt, was der Speicher bekam.
func einfuhrAus(t *testing.T, inhalt string) *speicherDouble {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), ort); err != nil {
		t.Fatalf("Import: %v", err)
	}

	return speicher
}

// **Die `keyword:`-Blöcke eines Sets werden gedeutet** —
// #1062.
//
// # Der Befund
//
// Bis hierher fielen sie in den Rohweg: sie überlebten den Rundlauf Byte für
// Byte und waren im Bestand **unsichtbar** — keine Suche fand sie, keine Seite
// zeigte sie, und pflegen ließ sich keines.
//
// MSE2 hat dafür eine eigene Seite (`src/gui/set/keywords_panel.cpp`, 394
// Zeilen, Fassung f9d9356 vom 2021-07-06): links die Liste, rechts der Editor,
// und der Unterschied zwischen **festen** (aus dem Spiel) und **eigenen** (aus
// dem Set).
//
// # Was in der Datei steht
//
// Dieselben fünf Felder wie im Spielpaket — `IMPLEMENT_REFLECTION(Keyword)`,
// `src/data/keyword.cpp`, Zeilen 86–93.
func TestDieSchluesselwoerterEinesSetsWerdenGedeutet(t *testing.T) {
	speicher := einfuhrAus(t, `mse version: 2.0.0
game: probe
keyword:
	keyword: Flügelschlag
	match: Flügelschlag
	mode: core
	reminder: Diese Kreatur kann nur von Fliegern geblockt werden.
	rules: Eine Erfindung dieses Sets.
keyword:
	keyword: Beben
	match: Beben <atom-param>zahl</atom-param>
	mode: custom
card:
	name: Probe
`)

	if len(speicher.set.Keywords) != 2 {
		t.Fatalf("%d Schlüsselwörter: %+v", len(speicher.set.Keywords), speicher.set.Keywords)
	}

	erstes := speicher.set.Keywords[0]
	if erstes.Word != "Flügelschlag" || erstes.Match != "Flügelschlag" ||
		erstes.Mode != "core" || erstes.Rules != "Eine Erfindung dieses Sets." {
		t.Errorf("das erste ist %+v", erstes)
	}

	if erstes.Reminder != "Diese Kreatur kann nur von Fliegern geblockt werden." {
		t.Errorf("die Erinnerung ist %q", erstes.Reminder)
	}

	// **Die Reihenfolge der Datei ist die Reihenfolge der Liste** — sie
	// entscheidet, welches Schlüsselwort zuerst auf einen Text passt
	// (`src/data/keyword.cpp`, Zeilen 148–208, Fassung f9d9356 vom
	// 2021-07-06).
	if speicher.set.Keywords[1].Word != "Beben" {
		t.Errorf("das zweite ist %+v", speicher.set.Keywords[1])
	}

	// **Und sie sind kein Rohblock mehr.** Ein Block, der gedeutet ist und
	// trotzdem roh mitreist, käme beim Schreiben zweimal heraus.
	for _, block := range speicher.roh {
		if block.Text != "" && block.Text[0] != '#' &&
			len(block.Text) > 7 && block.Text[:7] == "keyword" {
			t.Errorf("ein Schlüsselwort reist roh mit: %q", block.Text)
		}
	}
}

// **Ein Set ohne Schlüsselwörter hat keine.**
func TestEinSetOhneSchluesselwoerter(t *testing.T) {
	speicher := einfuhrAus(t, "mse version: 2.0.0\ngame: probe\ncard:\n\tname: Probe\n")

	if len(speicher.set.Keywords) != 0 {
		t.Errorf("gedeutet wurden %+v", speicher.set.Keywords)
	}
}
