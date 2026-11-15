package main

// **Der Ausgang an der Kommandozeile** — die Aufrufer für
// `service.Handelspost.Senden` und `Wiederaufnehmen` aus `T5`
// (#1470),
// nachgetragen mit
// #1637.
//
// # Drei Befehle neben `handelspost`
//
// **Weil vier Fragen an denselben Ausgang gehen**, und nur eine davon ist
// eine Auskunft: `handelspost` zählt, `handelspost-schicken` legt hinein,
// `handelspost-bestaetigen` räumt weg, `handelspost-wiederaufnahme` sagt, was
// nach einem Abriss fehlt.
//
// # Und warum „schicken" hier nichts über die Leitung trägt
//
// **Weil „abgeschickt" nicht „angekommen" heißt** (`T5`). Was hier
// hineingeht, bleibt liegen, bis die Gegenseite es bestätigt; dass der
// Ausgang von selbst leerläuft, verlangt eine Fläche für die Leitung, und die
// steht als
// #1634.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// versandsicht ist die Nachricht, die nun im Ausgang liegt.
type versandsicht struct {
	NachrichtID string `json:"nachrichtId"`
}

// nachlieferungssicht ist eine Nachricht, die noch einmal hinausmuss.
type nachlieferungssicht struct {
	NachrichtID string `json:"nachrichtId"`
	Vorgang     string `json:"vorgang"`
	Empfaenger  string `json:"empfaenger"`
	Folge       int64  `json:"folge"`
	Art         string `json:"art"`
}

// wiederaufnahmesicht sagt, was nach einem Abriss fehlt.
type wiederaufnahmesicht struct {
	Nachzuliefern []nachlieferungssicht `json:"nachzuliefern"`
	Klaerung      bool                  `json:"klaerung"`
}

// handelsschickenbefehl fährt `mse handelspost-schicken <vorgang> <kennung>`.
func handelsschickenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("handelspost-schicken nimmt den Vorgang und das Gegenüber.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return versandSchreiben(an,
				app.PeerflaecheUeber(db).HandelSchicken(ctx, rest[0], rest[1]))
		})
}

// handelsbestaetigtbefehl fährt `mse handelspost-bestaetigen <nachrichtID>`.
func handelsbestaetigtbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("handelspost-bestaetigen nimmt die Kennung der Nachricht.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return versandSchreiben(an,
				app.PeerflaecheUeber(db).HandelPostbestaetigung(ctx, rest[0]))
		})
}

// handelswiederaufnahmebefehl fährt
// `mse handelspost-wiederaufnahme <vorgang> <kennung> <revision> <abdruck> <folge>`.
//
// **Der Fortschritt der Gegenseite steht in den Argumenten** — er kommt von
// drüben und ist eine Behauptung, die gegen den eigenen Stand gehalten wird.
// Der eigene wird gelesen und nicht genannt.
func handelswiederaufnahmebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 5 {
		an.klage("handelspost-wiederaufnahme nimmt Vorgang, Gegenüber, "+
			"Revision, Abdruck und Folge.\n%s", hilfe)

		return CodeBedienung
	}

	revision, fehler := strconv.Atoi(rest[2])
	if fehler != nil {
		an.klage("die Revision %q ist keine Zahl.\n%s", rest[2], hilfe)

		return CodeBedienung
	}

	folge, fehler := strconv.ParseInt(rest[4], 10, 64)
	if fehler != nil {
		an.klage("die Folge %q ist keine Zahl.\n%s", rest[4], hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return wiederaufnahmeSchreiben(an,
				app.PeerflaecheUeber(db).HandelWiederaufnahme(
					ctx, rest[0], rest[1], app.HandelsfortschrittDTO{
						Vorgang: rest[0], Revision: revision,
						Abdruck: rest[3], Folge: folge,
					}))
		})
}

// versandSchreiben gibt die Kennung der Nachricht aus.
func versandSchreiben(an *schreiber, antwort app.VersandAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an,
			versandsicht{NachrichtID: antwort.NachrichtID})
	}

	an.sag("%s\n", antwort.NachrichtID)

	return an.abschluss()
}

// wiederaufnahmeSchreiben gibt aus, was nachzuliefern ist.
func wiederaufnahmeSchreiben(
	an *schreiber, antwort app.WiederaufnahmeAntwort,
) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	sicht := wiederaufnahmesicht{
		Nachzuliefern: make([]nachlieferungssicht, 0,
			len(antwort.Nachzuliefern)),
		Klaerung: antwort.Klaerung,
	}

	for _, zeile := range antwort.Nachzuliefern {
		sicht.Nachzuliefern = append(sicht.Nachzuliefern, nachlieferungssicht{
			NachrichtID: zeile.NachrichtID, Vorgang: zeile.Vorgang,
			Empfaenger: zeile.Empfaenger, Folge: zeile.Folge, Art: zeile.Art,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sicht)
	}

	// **Die Klärung zuerst** — sie ist die Auskunft, die den Rest wertlos
	// macht: stehen zwei Verläufe da, liefert niemand nach.
	an.sag("klaerung\t%v\n", sicht.Klaerung)

	for _, zeile := range sicht.Nachzuliefern {
		an.sag("%s\t%d\t%s\t%s\n",
			zeile.NachrichtID, zeile.Folge, zeile.Art, zeile.Empfaenger)
	}

	return an.abschluss()
}
