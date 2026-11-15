package msepaket_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die Rundlaufzusage, in einem Satz geprüft** (`docs/entscheidungen/0002`):
// was gelesen wurde, kommt Byte für Byte wieder heraus.
//
// Nicht „sieht gleich aus", nicht „hat dieselben Schlüssel" — dieselben Bytes.
// Alles andere hat der Zweitblick über Phase 2 schon einmal als unzureichend
// benannt: ein aus Schlüssel und Wert zusammengesetzter Block verliert
// `key:  wert` gegenüber `key: wert`.
func TestSchreibeGibtDasGeleseneZurueck(t *testing.T) {
	faelle := []struct {
		name   string
		inhalt string
	}{
		{
			"Schlüssel, Werte, Einrückung",
			"mse version: 0.3.8\ngame: magic\nset info:\n\tsymbol: s.mse-symbol\n",
		},
		{
			"zwei Abstände nach dem Doppelpunkt",
			"key:  wert\nandere:\twert\n",
		},
		{
			"gemischte Zeilenenden in derselben Datei",
			"eins: a\r\nzwei: b\ndrei: c\r\n",
		},
		{
			"ohne Zeilenende am Ende",
			"eins: a\nzwei: b",
		},
		{
			"Leerzeilen und Kommentare an ihrer Stelle",
			"eins: a\n\n# ein Kommentar\n\nzwei: b\n",
		},
		{
			"wiederholte Schlüssel in ihrer Reihenfolge",
			"depends on: a\ndepends on: b\ndepends on: c\n",
		},
		{
			"tiefe Verschachtelung",
			"a:\n\tb:\n\t\tc:\n\t\t\td: tief\n",
		},
		{
			"eine leere Datei",
			"",
		},
		{
			// **Ein Schlüsselwortblock, wie er im echten Pack steht**
			// (#1060).
			//
			// Gelesen wird er seit dem 2026-08-24 — und **geschrieben wird
			// weiterhin die Rohzeile**. Der Leser baut aus dem Baum einen
			// eigenen Verbund und rührt den Baum nicht an; ohne diesen Fall
			// stünde das nur im Kommentar.
			"ein Schlüsselwortblock",
			"keyword:\n\tkeyword: Fliegend\n\tmatch: Fliegend\n\tmode: core\n" +
				"\treminder: {this_or_that(\"creature\", upper:true)} kann fliegen.\n",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			baum, err := msepaket.Zerteile("probe", []byte(fall.inhalt))
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			if gebaut := string(msepaket.Schreibe(baum)); gebaut != fall.inhalt {
				t.Errorf("geschrieben %q, gelesen war %q", gebaut, fall.inhalt)
			}
		})
	}
}

// **Die BOM gehört dazu.**
//
// Sie steht am Anfang jeder gesichteten MSE-Datei (`docs/FORMATE.md`). Der
// Zerteiler schnitt sie ab und merkte sie sich nicht — ein Rundlauf hätte sie
// stillschweigend entfernt, und die Datei wäre um drei Bytes kürzer
// zurückgekommen als sie hineinging.
func TestDieBOMUeberlebtDenRundlauf(t *testing.T) {
	mitBOM := "\uFEFFmse version: 0.3.8\n"

	baum, err := msepaket.Zerteile("probe", []byte(mitBOM))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if gebaut := string(msepaket.Schreibe(baum)); gebaut != mitBOM {
		t.Errorf("geschrieben %q, gelesen war %q", gebaut, mitBOM)
	}

	// Und ohne BOM kommt keine dazu.
	ohne, err := msepaket.Zerteile("probe", []byte("mse version: 0.3.8\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if gebaut := string(msepaket.Schreibe(ohne)); gebaut != "mse version: 0.3.8\n" {
		t.Errorf("geschrieben %q", gebaut)
	}
}

// **Gegen eine echte Paketdatei**, nicht nur gegen selbst getippte Zeichenketten
// (AGENTS.md §3): eine Zeichenkette im Test enthält genau die Fälle, an die der
// Schreibende gedacht hat.
func TestDerRundlaufGegenEineEchtePaketdatei(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "probe.mse-game"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	roh, err := paket.Datei("game")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	baum, err := msepaket.Zerteile("game", roh)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if gebaut := msepaket.Schreibe(baum); !bytes.Equal(gebaut, roh) {
		t.Errorf("der Rundlauf änderte die Datei:\n%q\n%q", gebaut, roh)
	}
}

// Ein Baum, den niemand gelesen hat, lässt sich trotzdem schreiben.
//
// Für den Fall, dass später jemand eine Datei **erzeugt**: dann trägt kein
// Knoten eine Rohzeile, und der Schreiber muss sie aus Schlüssel und Wert
// bauen. Er tut das mit einem Tabulator je Ebene und einem Leerzeichen nach dem
// Doppelpunkt — die Form, die jede gesichtete Datei benutzt.
func TestEinSelbstGebauterBaumWirdLesbarGeschrieben(t *testing.T) {
	baum := &msepaket.Baum{Datei: "neu", Wurzel: []*msepaket.Knoten{
		{Schluessel: "mse version", Wert: "0.3.8"},
		{Schluessel: "card", Kinder: []*msepaket.Knoten{
			{Schluessel: "name", Wert: "Feuerball"},
		}},
	}}

	erwartet := "mse version: 0.3.8\ncard:\n\tname: Feuerball\n"

	if gebaut := string(msepaket.Schreibe(baum)); gebaut != erwartet {
		t.Errorf("geschrieben %q, erwartet war %q", gebaut, erwartet)
	}

	// Und was so entsteht, lässt sich wieder lesen.
	gelesen, err := msepaket.Zerteile("neu", msepaket.Schreibe(baum))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if knoten := gelesen.Ersten("card").Ersten("name"); knoten == nil ||
		knoten.Wert != "Feuerball" {
		t.Errorf("%+v", knoten)
	}
}

// **Ein eingebundener Baum behält seine BOM und seinen Namen.**
//
// Die Naht zwischen dem Einbinder (Phase 3) und dem Schreiber (Phase 6): beide
// waren für sich richtig, und zusammen verloren sie drei Bytes. Ein Rundlauf
// über ein Paket mit `include file:` hätte weniger zurückgegeben, als er bekam.
func TestEinEingebundenerBaumBehaeltBOMUndNamen(t *testing.T) {
	quelle := "\uFEFFa: eins\ninclude file: /probe.mse-include/teil\n"

	baum, err := msepaket.Zerteile("style", []byte(quelle))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	eingebunden := msepaket.NeuerEinbinder(nichtsGefunden{}).Einbinden(baum)

	if !eingebunden.MitBOM || eingebunden.Datei != "style" {
		t.Errorf("MitBOM %v, Datei %q", eingebunden.MitBOM, eingebunden.Datei)
	}

	if gebaut := string(msepaket.Schreibe(eingebunden)); !strings.HasPrefix(gebaut, "\uFEFF") {
		t.Errorf("geschrieben ohne BOM: %q", gebaut)
	}
}

// nichtsGefunden findet kein einziges Modul — der Einbinder macht daraus einen
// Vorbehalt, und der Baum bleibt lesbar.
type nichtsGefunden struct{}

func (nichtsGefunden) Lies(msepaket.Paketpfad) ([]byte, error) {
	return nil, errNichtsGefunden
}

var errNichtsGefunden = errors.New("hier ist nichts installiert")
