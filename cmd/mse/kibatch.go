package main

// **Ein Batch von Vorschlägen auf der Kommandozeile** — Werkstück A13.1 von
// Phase 5.3 (#1452).
//
// # Was hier geht und was nicht
//
// **Ein Batch an ein Modell auf diesem Rechner geht.** Einer an einen
// Cloudanbieter nicht: er verlangt die Freigabe eines Menschen, und die wird
// **am Fenster** ausgestellt — `Nachweise.Ausstellen` weist jeden anderen
// Modus ab ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).
// Dasselbe gilt für den einzelnen Vorschlag; hier steht es, damit niemand
// den Weg sucht.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kibatchbefehl fährt einen Batch über mehrere Karten.
func kibatchbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	const hilfe = "Aufruf: mse ki-batch <karte,karte,…> <anbieter> <modell> <auftrag>"

	if len(rest) != 4 {
		an.klage("ki-batch braucht die Karten, den Anbieter, das Modell "+
			"und den Auftrag.\n%s", hilfe)

		return CodeBedienung
	}

	karten, fehler := kartenliste(rest[0])
	if fehler != nil {
		an.klage("%v\n%s", fehler, hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIBatchVorschlagen(
			ctx, karten, rest[1], rest[2], rest[3], "")
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return batchzeilen(an, antwort)
	})
}

// batchzeilen schreibt die Lagen als Zeilen.
//
// **Je Ziel eine Zeile mit seinem Ausgang** — ein Batch, der nur „fertig"
// sagt, verschweigt, welche Karte nichts bekommen hat.
func batchzeilen(an *schreiber, antwort app.KIBatchAntwort) int {
	an.sag("batch\t%s\n", antwort.Kennung)

	gescheitert := 0

	for _, lage := range antwort.Ziele {
		if lage.Fehler != nil {
			gescheitert++

			an.sag("%d\tgescheitert\t%s\n", lage.Zielkarte, lage.Fehler.Meldung)

			continue
		}

		an.sag("%d\tvorgeschlagen\t%s\t%s\n", lage.Zielkarte, lage.Kennung,
			strings.Join(batchfeldnamen(lage.Felder), ", "))
	}

	// **Der Rückgabewert trägt dieselbe Aussage** (§16): ein Batch, in dem
	// ein Ziel scheiterte, ist kein voller Erfolg — und ein Skript liest den
	// Wert und nicht die Zeilen.
	if gescheitert > 0 {
		return CodeFachlich
	}

	return CodeGut
}

// feldnamen nennt die vorgeschlagenen Felder.
func batchfeldnamen(felder []app.KIFeldwertDTO) []string {
	namen := make([]string, 0, len(felder))
	for _, feld := range felder {
		namen = append(namen, feld.Feld)
	}

	return namen
}

// kartenliste zerteilt `4,7,12` in Kennungen.
//
// **Mit Komma und ohne Leerzeichen** — dieselbe Form wie `--felder=` beim
// einzelnen Vorschlag; zwei Trennzeichen für dieselbe Sache wären zwei
// Gewohnheiten.
func kartenliste(roh string) ([]int64, error) {
	teile := strings.Split(roh, ",")
	karten := make([]int64, 0, len(teile))

	for _, teil := range teile {
		karte, fehler := strconv.ParseInt(strings.TrimSpace(teil), 10, 64)
		if fehler != nil {
			return nil, &bedienfehler{
				satz: "die Kartenkennung " + teil + " ist keine Zahl",
			}
		}

		karten = append(karten, karte)
	}

	return karten, nil
}

// bedienfehler ist ein Satz für die Bedienung und kein Fachfehler.
type bedienfehler struct{ satz string }

func (b *bedienfehler) Error() string { return b.satz }
