package mcp

// **Die Spalten der Kartenliste und ihre Ordnung über MCP** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// # Wozu ein Modell das braucht
//
// **Weil `lesen.karten` alle Felder gibt und keine Ordnung.** Wer eine Liste
// zeigen oder eine Reihenfolge prüfen will, fragt hier: welche Spalten führt
// dieses Spiel, welche sind sichtbar, wonach wird sortiert.

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdliste ist, was der MCP-Server von der Kartenliste braucht.
//
// **Über einfache Werte und nicht über `service`-Typen**: `internal/mcp` darf
// nur `domain` kennen (§2), und der Schichtenwächter zählt nach.
type Fremdliste interface {
	Listenaufbau(ctx context.Context, setID int64) (domain.Kartenlistenaufbau, error)

	ListenspaltenSetzen(ctx context.Context, setID int64, wahl []domain.Spaltenwahl) error
	ListensortierungSetzen(ctx context.Context, setID int64, ordnung domain.Kartenordnung) error

	// KartenFiltern gibt die Kennungen der Karten, die durch den Schnellfilter
	// kommen — Werkstück U5 von Phase 5.1 (#1416).
	KartenFiltern(ctx context.Context, setID int64, anfrage string) ([]int64, error)
}

// listenwerkzeuge gibt die drei — oder keines ohne Naht.
func listenwerkzeuge(liste Fremdliste) (lesend, schreibend []Werkzeug) {
	if liste == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.kartenliste", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt die Spalten der Kartenliste eines Sets, ihre " +
			"Reihenfolge, welche sichtbar sind, die Zeilen in der " +
			"eingestellten Ordnung — und wonach sortiert wird. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, err := setkennung(roh)
			if err != nil {
				return "", err
			}

			aufbau, fehler := liste.Listenaufbau(ctx, setID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"spalten":     spaltenliste(aufbau.Spalten),
				"zeilen":      zeilenliste(aufbau.Zeilen),
				"sortierfeld": aufbau.Ordnung.Feld,
				"aufsteigend": aufbau.Ordnung.Aufsteigend,
				"ohneVorlage": aufbau.OhneVorlage,
			})
		},
	}, kartenfilterwerkzeug(liste)}, []Werkzeug{{
		Name: "schreiben.kartenliste.spalten", Art: Schreibend,
		Erklaerung: "Setzt, welche Spalten die Kartenliste zeigt und in " +
			"welcher Reihenfolge. Die REIHENFOLGE DER LISTE ist die " +
			"Reihenfolge der Spalten. Gilt für ALLE Sets desselben Spiels.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"spalten": map[string]any{
				"type":        "array",
				"description": "Die Spalten in ihrer Reihenfolge.",
				"items": objekt(map[string]any{
					"feld":     text("Der Feldname aus »lesen.kartenliste«."),
					"sichtbar": jaNein("Ob die Spalte zu sehen ist."),
				}, "feld", "sichtbar"),
			},
		}, "setID", "spalten"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, wahl, err := spaltenwahl(roh)
			if err != nil {
				return "", err
			}

			if fehler := liste.ListenspaltenSetzen(ctx, setID, wahl); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"spalten": len(wahl)})
		},
	}, {
		Name: "schreiben.kartenliste.sortierung", Art: Schreibend,
		Erklaerung: "Setzt, wonach die Kartenliste ordnet. Ein LEERES Feld " +
			"heißt: gar nicht sortieren — dann gilt die Reihenfolge des " +
			"Sets. Gilt für ALLE Sets desselben Spiels.",
		Schema: objekt(map[string]any{
			"setID":       zahl("Die Kennung des Sets aus »lesen.sets«."),
			"feld":        text("Der Feldname; leer heißt: gar nicht sortieren."),
			"aufsteigend": jaNein("Aufsteigend ordnen."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID       int64  `json:"setID"`
				Feld        string `json:"feld"`
				Aufsteigend bool   `json:"aufsteigend"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			fehler := liste.ListensortierungSetzen(ctx, argumente.SetID,
				domain.Kartenordnung{
					Feld: argumente.Feld, Aufsteigend: argumente.Aufsteigend,
				})
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"sortierfeld": argumente.Feld, "aufsteigend": argumente.Aufsteigend,
			})
		},
	}}
}

// setkennung liest die eine Kennung, die ein Ruf trägt.
func setkennung(roh json.RawMessage) (int64, error) {
	var argumente struct {
		SetID int64 `json:"setID"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, err
	}

	return argumente.SetID, pruefeKennung("setID", argumente.SetID)
}

// spaltenwahl liest Set und Spaltenliste eines Schreibrufs.
func spaltenwahl(roh json.RawMessage) (int64, []domain.Spaltenwahl, error) {
	var argumente struct {
		SetID   int64 `json:"setID"`
		Spalten []struct {
			Feld     string `json:"feld"`
			Sichtbar bool   `json:"sichtbar"`
		} `json:"spalten"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, nil, err
	}

	if err := pruefeKennung("setID", argumente.SetID); err != nil {
		return 0, nil, err
	}

	wahl := make([]domain.Spaltenwahl, 0, len(argumente.Spalten))

	// **Die Stelle kommt aus der Reihenfolge der Liste** — dieselbe Regel wie
	// am Fenster, damit beide Flächen dasselbe meinen.
	for stelle, eine := range argumente.Spalten {
		wahl = append(wahl, domain.Spaltenwahl{
			Feld: eine.Feld, Sichtbar: eine.Sichtbar, Stelle: stelle + 1,
		})
	}

	return argumente.SetID, wahl, nil
}

// spaltenliste macht aus den Spalten eine Liste — **immer eine**.
func spaltenliste(spalten []domain.Listenspalte) []map[string]any {
	liste := make([]map[string]any, 0, len(spalten))

	for _, spalte := range spalten {
		liste = append(liste, map[string]any{
			"feld":       spalte.Feld,
			"aufschrift": spalte.Aufschrift,
			"sichtbar":   spalte.Sichtbar,
		})
	}

	return liste
}

// zeilenliste macht aus den Zeilen eine Liste — ebenso.
func zeilenliste(zeilen []domain.Listenzeile) []map[string]any {
	liste := make([]map[string]any, 0, len(zeilen))

	for _, zeile := range zeilen {
		liste = append(liste, map[string]any{
			"karteID": zeile.KarteID,
			"nummer":  zeile.Nummer + 1,
			"werte":   zeile.Werte,
		})
	}

	return liste
}

// kartenfilterwerkzeug legt den Schnellfilter über die Karten eines Sets —
// Werkstück U5 von Phase 5.1 (#1416). Die Sprache ist die von MSE2
// (`src/data/filter.cpp`, Fassung f9d9356): Leerraum trennt, `"…"` hält
// zusammen, `feld:wert` bindet, ein führendes `-` verneint.
func kartenfilterwerkzeug(liste Fremdliste) Werkzeug {
	return Werkzeug{
		Name: "lesen.kartenfilter", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt die Kennungen der Karten eines Sets, die durch den " +
			"Schnellfilter kommen. Leerraum trennt Teile, »…« hält zusammen, " +
			"»feld:wert« bindet an Felder, ein führendes »-« verneint; " +
			"verglichen wird als Teilwort ohne Groß und Klein. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets aus »lesen.sets«."),
			"anfrage": text("Der Filter, etwa »type:creature -goblin«."),
		}, "setID", "anfrage"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				SetID   int64  `json:"setID"`
				Anfrage string `json:"anfrage"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if wahl.SetID <= 0 {
				return "", fmt.Errorf("%w: setID muss eine Kennung größer null sein", domain.ErrValidation)
			}

			kennungen, fehler := liste.KartenFiltern(ctx, wahl.SetID, wahl.Anfrage)
			if fehler != nil {
				return "", fehler
			}

			ids := make([]string, 0, len(kennungen))
			for _, id := range kennungen {
				ids = append(ids, strconv.FormatInt(id, 10))
			}

			return alsJSON(map[string]any{"kartenIDs": ids})
		},
	}
}
