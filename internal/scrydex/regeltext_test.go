package scrydex_test

// **`rules` ist ein Array, keine Zeichenkette** — Befund 2 des Zweitblicks
// über Phase 4.8
// (#1321).
//
// # Was falsch war
//
// `kartenDTO.Rules` stand als `string` da, und die eingecheckte Probe schrieb
// eine Zeichenkette hinein. Damit bestätigte die Vertragsprobe die **eigene
// Annahme** statt den fremden Vertrag (§10) — und jede echte Antwort scheiterte
// schon am `json.Unmarshal`, weil Scrydex ein Array schickt.
//
// # Die Quelle
//
// [Magic Cards](https://scrydex.com/docs/magicthegathering/cards), abgerufen am
// 2026-09-02: `rules` ist dort `<array of strings>` — „The various rules of the
// card describing the effects and abilities" —, und das Beispiel zeigt
//
//	"rules": ["Haste", "Zurgo attacks each combat if able…"]
//
// Daneben stehen `mana_cost`, `type`, `power` und `toughness` als `<string>`;
// die waren richtig.
//
// # Was hier geprüft wird
//
// Die **Zeilen werden mit `\n` verbunden**, so wie ein Regeltext auf einer
// Karte steht. Eine Zeichenkette bleibt zusätzlich lesbar: der Dienst ist der
// ungewisseste der vier, und eine Antwort abzuweisen, die sich lesen lässt,
// nützt niemandem (§8).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

func TestRegeltextKommtAlsArray(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		roh      string
		erwartet string
	}{
		{
			"wie die Dokumentation es zeigt",
			`{"data":[{"id":"a","name":"Zurgo","rules":["Haste","Zurgo attacks each combat if able."]}]}`,
			"Haste\nZurgo attacks each combat if able.",
		},
		{
			"eine einzige Zeile",
			`{"data":[{"id":"a","name":"X","rules":["Fliegend."]}]}`,
			"Fliegend.",
		},
		{
			"gar keine Regeln",
			`{"data":[{"id":"a","name":"X"}]}`,
			"",
		},
		{
			"ein leeres Array",
			`{"data":[{"id":"a","name":"X","rules":[]}]}`,
			"",
		},
		{
			// **`null` ist wie ein fehlendes Feld** — Zusatzbefund 2 des
			// Nachlaufs über Phase 4.8. Go entpackt JSON-`null` ohne Fehler
			// in ein nil-Slice, und der Kommentar daneben versprach einen
			// Fehler für „alles, was weder Zeichenkette noch Liste ist".
			//
			// **Der Kommentar war falsch, nicht der Code.** Eine Karte
			// wegzuwerfen, weil ihr Regeltext `null` ist, nützt niemandem;
			// `null` sagt dasselbe wie ein fehlendes Feld (§8).
			"null wie ein fehlendes Feld",
			`{"data":[{"id":"a","name":"X","rules":null}]}`,
			"",
		},
		{
			// **Toleranz statt Abweisung** (§8): der Dienst ist der
			// ungewisseste der vier. Schickt er doch einmal eine
			// Zeichenkette, wird sie gelesen statt die Karte zu verlieren.
			"eine Zeichenkette bleibt lesbar",
			`{"data":[{"id":"a","name":"X","rules":"Fliegend."}]}`,
			"Fliegend.",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			karten, err := scrydex.LiesKarten([]byte(fall.roh), 200)
			if err != nil {
				t.Fatalf("lesen: %v", err)
			}

			if len(karten) != 1 {
				t.Fatalf("%d Karten, erwartet 1", len(karten))
			}

			if karten[0].Regeltext != fall.erwartet {
				t.Errorf("Regeltext ist %q, erwartet %q",
					karten[0].Regeltext, fall.erwartet)
			}
		})
	}
}

// **Streng bleibt es, wo es zählt.** Was weder Zeichenkette noch Liste ist,
// wird gemeldet — mit dem Feldnamen und dem, was dastand (§8).
func TestEinRegeltextAusEtwasAnderemBrichtLaut(t *testing.T) {
	t.Parallel()

	_, err := scrydex.LiesKarten(
		[]byte(`{"data":[{"id":"a","name":"X","rules":42}]}`), 200)

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v, erwartet ErrValidation", err)
	}

	for _, teil := range []string{"rules", "42"} {
		if !strings.Contains(err.Error(), teil) {
			t.Errorf("die Meldung %q nennt %q nicht", err, teil)
		}
	}
}
