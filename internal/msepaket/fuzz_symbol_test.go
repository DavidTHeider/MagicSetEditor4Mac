package msepaket_test

// **Eine `.mse-symbol` von fremder Hand** — QA-06b
// (#1967),
// Fläche 8 aus `docs/FUZZZIELE.md`: [msepaket.SymbolzeichnungAus].
//
// # Warum der Rundlauf hier die Zusage ist
//
// **Ein Symbol wird gelesen, bearbeitet und zurückgeschrieben** — in ein Set,
// in die Symbolbibliothek. Was der Leser annimmt, schreibt der Schreiber
// wieder hin, und beim nächsten Öffnen liest es derselbe Leser. Ein Symbol,
// das sich öffnen, aber nach dem Speichern nicht wieder öffnen lässt, ist
// verloren, ohne dass jemand gewarnt wurde.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselben Bytes, zwei Ergebnisse
//   - ein Fehler **und** eine Zeichnung; ein Befund ohne Meldung
//   - eine angenommene Zeichnung über einer Grenze ([msepaket.PruefeSymbolumfang])
//   - **Schreiben ist kein Fixpunkt**: was angenommen ist, muss geschrieben
//     wieder gelesen werden — und dann so geschrieben werden wie beim ersten
//     Mal. Verglichen werden die geschriebenen Bytes und nicht die Zahlen:
//     `NaN` ist sich selbst ungleich, und MSE2 liest es wie wir
//     (`swscanf` mit `%lf`)
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzSymbolzeichnung -fuzztime 2m

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was angenommen ist, lässt sich schreiben und wieder lesen.**
func FuzzSymbolzeichnung(f *testing.F) {
	for _, saat := range []string{
		string(msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol())),
		"mse version: 0.3.5\npart:\n\ttype: shape\n\tname: a\n\tcombine: subtract\n" +
			"\tpoint:\n\t\tposition: (0.1,0.2)\n\t\tline_after: curve\n\t\thandle_after: (0.3, 0.4)\n" +
			"\tpoint:\n\t\tposition: (NaN,+Inf)\n\t\thandle_before: (1e400,0)\n\t\tlock: dir\n",
		"part:\n\ttype: symmetry\n\tkind: reflection\n\tcopies: 0400\n\tcenter: (0.5,0.5)\n" +
			"\tpart:\n\t\ttype: group\n\t\tpart:\n\t\t\tname: innen\n\tpart:\n\t\ttype: zauber\n",
		"part:\n\ttype: unbekannt\npart:\n",
		"",
	} {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		zeichnung, befunde, fehler := msepaket.SymbolzeichnungAus("probe.mse-symbol", roh)
		nochmal, nochmalBefunde, nochmalFehler := msepaket.SymbolzeichnungAus("probe.mse-symbol", roh)

		if (fehler == nil) != (nochmalFehler == nil) || !reflect.DeepEqual(befunde, nochmalBefunde) ||
			!bytes.Equal(msepaket.SchreibeSymbolzeichnung(zeichnung), msepaket.SchreibeSymbolzeichnung(nochmal)) {
			t.Fatalf("dieselben Bytes, zwei Ergebnisse: %v / %v", fehler, nochmalFehler)
		}

		for _, befund := range befunde {
			if befund.Meldung == "" {
				t.Error("ein Befund ohne Meldung")
			}
		}

		if fehler != nil {
			if len(zeichnung.Teile) != 0 {
				t.Errorf("abgelehnt **und** %d Teile: %v", len(zeichnung.Teile), fehler)
			}

			return
		}

		if umfang := msepaket.PruefeSymbolumfang(zeichnung); umfang != nil {
			t.Errorf("angenommen und über einer Grenze: %v", umfang)
		}

		geschrieben := msepaket.SchreibeSymbolzeichnung(zeichnung)

		zurueck, _, rueckfehler := msepaket.SymbolzeichnungAus("probe.mse-symbol", geschrieben)
		if rueckfehler != nil {
			t.Fatalf("angenommen, geschrieben (%d Byte) und nicht wieder lesbar: %v", len(geschrieben), rueckfehler)
		}

		if wieder := msepaket.SchreibeSymbolzeichnung(zurueck); !bytes.Equal(wieder, geschrieben) {
			t.Errorf("Schreiben ist kein Fixpunkt:\n%q\n%q", geschrieben, wieder)
		}
	})
}
