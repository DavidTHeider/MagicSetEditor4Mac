package domain_test

// **Die Grenze des Zwischenspeichers auf der Platte** — OPS-08.AK-03
// (#1725).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieGrenzeDesZwischenspeichers(t *testing.T) {
	t.Parallel()

	if ebene, _ := domain.EbeneVon(domain.ZwischenspeicherGrenze); ebene != domain.EbeneGeraet ||
		domain.EinstellungsWirkung(domain.ZwischenspeicherGrenze) != domain.WirkungNeustart {
		t.Errorf("Ebene %q", ebene)
	}

	for wert, bytes := range map[string]int64{
		"": 512 << 20, "1": 1 << 20, "8192": 8192 << 20, "0": 512 << 20, "viel": 512 << 20, "8193": 512 << 20,
	} {
		if ist := domain.ZwischenspeichergrenzeAus(wert); ist != bytes {
			t.Errorf("%q: %d", wert, ist)
		}
	}

	if domain.PruefeEinstellung(domain.ZwischenspeicherGrenze, "0") == nil ||
		domain.PruefeEinstellung(domain.ZwischenspeicherGrenze, "64") != nil {
		t.Error("die Prüfung kennt die Grenzen nicht")
	}
}
