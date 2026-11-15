package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// textmodell ist ein Modell, das in fester Form antworten kann — oder nicht.
func textmodell(alias string, fest bool) domain.KIModell {
	modell := domain.KIModell{Alias: alias, Anbieter: "ollama", Laufzeit: "ollama 0.35.1"}
	if fest {
		modell.Geprueft = []domain.KIGeprueft{{
			Faehigkeit: domain.KIStrukturierteGenerierung, Datum: "2026-10-04", Beleg: "Probe",
		}}
	}

	return modell
}

// **Welches Modell der Vorrang fragt** — #2054, aus AI-02 (#1743): ohne
// Wahl das erste, das in fester Form antworten kann, und das ist bei Ollama
// das zuletzt gezogene. Mit Wahl das gewählte — und fehlt es, wird das
// gesagt und nicht still das erste genommen.
func TestDasVorrangmodell(t *testing.T) {
	t.Parallel()

	modelle := []domain.KIModell{
		textmodell("embeddinggemma:300m", false),
		textmodell("qwen3:0.6b", true),
		textmodell("gemma3:1b", true),
	}

	for _, fall := range []struct {
		name, gewaehlt, erwartet string
		fehler                   error
	}{
		{name: "ohne Wahl das erste in fester Form", erwartet: "qwen3:0.6b"},
		{name: "mit Wahl das gewählte", gewaehlt: "gemma3:1b", erwartet: "gemma3:1b"},
		{
			name: "ohne Marke heißt :latest und nicht :1b", gewaehlt: "gemma3",
			fehler: domain.ErrNotFound,
		},
		{
			name: "ein gewähltes, das fehlt, wird gemeldet", gewaehlt: "mistral:7b",
			fehler: domain.ErrNotFound,
		},
		{
			name: "ein gewähltes ohne feste Form wird gemeldet", gewaehlt: "embeddinggemma:300m",
			fehler: domain.ErrValidation,
		},
		{name: "Leerraum um die Wahl zählt nicht", gewaehlt: "  gemma3:1b ", erwartet: "gemma3:1b"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			alias, fehler := domain.KIVorrangmodellAus(modelle, fall.gewaehlt)

			if fall.fehler != nil {
				if !errors.Is(fehler, fall.fehler) {
					t.Fatalf("Fehler %v, erwartet %v", fehler, fall.fehler)
				}

				if !strings.Contains(fehler.Error(), strings.TrimSpace(fall.gewaehlt)) {
					t.Errorf("die Meldung %q nennt das gewählte Modell nicht", fehler.Error())
				}

				return
			}

			if fehler != nil || alias != fall.erwartet {
				t.Errorf("%q, %v — erwartet %q", alias, fehler, fall.erwartet)
			}
		})
	}
}

// **Eine Wahl mit Marke :latest trifft auch den Namen ohne**, wie Ollama
// selbst es tut — `gemma3` und `gemma3:latest` sind dasselbe Modell.
func TestDasVorrangmodellKenntLatest(t *testing.T) {
	t.Parallel()

	modelle := []domain.KIModell{textmodell("gemma3:latest", true)}

	for _, gewaehlt := range []string{"gemma3", "gemma3:latest"} {
		if alias, fehler := domain.KIVorrangmodellAus(modelle, gewaehlt); fehler != nil || alias != "gemma3:latest" {
			t.Errorf("%q: %q, %v", gewaehlt, alias, fehler)
		}
	}
}

// **Ohne ein Modell in fester Form nennt die Regel keines** — und keinen
// Fehler: was dann zu sagen ist, sagt der Aufrufer, der den Anbieter kennt.
func TestOhneFesteFormKeinVorrangmodell(t *testing.T) {
	t.Parallel()

	alias, fehler := domain.KIVorrangmodellAus([]domain.KIModell{textmodell("x", false)}, "")
	if alias != "" || fehler != nil {
		t.Errorf("%q, %v", alias, fehler)
	}
}

// **Die Einstellung nimmt einen Modellnamen oder nichts.**
func TestDieWahlDesVorrangmodells(t *testing.T) {
	t.Parallel()

	for _, gut := range []string{"", "gemma3:1b", "x/flux2-klein:4b", "hf.co/org/modell:Q4_K_M"} {
		if fehler := domain.PruefeKIVorrangmodell(gut); fehler != nil {
			t.Errorf("%q abgewiesen: %v", gut, fehler)
		}
	}

	for _, schlecht := range []string{"gemma 3", "gemma3\n", "gemma3\x00", strings.Repeat("a", 201)} {
		if fehler := domain.PruefeKIVorrangmodell(schlecht); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%q angenommen: %v", schlecht, fehler)
		}
	}
}
