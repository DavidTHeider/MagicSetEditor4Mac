package mtgjson_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// Ein Schlüssel, der heißt wie ein belegtes Kartenfeld, wird benannt statt
// mitgeführt.
//
// Mitführen ginge nicht: `card_values` hat den Primärschlüssel
// `(card_id, feld)`, und zwei Werte für dasselbe Feld ließen den ganzen Import
// mit einem Konflikt scheitern, den niemand einer Stelle der Datei zuordnen
// könnte. Der Wert steht deshalb in der Meldung — benannt ist er nicht
// verloren.
func TestSchluesselWieEinBelegtesFeldWirdBenannt(t *testing.T) {
	quelle, err := mtgjson.Lies(pfad("kollision.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	karte := quelle.Sets[0].Karten[0]

	if text, _ := wertVon(karte.Werte, "rule text"); text != "Der zugeordnete Regeltext." {
		t.Errorf("»rule text«: %q — der zugeordnete Wert hat sich nicht durchgesetzt", text)
	}

	for _, wert := range karte.Rest {
		if wert.Field == "rule text" {
			t.Errorf("»rule text« liegt im Rest und kollidierte beim Ablegen: %+v", wert)
		}
	}

	if len(quelle.Befunde) != 1 {
		t.Fatalf("Befunde: %v, erwartet genau einen", quelle.Befunde)
	}

	meldung := quelle.Befunde[0].Meldung
	if !strings.Contains(meldung, "das Kartenfeld heißt genauso") {
		t.Errorf("Meldung: %q — sie sagt nicht, woran es liegt", meldung)
	}

	if !strings.Contains(meldung, "Der Schlüssel, der genauso heißt wie das Kartenfeld.") {
		t.Errorf("Meldung: %q — der Wert fehlt darin und ist damit verloren", meldung)
	}
}

// Was sich nicht lesen lässt, wird mit der Stelle gemeldet.
//
// Fremde Daten sind der Normalfall (§8). Eine Meldung, die nur „ungültig" sagt,
// lässt jemanden in einer Datei mit 200.000 Zeilen suchen.
func TestUnlesbaresWirdGemeldet(t *testing.T) {
	verzeichnis := t.TempDir()

	schreibe := func(name, inhalt string) string {
		t.Helper()

		ort := filepath.Join(verzeichnis, name)
		if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		return ort
	}

	leeresVerzeichnis := filepath.Join(verzeichnis, "leer")
	if err := os.Mkdir(leeresVerzeichnis, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	faelle := map[string]struct {
		ort      string
		enthaelt string
	}{
		"eine Datei, die es nicht gibt": {
			ort:      filepath.Join(verzeichnis, "gibtsnicht.json"),
			enthaelt: "gibtsnicht.json",
		},
		"kein JSON": {
			ort:      schreibe("kaputt.json", "{ das ist kein JSON"),
			enthaelt: "kaputt.json",
		},
		"weder ein Set noch eine Abbildung von Sets": {
			ort:      schreibe("liste.json", `["a", "b"]`),
			enthaelt: "liste.json",
		},
		"ein Verzeichnis ohne eine einzige JSON-Datei": {
			ort:      leeresVerzeichnis,
			enthaelt: "leer",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			_, err := mtgjson.Lies(fall.ort, mtgjson.ZuordnungMagic)
			if err == nil {
				t.Fatalf("Lies(%q): kein Fehler", fall.ort)
			}

			if !strings.Contains(err.Error(), fall.enthaelt) {
				t.Errorf("Fehler: %q — die Stelle %q fehlt darin", err, fall.enthaelt)
			}
		})
	}
}

// Ein unbekanntes Spiel bekommt keine erfundene Zuordnung.
func TestZuordnungFuerNenntNurBekannteSpiele(t *testing.T) {
	if _, bekannt := mtgjson.ZuordnungFuer("magic"); !bekannt {
		t.Error("»magic« ist nicht bekannt")
	}

	if _, bekannt := mtgjson.ZuordnungFuer("probe"); bekannt {
		t.Error("»probe« ist bekannt, obwohl niemand eine Zuordnung dafür aufgenommen hat")
	}
}

// Eine aufgenommene Zuordnung ist danach zu finden.
func TestAufgenommeneZuordnungIstZuFinden(t *testing.T) {
	mtgjson.Aufnehmen(mtgjson.Zuordnung{
		Spiel:  "probespiel",
		Felder: []mtgjson.Feldpaar{{JSON: []string{"name"}, Feld: "name"}},
	})

	zuordnung, bekannt := mtgjson.ZuordnungFuer("probespiel")
	if !bekannt {
		t.Fatal("die aufgenommene Zuordnung ist nicht zu finden")
	}

	if _, kennt := zuordnung.Feld("name"); !kennt {
		t.Error("die aufgenommene Zuordnung kennt »name« nicht")
	}

	if _, kennt := zuordnung.Feld("manaCost"); kennt {
		t.Error("die aufgenommene Zuordnung kennt »manaCost«, obwohl sie es nicht führt")
	}

	// Und sie steht danach in der Liste, die die Oberfläche anbietet.
	if !slices.Contains(mtgjson.Spiele(), "probespiel") {
		t.Errorf("Spiele: %v — die aufgenommene Zuordnung fehlt", mtgjson.Spiele())
	}
}

// wertVon sucht einen Feldwert in einer Werteliste.
func wertVon(werte []domain.Value, feld string) (string, bool) {
	for _, wert := range werte {
		if wert.Field == feld {
			return wert.Text, true
		}
	}

	return "", false
}

// **Ein unlesbares Kartenfeld kostet sein Feld, nicht die Datei** —
// #888.
//
// # Was sich geändert hat, und warum
//
// **Bis zum 2026-08-24 brach ein solcher Wert den ganzen Import ab.** Zwei
// Fälle standen dafür im Test darüber: eine Zahl in `name` und eine
// Zeichenkette in `types`. Beide sind hierher gewandert, und das ist eine
// **Änderung bestehender Semantik** — AGENTS.md §15 Regel 4 verlangt, dass sie
// benannt und begründet wird.
//
// **Der Grund steht in §8**: „Ein Set, von dem 300 Karten lesbar sind und zwei
// nicht, wird mit 300 Karten und zwei benannten Befunden geöffnet, nicht
// abgewiesen." Eine Zahl in **einem** Feld **einer** Karte ließ bis dahin ein
// Set mit 400 Karten ungeöffnet — und die Meldung sagte, welches Feld es war,
// half aber niemandem weiter, der die Datei nicht ändern kann.
//
// **Was es kostet**: wer sich auf den Abbruch verlassen hat, bekommt jetzt ein
// geöffnetes Set mit einem leeren Feld. Deshalb ist der Befund `JeKarte` — er
// wird nicht gebündelt, und die Stelle bleibt einzeln nachlesbar.
func TestEinUnlesbaresKartenfeldKostetNurSeinFeld(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		inhalt string
		feld   string
	}{
		"eine Zahl statt eines Namens": {
			`{"code":"ZAH","name":"Zahlen","cards":[{"name":42,"uuid":"x"}]}`, "name",
		},
		"eine Zeichenkette statt einer Liste": {
			`{"code":"TYP","name":"Typen","cards":[{"name":"Gut","types":"Creature","uuid":"x"}]}`,
			"super type",
		},

		// **Und ein gewöhnliches Textfeld** — es geht durch eine andere
		// Wandlung als die beiden darüber, und ohne diesen Fall bliebe ihr
		// Fehlerzweig unbetreten.
		"eine Zahl in einem Textfeld": {
			`{"code":"ZAH","name":"Zahlen","cards":[{"name":"Gut","power":7,"uuid":"x"}]}`,
			"power",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle, fehler := mtgjson.LiesInhalt(
				[]byte(fall.inhalt), "probe.json", mtgjson.ZuordnungMagic)
			if fehler != nil {
				t.Fatalf("LiesInhalt: %v — der Abbruch ist der alte Weg", fehler)
			}

			if len(quelle.Sets) != 1 || len(quelle.Sets[0].Karten) != 1 {
				t.Fatalf("Quelle = %+v", quelle)
			}

			gefunden := false

			for _, befund := range quelle.Befunde {
				if strings.Contains(befund.Meldung, fall.feld) {
					gefunden = true
				}
			}

			if !gefunden {
				t.Errorf("kein Befund über %q: %+v", fall.feld, quelle.Befunde)
			}
		})
	}
}
