package msepaket_test

// **`include file:` über fremde Dateien** — QA-06b
// (#1967),
// Fläche 5 aus `docs/FUZZZIELE.md`.
//
// # Warum hier
//
// **Ein Stil bestimmt, welche Dateien gelesen werden** — aus dem eigenen
// Paket, aus fremden, in der gewählten Sprache —, und jede davon kann wieder
// einbinden. Kreis, Raute und Kette haben je einen benannten Test
// (`einbindenbudget_test.go`); was sie nicht sehen, ist eine Mischung, die
// niemand aufgeschrieben hat.
//
// Der Fuzzer schreibt die Wurzel und drei Dateien in zwei Paketen, eine davon
// in zwei Sprachfassungen. Was die Verweise darin nennen, entscheidet er.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger — die Grenzen halten Kreis, Raute und Kette
//   - zweimal dieselbe Eingabe, zwei Ergebnisse oder zwei Befundlisten
//   - **der Eingabebaum ist danach ein anderer** (siehe `knotenliste`: ein
//     Ergebnis, das seine Eingabe verändert, ist an jeder zweiten Stelle eine
//     Überraschung)
//   - ein Befund ohne Meldung
//   - **ein Verweis überlebt das Einbinden** — aufgelöst oder benannt, aber
//     nie stehen gelassen. Verweis heißt, was MSE2 dafür hält: der Schlüssel
//     **in kanonischer Form** `include_file` (`Reader::unknownKey` vergleicht
//     ihn nach `canonical_name_form`, `src/util/io/reader.hpp` Zeile 175 und
//     `reader.cpp` Zeile 217, Fassung f9d9356 vom 2021-07-06), dazu die
//     erschlossene Form `include localized file` und der Verweis im Wert
//   - mehr gelesene Dateien, als die Grenzen erlauben: höchstens
//     [msepaket.GrenzeEinbindungen] Einbindungen, je Sprachverweis zwei
//     Versuche mehr
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzEinbinden -fuzztime 2m

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was eingebunden ist, steht ohne Verweis da — und die Eingabe bleibt,
// wie sie war.**
func FuzzEinbinden(f *testing.F) {
	for _, saat := range [][5]string{
		{"init script:\n\tinclude file: /a.mse-include/a\n", "x: 1\ninclude file: b\n", "y: 2\n", "z: 3\n", ""},
		{"include file: /a.mse-include/a\ninclude file: /a.mse-include/a\n", "include file: /a.mse-include/b\n", "include file: /a.mse-include/a\n", "", "de"},
		{"init script: include file: /b.mse-include/c\n", "", "", "include localized file: c\n", "de"},
		{"include localized file: /b.mse-include/c\n\tdarunter: 1\n", "", "", "q: 1\n", "fr"},
		{"include file: ../a.mse-include/a\ninclude file: /a.mse-include/../b.mse-include/c\n", "", "", "", ""},
		{"include_file: /a.mse-include/a\n", "x: 1\n", "", "", ""},
		{"a:\n\tinclude file: /a.mse-include/a\n\tb: 2\n", strings.Repeat("include file: b\n", 10), strings.Repeat("include file: a\n", 10), "", ""},
		{"", "", "", "", ""},
	} {
		f.Add(saat[0], saat[1], saat[2], saat[3], saat[4])
	}

	f.Fuzz(func(t *testing.T, wurzel, a, b, c, sprache string) {
		eingabe, fehler := msepaket.Zerteile("probe.mse-style/style", []byte(wurzel))
		if fehler != nil {
			return
		}

		// **Ein zweiter Baum aus denselben Bytes** ist der Vergleich: der
		// Zerteiler ist deterministisch (`FuzzZerteile`), der Einbinder soll
		// den ersten nicht anfassen.
		unberuehrt, _ := msepaket.Zerteile("probe.mse-style/style", []byte(wurzel)) //nolint:errcheck // eben gelesen

		quelle := &gezaehlteDateien{dateien: dateien{
			"a.mse-include/a": a, "a.mse-include/b": b,
			"b.mse-include/c_en": c, "b.mse-include/c_de": c + "\nde: ja\n",
		}}

		einbinder := msepaket.NeuerEinbinder(quelle).ImPaket("probe.mse-style").InSprache(sprache)
		ergebnis := einbinder.Einbinden(eingabe)

		if !reflect.DeepEqual(eingabe, unberuehrt) {
			t.Error("der Eingabebaum ist nach dem Einbinden ein anderer")
		}

		zweiter := msepaket.NeuerEinbinder(quelle.dateien).ImPaket("probe.mse-style").InSprache(sprache)
		if nochmal := zweiter.Einbinden(unberuehrt); !reflect.DeepEqual(ergebnis, nochmal) ||
			!reflect.DeepEqual(einbinder.Befunde(), zweiter.Befunde()) {
			t.Errorf("dieselbe Eingabe, zwei Ergebnisse: %d gegen %d Befunde",
				len(einbinder.Befunde()), len(zweiter.Befunde()))
		}

		for _, befund := range einbinder.Befunde() {
			if befund.Meldung == "" {
				t.Error("ein Befund ohne Meldung")
			}
		}

		keinVerweisUebrig(t, ergebnis.Wurzel)

		if obergrenze := 3*msepaket.GrenzeEinbindungen + 3*zaehleVerweise(wurzel, a, b, c); quelle.gelesen > obergrenze {
			t.Errorf("%d Dateien gelesen — über %d", quelle.gelesen, obergrenze)
		}
	})
}

// gezaehlteDateien zählt, wie oft gelesen wird.
type gezaehlteDateien struct {
	dateien dateien
	gelesen int
}

func (g *gezaehlteDateien) Lies(pfad msepaket.Paketpfad) ([]byte, error) {
	g.gelesen++

	return g.dateien.Lies(pfad)
}

// keinVerweisUebrig sucht Verweise im Ergebnis — als Schlüssel in
// kanonischer Form und im Wert.
func keinVerweisUebrig(t *testing.T, knoten []*msepaket.Knoten) {
	t.Helper()

	for _, kind := range knoten {
		if kind.Zeilenart == msepaket.ArtEintrag {
			switch kanonisch := msepaket.Kanonisch(kind.Schluessel); {
			case kanonisch == "include_file" || kanonisch == "include_localized_file":
				t.Errorf("%s:%d: der Verweis `%s: %s` steht noch da",
					kind.Stelle.Datei, kind.Stelle.Zeile, kind.Schluessel, kind.Wert)

			case strings.HasPrefix(strings.TrimSpace(kind.Wert), "include file:") &&
				strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(kind.Wert), "include file:")) != "":
				t.Errorf("%s:%d: der Verweis im Wert `%s` steht noch da",
					kind.Stelle.Datei, kind.Stelle.Zeile, kind.Wert)
			}
		}

		keinVerweisUebrig(t, kind.Kinder)
	}
}

// zaehleVerweise gibt, wie viele Zeilen überhaupt einbinden könnten — für
// die Sprachversuche, die vor der Budgetprüfung liegen.
func zaehleVerweise(texte ...string) int {
	zahl := 0
	for _, text := range texte {
		zahl += strings.Count(text, "include")
	}

	return zahl * msepaket.GrenzeEinbindungen
}
