// **Die Sicherung des ganzen Bestands am Dienst** —
// #1169.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// bestandssicherung erfüllt `rest.Bestandssicherung` über der Naht von `app`.
//
// **Der Pfad steht daneben**, weil die Naht ihn braucht: eine Sicherung
// beschreibt den Bestand, und wo der liegt, weiß nur der Composition Root.
type bestandssicherung struct {
	db  *sql.DB
	ort string
}

func (b bestandssicherung) BestandSichern(ctx context.Context, ziel, kennwort string, klartext bool) error {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Sichern(app.MitSicherungswunsch(ctx, domain.FlaecheREST, kennwort, klartext), ziel)
}

func (b bestandssicherung) SicherungSchauen(ctx context.Context, pfad, kennwort string) (
	rest.Sicherungsangaben, error,
) {
	db := b.db

	schau, fehler := app.BestandsnahtUeber(&db, b.ort).Beschauen(app.MitSicherungswunsch(ctx, domain.FlaecheREST, kennwort, false), pfad)
	if fehler != nil {
		return rest.Sicherungsangaben{}, fehler
	}

	return rest.Sicherungsangaben{
		AngelegtAm:  schau.AngelegtAm,
		Schemastand: schau.Schemastand,
		Sets:        schau.Sets,
		Karten:      schau.Karten,
		Bilder:      schau.Bilder,

		Verschluesselt: schau.Verschluesselt,

		PaketeEnthalten: schau.PaketeEnthalten,
		PaketeFehlen:    schau.PaketeFehlen,
	}, nil
}

// Sicherungsstand nennt das Alter der letzten Sicherung.
func (b bestandssicherung) Sicherungsstand(ctx context.Context) (any, error) {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Sicherungsstand(ctx)
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt daliegt.
func (b bestandssicherung) Verschluesselungsstand() (bool, error) {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Verschluesselungsstand()
}
