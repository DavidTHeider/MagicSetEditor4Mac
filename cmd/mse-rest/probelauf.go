// Der Probelauf des Dienstes — UXR-F7
// (#2015).
//
// Er rechnet, was `POST /sets/{setID}/bogen` und `POST /sets/{setID}/vorlagen`
// täten, und schreibt nichts — also steht hier **keine Reservierung**. Nur ein
// genanntes Ziel wird angesehen: liegt dort schon etwas, antwortet der
// Probelauf wie der schreibende Weg
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// probelauf erfüllt `rest.Probelauf` über den Nähten von `internal/app`.
type probelauf struct{ db *sql.DB }

// ortFrei sieht nach, ob [reserviereOrt] das Ziel bekäme — und legt nichts an.
// Leer heißt: kein Ziel genannt.
func ortFrei(ziel string) error {
	if ziel == "" {
		return nil
	}

	_, err := os.Lstat(ziel)

	switch {
	case err == nil:
		return fmt.Errorf("%w: an %s liegt schon etwas — dieses "+
			"Werkzeug überschreibt nichts, wähle einen anderen Namen",
			domain.ErrConflict, ziel)

	case errors.Is(err, os.ErrNotExist):
		return nil

	default:
		return fmt.Errorf("%w: %s lässt sich nicht anlegen: %w",
			domain.ErrValidation, ziel, err)
	}
}

// BogenProbelauf rechnet den Bogen durch, den `Bogen` legte.
func (p probelauf) BogenProbelauf(
	ctx context.Context, setID int64, ziel string,
) (rest.Bogenprobelauf, error) {
	if err := ortFrei(ziel); err != nil {
		return rest.Bogenprobelauf{}, err
	}

	werk, err := app.Zeichennaht(p.db)
	if err != nil {
		return rest.Bogenprobelauf{}, err
	}

	probe, fehler := werk.ProbeBogen(ctx, setID)
	if fehler != nil {
		return rest.Bogenprobelauf{}, fehler
	}

	return rest.Bogenprobelauf{
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
) (rest.Vorlagenprobelauf, error) {
	if err := ortFrei(ziel); err != nil {
		return rest.Vorlagenprobelauf{}, err
	}

	weg, err := app.VorlagenwegUeber(p.db, domain.UrheberServer)
	if err != nil {
		return rest.Vorlagenprobelauf{}, err
	}

	probe, fehler := weg.Probelauf(ctx, setID, paket, ziel, wahl)
	if fehler != nil {
		return rest.Vorlagenprobelauf{}, fehler
	}

	return rest.Vorlagenprobelauf{
		Karten: probe.Karten, Dateiname: probe.Dateiname, Bytes: probe.Bytes,
		Beigaben: probe.Beigaben, BeigabenBytes: probe.BeigabenBytes,
		Verzeichnis: probe.Verzeichnis, Geschaetzt: probe.Geschaetzt,
		Befunde: probe.Befunde,
	}, nil
}
