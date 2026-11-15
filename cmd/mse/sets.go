// Der Befehl `sets` — die erste Fähigkeit auf der Kommandozeile.
//
// **Eine Datei je Fähigkeit**, damit `lauf.go` die Verteilung bleibt und nicht
// zur Sammelstelle wird (§2).

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// setDTO ist ein Set, wie die Kommandozeile es ausgibt.
//
// **Ein eigener Typ und nicht `domain.SetHead`.** Was über diese Fläche geht,
// ist eine **Zusage** (§16 und [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)):
// wer sie in einem Skript liest, darf sich darauf verlassen. Eine Entität, die
// sich mit der Fachlichkeit ändert, wäre eine Zusage, die niemand gegeben hat.
type setDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Spiel string `json:"spiel"`
	Stil  string `json:"stil"`
}

// setsAntwort ist die Hülle — **mit einer Liste, die nie `null` ist**.
//
// `omitempty` steht hier nicht, und das ist keine Nachlässigkeit: eine leere
// Liste, die aus der Antwort verschwindet, kommt beim Aufrufer als `null` an
// (§3). Ein `jq '.sets | length'` gäbe dann einen Fehler statt einer Null.
type setsAntwort struct {
	Sets []setDTO `json:"sets"`
}

// setsAuflisten gibt die Sets des Bestands.
func setsAuflisten(ctx context.Context, db *sql.DB, an *schreiber) int {
	koepfe, err := storage.NeuerSetLeser(db).Setkoepfe(ctx)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	antwort := setsAntwort{Sets: make([]setDTO, 0, len(koepfe))}
	for _, kopf := range koepfe {
		antwort.Sets = append(antwort.Sets, setDTO{
			ID:    strconv.FormatInt(kopf.ID, 10),
			Name:  kopf.Name,
			Spiel: kopf.Game,
			Stil:  kopf.Style,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Sets) == 0 {
		// **„Kein Set" ist eine Antwort**, kein leerer Bildschirm (§6).
		an.sag("Kein Set im Bestand.\n")

		return an.abschluss()
	}

	for _, satz := range antwort.Sets {
		an.sag("%s\t%s\t%s\t%s\n", satz.ID, satz.Name, satz.Spiel, satz.Stil)
	}

	return an.abschluss()
}

// alsJSONSchreiben gibt eine Antwort maschinenlesbar aus.
//
// **Eingerückt**, weil ein Mensch sie auch liest, wenn er `--json` tippt, um
// nachzusehen, was ein Skript bekäme.
func alsJSONSchreiben(an *schreiber, antwort any) int {
	roh, err := json.MarshalIndent(antwort, "", "  ")
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	an.sag("%s\n", roh)

	return an.abschluss()
}

// setAnlegeAntwort ist die Hülle für ein angelegtes Set (#891).
type setAnlegeAntwort struct {
	ID string `json:"id"`
}

// setAnlegen legt ein leeres Set an.
//
// **Dieselbe Fähigkeit wie im Fenster** (§16): `SetAnlegen` dort, `set-neu`
// hier. Wer ein Set aus einem Skript heraus aufbauen will — Karten anlegen,
// Felder setzen, ausgeben —, brauchte bis dahin eine Datei, die es noch nicht
// gibt.
func setAnlegen(ctx context.Context, db *sql.DB, an *schreiber, name, spiel, stil string) int {
	kennung, err := app.Einfuhrnaht(db).Neu(ctx, name, spiel, stil)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	antwort := setAnlegeAntwort{ID: strconv.FormatInt(kennung, 10)}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.ID)

	return an.abschluss()
}

// setLoeschen nimmt ein Set aus dem Bestand — endgültig.
//
// **Der Name steht als zweites Argument und ist die Bestätigung.** Auf der
// Kommandozeile gibt es keine Rückfrage; wer den Namen tippt, hat ihn
// nachgesehen. Ein `--ja`-Schalter täte das nicht: er ließe sich vorsorglich
// in ein Skript schreiben, der Name nicht.
func setLoeschen(ctx context.Context, db *sql.DB, an *schreiber, setID, name string) int {
	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		an.klage("%q ist keine Setkennung.", setID)

		return CodeBedienung
	}

	befunde, loeschfehler := app.SetloeschnahtUeber(db).SetLoeschen(ctx, kennung, name)
	if loeschfehler != nil {
		an.klageUeber(loeschfehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, setLoeschAntwort{Geloescht: true, Befunde: ohneNilTexte(befunde)})
	}

	for _, befund := range befunde {
		an.sag("%s\n", befund)
	}

	return an.abschluss()
}

// setLoeschAntwort ist die JSON-Gestalt.
type setLoeschAntwort struct {
	Geloescht bool     `json:"geloescht"`
	Befunde   []string `json:"befunde"`
}

// ohneNilTexte macht aus `nil` eine leere Liste.
func ohneNilTexte(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}
