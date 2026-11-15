package domain_test

// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351).

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Einstellung hat Werte, eine Vorgabe, und die Vorgabe steht darunter.**
func TestJedeNutzereinstellungHatIhreWerte(t *testing.T) {
	t.Parallel()

	reihe := domain.Nutzereinstellungsreihe()
	if len(reihe) == 0 {
		t.Fatal("die Reihe ist leer")
	}

	for _, schluessel := range reihe {
		werte := domain.ErlaubteNutzerwerte(schluessel)

		// **Eine leere Werteliste ist eine Antwort** — Werkstück O3 von
		// Phase 5.2
		// (#1427):
		// eine Spaltenbreite entsteht am Ziehgriff und eine Kostengrenze am
		// Taschenrechner; beide lassen sich nicht aufzählen. **Die Zusage
		// schrumpft dabei nicht, sie wird genauer**: geprüft wird, dass genau
		// die Schlüssel ohne Werte auskommen, die es dürfen — eine
		// Aufzählung, die versehentlich leer bliebe, ist weiterhin ein Befund.
		if len(werte) < 2 && !slices.Contains(freieNutzereinstellungen, schluessel) {
			t.Errorf("%s hat %v", schluessel, werte)
		}

		// **Eine leere Vorgabe ist eine Antwort** — seit Werkstück 9 (#1351):
		// die Sprache hat keine, denn was ohne Wahl gilt, weiß nur das
		// Fenster (`navigator.language`). Hier eine zu setzen hieße, die
		// Systemsprache zu überschreiben, sobald jemand die Anwendung zum
		// ersten Mal öffnet.
		vorgabe := domain.NutzerVorgabe(schluessel)
		if vorgabe == "" {
			continue
		}

		// **Über die Prüfung und nicht über die Liste** — Werkstück O3.
		// Vorher stand hier `slices.Contains`; das hielt die halbe Zusage,
		// seit es Einstellungen ohne Werteliste gibt. `PruefeNutzereinstellung`
		// **enthält** die Liste und dazu die eigene Regel: eine Vorgabe von
		// 4000 Punkten fiele jetzt auf, eine Liste sähe sie nie.
		if fehler := domain.PruefeNutzereinstellung(schluessel, vorgabe); fehler != nil {
			t.Errorf("die eigene Vorgabe von %s wird abgewiesen: %v",
				schluessel, fehler)
		}
	}
}

// freieNutzereinstellungen sind die, deren Wert sich nicht aufzählen lässt.
//
// **Eine Aufzählung und keine Regel über Namen** — dieselbe Begründung wie bei
// [freieEinstellungen] nebenan: dass eine Breite freier Text ist, ist eine
// Entscheidung, und an einem Präfix festgemacht wäre sie beim nächsten Namen
// vergessen.
var freieNutzereinstellungen = []string{
	domain.ArbeitsplatzNavigatorbreite,
	domain.ArbeitsplatzInspektorbreite,
	domain.RichtlinieKostengrenze,

	// **Und drei des Handels** — Werkstück `T1` von Phase 5.6 (#1466): eine
	// Menge, ein Wert und eine Frist sind Zahlen und keine Auswahl.
	domain.RichtlinieHandelMenge,
	domain.RichtlinieHandelWert,
	domain.RichtlinieHandelFrist,
}

// **Die Vorgabe ist immer „so wie bisher".**
//
// **Ein frischer Bestand verhält sich, wie er sich immer verhalten hat** —
// eine Einstellung, die beim Einbau etwas ändert, ändert es für jeden, der
// nie danach gefragt hat.
func TestDieVorgabeIstSoWieBisher(t *testing.T) {
	t.Parallel()

	vorgaben := map[string]string{
		domain.NutzerBewegung:          domain.NachSystem,
		domain.NutzerKontrast:          domain.NachSystem,
		domain.NutzerFaktorbeimWechsel: domain.FaktorVerlangen,
	}

	for schluessel, erwartet := range vorgaben {
		if gegeben := domain.NutzerVorgabe(schluessel); gegeben != erwartet {
			t.Errorf("%s hat die Vorgabe %q, erwartet war %q",
				schluessel, gegeben, erwartet)
		}
	}
}

// **Ein unbekannter Schlüssel und ein unbekannter Wert werden abgewiesen —
// jeder mit seinem Feld** (§6).
func TestEineUnbekannteNutzereinstellungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel, wert, feld string
	}{
		"unbekannter Schlüssel": {"barrierefreiheit.farbe", "bunt", "schluessel"},
		"unbekannter Wert":      {domain.NutzerBewegung, "viel", "wert"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeNutzereinstellung(fall.schluessel, fall.wert)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("angenommen wurde: %v", fehler)
			}

			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Field != fall.feld {
				t.Errorf("gemeldet wurde %v", fehler)
			}
		})
	}
}

// **Eine Liste, die der Aufrufer ändert, ändert die Regel nicht.**
func TestDieWertelisteIstEineKopie(t *testing.T) {
	t.Parallel()

	geliehen := domain.ErlaubteNutzerwerte(domain.NutzerBewegung)
	geliehen[0] = "umgebogen"

	if wieder := domain.ErlaubteNutzerwerte(domain.NutzerBewegung); wieder[0] == "umgebogen" {
		t.Error("die Regel ließ sich von außen ändern")
	}
}

// **Ein unbekannter Schlüssel hat keine Werte und keine Vorgabe.**
func TestEinUnbekannterSchluesselHatNichts(t *testing.T) {
	t.Parallel()

	if werte := domain.ErlaubteNutzerwerte("gibtsnicht"); len(werte) != 0 {
		t.Errorf("es kamen %v", werte)
	}

	if vorgabe := domain.NutzerVorgabe("gibtsnicht"); vorgabe != "" {
		t.Errorf("es kam %q", vorgabe)
	}
}
