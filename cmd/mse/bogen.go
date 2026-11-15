// Der Befehl `bogen` — die Karten eines Sets als Druckbogen
// (#1291).
//
// # Warum es ihn gibt
//
// `docs/FAEHIGKEITEN.md` führte den Druckbogen mit drei Strichen und einem
// Haken: **nur am Fenster**. §16 verlangt, dass eine Fähigkeit an jeder Fläche
// steht — oder dass dort begründet steht, warum nicht. Für einen Druckbogen
// stand dort nichts.
//
// # Warum er ohne Fenster geht
//
// Weil der Dateidialog nur in der **Bindung** steckt: `Bindungen.BogenLegen`
// fragt nach einem Ort und ruft dann `App.BogenLegen`, das einen Pfad
// entgegennimmt. Der Weg dahinter kennt kein Fenster.

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bogenantwort sagt, was geschrieben wurde.
type bogenantwort struct {
	Pfad   string `json:"pfad"`
	Seiten int    `json:"seiten"`
	Karten int    `json:"karten"`

	// Befunde ist, was beim Zeichnen aufgefallen ist — **nicht `omitempty`**:
	// eine leere Liste ist die Aussage „nichts ist aufgefallen".
	Befunde []string `json:"befunde"`
}

// bogenLegen schreibt die Karten eines Sets als PDF.
func bogenLegen(ctx context.Context, db *sql.DB, an *schreiber, setID int64, ziel string) int {
	werk, err := app.Zeichennaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	// **Dieselben Maße wie am Fenster.** A4 mit neun Karten in
	// Druckauflösung; das Kartenmaß ist die Zusage aus #50 und keine
	// Einstellung.
	lauf, fehler := werk.LegeBogen(ctx, setID, ziel)
	if fehler != nil {
		an.klageUeber(fehler)

		// **Ein unbrauchbares Ziel ist ein Bedienfehler**, ein fehlendes Set
		// ein fachliches Nein — ein Skript soll das unterscheiden können.
		if errors.Is(fehler, domain.ErrValidation) {
			return CodeBedienung
		}

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bogenantwort{
			Pfad: lauf.Pfad, Seiten: lauf.Seiten, Karten: lauf.Karten, Befunde: lauf.Befunde,
		})
	}

	an.sag("Geschrieben nach %s — %d Seiten, %d Karten.\n",
		lauf.Pfad, lauf.Seiten, lauf.Karten)

	// **Und was dabei aufgefallen ist**, je Zeile eines — wie beim Zeichnen.
	for _, befund := range lauf.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}
