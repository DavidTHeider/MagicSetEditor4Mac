// **Der Probelauf des MCP-Servers** — UXR-F7
// (#2015).
//
// Er rechnet, was `schreiben.set.bogen` und `schreiben.vorlage.ausfuehren`
// täten, und schreibt nichts — also steht hier **keine Reservierung**. Nur
// ein genanntes Ziel wird angesehen: liegt dort schon etwas, antwortet der
// Probelauf wie der schreibende Aufruf ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// probelauf erfüllt `mcp.Probelauf` über den Nähten von `internal/app`.
type probelauf struct{ db *sql.DB }

// BogenProbelauf rechnet den Bogen durch, den `Bogen` legte.
func (p probelauf) BogenProbelauf(
	ctx context.Context, setID int64, ziel string,
) (mcp.Bogenprobelauf, error) {
	if err := ortFrei(ziel); err != nil {
		return mcp.Bogenprobelauf{}, err
	}

	werk, err := app.Zeichennaht(p.db)
	if err != nil {
		return mcp.Bogenprobelauf{}, err
	}

	probe, fehler := werk.ProbeBogen(ctx, setID)
	if fehler != nil {
		return mcp.Bogenprobelauf{}, fehler
	}

	return mcp.Bogenprobelauf{
		Karten: probe.Karten, Seiten: probe.Seiten, Spalten: probe.Spalten,
		Zeilen: probe.Zeilen, JeSeite: probe.JeSeite,
		AufLetzterSeite: probe.AufLetzterSeite, Bildbreite: probe.Bildbreite,
		Seitenbreite: probe.Seitenbreite, Seitenhoehe: probe.Seitenhoehe,
		Kartenbreite: probe.Kartenbreite, Kartenhoehe: probe.Kartenhoehe,
		Anschnitt: probe.Anschnitt, Rand: probe.Rand,
		DPI: probe.DPI, Schnittmarken: probe.Schnittmarken,
	}, nil
}

// VorlageProbelauf wertet die Vorlage aus, die `Ausfuehren` ausführte.
func (p probelauf) VorlageProbelauf(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) (mcp.Vorlagenprobelauf, error) {
	if err := ortFrei(ziel); err != nil {
		return mcp.Vorlagenprobelauf{}, err
	}

	weg, err := app.VorlagenwegUeber(p.db, domain.UrheberServer)
	if err != nil {
		return mcp.Vorlagenprobelauf{}, err
	}

	probe, fehler := weg.Probelauf(ctx, setID, paket, ziel, wahl)
	if fehler != nil {
		return mcp.Vorlagenprobelauf{}, fehler
	}

	return mcp.Vorlagenprobelauf{
		Karten: probe.Karten, Dateiname: probe.Dateiname, Bytes: probe.Bytes,
		Beigaben: probe.Beigaben, BeigabenBytes: probe.BeigabenBytes,
		Verzeichnis: probe.Verzeichnis, Geschaetzt: probe.Geschaetzt,
		Befunde: probe.Befunde,
	}, nil
}
