// **Was die Sitzung über sich sagt, und wie sie einem Wechsel folgt** — ID-02
// (#1701).
//
// **Eine eigene Datei neben `sitzung.go`**, damit jene unter der Zeilengrenze
// aus §2 bleibt: dort stehen die Übergänge, hier, was keiner ist.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sitzungsauskunft sagt, was ein Sperrbildschirm anbieten kann.
type Sitzungsauskunft struct {
	// Kennwort sagt, ob der Nutzer eines hat; Sensor, ob er auf diesem
	// Rechner lokal entsperren kann.
	Kennwort, Sensor bool
}

// Auskunft sagt es für den Nutzer dieser Sitzung — **ohne den Sensor zu
// fragen**.
func (s *Sitzung) Auskunft(ctx context.Context) (Sitzungsauskunft, error) {
	_, nutzerID := s.Stand()

	einer, fehler := s.nutzerMit(ctx, nutzerID)
	if fehler != nil {
		return Sitzungsauskunft{}, fehler
	}

	sensor, fehler := s.faktor.SensorMoeglich(ctx, nutzerID)
	if fehler != nil {
		return Sitzungsauskunft{}, fehler
	}

	return Sitzungsauskunft{Kennwort: einer.Kennwort != "", Sensor: sensor}, nil
}

// Umgestellt folgt einem Nutzerwechsel, den [Profilwechsel] geprüft hat —
// **nur angemeldet**: gesperrt und abgemeldet wechselt niemand.
func (s *Sitzung) Umgestellt(nutzerID int64) {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if s.zustand == domain.SitzungAngemeldet {
		s.nutzer = nutzerID
		s.zeigen()
	}
}
