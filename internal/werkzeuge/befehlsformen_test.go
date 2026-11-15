// Der Wächter über die Bedienbarkeit der Spielbefehle im Fenster.
//
// # Wofür
//
// **§16 sagt: dieselbe Fähigkeit an jeder Fläche.** Der Wächter daneben
// zählt **Bindungen**, und `PartieBefehl` gibt es an allen vier; sie nimmt
// jede Befehlsart entgegen. Was er nicht sieht: dass an drei Flächen ein
// Mensch die Art **tippen** kann und an der vierten niemand sie
// **anklicken**.
//
// **Gemessen am 2026-09-17**: der Kern kannte achtunddreißig Befehle, das
// Fenster konnte elf davon stellen — darunter drei aus `G1` selbst
// (#1623).
//
// # Was er prüft
//
// **Beide Richtungen.** Jede Art aus `service.Sandkastenbefehle()` und
// `service.Leitungsbefehle()` hat eine Form in
// `frontend/src/spiel/befehlsformen.ts`, und jede Form dort nennt eine Art,
// die es gibt. Eine Form ohne Befehl wäre ein Knopf, den der Kern
// abwiese.
//
// # Was er nicht prüft
//
// **Ob die Form die richtige Nutzlast einsammelt.** Dass `objekt.zaehler`
// nach „art" und „menge" fragt und nicht nach „an", steht in den Läufen des
// Fensters und im Kern. Dieser Wächter hält die **Liste** vollständig, nicht
// ihren Inhalt.
package werkzeuge_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// formenquelle ist die Tabelle, die das Fenster liest.
const formenquelle = "../../frontend/src/spiel/befehlsformen.ts"

// formenart findet eine Befehlsart in der Tabelle.
//
// **Zwei Gestalten, weil es zwei gibt**: `art: 'objekt.tappen'` in einem
// Verbund und `nurObjekt('objekt.tappen')` für die Kurzform.
var formenart = regexp.MustCompile(
	`(?:art: |nurObjekt\()'([a-z]+\.[a-z]+)'`)

// **Jeder Spielbefehl lässt sich im Fenster stellen.**
func TestJederSpielbefehlHatEineFormImFenster(t *testing.T) {
	t.Parallel()

	gefunden := formenAusDatei(t)

	for _, art := range spielbefehle() {
		if !gefunden[art] {
			t.Errorf("%q hat keine Form in %s — im Fenster ist er nicht "+
				"zu stellen (§16)", art, formenquelle)
		}
	}
}

// **Und keine Form nennt einen Befehl, den es nicht gibt.**
//
// **Die Gegenrichtung, und sie ist die unauffälligere**: ein Knopf für eine
// Art, die der Kern nicht kennt, sieht aus wie eine Fähigkeit und endet in
// „ist kein Befehl, den der Sandkasten kennt".
func TestKeineFormNenntEinenBefehlDenEsNichtGibt(t *testing.T) {
	t.Parallel()

	echte := map[string]bool{}
	for _, art := range spielbefehle() {
		echte[art] = true
	}

	for art := range formenAusDatei(t) {
		if !echte[art] {
			t.Errorf("%q steht in %s und ist kein Spielbefehl",
				art, formenquelle)
		}
	}
}

// spielbefehle sind alle Arten, die der Kern annimmt — die des Sandkastens
// und die der Leitung (die Rücknahme, P8, #2376).
func spielbefehle() []string {
	return append(service.Sandkastenbefehle(), service.Leitungsbefehle()...)
}

// formenAusDatei liest die Arten aus der Tabelle.
func formenAusDatei(t *testing.T) map[string]bool {
	t.Helper()

	roh, fehler := os.ReadFile(formenquelle)
	if fehler != nil {
		t.Fatalf("%s: %v", formenquelle, fehler)
	}

	gefunden := map[string]bool{}

	for _, zeile := range strings.Split(string(roh), "\n") {
		// **Kommentare zählen nicht.** Eine Art im Fließtext ist eine
		// Erklärung und kein Knopf.
		if strings.HasPrefix(strings.TrimSpace(zeile), "//") {
			continue
		}

		for _, treffer := range formenart.FindAllStringSubmatch(zeile, -1) {
			gefunden[treffer[1]] = true
		}
	}

	return gefunden
}
