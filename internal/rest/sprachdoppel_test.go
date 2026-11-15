package rest_test

// **Die Doppel der eigenen Übersetzungen** — LOC-02 (#1688).

import "context"

func (w williger) Sprachpakete(context.Context) (any, error) {
	return []any{map[string]any{"kennung": "da"}}, nil
}

func (w williger) Sprachpaket(context.Context, string) (any, error) {
	return map[string]any{"kennung": "da"}, nil
}

func (w williger) SprachpaketAnlegen(context.Context, string, string, string) (any, error) {
	return map[string]any{"kennung": "da"}, nil
}

func (w williger) SprachpaketAendern(context.Context, string, map[string]string, map[string]string) (any, error) {
	return map[string]any{"kennung": "da"}, nil
}

func (w williger) SprachpaketLoeschen(context.Context, string) (any, error) {
	return map[string]any{"geloescht": "da"}, nil
}

func (w williger) SprachpaketAusgabe(context.Context, string) (any, error) {
	return map[string]any{"name": "da.sprachpaket.json"}, nil
}

func (w williger) SprachpaketEinlesen(context.Context, string, bool) (any, error) {
	return map[string]any{"kennung": "da"}, nil
}

func (u unwilliger) Sprachpakete(context.Context) (any, error) { return nil, u.fehler }

func (u unwilliger) Sprachpaket(context.Context, string) (any, error) { return nil, u.fehler }

func (u unwilliger) SprachpaketAnlegen(context.Context, string, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SprachpaketAendern(context.Context, string, map[string]string, map[string]string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SprachpaketLoeschen(context.Context, string) (any, error) { return nil, u.fehler }

func (u unwilliger) SprachpaketAusgabe(context.Context, string) (any, error) { return nil, u.fehler }

func (u unwilliger) SprachpaketEinlesen(context.Context, string, bool) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) Sprachpakete(ctx context.Context) (any, error) {
	z.zaehle()

	return z.williger.Sprachpakete(ctx)
}

func (z *zaehlender) Sprachpaket(ctx context.Context, kennung string) (any, error) {
	z.zaehle()

	return z.williger.Sprachpaket(ctx, kennung)
}

func (z *zaehlender) SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) (any, error) {
	z.zaehle()

	return z.williger.SprachpaketAnlegen(ctx, kennung, name, basis)
}

func (z *zaehlender) SprachpaketAendern(
	ctx context.Context, kennung string, angaben, texte map[string]string,
) (any, error) {
	z.zaehle()

	return z.williger.SprachpaketAendern(ctx, kennung, angaben, texte)
}

func (z *zaehlender) SprachpaketLoeschen(ctx context.Context, kennung string) (any, error) {
	z.zaehle()

	return z.williger.SprachpaketLoeschen(ctx, kennung)
}

func (z *zaehlender) SprachpaketAusgabe(ctx context.Context, kennung string) (any, error) {
	z.zaehle()

	return z.williger.SprachpaketAusgabe(ctx, kennung)
}

func (z *zaehlender) SprachpaketEinlesen(ctx context.Context, inhalt string, ersetzen bool) (any, error) {
	z.zaehle()

	return z.williger.SprachpaketEinlesen(ctx, inhalt, ersetzen)
}
