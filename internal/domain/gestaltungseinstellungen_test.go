package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die vier Einstellungen aus UX-09** (#1681): Farbwelt, Reichweite der
// Schrift, Fensterbild und eigene Merkmale — mit Werten, Vorgabe und Ebene.
func TestDieGestaltungseinstellungenStehenInDerListe(t *testing.T) {
	faelle := []struct {
		name       string
		schluessel string
		werte      []string
		vorgabe    string
	}{
		{
			"die Farbwelt", domain.EinstellungFarbwelt,
			[]string{"pergament", "klar", "schiefer", "moos"},
			"pergament",
		},
		{
			"die Reichweite der Schrift", domain.EinstellungSchriftbereich,
			[]string{"ueberschriften", "ganz"},
			"ueberschriften",
		},
		{
			"das Fensterbild", domain.EinstellungHintergrund,
			[]string{"aus", "fuellen", "einpassen", "kacheln"},
			"aus",
		},
		{"die eigenen Merkmale", domain.EinstellungGestaltung, []string{}, ""},
	}
	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if !slices.Contains(domain.Einstellungsreihe(), fall.schluessel) {
				t.Fatalf("%q fehlt in der Reihe", fall.schluessel)
			}
			if werte := domain.ErlaubteEinstellungswerte(fall.schluessel); !slices.Equal(werte, fall.werte) {
				t.Fatalf("Werte %v, erwartet %v", werte, fall.werte)
			}
			if vorgabe := domain.EinstellungsVorgabe(fall.schluessel); vorgabe != fall.vorgabe {
				t.Fatalf("Vorgabe %q, erwartet %q", vorgabe, fall.vorgabe)
			}
			// **Die Arbeitsumgebung entscheidet** — wer ein Profil für
			// Turniere und eines für Entwürfe führt, sieht sie verschieden.
			if ebene, _ := domain.EbeneVon(fall.schluessel); ebene != domain.EbeneProfil {
				t.Fatalf("Ebene %q, erwartet profil", ebene)
			}
			// **Die Rücksetzung des Layouts fasst sie an** (AK-03): ein
			// sicherer Stand, auch wenn nichts mehr zu lesen ist.
			if !slices.Contains(domain.Layouteinstellungen(), fall.schluessel) {
				t.Fatalf("%q fehlt in der Layoutrücksetzung", fall.schluessel)
			}
		})
	}
}

// **Drei Textschriften kommen zur Anzeigeschrift** (UX-09.AK-02) — sie
// tragen, wenn gewünscht, die ganze Oberfläche; die drei Schmuckschriften
// bleiben dabei an den Überschriften.
func TestDieAnzeigeschriftKenntDreiTextschriften(t *testing.T) {
	werte := domain.ErlaubteEinstellungswerte(domain.EinstellungSchriftart)
	for _, neu := range []string{"buch", "rund", "serifenlos"} {
		if !slices.Contains(werte, neu) {
			t.Fatalf("%q fehlt in %v", neu, werte)
		}
		if fehler := domain.PruefeNutzereinstellung(domain.EinstellungSchriftart, neu); fehler != nil {
			t.Fatalf("beim Menschen abgewiesen: %v", fehler)
		}
	}
}
