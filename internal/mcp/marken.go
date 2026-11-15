package mcp

// **Die Markenwerkzeuge** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Warum ein Sprachmodell sie braucht
//
// Eine Marke ist eine Aussage über die **Arbeit** an einer Karte — „fertig",
// „prüfen", „noch kein bild". Wer ein Modell bittet, ein Set durchzusehen,
// will das Ergebnis irgendwo stehen haben; ohne Marken bliebe ihm nur, es in
// eine Antwort zu schreiben, die niemand wiederfindet.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdmarken ist, was der MCP-Server von den Marken braucht.
type Fremdmarken interface {
	MarkeSetzen(ctx context.Context, setID, karteID int64, marke string) error
	MarkeEntfernen(ctx context.Context, setID, karteID int64, marke string) error
	KartenMarken(ctx context.Context, setID, karteID int64) ([]string, error)
	Setmarken(ctx context.Context, setID int64) ([]string, error)
}

// markenwerkzeuge gibt die vier — oder keines ohne Naht.
//
// **Keines und nicht vier scheiternde** (#591): ein Werkzeug, das dasteht und
// bei jedem Ruf scheitert, ist schlimmer als keines.
func markenwerkzeuge(marken Fremdmarken) (lesend, schreibend []Werkzeug) {
	if marken == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.marken", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Marken, die in einem Set vorkommen — jede einmal. " +
			"ÄNDERT NICHTS. Eine Marke ist ein Wort, das jemand an eine Karte " +
			"gehängt hat: »fertig«, »prüfen«, »noch kein bild«. Sie steht neben " +
			"der Karte und nicht in ihr; eine weitergegebene .mse-set trägt sie " +
			"nicht mit.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID int64 `json:"setID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			gefunden, fehler := marken.Setmarken(ctx, argumente.SetID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"marken": listeOderLeer(gefunden)})
		},
	}, {
		Name: "lesen.kartenmarken", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Marken einer Karte. ÄNDERT NICHTS. Eine Karte ohne " +
			"Marken gibt eine leere Liste — das ist der Normalfall und kein Fehler.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
		}, "setID", "karteID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, karteID, err := markenstelle(roh)
			if err != nil {
				return "", err
			}

			gefunden, fehler := marken.KartenMarken(ctx, setID, karteID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"marken": listeOderLeer(gefunden)})
		},
	}}, []Werkzeug{{
		Name: "schreiben.marke", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: hängt eine Marke an eine Karte. Die Marke " +
			"wird getrimmt und kleingeschrieben — »Fertig« und »fertig« sind " +
			"dieselbe. Zweimal dieselbe zu setzen ist kein Fehler. Die Feldwerte " +
			"der Karte bleiben unberührt.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"marke":   text("Das Wort, das an die Karte soll."),
		}, "setID", "karteID", "marke"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, karteID, marke, err := markenruf(roh)
			if err != nil {
				return "", err
			}

			if fehler := marken.MarkeSetzen(ctx, setID, karteID, marke); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"karteID": karteID, "marke": domain.Markennormalform(marke), "gesetzt": true,
			})
		},
	}, {
		Name: "schreiben.markeentfernen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt eine Marke von einer Karte ab. Eine " +
			"Marke, die gar nicht dasteht, wird abgewiesen — so ist zu sehen, dass " +
			"die richtige gemeint war.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"marke":   text("Das Wort, das weg soll."),
		}, "setID", "karteID", "marke"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, karteID, marke, err := markenruf(roh)
			if err != nil {
				return "", err
			}

			if fehler := marken.MarkeEntfernen(ctx, setID, karteID, marke); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"karteID": karteID, "marke": domain.Markennormalform(marke),
				"entfernt": true,
			})
		},
	}}
}

// markenstelle liest Set- und Kartenkennung eines Rufs.
func markenstelle(roh json.RawMessage) (setID, karteID int64, fehler error) {
	var argumente struct {
		SetID   int64 `json:"setID"`
		KarteID int64 `json:"karteID"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, 0, err
	}

	if err := pruefeKennung("setID", argumente.SetID); err != nil {
		return 0, 0, err
	}

	if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
		return 0, 0, err
	}

	return argumente.SetID, argumente.KarteID, nil
}

// markenruf liest Kennungen **und** Marke.
//
// **Die Marke wird hier nicht geprüft** — das tut `service.Markenpflege`, und
// zwar für alle vier Flächen. Eine zweite Prüfung daneben wäre eine zweite
// Wahrheit (§16).
func markenruf(roh json.RawMessage) (setID, karteID int64, marke string, fehler error) {
	setID, karteID, fehler = markenstelle(roh)
	if fehler != nil {
		return 0, 0, "", fehler
	}

	var argumente struct {
		Marke string `json:"marke"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, 0, "", err
	}

	return setID, karteID, argumente.Marke, nil
}

// listeOderLeer gibt eine leere Liste statt `null`.
func listeOderLeer(werte []string) []string {
	if werte == nil {
		return []string{}
	}

	return werte
}
