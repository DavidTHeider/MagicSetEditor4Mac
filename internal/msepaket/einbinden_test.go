package msepaket_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// dateien ist eine Paketquelle aus dem Arbeitsspeicher.
type dateien map[string]string

func (d dateien) Lies(pfad msepaket.Paketpfad) ([]byte, error) {
	inhalt, da := d[pfad.Paket+"/"+pfad.Datei]
	if !da {
		return nil, errors.New("nicht vorhanden")
	}

	return []byte(inhalt), nil
}

// `include file: /magic-modules.mse-include/scripts` aus magic-m15 Zeile 57.
func TestZerlegePaketpfad(t *testing.T) {
	faelle := []struct {
		name    string
		verweis string
		paket   string
		datei   string
		gut     bool
	}{
		{
			name: "ein Verweis aus magic-m15", verweis: "/magic-modules.mse-include/scripts",
			paket: "magic-modules.mse-include", datei: "scripts", gut: true,
		},
		{
			name: "ein Verweis aus kamen-decade", verweis: "/kamen-icons.mse-include/V3.png",
			paket: "kamen-icons.mse-include", datei: "V3.png", gut: true,
		},
		{
			name: "eine Datei im Unterverzeichnis", verweis: "/a.mse-include/unter/datei",
			paket: "a.mse-include", datei: "unter/datei", gut: true,
		},
		{name: "mit Leerraum", verweis: "  /a.mse-include/b  ", paket: "a.mse-include", datei: "b", gut: true},
		{name: "ohne führenden Schrägstrich", verweis: "a.mse-include/b"},
		{name: "ohne Datei", verweis: "/a.mse-include"},
		{name: "ohne Paket", verweis: "/"},
		{name: "kein Paketname", verweis: "/etc/passwd"},
		{
			name: "ein Versuch, aus dem Paket zu laufen", verweis: "/a.mse-include/../../etc/passwd",
		},
		{
			// **Kein Ausbruch, aber ein Verweis, der etwas anderes tut, als er
			// sagt.** Der erste Anlauf kanonisierte erst und prüfte danach; damit
			// ging dieser hier durch und las aus `b.mse-include`. Gefunden vom
			// Zweitblick über Phase 3.
			name: "ein Wechsel des Pakets über ..", verweis: "/a.mse-include/../b.mse-include/x",
		},
		{name: "ein Punktsegment", verweis: "/a.mse-include/./x"},
		{name: "leer", verweis: ""},
		// **Kein gültiger Text ist kein gültiger Weg** — ein `fs.FS` öffnet
		// ihn nicht (`fs.ValidPath`). Gefunden von `FuzzPaketpfad` (QA-06b).
		{name: "eine Datei, die kein UTF-8 ist", verweis: "/a.mse-include/\xff"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			pfad, err := msepaket.ZerlegePaketpfad(fall.verweis)

			if !fall.gut {
				if !errors.Is(err, msepaket.ErrPaketpfad) {
					t.Errorf("Fehler %v, erwartet war ErrPaketpfad", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if pfad.Paket != fall.paket || pfad.Datei != fall.datei {
				t.Errorf("%+v, erwartet waren %q und %q", pfad, fall.paket, fall.datei)
			}
		})
	}
}

func einbinden(t *testing.T, quelle dateien, inhalt string) (*msepaket.Baum, []msepaket.Befund) {
	t.Helper()

	baum, err := msepaket.Zerteile("style", []byte(inhalt))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle)

	return einbinder.Einbinden(baum), einbinder.Befunde()
}

func TestEinbindenSetztDenInhaltAnDieStelle(t *testing.T) {
	quelle := dateien{"m.mse-include/scripts": "mana_t: eins\nmana_x: zwei\n"}

	baum, befunde := einbinden(t, quelle, "vorher: a\ninclude file: /m.mse-include/scripts\nnachher: b\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	schluessel := make([]string, 0, len(baum.Wurzel))
	for _, knoten := range baum.Wurzel {
		schluessel = append(schluessel, knoten.Schluessel)
	}

	if gebaut := strings.Join(schluessel, ","); gebaut != "vorher,mana_t,mana_x,nachher" {
		t.Errorf("%q — die Reihenfolge stimmt nicht", gebaut)
	}
}

// Ein eingebundener Baum darf selbst einbinden.
func TestEinbindenGehtUeberMehrereStufen(t *testing.T) {
	quelle := dateien{
		"a.mse-include/eins": "include file: /b.mse-include/zwei\n",
		"b.mse-include/zwei": "tief: ja\n",
	}

	baum, befunde := einbinden(t, quelle, "include file: /a.mse-include/eins\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(baum.Wurzel) != 1 || baum.Wurzel[0].Schluessel != "tief" {
		t.Errorf("%+v", baum.Wurzel)
	}
}

// Ein Kreis wird benannt, statt den Stapel zu füllen.
func TestEinKreisUeberIncludeWirdBenannt(t *testing.T) {
	quelle := dateien{
		"a.mse-include/eins": "include file: /b.mse-include/zwei\n",
		"b.mse-include/zwei": "include file: /a.mse-include/eins\n",
	}

	_, befunde := einbinden(t, quelle, "include file: /a.mse-include/eins\n")

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "im Kreis") {
		t.Errorf("die Meldung %q nennt den Kreis nicht", befunde[0].Meldung)
	}
}

// Ein fehlendes Paket ist ein Befund mit Stelle, kein Absturz.
func TestEinFehlendesPaketIstEinBefundMitStelle(t *testing.T) {
	faelle := []struct{ name, inhalt string }{
		{"ein Paket, das nicht da ist", "a: 1\ninclude file: /fehlt.mse-include/x\n"},
		{"ein Verweis, der keiner ist", "a: 1\ninclude file: etc/passwd\n"},
		{"eine Datei, die sich nicht lesen lässt", "a: 1\ninclude file: /kaputt.mse-include/x\n"},
	}

	quelle := dateien{"kaputt.mse-include/x": ": ohne Schlüssel\n"}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			baum, befunde := einbinden(t, quelle, fall.inhalt)

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet war einer", len(befunde))
			}

			if befunde[0].Stelle.Zeile != 2 {
				t.Errorf("Zeile %d, erwartet war 2 — %s",
					befunde[0].Stelle.Zeile, befunde[0].Meldung)
			}

			// Und der Rest des Baums steht noch.
			if len(baum.Wurzel) != 1 || baum.Wurzel[0].Schluessel != "a" {
				t.Errorf("der Rest ging verloren: %+v", baum.Wurzel)
			}
		})
	}
}

// Ein include innerhalb eines Blocks wird ebenfalls aufgelöst.
func TestEinIncludeInEinemBlock(t *testing.T) {
	quelle := dateien{"m.mse-include/s": "drin: ja\n"}

	baum, befunde := einbinden(t, quelle, "block:\n\tinclude file: /m.mse-include/s\n")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	kinder := baum.Wurzel[0].Kinder
	if len(kinder) != 1 || kinder[0].Schluessel != "drin" {
		t.Errorf("%+v", kinder)
	}
}

// Der Eingabebaum bleibt unberührt.
//
// Der erste Anlauf schrieb `kind.Kinder` neu und teilte die Knoten mit dem
// Original — der Aufrufer hielt danach einen Baum, dessen Tiefen teils
// aufgelöst waren und teils nicht. Gefunden vom Zweitblick über Phase 3.
func TestEinbindenLaesstDenEingabebaumUnberuehrt(t *testing.T) {
	quelle := dateien{"m.mse-include/s": "drin: ja\n"}

	baum, err := msepaket.Zerteile("style", []byte("block:\n\tinclude file: /m.mse-include/s\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	vorher := len(baum.Wurzel[0].Kinder)

	neu := msepaket.NeuerEinbinder(quelle).Einbinden(baum)

	if nachher := len(baum.Wurzel[0].Kinder); nachher != vorher {
		t.Errorf("der Eingabebaum hat jetzt %d Kinder statt %d", nachher, vorher)
	}

	if baum.Wurzel[0].Kinder[0].Schluessel != "include file" {
		t.Error("der include-Knoten im Eingabebaum wurde ersetzt")
	}

	// Und die Knoten sind nicht dieselben Zeiger.
	if neu.Wurzel[0] == baum.Wurzel[0] {
		t.Error("Ergebnis und Eingabe teilen sich einen Knoten")
	}
}

// Der Kreis wird als **Kette** genannt, nicht nur als Wort.
func TestDerIncludeKreisNenntDieGanzeKette(t *testing.T) {
	faelle := []struct {
		name   string
		quelle dateien
		kette  string
	}{
		{
			name: "ein Zweierkreis",
			quelle: dateien{
				"a.mse-include/x": "include file: /b.mse-include/y\n",
				"b.mse-include/y": "include file: /a.mse-include/x\n",
			},
			kette: "a.mse-include/x → b.mse-include/y → a.mse-include/x",
		},
		{
			name:   "eine Datei, die sich selbst einbindet",
			quelle: dateien{"a.mse-include/x": "include file: /a.mse-include/x\n"},
			kette:  "a.mse-include/x → a.mse-include/x",
		},
		{
			name: "ein Dreierkreis",
			quelle: dateien{
				"a.mse-include/x": "include file: /b.mse-include/y\n",
				"b.mse-include/y": "include file: /c.mse-include/z\n",
				"c.mse-include/z": "include file: /a.mse-include/x\n",
			},
			kette: "a.mse-include/x → b.mse-include/y → c.mse-include/z → a.mse-include/x",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, befunde := einbinden(t, fall.quelle, "include file: /a.mse-include/x\n")

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde: %v", len(befunde), befunde)
			}

			if !strings.Contains(befunde[0].Meldung, fall.kette) {
				t.Errorf("die Meldung %q nennt die Kette %q nicht",
					befunde[0].Meldung, fall.kette)
			}
		})
	}
}

// Auch tief verschachtelt teilen Ergebnis und Eingabe keinen Knoten.
func TestEinbindenTeiltAuchTiefKeinenKnoten(t *testing.T) {
	baum, err := msepaket.Zerteile("style",
		[]byte("a:\n\tb:\n\t\tc:\n\t\t\td: 1\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	neu := msepaket.NeuerEinbinder(dateien{}).Einbinden(baum)

	alt, kopie := baum.Wurzel[0], neu.Wurzel[0]
	for tiefe := 1; len(alt.Kinder) > 0; tiefe++ {
		if alt == kopie {
			t.Fatalf("in Tiefe %d teilen sich Ergebnis und Eingabe einen Knoten", tiefe)
		}

		alt, kopie = alt.Kinder[0], kopie.Kinder[0]
	}
}
