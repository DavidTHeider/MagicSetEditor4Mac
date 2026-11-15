// **Der zählende Durchreicher für die Turnierwege** — Werkstück 5 von Phase
// 4.7 (#1294).
//
// **Eine eigene Datei**, damit `vorrichtungdecks_test.go` unter der
// Zeilengrenze aus §2 bleibt.
//
// # Warum jedes Verfahren einzeln zählt
//
// Weil `TestJederEndpunktKenntSeineVierKlassen` prüft, dass ein gültiger Ruf
// den Anwendungsfall **erreicht**. Ohne diese Hüllen erbte `zaehlender` die
// Antworten von `williger` und zählte nichts — der Lauf meldete „0mal
// gerufen", und ein Weg, der nie ankommt, sähe wie ein Weg aus, den es nicht
// gibt.

package rest_test

import "context"

func (z *zaehlender) Turniere(ctx context.Context, spiel string) (any, error) {
	z.zaehle()

	return z.williger.Turniere(ctx, spiel)
}

func (z *zaehlender) Turnier(ctx context.Context, turnierID int64) (any, error) {
	z.zaehle()

	return z.williger.Turnier(ctx, turnierID)
}

func (z *zaehlender) TurnierRangliste(ctx context.Context, turnierID int64) (
	any, error,
) {
	z.zaehle()

	return z.williger.TurnierRangliste(ctx, turnierID)
}

func (z *zaehlender) TurnierPaarungsvorschlag(ctx context.Context, turnierID int64) (
	any, error,
) {
	z.zaehle()

	return z.williger.TurnierPaarungsvorschlag(ctx, turnierID)
}

func (z *zaehlender) TurnierAnlegen(
	ctx context.Context, name, spiel, notiz string, einzelspiel bool, formatID int64,
) (any, error) {
	z.zaehle()

	return z.williger.TurnierAnlegen(ctx, name, spiel, notiz, einzelspiel, formatID)
}

func (z *zaehlender) TurnierAbschliessen(ctx context.Context, turnierID int64) error {
	z.zaehle()

	return z.williger.TurnierAbschliessen(ctx, turnierID)
}

func (z *zaehlender) TurnierWiedereroeffnen(ctx context.Context, turnierID int64) error {
	z.zaehle()

	return z.williger.TurnierWiedereroeffnen(ctx, turnierID)
}

func (z *zaehlender) TurnierProtokoll(ctx context.Context, turnierID int64) (any, error) {
	z.zaehle()

	return z.williger.TurnierProtokoll(ctx, turnierID)
}

func (z *zaehlender) TurnierAusgabe(ctx context.Context, turnierID int64) (any, error) {
	z.zaehle()

	return z.williger.TurnierAusgabe(ctx, turnierID)
}

func (z *zaehlender) TurnierBericht(ctx context.Context, turnierID int64) (any, error) {
	z.zaehle()

	return z.williger.TurnierBericht(ctx, turnierID)
}

func (z *zaehlender) TurnierEinlesen(ctx context.Context, inhalt string) (any, error) {
	z.zaehle()

	return z.williger.TurnierEinlesen(ctx, inhalt)
}

func (z *zaehlender) TurnierKorrekturfolgen(
	ctx context.Context, turnierID, paarungID int64, stand map[string]any,
) (any, error) {
	z.zaehle()

	return z.williger.TurnierKorrekturfolgen(ctx, turnierID, paarungID, stand)
}

func (z *zaehlender) TurnierKopieren(
	ctx context.Context, turnierID int64, name string, mitTeilnehmenden bool,
) (any, error) {
	z.zaehle()

	return z.williger.TurnierKopieren(ctx, turnierID, name, mitTeilnehmenden)
}

func (z *zaehlender) TurnierLoeschen(ctx context.Context, turnierID int64) error {
	z.zaehle()

	return z.williger.TurnierLoeschen(ctx, turnierID)
}

func (z *zaehlender) TurnierAnmelden(
	ctx context.Context, turnierID int64, name string, deckID int64,
) (any, error) {
	z.zaehle()

	return z.williger.TurnierAnmelden(ctx, turnierID, name, deckID)
}

func (z *zaehlender) TurnierAussteigen(
	ctx context.Context, turnierID, teilnehmerID int64, ausgestiegen bool,
) error {
	z.zaehle()

	return z.williger.TurnierAussteigen(ctx, turnierID, teilnehmerID, ausgestiegen)
}

func (z *zaehlender) TurnierRundeLegen(ctx context.Context, turnierID int64) (
	any, error,
) {
	z.zaehle()

	return z.williger.TurnierRundeLegen(ctx, turnierID)
}

func (z *zaehlender) TurnierErgebnisSetzen(
	ctx context.Context, turnierID, paarungID int64, stand map[string]any,
) error {
	z.zaehle()

	return z.williger.TurnierErgebnisSetzen(ctx, turnierID, paarungID, stand)
}
