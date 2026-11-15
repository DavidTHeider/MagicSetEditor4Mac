package rest_test

// **Die Doppel für die Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Eine eigene Datei**, aus demselben Grund wie bei `indexdoppel_test.go`:
// die Vorrichtungen daneben gingen mit ihr über die Zeilengrenze aus §2.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// eineRecherche ist die Antwort, die alle drei Doppel geben.
func eineRecherche() map[string]any {
	return map[string]any{
		"treffer": []any{}, "modus": "ohne-index", "leere": "ohne-index",
	}
}

// einKontextpaket ist das Paket, das alle drei Doppel geben.
func einKontextpaket() map[string]any {
	return map[string]any{"paketID": "kp-1", "belege": []any{}}
}

// **Der Willige** — er antwortet auf alles.
func (w williger) Recherche(
	context.Context, string, string, int, string, string, string, []string,
) (any, error) {
	return eineRecherche(), nil
}

func (w williger) AehnlicheKarten(
	context.Context, int64, string, int, string, string, []string,
) (any, error) {
	return eineRecherche(), nil
}

func (w williger) Kontextpaket(context.Context, rest.Kontextwunsch) (
	any, error,
) {
	return einKontextpaket(), nil
}

// **Der Unwillige** — er antwortet auf nichts.
func (u unwilliger) Recherche(
	context.Context, string, string, int, string, string, string, []string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) AehnlicheKarten(
	context.Context, int64, string, int, string, string, []string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Kontextpaket(context.Context, rest.Kontextwunsch) (
	any, error,
) {
	return nil, u.fehler
}

// **Der Zähler** — er zählt jeden Weg.
func (z *zaehlender) Recherche(
	context.Context, string, string, int, string, string, string, []string,
) (any, error) {
	z.zaehle()

	return eineRecherche(), nil
}

func (z *zaehlender) AehnlicheKarten(
	context.Context, int64, string, int, string, string, []string,
) (any, error) {
	z.zaehle()

	return eineRecherche(), nil
}

func (z *zaehlender) Kontextpaket(context.Context, rest.Kontextwunsch) (
	any, error,
) {
	z.zaehle()

	return einKontextpaket(), nil
}
