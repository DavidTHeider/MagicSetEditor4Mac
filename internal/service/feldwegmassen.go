package service

// **Eine Massenänderung speichert, was gerechnet ist** — #2516 (Rest aus
// #2396, Vermerk 0203).
//
// Der Wert rechnet auf **jeder** Karte, und wer ihn liest, zieht nach — derselbe
// Nachzug wie beim einzelnen Feld ([feldnachzug.nachgezogen]). Geschrieben
// wird je Karte auf dem Stand, aus dem gerechnet wurde, alles in einem
// Schritt ([Feldablage.FelderGerechnetSetzen]). MSE2 hat keine
// Massenänderung; die Marke ist `new`.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MassenSetzen setzt `feld` auf vielen Karten und gibt den Vorgang zurück.
// `erwartet` ist die Erwartung aus der Vorschau (U6, #1417), `nil` keine.
//
// **Eine unlesbare Vorlage hält nichts auf** — die Karte bekommt den
// Rohwert, wie beim einzelnen Feld (#720), und ohne Stand, denn gerechnet
// wurde nichts.
func (f *Feldweg) MassenSetzen(
	ctx context.Context, karteIDs []int64, feld, wert string, erwartet []domain.Feldvorwert,
) (string, error) {
	schritte := make([]domain.Kartenschritt, 0, len(karteIDs))

	for _, karteID := range karteIDs {
		werte, stand := []domain.Value{{Field: feld, Text: wert}}, ""

		if karte, _, lesefehler := f.karteUndFelder(ctx, karteID); lesefehler == nil {
			werte, stand = f.nachzug.nachgezogen(ctx, karte, werte)
		}

		schritte = append(schritte, domain.Kartenschritt{KarteID: karteID, Werte: werte, Erwartet: stand})
	}

	return f.ablage.FelderGerechnetSetzen(ctx, schritte, erwartet, feld)
}

// Gerechnet gibt, was `wert` in `feld` dieser Karte gespeichert hieße — für
// die Vorschau, damit sie zeigt, was das Schreiben dann tut.
func (f *Feldweg) Gerechnet(ctx context.Context, karte *domain.CardRow, feld, wert string) string {
	werte, _ := f.nachzug.nachgezogen(ctx, karte, []domain.Value{{Field: feld, Text: wert}})

	return werte[0].Text
}
