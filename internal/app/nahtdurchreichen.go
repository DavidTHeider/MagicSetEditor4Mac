// **Die Weiterreichungen der beiden Nähte** — abgetrennt am 2026-09-02, weil
// `mcpnaht.go` mit `FelderSetzen` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **was zusammengesetzt wird** — welcher Speicher, welche
// Aufsicht, welcher Feldweg. Hier steht, **welche Methode wohin geht**: eine
// Reihe fast gleicher Zeilen, die genau eine Aussage tragen — dieser Aufruf
// nimmt den Umweg über den Dienst, jener nicht.
//
// **Und die Reihe ist kein Zufall.** Jede Methode hier verdeckt eine
// eingebettete; ohne sie ginge der Aufruf am Dienst vorbei, und die Setgrenze
// wäre eine Zusage ohne Weg
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
//
// # Warum `feldnaht` drüben geblieben ist
//
// Weil sie `*storage.Kartenschreiber` einbettet, und `storage` darf nur ein
// **Composition Root** kennen (§2). Diese Datei ist keiner — sie reicht
// weiter, sie setzt nichts zusammen. Der Schichtenwächter hat es angestrichen,
// bevor jemand darüber nachdenken musste.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (m mcpnaht) FeldSetzenMitVorwert(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	return m.aufsicht.FeldSetzenMitVorwert(ctx, setID, karteID, feld, wert)
}

// FeldSetzenVonAussen hält fest, woher der Wert kam (#888).
func (m mcpnaht) FeldSetzenVonAussen(
	ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return m.aufsicht.FeldSetzenVonAussen(ctx, setID, karteID, feld, wert, herkunft)
}

// FelderVonAussen schreibt mehrere übernommene Werte in einem Zug (#1026).
func (m mcpnaht) FelderVonAussen(
	ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return m.aufsicht.FelderVonAussen(ctx, setID, karteID, werte, herkunft)
}

// FelderVonAussenMitHerkunft schreibt Werte samt Herkunft in einem Zug
// (A11, #1450).
// GruppeVonAussenMitHerkunft reicht die atomare Gruppe durch (A13.1b, #1452).
func (m mcpnaht) GruppeVonAussenMitHerkunft(
	ctx context.Context, schritte []domain.Kartenschritt,
) error {
	return m.aufsicht.GruppeVonAussenMitHerkunft(ctx, schritte)
}

func (m mcpnaht) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	return m.aufsicht.FelderVonAussenMitHerkunft(
		ctx, setID, karteID, werte, herkunft, herkuenfte, angaben, erwartet)
}

func (m mcpnaht) FeldEntfernen(
	ctx context.Context, setID, karteID int64, feld string,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	return m.aufsicht.FeldEntfernen(ctx, setID, karteID, feld)
}

func (m mcpnaht) FeldSetzenAnStelle(
	ctx context.Context, setID, karteID int64, feld, wert string, stelle int64,
) (vorher string, hatte bool, fehler error) {
	return m.aufsicht.FeldSetzenAnStelle(ctx, setID, karteID, feld, wert, stelle)
}

func (m mcpnaht) KarteAnlegen(ctx context.Context, setID int64) (int64, error) {
	return m.aufsicht.KarteAnlegen(ctx, setID)
}

func (m mcpnaht) KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error) {
	return m.aufsicht.KarteKopieren(ctx, setID, karteID)
}

func (m mcpnaht) KartenLoeschen(
	ctx context.Context, setID int64, karteIDs []int64, trotzdem bool,
) error {
	return m.aufsicht.KartenLoeschen(ctx, setID, karteIDs, trotzdem)
}

func (m mcpnaht) KarteLoeschen(
	ctx context.Context, setID, karteID int64, trotzdem bool,
) error {
	return m.aufsicht.KarteLoeschen(ctx, setID, karteID, trotzdem)
}

func (m mcpnaht) KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error {
	return m.aufsicht.KartenTauschen(ctx, setID, eineID, andereID)
}

// FelderSetzen geht über die Aufsicht und nicht an den eingebetteten
// Schreiber — #1390.
//
// **Der Name steht hier zweimal, und das ist der Grund für diese Methode.**
// `storage.Kartenschreiber` hat ein `FelderSetzen` **ohne** Setkennung; das
// eingebettete verdeckte das der Aufsicht, und der Übersetzer hat es
// angestrichen. Ohne diese Zeile ginge die Massenänderung an der Setgrenze
// und an der Vorlage vorbei — genau der Befund aus
// [0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md),
// nur an einer neuen Stelle.
func (m mcpnaht) FelderSetzen(
	ctx context.Context, setID int64, karteIDs []int64, feld, wert string,
) (string, error) {
	return m.aufsicht.FelderSetzen(ctx, setID, karteIDs, feld, wert)
}

// FelderVorschau geht denselben Weg wie das Schreiben — Setgrenze und
// Feldregeln der Aufsicht (Werkstück U6, #1417).
func (m mcpnaht) FelderVorschau(
	ctx context.Context, setID int64, karteIDs []int64, feld, wert string,
) (domain.Massenvorschau, error) {
	return m.aufsicht.FelderVorschau(ctx, setID, karteIDs, feld, wert)
}

// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
// gezeigten Altwert steht.
func (m mcpnaht) FelderSetzenMitErwartung(
	ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string,
) (string, error) {
	return m.aufsicht.FelderSetzenMitErwartung(ctx, setID, erwartet, feld, wert)
}
