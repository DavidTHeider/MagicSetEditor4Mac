// **Die Schlüsselwörter eines Sets** —
// #1062.
//
// # Wofür ein Sprachmodell sie braucht
//
// Um zu **wissen**, welche das Spiel schon mitbringt: `magic.mse-game` führt
// 334. Ohne diese Auskunft schriebe ein Modell ein zweites »Fliegend« ins Set
// — und zwei Schlüsselwörter, die auf denselben Text passen, ergeben zwei
// Erinnerungstexte auf derselben Karte.

package mcp

import (
	"context"
	"encoding/json"
)

// Schluesselwoerter ist, was die Werkzeuge dieser Datei brauchen.
type Schluesselwoerter interface {
	Schluesselwoerter(ctx context.Context, setID int64) (Schluesselwortliste, error)
	SchluesselwortAnlegen(ctx context.Context, setID int64) (int64, error)
	SchluesselwortSetzen(ctx context.Context, setID, wortID int64, feld, wert string) error
	SchluesselwortLoeschen(ctx context.Context, setID, wortID int64) error
}

// Schluesselwortliste ist die Antwort auf »welche gibt es hier?«.
//
// **Die Marken sind die Zusage** — siehe [Satz].
type Schluesselwortliste struct {
	Woerter []Schluesselwort `json:"Woerter"`
	Modi    []string         `json:"Modi"`
}

// Schluesselwort ist eines des Sets oder seines Spiels.
type Schluesselwort struct {
	// ID ist **0 bei einem des Spiels**: das gehört keinem Set.
	ID int64 `json:"ID"`

	Wort       string `json:"Wort"`
	Muster     string `json:"Muster"`
	Erinnerung string `json:"Erinnerung"`
	Regeln     string `json:"Regeln"`
	Modus      string `json:"Modus"`

	Fest    bool `json:"Fest"`
	Gueltig bool `json:"Gueltig"`
}

// schluesselwortwerkzeuge gibt das lesende und die drei schreibenden.
func schluesselwortwerkzeuge(woerter Schluesselwoerter) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{lesenSchluesselwoerter(woerter)}, []Werkzeug{
		schreibenSchluesselwortAnlegen(woerter),
		schreibenSchluesselwortSetzen(woerter),
		schreibenSchluesselwortLoeschen(woerter),
	}
}

// lesenSchluesselwoerter gibt die des Sets und die des Spiels.
func lesenSchluesselwoerter(woerter Schluesselwoerter) Werkzeug {
	return Werkzeug{
		Name: "lesen.schluesselwoerter", Art: Lesend,
		Erklaerung: "Gibt die Schlüsselwörter eines Sets UND die seines Spiels. " +
			"ÄNDERT NICHTS. Ein Schlüsselwort erkennt seinen Text auf jeder Karte " +
			"und setzt den Erinnerungstext dahinter. »Fest« heißt: aus dem Spiel, " +
			"unveränderlich, ohne Kennung. »Gueltig« heißt: sein Muster greift — " +
			"ein leeres erkennt nichts. Im Muster steht " +
			"»<atom-param>art</atom-param>« an der Stelle eines Parameters.",
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

			gefunden, fehler := woerter.Schluesselwoerter(ctx, argumente.SetID)
			if fehler != nil {
				return "", fehler
			}

			// **Leere Listen und kein `null`** — [liste].
			return alsJSON(Schluesselwortliste{
				Woerter: liste(gefunden.Woerter), Modi: liste(gefunden.Modi),
			})
		},
	}
}

// schreibenSchluesselwortAnlegen hängt ein leeres ans Ende.
func schreibenSchluesselwortAnlegen(woerter Schluesselwoerter) Werkzeug {
	return Werkzeug{
		Name: "schreiben.schluesselwort.anlegen", Art: Schreibend,
		Erklaerung: "Legt ein leeres Schlüsselwort im Set an und gibt seine Kennung " +
			"zurück. ÄNDERT das Set: es steht danach in der Datei. Leer in jedem " +
			"Feld außer der Verwendung; gefüllt wird es mit " +
			"»schreiben.schluesselwort.setzen«. NICHT IDEMPOTENT — ein zweiter " +
			"Aufruf legt ein zweites an.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets."),
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

			neu, fehler := woerter.SchluesselwortAnlegen(ctx, argumente.SetID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"WortID": neu})
		},
	}
}

// schreibenSchluesselwortSetzen schreibt ein Feld.
func schreibenSchluesselwortSetzen(woerter Schluesselwoerter) Werkzeug {
	return Werkzeug{
		Name: "schreiben.schluesselwort.setzen", Art: Schreibend,
		Erklaerung: "Schreibt ein Feld eines Schlüsselworts des SETS. ÄNDERT es und " +
			"ÜBERSCHREIBT den bisherigen Wert. Die des Spiels tragen keine Kennung " +
			"und lassen sich nicht ändern.",
		Schema: objekt(map[string]any{
			"setID":  zahl("Die Kennung des Sets."),
			"wortID": zahl("Die Kennung des Schlüsselworts aus »lesen.schluesselwoerter«."),
			"feld": text("Das Feld: »wort«, »muster«, »erinnerung«, »regeln« " +
				"oder »modus«."),
			"wert": text("Der neue Wert. Leer nimmt die Zeile aus der Datei."),
		}, "setID", "wortID", "feld", "wert"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID  int64  `json:"setID"`
				WortID int64  `json:"wortID"`
				Feld   string `json:"feld"`
				Wert   string `json:"wert"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := beideSchluesselwortkennungen(argumente.SetID, argumente.WortID); err != nil {
				return "", err
			}

			if fehler := woerter.SchluesselwortSetzen(ctx, argumente.SetID,
				argumente.WortID, argumente.Feld, argumente.Wert); fehler != nil {
				return "", fehler
			}

			return "Das Feld " + argumente.Feld + " steht.", nil
		},
	}
}

// schreibenSchluesselwortLoeschen nimmt eines weg.
func schreibenSchluesselwortLoeschen(woerter Schluesselwoerter) Werkzeug {
	return Werkzeug{
		Name: "schreiben.schluesselwort.loeschen", Art: Schreibend,
		Erklaerung: "Löscht ein Schlüsselwort des SETS. ÄNDERT das Set: es steht " +
			"danach nicht mehr in der Datei. Die des Spiels lassen sich nicht " +
			"löschen. Umkehrbar über »schreiben.zuruecknehmen« — es kommt an " +
			"seine Stelle zurück, samt allen Feldern.",
		Schema: objekt(map[string]any{
			"setID":  zahl("Die Kennung des Sets."),
			"wortID": zahl("Die Kennung des Schlüsselworts."),
		}, "setID", "wortID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID  int64 `json:"setID"`
				WortID int64 `json:"wortID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := beideSchluesselwortkennungen(argumente.SetID, argumente.WortID); err != nil {
				return "", err
			}

			if fehler := woerter.SchluesselwortLoeschen(
				ctx, argumente.SetID, argumente.WortID); fehler != nil {
				return "", fehler
			}

			return "Das Schlüsselwort ist weg.", nil
		},
	}
}

// beideSchluesselwortkennungen prüft Set und Schlüsselwort.
func beideSchluesselwortkennungen(setID, wortID int64) error {
	if err := pruefeKennung("setID", setID); err != nil {
		return err
	}

	return pruefeKennung("wortID", wortID)
}
