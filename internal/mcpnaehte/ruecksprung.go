package mcpnaehte

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// ruecksprungauskunft reicht die drei Auskünfte an die Werkzeuge.
type ruecksprungauskunft struct {
	db  *sql.DB
	ort string
}

func (r ruecksprungauskunft) Ruecksprungarten(ctx context.Context) (
	[]mcp.Ruecksprungart, error,
) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Ruecksprungarten(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	arten := make([]mcp.Ruecksprungart, 0, len(antwort.Arten))
	for _, art := range antwort.Arten {
		arten = append(arten, mcp.Ruecksprungart{Art: art.Art, Hindernis: art.Hindernis})
	}

	return arten, nil
}

func (r ruecksprungauskunft) Ruecksprunglage(ctx context.Context, pfad string) (
	mcp.Ruecksprunglage, error,
) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Ruecksprunglage(ctx, pfad, "")
	if antwort.Fehler != nil {
		return mcp.Ruecksprunglage{}, antwort.Fehler.AlsFehler()
	}

	lage := mcp.Ruecksprunglage{
		Datenstand:         antwort.Datenstand,
		HandpruefungNoetig: antwort.HandpruefungNoetig,
		Zeilen:             make([]mcp.Vergleichszeile, 0, len(antwort.Zeilen)),
		Unklar:             mengen(antwort.Unklar),
	}

	for _, zeile := range antwort.Zeilen {
		lage.Zeilen = append(lage.Zeilen, mcp.Vergleichszeile{
			Gegenstand: zeile.Gegenstand, InSicherung: zeile.InSicherung,
			Jetzt: zeile.Jetzt, Verlust: zeile.Verlust, Unbekannt: zeile.Unbekannt,
		})
	}

	return lage, nil
}

func (r ruecksprungauskunft) Rueckkehren(ctx context.Context) ([]mcp.Rueckkehr, error) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Rueckkehren(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	rueckkehren := make([]mcp.Rueckkehr, 0, len(antwort.Rueckkehren))
	for _, eine := range antwort.Rueckkehren {
		rueckkehren = append(rueckkehren, mcp.Rueckkehr{
			ID: eine.ID, Zurueckgespielt: eine.Zurueckgespielt,
			Datenstand: eine.Datenstand, Quelle: eine.Quelle,
			Klaerung: eine.Klaerung, GeklaertAm: eine.GeklaertAm,
			Mengen: mengen(eine.Mengen),
		})
	}

	return rueckkehren, nil
}

// mengen übersetzt die Mengen — **eine Stelle für beide Antworten**.
func mengen(roh []app.MengeDTO) []mcp.Menge {
	uebersetzt := make([]mcp.Menge, 0, len(roh))
	for _, menge := range roh {
		uebersetzt = append(uebersetzt, mcp.Menge{
			Schluessel: menge.Schluessel, Zustand: menge.Zustand,
			Sprache: menge.Sprache, Foil: menge.Foil, Finish: menge.Finish, Menge: menge.Menge,
			Name: menge.Name, Setcode: menge.Setcode, Nummer: menge.Nummer,
		})
	}

	return uebersetzt
}
