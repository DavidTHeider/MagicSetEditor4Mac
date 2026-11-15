// **Die Liste der Handelsvorgänge und ihre Entwürfe an der Kommandozeile** —
// TRADE-01 (#1715).
//
// `mse handel liste [--lage=…] [--suche=…] …` fragt ohne Vorgang; die
// Entwurfsbefehle stehen dort, wo jeder andere Handelsbefehl steht — hinter
// `handel`, vor dem Vorgang.
package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// zeilensicht ist ein Vorgang der Liste, wie die Kommandozeile ihn schreibt.
type zeilensicht struct {
	Vorgang      vorgangssicht           `json:"vorgang"`
	Lage         string                  `json:"lage"`
	Naechster    string                  `json:"naechster"`
	Gegenueber   []app.HandelspartnerDTO `json:"gegenueber"`
	Karten       []string                `json:"karten"`
	Marken       []string                `json:"marken"`
	Angelegt     string                  `json:"angelegt"`
	Geaendert    string                  `json:"geaendert"`
	Archiviert   bool                    `json:"archiviert"`
	Unbestaetigt bool                    `json:"unbestaetigt"`
}

// istVerzeichnisbefehl sagt, ob ein Handelsbefehl Liste oder Entwurf meint.
func istVerzeichnisbefehl(befehl string) bool {
	switch befehl {
	case "entwurf", "kopieren", "verwerfen", "archivieren", "hervorholen":
		return true
	}

	return false
}

// handelslistebefehl fährt `mse handel liste [--…]`.
func handelslistebefehl(ctx context.Context, bestandsort string, an *schreiber, argumente []string) int {
	filter, gut := handelsfilterLesen(an, argumente)
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).HandelUebersicht(ctx, filter)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		sichten := make([]zeilensicht, 0, len(antwort.Zeilen))
		for _, zeile := range antwort.Zeilen {
			sichten = append(sichten, zeilensicht{
				Vorgang: sichtAus(zeile.Vorgang), Lage: zeile.Lage, Naechster: zeile.Naechster,
				Gegenueber: zeile.Gegenueber, Karten: zeile.Karten, Marken: zeile.Marken,
				Angelegt: zeile.Angelegt, Geaendert: zeile.Geaendert,
				Archiviert: zeile.Archiviert, Unbestaetigt: zeile.Unbestaetigt,
			})
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"zeilen": sichten})
		}

		for _, sicht := range sichten {
			an.sag("%s\t%s\t%s\t%s\t%s\t%s\n", sicht.Vorgang.Vorgang, sicht.Lage, sicht.Naechster,
				partnertext(sicht.Gegenueber), sicht.Vorgang.Zustand, sicht.Geaendert)
		}

		return an.abschluss()
	})
}

// partnertext nennt die Gegenüber — mit Namen, wo es einen gibt.
func partnertext(partner []app.HandelspartnerDTO) string {
	namen := make([]string, 0, len(partner))
	for _, gegenueber := range partner {
		namen = append(namen, gegenueber.Kennung)
		if gegenueber.Name != "" {
			namen[len(namen)-1] = gegenueber.Name
		}
	}

	return strings.Join(namen, ", ")
}

// handelsfilterLesen liest `--angabe=wert` — `--lage` darf mehrfach stehen
// oder mehrere durch Komma nennen.
func handelsfilterLesen(an *schreiber, argumente []string) (app.HandelsfilterDTO, bool) {
	var filter app.HandelsfilterDTO

	felder := map[string]*string{
		"--partner": &filter.Partner, "--suche": &filter.Suche, "--marke": &filter.Marke,
		"--seit": &filter.Seit, "--bis": &filter.Bis, "--ordnung": &filter.Ordnung,
	}

	for _, argument := range argumente {
		schluessel, wert, mitWert := strings.Cut(argument, "=")

		switch ziel, bekannt := felder[schluessel]; {
		case argument == "--archiv":
			filter.Archiv = true
		case schluessel == "--lage" && mitWert:
			filter.Lagen = append(filter.Lagen, strings.Split(wert, ",")...)
		case bekannt && mitWert:
			*ziel = wert
		default:
			an.klage("handel liste versteht %q nicht — erwartet sind --partner=, --lage=, "+
				"--suche=, --marke=, --seit=, --bis=, --ordnung= und --archiv.\n", argument)

			return app.HandelsfilterDTO{}, false
		}
	}

	return filter, true
}

// verzeichnisbefehl fährt die Entwurfs- und Archivbefehle an einem Vorgang.
func verzeichnisbefehl(ctx context.Context, flaeche *app.App, an *schreiber, befehl, vorgang string, rest []string) int {
	if befehl == "entwurf" {
		if len(rest) != 2 {
			an.klage("handel entwurf nimmt die Art und das Gegenüber.\n%s", hilfe)

			return CodeBedienung
		}

		positionen, gelesen := positionenLesen(an)
		if !gelesen {
			return CodeBedienung
		}

		return vorgangSchreiben(an, flaeche.HandelEntwurfAendern(ctx, app.Handelswunsch{
			Vorgang: vorgang, Art: rest[0], Gegenueber: []string{rest[1]}, Positionen: positionen,
		}))
	}

	if len(rest) > 0 {
		an.klage("handel %s nimmt nur den Vorgang.\n%s", befehl, hilfe)

		return CodeBedienung
	}

	switch befehl {
	case "kopieren":
		return vorgangSchreiben(an, flaeche.HandelKopieren(ctx, vorgang))
	case "verwerfen":
		return erledigt(an, flaeche.HandelEntwurfLoeschen(ctx, vorgang), "verworfen")
	case "archivieren":
		return erledigt(an, flaeche.HandelArchivieren(ctx, vorgang, true), "archiviert")
	}

	return erledigt(an, flaeche.HandelArchivieren(ctx, vorgang, false), "hervorgeholt")
}

// erledigt sagt in einem Wort, was geschah — oder, warum nicht.
func erledigt(an *schreiber, antwort app.SchreibAntwort, wort string) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	an.sag("%s\n", wort)

	return an.abschluss()
}
