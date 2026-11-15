package service_test

// **Die Anmeldung über einen Anbieter** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Wer zum ersten Mal kommt, bekommt einen Nutzer samt Profil.**
//
// # Warum nicht „bitte erst anlegen"
//
// **Weil das eine Hürde ohne Gegenwert wäre.** Wer sich über GitHub anmeldet,
// hat gesagt, wer er ist; ihn danach zu bitten, seinen Namen einzutippen,
// fragt zweimal dasselbe.
func TestWerZumErstenMalKommtBekommtEinenNutzer(t *testing.T) {
	t.Parallel()

	ablage := &anmeldedoppel{}

	nutzerID, fehler := service.NeueAnmeldung(ablage).Herein(t.Context(),
		domain.Fremdkennung{Anbieter: "github", Subjekt: "4711", Anzeigename: "Wer"})
	if fehler != nil {
		t.Fatalf("herein: %v", fehler)
	}

	if nutzerID == 0 {
		t.Fatal("kein Nutzer angelegt")
	}

	if ablage.angelegt != "Wer" {
		t.Errorf("angelegt wurde %q", ablage.angelegt)
	}

	if ablage.verbunden.Subjekt != "4711" {
		t.Errorf("verbunden wurde %+v", ablage.verbunden)
	}
}

// **Wer wiederkommt, bekommt seinen eigenen Nutzer.**
//
// **Und keinen zweiten.** Ein zweiter hieße: dieselbe Person, zwei Bestände —
// und der erste wäre unerreichbar.
func TestWerWiederkommtBekommtSeinenEigenenNutzer(t *testing.T) {
	t.Parallel()

	ablage := &anmeldedoppel{bekannt: 7}

	nutzerID, fehler := service.NeueAnmeldung(ablage).Herein(t.Context(),
		domain.Fremdkennung{Anbieter: "github", Subjekt: "4711"})
	if fehler != nil {
		t.Fatalf("herein: %v", fehler)
	}

	if nutzerID != 7 {
		t.Errorf("Nutzer %d, erwartet war 7", nutzerID)
	}

	if ablage.angelegt != "" {
		t.Errorf("es wurde trotzdem angelegt: %q", ablage.angelegt)
	}
}

// **Ein Anbieter ohne Anzeigenamen bekommt einen aus Anbieter und Subjekt.**
//
// **Weil ein Nutzer einen Namen haben muss** (`domain.Nutzer.Validate`), und
// wer sein Profil beim Anbieter nicht ausgefüllt hat, soll deswegen nicht
// draußen bleiben.
func TestEinAnbieterOhneAnzeigenamenBekommtEinen(t *testing.T) {
	t.Parallel()

	ablage := &anmeldedoppel{}

	if _, fehler := service.NeueAnmeldung(ablage).Herein(t.Context(),
		domain.Fremdkennung{Anbieter: "github", Subjekt: "4711"}); fehler != nil {
		t.Fatalf("herein: %v", fehler)
	}

	if ablage.angelegt != "github:4711" {
		t.Errorf("angelegt wurde %q", ablage.angelegt)
	}
}

// **Ein Name, den es schon gibt, hält niemanden auf.**
//
// # Der Fall, den man übersieht
//
// Zwei Menschen heißen bei zwei Anbietern gleich. Der Anzeigename ist im
// Bestand **eindeutig** (Werkstück 1) — der zweite käme also nicht herein,
// obwohl an seiner Anmeldung nichts falsch ist.
//
// **Deshalb bekommt er einen Namen mit Anbieter dahinter.** Das ist eine
// Abweichung von dem, was der Anbieter sagt, und sie ist sichtbar: der Mensch
// sieht seinen Namen und kann ihn ändern.
func TestEinBelegterNameHaeltNiemandenAuf(t *testing.T) {
	t.Parallel()

	ablage := &anmeldedoppel{belegt: "Wer"}

	if _, fehler := service.NeueAnmeldung(ablage).Herein(t.Context(),
		domain.Fremdkennung{Anbieter: "github", Subjekt: "4711", Anzeigename: "Wer"},
	); fehler != nil {
		t.Fatalf("herein: %v", fehler)
	}

	if ablage.angelegt != "Wer (github)" {
		t.Errorf("angelegt wurde %q", ablage.angelegt)
	}
}

// **Eine Kennung ohne Subjekt kommt nicht herein.**
func TestEineKennungOhneSubjektKommtNichtHerein(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeueAnmeldung(&anmeldedoppel{}).Herein(t.Context(),
		domain.Fremdkennung{Anbieter: "github"})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Jeder Fehlerweg der Anmeldung reicht durch.**
func TestJederFehlerwegDerAnmeldungReichtDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]*anmeldedoppel{
		"die Suche scheitert":     {suchfehler: errors.New("die Platte schweigt")},
		"das Anlegen scheitert":   {anlagefehler: errors.New("kein Platz")},
		"das Verbinden scheitert": {verbindfehler: errors.New("geht nicht")},
	}

	for name, ablage := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := service.NeueAnmeldung(ablage).Herein(t.Context(),
				domain.Fremdkennung{Anbieter: "github", Subjekt: "1"}); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// anmeldedoppel ist eine Kontenablage im Arbeitsspeicher.
type anmeldedoppel struct {
	// bekannt ist der Nutzer, den die Suche findet — 0 heißt: keiner.
	bekannt int64

	// belegt ist ein Name, den es schon gibt.
	belegt string

	angelegt  string
	verbunden domain.Anbieterkonto

	suchfehler    error
	anlagefehler  error
	verbindfehler error
}

func (d *anmeldedoppel) NutzerZuKonto(context.Context, string, string) (
	int64, error,
) {
	if d.suchfehler != nil {
		return 0, d.suchfehler
	}

	if d.bekannt == 0 {
		return 0, domain.ErrNotFound
	}

	return d.bekannt, nil
}

func (d *anmeldedoppel) NutzerAnlegen(_ context.Context, neu domain.Nutzer) (
	domain.Nutzer, error,
) {
	if d.anlagefehler != nil {
		return domain.Nutzer{}, d.anlagefehler
	}

	// **Der belegte Name antwortet wie der Bestand**: mit einem Konflikt.
	if neu.Name == d.belegt {
		return domain.Nutzer{}, domain.ErrConflict
	}

	d.angelegt = neu.Name
	neu.ID = 42

	return neu, nil
}

func (d *anmeldedoppel) KontoVerbinden(_ context.Context, neu domain.Anbieterkonto) (
	domain.Anbieterkonto, error,
) {
	if d.verbindfehler != nil {
		return domain.Anbieterkonto{}, d.verbindfehler
	}

	d.verbunden = neu

	return neu, nil
}
