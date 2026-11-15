package mcp

// **Karten kopieren und einfügen über MCP** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # Ohne Fenster ist die Zwischenablage der Text selbst
//
// Was das Fenster in die Zwischenablage legt, ist eine Zeichenkette in MSEs
// eigenem Format. Ein Sprachmodell bekommt sie hier zurück und gibt sie beim
// Einfügen wieder hinein — damit trägt der Weg auch über eine Grenze, an der
// es gar keine Zwischenablage gibt.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdablage ist, was der MCP-Server vom Kopieren und Einfügen braucht.
type Fremdablage interface {
	AblageKopieren(ctx context.Context, setID int64, karteIDs []int64) (string, error)
	AblageAusschneiden(ctx context.Context, setID int64, karteIDs []int64) (string, error)
	AblageEinfuegen(ctx context.Context, setID int64, inhalt string) (
		karten, befunde []string, fehler error)
}

// ablagewerkzeuge gibt die drei — oder keines ohne Naht.
func ablagewerkzeuge(ablage Fremdablage) (lesend, schreibend []Werkzeug) {
	if ablage == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.ablage.kopieren", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt die genannten Karten als Text im Format von Magic " +
			"Set Editor zurück. ÄNDERT NICHTS. Der Text geht unverändert in " +
			"»schreiben.ablage.einfuegen« — auch in ein anderes Set, solange " +
			"es dasselbe Spiel führt.",
		Schema: objekt(map[string]any{
			"setID":    zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteIDs": zahlenliste("Die Karten aus »lesen.karten«.", "Eine Kartenkennung."),
		}, "setID", "karteIDs"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, karteIDs, err := ablagekarten(roh)
			if err != nil {
				return "", err
			}

			inhalt, kopierfehler := ablage.AblageKopieren(ctx, setID, karteIDs)
			if kopierfehler != nil {
				return "", kopierfehler
			}

			return alsJSON(map[string]any{"inhalt": inhalt})
		},
	}}, []Werkzeug{{
		Name: "schreiben.ablage.ausschneiden", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: gibt die Karten als Text zurück und " +
			"ENTFERNT sie danach aus dem Set. Kopiert wird zuerst — scheitert " +
			"das Entfernen, steht der Text trotzdem da und die Karte auch.",
		Schema: objekt(map[string]any{
			"setID":    zahl("Die Kennung des Sets aus »lesen.sets«."),
			"karteIDs": zahlenliste("Die Karten aus »lesen.karten«.", "Eine Kartenkennung."),
		}, "setID", "karteIDs"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, karteIDs, err := ablagekarten(roh)
			if err != nil {
				return "", err
			}

			inhalt, schnittfehler := ablage.AblageAusschneiden(ctx, setID, karteIDs)
			if schnittfehler != nil {
				return "", schnittfehler
			}

			return alsJSON(map[string]any{"inhalt": inhalt})
		},
	}, {
		Name: "schreiben.ablage.einfuegen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt die Karten aus einem Text im Set " +
			"an. Der Text kommt aus »lesen.ablage.kopieren«. Führt das Set ein " +
			"anderes Spiel, wird abgewiesen und KEINE Karte angelegt — die " +
			"Felder eines fremden Spiels gibt es hier nicht.",
		Schema: objekt(map[string]any{
			"setID":  zahl("Die Kennung des Zielsets aus »lesen.sets«."),
			"inhalt": text("Der Text aus »lesen.ablage.kopieren«."),
		}, "setID", "inhalt"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID  int64  `json:"setID"`
				Inhalt string `json:"inhalt"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if strings.TrimSpace(argumente.Inhalt) == "" {
				return "", fmt.Errorf(
					"%w: es wurde kein Inhalt genannt", domain.ErrValidation)
			}

			karten, befunde, fehler := ablage.AblageEinfuegen(
				ctx, argumente.SetID, argumente.Inhalt)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"karten": karten, "befunde": befunde})
		},
	}}
}

// ablagekarten liest Set und Kartenliste eines Rufs.
//
// **Eine Stelle für zwei Werkzeuge** — zwei gleichlautende Zweige wären zwei,
// von denen nur einer je einen Test sähe (§3).
func ablagekarten(roh json.RawMessage) (setID int64, karteIDs []int64, fehler error) {
	var argumente struct {
		SetID    int64   `json:"setID"`
		KarteIDs []int64 `json:"karteIDs"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, nil, err
	}

	if err := pruefeKennung("setID", argumente.SetID); err != nil {
		return 0, nil, err
	}

	if len(argumente.KarteIDs) == 0 {
		return 0, nil, fmt.Errorf(
			"%w: es wurde keine Karte genannt", domain.ErrValidation)
	}

	for _, karteID := range argumente.KarteIDs {
		if err := pruefeKennung("karteIDs", karteID); err != nil {
			return 0, nil, err
		}
	}

	return argumente.SetID, argumente.KarteIDs, nil
}
