package mcp_test

// **Die Probe der Vorgabenpflege** — Werkstück 15 von Phase 4.12
// (#1351).
//
// **Sie hält ihren Stand**, weil die Idempotenzläufe zweimal gegen **denselben**
// Bestand rufen: eine Probe, die jedes Mal dasselbe erfindet, machte jedes
// Werkzeug idempotent (§10).

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// vorgabenprobe ist eine Vorgabenpflege im Arbeitsspeicher.
type vorgabenprobe struct {
	// gesetzt ist, worauf jede Art zeigt.
	gesetzt map[string][]string

	// moeglich ist, was es je Art gibt.
	moeglich map[string][]mcp.Vorgabeangebot

	// fehler ist der, den beide Wege stattdessen geben.
	fehler error
}

// arten sind die fünf, in Anzeigereihenfolge.
var arten = []string{"format", "spiel", "vorlage", "deck", "lagerort"}

func neueVorgabenprobe() *vorgabenprobe {
	return &vorgabenprobe{
		gesetzt: map[string][]string{},
		moeglich: map[string][]mcp.Vorgabeangebot{
			"deck": {{Wert: "1", Beschriftung: "Erstes (probe)"}},
		},
	}
}

func (p *vorgabenprobe) Vorgaben(_ context.Context, profilID string) (
	[]mcp.Vorgabenzeile, error,
) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	if profilID != "1" {
		return nil, fmt.Errorf("das Profil %s gibt es nicht", profilID)
	}

	zeilen := make([]mcp.Vorgabenzeile, 0, len(arten))

	for _, art := range arten {
		werte := p.gesetzt[art]
		if werte == nil {
			werte = []string{}
		}

		moeglich := p.moeglich[art]
		if moeglich == nil {
			moeglich = []mcp.Vorgabeangebot{}
		}

		zeilen = append(zeilen, mcp.Vorgabenzeile{
			Art:        art,
			Mehrwertig: art == "deck" || art == "lagerort",
			Werte:      werte,
			Moeglich:   moeglich,
			Verwaist:   []string{},
		})
	}

	return zeilen, nil
}

func (p *vorgabenprobe) VorgabeSetzen(
	ctx context.Context, profilID, art string, werte []string,
) (mcp.Vorgabengesetzt, error) {
	if p.fehler != nil {
		return mcp.Vorgabengesetzt{}, p.fehler
	}

	if !slices.Contains(arten, art) {
		return mcp.Vorgabengesetzt{}, fmt.Errorf("die Vorgabeart %q gibt es nicht", art)
	}

	// **Geprüft wird gegen dasselbe Angebot, das oben herausgeht** — eine
	// Probe, die alles annimmt, prüfte nicht die Zusage, sondern sich selbst.
	for _, wert := range werte {
		if !slices.ContainsFunc(p.moeglich[art], func(a mcp.Vorgabeangebot) bool {
			return a.Wert == wert
		}) {
			return mcp.Vorgabengesetzt{}, fmt.Errorf("für %q gibt es %q nicht", art, wert)
		}
	}

	p.gesetzt[art] = werte

	zeilen, fehler := p.Vorgaben(ctx, profilID)

	return mcp.Vorgabengesetzt{Vorgaben: zeilen}, fehler
}
