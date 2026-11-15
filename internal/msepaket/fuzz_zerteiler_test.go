package msepaket_test

// **Der Zerteiler gegen böswillige und halb geschriebene Dateien** —
// #1289.
//
// # Warum hier zuerst
//
// `msepaket` liest, was von außen kommt: Dateien von 2007, halb geschrieben,
// in Kodierungen, die niemand mehr benutzt. §8 nennt das den **Normalfall**
// und nicht den Ausnahmefall — ein Set, von dem 300 Karten lesbar sind und
// zwei nicht, wird mit 300 Karten geöffnet.
//
// Bis zum 2026-08-29 gab es dazu **kein einziges** Fuzzziel im Projekt
// (`grep -rn "^func Fuzz[A-Z]" --include='*_test.go'` → null Treffer).
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
// Aus dem Issue, wörtlich:
//
//   - Panic
//   - Hänger
//   - Laufzeit oder Speicher überproportional zur Eingabe
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - Erfolg gemeldet, Zustand strukturell ungültig
//   - **Rundlauf verliert getragene Information**
//
// Die letzten drei sind hier zugesichert; die ersten beiden hält `go test
// -fuzz` selbst. „Überproportional" ist **nicht** zugesichert, und das steht
// hier, damit niemand es für mehr hält: eine Zeitmessung im Fuzzlauf wäre eine
// Aussage über den Rechner.
//
// # Warum der Rundlauf die schärfste Zusage ist
//
// Weil er die **Zusage des Produkts** ist: ein zurückgeschriebenes Set muss in
// MSE2 wieder aufgehen (`docs/ROADMAP.md`, die vier Risiken). Ein Zerteiler,
// der nicht abstürzt, aber Zeilen verliert, wäre schlimmer als einer, der
// abbricht — der Verlust fiele erst dem auf, der die Datei später öffnet.
//
// # Wie er gefahren wird
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzZerteile -fuzztime 2m
//
// **In `make ci` läuft nur der Saatkorpus.** `go test` fährt jedes `FuzzX` mit
// seinen Seeds als gewöhnlichen Lauf; das eigentliche Fuzzen braucht Zeit und
// gehört auf die Liste vor der Freigabe.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// saatzeilen sind die Eingaben, mit denen der Fuzzer beginnt.
//
// **Echte Gestalten und keine erfundenen.** Sie stammen aus `testdaten/` und
// aus `docs/FORMATE.md`: die Stückliste eines Sets, ein mehrzeiliges Skript,
// ein unverstandener Block, ein Kommentar, eine BOM, ein Tabulator gegen
// Leerzeichen. Was der Fuzzer daraus macht, ist seine Sache — womit er
// anfängt, ist unsere.
var saatzeilen = []string{
	"",
	"\ufeffmse version: 2.0.0\ngame: probe\n",
	"card:\n\tname: Feuerball\n\tfarbe: rot\n",
	"set info:\n\ttitle: Probe\ncard:\n\tname: A\ncard:\n\tname: B\n",
	"unverstanden:\n\tdrin: wert\n\n\t# Kommentar im Block\n",
	"mana_sort := {\n\tif a then b\n\telse sort_text@(order: \"WUBRG\")()\n}\n",
	"a:\n\tb:\n\t\tc: tief\n",
	"schluessel: wert mit : Doppelpunkt\n",
	"# nur ein Kommentar\n",
	"\t\tzu tief eingerückt\n",
	"card:\r\n\tname: mit CRLF\r\n",
	"a: \x00 mit Nullbyte\n",
	"a: \xff\xfe kaputte Kodierung\n",
}

// **Der Zerteiler bricht nicht ab, und was er liest, schreibt er wieder.**
func FuzzZerteile(f *testing.F) {
	for _, saat := range saatzeilen {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		baum, fehler := msepaket.Zerteile("probe", roh)

		// **Ein Fehler ist kein Fund.** Fremde Daten dürfen unlesbar sein;
		// was nicht sein darf, ist ein Absturz oder ein halber Stand.
		if fehler != nil {
			if baum != nil {
				t.Errorf("abgelehnt **und** ein Baum zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein** (Nachlauf über Phase 4.2).
			// Ein Zerteiler, der abwechselnd annimmt und ablehnt oder zwei
			// Meldungen gibt, wäre sonst unsichtbar — und §8 verlangt für eine
			// Ablehnung die **Stelle**, die damit auch wandern könnte.
			nochmal, zweiter := msepaket.Zerteile("probe", roh)
			switch {
			case zweiter == nil:
				t.Errorf("derselbe Inhalt wurde erst abgelehnt (%v) und dann "+
					"angenommen: %+v", fehler, nochmal)

			case zweiter.Error() != fehler.Error():
				t.Errorf("derselbe Inhalt gab zwei Meldungen: %q gegen %q",
					fehler, zweiter)
			}

			return
		}

		if baum == nil {
			t.Fatal("kein Fehler und kein Baum — das ist keine Antwort")
		}

		// **Zweimal dieselbe Eingabe, dasselbe Ergebnis.**
		nochmal, nochmalFehler := msepaket.Zerteile("probe", roh)
		if nochmalFehler != nil {
			t.Fatalf("derselbe Inhalt wurde beim zweiten Mal abgelehnt: %v",
				nochmalFehler)
		}

		if !reflect.DeepEqual(baum, nochmal) {
			t.Fatal("derselbe Inhalt gab zwei verschiedene Bäume")
		}

		rundlaufPruefen(t, baum)
	})
}

// rundlaufPruefen schreibt den Baum und liest ihn wieder.
//
// **Die Zusage ist der zweite Baum und nicht die Bytefolge.** Ein Schreiber
// darf einen Tabulator setzen, wo ein Leerzeichen stand, und eine fehlende
// Endzeile ergänzen; was er **nicht** darf, ist einen Knoten verlieren oder
// einen Wert ändern. Deshalb wird nach dem Schreiben erneut zerteilt und die
// Bäume verglichen — nicht die Texte.
func rundlaufPruefen(t *testing.T, baum *msepaket.Baum) {
	t.Helper()

	geschrieben := msepaket.Schreibe(baum)

	zurueck, fehler := msepaket.Zerteile("probe", geschrieben)
	if fehler != nil {
		t.Fatalf("das Geschriebene ließ sich nicht wieder lesen: %v\n%q",
			fehler, kurzfassung(geschrieben))
	}

	if !reflect.DeepEqual(baum.Wurzel, zurueck.Wurzel) {
		t.Errorf("der Rundlauf hat den Baum verändert\ngeschrieben: %q",
			kurzfassung(geschrieben))
	}

	if baum.MitBOM != zurueck.MitBOM {
		t.Errorf("die BOM ging im Rundlauf verloren: %v → %v",
			baum.MitBOM, zurueck.MitBOM)
	}
}

// kurzfassung schneidet eine Ausgabe auf eine lesbare Länge.
func kurzfassung(roh []byte) string {
	const grenze = 400

	text := string(roh)
	if len(text) <= grenze {
		return text
	}

	return text[:grenze] + "… (" + strings.TrimSpace(text[grenze:grenze]) + "gekürzt)"
}
