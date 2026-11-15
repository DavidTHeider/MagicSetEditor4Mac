package rest_test

// **Die Doppel des Tauschabgleichs** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// **Eine eigene Datei**, weil `vorrichtungkontakte_test.go` sonst über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort die Kontakte
// und ihre Freigaben, hier der Abgleich.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Tauschabgleich(
	context.Context, string, rest.Fremdstand, int, int,
) (rest.Tauschabgleich, error) {
	return rest.Tauschabgleich{}, nil
}

func (u unwilliger) Tauschabgleich(
	context.Context, string, rest.Fremdstand, int, int,
) (rest.Tauschabgleich, error) {
	return rest.Tauschabgleich{}, u.fehler
}

func (z *zaehlender) Tauschabgleich(
	context.Context, string, rest.Fremdstand, int, int,
) (rest.Tauschabgleich, error) {
	z.zaehle()

	return rest.Tauschabgleich{}, nil
}
