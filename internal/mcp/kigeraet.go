package mcp

import (
	"context"
	"encoding/json"
)

// kigeraetsitzungwerkzeug schaltet Apples Gerätemodell für diesen Server ab
// oder wieder an — AI-05.AK-04
// (#2042).
//
// **Sitzung heißt hier: so lange dieser Server läuft.** Nach einem Neustart
// gilt wieder die Einstellung `ki.geraetemodell`.
func kigeraetsitzungwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.geraetemodell.sitzung", Art: Schreibend,
		Erklaerung: "Schaltet Apples Gerätemodell für DIESEN Server ab oder wieder an. " +
			"Ab heißt sofort: ein laufendes »fm serve« wird beendet, und bis zum " +
			"Neustart des Servers fragt »vorrang« es nicht. ÄNDERT den Bestand NICHT — " +
			"die Einstellung »ki.geraetemodell« bleibt, wie sie ist. " + vorbehalt,
		Schema: objekt(map[string]any{
			"an": jaNein("true: in dieser Sitzung fragen; false: nicht fragen."),
		}, "an"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wunsch struct {
				An bool `json:"an"`
			}

			if err := lies(roh, &wunsch); err != nil {
				return "", err
			}

			an, fehler := werkstatt.KIGeraetemodellFuerSitzung(ctx, wunsch.An)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"an": an})
		},
	}
}
