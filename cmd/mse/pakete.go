// Die Befehle `pakete` und `paket-beziehen`
// (#1107).
//
// **Der Bezug ist der Grund, warum es sie auf der Kommandozeile geben muss.**
// Wer einen Rechner aus einem Skript einrichtet, kann nicht erst ein Fenster
// aufschlagen, um die Vorlagen zu holen, ohne die kein Set einen Stil hat.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// paketDTO ist ein installiertes Paket, wie die Kommandozeile es ausgibt.
type paketDTO struct {
	Art      string `json:"art"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Pfad     string `json:"pfad"`
	Herkunft string `json:"herkunft"`

	// Die Übersicht — DOM-01.AK-01 (#1696), dieselbe wie an den übrigen
	// Flächen (§16).
	Pruefsumme   string   `json:"pruefsumme"`
	Herkunftsart string   `json:"herkunftsart"`
	Geaendert    bool     `json:"geaendert"`
	Braucht      []string `json:"braucht"`
	GebrauchtVon []string `json:"gebrauchtVon"`
	Benutzt      int      `json:"benutzt"`
}

// paketeAntwort trägt den Katalog samt Befunden.
type paketeAntwort struct {
	Pakete  []paketDTO `json:"pakete"`
	Befunde []string   `json:"befunde"`
}

// bezugAntwort sagt, was ein Bezug ergab.
//
// **Die übergangenen Einträge reisen mit** (§8): ein Archiv, das einen Pfad
// mit `..` enthält, wird nicht ausgepackt — und wer das nicht sähe, hielte ein
// halbes Paket für ein ganzes.
type bezugAntwort struct {
	Ort         string   `json:"ort"`
	Stand       string   `json:"stand"`
	Pruefsumme  string   `json:"pruefsumme"`
	Uebergangen []string `json:"uebergangen"`
}

// paketebefehl frischt den Katalog auf und zeigt ihn.
func paketebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("pakete nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		pakete, befunde, lesefehler := weg.Liste(ctx)
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		return paketeZeigen(an, pakete, befunde)
	})
}

// paketeZeigen schreibt den Katalog.
func paketeZeigen(an *schreiber, pakete []app.Paketzeile, befunde []domain.Befund) int {
	antwort := paketeAntwort{
		Pakete:  make([]paketDTO, 0, len(pakete)),
		Befunde: make([]string, 0, len(befunde)),
	}

	for _, zeile := range pakete {
		antwort.Pakete = append(antwort.Pakete, alsZeilenDTO(zeile))
	}

	for _, befund := range befunde {
		antwort.Befunde = append(antwort.Befunde, befund.String())
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Pakete) == 0 {
		an.sag("Kein Paket installiert.\n")
	}

	for _, paket := range antwort.Pakete {
		an.sag("%s\t%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version, paket.Pfad)
	}

	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}

// paketbezugbefehl bezieht ein Vorlagenpaket.
//
// **Vier Argumente und keine Vorgabe.** Die Paketseite des Fensters füllt
// Adresse und Stand vor; hier wäre eine eingebaute Adresse eine Entscheidung,
// die niemand gesehen hat — und §11 sagt, dass ein Bezug nur auf ausdrückliche
// Anweisung geschieht.
func paketbezugbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 4 {
		an.klage(
			"paket-beziehen braucht Adresse, Stand, Unterverzeichnis und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		ergebnis, bezugsfehler := weg.Beziehen(ctx, rest[0], rest[1], rest[2], rest[3])
		if bezugsfehler != nil {
			an.klageUeber(bezugsfehler)

			return CodeFachlich
		}

		antwort := bezugAntwort{
			Ort:         ergebnis.Ort,
			Stand:       ergebnis.Stand,
			Pruefsumme:  ergebnis.Pruefsumme,
			Uebergangen: make([]string, 0, len(ergebnis.Uebergangen)),
		}
		antwort.Uebergangen = append(antwort.Uebergangen, ergebnis.Uebergangen...)

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("Bezogen nach %s (Stand %s, %s).\n",
			antwort.Ort, antwort.Stand, antwort.Pruefsumme)

		for _, uebergangen := range antwort.Uebergangen {
			an.sag("  Übergangen: %s\n", uebergangen)
		}

		return an.abschluss()
	})
}

// alsPaketDTO übersetzt ein Paket der Domäne.
//
// **An einer Stelle** (#1291): das Einbinden und die Übernahme geben dasselbe
// aus wie die Liste, und zwei Abbildungen liefen auseinander, sobald eine
// etwas lernt.
func alsPaketDTO(paket domain.Paket) paketDTO {
	return paketDTO{
		Art: paket.Art, Name: paket.Name, Version: paket.Version,
		Pfad: paket.Pfad, Herkunft: paket.Herkunft,
		Pruefsumme: paket.Pruefsumme, Geaendert: paket.Geaendert,
		Braucht: app.BrauchtZeilen(paket), GebrauchtVon: []string{},
	}
}

// alsZeilenDTO nimmt die Übersicht dazu — Herkunftsart, wer es braucht, wie
// viele Sets es benutzen (DOM-01.AK-01, #1696).
func alsZeilenDTO(zeile app.Paketzeile) paketDTO {
	gelesen := alsPaketDTO(zeile.Paket)
	gelesen.Herkunftsart = zeile.Herkunftsart
	gelesen.GebrauchtVon = append(gelesen.GebrauchtVon, zeile.GebrauchtVon...)
	gelesen.Benutzt = zeile.Benutzt

	return gelesen
}
