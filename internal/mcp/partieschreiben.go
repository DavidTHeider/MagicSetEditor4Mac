// **Die drei schreibenden Partiewerkzeuge** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Herausgelöst aus `partie.go`**, damit beide Dateien unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort, **was** ein
// Modell sehen darf; hier, **was** es ändern darf.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// partienschreibwerkzeuge baut die drei.
func partienschreibwerkzeuge(quelle Partienwerkstatt) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.partie.anlegen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt eine NEUE Partie an. Sie " +
				"überschreibt keine. Jeder Sitz bringt ein Deck mit, und es " +
				"wird EINGEFROREN — wer sein Deck danach ändert, ändert " +
				"diese Partie nicht.",
			Schema: objekt(map[string]any{
				"spiel": text("Das Spielpaket, etwa »magic«."),
				"wirt": text("Die Sitzkennung dessen, der serialisiert — " +
					"leer heißt: der erste Sitz."),
				"sitze": objektliste("Mindestens zwei Sitze.", objekt(map[string]any{
					"kennung": text("Die Sitzkennung."),
					"name":    text("Was auf dem Tisch steht."),
					"deckID":  text("Die Kennung des Decks."),
					"commander": text("Wahlfrei: die Kartenkennung des Commanders " +
						"aus dem Deck — CR 903.3 (Fassung 2026-08-19). Leer heißt: keiner."),
					"profilID": text("Wahlfrei: die Kennung eines Profils " +
						"DESSELBEN Nutzers, an das der Sitz gebunden wird — " +
						"dann ist er in »lesen.partien« der eigene Sitz " +
						"(eigenerSitz, dran). Leer heißt: keines. Ein Profil, " +
						"das es nicht gibt, nennt die Antwort in »befunde«; " +
						"die Partie entsteht trotzdem."),
				}, "kennung", "name", "deckID")),
				"abgewaehlt": textliste("Wahlfrei: die Regelbereiche, die die "+
					"Runde SELBST übernimmt — sie werden mit der Partie "+
					"EINGEFROREN. Leer heißt: das Programm trägt, was es "+
					"tragen kann. Welche es gibt, nennt »lesen.regelfreigabe« "+
					"im Baum; dort lässt sich die Wahl vorher prüfen.",
					"Ein Regelbereich, etwa »zugstruktur«."),
			}, "spiel", "sitze"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Spiel      string       `json:"spiel"`
					Wirt       string       `json:"wirt"`
					Sitze      []Sitzwunsch `json:"sitze"`
					Abgewaehlt []string     `json:"abgewaehlt"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if argumente.Spiel == "" || len(argumente.Sitze) < 2 {
					return "", fmt.Errorf(
						"%w: es braucht ein Spiel und mindestens zwei Sitze",
						domain.ErrValidation)
				}

				angaben, err := quelle.PartieAnlegen(ctx, argumente.Spiel,
					argumente.Wirt, argumente.Sitze, argumente.Abgewaehlt)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
		{
			Name: "schreiben.partie.befehl", Art: Schreibend,
			Erklaerung: "ÄNDERT die Partie: hängt ein Ereignis an ihren " +
				"Strom. Es überschreibt nichts. Die Befehlskennung kommt " +
				"von DIR — mit derselben zweimal geschieht NICHTS ein " +
				"zweites Mal.",
			Schema: objekt(map[string]any{
				"partieID": text("Die Kennung der Partie."),
				"absender": text("Die Sitzkennung dessen, der den Befehl stellt."),
				"kennung": text("Die Kennung dieses Befehls — sie schützt " +
					"vor der doppelten Wirkung."),
				"art": text("Die Art, etwa »objekt.tappen« — was gerade " +
					"geht, sagt »erlaubt« aus »lesen.partie«."),
				"nutzlast": objekt(map[string]any{}),
			}, "partieID", "absender", "kennung", "art"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var befehl Spielbefehl

				if err := lies(roh, &befehl); err != nil {
					return "", err
				}

				if befehl.PartieID == "" || befehl.Absender == "" ||
					befehl.Kennung == "" || befehl.Art == "" {
					return "", fmt.Errorf(
						"%w: es braucht Partie, Absender, Befehlskennung "+
							"und Art", domain.ErrValidation)
				}

				angaben, err := quelle.PartieBefehl(ctx, befehl)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
		{
			Name: "schreiben.partie.stand", Art: Schreibend,
			Erklaerung: "ÄNDERT die Partie: hält sie an, nimmt sie wieder " +
				"auf oder beendet sie. EIN BEENDEN IST ENDGÜLTIG — eine " +
				"beendete Partie lässt sich nicht fortsetzen.",
			Schema: objekt(map[string]any{
				"partieID": text("Die Kennung der Partie."),
				"fuer":     text("Die Sitzkennung für die Antwort."),
				"stand": text("Eines von »pausieren«, »fortsetzen«, " +
					"»beenden«."),
			}, "partieID", "fuer", "stand"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					PartieID string `json:"partieID"`
					Fuer     string `json:"fuer"`
					Stand    string `json:"stand"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if argumente.PartieID == "" || argumente.Fuer == "" {
					return "", fmt.Errorf(
						"%w: es braucht die Partie und eine Sitzkennung",
						domain.ErrValidation)
				}

				angaben, err := quelle.PartieStand(
					ctx, argumente.PartieID, argumente.Fuer, argumente.Stand)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
	}
}
