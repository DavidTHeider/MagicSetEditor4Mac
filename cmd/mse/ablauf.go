// Die Befehle `oeffnen` und `sichern` — die beiden Enden des Arbeitsablaufs
// (#886).
//
// **Sie sind die zwei, die im Fenster an einem Dateidialog hängen.** Was der
// Dialog dort verdeckt — welche Datei? —, ist hier ein Argument, und das ist
// der Grund, warum der Fähigkeitskatalog nach Fähigkeiten fragt und nicht nach
// Bindungen.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// oeffnenAntwort sagt, was beim Einfahren herauskam.
//
// **Die Befunde reisen mit** (§8): ein Set, von dem 300 Karten lesbar sind und
// zwei nicht, wird mit 300 Karten und zwei benannten Befunden geöffnet — nicht
// abgewiesen. Wer das auf der Kommandozeile nicht sähe, hielte ein halbes Set
// für ein ganzes.
type oeffnenAntwort struct {
	SetID   string   `json:"setID"`
	Karten  int      `json:"karten"`
	Befunde []string `json:"befunde"`
}

// speicherAntwort sagt, wohin geschrieben wurde.
type speicherAntwort struct {
	Pfad    string   `json:"pfad"`
	Befunde []string `json:"befunde"`
}

// setEinfahren liest ein Setpaket in den Bestand.
func setEinfahren(ctx context.Context, db *sql.DB, an *schreiber, pfad string) int {
	ergebnis, err := app.Einfuhrnaht(db).Set(ctx, pfad)
	// **Der Teilerfolg wird genannt, bevor geklagt wird** — Befund 1 des
	// Nachlaufs über Phase 4.4 (#1291). `Importeur.Set` hängt einen
	// Schließfehler des Archivs an ein **geglücktes** Einfahren an; das Set
	// steht dann im Bestand, und seine Kennung ist das, was ein Skript
	// braucht (§7 Regel 3).
	if err != nil {
		if ergebnis != nil {
			an.klage("angelegt wurde Set %d\n", ergebnis.SetID)
		}

		an.klageUeber(err)

		return CodeFachlich
	}

	antwort := oeffnenAntwort{
		SetID:   strconv.FormatInt(ergebnis.SetID, 10),
		Karten:  ergebnis.Karten,
		Befunde: make([]string, 0, len(ergebnis.Befunde)),
	}

	for _, befund := range ergebnis.Befunde {
		antwort.Befunde = append(antwort.Befunde, befund.String())
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Eingefahren als Set %s mit %d Karten.\n", antwort.SetID, antwort.Karten)

	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}

// setSchreiben schreibt ein Set in eine Datei.
//
// **Das Ziel ist Pflicht**, anders als im Fenster. Dort heißt ein leeres Ziel
// „dorthin zurück, wo es herkam", und das ist richtig: wer auf Speichern
// drückt, meint seine Datei. Auf der Kommandozeile wäre dasselbe ein
// Überschreiben ohne Rückfrage — der Aufruf sähe aus wie ein Ausdruck und
// änderte eine Datei.
func setSchreiben(ctx context.Context, db *sql.DB, an *schreiber, setID int64, ziel string) int {
	if !kenntSetIm(ctx, db, an, setID) {
		return CodeFachlich
	}

	schreiber, err := app.Setschreibnaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	ergebnis, fehler := schreiber.Schreibe(ctx, setID, ziel)

	// **Der Teilerfolg wird genannt** (#1291, Nachlauf): scheitert nur das
	// Setzen des Arbeitsorts, **liegt die Datei schon am Ziel**.
	if fehler != nil {
		if ergebnis != nil {
			an.klage("geschrieben wurde %s\n", ergebnis.Pfad)
		}

		an.klageUeber(fehler)

		return CodeFachlich
	}

	antwort := speicherAntwort{Pfad: ziel, Befunde: make([]string, 0, len(ergebnis.Befunde))}
	for _, befund := range ergebnis.Befunde {
		antwort.Befunde = append(antwort.Befunde, befund.String())
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Geschrieben nach %s.\n", ziel)

	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}

// kenntSetIm sagt, ob es das Set gibt — und klagt, wenn nicht.
//
// **Dieselbe Frage wie in [kartenAuflisten]**, und derselbe Grund: auf der
// Kommandozeile ist eine Kennung ein Argument und damit ein möglicher
// Tippfehler.
func kenntSetIm(ctx context.Context, db *sql.DB, an *schreiber, setID int64) bool {
	koepfe, err := storage.NeuerSetLeser(db).Setkoepfe(ctx)
	if err != nil {
		an.klageUeber(err)

		return false
	}

	if !kenntSet(koepfe, setID) {
		// **Über den Domänenfehler und nicht als eigener Satz** — Befund 3
		// des Nachlaufs über Phase 3.8
		// (#1253).
		//
		// Ein eigener Satz über `an.klage` trüge seine **Klasse** nicht: der
		// Dienst antwortet `"code":"nichtGefunden"`, und ein Lauf, der „überall
		// gleich" zusagt, könnte das an der Kommandozeile nicht prüfen.
		an.klageUeber(fmt.Errorf("%w: das Set %d gibt es nicht",
			domain.ErrNotFound, setID))

		return false
	}

	return true
}
