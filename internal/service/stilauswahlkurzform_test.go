// **Auch die Stilauswahl findet den Stil aus der Kurzform**
// (#951).
//
// # Warum neben `stilkurzform_test.go` und nicht darin
//
// Jene Datei belegt die **Regel** — `stylesheet:` trägt die Kurzform, das
// Paket heißt `<spiel>-<stylesheet>` — an dem, der sie von Anfang an befolgt
// hat: dem Stilspeicher hinter dem Renderer. Hier steht der **zweite
// Verbraucher** derselben Regel, der sie zweieinhalb Monate lang nicht
// befolgte. Zwei Dateien, weil es zwei Aussagen sind: „die Regel stimmt" und
// „alle halten sich daran".
//
// **Die Stilauswahl befolgte sie nicht.** Gefunden bei der Sichtprüfung am
// 2026-08-23 an `Vanguard.mse-set`: die Lasche *Stil* zeigte „— kein
// installierter Stil —" und darunter „Der gesuchte Eintrag wurde nicht
// gefunden", während die Karte daneben korrekt gezeichnet wurde.
//
// **Und die alten Tests konnten es nicht sehen**: sie setzen `Style` auf den
// vollen Paketnamen — so schreibt unsere eigene Saat, und so steht es in
// keinem echten Set. Eine Vorrichtung, die nur die eigene Form kennt, ist
// blind für die fremde (§3).
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kurzformkopf ist ein Set, wie MSE2 es schreibt: Spiel und **Kurzform**.
func kurzformkopf() domain.SetHead {
	return domain.SetHead{ID: 1, Game: "probe", Style: "schlicht"}
}

func TestDieAuswahlErkenntDenStilAusDerKurzform(t *testing.T) {
	angebote, err := waehler(t, kurzformkopf(), &stilablage{}, &feldliste{}).
		Stile(t.Context(), 1)
	if err != nil {
		t.Fatalf("Stile: %v", err)
	}

	var gewaehlt string

	for _, angebot := range angebote {
		if angebot.Gewaehlt {
			gewaehlt = angebot.Paket
		}
	}

	if gewaehlt != "probe-schlicht" {
		t.Errorf("gewählt ist %q, erwartet war %q — die Kurzform wurde nicht aufgelöst",
			gewaehlt, "probe-schlicht")
	}
}

func TestDieStiloptionenFindenDenStilAusDerKurzform(t *testing.T) {
	optionen, err := waehler(t, kurzformkopf(), &stilablage{}, &feldliste{}).
		Optionen(t.Context(), 1)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("die Kurzform wurde nicht aufgelöst: %v", err)
		}

		t.Fatalf("Optionen: %v", err)
	}

	if len(optionen) == 0 {
		t.Error("der Stil der Saat hat Optionen, hier kam keine an")
	}
}

// **Ein Wechsel an einem Set, das es nicht gibt, wird gemeldet.**
//
// Der Zweig ist neu: `Wechseln` liest seit #951 zuerst den Setkopf, weil es
// ohne das Spiel keine Kurzform aufzulösen gibt. Ein Leseweg ohne Fehlerpfad
// wäre eine Behauptung darüber, was im Ernstfall geschieht (§3).
func TestEinWechselAnEinemUnbekanntenSetWirdGemeldet(t *testing.T) {
	err := waehler(t, kurzformkopf(), &stilablage{}, &feldliste{}).
		Wechseln(t.Context(), 4711, "probe-schlicht")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("erwartet war ErrNotFound, bekommen: %v", err)
	}
}
