// Der Befehl `symbol-nachzeichnen` — ein Bild als Setsymbol
// (#1340).
//
// # Warum er ohne Fenster geht
//
// Die übrigen Symbolfähigkeiten stehen mit Grund allein am Fenster: eine
// Grundform, ein Punktzug, eine Symmetrie sind Handgriffe auf einer
// Zeichenfläche.
//
// **Nachzeichnen ist keiner.** Bild hinein, Zeichnung heraus — und wer hundert
// Symbole aus PNG-Dateien machen will, tut das hier und nicht mit hundert
// Klicks (§16). Der Dateidialog steckt allein in `Bindungen.SymbolNachzeichnen`.

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nachzeichnungsantwort sagt, was entstanden ist.
type nachzeichnungsantwort struct {
	Name   string `json:"name"`
	Teile  int    `json:"teile"`
	Punkte int    `json:"punkte"`
}

// symbolnachzeichnenbefehl zeichnet ein Bild als Symbol eines Sets nach.
func symbolnachzeichnenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("symbol-nachzeichnen braucht die Setkennung und eine Bilddatei.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	// **Mit dem Bestand** (#1815): wem das Set gehört, weiß nur er.
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SymbolnachzeichnungUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		roh, err := naht.Nachzeichnen(ctx, setID, rest[1])
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return nachzeichnungMelden(an, roh)
	})
}

// nachzeichnungMelden gibt aus, was entstanden ist.
func nachzeichnungMelden(an *schreiber, roh any) int {
	// **Die Naht gibt `any`**, weil MCP und Dienst sie als JSON
	// weiterreichen; hier wird daraus wieder eine Struktur. Der Umweg ist der
	// Preis dafür, dass alle drei kopflosen Flächen **dieselbe** Naht rufen —
	// und der ist billiger als drei Fassungen derselben Kette.
	gemeldet, gemacht := roh.(map[string]any)
	if !gemacht {
		an.klageUeber(errors.New("die Naht gab eine unerwartete Antwort"))

		return CodeFachlich
	}

	// **Der Nullwert ist die Antwort auf „nicht da".** Eine Naht, die `name`
	// nicht setzt, hat nichts abgelegt — und `""` sagt das. Ein Fehler dafür
	// wäre eine zweite Meldung über denselben Umstand.
	name, _ := gemeldet["name"].(string)  //nolint:errcheck // der Nullwert ist die Aussage
	teile, _ := gemeldet["teile"].(int)   //nolint:errcheck // dito
	punkte, _ := gemeldet["punkte"].(int) //nolint:errcheck // dito

	if an.alsJSON {
		return alsJSONSchreiben(an, nachzeichnungsantwort{
			Name: name, Teile: teile, Punkte: punkte,
		})
	}

	an.sag("Nachgezeichnet nach %s — %d Formen, %d Punkte.\n", name, teile, punkte)

	return an.abschluss()
}
