package rest_test

// **Die Vorschauen der Einfuhr an den drei Doppeln** — DOM-03.AK-01
// (#1698).
//
// **Eine eigene Datei**, weil `willigerdoppel_test.go` und
// `unwilligeausfuhr_test.go` mit ihnen über die Zeilengrenze aus §2 gingen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) ListeVorschauen(context.Context, []rest.Listeneinfuhrzeile) (any, error) {
	return map[string]any{"posten": []any{}, "befunde": []string{}}, nil
}

func (williger) SammlungCSVVorschauen(context.Context, string) (any, error) {
	return map[string]any{"posten": []any{}, "befunde": []string{}}, nil
}

func (u unwilliger) ListeVorschauen(context.Context, []rest.Listeneinfuhrzeile) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SammlungCSVVorschauen(context.Context, string) (any, error) {
	return nil, u.fehler
}

// ListeVorschauen sagt, was das Einlesen täte, und zählt mit.
func (z *zaehlender) ListeVorschauen(context.Context, []rest.Listeneinfuhrzeile) (any, error) {
	z.rufe++

	return map[string]any{"posten": []any{}, "befunde": []string{}}, nil
}

// SammlungCSVVorschauen sagt, was das Einlesen täte, und zählt mit.
func (z *zaehlender) SammlungCSVVorschauen(context.Context, string) (any, error) {
	z.rufe++

	return map[string]any{"posten": []any{}, "befunde": []string{}}, nil
}
