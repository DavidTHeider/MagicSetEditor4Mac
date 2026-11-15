// **Die Vorrichtung für Sammlung und Preise** — Werkstück 2 und 3 von
// Phase 3.4 (#1165).
//
// # Sie ist so streng wie der Betrieb
//
// Eine Änderung an einer Zeile, die es nicht gibt, scheitert hier ebenso wie
// im Bestand — und ein zweites Anlegen desselben Exemplars auch. **Eine
// Vorrichtung, die großzügiger ist als der Betrieb, prüft den falschen Weg**
// (§3), und bei einer Zusage über Idempotenz ist genau das der Unterschied
// zwischen einem Beleg und einem Selbstgespräch.
package mcp_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sammlungsprobe ist eine Sammlung im Arbeitsspeicher.
type sammlungsprobe struct {
	zeilen   []mcp.Exemplarsicht
	naechste int64
}

// neueSammlungsprobe gibt eine Sammlung mit einem Exemplar.
func neueSammlungsprobe() *sammlungsprobe {
	return &sammlungsprobe{
		zeilen: []mcp.Exemplarsicht{{
			ID: 1, Fassung: 1, Printing: "probe-druck", Menge: 3,
			Zustand: "nearMint", Sprache: "en",
		}},
		naechste: 1,
	}
}

func (s *sammlungsprobe) Sammlung(context.Context) ([]mcp.Exemplarsicht, error) {
	return s.zeilen, nil
}

// **Der Ausschnitt** — Werkstück
// #1500.
// Die Probe gibt, was sie hat, und zählt es zweimal: die Sicht ist hier
// keine, und die Zahlen sollen trotzdem dastehen.
func (s *sammlungsprobe) Sammlungsfenster(
	_ context.Context, _, _, _ string, ab, anzahl int,
) (fenster []mcp.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	alle := domain.Fensterzahlen{Gesamt: len(s.zeilen), Insgesamt: len(s.zeilen), Stueck: len(s.zeilen)}

	if ab >= len(s.zeilen) {
		return []mcp.Exemplarsicht{}, alle, nil
	}

	fenster = s.zeilen[ab:]
	if anzahl > 0 && anzahl < len(fenster) {
		fenster = fenster[:anzahl]
	}

	return fenster, alle, nil
}

func (s *sammlungsprobe) SammlungFuer(_ context.Context, printing string) (
	[]mcp.Exemplarsicht, error,
) {
	gefunden := []mcp.Exemplarsicht{}

	for _, zeile := range s.zeilen {
		if zeile.Printing == printing {
			gefunden = append(gefunden, zeile)
		}
	}

	return gefunden, nil
}

func (s *sammlungsprobe) ExemplarAnlegen(_ context.Context, exemplar mcp.Exemplarsicht) (
	int64, error,
) {
	// **So streng wie der Bestand — und zwar über alle fünf Spalten**:
	// dieselbe Karte in demselben Zustand, derselben Sprache, derselben
	// Ausführung **und derselben Absicht** gibt es einmal.
	//
	// Die fünfte stand hier nicht, und die Vorrichtung war damit **strenger**
	// als der Bestand: sie wies einen Wunsch auf einen besessenen Druck ab,
	// den `UNIQUE (printing, zustand, sprache, foil, wunsch)` erlaubt
	// ([0026_ordnung.sql](https://github.com/DavidTHeider/MagicSetEditor4Mac/blob/v0.7.0/internal/storage/migrationen/0026_ordnung.sql), Vermerk
	// 0057). Angestrichen vom Nachlauf über Phase 4.5.
	for _, zeile := range s.zeilen {
		if zeile.Printing == exemplar.Printing && zeile.Zustand == exemplar.Zustand &&
			zeile.Sprache == exemplar.Sprache && zeile.Foil == exemplar.Foil &&
			zeile.Wunsch == exemplar.Wunsch {
			return 0, domain.ErrConflict
		}
	}

	s.naechste++
	exemplar.ID = s.naechste
	exemplar.Fassung = 1
	s.zeilen = append(s.zeilen, exemplar)

	return exemplar.ID, nil
}

func (s *sammlungsprobe) ExemplarAendern(_ context.Context, exemplar mcp.Exemplarsicht) error {
	for i, zeile := range s.zeilen {
		if zeile.ID == exemplar.ID && zeile.Fassung != exemplar.Fassung {
			return domain.ErrConflict
		}

		if zeile.ID == exemplar.ID {
			exemplar.Fassung = zeile.Fassung + 1
			s.zeilen[i] = exemplar

			return nil
		}
	}

	return domain.ErrNotFound
}

func (s *sammlungsprobe) ExemplarLoeschen(_ context.Context, id, fassung int64) error {
	for i, zeile := range s.zeilen {
		// **So streng wie der Bestand**: ein Löschen gegen einen überholten
		// Stand wird abgewiesen und nicht ausgeführt. Eine Vorrichtung, die
		// hier großzügiger wäre, prüfte den falschen Weg (§3).
		if zeile.ID == id && zeile.Fassung != fassung {
			return domain.ErrConflict
		}

		if zeile.ID == id {
			s.zeilen = append(s.zeilen[:i], s.zeilen[i+1:]...)

			return nil
		}
	}

	return domain.ErrNotFound
}
