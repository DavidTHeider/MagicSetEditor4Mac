package domain_test

// **Ein Nachweis bindet, was er erlaubt** — Werkstück O1 von Phase 5.2
// (#1425).
//
// # Was hier geprüft wird
//
// Ein Nachweis gilt für **einen** Nutzer, **ein** Profil, **einen** Zweck,
// **ein** Ziel und **einen** Stand — und einmal. Was davon abweicht, weist er
// ab, und die Absage sagt, was nicht passte.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var jetzt = time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

func nachweis() domain.Nachweis {
	return domain.Nachweis{
		Kennung: "n-1", NutzerID: 1, ProfilID: 7,
		Zweck: "handel.angebot", Ziel: "angebot-42", Revision: 3,
		Grenze:      domain.Kostengrenze{Betrag: 1500, Waehrung: "EUR"},
		Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(10 * time.Minute),
	}
}

func bindung() domain.Nachweisbindung {
	return domain.Nachweisbindung{
		NutzerID: 1, ProfilID: 7, Zweck: "handel.angebot", Ziel: "angebot-42",
		Revision: 3, Kosten: 1200, Waehrung: "EUR",
	}
}

func TestEinNachweisGiltGenauFuerSeineBindung(t *testing.T) {
	t.Parallel()

	if fehler := nachweis().Gilt(bindung(), jetzt); fehler != nil {
		t.Fatalf("der passende Nachweis wird abgewiesen: %v", fehler)
	}

	faelle := map[string]func(*domain.Nachweis, *domain.Nachweisbindung){
		"anderer Nutzer":    func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.NutzerID = 2 },
		"anderes Profil":    func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.ProfilID = 8 },
		"anderer Zweck":     func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.Zweck = "ki.auftrag" },
		"anderes Ziel":      func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.Ziel = "angebot-43" },
		"anderer Stand":     func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.Revision = 4 },
		"über der Grenze":   func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.Kosten = 1501 },
		"andere Währung":    func(_ *domain.Nachweis, b *domain.Nachweisbindung) { b.Waehrung = "USD" },
		"schon eingelöst":   func(n *domain.Nachweis, _ *domain.Nachweisbindung) { n.Eingeloest = true },
		"abgelaufen":        func(n *domain.Nachweis, _ *domain.Nachweisbindung) { n.Ablauf = jetzt.Add(-time.Second) },
		"noch nicht gültig": func(n *domain.Nachweis, _ *domain.Nachweisbindung) { n.Ausgestellt = jetzt.Add(time.Minute) },
	}

	for name, verschiebe := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			n, b := nachweis(), bindung()
			verschiebe(&n, &b)

			fehler := n.Gilt(b, jetzt)
			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Fatalf("%s: %v, erwartet ErrErlaubnis", name, fehler)
			}
		})
	}
}

// **Ohne Grenze ist jeder Betrag erlaubt** — die Grenze ist eine Wahl beim
// Ausstellen, keine Pflicht. Und ohne Kosten fragt niemand nach der Währung.
func TestOhneGrenzeZaehltDerBetragNicht(t *testing.T) {
	t.Parallel()

	n := nachweis()
	n.Grenze = domain.Kostengrenze{}

	b := bindung()
	b.Kosten = 999999
	b.Waehrung = "USD"

	if fehler := n.Gilt(b, jetzt); fehler != nil {
		t.Fatalf("ohne Grenze abgewiesen: %v", fehler)
	}

	n = nachweis()
	b = bindung()
	b.Kosten = 0
	b.Waehrung = ""

	if fehler := n.Gilt(b, jetzt); fehler != nil {
		t.Fatalf("ohne Kosten abgewiesen: %v", fehler)
	}
}

// **Ein Nachweis ohne Zweck oder Ziel ist keiner** — er würde für alles
// gelten, und genau das soll er nicht.
func TestEinNachweisBrauchtZweckUndZiel(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Nachweis){
		"ohne Zweck":     func(n *domain.Nachweis) { n.Zweck = "" },
		"ohne Ziel":      func(n *domain.Nachweis) { n.Ziel = " " },
		"ohne Nutzer":    func(n *domain.Nachweis) { n.NutzerID = 0 },
		"ohne Profil":    func(n *domain.Nachweis) { n.ProfilID = 0 },
		"Ablauf vor Aus": func(n *domain.Nachweis) { n.Ablauf = n.Ausgestellt.Add(-time.Second) },
		"negative Grenze": func(n *domain.Nachweis) {
			n.Grenze = domain.Kostengrenze{Betrag: -1, Waehrung: "EUR"}
		},
		"Grenze ohne Währung": func(n *domain.Nachweis) {
			n.Grenze = domain.Kostengrenze{Betrag: 5, Waehrung: ""}
		},
	}

	if fehler := nachweis().Pruefe(); fehler != nil {
		t.Fatalf("der gültige Nachweis wird abgewiesen: %v", fehler)
	}

	for name, verschiebe := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			n := nachweis()
			verschiebe(&n)

			if fehler := n.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%s: %v, erwartet ErrValidation", name, fehler)
			}
		})
	}
}
