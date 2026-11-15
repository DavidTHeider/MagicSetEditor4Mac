package rest_test

import "context"

// DateiAnsehen gibt eine leere Ansicht — die Vorrichtung prüft den Weg bis
// zum Anwendungsfall, nicht ihn selbst.
func (williger) DateiAnsehen(context.Context, string) (any, error) {
	return map[string]any{"eintraege": []any{}}, nil
}

// DateiAnsehen zählt mit.
func (z *zaehlender) DateiAnsehen(context.Context, string) (any, error) {
	z.zaehle()

	return map[string]any{"eintraege": []any{}}, nil
}

// DateiAnsehen gibt den Fehler, mit dem es gebaut ist.
func (u unwilliger) DateiAnsehen(context.Context, string) (any, error) { return nil, u.fehler }
