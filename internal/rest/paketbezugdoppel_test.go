package rest_test

// **Die Doubles für die Bezugsstände** — G04
// (#2251).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) PaketBezuege(context.Context, bool) ([]rest.Bezugsstand, error) {
	return []rest.Bezugsstand{{Name: "M15-Magic-Pack", Stand: "a", Lage: "neuer", Spitze: "b", Vorsprung: 2}}, nil
}

func (williger) PaketBezugVorschau(_ context.Context, name, stand string) (rest.Bezugsvorschau, error) {
	return rest.Bezugsvorschau{
		Name: name, Von: "a", Zu: stand, Lage: "ahead", Pakete: []rest.Paketaenderung{},
		Sets: []string{}, Gruende: []string{},
	}, nil
}

func (williger) PaketBezugAktualisieren(_ context.Context, name, stand string, _ bool) (rest.Bezug, error) {
	return rest.Bezug{Ort: "/pakete/" + name, Stand: stand, Uebergangen: []string{}}, nil
}

func (williger) PaketBezugZurueck(_ context.Context, name string) (rest.Bezug, error) {
	return rest.Bezug{Ort: "/pakete/" + name, Stand: "a", Uebergangen: []string{}}, nil
}

func (u unwilliger) PaketBezuege(context.Context, bool) ([]rest.Bezugsstand, error) {
	return nil, u.fehler
}

func (u unwilliger) PaketBezugVorschau(context.Context, string, string) (rest.Bezugsvorschau, error) {
	return rest.Bezugsvorschau{}, u.fehler
}

// **Der Baum liegt, das Vermerken scheiterte** — der Teilstand reist mit
// (§7, Regel 3), wie bei jedem Schreiben, das auf halber Strecke stehen bleibt.
func (u unwilliger) PaketBezugAktualisieren(context.Context, string, string, bool) (rest.Bezug, error) {
	return rest.Bezug{Ort: "/pakete/M15-Magic-Pack", Stand: "b", Uebergangen: []string{}}, u.fehler
}

func (u unwilliger) PaketBezugZurueck(context.Context, string) (rest.Bezug, error) {
	return rest.Bezug{}, u.fehler
}

func (z *zaehlender) PaketBezuege(ctx context.Context, fragen bool) ([]rest.Bezugsstand, error) {
	z.zaehle()

	return z.williger.PaketBezuege(ctx, fragen)
}

func (z *zaehlender) PaketBezugVorschau(ctx context.Context, name, stand string) (rest.Bezugsvorschau, error) {
	z.zaehle()

	return z.williger.PaketBezugVorschau(ctx, name, stand)
}

func (z *zaehlender) PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (rest.Bezug, error) {
	z.zaehle()

	return z.williger.PaketBezugAktualisieren(ctx, name, stand, bestaetigt)
}

func (z *zaehlender) PaketBezugZurueck(ctx context.Context, name string) (rest.Bezug, error) {
	z.zaehle()

	return z.williger.PaketBezugZurueck(ctx, name)
}
