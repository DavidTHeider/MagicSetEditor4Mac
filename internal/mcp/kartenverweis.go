package mcp

// **Die Verweiswerkzeuge** — Werkstück 2 von Phase 3.6
// (#1167).
//
// # Warum ein Sprachmodell sie braucht
//
// Ohne Verweis holt jeder Abgleich **irgendeinen** Druck: `/cards/named` wählt
// selbst aus. Ein Modell, das „nimm die Alpha-Fassung" hört, hätte keine
// Möglichkeit, das festzuhalten — beim nächsten Abgleich stünde wieder der
// neueste Druck da, und niemand sähe, warum.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verweisangabe ist ein Verweis, wie er über das Protokoll geht.
type Verweisangabe struct {
	KarteID  string `json:"karteID"`
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette wie jede Kennung, leer für keine; Vorlaeufig sagt, dass
	// der Druck keinen Alias hat. Beides **erweiternd**: alte Aufrufer lesen
	// weiter `printing`.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
	Adresse string `json:"adresse"`
	Gesetzt string `json:"gesetzt"`
}

// Fremdverweise ist, was der MCP-Server von den Verweisen braucht.
type Fremdverweise interface {
	VerweisSetzen(ctx context.Context, karteID int64, printing string) (Verweisangabe, error)
	Verweis(ctx context.Context, karteID int64) (*Verweisangabe, error)
	VerweisLoesen(ctx context.Context, karteID int64) error
}

// verweiswerkzeuge gibt die drei — oder keines ohne Naht.
//
// **Keines und nicht drei scheiternde** (#591): ein Werkzeug, das dasteht und
// bei jedem Ruf scheitert, ist schlimmer als keines — ein Modell versucht es
// dreimal.
func verweiswerkzeuge(verweise Fremdverweise) (lesend, schreibend []Werkzeug) {
	if verweise == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.verweis", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt den Druck, auf den eine Karte verweist. ÄNDERT NICHTS. " +
			"Ohne Verweis kommt »verweis: null« — das ist der Normalfall und " +
			"kein Fehler. Mit Verweis holt »lesen.scryfall.vorschlag« genau " +
			"diesen Druck statt irgendeines.",
		Schema: objekt(map[string]any{
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
		}, "karteID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			karteID, err := verweiskarte(roh)
			if err != nil {
				return "", err
			}

			verweis, lesefehler := verweise.Verweis(ctx, karteID)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{"verweis": verweis})
		},
	}}, []Werkzeug{{
		Name: "schreiben.verweis", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: bezieht eine Karte auf einen bestimmten " +
			"Druck bei Scryfall. printing ist die KENNUNG aus »lesen.scryfall.suche«. " +
			"Der Druck wird geholt, bevor der Verweis entsteht — eine Kennung, die " +
			"es nicht gibt, wird abgewiesen. Danach holt »lesen.scryfall.vorschlag« " +
			"genau diesen Druck. Zurückgenommen wird er mit " +
			"»schreiben.verweisloesen«.",
		Schema: objekt(map[string]any{
			"karteID":  zahl("Die Kennung der Karte aus »lesen.karten«."),
			"printing": text("Die Scryfall-Kennung des Drucks aus »lesen.scryfall.suche«."),
		}, "karteID", "printing"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				KarteID  int64  `json:"karteID"`
				Printing string `json:"printing"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
				return "", err
			}

			if strings.TrimSpace(argumente.Printing) == "" {
				return "", fmt.Errorf(
					"%w: es wurde kein Druck genannt", domain.ErrValidation)
			}

			verweis, fehler := verweise.VerweisSetzen(ctx,
				argumente.KarteID, argumente.Printing)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"verweis": verweis})
		},
	}, {
		Name: "schreiben.verweisloesen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: nimmt den Verweis einer Karte zurück. " +
			"Eine Karte ohne Verweis wird abgewiesen — so ist zu sehen, dass die " +
			"richtige Karte gemeint war. Die Feldwerte bleiben, wie sie sind.",
		Schema: objekt(map[string]any{
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
		}, "karteID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			karteID, err := verweiskarte(roh)
			if err != nil {
				return "", err
			}

			if fehler := verweise.VerweisLoesen(ctx, karteID); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"karteID": karteID, "geloest": true})
		},
	}}
}

// verweiskarte liest und prüft die Kartenkennung eines Rufs.
//
// **Eine Stelle für zwei Werkzeuge**: zwei gleichlautende Zweige wären zwei,
// von denen nur einer je einen Test sähe (§3).
func verweiskarte(roh json.RawMessage) (int64, error) {
	var argumente struct {
		KarteID int64 `json:"karteID"`
	}

	if err := lies(roh, &argumente); err != nil {
		return 0, err
	}

	if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
		return 0, err
	}

	return argumente.KarteID, nil
}
