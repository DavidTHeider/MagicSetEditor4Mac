package service_test

// **Was einen Aufbau anhält** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Der Schnitt folgt der Frage**: dort steht, was ein geglückter Aufbau tut;
// hier, was ihn anhält — ein Abbruch, ein fehlender Einbetter, ein
// gescheiterter Schritt. Und der Stapel, der geteilt wird, weil er sonst als
// ein Auftrag hinausginge.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein abgebrochener Aufbau verwirft seine Generation.**
//
// **Und aktiviert sie nicht** — das ist der Satz aus dem Issue: nie gemischte
// Vektoren.
func TestEinAbgebrochenerAufbauVerwirftSeineGeneration(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}
	kontext, abbrechen := context.WithCancel(t.Context())

	schreiber.vorSchreiben = abbrechen

	_, fehler := einAufbau(schreiber, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}).Aufbauen(kontext, einAufbauwunsch(), einKorpus())
	if !errors.Is(fehler, context.Canceled) {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if len(schreiber.aktiviert) != 0 {
		t.Errorf("aktiviert wurde %v", schreiber.aktiviert)
	}

	if len(schreiber.verworfen) != 1 {
		t.Errorf("verworfen wurde %v", schreiber.verworfen)
	}
}

// **Ein Aufbau ohne Einbetter findet nicht statt.**
//
// **Und zwar mit Ansage.** Ein Index ohne Vektoren wäre kein hybrider Index,
// sondern die Volltextsuche unter neuem Namen — die gibt es schon.
func TestEinAufbauOhneEinbetterFindetNichtStatt(t *testing.T) {
	t.Parallel()

	if _, fehler := einAufbau(&schreiberdoppel{}, nil).Aufbauen(
		t.Context(), einAufbauwunsch(), einKorpus(),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Aufbauen: %v", fehler)
	}
}

// **Scheitert etwas unterwegs, wird verworfen und der Stand sagt es.**
func TestEinGescheiterterAufbauVerwirftUndSagtEs(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		schreiber *schreiberdoppel
		korpus    *korpusdoppel
	}{
		"beim Schreiben": {
			&schreiberdoppel{schreibfehler: domain.MitGrund(
				domain.ErrConflict, nil, "die Platte ist voll")}, einKorpus(),
		},
		"beim Umschalten": {
			&schreiberdoppel{aktivfehler: domain.MitGrund(
				domain.ErrConflict, nil, "eine andere ist schon aktiv")}, einKorpus(),
		},
		"beim Lesen des Korpus": {
			&schreiberdoppel{}, &korpusdoppel{
				staende: einKorpus().staende,
				fehler: domain.MitGrund(domain.ErrNotFound, nil,
					"die Quelldatei fehlt"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			aufbau := einAufbau(fall.schreiber, &retrievaleinbetterdoppel{
				raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
			})

			if _, fehler := aufbau.Aufbauen(t.Context(), einAufbauwunsch(),
				fall.korpus); fehler == nil {
				t.Fatal("der Aufbau kam durch")
			}

			if len(fall.schreiber.verworfen) != 1 {
				t.Errorf("verworfen wurde %v", fall.schreiber.verworfen)
			}

			stand := aufbau.Stand()
			if stand.Laeuft || stand.Fehler == "" {
				t.Errorf("der Stand ist %+v", stand)
			}
		})
	}
}

// **Ein Stapel wird nicht am Stück eingebettet.**
//
// **Das Issue verlangt „begrenzte Batches"**, und der Grund steht in
// `internal/ollama`: ein Stapel wird als **ein** Auftrag geladen.
func TestEinGrosserStapelWirdGeteilt(t *testing.T) {
	t.Parallel()

	gross := make([]domain.Retrievaldokument, 0, service.StapelHoechstens+3)
	for i := range service.StapelHoechstens + 3 {
		gross = append(gross, einRetrievaltreffer(
			"karte"+string(rune('a'+i%26))+string(rune('a'+i/26)),
			1, domain.QuelleOracle).Dokument)
	}

	einbetter := &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}

	schreiber := &schreiberdoppel{}

	if _, fehler := einAufbau(schreiber, einbetter).Aufbauen(
		t.Context(), einAufbauwunsch(), &korpusdoppel{
			staende: einKorpus().staende,
			stapel:  [][]domain.Retrievaldokument{gross},
		}); fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if einbetter.groesster > service.StapelHoechstens {
		t.Errorf("ein Stapel war %d Texte groß", einbetter.groesster)
	}

	if len(schreiber.geschrieben) != len(gross) {
		t.Errorf("geschrieben wurden %d von %d",
			len(schreiber.geschrieben), len(gross))
	}
}
