package mcpnaehte

// **Die Regelfreigabe am MCP-Server** — Werkstück G4 von Phase 5.5
// (#1461).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// regelfreigabe reicht die Frage an dieselbe Fläche wie das Fenster.
type regelfreigabe struct{ db *sql.DB }

// Regelfreigabe gibt die Auskunft über das, was durchgesetzt wird.
func (r regelfreigabe) Regelfreigabe(
	ctx context.Context, spiel, strenge string, sitze int,
	modus string, abgewaehlt []string,
) (mcp.Regelfreigabeangaben, error) {
	antwort := app.RegelflaecheUeber(r.db).Regelfreigabe(
		ctx, spiel, strenge, sitze, modus, abgewaehlt)
	if antwort.Fehler != nil {
		return mcp.Regelfreigabeangaben{}, antwort.Fehler.AlsFehler()
	}

	freigabe := antwort.Freigabe

	angaben := mcp.Regelfreigabeangaben{
		Erlaubt:   freigabe.Erlaubt,
		Regelwerk: freigabe.Regelwerk,
		Fassung:   freigabe.Fassung,
		Wirksam:   freigabe.Wirksam,
		Hoechste:  freigabe.Hoechste,

		// **Leere Listen und kein `nil`** — dieselbe Zusage wie an jeder
		// anderen Fläche.
		Matrix:      make([]mcp.Regelbereichsangaben, 0, len(freigabe.Matrix)),
		Luecken:     gefuelltOderLeer(freigabe.Luecken),
		Hindernisse: gefuelltOderLeer(freigabe.Hindernisse),

		// **Modus, Baum und Widersprüche kamen mit PLAY-02** (#1710).
		Modus: freigabe.Modus,
		Baum:  make([]mcp.Regelgruppenangaben, 0, len(freigabe.Baum)),
		Widersprueche: make([]mcp.Regelwiderspruchsangaben, 0,
			len(freigabe.Widersprueche)),
	}

	for _, gruppe := range freigabe.Baum {
		angaben.Baum = append(angaben.Baum, mcp.Regelgruppenangaben{
			Kennung:  gruppe.Kennung,
			Bereiche: gefuelltOderLeer(gruppe.Bereiche),
		})
	}

	for _, widerspruch := range freigabe.Widersprueche {
		angaben.Widersprueche = append(angaben.Widersprueche,
			mcp.Regelwiderspruchsangaben{
				Bereich: widerspruch.Bereich,
				Haengt:  widerspruch.Haengt,
				Quelle:  widerspruch.Quelle,
			})
	}

	for _, eintrag := range freigabe.Matrix {
		angaben.Matrix = append(angaben.Matrix, mcp.Regelbereichsangaben{
			Bereich: eintrag.Bereich,
			Grad:    eintrag.Grad,
			Quelle:  eintrag.Quelle,
			Grenzen: gefuelltOderLeer(eintrag.Grenzen),
		})
	}

	return angaben, nil
}
