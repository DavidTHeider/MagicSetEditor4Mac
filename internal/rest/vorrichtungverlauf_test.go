// **Der Verlauf an den drei Doppeln** — Werkstück U6 von Phase 5.1 (#1417),
// in eigener Datei, weil `vorrichtungbestand_test.go` an der Zeilengrenze
// steht (§2).
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) Verlauf(context.Context, int64) ([]rest.Verlaufsschritt, error) {
	return nil, u.fehler
}

func (williger) Verlauf(context.Context, int64) ([]rest.Verlaufsschritt, error) {
	return []rest.Verlaufsschritt{{
		ID: "3", Vorgang: "g1", Umfang: 2, KarteID: "8", Handlung: "feld", Feld: "typ",
		Vorher: "Zauber", GabVorher: true, Nachher: "Land", GabNachher: true,
		Urheber: "server", Zeitpunkt: "2026-09-09T02:01:00Z",
	}}, nil
}

func (z *zaehlender) Verlauf(context.Context, int64) ([]rest.Verlaufsschritt, error) {
	z.rufe++

	return []rest.Verlaufsschritt{{
		ID: "3", Vorgang: "g1", Umfang: 2, KarteID: "8", Handlung: "feld", Feld: "typ",
		Vorher: "Zauber", GabVorher: true, Nachher: "Land", GabNachher: true,
		Urheber: "server", Zeitpunkt: "2026-09-09T02:01:00Z",
	}}, nil
}
