// **Eine Karte aus einem Namen** —
// #1032.
//
// Eine eigene Datei, weil es ein eigener Vorgang ist: `scryfall.go` gleicht
// eine **vorhandene** Karte ab — hier entsteht eine.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Fremdanlage ist, was das Anlegewerkzeug braucht.
type Fremdanlage interface {
	// Anlegen legt eine Karte an und füllt sie aus Scryfall.
	Anlegen(ctx context.Context, setID int64, name string) (Anlage, error)
}

// Anlage ist, was ein Anlegen zurückmeldet.
//
// **Die Marken sind die Zusage** — dieselbe Überlegung wie bei den Zeilen in
// `gestalt.go`: ohne sie nähme `encoding/json` die Go-Feldnamen, und ein
// umbenanntes Feld verschöbe stillschweigend eine Fläche, die
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// über Fassungen hinweg zusagt. Angestrichen von den Flächenläufen
// (#1028):
// die Namen standen groß, aber nur, weil Go sie so schreibt.
type Anlage struct {
	KarteID      int64    `json:"KarteID"`
	Felder       []string `json:"Felder"`
	Befunde      []string `json:"Befunde"`
	Herkunft     string   `json:"Herkunft"`
	AlterStunden float64  `json:"AlterStunden"`
}

// scryfallkartenwerkzeug gibt das Werkzeug, das eine Karte anlegt.
//
// # Warum es keinen Vorschlag davor gibt
//
// Der Feldabgleich zeigt erst und schreibt dann, weil ein Wert von Scryfall
// einen **getippten überschreiben** könnte. Eine Karte, die es noch nicht
// gibt, hat diesen Fall nicht: es gibt nichts zu überschreiben, und der
// Vorschlag wäre eine Rückfrage auf eine leere Karte.
//
// # Und warum 0044 es trägt
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// misst die **Umkehrbarkeit**: das Werkzeug meldet die angelegte Kartenkennung
// zurück, und zurückgenommen wird über `schreiben.karte.loeschen`. Ein
// Werkzeug, das anlegt und die Kennung verschwiege, wäre nicht umkehrbar.
//
// **Ohne Bild.** Das ist ein eigener Weg mit einer eigenen Bedingung; wer es
// will, ruft danach `schreiben.scryfall.bild` — mit der Kennung von hier.
func scryfallkartenwerkzeug(anlage Fremdanlage) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.scryfall.karte", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt eine NEUE Karte im Set an und füllt " +
				"sie mit den Feldern, die Scryfall zu diesem Namen führt — Name, " +
				"Manakosten, Typen, Regeltext, Aromatext, Stärke, Widerstand, " +
				"Seltenheit, Illustrator. Gibt die Kennung der angelegten Karte " +
				"zurück — AUCH wenn das Füllen danach scheitert, damit der halbe " +
				"Stand zurücknehmbar bleibt; zurückgenommen wird über " +
				"»schreiben.karte.loeschen«. " +
				"Das Bild kommt NICHT mit — dafür danach »schreiben.scryfall.bild« " +
				"mit dieser Kennung.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"name":  text("Der Kartenname, wie Scryfall ihn führt."),
			}, "setID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Name  string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				angelegt, fehler := anlage.Anlegen(ctx, argumente.SetID, argumente.Name)
				if fehler != nil {
					return "", mitTeilstand(fehler, angelegt)
				}

				return alsJSON(angelegt)
			},
		},
	}
}

// mitTeilstand hängt an den Fehler, was schon geschrieben wurde
// (#1041).
//
// # Warum in die Meldung und nicht in ein eigenes Feld
//
// **Weil das Protokoll dafür keinen Platz hat.** Ein gescheiterter
// Werkzeugaufruf trägt `isError: true` und **einen** Text; ein zweites Feld
// daneben läse kein Gegenüber. Der Text ist damit der einzige Ort, an dem der
// halbe Stand ankommt — und er muss dort ankommen: die Erklärung dieses
// Werkzeugs sagt zu, dass die Kennung zurückkommt, und ohne sie ist der
// Aufruf nicht umkehrbar (0010, 0044).
//
// **Ohne angelegte Karte bleibt der Fehler, wie er ist.** Ein Aufruf, der vor
// der Anlage scheitert, hat keinen halben Stand.
func mitTeilstand(fehler error, angelegt Anlage) error {
	if angelegt.KarteID == 0 {
		return fehler
	}

	// **Aus den Angaben gebaut und nicht als JSON.** Ein `alsJSON` hier hätte
	// einen Fehlerzweig, den kein Aufruf betreten kann — [Anlage] ist ein
	// schlichter Verbund und lässt sich immer schreiben (§3).
	return fmt.Errorf(
		"%w — die Karte %d ist angelegt und teilweise gefüllt (%s); "+
			"zurückgenommen wird sie über »schreiben.karte.loeschen«",
		fehler, angelegt.KarteID, gefuellt(angelegt.Felder))
}

// gefuellt nennt die geschriebenen Felder.
func gefuellt(felder []string) string {
	if len(felder) == 0 {
		return "kein Feld gefüllt"
	}

	return "gefüllt: " + strings.Join(felder, ", ")
}
