// **Die Doppel für Weitergabe und Einlesen** — FMT-02.AK-04 (#1691).
// Eine eigene Datei, weil `vorrichtungbestand_test.go` an der Zeilengrenze
// aus §2 steht.

package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) PaketWeitergeben(_ context.Context, _ []string, ziel string) (string, error) {
	return ziel, nil
}

func (williger) PaketEinlesen(context.Context, string) (rest.Uebernahme, error) {
	return rest.Uebernahme{Uebernommen: []rest.Paket{}, Uebergangen: []string{}, Befunde: []string{}}, nil
}

func (z *zaehlender) PaketWeitergeben(ctx context.Context, pakete []string, ziel string) (string, error) {
	z.zaehle()

	return z.williger.PaketWeitergeben(ctx, pakete, ziel)
}

func (z *zaehlender) PaketEinlesen(ctx context.Context, datei string) (rest.Uebernahme, error) {
	z.zaehle()

	return z.williger.PaketEinlesen(ctx, datei)
}

func (u unwilliger) PaketWeitergeben(context.Context, []string, string) (string, error) {
	return "", u.fehler
}

func (u unwilliger) PaketEinlesen(context.Context, string) (rest.Uebernahme, error) {
	return rest.Uebernahme{}, u.fehler
}
