// **Die Maschinenidentität im Kontext** — Befund 5 des Zweitblicks über
// Phase 5.2 (#1435).
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was hineingeht, kommt heraus — und ohne sie ist sie leer.**
func TestDieTokenidentitaetReistImKontext(t *testing.T) {
	t.Parallel()

	if identitaet := service.TokenidentitaetAus(t.Context()); identitaet != "" {
		t.Errorf("ohne Token steht %q da", identitaet)
	}

	mit := service.MitTokenidentitaet(t.Context(), "7:Rechner")

	if identitaet := service.TokenidentitaetAus(mit); identitaet != "7:Rechner" {
		t.Errorf("die Identität ist %q", identitaet)
	}
}
