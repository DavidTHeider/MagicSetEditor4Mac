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

func (williger) KIVorschlag(context.Context, rest.KIVorschlagsbezug) (
	rest.KIVorschlagssicht, error,
) {
	return rest.KIVorschlagssicht{
		Kennung: "ki-1-1", Modell: "beispiel:3b",
		Felder: []rest.KIFeldwertsicht{{Feld: "name", Werte: []string{"Probe"}}},
	}, nil
}

// **Und die zwei Wege der Messung** — A9, #1448.
func (williger) KIModellPruefen(_ context.Context, bezug rest.KIProbenbezug) (
	rest.KIProbensicht, error,
) {
	return rest.KIProbensicht{Proben: []rest.KIProbesicht{{
		Anbieter: bezug.Anbieter, Alias: bezug.Alias,
		Faehigkeit: "strukturiert", Datum: "2026-09-12",
		Beleg: "gemessen", Modell: bezug.Alias,
	}}}, nil
}

func (williger) KIModellVergessen(context.Context, rest.KIProbenbezug) error {
	return nil
}

func (williger) KIGeraetemodellFuerSitzung(_ context.Context, an bool) (bool, error) {
	return an, nil
}

// KIBatch antwortet für jedes Ziel — A13.1 (#1452).
func (williger) KIBatch(_ context.Context, bezug rest.KIBatchbezug) (
	rest.KIBatchsicht, error,
) {
	ziele := make([]rest.KIBatchlagensicht, 0, len(bezug.KarteIDs))

	for _, karte := range bezug.KarteIDs {
		ziele = append(ziele, rest.KIBatchlagensicht{
			KarteID: karte, Kennung: "ki-1-1", Modell: bezug.Alias,
			Felder: []rest.KIFeldwertsicht{
				{Feld: "name", Werte: []string{"Probe"}},
			},
			Befunde: []string{},
		})
	}

	return rest.KIBatchsicht{
		Kennung: "kibatch-1", Anbieter: bezug.Anbieter, Ziele: ziele,
	}, nil
}

// KIEinschaetzung antwortet mit einer belegten Einschätzung — A13.2a (#1452).
func (williger) KIEinschaetzung(
	_ context.Context, bezug rest.KIEinschaetzungsbezug,
) (any, error) {
	return map[string]any{
		"art": bezug.Art, "satz": "Der Block ist legal [B1].",
		"belegt": true, "modell": bezug.Alias,
	}, nil
}

// KIGruppe übernimmt die gewählten Vorschläge — A13.1b (#1452).
//
// **Es antwortet und scheitert nicht**: das willige Doppel ist der Weg, auf
// dem ein Endpunkt seinen Erfolgsfall zeigt. Wer eine Gruppe prüfen will, die
// scheitert, nimmt `unwilliger`.
func (williger) KIGruppe(_ context.Context, bezug rest.KIGruppenbezug) (
	rest.KIGruppensicht, error,
) {
	namen := make([]rest.KIGruppenkartensicht, 0, len(bezug.Wahl))

	felder := 0

	for i, eine := range bezug.Wahl {
		felder += len(eine.Felder)

		namen = append(namen, rest.KIGruppenkartensicht{
			KarteID: int64(i + 1), Felder: eine.Felder,
		})
	}

	return rest.KIGruppensicht{
		Karten: len(namen), Felder: felder, Namen: namen,
	}, nil
}

// KIErkennung liest ein Kartenfoto — A13.3 (#1452).
//
// **Es antwortet und scheitert nicht**: das willige Doppel ist der Weg, auf
// dem ein Endpunkt seinen Erfolgsfall zeigt.
func (williger) KIErkennung(_ context.Context, bezug rest.KIErkennungsbezug) (
	any, error,
) {
	return map[string]any{
		"beobachtungen": []any{map[string]any{
			"art": "name", "wert": "Probe", "lesbar": true,
		}},
		"nichtGelesen": []string{},
		"befunde":      []any{},
		"kandidaten":   []any{},
		"drucke":       0,
		"anbieter":     bezug.Anbieter,
	}, nil
}
