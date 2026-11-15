package domain_test

// **Die Regelquelle** — Werkstück G4 von Phase 5.5
// (#1461).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Quelle ohne Dokument, ohne Fassung oder ohne Nummer belegt nichts.**
//
// # Warum jedes der drei einzeln geprüft wird
//
// **Weil jedes einzeln fehlen kann und jedes allein die Aussage kippt.** Eine
// Nummer ohne Fassung sagt nicht, welche 704 gemeint war; eine Fassung ohne
// Nummer sagt gar nichts; ein erfundenes Dokument sieht aus wie ein Beleg.
func TestEineRegelquelleBrauchtAlleDrei(t *testing.T) {
	t.Parallel()

	gut := domain.Regelquelle{
		Dokument: domain.DokumentCR,
		Fassung:  "2026-08-19",
		Nummern:  []string{"704"},
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("eine vollständige Quelle wird abgewiesen: %v", fehler)
	}

	faelle := map[string]domain.Regelquelle{
		"ohne Dokument":    {Fassung: "2026-08-19", Nummern: []string{"704"}},
		"fremdes Dokument": {Dokument: "IPG", Fassung: "x", Nummern: []string{"1"}},
		"ohne Fassung":     {Dokument: domain.DokumentCR, Nummern: []string{"704"}},
		"leere Fassung":    {Dokument: domain.DokumentCR, Fassung: "  ", Nummern: []string{"704"}},
		"ohne Nummer":      {Dokument: domain.DokumentCR, Fassung: "2026-08-19"},
		"leere Nummer":     {Dokument: domain.DokumentCR, Fassung: "2026-08-19", Nummern: []string{" "}},
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := quelle.Pruefe()
			if fehler == nil {
				t.Fatal("kommt durch")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}
		})
	}
}

// **Die Quelle liest sich wie die Herkunftsangabe im Quelltext.**
//
// **Dieselbe Form wie bei MSE2** (§11): Nummer **und** Fassung. Wer sie in
// einen Kommentar schreibt, schreibt dasselbe wie hier.
func TestEineRegelquelleLiestSichWieImQuelltext(t *testing.T) {
	t.Parallel()

	quelle := domain.Regelquelle{
		Dokument: domain.DokumentCR,
		Fassung:  "2026-08-19",
		Nummern:  []string{"400", "704"},
	}

	gesehen := quelle.String()
	if !strings.Contains(gesehen, "CR 400, 704") ||
		!strings.Contains(gesehen, "Fassung 2026-08-19") {
		t.Errorf("%q", gesehen)
	}
}

// **Die Aufzählung der Dokumente trennt.**
func TestIstRegeldokumentTrennt(t *testing.T) {
	t.Parallel()

	if len(domain.Regeldokumente()) != 2 {
		t.Errorf("%d Dokumente", len(domain.Regeldokumente()))
	}

	for _, gilt := range domain.Regeldokumente() {
		if !domain.IstRegeldokument(gilt) {
			t.Errorf("%q gilt nicht", gilt)
		}
	}

	if domain.IstRegeldokument("IPG") {
		t.Error("ein fremdes Dokument gilt")
	}
}
