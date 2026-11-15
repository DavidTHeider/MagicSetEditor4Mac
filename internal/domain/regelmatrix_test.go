package domain_test

// **Die Unterstützungsmatrix** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass „unbekannt" die Vorgabe ist.** Das Issue verlangt es wörtlich:
// „Unbekannt/ununterstützt ist ein eigener Zustand, niemals still erlaubt
// oder gültig."
//
// **Dass Strenge nur trägt, wo jeder Bereich durchgesetzt wird** — eine
// Zwischenstufe gibt es nicht.
//
// **Und dass eine Zusage ihre Grenzen nennt.** Ein Bereich, der als
// `automatisch` dasteht und keine Grenze nennt, behauptet Vollständigkeit.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineQuelle gibt eine gültige Regelquelle.
func eineQuelle(nummern ...string) domain.Regelquelle {
	return domain.Regelquelle{
		Dokument: domain.DokumentCR,
		Fassung:  "2026-08-19",
		Nummern:  nummern,
	}
}

// volleMatrix gibt eine Matrix, in der jeder Bereich durchgesetzt wird.
func volleMatrix() domain.Regelmatrix {
	eintraege := make([]domain.Matrixeintrag, 0, len(domain.Regelbereiche()))

	for _, bereich := range domain.Regelbereiche() {
		eintraege = append(eintraege, domain.Matrixeintrag{
			Bereich: bereich,
			Grad:    domain.UnterstuetzungAutomatisch,
			Quelle:  eineQuelle("704"),
			Grenzen: []string{"keine Karte mit eigenem Text"},
		})
	}

	return domain.Regelmatrix{Eintraege: eintraege}
}

// **Ein Bereich ohne Eintrag ist unbekannt — nicht erlaubt und nicht
// verboten.**
func TestEinBereichOhneEintragIstUnbekannt(t *testing.T) {
	t.Parallel()

	leer := domain.Regelmatrix{}

	for _, bereich := range domain.Regelbereiche() {
		if grad := leer.Grad(bereich); grad != domain.UnterstuetzungUnbekannt {
			t.Errorf("%q ist %q, erwartet unbekannt", bereich, grad)
		}
	}

	if _, da := leer.Eintrag(domain.BereichZonen); da {
		t.Error("eine leere Matrix hat einen Eintrag")
	}

	// **Und sie trägt nur den Sandkasten.**
	switch {
	case !leer.TraegtStrenge(domain.StrengeSandkasten):
		t.Error("der Sandkasten wird nicht getragen")
	case leer.TraegtStrenge(domain.StrengeBegleitet):
		t.Error("eine leere Matrix trägt begleitet")
	case leer.TraegtStrenge(domain.StrengeStreng):
		t.Error("eine leere Matrix trägt streng")
	}
}
