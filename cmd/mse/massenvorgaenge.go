// **Massenvorgänge und Preise in Masse, an der Kommandozeile** —
// Werkstück 6 von Phase 4.5
// (#1292).
//
//	mse masse-aendern lagerort="Ordner Rot" 12:1 13:1 14:2
//	mse masse-loeschen 12:1 13:1
//	mse preise-alle [stunden]
//
// **Jede Zeile trägt ihre Fassung** — `<id>:<fassung>` (§7, Regel 2). Ohne sie
// überschriebe ein Massenvorgang jede Änderung, die seit dem Lesen der Liste
// geschehen ist, und zwar fünfzigfach auf einmal.

package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// masseAendernbefehl ändert die genannten Zeilen.
func masseAendernbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	aenderung, zeilen, gut := massenargumente(an, rest)
	if !gut {
		return CodeBedienung
	}

	aenderung.Zeilen = zeilen

	return mitMasse(ctx, bestandsort, an, func(ctx context.Context, naht *app.Massennaht) int {
		bericht, err := naht.Aendern(ctx, aenderung)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return massenbericht(an, bericht)
	})
}

// masseLoeschenbefehl nimmt die genannten Zeilen heraus.
func masseLoeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	zeilen, gut := massenzeilenAus(an, rest)
	if !gut {
		return CodeBedienung
	}

	return mitMasse(ctx, bestandsort, an, func(ctx context.Context, naht *app.Massennaht) int {
		bericht, err := naht.Loeschen(ctx, zeilen)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return massenbericht(an, bericht)
	})
}

// preiseAllebefehl holt einen Schub Preise.
func preiseAllebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	stunden := 24

	if len(rest) > 0 {
		gelesen, fehler := strconv.Atoi(rest[0])
		if fehler != nil {
			an.klage("»%s« ist keine Stundenzahl.\n%s", rest[0], hilfe)

			return CodeBedienung
		}

		stunden = gelesen
	}

	return mitMasse(ctx, bestandsort, an, func(ctx context.Context, naht *app.Massennaht) int {
		schub, err := naht.PreiseInMasse(ctx, stunden)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, schub)
		}

		an.sag("%d geholt, %d offen\n", schub.Geholt, schub.Offen)

		// **Nur wenn etwas schiefging** — sonst wäre die Zahl bei jedem Lauf
		// da und verlöre ihre Bedeutung. Steht der Schub, sagt er es: ein
		// weiterer Ruf gäbe dasselbe (Nachlauf über Phase 4.5, Befund 7).
		if schub.Fehlgeschlagen > 0 {
			an.sag("%d fehlgeschlagen\n", schub.Fehlgeschlagen)
		}

		if schub.Steht {
			an.sag("dieser Schub hat nichts bewegt — " +
				"ein weiterer Ruf gäbe dasselbe\n")
		}

		for _, befund := range schub.Befunde {
			an.sag("!\t%s\n", befund)
		}

		return an.abschluss()
	})
}

// massenargumente trennt die Änderungen von den Zeilen.
//
// **Ein `feld=wert` ist eine Änderung, ein `id:fassung` eine Zeile.** Der
// Doppelpunkt trennt sie eindeutig; ein Lagerort mit Doppelpunkt bräuchte
// Anführungszeichen, und die setzt die Schale.
func massenargumente(an *schreiber, rest []string) (
	app.MassenaenderungDTO, []app.MassenzeileDTO, bool,
) {
	aenderung := app.MassenaenderungDTO{}
	zeilen := []string{}

	for _, argument := range rest {
		name, wert, istAenderung := strings.Cut(argument, "=")
		if !istAenderung {
			zeilen = append(zeilen, argument)

			continue
		}

		if !setzeAenderung(an, &aenderung, name, wert) {
			return aenderung, nil, false
		}
	}

	gelesen, gut := massenzeilenAus(an, zeilen)

	return aenderung, gelesen, gut
}

// setzeAenderung trägt ein `feld=wert` ein.
func setzeAenderung(
	an *schreiber, aenderung *app.MassenaenderungDTO, name, wert string,
) bool {
	switch name {
	case "menge":
		menge, fehler := strconv.Atoi(wert)
		if fehler != nil {
			an.klage("»%s« ist keine Menge.\n%s", wert, hilfe)

			return false
		}

		aenderung.Menge = &menge

	case "zustand":
		aenderung.Zustand = &wert

	case "lagerort":
		aenderung.Lagerort = &wert

	case "wunsch":
		// **`wunsch=ja` und `wunsch=nein`** — und nichts sonst: ein
		// stillschweigendes „alles außer ja heißt nein" machte aus einem
		// Tippfehler eine Änderung.
		if wert != "ja" && wert != "nein" {
			an.klage("»wunsch« nimmt ja oder nein, nicht »%s«.\n%s", wert, hilfe)

			return false
		}

		gesetzt := wert == "ja"
		aenderung.Wunsch = &gesetzt

	default:
		an.klage("»%s« ist kein Feld für einen Massenvorgang.\n%s", name, hilfe)

		return false
	}

	return true
}

// massenzeilenAus liest `<id>:<fassung>`.
func massenzeilenAus(an *schreiber, roh []string) ([]app.MassenzeileDTO, bool) {
	if len(roh) == 0 {
		an.klage("Ein Massenvorgang braucht Zeilen als <id>:<fassung>.\n%s", hilfe)

		return nil, false
	}

	zeilen := make([]app.MassenzeileDTO, 0, len(roh))

	for _, eintrag := range roh {
		kennung, fassung, getrennt := strings.Cut(eintrag, ":")
		if !getrennt {
			an.klage("»%s« ist keine Zeile — erwartet wird <id>:<fassung>.\n%s",
				eintrag, hilfe)

			return nil, false
		}

		id, kennungsfehler := strconv.ParseInt(kennung, 10, 64)
		stand, fassungsfehler := strconv.ParseInt(fassung, 10, 64)

		if kennungsfehler != nil || fassungsfehler != nil {
			an.klage("»%s« ist keine Zeile — erwartet wird <id>:<fassung>.\n%s",
				eintrag, hilfe)

			return nil, false
		}

		zeilen = append(zeilen, app.MassenzeileDTO{ID: id, Fassung: stand})
	}

	return zeilen, true
}

// massenbericht schreibt das Ergebnis.
func massenbericht(an *schreiber, bericht *app.MassenberichtDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, bericht)
	}

	an.sag("%d geändert\n", bericht.Geaendert)

	for _, befund := range bericht.Befunde {
		an.sag("!\t%s\n", befund)
	}

	return an.abschluss()
}

// mitMasse baut die Naht über dem Bestand.
func mitMasse(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Massennaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return arbeit(ctx, app.MassennahtUeber(db))
	})
}
