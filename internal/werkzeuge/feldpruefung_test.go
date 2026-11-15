// Der Wächter über die Liste der freien Eingaben.
//
// # Wofür
//
// **Eine neue Eingabe soll die Entscheidung erzwingen, statt still in der
// Menge zu verschwinden** — Werkstück F von Phase 4.10
// (#1325).
// `AGENTS.md` §6 verlangt, dass ein Fehler am verursachenden Feld erscheint;
// gemessen am 2026-09-03 gab es **66 freie Eingaben in 44 Dateien** und
// **zwei** Dateien, die `aria-invalid` setzen.
//
// `docs/FELDPRUEFUNG.md` führt je Datei, was ihre Eingaben erwarten und ob
// hier geprüft wird — **oder ausdrücklich, dass nicht geprüft wird, und
// warum**. Dieselbe Bauform wie `docs/SCHREIBWEGE.md`.
//
// # Was er nicht hält
//
// **Ob die Prüfung stimmt.** Er zählt Eingaben je Datei und hält die Zahl
// gegen die Tabelle. Eine Datei, die `aria-invalid` setzt und die falsche
// Regel prüft, käme durch — dafür stehen die Tests neben den Bauteilen.
//
// **Und nicht, ob der Satz am Feld gut ist.** §6 verlangt, dass er sagt, was
// falsch ist und was zu tun ist; das prüft ein Mensch.
package werkzeuge_test

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// freieEingabe findet den Anfang eines Eingabefelds.
var freieEingabe = regexp.MustCompile(`<(input|textarea)\b`)

// eingabeart liest `type="…"`.
var eingabeart = regexp.MustCompile(`type="(\w+)"`)

// gebundeneArten nehmen nur gültige Werte an und zählen deshalb nicht mit.
var gebundeneArten = map[string]bool{
	"checkbox": true, "radio": true, "date": true,
	"color": true, "file": true, "range": true,
}

// tabellenzeile liest „| `seiten/Suche.tsx` | 1 | search | **ungeprüft** | … |".
//
// **`\p{L}` und nicht `\w`.** Gos `\w` ist ASCII; zwei der drei Marken tragen
// ein „ü", und der erste Anlauf las **12 von 44** Zeilen. Er wäre nicht
// stillschweigend durchgelaufen — die Zahl unten fängt einen leeren Leser —,
// aber er hätte 32 Dateien als „steht nicht in der Datei" gemeldet, wo sie
// dastanden.
var tabellenzeile = regexp.MustCompile(
	"^\\| `([^`]+\\.tsx)` \\| (\\d+) \\|[^|]*\\| \\*\\*(\\p{L}+)\\*\\* \\|")

// zulaessigeMarken sind die drei Klassen aus der Datei.
var zulaessigeMarken = map[string]bool{
	"geprüft": true, "offen": true, "ungeprüft": true,
}

func TestJedeFreieEingabeStehtInDerFeldpruefung(t *testing.T) {
	t.Parallel()

	gefuehrt := feldpruefungLesen(t)
	gezaehlt := freieEingabenZaehlen(t)

	for datei, zahl := range gezaehlt {
		steht, ist := gefuehrt[datei]
		if !ist {
			t.Errorf("%s trägt %d freie Eingaben und steht nicht in "+
				"docs/FELDPRUEFUNG.md — geprüft oder ausdrücklich ungeprüft "+
				"eintragen (AGENTS.md §6)", datei, zahl)

			continue
		}

		if steht != zahl {
			t.Errorf("docs/FELDPRUEFUNG.md führt für %s %d Eingaben, gezählt "+
				"sind %d", datei, steht, zahl)
		}
	}

	for datei := range gefuehrt {
		if _, ist := gezaehlt[datei]; !ist {
			t.Errorf("docs/FELDPRUEFUNG.md führt %s — die Datei hat keine "+
				"freie Eingabe mehr", datei)
		}
	}
}

// feldpruefungLesen gibt je Datei die geführte Zahl.
func feldpruefungLesen(t *testing.T) map[string]int {
	t.Helper()

	inhalt := gelesen(t, filepath.Join("..", "..", "docs", "FELDPRUEFUNG.md"))

	gefuehrt := map[string]int{}

	for _, zeile := range strings.Split(inhalt, "\n") {
		treffer := tabellenzeile.FindStringSubmatch(zeile)
		if treffer == nil {
			continue
		}

		if !zulaessigeMarken[treffer[3]] {
			t.Errorf("%s trägt die Marke %q — es gibt nur geprüft, offen "+
				"und ungeprüft", treffer[1], treffer[3])
		}

		zahl, err := strconv.Atoi(treffer[2])
		if err != nil {
			t.Fatalf("%s: %q ist keine Zahl", treffer[1], treffer[2])
		}

		gefuehrt["frontend/src/"+treffer[1]] = zahl
	}

	if len(gefuehrt) < 30 {
		t.Fatalf("nur %d Zeilen gelesen — der Leser greift ins Leere",
			len(gefuehrt))
	}

	return gefuehrt
}

// freieEingabenZaehlen zählt je Datei die Eingaben, die freien Text nehmen.
func freieEingabenZaehlen(t *testing.T) map[string]int {
	t.Helper()

	gezaehlt := map[string]int{}
	wurzel := filepath.Join("..", "..", "frontend", "src")

	err := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".tsx") ||
			strings.Contains(pfad, ".test.") || strings.Contains(pfad, "/test/") {
			return nil
		}

		zahl := freieEingabenIn(gelesen(t, pfad))
		if zahl == 0 {
			return nil
		}

		schluessel := filepath.ToSlash(
			strings.TrimPrefix(filepath.ToSlash(pfad), "../../"))
		gezaehlt[schluessel] = zahl

		return nil
	})
	if err != nil {
		t.Fatalf("%s durchgehen: %v", wurzel, err)
	}

	if len(gezaehlt) < 30 {
		t.Fatalf("nur %d Dateien mit freier Eingabe gefunden — der Leser "+
			"greift ins Leere", len(gezaehlt))
	}

	return gezaehlt
}

// ohneKommentar streicht Zeilen- und Blockkommentare aus TSX.
//
// **Weil ein Kommentar über ein Feld sprechen darf.** Der erste Anlauf zählte
// rohen Text und fand in `Bildereinstellungen.tsx` zwei Eingaben, wo eine
// steht — die zweite war das Wort `<input type="number">` in einem
// Kommentar, der erklärt, was dort **früher** stand. Ein Wächter, den man
// nicht erklären darf, ohne ihn rot zu machen, erzieht zum Schweigen.
//
// **Grob und mit Absicht**: eine Zeichenkette, die `//` enthält, verliert
// hier ihren Rest. Das kostet nichts — gesucht wird `<input`, und das steht
// in keiner Zeichenkette dieses Baums.
func ohneKommentar(quelle string) string {
	ohneBloecke := blockkommentar.ReplaceAllString(quelle, " ")

	zeilen := strings.Split(ohneBloecke, "\n")
	for i, zeile := range zeilen {
		if stelle := strings.Index(zeile, "//"); stelle >= 0 {
			zeilen[i] = zeile[:stelle]
		}
	}

	return strings.Join(zeilen, "\n")
}

// blockkommentar findet `/* … */` — und damit auch `{/* … */}` aus JSX.
var blockkommentar = regexp.MustCompile(`(?s)/\*.*?\*/`)

// freieEingabenIn zählt die Felder einer Datei, die freien Text nehmen.
func freieEingabenIn(roh string) int {
	quelle := ohneKommentar(roh)

	var zahl int

	for _, stelle := range freieEingabe.FindAllStringIndex(quelle, -1) {
		schluss := strings.Index(quelle[stelle[0]:], ">")
		if schluss < 0 {
			continue
		}

		block := quelle[stelle[0] : stelle[0]+schluss+1]

		art := "text"
		if treffer := eingabeart.FindStringSubmatch(block); treffer != nil {
			art = treffer[1]
		}

		if gebundeneArten[art] {
			continue
		}

		zahl++
	}

	return zahl
}
