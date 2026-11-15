package mcp_test

// **Das Profilbild über MCP** — Werkstück 8 von Phase 4.12
// (#1351).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// NutzerbildSetzen merkt sich den Pfad und gibt den Nutzer mit Bild.
func (p *nutzerprobe) NutzerbildSetzen(_ context.Context, nutzerID, _, pfad string) (
	mcp.Nutzerzeile, error,
) {
	if p.fehler != nil {
		return mcp.Nutzerzeile{}, p.fehler
	}

	p.gebildet = pfad

	return mcp.Nutzerzeile{ID: nutzerID, Name: "Ich", Bild: "nutzer-" + nutzerID + ".png"}, nil
}

// NutzerbildEntfernen gibt den Nutzer ohne Bild.
func (p *nutzerprobe) NutzerbildEntfernen(_ context.Context, nutzerID, _ string) (
	mcp.Nutzerzeile, error,
) {
	if p.fehler != nil {
		return mcp.Nutzerzeile{}, p.fehler
	}

	p.gebildet = ""

	return mcp.Nutzerzeile{ID: nutzerID, Name: "Ich"}, nil
}

// **Das Werkzeug reicht den Pfad durch und gibt den Nutzer mit Bild.**
func TestDasProfilbildwerkzeugSetztEines(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{}

	antwort, fehler := mitNutzern(t, probe, "schreiben.nutzer.bild").Ruf(
		t.Context(), json.RawMessage(`{"nutzerID":"1","pfad":"/wo/gesicht.png","fassung":"1"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if probe.gebildet != "/wo/gesicht.png" {
		t.Errorf("gereicht wurde %q", probe.gebildet)
	}

	if !strings.Contains(antwort, "nutzer-1.png") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Und das zweite nimmt es weg.**
func TestDasProfilbildwerkzeugNimmtEinesWeg(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{gebildet: "/wo/gesicht.png"}

	antwort, fehler := mitNutzern(t, probe, "schreiben.nutzer.bild.loeschen").Ruf(
		t.Context(), json.RawMessage(`{"nutzerID":"1","fassung":"1"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if probe.gebildet != "" {
		t.Errorf("es blieb %q", probe.gebildet)
	}

	var gelesen struct {
		Nutzer mcp.Nutzerzeile `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("die Antwort lesen: %v", err)
	}

	if gelesen.Nutzer.Bild != "" {
		t.Errorf("der Nutzer trägt noch %q", gelesen.Nutzer.Bild)
	}
}

// **Beide reichen einen Fehler durch.**
func TestDieProfilbildwerkzeugeReichenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"schreiben.nutzer.bild":          `{"nutzerID":"1","pfad":"/x.png","fassung":"1"}`,
		"schreiben.nutzer.bild.loeschen": `{"nutzerID":"1","fassung":"1"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &nutzerprobe{fehler: errors.New("die Platte schweigt")}

			if !rufeMitFehler(t, mitNutzern(t, probe, name), argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und beide weisen kaputte Argumente ab.**
func TestDieProfilbildwerkzeugeWeisenKaputteArgumenteAb(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"schreiben.nutzer.bild", "schreiben.nutzer.bild.loeschen",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !rufeMitFehler(t, mitNutzern(t, &nutzerprobe{}, name), `{"nutzerID":`) {
				t.Error("der kaputte Rumpf kam durch")
			}
		})
	}
}

// ProfilbadgeSetzen merkt sich den Pfad und gibt das Profil mit Abzeichen.
func (p *nutzerprobe) ProfilbadgeSetzen(_ context.Context, profilID, _, pfad string) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.gebildet = pfad

	return mcp.Profilzeile{
		ID: profilID, NutzerID: "1", Name: "Mein Profil",
		Badge: "profil-" + profilID + ".png",
	}, nil
}

// ProfilbadgeEntfernen gibt das Profil ohne Abzeichen.
func (p *nutzerprobe) ProfilbadgeEntfernen(_ context.Context, profilID, _ string) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.gebildet = ""

	return mcp.Profilzeile{ID: profilID, NutzerID: "1", Name: "Mein Profil"}, nil
}

// **Das Werkzeug reicht den Pfad durch und gibt das Profil mit Abzeichen.**
func TestDasAbzeichenwerkzeugSetztEines(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{}

	antwort, fehler := mitNutzern(t, probe, "schreiben.profil.badge").Ruf(
		t.Context(), json.RawMessage(`{"profilID":"7","pfad":"/wo/wappen.png","fassung":"1"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if probe.gebildet != "/wo/wappen.png" {
		t.Errorf("gereicht wurde %q", probe.gebildet)
	}

	if !strings.Contains(antwort, "profil-7.png") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Und das zweite nimmt es weg.**
func TestDasAbzeichenwerkzeugNimmtEinesWeg(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{gebildet: "/wo/wappen.png"}

	if _, fehler := mitNutzern(t, probe, "schreiben.profil.badge.loeschen").Ruf(
		t.Context(), json.RawMessage(`{"profilID":"7","fassung":"1"}`)); fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if probe.gebildet != "" {
		t.Errorf("es blieb %q", probe.gebildet)
	}
}

// **Beide reichen einen Fehler durch und weisen kaputte Argumente ab.**
func TestDieAbzeichenwerkzeugeReichenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"schreiben.profil.badge":          `{"profilID":"7","pfad":"/x.png","fassung":"1"}`,
		"schreiben.profil.badge.loeschen": `{"profilID":"7","fassung":"1"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &nutzerprobe{fehler: errors.New("die Platte schweigt")}

			if !rufeMitFehler(t, mitNutzern(t, probe, name), argumente) {
				t.Error("der Fehler ging verloren")
			}

			if !rufeMitFehler(t, mitNutzern(t, &nutzerprobe{}, name), `{"profilID":`) {
				t.Error("der kaputte Rumpf kam durch")
			}
		})
	}
}
