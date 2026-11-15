// **Die Fragen an eine geordnete Sammlung, für alle vier Flächen** —
// Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// **Ohne Netz**, wie die Ausfuhr: alle vier Fragen lesen den Bestand.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// BesitzDTO sagt, was von einem Druck da ist.
type BesitzDTO struct {
	Printing string `json:"printing"`

	// Stueck ist die Zahl der **besessenen** Stücke.
	Stueck int `json:"stueck"`

	// Gewuenscht ist die Zahl der gewünschten Stücke — getrennt gezählt (0057).
	Gewuenscht int `json:"gewuenscht"`

	// Besitzt sagt, ob mindestens ein Stück da ist.
	//
	// **Mitgeschickt und nicht drüben gerechnet**: die Regel „ein Wunsch ist
	// kein Besitz" gehört in eine Schicht.
	Besitzt bool `json:"besitzt"`

	Zeilen []ExemplarDTO `json:"zeilen"`
}

// LagerortDTO ist ein Lagerort samt dem, was dort liegt.
type LagerortDTO struct {
	// Lagerort ist der Freitext — leer heißt „ohne Angabe".
	Lagerort string `json:"lagerort"`
	Zeilen   int    `json:"zeilen"`
	Stueck   int    `json:"stueck"`
}

// DublettenDTO ist ein Exemplar, von dem mehr da ist als gewollt.
type DublettenDTO struct {
	Exemplar ExemplarDTO `json:"exemplar"`

	// Ueberzaehlig ist, wie viele über der behaltenen Zahl liegen.
	Ueberzaehlig int `json:"ueberzaehlig"`
}

// Sammlungssichtennaht beantwortet die vier Fragen.
type Sammlungssichtennaht struct{ sichten *service.Sammlungssichten }

// SichtenUeber baut sie über einer Datenbank.
func SichtenUeber(db *sql.DB) *Sammlungssichtennaht {
	return &Sammlungssichtennaht{
		// **Mit der Naht zum Handeln** (T2, #1467, 0120): was zugesagt ist,
		// steht nicht als überzählig da.
		sichten: service.NeueSammlungssichten(storage.NeuerSammlungsspeicher(db)).
			MitReservierungen(storage.NeueReservierungsablage(db)),
	}
}

// Besitz sagt, was von einem Druck da ist.
func (s *Sammlungssichtennaht) Besitz(ctx context.Context, printing string) (
	*BesitzDTO, error,
) {
	stand, err := s.sichten.Besitz(ctx, printing)
	if err != nil {
		return nil, err
	}

	return &BesitzDTO{
		Printing: printing, Stueck: stand.Stueck,
		Gewuenscht: stand.Gewuenscht, Besitzt: stand.Besitzt(),
		Zeilen: zuExemplarDTOs(stand.Zeilen),
	}, nil
}

// Lagerorte gibt die Lagerorte samt dem, was dort liegt.
func (s *Sammlungssichtennaht) Lagerorte(ctx context.Context) ([]LagerortDTO, error) {
	orte, err := s.sichten.Lagerorte(ctx)
	if err != nil {
		return nil, err
	}

	zeilen := make([]LagerortDTO, 0, len(orte))
	for _, ort := range orte {
		zeilen = append(zeilen, LagerortDTO{
			Lagerort: ort.Lagerort, Zeilen: ort.Zeilen, Stueck: ort.Stueck,
		})
	}

	return zeilen, nil
}

// Dubletten gibt die Exemplare, von denen mehr da ist als `behalten`.
func (s *Sammlungssichtennaht) Dubletten(ctx context.Context, behalten int) (
	[]DublettenDTO, error,
) {
	dubletten, err := s.sichten.Dubletten(ctx, behalten)
	if err != nil {
		return nil, err
	}

	zeilen := make([]DublettenDTO, 0, len(dubletten))
	for _, dublette := range dubletten {
		zeilen = append(zeilen, DublettenDTO{
			Exemplar:     zuExemplarDTO(dublette.Exemplar),
			Ueberzaehlig: dublette.Ueberzaehlig,
		})
	}

	return zeilen, nil
}

// Wuensche gibt die Zeilen, die kein Besitz sind.
func (s *Sammlungssichtennaht) Wuensche(ctx context.Context) ([]ExemplarDTO, error) {
	wuensche, err := s.sichten.Wuensche(ctx)
	if err != nil {
		return nil, err
	}

	return zuExemplarDTOs(wuensche), nil
}
