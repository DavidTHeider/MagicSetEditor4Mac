package service_test

// **Einen Handel wählen, statt seine Kennung zu tippen** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind."*
//
// # Was hier gemessen wird
//
// **Dass die Leitung die Kennung eines neuen Handels selbst vergibt**, wenn
// niemand eine nennt — und eine genannte stehen lässt. **Und dass die Handel
// mit einem Gegenüber eine Liste sind**, aus der man wählt: wer einen Handel
// empfängt, erfuhr seine Kennung bisher nirgends im Fenster.

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne genannte Kennung vergibt die Leitung eine.**
func TestDieLeitungVergibtDieKennungEinesNeuenHandels(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	leitung.MitKennungen(func() string { return "h-probe" })

	angebot := einHandelsangebot()
	angebot.Vorgang = ""

	angelegt, fehler := leitung.Eroeffnen(context.Background(), angebot)
	if fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	if angelegt.Vorgang != "h-probe" {
		t.Errorf("eröffnet wurde %q", angelegt.Vorgang)
	}

	if _, da := ablage.staende["h-probe"]; !da {
		t.Errorf("abgelegt wurde %v", ablage.staende)
	}
}

// **Eine genannte Kennung bleibt** — ein Handel, den die Gegenseite eröffnet
// hat, kommt mit seiner, und die ist verabredet.
func TestEineGenannteKennungBleibt(t *testing.T) {
	t.Parallel()

	leitung, _ := eineHandelsleitung()
	leitung.MitKennungen(func() string { return "h-probe" })

	angelegt, fehler := leitung.Eroeffnen(context.Background(),
		einHandelsangebot())
	if fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	if angelegt.Vorgang != "h1" {
		t.Errorf("eröffnet wurde %q", angelegt.Vorgang)
	}
}

// **Die vergebene Kennung ist zufällig und lesbar** — zwei Handel, zwei
// Kennungen; ein fester Vorsatz sagt, was sie ist.
func TestDieVergebeneKennungIstZufaelligUndLesbar(t *testing.T) {
	t.Parallel()

	leitung, _ := eineHandelsleitung()
	gestalt := regexp.MustCompile(`^h-[0-9a-f]{16}$`)
	gesehen := map[string]bool{}

	for range 2 {
		angebot := einHandelsangebot()
		angebot.Vorgang = ""

		angelegt, fehler := leitung.Eroeffnen(context.Background(), angebot)
		if fehler != nil {
			t.Fatalf("Eroeffnen: %v", fehler)
		}

		if !gestalt.MatchString(angelegt.Vorgang) {
			t.Errorf("vergeben wurde %q", angelegt.Vorgang)
		}

		gesehen[angelegt.Vorgang] = true
	}

	if len(gesehen) != 2 {
		t.Errorf("zweimal dieselbe Kennung: %v", gesehen)
	}
}

// **Die Handel mit einem Gegenüber — und nur die.**
func TestDieHandelMitEinemGegenueber(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	mitBert := einHandelsangebot()
	mitCara := einHandelsangebot()
	mitCara.Vorgang = "h2"
	mitCara.Parteien = []string{"anna", "cara"}
	ablage.letzte = []domain.Handelsangebot{mitCara, mitBert}

	liste, fehler := leitung.Vorgaenge(context.Background(), "bert")
	if fehler != nil {
		t.Fatalf("Vorgaenge: %v", fehler)
	}

	if len(liste) != 1 || liste[0].Vorgang != "h1" {
		t.Errorf("mit bert: %+v", liste)
	}

	// **Keiner ist eine leere Liste und kein `nil`.**
	keiner, fehler := leitung.Vorgaenge(context.Background(), "dora")
	if fehler != nil || keiner == nil || len(keiner) != 0 {
		t.Errorf("mit dora: %+v, %v", keiner, fehler)
	}
}

// **Ohne Gegenüber keine Liste** — und was der Bestand klagt, kommt zurück.
func TestDieHandelOhneGegenueberOderMitKlage(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	if _, fehler := leitung.Vorgaenge(context.Background(), " "); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("ohne Gegenüber kam %v", fehler)
	}

	ablage.klagt = errors.New("die Platte klemmt")

	if _, fehler := leitung.Vorgaenge(context.Background(), "bert"); !errors.Is(
		fehler, ablage.klagt) {
		t.Errorf("die Klage kam als %v", fehler)
	}
}
