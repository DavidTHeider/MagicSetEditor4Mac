// **Die Dateien eines installierten Pakets an der Kommandozeile** —
// #1164.
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Wer ein Skript in
// `card_fields` ändert, will es danach mit `mse zeichnen` ansehen — beides
// ohne Fenster.
package main

import (
	"context"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// paketdateienbefehl zählt die bearbeitbaren Dateien eines Pakets auf.
//
// **Das Beiwerk steht nur im JSON** (FMT-02, #1691). In der Textausgabe wäre
// es eine brechende Änderung (0033, Regel 6): eine Schleife über die Zeilen
// bekäme Namen, die `paketdatei` als Beiwerk abweist.
func paketdateienbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("paketdateien braucht den Paketnamen.\n%s", hilfe)

		return CodeBedienung
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	dateien, beiwerk, listenfehler := naht.DateienUndBeiwerk(ctx, rest[0])
	if listenfehler != nil {
		an.klageUeber(listenfehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"dateien": dateien, "beiwerk": beiwerk})
	}

	for _, datei := range dateien {
		an.sag("%s\n", datei)
	}

	return an.abschluss()
}

// paketdateibefehl gibt eine Paketdatei aus.
//
// **Der Stand steht nur im JSON.** Wer die Datei liest, um sie anzusehen,
// will ihn nicht mitten im Text; wer sie ändern will, ruft mit `--json`.
func paketdateibefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paketdatei braucht Paket und Datei.\n%s", hilfe)

		return CodeBedienung
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	inhalt, stand, lesefehler := naht.Lesen(ctx, rest[0], rest[1])
	if lesefehler != nil {
		an.klageUeber(lesefehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"inhalt": inhalt, "stand": stand})
	}

	an.sag("%s", inhalt)

	return an.abschluss()
}

// paketdateisetzenbefehl schreibt eine Paketdatei aus einer lokalen Datei.
//
// **Aus einer Datei und nicht von der Kommandozeile.** Eine Vorlage hat
// Zeilenumbrüche, Tabulatoren und Anführungszeichen; sie durch eine Shell zu
// reichen hieße, sie dabei zu verändern.
//
// **Der Stand steht als Argument und nicht als Flagge.** Er ist keine
// Ausgabeform, sondern die Bedingung des Schreibens — wer ihn weglassen
// könnte, schriebe ohne Schutz.
func paketdateisetzenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 4 {
		an.klage("paketdatei-setzen braucht Paket, Datei, Quelle und Stand.\n%s", hilfe)

		return CodeBedienung
	}

	inhalt, lesefehler := os.ReadFile(rest[2]) // #nosec G304,G703 -- der Aufrufer sitzt an dieser Kommandozeile
	if lesefehler != nil {
		an.klageUeber(lesefehler)

		return CodeFachlich
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	stand, schreibfehler := naht.Schreiben(ctx, rest[0], rest[1], string(inhalt), rest[3])
	if schreibfehler != nil {
		an.klageUeber(schreibfehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"stand": stand})
	}

	an.sag("%s\n", stand)

	return an.abschluss()
}

// paketdateipruefenbefehl prüft die Skripte einer Paketdatei.
//
// **§16: dieselbe Fähigkeit, überall dieselbe Antwort.** Für die Einfärbung
// steht in `docs/FAEHIGKEITEN.md` „ohne Fenster wird nicht getippt" — und für
// eine **Prüfung** trägt das nicht: sie ist eine Aussage über die Datei, nicht
// die Darstellung eines Textfelds. Wer eine Vorlage in einem Ablauf ändert,
// will wissen, ob sie noch zerteilt, bevor er sie ausliefert.
//
// **Die Datei kommt aus dem Paket und nicht von der Platte**: geprüft wird,
// was installiert ist, damit dieselbe Frage dieselbe Antwort gibt wie im
// Fenster.
func paketdateipruefenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paketdatei-pruefen braucht Paket und Datei.\n%s", hilfe)

		return CodeBedienung
	}

	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	inhalt, _, lesefehler := naht.Lesen(ctx, rest[0], rest[1])
	if lesefehler != nil {
		an.klageUeber(lesefehler)

		return CodeFachlich
	}

	zeile, spalte, meldung := app.PaketSkripteGeprueft(inhalt)

	// **`--json` ändert die Ausgabe und nicht den Ausgang** (#2400, Vermerk
	// 0033): bis dahin endete ein Befund hier mit 0, und ein Ablauf, der
	// maschinenlesbar prüfte, sah einen kaputten Stil als Erfolg.
	if an.alsJSON {
		if code := alsJSONSchreiben(an, pruefbericht(zeile, spalte, meldung)); code != 0 {
			return code
		}

		return urteilscode(zeile == 0)
	}

	if zeile == 0 {
		an.sag("Kein Syntaxfehler.\n")

		return an.abschluss()
	}

	an.sag("Zeile %d, Spalte %d: %s\n", zeile, spalte, meldung)

	// **Ein Befund ist ein fachliches Ergebnis, kein Fehler des Aufrufs.**
	// Ein Ablauf soll daran scheitern können — deshalb der Code, und nicht
	// bloß eine Zeile auf der Ausgabe.
	return CodeFachlich
}

// pruefbericht baut die JSON-Gestalt eines Befunds.
//
// **`gefunden` und nicht bloß ein fehlendes Feld**: ein Ablauf, der `.stelle`
// liest, bekommt bei einer heilen Datei `null` — und `null` heißt in JSON
// beides, „kein Befund" und „das Feld gibt es nicht mehr".
func pruefbericht(zeile, spalte int, meldung string) map[string]any {
	if zeile == 0 {
		return map[string]any{"gefunden": false}
	}

	return map[string]any{
		"gefunden": true,
		"zeile":    zeile,
		"spalte":   spalte,
		"meldung":  meldung,
	}
}
