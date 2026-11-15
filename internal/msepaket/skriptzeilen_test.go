// **Ein Skriptrumpf samt der Stelle jeder seiner Zeilen** — Werkstück B von
// Phase 4.9
// (#1324).
//
// # Warum die Stellen dazugehören
//
// `Skripttext` gibt den Rumpf ohne die Einrückung des Paketformats — richtig
// so, denn in MSE-Script bedeutet sie nichts. Wer aber einen Fehler **in der
// Datei** anstreichen will, braucht sie zurück: Zeile 2 des Rumpfes ist Zeile
// 47 der Datei, und ihre Spalte 1 ist dort Spalte 3.
package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestSkriptzeilenNennenIhreStelleInDerDatei(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		erwartet []msepaket.Skriptzeile
	}{
		// **Der einzeilige Fall ist keiner für sich.** Bis hierher gab
		// `Skripttext` dafür die leere Zeichenkette, und jeder Aufrufer hatte
		// seinen eigenen Rückfall auf `Wert` — zwei Wege für eine Frage.
		{
			name: "ein Wert in derselben Zeile", quelle: "script: 1 + 2\n",
			erwartet: []msepaket.Skriptzeile{{Text: "1 + 2", Zeile: 1, Spalte: 9}},
		},
		{
			name: "ein Rumpf über zwei Zeilen", quelle: "script:\n\t1 +\n\t2\n",
			erwartet: []msepaket.Skriptzeile{
				{Text: "1 +", Zeile: 2, Spalte: 2},
				{Text: "2", Zeile: 3, Spalte: 2},
			},
		},
		{
			name:   "ein eingerückter Rumpf",
			quelle: "style:\n\tscript:\n\t\tf(1,\n\t\t2\n",
			erwartet: []msepaket.Skriptzeile{
				{Text: "f(1,", Zeile: 3, Spalte: 3},
				{Text: "2", Zeile: 4, Spalte: 3},
			},
		},
		// **Ein Eintrag im Rumpf zählt mit.** `include file:` steht in einem
		// `init script:` und ist dort eine Zeile wie jede andere —
		// `Skripttext` nimmt sie, und diese Liste muss dieselbe Reihenfolge
		// haben, sonst zeigt jede Stelle danach auf die falsche Zeile.
		{
			name:   "ein Eintrag mitten im Rumpf",
			quelle: "init script:\n\ta := 1\n\tinclude file: x\n\tb := 2\n",
			erwartet: []msepaket.Skriptzeile{
				{Text: "a := 1", Zeile: 2, Spalte: 2},
				{Text: "include file: x", Zeile: 3, Spalte: 2},
				{Text: "b := 2", Zeile: 4, Spalte: 2},
			},
		},
		{
			name: "kein Rumpf", quelle: "script:\n",
			erwartet: nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			baum, err := msepaket.Zerteile("probe", []byte(fall.quelle))
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			knoten := ersterSkriptknoten(baum.Wurzel)
			if knoten == nil {
				t.Fatal("kein script-Knoten gefunden")
			}

			zeilen := knoten.Skriptzeilen()

			if len(zeilen) != len(fall.erwartet) {
				t.Fatalf("%d Zeilen %v, erwartet %d %v",
					len(zeilen), zeilen, len(fall.erwartet), fall.erwartet)
			}

			for nummer, zeile := range zeilen {
				if zeile != fall.erwartet[nummer] {
					t.Errorf("Zeile %d ist %v, erwartet %v",
						nummer, zeile, fall.erwartet[nummer])
				}
			}
		})
	}
}

// Und `Skripttext` bleibt, was er war — für beide Formen.
//
// **Die Gegenprobe zur Änderung.** `Skripttext` gab bisher für einen Wert in
// derselben Zeile die leere Zeichenkette; wer ihn auf die neue Liste stellt,
// ändert das. Genau darum steht es hier: als Zusage und nicht als Nebenwirkung.
func TestSkripttextGibtDenRumpfBeiderFormen(t *testing.T) {
	faelle := map[string]string{
		"script: 1 + 2\n":       "1 + 2",
		"script:\n\t1 +\n\t2\n": "1 +\n2",
		"style:\n\tscript: a\n": "a",
		"script:\n":             "",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			baum, err := msepaket.Zerteile("probe", []byte(quelle))
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			knoten := ersterSkriptknoten(baum.Wurzel)
			if knoten == nil {
				t.Fatal("kein script-Knoten gefunden")
			}

			if text := knoten.Skripttext(); text != erwartet {
				t.Errorf("Skripttext %q, erwartet %q", text, erwartet)
			}
		})
	}
}

// ersterSkriptknoten sucht den ersten `script`-Eintrag im Baum.
func ersterSkriptknoten(knoten []*msepaket.Knoten) *msepaket.Knoten {
	for _, eins := range knoten {
		// **`Kanonisch` ersetzt Leerzeichen durch `_`** — `init script` heißt
		// hier `init_script`, und ein Vergleich mit dem Leerzeichen träfe nie.
		if msepaket.Kanonisch(eins.Schluessel) == "script" ||
			msepaket.Kanonisch(eins.Schluessel) == "init_script" {
			return eins
		}

		if tiefer := ersterSkriptknoten(eins.Kinder); tiefer != nil {
			return tiefer
		}
	}

	return nil
}

// Und die Einfärbung nimmt den ganzen Rumpf als **einen** Wert.
//
// # Die Gegenprobe zu der Änderung
//
// Bis zum 2026-09-02 kam ein eingerückter Rumpf als Folge von
// Fortsetzungszeilen an und wurde als gewöhnlicher Text gefärbt — **2.240**
// der **5.553** Skriptwerte des vollen Packs. Dieser Test hält beides fest:
// dass es jetzt **ein** Skriptmerkmal ist, und dass die Zeilen darunter kein
// zweites Mal als Wert dastehen.
func TestEineMehrzeiligeSkriptzeileWirdEinSkriptwert(t *testing.T) {
	quelle := "style:\n\tscript:\n\t\tf(1,\n\t\t2\n\tname: x\n"

	var skriptwerte, werte []msepaket.Farbmerkmal

	for _, merkmal := range msepaket.Einfaerben(quelle) {
		switch merkmal.Art {
		case msepaket.FarbeSkriptwert:
			skriptwerte = append(skriptwerte, merkmal)

		case msepaket.FarbeWert:
			werte = append(werte, merkmal)
		}
	}

	if len(skriptwerte) != 1 {
		t.Fatalf("%d Skriptwerte, erwartet einer: %v", len(skriptwerte), skriptwerte)
	}

	if skriptwerte[0].Text != "f(1,\n2" {
		t.Errorf("Text %q, erwartet %q", skriptwerte[0].Text, "f(1,\n2")
	}

	if len(skriptwerte[0].Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei: %v", len(skriptwerte[0].Zeilen), skriptwerte[0].Zeilen)
	}

	if skriptwerte[0].Zeilen[1] != (msepaket.Skriptzeile{Text: "2", Zeile: 4, Spalte: 3}) {
		t.Errorf("zweite Zeile %v", skriptwerte[0].Zeilen[1])
	}

	// **`name: x` bleibt ein Wert, der Rumpf wird keiner.** Ohne diese
	// Zusicherung wäre der Test auch grün, wenn gar nichts mehr gefärbt würde.
	if len(werte) != 1 || werte[0].Text != "x" {
		t.Errorf("Werte %v, erwartet allein `x`", werte)
	}
}

// Und ein Skriptschlüssel ohne Rumpf bleibt ein Skriptwert der Breite null.
//
// **Der Zweig, den sonst kein Lauf betritt.** `script:` allein am Ende einer
// Datei kommt vor — beim Tippen, und in einer Vorlage, deren Verfasser den
// Block noch schreibt.
func TestEinSkriptschluesselOhneRumpf(t *testing.T) {
	var gefunden []msepaket.Farbmerkmal

	for _, merkmal := range msepaket.Einfaerben("script:\n") {
		if merkmal.Art == msepaket.FarbeSkriptwert {
			gefunden = append(gefunden, merkmal)
		}
	}

	if len(gefunden) != 1 {
		t.Fatalf("%d Skriptwerte, erwartet einer: %v", len(gefunden), gefunden)
	}

	if gefunden[0].Breite != 0 || gefunden[0].Zeilen != nil {
		t.Errorf("Merkmal %v, erwartet Breite null ohne Zeilen", gefunden[0])
	}
}
