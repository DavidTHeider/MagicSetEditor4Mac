package rest_test

// **Die Doppel der Tokennaht** — Werkstück 10 von Phase 4.12
// (#1351).
//
// **Beide Richtungen an einer Stelle**: das willige antwortet, das unwillige
// scheitert. Sie stehen zusammen, weil sie dieselbe Naht bedienen — und
// getrennt von den Nutzerdoppeln, weil es eine andere Naht ist.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// williger bedient auch die Tokennaht.
func (williger) Token(context.Context, string) ([]rest.Token, error) {
	return []rest.Token{{
		ID: "1", ProfilID: "1", Bezeichnung: "Nachtlauf", Scope: "lesen",
		Angelegt: "2026-09-05T12:00:00Z", Fassung: "1",
	}}, nil
}

func (williger) TokenAnlegen(_ context.Context, angaben rest.Tokenangaben) (
	rest.Token, string, error,
) {
	return rest.Token{
		ID: "2", ProfilID: angaben.ProfilID, Bezeichnung: angaben.Bezeichnung,
		Scope: angaben.Scope, Ablauf: angaben.Ablauf,
		Angelegt: "2026-09-05T12:00:00Z", Fassung: "1",
	}, "mse_2_GEHEIMNIS", nil
}

func (williger) TokenAendern(_ context.Context, angaben rest.Tokenangaben) (
	rest.Token, error,
) {
	return rest.Token{
		ID: angaben.TokenID, ProfilID: "1", Bezeichnung: angaben.Bezeichnung,
		Scope: angaben.Scope, Ablauf: angaben.Ablauf,
		Angelegt: "2026-09-05T12:00:00Z", Fassung: "2",
	}, nil
}

func (williger) TokenLoeschen(context.Context, string) error { return nil }

// **Werkstück O2** (#1426): Rotation, Sammelwiderruf und Ablaufwahlen.
func (williger) TokenRotieren(_ context.Context, tokenID string) (rest.Token, string, error) {
	return rest.Token{
		ID: tokenID, ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Rechner",
		Scope: "lesen", VorgaengerBis: "2026-09-08T13:00:00Z", Fassung: "2",
	}, "mse_" + tokenID + "_NEU", nil
}

func (williger) TokenAlleWiderrufen(context.Context, string) (int64, error) { return 2, nil }

func (williger) Tokenablaufwahlen(context.Context) ([]rest.Ablaufwahl, error) {
	return []rest.Ablaufwahl{{Name: "tag", Ablauf: "2026-09-06T12:00:00Z"}, {Name: "nie"}}, nil
}

// unwilliger scheitert auf jedem der vier Wege.
func (u unwilliger) Token(context.Context, string) ([]rest.Token, error) {
	return nil, u.fehler
}

func (u unwilliger) TokenAnlegen(context.Context, rest.Tokenangaben) (
	rest.Token, string, error,
) {
	return rest.Token{}, "", u.fehler
}

func (u unwilliger) TokenAendern(context.Context, rest.Tokenangaben) (
	rest.Token, error,
) {
	return rest.Token{}, u.fehler
}

func (u unwilliger) TokenLoeschen(context.Context, string) error { return u.fehler }

func (u unwilliger) TokenRotieren(context.Context, string) (rest.Token, string, error) {
	return rest.Token{}, "", u.fehler
}

func (u unwilliger) TokenAlleWiderrufen(context.Context, string) (int64, error) {
	return 0, u.fehler
}

func (u unwilliger) Tokenablaufwahlen(context.Context) ([]rest.Ablaufwahl, error) {
	return nil, u.fehler
}

// zaehlender zählt die vier mit — sonst belegte „gültig" nur einen Status und
// nicht den Weg bis zum Anwendungsfall.

func (z *zaehlender) Token(ctx context.Context, profilID string) ([]rest.Token, error) {
	z.zaehle()

	return z.williger.Token(ctx, profilID)
}

func (z *zaehlender) TokenAnlegen(ctx context.Context, angaben rest.Tokenangaben) (
	rest.Token, string, error,
) {
	z.zaehle()

	return z.williger.TokenAnlegen(ctx, angaben)
}

func (z *zaehlender) TokenAendern(ctx context.Context, angaben rest.Tokenangaben) (
	rest.Token, error,
) {
	z.zaehle()

	return z.williger.TokenAendern(ctx, angaben)
}

func (z *zaehlender) TokenLoeschen(ctx context.Context, tokenID string) error {
	z.zaehle()

	return z.williger.TokenLoeschen(ctx, tokenID)
}

func (z *zaehlender) TokenRotieren(ctx context.Context, tokenID string) (rest.Token, string, error) {
	z.zaehle()

	return z.williger.TokenRotieren(ctx, tokenID)
}

func (z *zaehlender) TokenAlleWiderrufen(ctx context.Context, profilID string) (int64, error) {
	z.zaehle()

	return z.williger.TokenAlleWiderrufen(ctx, profilID)
}

func (z *zaehlender) Tokenablaufwahlen(ctx context.Context) ([]rest.Ablaufwahl, error) {
	z.zaehle()

	return z.williger.Tokenablaufwahlen(ctx)
}
