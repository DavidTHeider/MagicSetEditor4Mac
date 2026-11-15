//go:build fremdpakete

// **Der Fähigkeitsbestand von MSE2 bleibt vollständig** —
// #1291.
//
// # Warum es diesen Wächter gibt
//
// `docs/ROADMAP.md` sagt: *MagicSetEditor4Mac enthält die Fähigkeiten von MSE2
// als echtes Subset und geht darüber hinaus.* Ein Subset lässt sich nur
// behaupten, wenn die **Obermenge** dasteht — und eine Obermenge, die von Hand
// geführt wird, veraltet still.
//
// `docs/MSE2BESTAND.md` führt sie. Dieser Lauf hält sie gegen ihre Quelle.
//
// # Warum hinter `fremdpakete`
//
// Weil er den **fremden** Quelltext braucht: `data/en.mse-locale/locale` aus
// `mse2-quelle`. Der liegt neben dem Repositorium (§11) und nicht darin.
//
//	MSE2_QUELLE=/Users/david/mse2-quelle \
//	  go test -tags fremdpakete ./internal/werkzeuge/ -run Mse2Bestand -v
//
// # Was er hält
//
//   - **Jeder Menüeintrag von MSE2 hat eine Zeile.** Ein neuer Eintrag — oder
//     ein bisher übersehener — macht ihn rot.
//   - **Jede Zeile nennt einen Eintrag, den es gibt.** Ein Tippfehler im
//     Namen fällt damit auf.
//
// # Was er **nicht** hält
//
//   - **Nicht, dass die Einordnung stimmt.** Ob ein ✓ verdient ist, sagt kein
//     Wächter; das liest ein Mensch.
//   - **Nicht die Vollständigkeit über das Menü hinaus.** Kontextmenüs,
//     Panels, Ziehen und Ablegen stehen in keiner der beiden Listen — die
//     Datei sagt das selbst.

package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// menueeintrag findet eine Zeile des Abschnitts `menu:` in MSEs Locale.
//
//	new set:			&New...	Ctrl+N
var menueeintrag = regexp.MustCompile(`(?m)^\t([a-z_ ]+):`)

// bestandszeile findet den Befehlsnamen einer Zeile der Bestandstafel.
var bestandszeile = regexp.MustCompile("(?m)^\\| `([a-z_ ]+)` \\|")

// ueberschriften sind die Einträge, die kein Befehl sind.
//
// **Sie stehen hier und nicht in der Tafel**: eine Zeile „Datei — Überschrift"
// sagte nichts und stünde elfmal da.
var ueberschriften = map[string]bool{
	"file": true, "edit": true, "cards": true, "keywords": true,
	"format": true, "graph": true, "window": true, "help": true,
	"editor": true, "tool": true, "export": true, "symbols": true,
}

func TestDerMse2BestandIstVollstaendig(t *testing.T) {
	t.Parallel()

	quelle := os.Getenv("MSE2_QUELLE")
	if quelle == "" {
		t.Skip("MSE2_QUELLE nicht gesetzt — der Lauf braucht den Quelltext von MSE2")
	}

	locale, err := os.ReadFile( //nolint:gosec // Weg aus der Umgebung, bewusst
		filepath.Join(quelle, "data", "en.mse-locale", "locale"))
	if err != nil {
		t.Fatalf("MSEs Locale lesen: %v", err)
	}

	// **Die Datei kommt mit CRLF** — sie stammt aus einer Windows-Anwendung.
	// Ohne diese Zeile findet der Abschnittsmarker `\nmenu:\n` nichts, und der
	// Lauf meldete „kein Menüeintrag gefunden": richtig gemeldet, und die
	// Ursache lag im Leser (§8 — fremde Daten sind der Normalfall).
	imMenue := menuebefehle(strings.ReplaceAll(string(locale), "\r\n", "\n"))
	if len(imMenue) == 0 {
		t.Fatal("kein Menüeintrag gefunden — der Leser greift daneben")
	}

	inDerTafel := map[string]bool{}
	for _, treffer := range bestandszeile.FindAllStringSubmatch(lies(t, "docs/MSE2BESTAND.md"), -1) {
		inDerTafel[treffer[1]] = true
	}

	for befehl := range imMenue {
		if !inDerTafel[befehl] {
			t.Errorf("MSE2 führt den Menübefehl %q — docs/MSE2BESTAND.md kennt ihn nicht", befehl)
		}
	}

	for befehl := range inDerTafel {
		if !imMenue[befehl] {
			t.Errorf("docs/MSE2BESTAND.md führt %q — MSEs Menü kennt ihn nicht", befehl)
		}
	}
}

// menuebefehle gibt die Befehle aus dem Abschnitt `menu:` — ohne Überschriften.
func menuebefehle(locale string) map[string]bool {
	anfang := strings.Index(locale, "\nmenu:\n")
	if anfang < 0 {
		return nil
	}

	rest := locale[anfang+len("\nmenu:\n"):]

	// **Der Abschnitt endet bei der nächsten Zeile ohne Einrückung.**
	if ende := regexp.MustCompile(`(?m)^[a-z]`).FindStringIndex(rest); ende != nil {
		rest = rest[:ende[0]]
	}

	befehle := map[string]bool{}

	for _, treffer := range menueeintrag.FindAllStringSubmatch(rest, -1) {
		if !ueberschriften[treffer[1]] {
			befehle[treffer[1]] = true
		}
	}

	return befehle
}
