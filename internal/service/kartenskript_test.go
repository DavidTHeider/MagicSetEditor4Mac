package service_test

// **Ein Kartentext, der im Spiel wirkt** — Werkstück M8 von Phase 5.5
// (#1610).
//
// # Die drei Abnahmen des Issues, der Reihe nach
//
//  1. Eine selbst gebaute Karte mit einer Fähigkeit in MSE-Script **wirkt im
//     Spiel** — „das ist der Punkt dieser Phase".
//  2. Ein Skript, das nicht endet, **bricht ab und wird zum Befund**, statt
//     die Partie anzuhalten.
//  3. Ein Skript, das etwas versucht, was die Regeln nicht erlauben, wird
//     **abgewiesen** (CR 101.2).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Skript schreibt Wirkungen auf** — und nichts sonst.
func TestEinSkriptSchreibtWirkungenAuf(t *testing.T) {
	t.Parallel()

	wirkungen, fehler := service.NeuesKartenskript().Deuten(
		t.Context(),
		`draw(amount: 2)
		 change_life(player: target, amount: -3)`,
		service.Skriptumfeld{
			Beherrscher: "sitz-1", Objekt: "sitz-1/blitz/0",
			Ziele: []string{"sitz-2"},
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(wirkungen) != 2 {
		t.Fatalf("%d Wirkungen: %+v", len(wirkungen), wirkungen)
	}

	// **Ohne `player:` gilt der Beherrscher** — CR 113.8.
	if wirkungen[0] != (domain.Spielwirkung{
		Art: domain.WirkungZiehen, Wer: "sitz-1",
		Objekt: "sitz-2", Marke: domain.MarkePlusEins, Menge: 2,
	}) {
		t.Errorf("%+v", wirkungen[0])
	}

	if wirkungen[1].Wer != "sitz-2" || wirkungen[1].Menge != -3 {
		t.Errorf("%+v", wirkungen[1])
	}
}

// **Ein leeres Skript gibt nichts und ist kein Fehler** — der Normalfall.
func TestEinLeeresSkriptGibtNichts(t *testing.T) {
	t.Parallel()

	wirkungen, fehler := service.NeuesKartenskript().Deuten(
		t.Context(), "", service.Skriptumfeld{Beherrscher: "sitz-1"})

	if fehler != nil || wirkungen != nil {
		t.Errorf("%+v, %v", wirkungen, fehler)
	}
}

// **Ein Skript ohne Ende bricht ab** — die zweite Abnahme.
//
// **Zwei Wände, und beide werden gefahren**: die Tiefengrenze des Auswerters
// gegen `a := {a()}` und die **Frist** gegen eine Schleife, die rechnet.
func TestEinSkriptOhneEndeBrichtAb(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		quelle string
	}{
		{name: "endlose Rekursion", quelle: "a := { a() }\na()"},
		{
			name: "endlose Wiederholung",
			quelle: `a := { draw() + a() }
			         a()`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := service.NeuesKartenskript().
				MitFrist(50*time.Millisecond).
				Deuten(t.Context(), fall.quelle, service.Skriptumfeld{
					Beherrscher: "sitz-1", Objekt: "sitz-1/schlinge/0",
				})

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("das Skript lief durch: %v", fehler)
			}

			// **Und der Satz nennt die Karte** — wer ihn liest, soll
			// wissen, welche es war (§8).
			if befund := service.SkriptBefund(fehler); befund == "" {
				t.Error("der Befund ist leer")
			}
		})
	}
}

// **Ein Skript, das etwas Unbekanntes ruft, wird abgewiesen** — es kommt
// nicht an die Partie.
func TestEinSkriptMitUnbekanntemNamenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuesKartenskript().Deuten(
		t.Context(), `win_the_game()`, service.Skriptumfeld{
			Beherrscher: "sitz-1", Objekt: "sitz-1/schummel/0",
		})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das Skript lief durch: %v", fehler)
	}
}

// **Und was die Regeln nicht erlauben, wird ausgelassen** — CR 101.2, die
// dritte Abnahme.
//
// **Die Wirkung entsteht und geschieht nicht**: der Sandkasten schreibt sie
// auf, und [domain.Partie.Anwenden] lässt sie aus. Das ist die Trennung, um
// derentwillen es die benannten Wege gibt.
func TestWasDieRegelnNichtErlaubenWirdAusgelassen(t *testing.T) {
	t.Parallel()

	wirkungen, fehler := service.NeuesKartenskript().Deuten(
		t.Context(), `destroy(object: "gibtesnicht")`,
		service.Skriptumfeld{Beherrscher: "sitz-1"})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(wirkungen) != 1 {
		t.Fatalf("%d Wirkungen", len(wirkungen))
	}

	vorher := einHandlungsstand(t)

	danach, ging := vorher.Anwenden(wirkungen[0])
	if ging {
		t.Error("die Wirkung ging")
	}

	if len(danach.Objekte) != len(vorher.Objekte) {
		t.Error("die Partie hat sich geändert")
	}
}

// **Und ein Skript kann die Bibliothek nicht verbiegen.**
//
// **Eine Zuweisung legt einen Namen in der Kindebene an** — die Bibliothek
// bleibt, wie sie ist, und der nächste Zauber sieht keinen `draw`, den
// jemand ersetzt hat.
func TestEinSkriptVerbiegtDieBibliothekNicht(t *testing.T) {
	t.Parallel()

	skript := service.NeuesKartenskript()

	if _, fehler := skript.Deuten(t.Context(), `draw := { 1 }`,
		service.Skriptumfeld{Beherrscher: "sitz-1"}); fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	wirkungen, fehler := skript.Deuten(t.Context(), `draw()`,
		service.Skriptumfeld{Beherrscher: "sitz-1"})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(wirkungen) != 1 || wirkungen[0].Art != domain.WirkungZiehen {
		t.Errorf("%+v", wirkungen)
	}
}

// **Eine Menge, die keine Zahl ist, ist keine Zahl und kein Fehler** — §8.
//
// **Die Vorgabe greift**: ein Kartentext mit einem Tippfehler zieht eine
// Karte statt die Partie anzuhalten.
func TestEineMengeDieKeineZahlIstNimmtDieVorgabe(t *testing.T) {
	t.Parallel()

	wirkungen, fehler := service.NeuesKartenskript().Deuten(
		t.Context(), `draw(amount: "viele")`,
		service.Skriptumfeld{Beherrscher: "sitz-1"})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(wirkungen) != 1 || wirkungen[0].Menge != 1 {
		t.Errorf("%+v", wirkungen)
	}
}

// **Eine überschrittene Frist wird als solche benannt** — §8.
//
// **Eine Frist von einer Nanosekunde ist schon abgelaufen**, wenn der erste
// Knoten gerechnet wird; damit misst dieser Lauf den Abbruch und nicht die
// Geschwindigkeit dieses Rechners.
func TestEineUeberschritteneFristWirdBenannt(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuesKartenskript().MitFrist(time.Nanosecond).
		Deuten(t.Context(), `draw()`, service.Skriptumfeld{
			Beherrscher: "sitz-1", Objekt: "sitz-1/langsam/0",
		})

	if !errors.Is(fehler, context.DeadlineExceeded) {
		t.Fatalf("die Frist hielt: %v", fehler)
	}

	if befund := service.SkriptBefund(fehler); !strings.Contains(
		befund, "Frist") {
		t.Errorf("der Befund nennt die Frist nicht: %q", befund)
	}

	// **Und ohne Fehler gibt es keinen Befund** — die Gegenprobe.
	if befund := service.SkriptBefund(nil); befund != "" {
		t.Errorf("ein Befund ohne Fehler: %q", befund)
	}
}

// **`grant` nennt sein Schlüsselwort unter `keyword:`** — CR 613.1f über den
// Sandkasten (M10).
func TestGrantNenntSeinSchluesselwort(t *testing.T) {
	t.Parallel()

	wirkungen, fehler := service.NeuesKartenskript().Deuten(
		t.Context(), `grant(object: this, keyword: "flying")`,
		service.Skriptumfeld{
			Beherrscher: "sitz-1", Objekt: "sitz-1/baer/0",
		})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if len(wirkungen) != 1 || wirkungen[0].Marke != domain.SchluesselFlug {
		t.Errorf("%+v", wirkungen)
	}
}

// **Und die elf Wege sind dieselben wie in der Domäne** — sonst gäbe es
// eine Wirkungsart, die kein Skript rufen kann (§10).
func TestJedeWirkungsartHatEineFunktion(t *testing.T) {
	t.Parallel()

	gefunden := map[domain.Wirkungsart]bool{}

	for _, art := range domain.Wirkungsarten() {
		wirkungen, fehler := service.NeuesKartenskript().Deuten(
			t.Context(), service.Skriptname(art)+"()",
			service.Skriptumfeld{Beherrscher: "sitz-1"})
		if fehler != nil {
			t.Fatalf("%q: %v", art, fehler)
		}

		if len(wirkungen) != 1 {
			t.Fatalf("%q gibt %d Wirkungen", art, len(wirkungen))
		}

		gefunden[wirkungen[0].Art] = true
	}

	if len(gefunden) != len(domain.Wirkungsarten()) {
		t.Errorf("%d von %d Arten", len(gefunden), len(domain.Wirkungsarten()))
	}
}
