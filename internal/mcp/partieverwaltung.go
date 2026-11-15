// **Die vier Verwaltungswerkzeuge einer Partie** — PLAY-03.AK-05
// (#1711).
//
// **Eine eigene Datei**, damit `partie.go` und `partieschreiben.go` unter
// der Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort, was
// ein Modell **sieht** und was es an einer laufenden Partie **ändert**; hier,
// was es mit der Partie als **Ganzem** tut.
//
// # Warum die Ausfuhr lesend ist und trotzdem alles zeigt
//
// **Weil sie den vollen Zustand gibt — auch fremde Hände.** `lesen.partie`
// kürzt auf die Sicht eines Sitzes; diese hier tut es ausdrücklich nicht,
// denn eine Lage ohne ihre verdeckten Teile ließe sich nicht nachstellen.
// Das steht in der Erklärung, damit ein Modell es weiß, bevor es ruft.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// partienverwaltungswerkzeuge baut die vier.
func partienverwaltungswerkzeuge(
	quelle Partienwerkstatt,
) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{ausfuhrwerkzeug(quelle)},
		[]Werkzeug{
			loeschwerkzeug(quelle),
			kopierwerkzeug(quelle),
			einfuhrwerkzeug(quelle),
		}
}

// ausfuhrwerkzeug gibt eine Partie als Text.
func ausfuhrwerkzeug(quelle Partienwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "lesen.partie.ausfuhr", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt eine Partie als Text, wie sie in einer Datei " +
			"steht. ÄNDERT NICHTS. ACHTUNG: darin steht der VOLLE Zustand " +
			"— auch fremde Hände und die Reihenfolge jeder Bibliothek. " +
			"Wer nur die Sicht eines Sitzes will, nimmt »lesen.partie«.",
		Schema: objekt(map[string]any{
			"partieID": text("Die Kennung der Partie."),
		}, "partieID"),
		Ruf: mitPartiekennung(quelle.PartieAusfuhr),
	}
}

// loeschwerkzeug nimmt eine Partie aus dem Bestand.
func loeschwerkzeug(quelle Partienwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.partie.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt eine Partie samt ihrem " +
			"Verlauf heraus. Geht NUR bei einem Entwurf — eine beendete " +
			"bleibt nachvollziehbar, und eine, für die jemand eine " +
			"Einwilligung gegeben hat, bleibt auch.",
		Schema: objekt(map[string]any{
			"partieID": text("Die Kennung der Partie."),
		}, "partieID"),
		Ruf: mitPartiekennung(quelle.PartieLoeschen),
	}
}

// kopierwerkzeug legt eine Partie als neuen Übungszustand daneben.
func kopierwerkzeug(quelle Partienwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.partie.kopieren", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine NEUE Partie mit der " +
			"Stellung einer vorhandenen an. Die Vorlage bleibt, wie sie " +
			"ist. Der Verlauf geht NICHT mit, und alte Einwilligungen " +
			"auch nicht — die Kopie läuft, auch wenn die Vorlage beendet ist.",
		Schema: objekt(map[string]any{
			"partieID": text("Die Kennung der Vorlage."),
		}, "partieID"),
		Ruf: mitPartiekennung(quelle.PartieKopieren),
	}
}

// einfuhrwerkzeug legt eine Partie aus einem Text an.
func einfuhrwerkzeug(quelle Partienwerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.partie.einfuhr", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt eine NEUE Partie aus dem " +
			"Text von »lesen.partie.ausfuhr« an. Sie bekommt eine NEUE " +
			"Kennung und fängt bei Fassung 1 an; eine fremde Partie wird " +
			"damit NICHT fortgesetzt.",
		Schema: objekt(map[string]any{
			"inhalt": text("Der Text aus »lesen.partie.ausfuhr«."),
		}, "inhalt"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Inhalt string `json:"inhalt"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if argumente.Inhalt == "" {
				return "", fmt.Errorf("%w: es braucht den Text der Partie",
					domain.ErrValidation)
			}

			angaben, err := quelle.PartieEinfuhr(ctx, argumente.Inhalt)
			if err != nil {
				return "", err
			}

			return alsJSON(angaben)
		},
	}
}

// mitPartiekennung baut den Ruf für die drei Wege, die nur eine Kennung
// brauchen.
//
// **Drei Werkzeuge und ein Ruf**: dieselbe Prüfung dreimal geschrieben
// hieße, sie dreimal ändern zu müssen — und beim dritten Mal wäre sie anders.
func mitPartiekennung(
	arbeit func(context.Context, string) (Partieverwaltungsangaben, error),
) func(context.Context, json.RawMessage) (string, error) {
	return func(ctx context.Context, roh json.RawMessage) (string, error) {
		var argumente struct {
			PartieID string `json:"partieID"`
		}

		if err := lies(roh, &argumente); err != nil {
			return "", err
		}

		if argumente.PartieID == "" {
			return "", fmt.Errorf("%w: es braucht die Kennung der Partie",
				domain.ErrValidation)
		}

		angaben, err := arbeit(ctx, argumente.PartieID)
		if err != nil {
			return "", err
		}

		return alsJSON(angaben)
	}
}
