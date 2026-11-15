// Die Wortlisten für den Dienst
// (#1074).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// wortlisten führt die beiden Endpunkte an die Naht dieser Schicht.
type wortlisten struct{ db *sql.DB }

// Wortlisten gibt die Auswahl eines Feldes.
func (w wortlisten) Wortlisten(ctx context.Context, karteID int64, feld string) (
	[]rest.Wortbereich, error,
) {
	naht, err := app.WortlistennahtUeber(w.db, domain.UrheberDienst)
	if err != nil {
		return nil, err
	}

	bereiche, lesefehler := naht.Wortlisten(ctx, karteID, feld)
	if lesefehler != nil {
		return nil, lesefehler
	}

	umgesetzt := make([]rest.Wortbereich, 0, len(bereiche))

	for _, bereich := range bereiche {
		umgesetzt = append(umgesetzt, rest.Wortbereich{
			Liste:     bereich.Liste,
			Text:      bereich.Text,
			Kaestchen: bereich.Kaestchen,
			Gewaehlt:  bereich.Gewaehlt,
			Eintraege: eintraege(bereich.Eintraege),
		})
	}

	return umgesetzt, nil
}

// Wortwahl wählt einen Eintrag und schreibt das Feld.
func (w wortlisten) Wortwahl(
	ctx context.Context, setID, karteID int64, feld string, bereich int, weg []int,
) error {
	naht, err := app.WortlistennahtUeber(w.db, domain.UrheberDienst)
	if err != nil {
		return err
	}

	return naht.Waehlen(ctx, setID, karteID, feld, bereich, weg)
}

// eintraege setzt ein Menü samt seiner Untermenüs um.
func eintraege(quelle []app.WortmenueeintragDTO) []rest.Worteintrag {
	umgesetzt := make([]rest.Worteintrag, 0, len(quelle))

	for _, eintrag := range quelle {
		umgesetzt = append(umgesetzt, rest.Worteintrag{
			Name:          eintrag.Name,
			IstPraefix:    eintrag.IstPraefix,
			ZeileDarunter: eintrag.ZeileDarunter,
			Aktiv:         eintrag.Aktiv,
			Kinder:        eintraege(eintrag.Kinder),
		})
	}

	return umgesetzt
}
