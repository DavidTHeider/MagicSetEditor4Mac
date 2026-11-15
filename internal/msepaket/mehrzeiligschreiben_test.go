package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein mehrzeiliger Wert wird mehrzeilig geschrieben** —
// #704.
//
// # Der Befund
//
// Gefunden beim Rundlauf über das mitgelieferte Set: von 18 Feldwerten der
// ersten Karte kamen nach „sichern, wieder öffnen" **zwei** zurück.
//
// Die Karte trägt ein `notes` über drei Zeilen. Geschrieben wurde
//
//	notes: Sets: [pmoa];
//
//	Formats: ;
//
// — die zweite und dritte Zeile ohne Einrückung. Eine Zeile ohne Einrückung
// beendet den `card:`-Block, und alles danach las sich als oberste Ebene. Die
// Datei sah aus wie eine Datei und war eine andere.
//
// # Die Form
//
//	} else if (value.empty()) {
//	  // a multiline string
//	  readLine(true);
//	  while (indent >= expected_indent && !input.Eof()) {
//	    previous_value += line.substr(expected_indent);
//
// `src/util/io/reader.cpp`, Zeilen 258–272, Fassung f9d9356 vom 2021-07-06:
// **hinter dem Doppelpunkt steht nichts**, und der Wert folgt eine Ebene
// tiefer eingerückt.
func TestEinMehrzeiligerWertWirdMehrzeiligGeschrieben(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name  string
		wert  string
		zeige string
	}{
		{
			name:  "der Fall aus dem Kanon",
			wert:  "Sets: [pmoa];\n\nFormats: ;",
			zeige: "card:\n\tnotes:\n\t\tSets: [pmoa];\n\t\t\n\t\tFormats: ;\n",
		},
		{
			name:  "zwei Zeilen",
			wert:  "erste\nzweite",
			zeige: "card:\n\tnotes:\n\t\terste\n\t\tzweite\n",
		},
		{
			name:  "ein abschließender Umbruch zählt mit",
			wert:  "erste\n",
			zeige: "card:\n\tnotes:\n\t\terste\n\t\t\n",
		},
		{
			name:  "einzeilig bleibt einzeilig",
			wert:  "kurz",
			zeige: "card:\n\tnotes: kurz\n",
		},

		// **Führender Leerraum steht als Block** (FMT-06.AK-02, #1695) — wie
		// bei MSE2 (`src/util/io/writer.cpp`, Zeile 74, Fassung f9d9356 vom
		// 2021-07-06): hinter dem Doppelpunkt schnitte ihn das Lesen ab.
		{
			name:  "führender Leerraum steht als Block",
			wert:  "  eingerückt",
			zeige: "card:\n\tnotes:\n\t\t  eingerückt\n",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			baum := &msepaket.Baum{Wurzel: []*msepaket.Knoten{{
				Zeilenart:  msepaket.ArtEintrag,
				Schluessel: "card",
				Kinder: []*msepaket.Knoten{{
					Zeilenart:  msepaket.ArtEintrag,
					Schluessel: "notes",
					Wert:       fall.wert,
				}},
			}}}

			if hat := string(msepaket.Schreibe(baum)); hat != fall.zeige {
				t.Errorf("geschrieben\n%q\nerwartet\n%q", hat, fall.zeige)
			}
		})
	}
}

// **Und was geschrieben wurde, liest sich wieder als derselbe Wert.**
//
// Die Zusage, an der es hängt: der Weg hinaus ist die Umkehrung des Weges
// hinein. Ein Schreiber, der eine Form erzeugt, die der Leser anders versteht,
// verliert einen Wert — genau das ist geschehen.
func TestEinMehrzeiligerWertKommtZurueck(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{
		"Sets: [pmoa];\n\nFormats: ;",
		"erste\nzweite\ndritte",
		"einzeilig",
		"  eingerückt",
		" ein Leerzeichen",
	} {
		t.Run(strings.SplitN(wert, "\n", 2)[0], func(t *testing.T) {
			t.Parallel()

			baum := &msepaket.Baum{Wurzel: []*msepaket.Knoten{{
				Zeilenart:  msepaket.ArtEintrag,
				Schluessel: "card",
				Kinder: []*msepaket.Knoten{{
					Zeilenart:  msepaket.ArtEintrag,
					Schluessel: "notes",
					Wert:       wert,
				}},
			}}}

			gelesen, err := msepaket.Zerteile("set", msepaket.Schreibe(baum))
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			karte := gelesen.Ersten("card")
			if karte == nil {
				t.Fatal("kein card-Block")
			}

			notiz := karte.Ersten("notes")
			if notiz == nil {
				t.Fatal("kein notes")
			}

			hat := notiz.Wert
			if hat == "" {
				hat = notiz.Skripttext()
			}

			if hat != wert {
				t.Errorf("gelesen %q, geschrieben war %q", hat, wert)
			}
		})
	}
}
