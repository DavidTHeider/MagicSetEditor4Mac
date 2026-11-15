// **Die Setangaben, die Vorbehalte und die mitgelieferten Sets als Werkzeuge**
// (#1107).
//
// # Warum ein Modell sie braucht
//
// Ein Set hat Titel, Setcode und Copyright, und ein Modell, das gebeten wird
// „bau mir ein Set", konnte sie bis hierher nicht setzen — nur Karten
// anlegen. Die Vorbehalte sind die Gegenprobe zum Einfahren: sie sagen, was
// von einer fremden Datei nicht gedeutet wurde, und ohne sie hielte ein Modell
// ein halb gelesenes Set für ein ganzes (§8).

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Setangabenquelle ist, was die vier Werkzeuge brauchen.
type Setangabenquelle interface {
	Angaben(ctx context.Context, setID int64) (
		felder []Angabenfeld, werte []Angabenwert, fehler error)
	Setzen(ctx context.Context, setID int64, feld, wert string) (
		vorher string, hatte bool, fehler error)
	Vorbehalte(ctx context.Context, setID int64) ([]Vorbehalt, error)
	Beispielsets() ([]Beispielset, error)
}

// Angabenfeld ist ein Angabenfeld, wie der Server es nennt.
//
// **Eine eigene Gestalt und nicht `app.FeldbeschreibungDTO`**: `internal/mcp`
// ist ein Blatt und kennt `app` nicht (§2).
type Angabenfeld struct {
	Name       string `json:"name"`
	Aufschrift string `json:"aufschrift"`
	Art        string `json:"art"`
}

// Angabenwert ist eine gesetzte Angabe.
type Angabenwert struct {
	Feld string `json:"feld"`
	Wert string `json:"wert"`
}

// Vorbehalt ist ein Block, den das Einfahren nicht gedeutet hat.
type Vorbehalt struct {
	Datei  string `json:"datei"`
	Zeile  int    `json:"zeile"`
	Inhalt string `json:"inhalt"`
}

// Beispielset ist ein mitgeliefertes Set.
type Beispielset struct {
	Name string `json:"name"`
	Pfad string `json:"pfad"`
}

// setangabenwerkzeuge gibt die drei lesenden und das eine schreibende.
func setangabenwerkzeuge(quelle Setangabenquelle) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{
			angabenLesen(quelle), vorbehalteLesen(quelle), beispielsetsLesen(quelle),
		},
		[]Werkzeug{angabeSetzen(quelle)}
}

// angabenLesen nennt die Angabenfelder eines Sets samt ihren Werten.
func angabenLesen(quelle Setangabenquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.setangaben", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Angabenfelder eines Sets — Titel, Setcode, Copyright — samt Werten.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			felder, werte, lesefehler := quelle.Angaben(ctx, wahl)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{
				"felder": liste(felder), "werte": liste(werte),
			})
		},
	}
}

// angabeSetzen schreibt eine Angabe und nennt die vorige.
func angabeSetzen(quelle Setangabenquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.setangabe", Art: Schreibend,
		Erklaerung: "ÄNDERT eine Angabe eines Sets — Titel, Setcode, Copyright. Gibt den vorigen Wert zurück, damit sich die Änderung zurücknehmen lässt.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"feld":  text("Der Feldname aus »lesen.setangaben«."),
			"wert":  text("Der neue Wert. Leer löscht ihn nicht, sondern setzt ihn leer."),
		}, "setID", "feld", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				SetID int64  `json:"setID"`
				Feld  string `json:"feld"`
				Wert  string `json:"wert"`
			}

			if err := json.Unmarshal(argumente, &wahl); err != nil {
				return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
			}

			// **Die Kennung wird geprüft, bevor der Bestand sie sieht** (§8).
			// Sonst liefe eine Null bis zum Schreiber und käme als „nicht
			// gefunden" zurück — richtig abgewiesen, aber am falschen Ort.
			if err := pruefeKennung("setID", wahl.SetID); err != nil {
				return "", err
			}

			vorher, hatte, fehler := quelle.Setzen(ctx, wahl.SetID, wahl.Feld, wahl.Wert)
			if fehler != nil {
				return "", fehler
			}

			// **Der vorige Wert reist mit**, damit ein Modell zurückkann —
			// dieselbe Zusage wie beim Kartenfeld
			// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
			return alsJSON(map[string]any{"vorher": vorher, "hatte": hatte})
		},
	}
}

// vorbehalteLesen gibt, was beim Einfahren nicht gedeutet wurde.
func vorbehalteLesen(quelle Setangabenquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.vorbehalte", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Blöcke, die beim Einfahren eines Sets nicht gedeutet wurden.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			bloecke, lesefehler := quelle.Vorbehalte(ctx, wahl)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{"vorbehalte": liste(bloecke)})
		},
	}
}

// beispielsetsLesen nennt die mitgelieferten Sets samt Pfad.
//
// **Ohne Argument und ohne Bestand.** Ein Modell, das ein Set aufschlagen
// soll, kennt sonst keinen einzigen Pfad — und die mitgelieferten liegen im
// Bündel, nicht im Dateisystem der Nutzenden.
func beispielsetsLesen(quelle Setangabenquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.beispielsets", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die mitgelieferten Sets samt Pfad — für »schreiben.set.einfahren«.",
		Schema:     objekt(map[string]any{}),
		Ruf: func(_ context.Context, _ json.RawMessage) (string, error) {
			sets, fehler := quelle.Beispielsets()
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"sets": liste(sets)})
		},
	}
}
