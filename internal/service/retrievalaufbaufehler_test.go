package service_test

// **Die Fehlerwege des Aufbaus** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Jeder ist betreten worden, und keiner lässt einen halben Stand zurück**
// (§10, „Fehlerwege"). Die vier hier sind die, die `retrievalaufbau_test.go`
// nicht erreicht: ein besetzter Haushalt, eine Generation, die gar nicht
// entsteht, ein unvollständiges Manifest und ein Einbetter, der unterwegs
// aufgibt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einGutesDoppel gibt einen Einbetter, der nichts zu beanstanden hat.
func einGutesDoppel() *retrievaleinbetterdoppel {
	return &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}
}

// **Ein besetzter Haushalt hält den Aufbau auf — und der Riegel fällt wieder.**
//
// **Der zweite Teil ist der wichtige.** Ein Riegel, der bei einem
// abgewiesenen Platz stehen bliebe, sperrte jeden weiteren Aufbau für immer
// — und zwar ohne dass je einer gelaufen wäre.
func TestEinBesetzterHaushaltHaeltDenAufbauAuf(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel = map[domain.Ressourcenklasse]int{
		domain.RessourceIndexbau: 1,
	}

	haushalt := service.NeuerHaushalt(grenzen)

	freigeben, genommen := haushalt.Versuchen(domain.RessourceIndexbau)
	if !genommen {
		t.Fatal("der einzige Platz war schon weg")
	}

	schreiber := &schreiberdoppel{}
	aufbau := &aufbaupaar{
		aufbau:    service.NeuerIndexaufbau(schreiber, haushalt),
		einbetter: einGutesDoppel(),
	}

	kontext, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := aufbau.Aufbauen(kontext, einAufbauwunsch(),
		einKorpus()); !errors.Is(fehler, context.Canceled) {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if stand := aufbau.Stand(); stand.Laeuft {
		t.Errorf("der Riegel blieb liegen: %+v", stand)
	}

	freigeben()

	// **Und danach geht es wieder** — der Beleg, dass der Riegel wirklich
	// gefallen ist und nicht nur so aussieht.
	if _, fehler := aufbau.Aufbauen(t.Context(), einAufbauwunsch(),
		einKorpus()); fehler != nil {
		t.Fatalf("der zweite Aufbau: %v", fehler)
	}
}

// **Entsteht gar keine Generation, wird auch nichts verworfen.**
func TestOhneGenerationWirdNichtsVerworfen(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{
		anlegefehler: domain.MitGrund(domain.ErrConflict, nil,
			"der Bestand nimmt nichts an"),
	}

	if _, fehler := einAufbau(schreiber, einGutesDoppel()).Aufbauen(
		t.Context(), einAufbauwunsch(), einKorpus()); fehler == nil {
		t.Fatal("der Aufbau kam durch")
	}

	if len(schreiber.verworfen) != 0 {
		t.Errorf("verworfen wurde %v", schreiber.verworfen)
	}
}

// **Ein unvollständiges Manifest entsteht gar nicht erst.**
//
// **Geprüft wird, bevor angelegt wird.** Andersherum stünde eine Generation
// im Bestand, deren Manifest niemand deuten kann.
func TestEinUnvollstaendigesManifestEntstehtNicht(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wunsch service.Aufbauwunsch
		korpus *korpusdoppel
	}{
		"ohne Zerteilfassung": {
			service.Aufbauwunsch{Schemafassung: 1, Indexfassung: 1}, einKorpus(),
		},
		"ohne Korpusstand": {
			einAufbauwunsch(), &korpusdoppel{stapel: einKorpus().stapel},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schreiber := &schreiberdoppel{}

			if _, fehler := einAufbau(schreiber, einGutesDoppel()).Aufbauen(
				t.Context(), fall.wunsch, fall.korpus,
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("Aufbauen: %v", fehler)
			}

			if schreiber.naechste != 0 {
				t.Errorf("es wurde %d angelegt", schreiber.naechste)
			}
		})
	}
}

// **Ein unlesbarer Korpusstand bricht den Aufbau, bevor etwas entsteht.**
func TestEinUnlesbarerKorpusstandBrichtDenAufbau(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}

	if _, fehler := einAufbau(schreiber, einGutesDoppel()).Aufbauen(
		t.Context(), einAufbauwunsch(), &korpusdoppel{
			staendefehler: domain.MitGrund(domain.ErrNotFound, nil,
				"die Standdatei fehlt"),
		}); fehler == nil {
		t.Fatal("der Aufbau kam durch")
	}

	if schreiber.naechste != 0 {
		t.Errorf("es wurde %d angelegt", schreiber.naechste)
	}
}

// **Ein Einbetter, der unterwegs aufgibt, verwirft die Generation.**
func TestEinAufgebenderEinbetterVerwirftDieGeneration(t *testing.T) {
	t.Parallel()

	for name, einbetter := range map[string]*retrievaleinbetterdoppel{
		"er antwortet nicht": {
			raum:   einRetrievalmanifest().Raum(),
			fehler: domain.MitGrund(domain.ErrConflict, nil, "die Laufzeit schweigt"),
		},
		"er gibt einen Vektor zu viel": {
			raum:   einRetrievalmanifest().Raum(),
			vektor: []float32{1, 0, 0}, zweimal: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schreiber := &schreiberdoppel{}

			if _, fehler := einAufbau(schreiber, einbetter).Aufbauen(
				t.Context(), einAufbauwunsch(), einKorpus()); fehler == nil {
				t.Fatal("der Aufbau kam durch")
			}

			if len(schreiber.verworfen) != 1 {
				t.Errorf("verworfen wurde %v", schreiber.verworfen)
			}

			if len(schreiber.aktiviert) != 0 {
				t.Errorf("aktiviert wurde %v", schreiber.aktiviert)
			}
		})
	}
}
