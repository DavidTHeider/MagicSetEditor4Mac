package mcp

// **Ein Modell messen** — Werkstück A9 von Phase 5.3
// (#1448).
//
// **Eine eigene Datei**, weil `ki.go` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: dort steht, was ein Modell **tut**,
// hier, was es **kann** — und wie man das misst.

import (
	"context"
	"encoding/json"
)

// kimodellpruefenwerkzeug gibt das Werkzeug für eine Messung.
//
// **Schreibend, obwohl es keine Karte anfasst.** Es kostet Geld und legt
// eine Aussage im Bestand ab; wer es als lesend führte, ließe es an einer
// Sperre vorbei, die genau dafür da ist.
func kimodellpruefenwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.modell.pruefen", Art: Schreibend,
		Erklaerung: "MISST, was ein Modell kann, und ÄNDERT dabei den Bestand — es legt die Messung dort ab.  ein kleiner Auftrag mit " +
			"festem Schema geht an den Anbieter, und kommt etwas " +
			"Schemakonformes zurück, gelten Textgenerierung und " +
			"strukturierte Ausgabe als BELEGT. " +
			"KOSTET GELD, wenn der Anbieter anderswo rechnet — der Ruf " +
			"braucht dieselbe Freigabe wie ein Vorschlag und scheitert ohne " +
			"sie. " +
			"WOZU: manche Anbieter nennen die Fähigkeiten ihrer Modelle " +
			"NICHT (OpenAI etwa), und ohne Beleg steht ein Modell nicht im " +
			"Katalog. Eine Messung ist der einzige Weg dorthin. " +
			"Sie belegt NICHT, dass das Modell Bilder liest oder einbettet — " +
			"und nicht, dass es morgen dasselbe tut. " + vorbehalt,
		Schema: objekt(map[string]any{
			"anbieter": text("Der Anbieter, wie »lesen.ki.modelle« ihn nennt."),
			"alias":    text("Das Modell, das gemessen werden soll."),
		}, "anbieter", "alias"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIProbenbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIModellPruefen(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen bleiben Listen** — ein `null` sähe für ein
			// Sprachmodell aus wie „konnte nicht beantwortet werden".
			if sicht.Proben == nil {
				sicht.Proben = []KIProbesicht{}
			}

			return alsJSON(sicht)
		},
	}
}

// kimodellvergessenwerkzeug nimmt die Messungen eines Modells zurück.
func kimodellvergessenwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.modell.vergessen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: es NIMMT die Messungen eines Modells ZURÜCK. Danach steht " +
			"über das Modell nur noch das fest, was sein Anbieter selbst " +
			"nennt — bei manchen ist das NICHTS. " +
			"KOSTET NICHTS und fragt NICHT im Netz. " + vorbehalt,
		Schema: objekt(map[string]any{
			"anbieter": text("Der Anbieter, wie »lesen.ki.modelle« ihn nennt."),
			"alias":    text("Das Modell, dessen Messungen fort sollen."),
		}, "anbieter", "alias"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIProbenbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			if fehler := werkstatt.KIModellVergessen(ctx, bezug); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"vergessen": true})
		},
	}
}
