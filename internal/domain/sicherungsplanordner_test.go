package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerOrdnerDesPlansIstEineEinstellungDesRechners(t *testing.T) {
	t.Parallel()

	if ebene, da := domain.EbeneVon(domain.SicherungsplanOrdner); !da || ebene != domain.EbeneGeraet {
		t.Errorf("Ebene %q (%v) — ein Ordner beschreibt diesen Rechner und reist nie", ebene, da)
	}

	if bereich := domain.EinstellungsBereich(domain.SicherungsplanOrdner); bereich != domain.BereichSicherung {
		t.Errorf("Bereich %q — er steht beim Sichern und nicht unter den Vorgaben", bereich)
	}

	if !domain.IstOrdnerangabe(domain.SicherungsplanOrdner) {
		t.Error("kein Ordner — dann fragte beim Setzen niemand die Platte, ob es ihn gibt")
	}

	if vorgabe := domain.EinstellungsVorgabe(domain.SicherungsplanOrdner); vorgabe != "" {
		t.Errorf("Vorgabe %q — der Plan ist freiwillig, ohne Ordner gibt es keinen", vorgabe)
	}

	for wert, gueltig := range map[string]bool{
		"": true, "~/Sicherungen": true, "/Volumes/Stick/Bestand": true,
		"Sicherungen": false, "/Volumes/../etc": false,
	} {
		if fehler := domain.PruefeEinstellung(domain.SicherungsplanOrdner, wert); (fehler == nil) != gueltig {
			t.Errorf("%q: %v", wert, fehler)
		}
	}
}
