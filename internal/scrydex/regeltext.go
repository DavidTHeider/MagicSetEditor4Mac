package scrydex

// **Der Regeltext einer Scrydex-Karte** — Befund 2 des Zweitblicks über
// Phase 4.8 (#1321).
//
// # Warum ein eigener Typ
//
// **Weil das Feld ein Array ist.** [Magic Cards](https://scrydex.com/docs/magicthegathering/cards),
// abgerufen am 2026-09-02: `rules` ist dort `<array of strings>` — „The various
// rules of the card describing the effects and abilities" —, und das Beispiel
// zeigt `"rules": ["Haste", "Zurgo attacks each combat if able…"]`.
//
// # Und warum er beides liest
//
// **Weil Scrydex der ungewisseste der vier Dienste ist** (Vermerk 0056): es
// gibt hier kein Konto, eine Anfrage ohne Zugang endet mit 401, und ein Schema
// zum Nachsehen gibt es auch nicht. Eine Antwort abzuweisen, die sich lesen
// lässt, nützt niemandem — das ist §8 für fremde Daten, eine Ebene höher.
//
// **Streng bleibt es trotzdem, wo es zählt**: eine Zahl, ein Verbund, ein
// Wahrheitswert — das ist ein Fehler und wird auch so gemeldet.
//
// **`null` ist die benannte Ausnahme**, und sie steht hier, weil der erste
// Kommentar sie verschwieg (Zusatzbefund 2 des Nachlaufs): Go entpackt
// JSON-`null` ohne Fehler in ein nil-Slice, also kam es durch — während
// darüber „was weder Zeichenkette noch Array ist, ist ein Fehler" stand.
//
// **Der Kommentar war falsch, nicht der Code.** `null` sagt dasselbe wie ein
// fehlendes Feld, und eine Karte dafür wegzuwerfen nützt niemandem (§8).

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// regeltext ist der Regeltext, wie Scrydex ihn schickt.
type regeltext string

// UnmarshalJSON liest ein Array von Zeilen — oder eine einzelne Zeichenkette.
//
// **Verbunden mit `\n`**, so wie ein Regeltext auf einer Karte steht: die
// Zeilen des Arrays sind die Absätze des Kastens.
func (r *regeltext) UnmarshalJSON(roh []byte) error {
	var zeilen []string

	if err := json.Unmarshal(roh, &zeilen); err == nil {
		*r = regeltext(strings.Join(zeilen, "\n"))

		return nil
	}

	var eine string

	if err := json.Unmarshal(roh, &eine); err == nil {
		*r = regeltext(eine)

		return nil
	}

	return fmt.Errorf(
		"%w: das Feld rules ist weder eine Zeichenkette noch eine Liste von "+
			"Zeichenketten, sondern %s", domain.ErrValidation, string(roh))
}
