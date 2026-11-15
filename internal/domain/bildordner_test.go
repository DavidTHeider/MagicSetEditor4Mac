package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestDerOrdnerDerBilderIstEinGemerkterOrdnerDesRechners hält AP-10
// (#2367): der zuletzt gewählte Ordner der Bildausgabe wird gemerkt wie in
// MSE2 `settings.default_export_dir` — einer je Installation, ein Pfad
// dieser Platte.
func TestDerOrdnerDerBilderIstEinGemerkterOrdnerDesRechners(t *testing.T) {
	t.Parallel()

	if ebene, da := domain.EbeneVon(domain.EinstellungBildordner); !da || ebene != domain.EbeneGeraet {
		t.Errorf("Ebene %q (%v) — ein Ordner beschreibt diesen Rechner und reist nie", ebene, da)
	}

	if bereich := domain.EinstellungsBereich(domain.EinstellungBildordner); bereich != domain.BereichBildausfuhr {
		t.Errorf("Bereich %q — er steht bei der Bildausgabe", bereich)
	}

	if !domain.IstOrdnerangabe(domain.EinstellungBildordner) {
		t.Error("kein Ordner — dann fragte beim Setzen niemand die Platte, ob es ihn gibt")
	}

	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungBildordner); vorgabe != "" {
		t.Errorf("Vorgabe %q — ohne Wahl ist noch kein Ordner gemerkt", vorgabe)
	}

	for wert, gueltig := range map[string]bool{
		"": true, "~/Karten/Tischprobe": true, "/Volumes/Stick/Bilder": true,
		"Bilder": false, "/Volumes/../etc": false,
	} {
		if err := domain.PruefeEinstellung(domain.EinstellungBildordner, wert); (err == nil) != gueltig {
			t.Errorf("%q: %v", wert, err)
		}
	}
}
