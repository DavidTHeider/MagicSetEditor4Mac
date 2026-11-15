package rest_test

// **Der Preisanhalt der drei Doppel** — TRADE-02b
// (#1951).

import "context"

func (williger) HandelPreisanhalt(context.Context, string) (any, error) {
	return map[string]any{"hauptwaehrung": "EUR"}, nil
}

func (u unwilliger) HandelPreisanhalt(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) HandelPreisanhalt(context.Context, string) (any, error) {
	z.zaehle()

	return map[string]any{}, nil
}
