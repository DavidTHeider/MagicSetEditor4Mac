package service_test

// **Das Doppel der Formatpflege** — Werkstück 4 von Phase 4.6
// (#1293).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// formatdoppel ist ein Formatspeicher im Arbeitsspeicher.
//
// **Er hält die Fassung wie der Bestand** (§3): `Anlegen` setzt 1, `Aendern`
// prüft gegen die mitgebrachte und erhöht.
type formatdoppel struct {
	formate  map[int64]domain.Format
	naechste int64
	fehler   error
}

func (f *formatdoppel) Anlegen(_ context.Context, format domain.Format) (
	domain.Format, error,
) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	if f.formate == nil {
		f.formate = map[int64]domain.Format{}
	}

	f.naechste++
	format.ID = f.naechste
	format.Fassung = 1
	f.formate[format.ID] = format

	return format, nil
}

func (f *formatdoppel) Aendern(_ context.Context, format domain.Format) (
	domain.Format, error,
) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	stand, da := f.formate[format.ID]
	if !da {
		return domain.Format{}, fmt.Errorf(
			"%w: das Format gibt es nicht", domain.ErrNotFound)
	}

	if stand.Fassung != format.Fassung {
		return domain.Format{}, fmt.Errorf(
			"%w: das Format hat sich seit dem Lesen geändert", domain.ErrConflict)
	}

	format.Fassung++
	f.formate[format.ID] = format

	return format, nil
}

func (f *formatdoppel) Format(_ context.Context, kennung int64) (
	domain.Format, error,
) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	format, da := f.formate[kennung]
	if !da {
		return domain.Format{}, fmt.Errorf(
			"%w: das Format gibt es nicht", domain.ErrNotFound)
	}

	return format, nil
}

func (f *formatdoppel) Formate(_ context.Context, spiel string) (
	[]domain.Format, error,
) {
	if f.fehler != nil {
		return nil, f.fehler
	}

	formate := []domain.Format{}

	for _, format := range f.formate {
		if spiel == "" || format.Spiel == spiel {
			formate = append(formate, format)
		}
	}

	return formate, nil
}

func (f *formatdoppel) Loeschen(_ context.Context, kennung int64) error {
	if f.fehler != nil {
		return f.fehler
	}

	delete(f.formate, kennung)

	return nil
}
