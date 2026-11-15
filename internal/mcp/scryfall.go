package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdabgleich ist, was die beiden Scryfall-Werkzeuge brauchen.
//
// Die Schnittstelle sagt, was `mcp` **braucht** (§2). Erfüllt wird sie im
// Composition Root des Servers.
type Fremdabgleich interface {
	// Vorschlag sagt, was sich ändern würde — und ändert nichts.
	Vorschlag(ctx context.Context, setID, karteID int64, name string) (Vorschlag, error)

	// Uebernehmen schreibt die genannten Felder — und nur die.
	//
	// **Es gibt die vorigen Werte zurück**, denn ein schreibendes Werkzeug
	// sagt hier, wie es zurückzunehmen ist
	// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
	// Ohne sie müsste ein Gegenüber den Vorschlag aufgehoben haben — und ein
	// Modell, das dazwischen den Faden verliert, hätte keinen Rückweg.
	Uebernehmen(ctx context.Context, setID, karteID int64, name string, felder []string) (
		vorher []Aenderung, fehler error)
}

// Vorschlag ist, was eine Übernahme ändern würde.
//
// **Eine eigene Gestalt und nicht `service.Vorschlag`** — dieselbe Überlegung
// wie in `gestalt.go`: MCP ist eine Fläche nach außen (0033, Regel 5), und
// eine Entität, die nach draußen geht, ändert ihren Vertrag mit der
// Fachlichkeit.
type Vorschlag struct {
	Aenderungen  []Aenderung
	Befunde      []string
	Herkunft     string
	AlterStunden int
}

// Aenderung ist ein Feld, das sich unterscheidet.
type Aenderung struct {
	Feld string
	Alt  string
	Neu  string
}

// scryfallwerkzeuge gibt die beiden Werkzeuge für den Abgleich.
//
// # Warum zwei und nicht eines
//
// §16, Regel 6: „Nichts wird still übernommen. Was von außen in ein Kartenfeld
// will, zeigt vorher, was sich ändert." Für ein Sprachmodell ist das die
// wichtigere Hälfte der Regel — es ruft, während niemand hinsieht.
//
// **Das erste ist lesend**, das zweite schreibend, und die Namen sagen es.
//
// # Und warum das Schreibende nach 0044 erlaubt ist
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// misst die **Umkehrbarkeit**: ein Werkzeug darf schreiben, was es nicht
// überschreibt. Eine Übernahme **überschreibt** einen Feldwert — aber sie gibt
// den vorigen zurück, und der Weg dorthin ist derselbe wie bei
// `schreiben.feld`, das seit 0010 erlaubt ist. Wer zurück will, ruft
// `schreiben.feld` mit dem alten Wert.
func scryfallLesend(abgleich Fremdabgleich) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.scryfall.vorschlag", Art: Lesend,
			Erklaerung: "Zeigt, was sich an einer Karte ändern würde, wenn man die Angaben " +
				"von Scryfall übernimmt. ÄNDERT NICHTS. Die Antwort nennt je Feld den " +
				"alten und den neuen Wert, die Herkunft und wie alt die Auskunft ist.",
			Schema: objekt(map[string]any{
				"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
				"name":    text("Der Kartenname, wie Scryfall ihn führt."),
			}, "setID", "karteID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID   int64  `json:"setID"`
					KarteID int64  `json:"karteID"`
					Name    string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
					return "", err
				}

				vorschlag, fehler := abgleich.Vorschlag(ctx,
					argumente.SetID, argumente.KarteID, argumente.Name)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(alsVorschlagsgestalt(vorschlag))
			},
		},
	}
}

// scryfallSchreibend gibt das Werkzeug, das übernimmt.
//
// **Getrennt vom lesenden, und das ist die Reihenfolge aus 0010**: „die
// lesenden zuerst, die schreibenden danach — ein Gegenüber, das die Liste
// kürzt, behält damit die harmlosen." Ein Wächter hält sie.
func scryfallSchreibend(abgleich Fremdabgleich) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.scryfall.uebernehmen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: übernimmt die genannten Felder aus " +
				"»lesen.scryfall.vorschlag«. Ein Feld, das dort nicht steht, wird " +
				"abgewiesen. Zurücknehmen mit »schreiben.feld« und dem alten Wert, " +
				"den der Vorschlag nennt.",
			Schema: objekt(map[string]any{
				"setID":   zahl("Die Kennung des Sets."),
				"karteID": zahl("Die Kennung der Karte."),
				"name":    text("Derselbe Kartenname wie beim Vorschlag."),
				"felder": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Die Feldnamen aus dem Vorschlag, die übernommen werden sollen.",
				},
			}, "setID", "karteID", "name", "felder"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID   int64    `json:"setID"`
					KarteID int64    `json:"karteID"`
					Name    string   `json:"name"`
					Felder  []string `json:"felder"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
					return "", err
				}

				vorher, fehler := abgleich.Uebernehmen(ctx, argumente.SetID,
					argumente.KarteID, argumente.Name, argumente.Felder)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{
					"uebernommen":  argumente.Felder,
					"vorher":       alsAenderungen(vorher),
					"zuruecknahme": "je Feld »schreiben.feld« mit dem Wert aus »vorher«",
				})
			},
		},
	}
}

// alsAenderungen bringt die vorigen Werte in die Gestalt dieser Fläche.
func alsAenderungen(aenderungen []Aenderung) []map[string]any {
	gestalt := make([]map[string]any, 0, len(aenderungen))
	for _, eine := range aenderungen {
		gestalt = append(gestalt, map[string]any{
			"Feld": eine.Feld, "Alt": eine.Alt, "Neu": eine.Neu,
		})
	}

	return gestalt
}

// alsVorschlagsgestalt bringt einen Vorschlag in die Gestalt dieser Fläche.
//
// **Die leeren Listen bleiben Listen.** Ein `nil`-Slice würde zu `null`, und
// ein Gegenüber, das darüber schleift, bekäme einen Fehler statt keiner Runde.
func alsVorschlagsgestalt(vorschlag Vorschlag) map[string]any {
	aenderungen := make([]map[string]any, 0, len(vorschlag.Aenderungen))
	for _, eine := range vorschlag.Aenderungen {
		aenderungen = append(aenderungen, map[string]any{
			"Feld": eine.Feld, "Alt": eine.Alt, "Neu": eine.Neu,
		})
	}

	befunde := vorschlag.Befunde
	if befunde == nil {
		befunde = []string{}
	}

	return map[string]any{
		"Aenderungen":  aenderungen,
		"Befunde":      befunde,
		"Herkunft":     vorschlag.Herkunft,
		"AlterStunden": vorschlag.AlterStunden,
	}
}

// Fremdauskuenfte ist, was die lesenden Auskunftswerkzeuge brauchen.
type Fremdauskuenfte interface {
	Regelhinweise(ctx context.Context, name string) (Regelauskunft, error)
	Manakosten(ctx context.Context, kosten string) (unbekannt []string, fehler error)
}

// Regelauskunft sind die Hinweise zu einer Karte.
type Regelauskunft struct {
	Kartenname   string
	Hinweise     []Regelhinweis
	Adresse      string
	Befunde      []string
	AlterStunden int
}

// Regelhinweis ist ein einzelner Hinweis.
type Regelhinweis struct {
	// Offiziell sagt, ob er von Wizards of the Coast stammt.
	//
	// **Für ein Sprachmodell ist das die wichtigere Hälfte.** Eine Anmerkung
	// des Scryfall-Teams erklärt ein inoffizielles Format; sie als Regel
	// weiterzugeben wäre falsch, und der bloße Wert `scryfall` sagt das nicht
	// von selbst.
	Offiziell bool

	Quelle string
	Datum  string
	Text   string
}

// scryfallauskunftswerkzeuge gibt die beiden lesenden Werkzeuge.
//
// **Beide sind lesend**, und deshalb stehen sie vorn in der Liste — die
// Reihenfolge aus [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md).
func scryfallauskunftswerkzeuge(auskunft Fremdauskuenfte) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.scryfall.regeln", Art: Lesend, Idempotent: true,
			Erklaerung: "Gibt die offiziellen Regelhinweise zu einer Karte. ÄNDERT NICHTS. " +
				"Jeder Hinweis sagt, ob er von Wizards of the Coast stammt (offiziell) " +
				"oder eine Anmerkung des Scryfall-Teams zu einem inoffiziellen Format ist.",
			Schema: objekt(map[string]any{
				"name": text("Der Kartenname, wie Scryfall ihn führt."),
			}, "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if argumente.Name == "" {
					return "", fmt.Errorf(
						"%w: es wurde kein Kartenname genannt", domain.ErrValidation)
				}

				geholt, fehler := auskunft.Regelhinweise(ctx, argumente.Name)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(geholt)
			},
		},
		{
			Name: "lesen.scryfall.manakosten", Art: Lesend, Idempotent: true,
			Erklaerung: "Prüft einen Manakosten wie »{2}{R}« gegen die Symbolliste von " +
				"Scryfall. ÄNDERT NICHTS. Die Antwort nennt die Symbole, die es " +
				"in einem Manakosten nicht gibt — eine leere Liste heißt: in Ordnung.",
			Schema: objekt(map[string]any{
				"kosten": text("Der Manakosten in geschweiften Klammern, etwa »{2}{R}«."),
			}, "kosten"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Kosten string `json:"kosten"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Ein leerer Kosten wird abgewiesen** (#888) — wie an der
				// Kommandozeile und über REST. Er käme sonst mit „in
				// Ordnung" zurück: eine Auskunft über nichts, und §16 sagt,
				// dass dieselbe Fähigkeit überall dieselbe Antwort gibt.
				if argumente.Kosten == "" {
					return "", fmt.Errorf(
						"%w: es wurde kein Manakosten genannt", domain.ErrValidation)
				}

				unbekannt, fehler := auskunft.Manakosten(ctx, argumente.Kosten)
				if fehler != nil {
					return "", fehler
				}

				if unbekannt == nil {
					unbekannt = []string{}
				}

				return alsJSON(map[string]any{"unbekannt": unbekannt})
			},
		},
	}
}
