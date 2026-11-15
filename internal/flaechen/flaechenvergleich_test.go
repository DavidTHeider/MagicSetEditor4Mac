// **Die Vorrichtungen für den Flächenvergleich** — wie geprüft wird, dass
// drei Flächen dasselbe sagen.
//
// **Aus `setauskunft_test.go` herausgelöst**, als es über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort stehen die **Läufe**, hier die
// **Maßstäbe** — und die gelten für jede Datei dieses Pakets, nicht nur für
// die Setauskunft.

package flaechen_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// fuerAlleDrei verlangt, dass jede Fläche denselben Kern nennt.
//
// # Was er belegt — und was nicht
//
// **Er belegt, dass jede Fläche antwortet und dass dieselbe Angabe darin
// steht.** Die drei Gestalten sind verschieden — Tabelle, JSON-Liste,
// JSON-Verbund —, und das dürfen sie sein: gleich ist die **Auskunft**, nicht
// die Ausgabe.
//
// **Er belegt nicht die Gleichheit der ganzen Antwort.** Das hat der
// Zweitblick über Phase 3.8 als Befund 4 angestrichen, und er hat recht: wo
// zwei Flächen dieselbe **Struktur** liefern, gehört sie verglichen und nicht
// nach einem Wort durchsucht. Dafür steht [gleicheAuskunft] daneben — dieser
// hier bleibt für die Fälle, in denen eine der drei nur Text gibt.
//
// **Wo er benutzt wird, ist das eine Entscheidung** und kein Rückfall: `mse
// stile` schreibt eine Tabelle mit einem Stern am gewählten Eintrag, und die
// mit einem JSON-Verbund zu vergleichen hieße, eine der beiden Gestalten zur
// richtigen zu erklären.
func fuerAlleDrei(t *testing.T, was, kern string, cli, mcp, rest func() string) {
	t.Helper()

	fuerAlleDreiMit(t, was, [3]string{kern, kern, kern}, cli, mcp, rest)
}

// fuerAlleDreiMit verlangt je Fläche einen eigenen Kern.
//
// # Wozu drei statt einem
//
// **Weil die Gestalt verschieden sein darf und der Kern damit auch.** Die
// Auszählung schreibt an der Kommandozeile `Karten\t3` und im JSON
// `"karten":3` — es ist dieselbe Auskunft in zwei Schreibweisen, und ein
// gemeinsamer Text dafür wäre die bloße Ziffer `3`.
//
// **Und die steht auch in `13`.** Genau das hat der Nachlauf über Phase 3.8
// angestrichen (Befund 4,
// #1253):
// eine Auszählung, die drei Karten zusagt, war auch bei dreizehn grün.
//
// **Die Reihenfolge ist Kommandozeile, MCP, Dienst** — dieselbe wie überall in
// dieser Datei.
func fuerAlleDreiMit(t *testing.T, was string, kerne [3]string, cli, mcp, rest func() string) {
	t.Helper()

	for stelle, kern := range kerne {
		if len(kern) < 2 {
			t.Fatalf("der Kern %q ist zu kurz, um etwas zu unterscheiden — "+
				"eine Ziffer steht in jeder größeren Zahl", kern)
		}

		_ = stelle
	}

	for stelle, hole := range [3]func() string{cli, mcp, rest} {
		name := [3]string{"Kommandozeile", "MCP", "REST"}[stelle]

		if antwort := hole(); !strings.Contains(antwort, kerne[stelle]) {
			t.Errorf("%s nennt bei %q nicht %q:\n%s",
				name, was, kerne[stelle], antwort)
		}
	}
}

// listenlaenge liest die Länge der **benannten** Liste in einer Antwort.
//
// **Aus zwei Gestalten**: `{"<feld>":[…]}` und eine nackte Liste. Beides ist
// erlaubt; was hier zählt, ist die **Zahl** der Einträge.
//
// **Und eine fehlende Liste ist ein Fehler und keine Null.** Sonst wäre der
// Lauf darüber auch dann grün, wenn die Antwort das Feld gar nicht trägt —
// genau der Fall, gegen den es ihn gibt.
//
// # Warum das Feld genannt wird
//
// Der erste Anlauf ging über **alle** Werte des Verbunds und nahm die erste
// Liste. `{"sets":"falsch","beifang":[]}` bestand damit, obwohl das zugesagte
// Feld gar keine Liste ist — und bei zwei Listen entschied die Laune der
// Map-Iteration, welche gezählt wird. Angestrichen vom Nachlauf über Phase
// 3.8, Befund 5
// (#1253).
func listenlaenge(t *testing.T, roh, feld string) int {
	t.Helper()

	var nackt []any
	if err := json.Unmarshal([]byte(roh), &nackt); err == nil {
		return len(nackt)
	}

	var verbund map[string]any
	if err := json.Unmarshal([]byte(roh), &verbund); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	wert, gefuehrt := verbund[feld]
	if !gefuehrt {
		t.Fatalf("die Antwort führt kein Feld %q:\n%s", feld, roh)
	}

	liste, istListe := wert.([]any)
	if !istListe {
		t.Fatalf("das Feld %q ist keine Liste, sondern %T:\n%s", feld, wert, roh)
	}

	return len(liste)
}

// gleicheAuskunft verlangt, dass zwei Flächen **denselben Verbund** geben.
//
// # Warum er neben [fuerAlleDrei] steht
//
// Weil „enthält das Wort" und „sagt dasselbe" zwei verschiedene Zusagen sind.
// Beim Hineinsehen in eine Sicherung geben MCP und der Dienst denselben
// Verbund — Zeitpunkt, Schemastand, Sets, Karten, Bilder —, und ein Lauf, der
// nur `"sets": 1` sucht, ließe eine Fläche mit `"karten": 0` durch.
//
// **Verglichen wird der gelesene Wert und nicht der Text**: Einrückung und
// Feldreihenfolge sind Gestalt.
func gleicheAuskunft(t *testing.T, was, erste, zweite string) {
	t.Helper()

	var eine, andere any

	if err := json.Unmarshal([]byte(erste), &eine); err != nil {
		t.Fatalf("%s: die erste Antwort ist kein JSON: %v\n%s", was, err, erste)
	}

	if err := json.Unmarshal([]byte(zweite), &andere); err != nil {
		t.Fatalf("%s: die zweite Antwort ist kein JSON: %v\n%s", was, err, zweite)
	}

	if !reflect.DeepEqual(eine, andere) {
		t.Errorf("%s: die beiden Flächen sagen Verschiedenes:\n%s\n%s", was, erste, zweite)
	}
}
