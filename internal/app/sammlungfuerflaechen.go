// **Die Sammlung für die Flächen ohne Fenster** — Werkstück 2 und 3 von
// Phase 3.4 (#1165).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Warum eine Naht und nicht dreimal derselbe Aufbau
//
// Weil der **Preis als Text** eine Entscheidung dieser Schicht ist und keine
// der Domäne. Wer sie in jeder Fläche wiederholte, hätte drei Umrechnungen —
// und die dritte rundete irgendwann.

package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Sammlungsnaht ist die Sammlung, wie eine Fläche ohne Fenster sie ruft.
type Sammlungsnaht struct {
	pflege *service.Sammlungspflege

	// fenster ist der Ausschnitt einer Sicht — Werkstück
	// #1500.
	fenster *service.Sammlungsfenster
	preise  *service.Preispflege
	uhr     func() time.Time
}

// SammlungsnahtUeber baut sie über einer Datenbank.
//
// **Ohne Zuhause gibt es keine Preise, aber eine Sammlung.** Exemplare zu
// führen braucht nur den Bestand; wer Preise holen will, braucht einen
// Zwischenspeicher — und ohne den scheitert dieser Aufbau, statt eine halbe
// Naht herauszugeben.
func SammlungsnahtUeber(db *sql.DB) (*Sammlungsnaht, error) {
	// **Über die eine Stelle, an der ein Klient entsteht** (Phase 4.8) — dort
	// wird auch der Dienstschalter geprüft, und jede eigene Konstruktion wäre
	// ein Loch: ein Weg, der weiter hinausgeht, während der Dienst gesperrt ist.
	klient, err := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if err != nil {
		return nil, err
	}

	return &Sammlungsnaht{
		pflege:  service.NeueSammlungspflege(storage.NeuerSammlungsspeicher(db)),
		fenster: service.NeuesSammlungsfenster(storage.NeuerSammlungsspeicher(db)),
		preise: service.NeuePreispflege(storage.NeuerPreisspeicher(db),
			storage.NeuerSammlungsspeicher(db), klient, time.Now),
		uhr: time.Now,
	}, nil
}

// Sammlung gibt die ganze Sammlung.
func (s *Sammlungsnaht) Sammlung(ctx context.Context) ([]ExemplarDTO, error) {
	exemplare, err := s.pflege.Liste(ctx)
	if err != nil {
		return nil, err
	}

	return zuExemplarDTOs(exemplare), nil
}

// Sammlungsfenster gibt den Ausschnitt einer Sicht — samt beider Zahlen.
//
// **Derselbe Weg wie am Fenster** (Werkstück
// #1500):
// filtern und ordnen dort, wo die ganze Liste liegt. Ein Aufrufer ohne
// Ausschnitt bekommt weiterhin alles.
func (s *Sammlungsnaht) Sammlungsfenster(
	ctx context.Context, sicht domain.Sammlungssicht, ab, anzahl int,
) (fenster []ExemplarDTO, zahlen domain.Fensterzahlen, fehler error) {
	ausschnitt, zahlen, fehler := s.fenster.Sicht(ctx, sicht, ab, anzahl)
	if fehler != nil {
		return nil, domain.Fensterzahlen{}, fehler
	}

	return zuExemplarDTOs(ausschnitt), zahlen, nil
}

// SammlungFuer gibt die Exemplare eines Drucks.
func (s *Sammlungsnaht) SammlungFuer(ctx context.Context, printing string) (
	[]ExemplarDTO, error,
) {
	exemplare, err := s.pflege.Fuer(ctx, printing)
	if err != nil {
		return nil, err
	}

	return zuExemplarDTOs(exemplare), nil
}

// ExemplarAnlegen nimmt ein Exemplar auf.
func (s *Sammlungsnaht) ExemplarAnlegen(ctx context.Context, dto ExemplarDTO) (int64, error) {
	exemplar, err := ausExemplarDTO(dto)
	if err != nil {
		return 0, err
	}

	return s.pflege.Anlegen(ctx, exemplar)
}

// ExemplarAendern schreibt ein Exemplar zurück.
func (s *Sammlungsnaht) ExemplarAendern(ctx context.Context, dto ExemplarDTO) error {
	exemplar, err := ausExemplarDTO(dto)
	if err != nil {
		return err
	}

	return s.pflege.Aendern(ctx, exemplar)
}

// ExemplarLoeschen nimmt ein Exemplar aus der Sammlung.
func (s *Sammlungsnaht) ExemplarLoeschen(ctx context.Context, id, fassung int64) error {
	return s.pflege.Loeschen(ctx, id, fassung)
}

// PreiseHolen fragt die Quelle und legt die Stände ab.
func (s *Sammlungsnaht) PreiseHolen(ctx context.Context, printing string) (
	[]PreisstandDTO, []string, error,
) {
	staende, alter, befunde, err := s.preise.Holen(ctx, printing)
	if err != nil {
		return nil, alsMeldungen(befunde), err
	}

	return zuPreisstandDTOs(staende, alter), alsMeldungen(befunde), nil
}

// Preise gibt die abgelegten Stände eines Drucks.
func (s *Sammlungsnaht) Preise(ctx context.Context, printing string) (
	[]PreisstandDTO, error,
) {
	staende, err := s.preise.Neueste(ctx, printing)
	if err != nil {
		return nil, err
	}

	return zuPreisstandDTOs(staende, alterSeit(s.uhr(), staende)), nil
}

// Preisverlauf gibt eine Reihe in der Zeit.
//
// **Die Reihe wird genannt und nicht erraten** (A3, #1442).
func (s *Sammlungsnaht) Preisverlauf(
	ctx context.Context, reihe PreisreiheDTO,
) ([]PreisstandDTO, error) {
	staende, err := s.preise.Verlauf(ctx, reihe.Printing, reihe.alsReihe())
	if err != nil {
		return nil, err
	}

	return zuPreisstandDTOs(staende, alterSeit(s.uhr(), staende)), nil
}

// Portfolio gibt die Übersicht.
func (s *Sammlungsnaht) Portfolio(ctx context.Context) ([]PortfoliozeileDTO, int, error) {
	uebersicht, err := s.preise.Portfolio(ctx)
	if err != nil {
		return nil, 0, err
	}

	return zuPortfoliozeilen(uebersicht.Zeilen, s.uhr()), uebersicht.OhnePreis, nil
}
