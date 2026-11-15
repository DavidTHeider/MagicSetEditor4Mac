package rest_test

// **Die Doppel für den Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Eine eigene Datei**, weil die drei Vorrichtungen daneben mit ihr über die
// Zeilengrenze aus §2 gingen — dieselbe Überlegung wie bei
// `zaehlerruecksetzung_test.go`.

import "context"

// **Der Willige** — er antwortet auf alles.
func (w williger) Indexstand(context.Context) (any, error) {
	return map[string]any{"aktiv": false}, nil
}

func (w williger) Indexsuche(
	context.Context, string, int, map[string][]string,
) (any, error) {
	return map[string]any{"treffer": []any{}, "modus": "ohne-index"}, nil
}

func (w williger) IndexAufbauen(
	context.Context, string, string, string, string, bool,
) (any, error) {
	return map[string]any{"generation": 1}, nil
}

func (w williger) IndexAbbrechen(context.Context) (any, error) {
	return map[string]bool{"abgebrochen": false}, nil
}

// **Der Unwillige** — er antwortet auf nichts.
func (u unwilliger) Indexstand(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Indexsuche(
	context.Context, string, int, map[string][]string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) IndexAufbauen(
	context.Context, string, string, string, string, bool,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) IndexAbbrechen(context.Context) (any, error) {
	return nil, u.fehler
}

// **Der Zähler** — er zählt jeden Weg.
func (z *zaehlender) Indexstand(context.Context) (any, error) {
	z.zaehle()

	return map[string]any{"aktiv": false}, nil
}

func (z *zaehlender) Indexsuche(
	context.Context, string, int, map[string][]string,
) (any, error) {
	z.zaehle()

	return map[string]any{"treffer": []any{}, "modus": "ohne-index"}, nil
}

func (z *zaehlender) IndexAufbauen(
	context.Context, string, string, string, string, bool,
) (any, error) {
	z.zaehle()

	return map[string]any{"generation": 1}, nil
}

func (z *zaehlender) IndexAbbrechen(context.Context) (any, error) {
	z.zaehle()

	return map[string]bool{"abgebrochen": false}, nil
}
