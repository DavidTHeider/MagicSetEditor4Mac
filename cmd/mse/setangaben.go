// Die Befehle `setangaben`, `vorbehalte` und `beispielsets`
// (#1107).
//
// **Drei Fähigkeiten des Fensters, die kein Fenster brauchen.** Ein Set hat
// Titel, Setcode und Copyright; wer hundert Sets aus einem Skript anlegt, will
// sie setzen können, ohne hundertmal eine Lasche aufzuschlagen.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// angabenAntwort trägt Felder und Werte eines Sets.
type angabenAntwort struct {
	Felder []app.FeldbeschreibungDTO `json:"felder"`
	Werte  []app.WertDTO             `json:"werte"`
}

// vorbehalteAntwort trägt, was beim Einfahren nicht gedeutet wurde.
type vorbehalteAntwort struct {
	Vorbehalte []app.VorbehaltDTO `json:"vorbehalte"`
}

// beispielsetsAntwort trägt die mitgelieferten Sets.
type beispielsetsAntwort struct {
	Sets []app.BeispielsetDTO `json:"sets"`
}

// setangabenbefehl liest oder setzt die Angaben eines Sets.
func setangabenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("setangaben braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	weitere := rest[1:]

	if len(weitere) != 0 && (weitere[0] != "setzen" || len(weitere) != 3) {
		an.klage("setangaben nimmt entweder nichts oder »setzen <feld> <wert>«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.SetangabennahtUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if len(weitere) == 3 {
			return setwertSchreiben(ctx, an, weg, setID, weitere[1], weitere[2])
		}

		return angabenZeigen(ctx, an, weg, setID)
	})
}

// angabenZeigen schreibt Felder und Werte.
func angabenZeigen(
	ctx context.Context, an *schreiber, weg *app.Setangabenweg, setID int64,
) int {
	felder, werte, fehler := weg.Angaben(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	antwort := angabenAntwort{Felder: felder, Werte: werte}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Werte) == 0 {
		an.sag("Keine Angabe gesetzt. Mögliche Felder: %d\n", len(antwort.Felder))

		return an.abschluss()
	}

	for _, wert := range antwort.Werte {
		an.sag("%s\t%s\n", wert.Feld, wert.Wert)
	}

	return an.abschluss()
}

// setwertSchreiben setzt eine Angabe und nennt die vorige.
func setwertSchreiben(
	ctx context.Context, an *schreiber, weg *app.Setangabenweg, setID int64, feld, wert string,
) int {
	// **Der Feldname wird geprüft, der Wert nicht** — dieselbe Regel wie im
	// Fenster: ein leerer Name ergäbe eine Zeile, die niemand wiederfindet;
	// ein leerer Wert ist ein gültiger Wert.
	if befund := domain.PruefeFeldname(feld); befund != nil {
		an.klageUeber(befund)

		return CodeFachlich
	}

	vorher, hatte, fehler := weg.Setzen(ctx, setID, feld, wert)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, feldAntwort{Vorher: vorher, Hatte: hatte})
	}

	if !hatte {
		an.sag("Gesetzt. Die Angabe gab es vorher nicht.\n")

		return an.abschluss()
	}

	an.sag("Gesetzt. Vorher: %s\n", vorher)

	return an.abschluss()
}

// vorbehaltebefehl zeigt, was beim Einfahren nicht gedeutet wurde.
//
// **Es ist die Gegenprobe zu `mse oeffnen`.** Jenes nennt die Befunde eines
// Laufs; dieses sagt, was von einem längst eingefahrenen Set übrig blieb —
// und ob eine Rücksicherung es wieder mitnähme.
func vorbehaltebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("vorbehalte braucht genau eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.SetangabennahtUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		vorbehalte, fehler := weg.Vorbehalte(ctx, setID)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return vorbehalteZeigen(an, vorbehalteAntwort{Vorbehalte: vorbehalte})
	})
}

// vorbehalteZeigen schreibt die Vorbehalte.
func vorbehalteZeigen(an *schreiber, antwort vorbehalteAntwort) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Vorbehalte) == 0 {
		an.sag("Kein Vorbehalt — alles gedeutet.\n")

		return an.abschluss()
	}

	for _, vorbehalt := range antwort.Vorbehalte {
		an.sag("%s:%d\t%s\n", vorbehalt.Datei, vorbehalt.Zeile, vorbehalt.Inhalt)
	}

	return an.abschluss()
}

// beispielsetsbefehl nennt die mitgelieferten Sets — **ohne Bestand**.
//
// **Und deshalb ohne `mitBestand`**: er liest ein Verzeichnis im Bündel. Wer
// dafür eine Datenbank öffnete, machte ihn von einem Zustand abhängig, den er
// nicht hat — dieselbe Begründung wie bei `scryfall-regeln` (#888).
func beispielsetsbefehl(an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("beispielsets nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	sets, befund := app.Beispielsetliste()
	if befund != nil {
		an.klageUeberFehler(befund)

		return CodeFachlich
	}

	antwort := beispielsetsAntwort{Sets: sets}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Sets) == 0 {
		an.sag("Kein mitgeliefertes Set gefunden.\n")

		return an.abschluss()
	}

	for _, set := range antwort.Sets {
		an.sag("%s\t%s\n", set.Name, set.Pfad)
	}

	return an.abschluss()
}
