// Die Doubles des Probelaufs — UXR-F7
// (#2015).
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// BogenProbelauf antwortet willig: zehn Karten auf zwei Seiten.
func (williger) BogenProbelauf(context.Context, int64, string) (rest.Bogenprobelauf, error) {
	return rest.Bogenprobelauf{
		Karten: 10, Seiten: 2, Spalten: 3, Zeilen: 3, JeSeite: 9, AufLetzterSeite: 1,
		Seitenbreite: 210, Seitenhoehe: 297, Kartenbreite: 63, Kartenhoehe: 88,
		Rand: 10, DPI: 300, Schnittmarken: true,
	}, nil
}

// VorlageProbelauf antwortet willig: zwei Karten, eine kleine Datei.
func (williger) VorlageProbelauf(
	context.Context, int64, string, string, map[string]string,
) (rest.Vorlagenprobelauf, error) {
	return rest.Vorlagenprobelauf{
		Karten: 2, Dateiname: "ausfuhr.txt", Bytes: 12, Befunde: []string{},
	}, nil
}

func (u unwilliger) BogenProbelauf(context.Context, int64, string) (rest.Bogenprobelauf, error) {
	return rest.Bogenprobelauf{}, u.fehler
}

func (u unwilliger) VorlageProbelauf(
	context.Context, int64, string, string, map[string]string,
) (rest.Vorlagenprobelauf, error) {
	return rest.Vorlagenprobelauf{}, u.fehler
}

func (z *zaehlender) BogenProbelauf(
	ctx context.Context, setID int64, ziel string,
) (rest.Bogenprobelauf, error) {
	z.zaehle()

	return z.williger.BogenProbelauf(ctx, setID, ziel)
}

func (z *zaehlender) VorlageProbelauf(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) (rest.Vorlagenprobelauf, error) {
	z.zaehle()

	return z.williger.VorlageProbelauf(ctx, setID, paket, ziel, wahl)
}
