package service_test

// **Was eine Zusage aus einer Dublette macht** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum das eine eigene Frage ist
//
// **Weil „überzählig" heißt: kann weg.** Ein Stück, das einem Handel zugesagt
// ist, kann nicht weg — es ist bereits versprochen. Wer es weiter als
// überzählig führte, böte es ein zweites Mal an.
//
// # Und was dabei **nicht** geändert wird
//
// **Die Dublettendefinition bleibt je Zeile** (0057): „mehr als ich behalten
// will", gerechnet über die vorhandene Zeile und nicht über eine Drucksumme.
// Das Issue sagt es ausdrücklich — die Zusage **zieht ab**, sie stellt nicht
// um.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zusagenprobe sagt, was je Exemplar festgehalten ist.
type zusagenprobe struct {
	fest  map[int64]int
	klagt error
}

func (z zusagenprobe) ReserviertJeExemplar(context.Context) (map[int64]int, error) {
	return z.fest, z.klagt
}

// **Zugesagte Stücke sind nicht überzählig.**
func TestEineZusageNimmtDerDubletteIhreUeberzahl(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{zeilen: []*domain.Exemplar{{
		ID: 7, Printing: "abc", Menge: 4, Zustand: domain.ZustandNearMint,
		Sprache: "en",
	}}}

	for name, fall := range map[string]struct {
		fest     int
		erwartet int
	}{
		"ohne Zusage":     {0, 3},
		"eine zugesagt":   {1, 2},
		"alles zugesagt":  {4, 0},
		"mehr als da ist": {9, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sichten := service.NeueSammlungssichten(ablage).MitReservierungen(
				zusagenprobe{fest: map[int64]int{7: fall.fest}})

			dubletten, err := sichten.Dubletten(context.Background(), 1)
			if err != nil {
				t.Fatalf("Dubletten: %v", err)
			}

			// **Null überzählig heißt: keine Dublette.** Eine Zeile mit
			// „überzählig 0" in der Liste wäre eine Aufforderung, die auf
			// nichts zeigt.
			if fall.erwartet == 0 {
				if len(dubletten) != 0 {
					t.Fatalf("es stehen %d Dubletten da: %+v", len(dubletten), dubletten)
				}

				return
			}

			if len(dubletten) != 1 || dubletten[0].Ueberzaehlig != fall.erwartet {
				t.Errorf("die Dubletten sind %+v", dubletten)
			}
		})
	}
}

// **Ohne Naht bleibt es, wie es war.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne diesen Lauf hielte auch eine
// Fassung, die immer null abzieht — und die Zusage wäre wieder unsichtbar.
func TestOhneNahtBleibtDieDubletteWieSieWar(t *testing.T) {
	t.Parallel()

	sichten := service.NeueSammlungssichten(&sammlungsdoppel{
		zeilen: []*domain.Exemplar{{
			ID: 7, Printing: "abc", Menge: 4, Zustand: domain.ZustandNearMint,
			Sprache: "en",
		}},
	})

	dubletten, err := sichten.Dubletten(context.Background(), 1)
	if err != nil {
		t.Fatalf("Dubletten: %v", err)
	}

	if len(dubletten) != 1 || dubletten[0].Ueberzaehlig != 3 {
		t.Errorf("die Dubletten sind %+v", dubletten)
	}
}

// **Was die Zusagen klagen, kommt zurück — und keine halbe Liste.**
//
// **Sonst stünde eine Dublettenliste da, in der die Zusagen fehlen**, und sie
// sähe aus wie eine, in der es keine gibt (§8).
func TestWasDieZusagenKlagenKommtZurueck(t *testing.T) {
	t.Parallel()

	klage := errors.New("die Platte ist voll")

	sichten := service.NeueSammlungssichten(&sammlungsdoppel{
		zeilen: []*domain.Exemplar{{
			ID: 7, Printing: "abc", Menge: 4, Zustand: domain.ZustandNearMint,
			Sprache: "en",
		}},
	}).MitReservierungen(zusagenprobe{klagt: klage})

	dubletten, fehler := sichten.Dubletten(context.Background(), 1)
	if !errors.Is(fehler, klage) {
		t.Fatalf("es kam %v", fehler)
	}

	if dubletten != nil {
		t.Errorf("es kam eine halbe Liste zurück: %+v", dubletten)
	}
}
