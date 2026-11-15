package mcp_test

// **Die Probe der Freigabenpflege** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Sie hält ihren Stand**, weil die Idempotenzläufe zweimal gegen **denselben**
// Bestand rufen: eine Probe, die jedes Mal dasselbe erfindet, machte jedes
// Werkzeug idempotent (§10).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// freigabenprobe ist eine Freigabenpflege im Arbeitsspeicher.
type freigabenprobe struct {
	// abgelegt ist die Rolle je Nutzerkennung.
	abgelegt map[string]string

	// rolle ist die eigene — **`verwalter`**, sonst käme kein Lauf durch.
	rolle string

	fehler error
}

func neueFreigabenprobe() *freigabenprobe {
	return &freigabenprobe{abgelegt: map[string]string{}, rolle: "verwalter"}
}

func (p *freigabenprobe) Freigaben(_ context.Context, profilID string) (
	mcp.Freigabenstand, error,
) {
	if p.fehler != nil {
		return mcp.Freigabenstand{}, p.fehler
	}

	if profilID != "1" {
		return mcp.Freigabenstand{}, fmt.Errorf(
			"das Profil %s gibt es nicht", profilID)
	}

	return p.stand(), nil
}

func (p *freigabenprobe) FreigabeSetzen(
	ctx context.Context, profilID, nutzerID, rolle string,
) (mcp.Freigabenstand, error) {
	if p.fehler != nil {
		return mcp.Freigabenstand{}, p.fehler
	}

	// **Geprüft wird gegen dieselbe Rollenliste, die auch herausgeht** — eine
	// Probe, die alles annimmt, prüfte nicht die Zusage, sondern sich selbst.
	if rolle != "leser" && rolle != "mitarbeiter" && rolle != "verwalter" {
		return mcp.Freigabenstand{}, fmt.Errorf("die Rolle %q gibt es nicht", rolle)
	}

	p.abgelegt[nutzerID] = rolle

	return p.Freigaben(ctx, profilID)
}

func (p *freigabenprobe) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) (mcp.Freigabenstand, error) {
	if p.fehler != nil {
		return mcp.Freigabenstand{}, p.fehler
	}

	delete(p.abgelegt, nutzerID)

	return p.Freigaben(ctx, profilID)
}

// stand baut die Antwort aus dem, was abgelegt ist.
func (p *freigabenprobe) stand() mcp.Freigabenstand {
	zeilen := make([]mcp.Freigabezeile, 0, len(p.abgelegt))

	for nutzerID, rolle := range p.abgelegt {
		zeilen = append(zeilen, mcp.Freigabezeile{
			NutzerID: nutzerID, Name: "Zweiter", Rolle: rolle,
			Angelegt: "2026-09-06T00:00:00Z",
		})
	}

	return mcp.Freigabenstand{
		Freigaben: zeilen, Sichtbarkeit: "privat", EigeneRolle: p.rolle,
	}
}
