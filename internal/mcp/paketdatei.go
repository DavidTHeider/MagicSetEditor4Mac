// **Die Dateien eines installierten Pakets** —
// #1164.
//
// # Wofür ein Sprachmodell das braucht
//
// Weil hier die **Fachlogik einer Vorlage** steht. Welche Felder eine Karte
// hat, wie ein Manasymbol gesetzt wird, was `card_fields` rechnet — das ist
// MSE-Script in einer Textdatei, und ohne diese Werkzeuge kann ein Modell es
// lesen, aber nicht ändern.

package mcp

import (
	"context"
	"encoding/json"
)

// Paketwerkstatt ist, was die Werkzeuge für Paketdateien brauchen.
type Paketwerkstatt interface {
	DateienUndBeiwerk(ctx context.Context, paket string) (dateien, beiwerk []string, fehler error)
	Lesen(ctx context.Context, paket, datei string) (inhalt, stand string, fehler error)
	Schreiben(ctx context.Context, paket, datei, inhalt, erwartet string) (string, error)

	// Pruefen nennt Zeile, Spalte und Satz des ersten Syntaxfehlers in den
	// Skripten eines Paketdateitextes — **Zeile 0**, wenn keiner darin steht.
	Pruefen(inhalt string) (zeile, spalte int, meldung string)

	// Der Dateibaum (FMT-02, #1691) — **ohne Einfügen von der Platte**: ein
	// Modell könnte damit eine beliebige Datei des Rechners in ein Paket
	// legen und über »lesen.paketdatei« wieder hinauslesen.
	Anlegen(ctx context.Context, paket, datei, inhalt string) (string, error)
	Umbenennen(ctx context.Context, paket, alt, neu, erwartet string) error
	Entfernen(ctx context.Context, paket, datei, erwartet string) error
}

// werkstattwerkzeuge gibt die drei Werkzeuge der Paketwerkstatt.
func werkstattwerkzeuge(werkstatt Paketwerkstatt) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.paketdateien", Art: Lesend,
			Erklaerung: "Zählt die Dateien eines installierten Pakets auf — unter " +
				"»Dateien« die bearbeitbaren Textdateien mit Feldern, Skripten und Maßen, " +
				"unter »Beiwerk« Bilder und Schriften, die sich nicht als Text lesen " +
				"lassen. ÄNDERT NICHTS. Der Paketname kommt aus »lesen.pakete«.",
			Schema: objekt(map[string]any{
				"paket": text("Der Name des Pakets, etwa »magic-m15.mse-style«."),
			}, "paket"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Paket string `json:"paket"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				dateien, beiwerk, fehler := werkstatt.DateienUndBeiwerk(ctx, argumente.Paket)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Dateien": leerStattNil(dateien), "Beiwerk": leerStattNil(beiwerk)})
			},
		},
		{
			Name: "lesen.paketdatei", Art: Lesend,
			Namensfelder: []string{"datei"},
			Erklaerung: "Gibt den Text einer Paketdatei samt ihrem STAND. ÄNDERT NICHTS. " +
				"Der Stand gehört zu »schreiben.paketdatei« — ohne ihn ist nicht zu " +
				"erkennen, ob inzwischen jemand anderes geschrieben hat.",
			Schema: objekt(map[string]any{
				"paket": text("Der Name des Pakets aus »lesen.pakete«."),
				"datei": text("Der Name der Datei aus »lesen.paketdateien«."),
			}, "paket", "datei"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Paket string `json:"paket"`
					Datei string `json:"datei"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				inhalt, stand, fehler := werkstatt.Lesen(ctx, argumente.Paket, argumente.Datei)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Inhalt": inhalt, "Stand": stand})
			},
		},
		{
			Name: "lesen.skriptbefund", Art: Lesend,
			Erklaerung: "Prüft die Skripte einer Paketdatei auf Syntaxfehler und nennt " +
				"die Stelle des ersten. ÄNDERT NICHTS. Geprüft wird der übergebene " +
				"Text — nicht die installierte Datei —, damit eine Änderung vor dem " +
				"Schreiben zu prüfen ist. Ohne Befund ist »Gefunden« falsch.",
			Schema: objekt(map[string]any{
				"inhalt": text("Der Text der Paketdatei — vollständig, nicht nur ein Teil."),
			}, "inhalt"),
			Ruf: func(_ context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Inhalt string `json:"inhalt"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				zeile, spalte, meldung := werkstatt.Pruefen(argumente.Inhalt)
				if zeile == 0 {
					return alsJSON(map[string]any{"Gefunden": false})
				}

				return alsJSON(map[string]any{
					"Gefunden": true,
					"Zeile":    zeile,
					"Spalte":   spalte,
					"Meldung":  meldung,
				})
			},
		},
		{
			Name: "schreiben.paketdatei", Art: Schreibend,
			Namensfelder: []string{"datei"},
			Erklaerung: "Schreibt eine Paketdatei zurück. ÄNDERT DIE VORLAGE FÜR ALLE " +
				"SETS, die dieses Paket benutzen. Der STAND muss der aus " +
				"»lesen.paketdatei« sein; hat inzwischen jemand anderes geschrieben, " +
				"wird abgewiesen statt überschrieben. Der vorige Stand bleibt als " +
				"».bak« neben dem Paket liegen. Zurückzunehmen ist die Änderung, " +
				"indem der vorige Text wieder geschrieben wird.",
			Schema: objekt(map[string]any{
				"paket":  text("Der Name des Pakets aus »lesen.pakete«."),
				"datei":  text("Der Name der Datei aus »lesen.paketdateien«."),
				"inhalt": text("Der neue Text der Datei — vollständig, nicht nur der Teil."),
				"stand":  text("Der Stand aus »lesen.paketdatei«."),
			}, "paket", "datei", "inhalt", "stand"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Paket  string `json:"paket"`
					Datei  string `json:"datei"`
					Inhalt string `json:"inhalt"`
					Stand  string `json:"stand"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				stand, fehler := werkstatt.Schreiben(ctx,
					argumente.Paket, argumente.Datei, argumente.Inhalt, argumente.Stand)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"Stand": stand})
			},
		},
		dateianlage(werkstatt),
		dateiumbenennung(werkstatt),
		dateientfernung(werkstatt),
	}
}

// leerStattNil gibt eine leere Liste statt nil — im JSON stünde sonst `null`.
func leerStattNil(liste []string) []string {
	if liste == nil {
		return []string{}
	}

	return liste
}

// geteilt trennt eine Werkzeugliste nach ihrer Art.
//
// **Nach `Art` und nicht nach der Reihenfolge**: die Trennung ist die Zusage
// aus dem Namen (`lesen.…` gegen `schreiben.…`), und wer sie aus einer
// Position ableitete, verlöre sie beim ersten Umsortieren.
func geteilt(alle []Werkzeug) (lesend, schreibend []Werkzeug) {
	for _, werkzeug := range alle {
		if werkzeug.Art == Lesend {
			lesend = append(lesend, werkzeug)

			continue
		}

		schreibend = append(schreibend, werkzeug)
	}

	return lesend, schreibend
}
