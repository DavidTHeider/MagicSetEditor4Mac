package domain_test

// **Eine Vorgangskennung von außen** — OPS-05.AK-02
// (#1722):
// „Eine Korrelations-ID verbindet Frontend, REST/MCP, Service und
// Datenhaltung." Wer den Dienst ruft, darf seine eigene mitgeben — aber nur
// in der Gestalt, die dieses Programm selbst vergibt, und nicht beliebigen
// Text, der dann in jeder Protokollzeile stünde.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineVorgangskennungVonAussen(t *testing.T) {
	t.Parallel()

	for text, gut := range map[string]bool{
		"0a1b2c3d":                          true,
		"0123456789abcdef0123456789abcdef":  true,
		"0a1b2c3":                           false,
		"0123456789abcdef0123456789abcdef0": false,
		"0A1B2C3D":                          false,
		"0a1b2c3d\nboese":                   false,
		"":                                  false,
	} {
		kennung, ok := domain.VorgangskennungAus(text)
		if ok != gut || (gut && string(kennung) != text) {
			t.Errorf("%q: %q, %v", text, kennung, ok)
		}
	}

	if _, ok := domain.VorgangskennungAus(string(domain.NeueVorgangskennung())); !ok {
		t.Error("die eigene Gestalt wird abgewiesen")
	}
}

func TestDreiAusgaengeEinesRufs(t *testing.T) {
	t.Parallel()

	if len(domain.Rufausgaenge()) != 3 || domain.Rufausgaenge()[0] != domain.RufGelungen {
		t.Errorf("%v", domain.Rufausgaenge())
	}
}
