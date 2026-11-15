package rest_test

// **Der KI-Vorschlag am Doppel** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Eine eigene Datei**, weil das Doppel daneben mit ihr über die
// Zeilengrenze aus §2 ginge.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (z *zaehlender) KIVorschlag(context.Context, rest.KIVorschlagsbezug) (
	rest.KIVorschlagssicht, error,
) {
	z.rufe++

	return rest.KIVorschlagssicht{}, nil
}

// **Und die zwei Wege der Messung** — A9, #1448.
func (z *zaehlender) KIModellPruefen(context.Context, rest.KIProbenbezug) (
	rest.KIProbensicht, error,
) {
	z.rufe++

	return rest.KIProbensicht{}, nil
}

func (z *zaehlender) KIModellVergessen(context.Context, rest.KIProbenbezug) error {
	z.rufe++

	return nil
}

// KIGeraetemodellFuerSitzung zählt mit — AI-05 (#2042).
func (z *zaehlender) KIGeraetemodellFuerSitzung(_ context.Context, an bool) (bool, error) {
	z.rufe++

	return an, nil
}

// KIBatch zählt mit — A13.1 (#1452).
func (z *zaehlender) KIBatch(context.Context, rest.KIBatchbezug) (
	rest.KIBatchsicht, error,
) {
	z.rufe++

	return rest.KIBatchsicht{Ziele: []rest.KIBatchlagensicht{}}, nil
}

// KIEinschaetzung zählt mit — A13.2a (#1452).
func (z *zaehlender) KIEinschaetzung(
	context.Context, rest.KIEinschaetzungsbezug,
) (any, error) {
	z.rufe++

	return map[string]any{}, nil
}
