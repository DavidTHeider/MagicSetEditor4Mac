package rest_test

// **Die Doppel der Marktlinks** — Werkstück `T8` von Phase 5.6
// (#1473).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Marktlinks(
	context.Context, string, string,
) ([]rest.Marktlink, error) {
	return []rest.Marktlink{}, nil
}

func (u unwilliger) Marktlinks(
	context.Context, string, string,
) ([]rest.Marktlink, error) {
	return nil, u.fehler
}

func (z *zaehlender) Marktlinks(
	context.Context, string, string,
) ([]rest.Marktlink, error) {
	z.zaehle()

	return []rest.Marktlink{}, nil
}

// **Und die des Marktkontos** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// **Sie stehen hier und nicht in einer eigenen Datei**: beide sprechen über
// denselben Marktplatz, und der Schnitt folgt der Frage (§2).

func (williger) Marktkonto(context.Context, string, int, int) (any, error) {
	return map[string]any{
		"name": "Probe", "nutzerID": 7,
		"angebote": []any{}, "bestellungen": []any{}, "befunde": []string{},
	}, nil
}

func (u unwilliger) Marktkonto(context.Context, string, int, int) (any, error) {
	return nil, u.fehler
}

func (z *zaehlender) Marktkonto(context.Context, string, int, int) (any, error) {
	z.zaehle()

	return map[string]any{"name": "Probe", "nutzerID": 7}, nil
}
