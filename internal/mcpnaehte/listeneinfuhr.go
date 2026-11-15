// **Die Listeneinfuhr des MCP-Servers** —
// #1292.
//
// **Ohne Wache** (0044): `Deuten` schreibt nichts, und `Uebernehmen` schreibt
// nur in den Bestand — es überschreibt keine Datei auf der Platte. Was es
// nicht tut, ist selbst entscheiden, welcher von 72 Drucken gemeint ist; das
// steht in `internal/mcp/listeneinfuhr.go`.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// listeneinfuhr erfüllt `mcp.Listeneinfuhr` über der Naht von `internal/app`.
type listeneinfuhr struct{ db *sql.DB }

func (l listeneinfuhr) Deuten(ctx context.Context, text string) (mcp.Listendeutung, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return mcp.Listendeutung{}, err
	}

	deutung, err := naht.Deuten(ctx, text)
	if err != nil {
		return mcp.Listendeutung{}, err
	}

	zeilen := make([]mcp.Listendeutungszeile, 0, len(deutung.Zeilen))

	for _, zeile := range deutung.Zeilen {
		treffer := make([]mcp.Listentreffer, 0, len(zeile.Treffer))
		for _, einer := range zeile.Treffer {
			treffer = append(treffer, mcp.Listentreffer{
				Printing: einer.Printing, Name: einer.Name,
				Setcode: einer.Setcode, Setname: einer.Setname,
				Nummer: einer.Nummer, Seltenheit: einer.Seltenheit,
			})
		}

		var fehler *mcp.Zeilenfehler
		if zeile.Fehler != nil {
			fehler = &mcp.Zeilenfehler{
				Code: string(zeile.Fehler.Code), Schluessel: string(zeile.Fehler.Schluessel), Meldung: zeile.Fehler.Meldung,
			}
		}

		zeilen = append(zeilen, mcp.Listendeutungszeile{
			Roh: zeile.Roh, Zeilennummer: zeile.Zeilennummer,
			Menge: zeile.Menge, Name: zeile.Name, Setcode: zeile.Setcode,
			Nummer: zeile.Nummer, Treffer: treffer,
			Foil:   zeile.Finish == string(domain.FinishFoil),
			Finish: zeile.Finish,
			Drucke: zeile.Drucke, Eindeutig: zeile.Eindeutig, Befund: zeile.Befund,
			Weitere: zeile.Weitere, Fehler: fehler,
		})
	}

	return mcp.Listendeutung{
		Zeilen: zeilen, Befunde: deutung.Befunde,
		AlterSekunden: deutung.AlterSekunden,
	}, nil
}

func (l listeneinfuhr) Einlesen(
	ctx context.Context, zeilen []mcp.Listeneinfuhrzeile, zusammenfuehrung string,
) (mcp.Einfuhrbericht, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return mcp.Einfuhrbericht{}, err
	}

	bericht, err := naht.Einlesen(ctx, alsNahtzeilen(zeilen), zusammenfuehrung)
	if bericht == nil {
		return mcp.Einfuhrbericht{Angelegt: []int64{}, Befunde: []string{}}, err
	}

	return alsMCPBericht(bericht), err
}

// Vorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
func (l listeneinfuhr) Vorschauen(
	ctx context.Context, zeilen []mcp.Listeneinfuhrzeile,
) (mcp.Sammelvorschau, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return mcp.Sammelvorschau{}, err
	}

	vorschau, err := naht.Vorschauen(ctx, alsNahtzeilen(zeilen))
	if err != nil {
		return mcp.Sammelvorschau{}, err
	}

	return alsMCPVorschau(vorschau), nil
}

// alsNahtzeilen übersetzt die Zeilen von MCP in die der Naht — **eine**
// Stelle für Einlesen und Vorschau, samt dem abgekündigten `foil`.
func alsNahtzeilen(zeilen []mcp.Listeneinfuhrzeile) []app.Listeneinfuhrzeile {
	uebersetzt := make([]app.Listeneinfuhrzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		uebersetzt = append(uebersetzt, app.Listeneinfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Zustand: zeile.Zustand, Sprache: zeile.Sprache,
			Finish: app.AusfuehrungAus(zeile.Finish, zeile.Foil),
			Name:   zeile.Name, Setcode: zeile.Setcode, Nummer: zeile.Nummer,
		})
	}

	return uebersetzt
}
