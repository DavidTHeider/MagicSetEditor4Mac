package service_test

// **Die Doppel der Vorgabenverwaltung** — Werkstück 15 von Phase 4.12
// (#1351).
//
// **Sie stehen eigens hier**, wie `kennungsdoppel` und `tokendoppel`: die
// Testdatei hält die Verhaltensklassen zusammen (§2).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorgabendoppel ist eine Ablage im Arbeitsspeicher.
type vorgabendoppel struct {
	abgelegt map[int64]map[domain.Vorgabeart][]string
	kaputt   error
}

func neuesVorgabendoppel() *vorgabendoppel {
	return &vorgabendoppel{abgelegt: map[int64]map[domain.Vorgabeart][]string{}}
}

func (d *vorgabendoppel) Vorgaben(_ context.Context, profilID int64) (
	map[domain.Vorgabeart][]string, error,
) {
	if d.kaputt != nil {
		return nil, d.kaputt
	}

	if d.abgelegt[profilID] == nil {
		return map[domain.Vorgabeart][]string{}, nil
	}

	return d.abgelegt[profilID], nil
}

func (d *vorgabendoppel) VorgabeSetzen(
	_ context.Context, profilID int64, vorgabe domain.Profilvorgabe,
) error {
	if d.kaputt != nil {
		return d.kaputt
	}

	if d.abgelegt[profilID] == nil {
		d.abgelegt[profilID] = map[domain.Vorgabeart][]string{}
	}

	if len(vorgabe.Werte) == 0 {
		delete(d.abgelegt[profilID], vorgabe.Art)

		return nil
	}

	d.abgelegt[profilID][vorgabe.Art] = vorgabe.Werte

	return nil
}

// angebotsdoppel sagt, was es je Art gibt.
//
// **Es hält seine Wahlen in einer Tabelle, die sich ändern lässt** — das ist
// der Fall, für den es dieses Doppel gibt: ein Deck, das gelöscht wird,
// verschwindet aus dem Angebot und lässt eine Vorgabe verwaist zurück.
type angebotsdoppel struct {
	wahlen map[domain.Vorgabeart][]service.Vorgabewahl
	kaputt error
}

func (d *angebotsdoppel) Angebot(_ context.Context, art domain.Vorgabeart) (
	[]service.Vorgabewahl, error,
) {
	if d.kaputt != nil {
		return nil, d.kaputt
	}

	return d.wahlen[art], nil
}
