package rest_test

// **Der Zähler für die abgestufte Rücksetzung** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Eine eigene Datei**, weil `vorrichtungzaehler_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge.

import "context"

// **Und die abgestufte Rücksetzung** (O4, #1428).
func (z *zaehlender) Ruecksetzungsarten(context.Context) (any, error) {
	z.zaehle()

	return []string{"layout"}, nil
}

func (z *zaehlender) Ruecksetzungsvorschau(
	context.Context, string, string, string,
) (any, error) {
	z.zaehle()

	return map[string]any{"betroffen": []any{}, "hindernisse": []any{}}, nil
}

func (z *zaehlender) Zuruecksetzen(
	context.Context, string, string, string, string, string, string,
) (any, error) {
	z.zaehle()

	return map[string]any{"zurueckgesetzt": 0}, nil
}

// **Und der Zähler für das Kostenjournal** (O12, #1439).
func (z *zaehlender) Kostenstand(context.Context, string, string) (any, error) {
	z.zaehle()

	return map[string]any{"verfuegbar": map[string]any{"wert": 0}}, nil
}

func (z *zaehlender) Kostenposten(context.Context, string, string) (any, error) {
	z.zaehle()

	return map[string]any{"posten": []any{}}, nil
}

func (z *zaehlender) KostenbudgetSetzen(
	context.Context, string, string, int64, string,
) (any, error) {
	z.zaehle()

	return map[string]any{"fehler": nil}, nil
}
