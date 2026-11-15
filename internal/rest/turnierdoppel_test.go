// **Die Doppel für die Turnierwege** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `deckdoppel_test.go` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort die **Sammlung**, hier der
// **Verlauf**.
//
// # Warum beide Doppel jedes Verfahren haben
//
// Weil `rest.Naehte` sie als **eine** Schnittstelle nimmt. Ein Doppel, dem
// eines fehlt, ist kein `Turnierwerkstatt` — und ein Feld, das `nil` bleibt,
// lässt den Ruf abstürzen statt antworten. Genau das ist beim ersten Anlauf
// passiert: `TestJederEndpunktReichtSeinenFehlerDurch` gab einen
// Nil-Zeigerfehler, weil `Turniere` in der feindlichen Vorrichtung fehlte.

package rest_test

import "context"

func (w williger) Turniere(context.Context, string) (any, error) {
	return []any{}, nil
}

func (w williger) Turnier(context.Context, int64) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) TurnierRangliste(context.Context, int64) (any, error) {
	return []any{}, nil
}

func (w williger) TurnierPaarungsvorschlag(context.Context, int64) (any, error) {
	return []any{}, nil
}

func (w williger) TurnierAnlegen(
	context.Context, string, string, string, bool, int64,
) (any, error) {
	return map[string]any{"id": "1"}, nil
}

func (w williger) TurnierAbschliessen(context.Context, int64) error    { return nil }
func (w williger) TurnierWiedereroeffnen(context.Context, int64) error { return nil }

func (w williger) TurnierProtokoll(context.Context, int64) (any, error) {
	return []any{}, nil
}

func (w williger) TurnierAusgabe(context.Context, int64) (any, error) {
	return map[string]any{"inhalt": "{}"}, nil
}

func (w williger) TurnierBericht(context.Context, int64) (any, error) {
	return map[string]any{"inhalt": "# T"}, nil
}

func (w williger) TurnierEinlesen(context.Context, string) (any, error) {
	return map[string]any{"kennung": "2"}, nil
}

func (w williger) TurnierKorrekturfolgen(
	context.Context, int64, int64, map[string]any,
) (any, error) {
	return map[string]any{"runde": 1}, nil
}

func (w williger) TurnierKopieren(context.Context, int64, string, bool) (any, error) {
	return map[string]any{"id": "2"}, nil
}

func (w williger) TurnierLoeschen(context.Context, int64) error { return nil }

func (w williger) TurnierAnmelden(
	context.Context, int64, string, int64,
) (any, error) {
	return map[string]any{"id": "2"}, nil
}

func (w williger) TurnierAussteigen(context.Context, int64, int64, bool) error {
	return nil
}

func (w williger) TurnierRundeLegen(context.Context, int64) (any, error) {
	return map[string]any{"id": "3", "nummer": 1}, nil
}

func (w williger) TurnierErgebnisSetzen(
	context.Context, int64, int64, map[string]any,
) error {
	return nil
}

func (u unwilliger) Turniere(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Turnier(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierRangliste(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierPaarungsvorschlag(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierAnlegen(
	context.Context, string, string, string, bool, int64,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierAbschliessen(context.Context, int64) error    { return u.fehler }
func (u unwilliger) TurnierWiedereroeffnen(context.Context, int64) error { return u.fehler }

func (u unwilliger) TurnierProtokoll(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierAusgabe(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierBericht(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierEinlesen(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierKorrekturfolgen(
	context.Context, int64, int64, map[string]any,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierKopieren(context.Context, int64, string, bool) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierLoeschen(context.Context, int64) error { return u.fehler }

func (u unwilliger) TurnierAnmelden(
	context.Context, int64, string, int64,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierAussteigen(context.Context, int64, int64, bool) error {
	return u.fehler
}

func (u unwilliger) TurnierRundeLegen(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) TurnierErgebnisSetzen(
	context.Context, int64, int64, map[string]any,
) error {
	return u.fehler
}
