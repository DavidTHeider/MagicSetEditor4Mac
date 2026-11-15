package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Wege der Feldnaht, die rechnen** — in einer eigenen Datei, weil
// `mcpnaht.go` an der Zeilengrenze steht (§2).

// FeldSetzenMitErgebnis geht denselben Weg wie das Setzen und nennt, **was
// gespeichert ist** (#2535) — für den Entwurf im Fenster, der genau diesen
// Wert als seinen bestätigt.
func (f feldnaht) FeldSetzenMitErgebnis(
	ctx context.Context, karteID int64, feld, wert string,
) (gespeichert string, bekannt bool, fehler error) {
	return f.weg.FeldSetzenMitErgebnis(ctx, karteID, feld, wert)
}

// MassenSetzen rechnet eine Massenänderung je Karte (#2516).
func (f feldnaht) MassenSetzen(
	ctx context.Context, karteIDs []int64, feld, wert string, erwartet []domain.Feldvorwert,
) (string, error) {
	return f.weg.MassenSetzen(ctx, karteIDs, feld, wert, erwartet)
}

// Gerechnet sagt der Vorschau, was gespeichert würde (#2516).
func (f feldnaht) Gerechnet(ctx context.Context, karte *domain.CardRow, feld, wert string) string {
	return f.weg.Gerechnet(ctx, karte, feld, wert)
}

// KarteGerechnetAnlegen legt eine Karte mit dem an, was ihre Skripte rechnen
// (#2516) — in einem Schritt mit der Karte selbst.
func (f feldnaht) KarteGerechnetAnlegen(ctx context.Context, setID int64) (int64, error) {
	return f.KarteAnlegenMitWerten(ctx, setID, f.weg.Anfangswerte(ctx, setID, nil))
}
