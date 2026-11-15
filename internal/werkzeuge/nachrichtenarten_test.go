// **Jede Nachrichtenart des Handels hat einen Sender** — SEC04
// (#2357),
// Befund Z-02: die Erreichbarkeit der Arten aus `domain/handelsumschlag.go`.
//
// # Was er prüft
//
// Je Art `Nachricht…`: **ein Sender** — sie steht in
// `domain.NachrichtenartFuer` (der Stand eines Handels wählt seine Art) oder
// außerhalb von `internal/domain` in Code, der kein Test ist. Eine Art ohne
// Sender steht unten mit Grund, und eine, die wieder einen hat, muss aus der
// Liste.
//
// **Und einen Empfänger** — P11
// (#2379):
// eine Art, die einen **Zustand** trägt (die Wahl in
// `handelsnachrichtenwahl.go`), empfängt die Post, ohne nach ihr zu
// verzweigen — sie übernimmt den Zustand aus dem Angebot. **Jede andere** muss
// außerhalb von `internal/domain` mit `== domain.Nachricht…` oder
// `!= domain.Nachricht…` verglichen werden, also eine eigene Behandlung haben,
// oder unten mit Grund stehen.
//
// # Was er nicht prüft
//
//   - **Ob die Behandlung das Richtige tut** — er sieht den Vergleich, nicht
//     den Zweig dahinter.
//   - **Ob gesendet ankommt** — das halten die Läufe des Handels
//     (`internal/flaechen/handels*_test.go`) und für die Sicht
//     `service/handelssicht_test.go`.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ohneSender sind Arten im Wortschatz, die nichts sendet — je mit Grund.
var ohneSender = map[string]string{
	"NachrichtHallo": "im Wortschatz seit T5 (#1636, `76b4e65fe`), **von keiner Fassung je gesendet** " +
		"(`git log -S domain.NachrichtHallo` über `internal` findet nichts); der Prüfer nimmt sie an. " +
		"Sie zu streichen änderte, was das Drahtformat annimmt — benannt in SEC04, nicht entfernt",
}

// ohneEmpfaenger sind Arten ohne Zustand, die nichts eigens behandelt — je
// mit Grund.
var ohneEmpfaenger = map[string]string{
	"NachrichtHallo": "dieselbe Lage wie in ohneSender: nie gesendet, also nie empfangen; ein " +
		"Prüfer, der sie annimmt, ist kein Empfänger",
}

func TestJedeNachrichtenartHatEinenSender(t *testing.T) {
	roh, fehler := os.ReadFile("../domain/handelsumschlag.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	arten := regexp.MustCompile(`(?m)^\s*(Nachricht[A-Z]\w*)\s+Handelsnachrichtenart\s*=`).FindAllStringSubmatch(string(roh), -1)
	if len(arten) < 10 {
		t.Fatalf("nur %d Arten gefunden — die Gestalt der Datei hat sich geändert", len(arten))
	}

	wahl, fehler := os.ReadFile("../domain/handelsnachrichtenwahl.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	alles := string(wahl) + "\n" + quelltexteAusserhalbDerDomaene(t)

	for _, art := range arten {
		name := art[1]
		gesendet := regexp.MustCompile(`return ` + name + `, true|\bdomain\.` + name + `\b`).MatchString(alles)
		_, benannt := ohneSender[name]

		switch {
		case !gesendet && !benannt:
			t.Errorf("%s: nichts sendet sie, und sie steht ohne Grund nicht in ohneSender", name)
		case gesendet && benannt:
			t.Errorf("%s hat einen Sender und steht trotzdem in ohneSender", name)
		}
	}
}

func TestJedeNachrichtenartOhneZustandHatEinenEmpfaenger(t *testing.T) {
	roh, fehler := os.ReadFile("../domain/handelsumschlag.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	wahl, fehler := os.ReadFile("../domain/handelsnachrichtenwahl.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	ausserhalb := quelltexteAusserhalbDerDomaene(t)
	arten := regexp.MustCompile(`(?m)^\s*(Nachricht[A-Z]\w*)\s+Handelsnachrichtenart\s*=`).FindAllStringSubmatch(string(roh), -1)

	for _, art := range arten {
		name := art[1]
		mitZustand := regexp.MustCompile(`return ` + name + `, true`).Match(wahl)
		behandelt := regexp.MustCompile(`[!=]= domain\.` + name + `\b`).MatchString(ausserhalb)
		_, benannt := ohneEmpfaenger[name]

		switch {
		case mitZustand && benannt:
			t.Errorf("%s trägt einen Zustand und steht trotzdem in ohneEmpfaenger", name)
		case !mitZustand && !behandelt && !benannt:
			t.Errorf("%s: nichts empfängt sie, und sie steht ohne Grund nicht in ohneEmpfaenger", name)
		case behandelt && benannt:
			t.Errorf("%s wird behandelt und steht trotzdem in ohneEmpfaenger", name)
		}
	}
}

// quelltexteAusserhalbDerDomaene gibt den Code der Schichten über `domain`,
// ohne Tests.
func quelltexteAusserhalbDerDomaene(t *testing.T) string {
	t.Helper()

	var texte []string

	for _, ort := range []string{"../service", "../app", "../rest", "../mcp", "../storage", "../../cmd"} {
		if fehler := filepath.WalkDir(ort, func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
			texte = append(texte, string(inhalt))

			return lesefehler
		}); fehler != nil {
			t.Fatalf("%s lesen: %v", ort, fehler)
		}
	}

	return strings.Join(texte, "\n")
}
