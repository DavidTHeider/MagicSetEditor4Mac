package service_test

// **Das Doppel der Freigabenablage** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// freigabendoppel ist eine Ablage im Arbeitsspeicher.
type freigabendoppel struct {
	abgelegt []domain.Profilfreigabe

	// gefragt zählt die Lesevorgänge.
	//
	// **Damit „der Eigner wird nicht gefragt" eine Aussage ist** und nicht
	// bloß eine Vermutung über den Weg.
	gefragt int

	kaputt error
}

func neuesFreigabendoppel() *freigabendoppel { return &freigabendoppel{} }

func (d *freigabendoppel) Freigaben(_ context.Context, profilID int64) (
	[]domain.Profilfreigabe, error,
) {
	d.gefragt++

	if d.kaputt != nil {
		return nil, d.kaputt
	}

	gefunden := []domain.Profilfreigabe{}

	for _, eine := range d.abgelegt {
		if eine.ProfilID == profilID {
			gefunden = append(gefunden, eine)
		}
	}

	return gefunden, nil
}

func (d *freigabendoppel) FreigabenFuer(_ context.Context, nutzerID int64) (
	[]domain.Profilfreigabe, error,
) {
	if d.kaputt != nil {
		return nil, d.kaputt
	}

	gefunden := []domain.Profilfreigabe{}

	for _, eine := range d.abgelegt {
		if eine.NutzerID == nutzerID {
			gefunden = append(gefunden, eine)
		}
	}

	return gefunden, nil
}

func (d *freigabendoppel) FreigabeSetzen(
	_ context.Context, freigabe domain.Profilfreigabe,
) error {
	if d.kaputt != nil {
		return d.kaputt
	}

	for i, eine := range d.abgelegt {
		if eine.ProfilID == freigabe.ProfilID &&
			eine.NutzerID == freigabe.NutzerID {
			d.abgelegt[i].Rolle = freigabe.Rolle

			return nil
		}
	}

	d.abgelegt = append(d.abgelegt, freigabe)

	return nil
}

func (d *freigabendoppel) FreigabeLoeschen(
	_ context.Context, profilID, nutzerID int64,
) error {
	if d.kaputt != nil {
		return d.kaputt
	}

	for i, eine := range d.abgelegt {
		if eine.ProfilID == profilID && eine.NutzerID == nutzerID {
			d.abgelegt = append(d.abgelegt[:i], d.abgelegt[i+1:]...)

			return nil
		}
	}

	return nil
}
