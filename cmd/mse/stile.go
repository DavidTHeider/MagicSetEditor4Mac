// Die Befehle `stile` und `stiloption`
// (#1107).
//
// **Der Wechsel bleibt zweistufig.** `stile <set> folgen <paket>` sagt, was er
// kostet; `stile <set> wechseln <paket>` tut ihn. Im Fenster steht dazwischen
// eine Rückfrage, und eine Kommandozeile ohne Rückfrage muss wenigstens
// **fragen** können (§16, Regel 6).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// stileAntwort trägt die wählbaren Stile.
type stileAntwort struct {
	Stile []app.StilDTO `json:"stile"`
}

// stiloptionenAntwort trägt die Optionen des gewählten Stils.
type stiloptionenAntwort struct {
	Optionen []app.StiloptionDTO `json:"optionen"`
}

// stilebefehl liest, prüft oder wechselt den Stil eines Sets.
func stilebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("stile braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	weitere := rest[1:]
	if len(weitere) != 0 && len(weitere) != 2 {
		an.klage("stile nimmt entweder nichts oder »folgen|wechseln <paket>«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.StilnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if len(weitere) == 0 {
			return stileZeigen(ctx, an, weg, setID)
		}

		return stilbewegung(ctx, an, weg, setID, weitere[0], weitere[1])
	})
}

// stileZeigen schreibt die wählbaren Stile.
func stileZeigen(ctx context.Context, an *schreiber, weg *app.Stilweg, setID int64) int {
	stile, fehler := weg.Stile(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, stileAntwort{Stile: stile})
	}

	if len(stile) == 0 {
		an.sag("Kein Stil verfügbar — fehlt ein Paket?\n")

		return an.abschluss()
	}

	for _, stil := range stile {
		marke := " "
		if stil.Gewaehlt {
			marke = "*"
		}

		an.sag("%s %s\t%s\t%s\n", marke, stil.Paket, stil.Spiel, stil.Version)
	}

	return an.abschluss()
}

// stilbewegung prüft die Folgen oder wechselt.
func stilbewegung(
	ctx context.Context, an *schreiber, weg *app.Stilweg, setID int64, was, paket string,
) int {
	switch was {
	case "folgen":
		folgen, fehler := weg.Folgen(ctx, setID, paket)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return folgenZeigen(an, folgen)

	case "wechseln":
		return getan(an, weg.Wechseln(ctx, setID, paket), "Gewechselt.\n")

	default:
		an.klage("unbekannter Unterbefehl %q von stile.\n%s", was, hilfe)

		return CodeBedienung
	}
}

// folgenZeigen schreibt, was ein Wechsel kostet.
func folgenZeigen(an *schreiber, folgen *app.FolgenDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, folgen)
	}

	if folgen.Spielwechsel {
		an.sag("Achtung: das Spiel wechselt mit.\n")
	}

	for _, feld := range folgen.UngedeuteteFelder {
		an.sag("Feld ohne Entsprechung: %s\n", feld)
	}

	for _, option := range folgen.VerloreneOptionen {
		an.sag("Stiloption geht verloren: %s\n", option)
	}

	if !folgen.Spielwechsel &&
		len(folgen.UngedeuteteFelder) == 0 && len(folgen.VerloreneOptionen) == 0 {
		an.sag("Der Wechsel kostet nichts.\n")
	}

	return an.abschluss()
}

// stiloptionbefehl liest, setzt oder entfernt eine Stiloption.
func stiloptionbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("stiloption braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	weitere := rest[1:]
	if !stilargumenteStimmen(weitere) {
		an.klage(
			"stiloption nimmt nichts, »setzen <feld> <wert>« oder »zuruecksetzen <feld>«.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.StilnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return stiloptionAusfuehren(ctx, an, weg, setID, weitere)
	})
}

// stilargumenteStimmen prüft die drei erlaubten Gestalten.
func stilargumenteStimmen(weitere []string) bool {
	switch len(weitere) {
	case 0:
		return true
	case 2:
		return weitere[0] == "zuruecksetzen"
	case 3:
		return weitere[0] == "setzen"
	default:
		return false
	}
}

// stiloptionAusfuehren wählt zwischen Lesen, Setzen und Zurücksetzen.
func stiloptionAusfuehren(
	ctx context.Context, an *schreiber, weg *app.Stilweg, setID int64, weitere []string,
) int {
	switch len(weitere) {
	case 3:
		return getan(an, weg.OptionSetzen(ctx, setID, weitere[1], weitere[2]), "Gesetzt.\n")

	case 2:
		return getan(an, weg.OptionZuruecksetzen(ctx, setID, weitere[1]),
			"Auf die Vorgabe zurückgesetzt.\n")

	default:
		return stiloptionenZeigen(ctx, an, weg, setID)
	}
}

// stiloptionenZeigen schreibt die Optionen samt Werten.
func stiloptionenZeigen(ctx context.Context, an *schreiber, weg *app.Stilweg, setID int64) int {
	optionen, fehler := weg.Optionen(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, stiloptionenAntwort{Optionen: optionen})
	}

	if len(optionen) == 0 {
		an.sag("Der gewählte Stil hat keine Optionen.\n")

		return an.abschluss()
	}

	for _, option := range optionen {
		// **Gesetzt und Vorgabe sind zweierlei** — ein Stern sagt, dass der
		// Wert aus dem Bestand kommt und nicht aus `default:`.
		marke := " "
		if option.Gesetzt {
			marke = "*"
		}

		an.sag("%s %s\t%s\t%s\n", marke, option.Name, option.Art, option.Wert)
	}

	return an.abschluss()
}
