// **Das Zufallspack** —
// #889,
// Befund 7 des Zweitblicks über Phase W.
//
// # Wofür ein Sprachmodell es braucht
//
// Um zu **sehen, was ein Set als Booster hergibt** — und um dieselbe Ziehung
// herzustellen, über die jemand im Fenster spricht. Der Samen ist die Zusage:
// dieselbe Zahl, dieselben Karten. Ohne ihn wäre jede Antwort einmalig und
// über keine ließe sich reden.
//
// **Und um zu sehen, wann ein Set noch nichts hergibt**: eine Packart mit null
// erreichbaren Karten zieht nichts. Ohne die Zahl sähe das aus wie ein Fehler.

package mcp

import (
	"context"
	"encoding/json"
)

// Pack ist, was die Werkzeuge dieser Datei brauchen.
type Pack interface {
	Packarten(ctx context.Context, setID int64) (Packangebot, error)
	PackZiehen(ctx context.Context, setID int64, art, same string) (Packziehung, error)
}

// Packangebot ist die Antwort auf »was lässt sich hier ziehen?«.
type Packangebot struct {
	Arten   []Packart `json:"Arten"`
	Befunde []string  `json:"Befunde"`
}

// Packart ist eine Art samt der Zahl erreichbarer Karten.
type Packart struct {
	// Name steht **unübersetzt** da — so heißt er im Paket (§9).
	Name     string `json:"Name"`
	Waehlbar bool   `json:"Waehlbar"`
	Karten   int    `json:"Karten"`
}

// Packziehung sind die gezogenen Karten in ihrer Ziehreihenfolge.
type Packziehung struct {
	Karten  []Packkarte `json:"Karten"`
	Befunde []string    `json:"Befunde"`
}

// Packkarte ist eine gezogene Karte.
type Packkarte struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

// packwerkzeuge gibt die beiden lesenden.
//
// **Beide lesend.** Ziehen ändert nichts am Bestand — es rechnet aus einem
// Samen eine Auswahl. Es als schreibend zu führen hieße, ein Modell ohne Not
// vor eine Rückfrage zu stellen.
func packwerkzeuge(pack Pack) []Werkzeug {
	return []Werkzeug{lesenPackarten(pack), lesenPackZiehen(pack)}
}

// lesenPackarten gibt die Packarten eines Sets.
func lesenPackarten(pack Pack) Werkzeug {
	return Werkzeug{
		Name: "lesen.packarten", Art: Lesend,
		Erklaerung: "Gibt die Packarten, die das Spiel hinter einem Set kennt. " +
			"ÄNDERT NICHTS. »Waehlbar« heißt: MSE bietet sie zum Ziehen an — die " +
			"übrigen sind Bausteine, aus denen die wählbaren bestehen. »Karten« " +
			"ist die Zahl der Karten des Sets, die diese Art ERREICHEN kann; bei " +
			"null zieht sie nichts, und das ist kein Fehler, sondern ein Set ohne " +
			"passende Karten. »Befunde« nennt Arten, deren Filter sich nicht " +
			"ausrechnen ließ.",
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

			angebot, fehler := pack.Packarten(ctx, argumente.SetID)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen und kein `null`** — [liste].
			return alsJSON(Packangebot{
				Arten: liste(angebot.Arten), Befunde: liste(angebot.Befunde),
			})
		},
	}
}

// lesenPackZiehen zieht ein Pack.
func lesenPackZiehen(pack Pack) Werkzeug {
	return Werkzeug{
		Name: "lesen.pack.ziehen", Art: Lesend,
		Erklaerung: "Zieht ein Pack aus einem Set und gibt die Karten in ihrer " +
			"Ziehreihenfolge. ÄNDERT NICHTS am Set. »art« ist ein Name aus " +
			"»lesen.packarten«. »same« ist der Samen ALS TEXT: dieselbe Zahl gibt " +
			"dieselbe Ziehung, eine andere eine andere. Als Text, weil eine Zahl " +
			"über 2^53 in JSON ihre letzten Stellen verlöre. Leer heißt null — " +
			"auch das ist ein gültiger Samen.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets."),
			"art":   text("Der Name der Packart aus »lesen.packarten«."),
			"same":  text("Der Samen als Text, etwa »12345«. Leer heißt null."),
		}, "setID", "art"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID int64  `json:"setID"`
				Art   string `json:"art"`
				Same  string `json:"same"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			ziehung, fehler := pack.PackZiehen(ctx, argumente.SetID, argumente.Art, argumente.Same)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(Packziehung{
				Karten: liste(ziehung.Karten), Befunde: liste(ziehung.Befunde),
			})
		},
	}
}
