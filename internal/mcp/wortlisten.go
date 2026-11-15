// **Die Auswahllisten eines Kartenfeldes** —
// #1074.
//
// # Wofür ein Sprachmodell sie braucht
//
// Um zu **wissen**, welche Untertypen ein Spiel kennt, statt sie zu erfinden.
// `magic.mse-game` führt 51 Listen; ohne sie schreibt ein Modell `Beast` und
// `Beasts` und `beast` in drei Karten desselben Sets.

package mcp

import (
	"context"
	"encoding/json"
)

// Wortlisten ist, was die Werkzeuge dieser Datei brauchen.
type Wortlisten interface {
	Wortlisten(ctx context.Context, karteID int64, feld string) ([]Wortbereich, error)

	Wortwahl(
		ctx context.Context, setID, karteID int64, feld string, bereich int, weg []int,
	) error
}

// Wortbereich ist eine Stelle eines Feldes mit ihrem Menü.
//
// **Die Marken sind die Zusage** — siehe [Satz]: ohne sie hinge der Name eines
// Feldes im Protokoll an Gos Vorgabe, und ein Umbenennen hier änderte still das
// Protokoll.
type Wortbereich struct {
	Liste     string        `json:"Liste"`
	Text      string        `json:"Text"`
	Kaestchen bool          `json:"Kaestchen"`
	Gewaehlt  int           `json:"Gewaehlt"`
	Eintraege []Worteintrag `json:"Eintraege"`
}

// Worteintrag ist eine Zeile im Menü.
type Worteintrag struct {
	Name          string        `json:"Name"`
	IstPraefix    bool          `json:"IstPraefix"`
	ZeileDarunter bool          `json:"ZeileDarunter"`
	Aktiv         bool          `json:"Aktiv"`
	Kinder        []Worteintrag `json:"Kinder"`
}

// wortlistenwerkzeuge gibt das lesende und das schreibende Werkzeug.
func wortlistenwerkzeuge(listen Wortlisten) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{lesenWortlisten(listen)}, []Werkzeug{schreibenWortwahl(listen)}
}

// lesenWortlisten gibt die Auswahl eines Feldes.
func lesenWortlisten(listen Wortlisten) Werkzeug {
	return Werkzeug{
		Name: "lesen.wortlisten", Art: Lesend,
		Erklaerung: "Gibt die Auswahllisten eines Kartenfeldes — die Wörter, die das " +
			"Spiel für diese Stelle kennt. ÄNDERT NICHTS. Ein Feld kann mehrere " +
			"Bereiche haben (Obertyp und Untertyp stehen im selben Feld); der " +
			"Bereich wird über seine Nummer angesprochen, beginnend bei 0. " +
			"»IstPraefix« heißt: das Wort wird vorangestellt und lässt sich mit " +
			"anderen verbinden, statt sie zu ersetzen. Eine leere Liste heißt, " +
			"dass dieses Feld keine Auswahl hat.",
		Schema: objekt(map[string]any{
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"feld":    text("Der Feldname, etwa »type« oder »sub type«."),
		}, "karteID", "feld"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				KarteID int64  `json:"karteID"`
				Feld    string `json:"feld"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
				return "", err
			}

			bereiche, fehler := listen.Wortlisten(ctx, argumente.KarteID, argumente.Feld)
			if fehler != nil {
				return "", fehler
			}

			// **Eine leere Liste und kein `null`** — [liste].
			return alsJSON(liste(bereiche))
		},
	}
}

// schreibenWortwahl wählt einen Eintrag und schreibt das Feld.
//
// **Der Weg ist eine Liste, keine Zahl.** Eine Gruppe öffnet ein Untermenü —
// im Pack tun das 103 Einträge —, und `[2, 0]` heißt: der erste Eintrag im
// Untermenü des dritten. Ohne diese Form bliebe die Hälfte der Wörter
// unerreichbar.
func schreibenWortwahl(listen Wortlisten) Werkzeug {
	return Werkzeug{
		Name: "schreiben.wortwahl", Art: Schreibend,
		Erklaerung: "Wählt einen Eintrag aus einer Auswahlliste und schreibt das Feld. " +
			"ÄNDERT die Karte: ÜBERSCHREIBT den Bereich, auf den sich die Wahl " +
			"bezieht — der Rest des " +
			"Feldes bleibt stehen. Ein Präfix wird UMGESCHALTET: steht es schon da, " +
			"nimmt dieselbe Wahl es wieder weg. Die Nummern kommen aus " +
			"»lesen.wortlisten« und gelten nur für den Stand, aus dem sie stammen.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets, zu dem die Karte gehört."),
			"karteID": zahl("Die Kennung der Karte."),
			"feld":    text("Der Feldname."),
			"bereich": zahl("Der wievielte Bereich des Feldes, beginnend bei 0."),
			"weg": zahlenliste(
				"Der Weg zum Eintrag: [3] ist der vierte Eintrag des Menüs, "+
					"[2, 0] der erste im Untermenü des dritten.",
				"Eine Eintragsnummer, beginnend bei 0.",
			),
		}, "setID", "karteID", "feld", "bereich", "weg"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID   int64  `json:"setID"`
				KarteID int64  `json:"karteID"`
				Feld    string `json:"feld"`
				Bereich int    `json:"bereich"`
				Weg     []int  `json:"weg"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
				return "", err
			}

			if fehler := listen.Wortwahl(ctx, argumente.SetID, argumente.KarteID,
				argumente.Feld, argumente.Bereich, argumente.Weg); fehler != nil {
				return "", fehler
			}

			return "Der Eintrag steht im Feld " + argumente.Feld + ".", nil
		},
	}
}
