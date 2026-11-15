// **Der Verweis auf einen echten Druck für die Flächen ohne Fenster** —
// Werkstück 2 von Phase 3.6
// (#1167).
//
// **Dieselbe Fähigkeit, überall dieselbe Antwort** (§16): was das Fenster
// kann, können Kommandozeile, MCP und Dienst. Die Naht baut den Zugang je
// Aufruf, weil es dort keinen Start gibt, an dem sie einmal entstünde.

package app

import (
	"context"
	"database/sql"
)

// Fremdverweise ist der Verweisdienst für eine Fläche ohne Fenster.
type Fremdverweise struct{ quelle Verweisquelle }

// Verweisflaechennaht baut ihn über einem Bestand.
func Verweisflaechennaht(db *sql.DB) (*Fremdverweise, error) {
	dienst, err := Verweisnaht(dienstwahlUeber(db), db)
	if err != nil {
		return nil, err
	}

	return &Fremdverweise{quelle: dienst}, nil
}

// MitVerweisquelle setzt die Naht — für Tests.
func (f *Fremdverweise) MitVerweisquelle(quelle Verweisquelle) *Fremdverweise {
	f.quelle = quelle

	return f
}

// Setzen bezieht eine Karte auf einen Druck.
//
// **Dieselbe Gestalt wie am Fenster** ([KartenverweisDTO]), ohne den Fehler:
// die Flächen ohne Fenster tragen ihn als `error` (§8).
func (f *Fremdverweise) Setzen(ctx context.Context, karteID int64, printing string) (
	KartenverweisDTO, error,
) {
	verweis, _, fehler := f.quelle.Setzen(ctx, karteID, printing)
	if fehler != nil {
		return KartenverweisDTO{}, fehler
	}

	return alsVerweisDTO(karteID, *verweis), nil
}

// Fuer gibt den Verweis einer Karte — `nil`, wenn sie keinen hat.
func (f *Fremdverweise) Fuer(ctx context.Context, karteID int64) (
	*KartenverweisDTO, error,
) {
	verweis, fehler := f.quelle.Fuer(ctx, karteID)
	if fehler != nil || verweis == nil {
		return nil, fehler
	}

	gestalt := alsVerweisDTO(karteID, *verweis)

	return &gestalt, nil
}

// Loesen nimmt den Verweis einer Karte zurück.
func (f *Fremdverweise) Loesen(ctx context.Context, karteID int64) error {
	return f.quelle.Loesen(ctx, karteID)
}

// JeSet gibt die Verweise eines Sets, nach Kartenkennung geordnet.
func (f *Fremdverweise) JeSet(ctx context.Context, setID int64) (
	[]KartenverweisDTO, error,
) {
	verweise, fehler := f.quelle.JeSet(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	gestalt := make([]KartenverweisDTO, 0, len(verweise))
	for karteID, verweis := range verweise {
		gestalt = append(gestalt, alsVerweisDTO(karteID, verweis))
	}

	sortiereVerweise(gestalt)

	return gestalt, nil
}
