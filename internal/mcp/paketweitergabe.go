// **Pakete als `.mse-installer` weitergeben und einlesen** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **erweiternd** — zwei Werkzeuge kommen hinzu.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// paketweitergabe gibt die zwei Werkzeuge rund um `.mse-installer`.
func paketweitergabe(pflege Paketpflege) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.paket.weitergeben", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: schreibt Vorlagenpakete als .mse-installer — " +
				"die Datei, mit der MSE2 Vorlagen weitergibt. Die Pakete stehen mit " +
				"Endung da, wie »lesen.pakete« sie nennt, etwa »probe.mse-game«; " +
				"Abhängigkeiten kommen NICHT von selbst mit. Ein Ordner als Ziel bekommt " +
				"den Namen des ersten Pakets; eine vorhandene Datei wird ersetzt.",
			Schema: objekt(map[string]any{
				"pakete": textliste("Die Pakete, in dieser Reihenfolge.", "Ein Paket mit Endung, etwa »probe.mse-game«."),
				"ziel":   text("Ein Ordner oder eine Datei auf .mse-installer."),
			}, "pakete", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Pakete []string `json:"pakete"`
					Ziel   string   `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				datei, err := pflege.PaketWeitergeben(ctx, argumente.Pakete, argumente.Ziel)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]string{"datei": datei})
			},
		},
		{
			Name: "schreiben.paket.einlesen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Rechner: legt die Pakete eines .mse-installer ins " +
				"verwaltete Paketverzeichnis. Was schon eingebunden ist, wird " +
				"ÜBERGANGEN und nicht überschrieben; ein Paket, das sich nicht lesen " +
				"lässt, bleibt draußen und steht unter den Befunden.",
			Schema: objekt(map[string]any{
				"datei": text("Der Installer, etwa »/Users/…/probe.mse-installer«."),
			}, "datei"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Datei string `json:"datei"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Der Teilerfolg reist mit dem Fehler** — wie bei
				// »schreiben.paket.uebernehmen«: die Pakete liegen schon da.
				uebernahme, err := pflege.PaketEinlesen(ctx, argumente.Datei)
				if err != nil {
					return "", fmt.Errorf("%w — eingelesen wurden bereits %s",
						err, namenVon(uebernahme.Uebernommen))
				}

				return alsJSON(uebernahme)
			},
		},
	}
}
