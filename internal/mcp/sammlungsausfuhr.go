// **Die Sammlung ausgeben und zurücklesen — über MCP** — Werkstück 4 von
// Phase 4.5 (#1292).

package mcp

import (
	"context"
	"encoding/json"
)

// Sammlungsausfuhr ist, was MCP zum Ausgeben und Zurücklesen braucht.
type Sammlungsausfuhr interface {
	Ausgeben(ctx context.Context, form string) (string, error)
	CSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (Einfuhrbericht, error)

	// CSVVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	CSVVorschauen(ctx context.Context, text string) (Sammelvorschau, error)
}

// sammlungsausfuhr gibt die beiden Werkzeuge — lesend zuerst.
func sammlungsausfuhr(ausfuhr Sammlungsausfuhr) []Werkzeug {
	if ausfuhr == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.sammlung.ausgeben", Art: Lesend,
		Erklaerung: "Gibt die ganze Sammlung als Text. »csv« trägt jede Spalte " +
			"— Zustand, Sprache, Erwerbsdatum, Preis, Währung und Notiz — und " +
			"lässt sich mit »schreiben.sammlung.einlesen« vollständig " +
			"zurücklesen. »liste« gibt die einfügbare Gestalt (»4 Lightning " +
			"Bolt (M11) 149«) und ist VERLUSTBEHAFTET: sie trägt nur Menge, " +
			"Karte, Druck und Ausführung.",
		Schema: objekt(map[string]any{
			"form": text("»csv« für alles, »liste« für die einfügbare Gestalt."),
		}, "form"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Form string `json:"form"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			return ausfuhr.Ausgeben(ctx, argumente.Form)
		},
	}, {
		Name: "schreiben.sammlung.einlesen", Art: Schreibend,
		Erklaerung: "ÄNDERT die Sammlung: liest ein CSV zurück, wie " +
			"»lesen.sammlung.ausgeben« es schreibt. Jede Zeile nennt ihre " +
			"Druckkennung — es wird NICHTS gedeutet und nichts geraten. Eine " +
			"Zeile, die es schon gibt, wird ERHÖHT. Für eine Liste mit " +
			"Kartennamen statt Kennungen ist »lesen.liste.deuten« der Weg.",
		Schema: objekt(map[string]any{
			"text": text("Das CSV, mit einer Kopfzeile. Pflichtspalten sind " +
				"»printing« und »menge«; unbekannte Spalten werden übergangen."),

			// **Erweiternd im Sinne von 0033**: ein Feld kommt hinzu, und es
			// ist freiwillig. Ein älterer Aufrufer schickt es nicht und
			// bekommt dasselbe wie bisher.
			"quelle": text("Woher das CSV kommt — ein Dateiname oder eine " +
				"Bezeichnung. Sie wird an jeder NEU entstehenden Zeile " +
				"notiert; eine erhöhte Zeile bekommt keine, sie war schon " +
				"da. Leer heißt: es wird nichts behauptet."),
			"zusammenfuehrung": zusammenfuehrungsfeld(),
		}, "text"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Text             string `json:"text"`
				Quelle           string `json:"quelle"`
				Zusammenfuehrung string `json:"zusammenfuehrung"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			bericht, err := ausfuhr.CSVEinlesen(ctx, argumente.Text, argumente.Quelle,
				argumente.Zusammenfuehrung)
			if err != nil {
				return "", err
			}

			return alsJSON(bericht)
		},
	}, lesenSammlungsvorschau(ausfuhr)}
}
