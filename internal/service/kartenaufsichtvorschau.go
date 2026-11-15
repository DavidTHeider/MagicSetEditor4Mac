package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau einer Massenänderung und das Anwenden mit Erwartung** —
// Werkstück U6 von Phase 5.1
// (#1417).
//
// # Warum hier und nicht im Speicher
//
// Die Vorschau ist eine **Lesung** mit denselben Regeln wie das Schreiben:
// Setgrenze ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md))
// und Feldregeln (#720). Wer sie im Speicher rechnete, hätte sie ohne beide
// — und eine Vorschau, die etwas zeigt, was das Schreiben dann abweist, ist
// eine Behauptung.
//
// **Die Zeilen entstehen in der Domäne** ([domain.Vorschauzeile]); hier
// steht nur, welche Karten gelesen werden und in welcher Reihenfolge — der,
// in der sie genannt wurden, denn so stehen sie in der Auswahl.

// FelderVorschau sagt, was ein Feldwert auf diesen Karten hieße — ohne zu
// schreiben.
func (a *Kartenaufsicht) FelderVorschau(
	ctx context.Context, setID int64, karteIDs []int64, feld, wert string,
) (domain.Massenvorschau, error) {
	vorschau := domain.Massenvorschau{
		Feld: feld, Wert: wert, Zeilen: make([]domain.Massenvorschauzeile, 0, len(karteIDs)),
	}

	if fehler := a.darfInDieMasse(ctx, setID, feld); fehler != nil {
		return domain.Massenvorschau{}, fehler
	}

	for _, karteID := range karteIDs {
		karte, fehler := a.gelesenAus(ctx, setID, karteID)
		if fehler != nil {
			return domain.Massenvorschau{}, fehler
		}

		// **Die Vorschau zeigt, was gespeichert würde** (#2516) — sonst stünde
		// dort der eingegebene Wert und im Bestand danach der gerechnete.
		neu := wert
		if rechner, rechnet := a.feld.(Massenrechner); rechnet {
			neu = rechner.Gerechnet(ctx, karte, feld, wert)
		}

		vorschau.Zeilen = append(vorschau.Zeilen, domain.Vorschauzeile(karte, feld, neu))
	}

	return vorschau, nil
}

// FelderSetzenMitErwartung schreibt, was die Vorschau gezeigt hat — und nur
// dann, wenn es noch so dasteht (siehe den Speicher).
func (a *Kartenaufsicht) FelderSetzenMitErwartung(
	ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string,
) (string, error) {
	if fehler := a.darfInDieMasse(ctx, setID, feld); fehler != nil {
		return "", fehler
	}

	for _, vorwert := range erwartet {
		if fehler := a.gehoertZu(ctx, setID, vorwert.KarteID); fehler != nil {
			return "", fehler
		}
	}

	karteIDs := make([]int64, 0, len(erwartet))
	for _, vorwert := range erwartet {
		karteIDs = append(karteIDs, vorwert.KarteID)
	}

	if rechner, rechnet := a.feld.(Massenrechner); rechnet {
		return rechner.MassenSetzen(ctx, karteIDs, feld, wert, erwartet)
	}

	return a.massen.FelderSetzenMitErwartung(ctx, erwartet, feld, wert)
}

// gelesenAus liest eine Karte und prüft die Setgrenze in einem Zug — die
// Vorschau braucht die Werte, die `gehoertZu` wegwirft.
func (a *Kartenaufsicht) gelesenAus(
	ctx context.Context, setID, karteID int64,
) (*domain.CardRow, error) {
	karte, err := a.sicht.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, err
	}

	if karte.SetID != setID {
		return nil, fremdeKarte(setID, karteID)
	}

	return karte, nil
}
