package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Welche Einstellung bestimmt, wohin gesendet wird** — SEC-P4
// (#2000).
//
// **Jede Adresse, die ein Dienst tragen darf, gehört dazu** — und keine
// darüber hinaus: ein Kartendienst hat keine einstellbare Adresse, und eine
// Sprache bestimmt kein Ziel.
func TestEinNetzwegIstWasDasZielBestimmt(t *testing.T) {
	t.Parallel()

	faelle := map[string]bool{
		domain.KIAdresse:      true,
		domain.NetzProxyHTTP:  true,
		domain.NetzProxyHTTPS: true,
		domain.NetzProxyAus:   true,

		domain.EinstellungSprache:                   false,
		domain.KILaufzeit:                           false,
		domain.NetzFristKopf:                        false,
		domain.DienstAdresse(domain.DienstScryfall): false,
		"unbekannt":                                 false,
	}

	for _, dienst := range domain.Dienste() {
		faelle[domain.DienstAdresse(dienst)] = dienst.AdresseEinstellbar()
	}

	if !faelle[domain.DienstAdresse(domain.DienstAnthropic)] {
		t.Fatal("die Anthropic-Adresse fehlt unter den einstellbaren — die Tafel prüfte nichts")
	}

	for schluessel, erwartet := range faelle {
		if domain.IstNetzweg(schluessel) != erwartet {
			t.Errorf("IstNetzweg(%q) = %v, erwartet %v", schluessel, !erwartet, erwartet)
		}
	}
}
