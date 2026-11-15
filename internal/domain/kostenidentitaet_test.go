// **Die Identität eines Auftrags** — Befund 8 des Zweitblicks über Phase 5.2
// (#1435).
package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probeposten ist der Auftrag, den jeder Fall abwandelt.
func probeposten() domain.Kostenposten {
	return domain.Kostenposten{
		Kennung: "eins", ProfilID: 1, Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 300, Waehrung: "EUR"},
	}
}

// **Derselbe Auftrag ist derselbe — auch mit anderem Stand und Betrag.**
func TestDerselbeAuftragMeintDasselbe(t *testing.T) {
	t.Parallel()

	laeuft := probeposten()
	laeuft.Stand = domain.KostenAbgerechnet
	laeuft.Betrag = domain.Betrag{Wert: 120, Waehrung: "EUR"}
	laeuft.Grund = "fertig"

	if fehler := laeuft.MeintDasselbeWie(probeposten()); fehler != nil {
		t.Errorf("derselbe Auftrag gilt als anderer: %v", fehler)
	}
}

// **Und ein anderer ist ein Konflikt — in jedem der vier Felder.**
func TestEinAndererAuftragIstEinKonflikt(t *testing.T) {
	t.Parallel()

	for name, anders := range map[string]func(domain.Kostenposten) domain.Kostenposten{
		"Profil":       func(p domain.Kostenposten) domain.Kostenposten { p.ProfilID = 2; return p },
		"Konto":        func(p domain.Kostenposten) domain.Kostenposten { p.Konto = "woanders"; return p },
		"Aufgabe":      func(p domain.Kostenposten) domain.Kostenposten { p.Aufgabe = "bild"; return p },
		"Höchstbetrag": func(p domain.Kostenposten) domain.Kostenposten { p.Hoechstbetrag.Wert = 900; return p },
		"Währung":      func(p domain.Kostenposten) domain.Kostenposten { p.Hoechstbetrag.Waehrung = "USD"; return p },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := probeposten().MeintDasselbeWie(anders(probeposten()))
			if !errors.Is(fehler, domain.ErrConflict) {
				t.Errorf("abgewiesen wurde mit %v", fehler)
			}
		})
	}
}
