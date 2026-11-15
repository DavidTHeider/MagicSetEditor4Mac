package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Karten kopieren, ausschneiden und einfügen** — Werkstück 1 von Phase 3.7
// (#1168).

// ablagequelle ist ein Doppel des Anwendungsfalls.
type ablagequelle struct {
	inhalt      string
	ergebnis    *service.Einfuegung
	fehler      error
	geschnitten bool
	gefragt     []int64
	eingefuegt  string
}

func (a *ablagequelle) Kopieren(_ context.Context, _ int64, karteIDs []int64) (string, error) {
	a.gefragt = karteIDs

	return a.inhalt, a.fehler
}

// **Der Doppelgänger legt wirklich ab** — Befund 2 des Nachlaufs über
// Phase 3.7: `ablegen` muss glücken, bevor eine Karte fällt, und ein
// Doppelgänger, der es überspränge, wäre großzügiger als der Betrieb (§3).
func (a *ablagequelle) Ausschneiden(
	_ context.Context, _ int64, karteIDs []int64, ablegen func(string) error,
) (string, error) {
	if fehler := ablegen(a.inhalt); fehler != nil {
		return "", fehler
	}

	a.geschnitten = true
	a.gefragt = karteIDs

	return a.inhalt, a.fehler
}

func (a *ablagequelle) Einfuegen(
	_ context.Context, _ int64, inhalt string,
) (*service.Einfuegung, error) {
	a.eingefuegt = inhalt

	return a.ergebnis, a.fehler
}

// mitAblage baut eine Anwendung samt Zwischenablage im Arbeitsspeicher.
func mitAblage(quelle *ablagequelle) (fassade *app.App, abgelegt *string) {
	anwendung := app.NeueApp("")
	anwendung.MitKartenablage(quelle)

	tafel := new(string)

	anwendung.MitZwischenablage(
		func(context.Context) (string, error) { return *tafel, nil },
		func(_ context.Context, inhalt string) error {
			*tafel = inhalt

			return nil
		})

	return anwendung, tafel
}

// **Kopieren legt den Inhalt in die Zwischenablage.**
func TestAblageKopierenSchreibtInDieZwischenablage(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{inhalt: "game_name: magic\ncard:\n"}
	anwendung, tafel := mitAblage(quelle)

	if antwort := anwendung.AblageKopieren(
		probekontext(t), "1", []string{"11", "12"}); antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if *tafel != quelle.inhalt {
		t.Errorf("in der Ablage steht %q", *tafel)
	}

	if len(quelle.gefragt) != 2 || quelle.gefragt[0] != 11 {
		t.Errorf("gefragt wurde nach %v", quelle.gefragt)
	}
}

// **Ausschneiden geht denselben Weg — über den anderen Anwendungsfall.**
func TestAblageAusschneidenNimmtDenAnderenWeg(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{inhalt: "game_name: magic\ncard:\n"}
	anwendung, tafel := mitAblage(quelle)

	if antwort := anwendung.AblageAusschneiden(
		probekontext(t), "1", []string{"11"}); antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if !quelle.geschnitten || *tafel == "" {
		t.Errorf("geschnitten %v, Ablage %q", quelle.geschnitten, *tafel)
	}
}

// **Ein gescheitertes Löschen gibt den Inhalt trotzdem in die Ablage** — und
// meldet den Fehler daneben.
//
// Er in der Ablage und die Karte im Set ist wiederholbar; beides zu verlieren
// nicht.
func TestEinGescheitertesAusschneidenSchreibtTrotzdem(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{
		inhalt: "game_name: magic\ncard:\n", fehler: domain.ErrBelegt,
	}

	anwendung, tafel := mitAblage(quelle)

	antwort := anwendung.AblageAusschneiden(t.Context(), "1", []string{"11"})
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if *tafel == "" {
		t.Error("die Ablage blieb leer")
	}
}

// **Ohne Inhalt wird nichts geschrieben.**
func TestOhneInhaltWirdNichtsGeschrieben(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{fehler: domain.ErrNotFound}
	anwendung, tafel := mitAblage(quelle)

	if antwort := anwendung.AblageKopieren(
		probekontext(t), "1", []string{"11"}); antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if *tafel != "" {
		t.Errorf("in der Ablage steht %q", *tafel)
	}
}

// **Einfügen liest die Zwischenablage und gibt die neuen Kennungen.**
func TestAblageEinfuegenGibtDieNeuenKennungen(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{ergebnis: &service.Einfuegung{Karten: []int64{100, 101}}}
	anwendung, tafel := mitAblage(quelle)
	*tafel = "game_name: magic\ncard:\n\tname: Erste\n"

	antwort := anwendung.AblageEinfuegen(t.Context(), "2")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Karten) != 2 || antwort.Karten[0] != "100" {
		t.Errorf("es kamen %v", antwort.Karten)
	}

	if quelle.eingefuegt != *tafel {
		t.Errorf("eingefügt wurde %q", quelle.eingefuegt)
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, bevor etwas geschieht.**
func TestEineUnbrauchbareKennungErreichtDieAblageNicht(t *testing.T) {
	t.Parallel()

	for name, ruf := range map[string]func(*app.App) *app.Fehler{
		"setID beim Kopieren": func(a *app.App) *app.Fehler {
			return a.AblageKopieren(probekontext(t), "x", []string{"11"}).Fehler
		},
		"karteID beim Kopieren": func(a *app.App) *app.Fehler {
			return a.AblageKopieren(probekontext(t), "1", []string{"x"}).Fehler
		},
		"setID beim Einfügen": func(a *app.App) *app.Fehler {
			return a.AblageEinfuegen(probekontext(t), "x").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := &ablagequelle{inhalt: "game_name: magic\n"}
			anwendung, _ := mitAblage(quelle)

			if fehler := ruf(anwendung); fehler == nil {
				t.Fatal("die unbrauchbare Kennung kam als Erfolg zurück")
			}

			if quelle.gefragt != nil {
				t.Errorf("es ging %v hinein", quelle.gefragt)
			}
		})
	}
}

// **Ohne eingerichtete Ablage sagt jede der drei Bindungen es.**
func TestOhneAblageSagenDieBindungenEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	for name, fehler := range map[string]*app.Fehler{
		"kopieren":     anwendung.AblageKopieren(t.Context(), "1", []string{"11"}).Fehler,
		"ausschneiden": anwendung.AblageAusschneiden(t.Context(), "1", []string{"11"}).Fehler,
		"einfügen":     anwendung.AblageEinfuegen(t.Context(), "1").Fehler,
	} {
		if fehler == nil {
			t.Errorf("%s: die fehlende Ablage kam als Erfolg zurück", name)
		}
	}
}

// **Und ein Lesefehler der Zwischenablage wird übersetzt.**
func TestEinLesefehlerDerZwischenablageWirdUebersetzt(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenablage(&ablagequelle{})
	anwendung.MitZwischenablage(
		func(context.Context) (string, error) { return "", domain.ErrBelegt },
		func(context.Context, string) error { return nil })

	if antwort := anwendung.AblageEinfuegen(t.Context(), "1"); antwort.Fehler == nil {
		t.Fatal("der Lesefehler kam als Erfolg zurück")
	}
}

// **Und ein Schreibfehler ebenso.**
func TestEinSchreibfehlerDerZwischenablageWirdUebersetzt(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenablage(&ablagequelle{inhalt: "game_name: magic\n"})
	anwendung.MitZwischenablage(
		func(context.Context) (string, error) { return "", nil },
		func(context.Context, string) error { return domain.ErrBelegt })

	if antwort := anwendung.AblageKopieren(
		probekontext(t), "1", []string{"11"}); antwort.Fehler == nil {
		t.Fatal("der Schreibfehler kam als Erfolg zurück")
	}
}

// **Ein Fehler beim Einfügen kommt mit den schon entstandenen Karten.**
func TestEinFehlerBeimEinfuegenNenntDieEntstandenenKarten(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{
		ergebnis: &service.Einfuegung{Karten: []int64{100}},
		fehler:   domain.ErrBelegt,
	}

	anwendung, tafel := mitAblage(quelle)
	*tafel = "game_name: magic\ncard:\n"

	antwort := anwendung.AblageEinfuegen(t.Context(), "2")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if len(antwort.Karten) != 1 || antwort.Karten[0] != "100" {
		t.Errorf("es kamen %v", antwort.Karten)
	}
}

// **Und ein Fehler ohne Ergebnis ebenfalls** — dann gibt es nichts zu nennen.
func TestEinFehlerOhneErgebnisGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	quelle := &ablagequelle{fehler: domain.ErrValidation}
	anwendung, tafel := mitAblage(quelle)
	*tafel = "unsinn"

	antwort := anwendung.AblageEinfuegen(t.Context(), "2")
	if !errors.Is(domain.ErrValidation, domain.ErrValidation) || antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if antwort.Karten == nil || len(antwort.Karten) != 0 {
		t.Errorf("es kamen %v", antwort.Karten)
	}
}
