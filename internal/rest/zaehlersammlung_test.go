// Die zählenden Sammlungswege — Werkstück 5 von Phase 4.5
// (#1292).
//
// **Eine eigene Datei**, weil `zaehlerdoppel_test.go` an der Zeilengrenze aus
// §2 stand.
package rest_test

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// Die vier Sammlungssichten zählen mit (#1292).
func (z *zaehlender) SammlungBesitz(context.Context, string) (any, error) {
	z.rufe++

	return map[string]any{"stueck": 1, "besitzt": true}, nil
}

func (z *zaehlender) SammlungLagerorte(context.Context) (any, error) {
	z.rufe++

	return []any{}, nil
}

func (z *zaehlender) SammlungDubletten(context.Context, int) (any, error) {
	z.rufe++

	return []any{}, nil
}

func (z *zaehlender) SammlungWuensche(context.Context) (any, error) {
	z.rufe++

	return []any{}, nil
}

// Die drei Massenvorgänge zählen mit (#1292).
func (z *zaehlender) MasseAendern(context.Context, rest.Massenaenderung) (any, error) {
	z.rufe++

	return map[string]any{"geaendert": 1, "befunde": []string{}}, nil
}

func (z *zaehlender) MasseLoeschen(context.Context, []rest.Massenzeile) (any, error) {
	z.rufe++

	return map[string]any{"geaendert": 1, "befunde": []string{}}, nil
}

func (z *zaehlender) MassePreise(context.Context, int) (any, error) {
	z.rufe++

	return map[string]any{"geholt": 0, "offen": 0, "fertig": true}, nil
}

// **Die Kartenskripte** (#1337).
func (z *zaehlender) Kartenskripte(context.Context, int64) (any, error) {
	z.rufe++

	return []map[string]any{}, nil
}

func (z *zaehlender) KartenskriptAusfuehren(context.Context, int64, string) (any, error) {
	z.rufe++

	return map[string]any{"karten": []string{}, "befunde": []string{}}, nil
}

// **Die Auszählung** (#1338).
func (z *zaehlender) Auszaehldimensionen(context.Context, int64) (any, error) {
	z.rufe++

	return []map[string]any{}, nil
}

func (z *zaehlender) Auszaehlen(context.Context, int64, string) (any, error) {
	z.rufe++

	return map[string]any{
		"dimension": "farbe", "gruppen": []map[string]any{}, "befunde": []string{},
	}, nil
}

// **Die Parameterhilfen** (#1342).
func (z *zaehlender) Parameterhilfen(context.Context, int64, int64) (any, error) {
	z.rufe++

	return map[string]any{
		"arten": []map[string]any{}, "verweise": []map[string]any{},
		"befunde": []string{},
	}, nil
}

// **Die Autoersetzungen** (#1341).
func (z *zaehlender) Autoersetzungen(context.Context, int64) (any, error) {
	z.rufe++

	return map[string]any{"regeln": []map[string]any{}, "an": true}, nil
}

func (z *zaehlender) AutoersetzungSetzen(
	context.Context, int64, string, string, bool, bool,
) error {
	z.rufe++

	return nil
}

func (z *zaehlender) AutoersetzungLoeschen(context.Context, int64, string) error {
	z.rufe++

	return nil
}

func (z *zaehlender) AutoersetzungenZuruecksetzen(context.Context, int64) error {
	z.rufe++

	return nil
}

func (z *zaehlender) AutoersetzungSchalten(context.Context, int64, bool) error {
	z.rufe++

	return nil
}

// **Und das Nachzeichnen** (#1340).
func (z *zaehlender) SymbolNachzeichnen(context.Context, int64, string) (any, error) {
	z.zaehle()

	return map[string]any{"name": "symbol.mse-symbol", "teile": 1, "punkte": 12}, nil
}

// **Der Symboleditor** (#1391) — jeder Weg zählt mit.
func (z *zaehlender) SymbolLesen(context.Context, int64, string) (any, error) {
	z.rufe++

	return []map[string]any{}, nil
}

func (z *zaehlender) SymbolAblegen(context.Context, int64, json.RawMessage, string) (string, error) {
	z.rufe++

	return "probe.mse-symbol", nil
}

func (z *zaehlender) SymbolRechnen(string, json.RawMessage) (any, error) {
	z.rufe++

	return []map[string]any{}, nil
}

func (z *zaehlender) Palette(context.Context, int64, string) (any, error) {
	z.rufe++

	return []map[string]any{}, nil
}
