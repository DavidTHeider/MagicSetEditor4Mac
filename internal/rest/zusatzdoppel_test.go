package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die Zusatzfelder an den Allzweck-Doubles** (#2298) — eine eigene Datei,
// weil die der Doubles an der Zeilengrenze aus §2 stehen.

func (williger) Zusatzfelder(context.Context, int64, int64) ([]rest.Zusatzfeld, error) {
	return []rest.Zusatzfeld{{Name: "vermerk", Art: "text"}}, nil
}

func (williger) ZusatzwertSetzen(context.Context, int64, int64, string, string) error { return nil }

func (u unwilliger) Zusatzfelder(context.Context, int64, int64) ([]rest.Zusatzfeld, error) {
	return nil, u.fehler
}

func (u unwilliger) ZusatzwertSetzen(context.Context, int64, int64, string, string) error {
	return u.fehler
}

func (z *zaehlender) Zusatzfelder(ctx context.Context, setID, karteID int64) ([]rest.Zusatzfeld, error) {
	z.zaehle()

	return z.williger.Zusatzfelder(ctx, setID, karteID)
}

func (z *zaehlender) ZusatzwertSetzen(ctx context.Context, setID, karteID int64, feld, wert string) error {
	z.zaehle()

	return z.williger.ZusatzwertSetzen(ctx, setID, karteID, feld, wert)
}
