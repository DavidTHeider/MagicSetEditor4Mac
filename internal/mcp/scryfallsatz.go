// **Ein ganzer Satz von Scryfall** —
// #1033.

package mcp

import (
	"context"
	"encoding/json"
)

// Fremdsaetze ist, was das Satzwerkzeug braucht.
type Fremdsaetze interface {
	Einlesen(ctx context.Context, code, spiel string) (Satz, error)
}

// Satz ist, was ein Einlesen zurückmeldet.
//
// **Die Marken sind die Zusage** — dieselbe Überlegung wie bei den Zeilen in
// `gestalt.go`: ohne sie nähme `encoding/json` die Go-Feldnamen, und ein
// umbenanntes Feld verschöbe stillschweigend eine Fläche, die
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// über Fassungen hinweg zusagt. Angestrichen von den Flächenläufen
// (#1028):
// die Namen standen groß, aber nur, weil Go sie so schreibt.
type Satz struct {
	SetID      int64    `json:"SetID"`
	Karten     int      `json:"Karten"`
	Befunde    []string `json:"Befunde"`
	SchonOffen bool     `json:"SchonOffen"`
}

// scryfallsatzwerkzeug gibt das Werkzeug, das einen Satz einliest.
//
// # Warum 0044 es trägt
//
// **Es überschreibt nichts** — „eine Quelle, ein Set" (#886): ein zweiter
// Aufruf mit demselben Satzcode legt kein zweites Set an, sondern meldet das
// vorhandene. Es steht damit in derselben Gruppe wie `schreiben.set.einfahren`.
//
// **Und es sagt trotzdem keine Idempotenz zu.** Der erste Aufruf schreibt
// einen Verlaufsschritt, der zweite nicht — die Antworten sind gleich, die
// **Wirkung** ist es nicht. Dieselbe Haltung wie beim Einfahren: die Zusage
// gilt für die Wirkung, nicht für den Text der Antwort.
//
// **Zurückgenommen wird über `schreiben.set.loeschen`** — den es nicht gibt.
// Das gehört gesagt: ein eingelesenes Set bleibt, bis es jemand im Fenster
// entfernt. Deshalb steht die Setkennung in der Antwort.
func scryfallsatzwerkzeug(saetze Fremdsaetze) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.scryfall.set", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: liest einen GANZEN Satz von Scryfall als " +
				"neues Set ein — alle Karten mit Name, Manakosten, Typen, Regeltext, " +
				"Aromatext, Stärke, Widerstand, Seltenheit und Illustrator. " +
				"code ist das Satzkürzel (»lea«, »m15«, »neo«), spiel das MSE-Spiel " +
				"(meist »magic«). Bilder kommen NICHT mit — dafür je Karte " +
				"»schreiben.scryfall.bild«. " +
				"Ein zweiter Aufruf mit demselben code legt KEIN zweites Set an, " +
				"sondern meldet das vorhandene.",
			Schema: objekt(map[string]any{
				"code":  text("Das Satzkürzel bei Scryfall, etwa »lea«."),
				"spiel": text("Das MSE-Spiel, meist »magic«."),
			}, "code", "spiel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Code  string `json:"code"`
					Spiel string `json:"spiel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				eingelesen, fehler := saetze.Einlesen(ctx, argumente.Code, argumente.Spiel)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(eingelesen)
			},
		},
	}
}
