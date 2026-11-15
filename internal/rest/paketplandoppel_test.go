package rest_test

// **Die Doubles für Plan, Ersetzen und Zurück** — DOM-01
// (#1696).
//
// **Eine eigene Datei**, weil die drei Doubles in ihren Dateien über die
// Zeilengrenze aus §2 gingen; sie stehen hier zusammen, weil sie dieselbe
// Naht erfüllen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) PaketPlan(_ context.Context, vorgang, art, name, _ string) (rest.Paketplan, error) {
	return rest.Paketplan{
		Vorgang: vorgang, Paket: name + "." + art, Von: "2026-01-01", Moeglich: true,
		Sets: []string{"Erstes"}, Konflikte: []rest.Paketkonflikt{},
	}, nil
}

func (williger) PaketErsetzen(context.Context, string, bool) (rest.Paket, error) {
	return rest.Paket{Art: "mse-style", Name: "probe", Version: "2026-02-02"}, nil
}

func (williger) PaketZurueck(_ context.Context, art, name string) (rest.Paket, error) {
	return rest.Paket{Art: art, Name: name, Version: "2026-01-01"}, nil
}

func (u unwilliger) PaketPlan(context.Context, string, string, string, string) (rest.Paketplan, error) {
	return rest.Paketplan{}, u.fehler
}

func (u unwilliger) PaketErsetzen(context.Context, string, bool) (rest.Paket, error) {
	return rest.Paket{}, u.fehler
}

func (u unwilliger) PaketZurueck(context.Context, string, string) (rest.Paket, error) {
	return rest.Paket{}, u.fehler
}

func (z *zaehlender) PaketPlan(ctx context.Context, vorgang, art, name, quelle string) (rest.Paketplan, error) {
	z.zaehle()

	return z.williger.PaketPlan(ctx, vorgang, art, name, quelle)
}

func (z *zaehlender) PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) (rest.Paket, error) {
	z.zaehle()

	return z.williger.PaketErsetzen(ctx, quelle, bestaetigt)
}

func (z *zaehlender) PaketZurueck(ctx context.Context, art, name string) (rest.Paket, error) {
	z.zaehle()

	return z.williger.PaketZurueck(ctx, art, name)
}

func (williger) PaketBedarf(context.Context, int64) ([]rest.Paketkonflikt, error) {
	return []rest.Paketkonflikt{{Paket: "Erstes", Braucht: "probe.mse-game", Art: "fehlt"}}, nil
}

func (u unwilliger) PaketBedarf(context.Context, int64) ([]rest.Paketkonflikt, error) {
	return nil, u.fehler
}

func (z *zaehlender) PaketBedarf(ctx context.Context, setID int64) ([]rest.Paketkonflikt, error) {
	z.zaehle()

	return z.williger.PaketBedarf(ctx, setID)
}
