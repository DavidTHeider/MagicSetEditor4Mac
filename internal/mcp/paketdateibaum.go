// **Der Dateibaum eines Pakets** — FMT-02.AK-01
// (#1691).
//
// Einordnung nach §16: **erweiternd** — drei Werkzeuge kommen hinzu, und
// `lesen.paketdateien` nennt zusätzlich das Beiwerk.
//
// **Einfügen von der Platte gibt es hier nicht**, und das ist eine
// Entscheidung (`docs/FAEHIGKEITEN.md`): ein Modell könnte damit eine
// beliebige Datei des Rechners in ein Paket legen und über
// »lesen.paketdatei« wieder hinauslesen.

package mcp

import (
	"context"
	"encoding/json"
)

// dateianlage legt eine neue Vorlage an.
func dateianlage(werkstatt Paketwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.paketdatei.anlegen", Art: Schreibend,
		Namensfelder: []string{"datei"},
		Erklaerung: "Legt eine NEUE Textdatei in einem Paket an — eine Vorlage wie " +
			"»card_fields« oder ein eigenes Skript. ÄNDERT DAS PAKET FÜR ALLE SETS, die " +
			"es benutzen. Ein Name, unter dem schon etwas steht, wird abgewiesen und " +
			"nichts überschrieben; Ordner im Namen werden angelegt. Nur an einem " +
			"entpackten Paket — ein gepacktes zuerst mit »schreiben.paket.kopieren« " +
			"kopieren. Gibt den STAND für »schreiben.paketdatei« zurück.",
		Schema: objekt(map[string]any{
			"paket":  text("Der Name des Pakets aus »lesen.pakete«."),
			"datei":  text("Der Name der neuen Datei, auch mit Ordner, etwa »skripte/farben«."),
			"inhalt": text("Der Text der neuen Datei."),
		}, "paket", "datei", "inhalt"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Paket  string `json:"paket"`
				Datei  string `json:"datei"`
				Inhalt string `json:"inhalt"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, fehler := werkstatt.Anlegen(ctx, argumente.Paket, argumente.Datei, argumente.Inhalt)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Stand": stand})
		},
	}
}

// dateiumbenennung gibt einer Datei einen neuen Namen.
func dateiumbenennung(werkstatt Paketwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.paketdatei.umbenennen", Art: Schreibend,
		Namensfelder: []string{"datei", "neu"},
		Erklaerung: "Gibt einer Datei eines Pakets einen neuen Namen, auch in einem " +
			"anderen Ordner. ÄNDERT DAS PAKET FÜR ALLE SETS, die es benutzen — ein " +
			"Verweis auf den alten Namen findet danach nichts mehr. Die Hauptdatei " +
			"(»style«, »game« …) lässt sich nicht umbenennen; ein belegter neuer Name " +
			"wird abgewiesen. Mit STAND wird geprüft, dass niemand inzwischen " +
			"geschrieben hat. Zurück geht es mit demselben Werkzeug in die andere Richtung.",
		Schema: objekt(map[string]any{
			"paket": text("Der Name des Pakets aus »lesen.pakete«."),
			"datei": text("Der bisherige Name aus »lesen.paketdateien«."),
			"neu":   text("Der neue Name, auch mit Ordner."),
			"stand": text("Der Stand aus »lesen.paketdatei« — Pflicht bei einer Textdatei, leer bei Beiwerk."),
		}, "paket", "datei", "neu"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Paket string `json:"paket"`
				Datei string `json:"datei"`
				Neu   string `json:"neu"`
				Stand string `json:"stand"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if fehler := werkstatt.Umbenennen(ctx,
				argumente.Paket, argumente.Datei, argumente.Neu, argumente.Stand); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Datei": argumente.Neu})
		},
	}
}

// dateientfernung nimmt eine Datei aus dem Paket.
func dateientfernung(werkstatt Paketwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.paketdatei.entfernen", Art: Schreibend,
		Namensfelder: []string{"datei"},
		Erklaerung: "Nimmt eine Datei aus einem Paket. ÄNDERT DAS PAKET FÜR ALLE SETS, " +
			"die es benutzen. Die Datei liegt danach in der Sicherung »<paket>.bak« " +
			"neben dem Paket und lässt sich von dort zurückholen. Die Hauptdatei " +
			"lässt sich nicht entfernen. Mit STAND wird geprüft, dass niemand " +
			"inzwischen geschrieben hat.",
		Schema: objekt(map[string]any{
			"paket": text("Der Name des Pakets aus »lesen.pakete«."),
			"datei": text("Der Name der Datei aus »lesen.paketdateien«."),
			"stand": text("Der Stand aus »lesen.paketdatei« — Pflicht bei einer Textdatei, leer bei Beiwerk."),
		}, "paket", "datei"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Paket string `json:"paket"`
				Datei string `json:"datei"`
				Stand string `json:"stand"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if fehler := werkstatt.Entfernen(ctx, argumente.Paket, argumente.Datei, argumente.Stand); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Entfernt": argumente.Datei})
		},
	}
}
