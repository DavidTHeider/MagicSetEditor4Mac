package mcp_test

// **Die Einstellungen eines Menschen und seine Auskunft über MCP** —
// Werkstück 8 von Phase 4.12
// (#1351).

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// Nutzereinstellungen gibt die Reihe aus der Domäne samt Gesetztem.
func (p *nutzerprobe) Nutzereinstellungen(context.Context, string) (
	[]mcp.Nutzereinstellungszeile, error,
) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	return p.zeilenAusDerReihe(), nil
}

func (p *nutzerprobe) NutzereinstellungSetzen(_ context.Context, _, schluessel, wert string) (
	[]mcp.Nutzereinstellungszeile, error,
) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	if p.gesetzteWerte == nil {
		p.gesetzteWerte = map[string]string{}
	}

	p.gesetzteWerte[schluessel] = wert

	return p.zeilenAusDerReihe(), nil
}

func (p *nutzerprobe) Personendaten(context.Context, string) (string, error) {
	if p.fehler != nil {
		return "", p.fehler
	}

	return `{"erhoben":"2026-09-05T00:00:00Z","nutzer":{"name":"Ich"}}`, nil
}

// zeilenAusDerReihe baut die Reihe der Domäne mit dem, was gesetzt wurde.
//
// **Aus der Domäne und nicht aus einer Kopie**: ein Doppel mit eigenen
// Schlüsseln prüfte etwas, das es im Betrieb nicht gibt (§3).
func (p *nutzerprobe) zeilenAusDerReihe() []mcp.Nutzereinstellungszeile {
	reihe := domain.Nutzereinstellungsreihe()
	zeilen := make([]mcp.Nutzereinstellungszeile, 0, len(reihe))

	for _, schluessel := range reihe {
		wert, gesetzt := p.gesetzteWerte[schluessel]
		if !gesetzt {
			wert = domain.NutzerVorgabe(schluessel)
		}

		zeilen = append(zeilen, mcp.Nutzereinstellungszeile{
			Schluessel: schluessel, Wert: wert,
			Werte: domain.ErlaubteNutzerwerte(schluessel),
		})
	}

	return zeilen
}

// **Das lesende Werkzeug nennt jede Einstellung samt ihren Werten.**
func TestDasNutzereinstellungswerkzeugNenntAlle(t *testing.T) {
	t.Parallel()

	antwort, fehler := mitNutzern(t, &nutzerprobe{}, "lesen.nutzereinstellungen").
		Ruf(t.Context(), json.RawMessage(`{"nutzerID":"1"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	var gelesen struct {
		Einstellungen []mcp.Nutzereinstellungszeile `json:"einstellungen"`
	}

	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("die Antwort lesen: %v", err)
	}

	if len(gelesen.Einstellungen) != len(domain.Nutzereinstellungsreihe()) {
		t.Fatalf("es kamen %d", len(gelesen.Einstellungen))
	}

	// **Gegen die Domäne und nicht gegen »nicht leer«** — Werkstück O3 von
	// Phase 5.2 (#1427):
	// eine Spaltenbreite hat keine Werteliste, und für sie wäre „nicht leer"
	// falsch. Der Vergleich ist die **schärfere** Zusage — er sieht auch eine
	// Liste, die etwas anderes enthält als die Domäne erlaubt.
	for _, eine := range gelesen.Einstellungen {
		erlaubt := domain.ErlaubteNutzerwerte(eine.Schluessel)
		if !slices.Equal(eine.Werte, erlaubt) {
			t.Errorf("%q nennt %v statt %v — ein Modell müsste raten",
				eine.Schluessel, eine.Werte, erlaubt)
		}
	}
}

// **Das schreibende setzt eine und gibt alle zurück.**
func TestDasNutzereinstellungswerkzeugSetztEine(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{}

	antwort, fehler := mitNutzern(t, probe, "schreiben.nutzereinstellung").
		Ruf(t.Context(), json.RawMessage(
			`{"nutzerID":"1","schluessel":"`+domain.NutzerKontrast+
				`","wert":"`+domain.KontrastHoch+`"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if probe.gesetzteWerte[domain.NutzerKontrast] != domain.KontrastHoch {
		t.Errorf("gesetzt wurde %v", probe.gesetzteWerte)
	}

	if !strings.Contains(antwort, domain.KontrastHoch) {
		t.Errorf("die Antwort trägt den Stand nicht: %s", antwort)
	}
}

// **Die Auskunft kommt als Dokument.**
func TestDasAuskunftswerkzeugGibtDasDokument(t *testing.T) {
	t.Parallel()

	antwort, fehler := mitNutzern(t, &nutzerprobe{}, "lesen.personendaten").
		Ruf(t.Context(), json.RawMessage(`{"nutzerID":"1"}`))
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	var gelesen map[string]any
	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v — %s", err, antwort)
	}

	if gelesen["erhoben"] == nil {
		t.Error("der Stichtag fehlt")
	}
}

// **Alle drei reichen einen Fehler durch und weisen kaputte Argumente ab.**
func TestDieEinstellungswerkzeugeReichenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"lesen.nutzereinstellungen": `{"nutzerID":"1"}`,
		"lesen.personendaten":       `{"nutzerID":"1"}`,
		"schreiben.nutzereinstellung": `{"nutzerID":"1","schluessel":"` +
			domain.NutzerKontrast + `","wert":"` + domain.KontrastHoch + `"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &nutzerprobe{fehler: errors.New("die Platte schweigt")}

			if !rufeMitFehler(t, mitNutzern(t, probe, name), argumente) {
				t.Error("der Fehler ging verloren")
			}

			if !rufeMitFehler(t, mitNutzern(t, &nutzerprobe{}, name), `{"nutzerID":`) {
				t.Error("der kaputte Rumpf kam durch")
			}
		})
	}
}
