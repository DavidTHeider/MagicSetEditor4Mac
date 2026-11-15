package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die vier Verhaltensweisen bei einem Namenskonflikt** —
// #491.
//
//	conflicts->Append(_BUTTON_("keep old"));         // 0
//	conflicts->Append(_BUTTON_("overwrite"));        // 1
//	conflicts->Append(_BUTTON_("number"));           // 2
//	conflicts->Append(_BUTTON_("number overwrite")); // 3
//
// `src/gui/images_export_window.cpp`, Zeilen 29–32, Fassung f9d9356 vom
// 2021-07-06. Die Bildausgabe dieses Projekts kannte keine davon: sie
// überschrieb.
//
// **Warum das materiell ist:** das Muster `{card.name}` erzeugt für zwei
// Länder desselben Namens denselben Dateinamen. Heute blieb eines übrig, und
// niemand erfuhr es.
func TestDieVierVerhaltensweisenStehenInMSEsReihenfolge(t *testing.T) {
	t.Parallel()

	werte := domain.ErlaubteEinstellungswerte(domain.EinstellungBildkonflikt)

	erwartet := []string{"altbehalten", "ueberschreiben", "nummerieren", "nummerierenneu"}

	if len(werte) != len(erwartet) {
		t.Fatalf("%v", werte)
	}

	for i := range erwartet {
		if werte[i] != erwartet[i] {
			t.Errorf("an Stelle %d steht %q, erwartet %q", i, werte[i], erwartet[i])
		}
	}
}

// **Die Vorgabe ist `nummerierenneu`** — und sie ist nachgelesen, nicht
// gewählt:
//
//	, images_export_conflicts(CONFLICT_NUMBER_OVERWRITE)
//
// `src/data/settings.cpp`, Zeile 74, Fassung f9d9356 vom 2021-07-06. Damit
// ist auch der letzte Abnahmekasten
// aus #491 eingelöst: die Vorgabe ist **nicht** `ueberschreiben`.
func TestDieVorgabeIstNummerierenNeu(t *testing.T) {
	t.Parallel()

	if gefunden := domain.EinstellungsVorgabe(domain.EinstellungBildkonflikt); gefunden != "nummerierenneu" {
		t.Errorf("%q", gefunden)
	}
}

// **Jede Einstellung sagt, wo sie hingehört.**
//
// Die Wahl beim Namenskonflikt steht bei der **Bildausgabe** und nicht auf der
// Einstellungsseite: sie gehört zu dem Lauf, den man gerade startet. MSE hält
// sie ebenso im Ausgabefenster (`images_export_window.cpp`) und nicht in den
// Programmeinstellungen.
func TestJedeEinstellungNenntIhrenBereich(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		domain.EinstellungSprache:          domain.BereichAnwendung,
		domain.EinstellungErscheinungsbild: domain.BereichAnwendung,
		domain.EinstellungSchriftgroesse:   domain.BereichAnwendung,
		domain.EinstellungBildkonflikt:     domain.BereichBildausfuhr,
	}

	for schluessel, erwartet := range faelle {
		if gefunden := domain.EinstellungsBereich(schluessel); gefunden != erwartet {
			t.Errorf("%s: %q, erwartet %q", schluessel, gefunden, erwartet)
		}
	}
}

func TestEinUnbekannterSchluesselHatKeinenBereich(t *testing.T) {
	t.Parallel()

	if gefunden := domain.EinstellungsBereich("erfunden"); gefunden != "" {
		t.Errorf("%q", gefunden)
	}
}
