package domain_test

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Sechs Arten, jede genau einmal, in der Reihenfolge ihrer Schwere.**
func TestEsGibtSechsRuecksetzungsarten(t *testing.T) {
	t.Parallel()

	arten := domain.Ruecksetzungsarten()
	if len(arten) != 6 {
		t.Fatalf("es sind %d Arten: %v", len(arten), arten)
	}

	// **Die letzte ist die schwerste** — daran hängt die Reihenfolge im
	// Fenster.
	if arten[0] != domain.RuecksetzungLayout ||
		arten[5] != domain.RuecksetzungAnwendung {
		t.Errorf("die Reihe fängt mit %q an und endet mit %q", arten[0], arten[5])
	}

	gesehen := map[domain.Ruecksetzungsart]bool{}

	for _, art := range arten {
		if gesehen[art] {
			t.Errorf("%q steht zweimal in der Reihe", art)
		}

		gesehen[art] = true

		if !domain.IstRuecksetzungsart(string(art)) {
			t.Errorf("%q gilt nicht als Art", art)
		}
	}

	if domain.IstRuecksetzungsart("alles") {
		t.Error("eine erfundene Art gilt als Art")
	}
}

// **Nur die ganze Anwendung braucht einen Nachweis.**
//
// Ein Profilverwalter darf sein Profil zurücksetzen und seinen Nutzer
// löschen; die ganze Anwendung zurückzusetzen nähme **anderen Eigentümern**
// ihre Daten weg, und dafür reicht keine Rolle.
func TestNurDieGanzeAnwendungBrauchtEinenNachweis(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Ruecksetzungsarten() {
		braucht := art.BrauchtNachweis()
		erwartet := art == domain.RuecksetzungAnwendung

		if braucht != erwartet {
			t.Errorf("%q braucht %v, erwartet war %v", art, braucht, erwartet)
		}
	}
}

// **Nur die beiden, die etwas Benanntes löschen, verlangen einen Namen.**
//
// **Nicht dieselbe Menge wie [LoeschtDaten]**: die ganze Anwendung löscht am
// meisten und hat keinen Namen, den man abtippen könnte.
func TestNurBenanntesVerlangtEinenNamen(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Ruecksetzungsarten() {
		braucht := art.BrauchtNamen()
		erwartet := art == domain.RuecksetzungProfil || art == domain.RuecksetzungNutzer

		if braucht != erwartet {
			t.Errorf("%q braucht %v, erwartet war %v", art, braucht, erwartet)
		}
	}
}

// **Was Daten wegnimmt, bekommt eine Sicherung angeboten — was eine Wahl
// zurücknimmt, nicht.**
func TestNurWasDatenWegnimmtIstEineLoeschung(t *testing.T) {
	t.Parallel()

	for art, erwartet := range map[domain.Ruecksetzungsart]bool{
		domain.RuecksetzungLayout:              false,
		domain.RuecksetzungProfileinstellungen: false,
		domain.RuecksetzungNutzereinstellungen: false,
		domain.RuecksetzungProfil:              true,
		domain.RuecksetzungNutzer:              true,
		domain.RuecksetzungAnwendung:           true,
	} {
		if loescht := art.LoeschtDaten(); loescht != erwartet {
			t.Errorf("%q löscht %v, erwartet war %v", art, loescht, erwartet)
		}
	}

	// **Und eine erfundene Art löscht nichts** — im Zweifel wird nicht
	// gelöscht.
	if domain.Ruecksetzungsart("zeitreise").LoeschtDaten() {
		t.Error("eine erfundene Art gilt als Löschung")
	}
}

// **Die Layoutliste ist die Zusage: was hier nicht steht, wird nicht
// angefasst.**
//
// Und jeder Schlüssel darin ist einer, den es gibt — sonst setzte die
// Rücksetzung etwas zurück, das niemand kennt.
func TestJedeLayouteinstellungGibtEsWirklich(t *testing.T) {
	t.Parallel()

	reihe := domain.Einstellungsreihe()

	for _, schluessel := range domain.Layouteinstellungen() {
		if !slices.Contains(reihe, schluessel) {
			t.Errorf("%q steht in der Layoutliste, aber in keiner Einstellungsreihe", schluessel)
		}
	}

	// **Die Sprache gehört nicht dazu** — eine Sprache ist kein Layout, und
	// eine Rücksetzung, die sie mitnähme, spräche danach plötzlich Englisch.
	if slices.Contains(domain.Layouteinstellungen(), domain.EinstellungSprache) {
		t.Error("die Sprache steht in der Layoutliste")
	}
}

// **Eine Vorschau ohne Hindernis ist bereit** — und eine mit einem nicht.
func TestEineVorschauMitHindernisIstNichtBereit(t *testing.T) {
	t.Parallel()

	if !(domain.Ruecksetzungsvorschau{Art: domain.RuecksetzungLayout}).Bereit() {
		t.Error("eine Vorschau ohne Hindernis gilt als nicht bereit")
	}

	mit := domain.Ruecksetzungsvorschau{
		Art: domain.RuecksetzungProfil,
		Hindernisse: []domain.Hindernis{
			{Art: domain.RuecksetzungOffeneRueckkehr, Stelle: "7"},
		},
	}

	if mit.Bereit() {
		t.Error("eine Vorschau mit Hindernis gilt als bereit")
	}
}
