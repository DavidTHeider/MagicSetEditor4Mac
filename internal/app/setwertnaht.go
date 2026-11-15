package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// setwertnaht führt Lesen und Schreiben der Setangaben zusammen
// (#567).
//
// # Warum zwei Quellen und eine Naht
//
// Gelesen wird über den `SetLeser`, geschrieben über den `Kartenschreiber` —
// **derselbe**, der auch Kartenfelder schreibt, denn er führt den Verlauf. Zwei
// Schreiber wären zwei Urheber im Verlauf für dieselbe Handlung.
//
// Die Oberfläche stellt beides als **eine** Frage ([Setwertquelle]): sie will
// die Angaben eines Sets, nicht zwei Nähte dorthin. Hier steht die Naht, die
// aus zweien eine macht — im Composition Root, wo die Zusammensetzung hingehört
// (§2).
type setwertnaht struct {
	leser     Setwertleser
	schreiber Setwertschreiber
}

// Setwertleser gibt die Angaben eines Sets.
//
// **Als Schnittstelle und nicht als `*storage.SetLeser`** (§2): `storage` darf
// nur der Composition Root kennen, und ein Wächter zählt es nach
// (`internal/schichten`). Wegen Gos struktureller Typisierung erfüllt der Leser
// sie ohne davon zu wissen.
type Setwertleser interface {
	Setwerte(ctx context.Context, setID int64) ([]domain.Value, error)

	// Setkopf gibt die fünf Formatangaben (#1325).
	Setkopf(ctx context.Context, setID int64) (domain.Setkopf, error)
}

// Setwertschreiber schreibt eine Angabe und gibt die vorige zurück.
type Setwertschreiber interface {
	SetwertSetzenMitVorwert(ctx context.Context, setID int64, feld, wert string) (
		vorher string, hatte bool, fehler error,
	)
}

func (n setwertnaht) Setwerte(ctx context.Context, setID int64) ([]domain.Value, error) {
	return n.leser.Setwerte(ctx, setID)
}

func (n setwertnaht) Setkopf(
	ctx context.Context, setID int64,
) (domain.Setkopf, error) {
	return n.leser.Setkopf(ctx, setID)
}

func (n setwertnaht) SetwertSetzenMitVorwert(
	ctx context.Context, setID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	return n.schreiber.SetwertSetzenMitVorwert(ctx, setID, feld, wert)
}
