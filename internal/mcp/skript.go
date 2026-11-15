// **Ein MSE-Script-Ausdruck im Kontext einer Karte** —
// #1070.

package mcp

import (
	"context"
	"encoding/json"
)

// Skriptkonsole ist, was das Konsolenwerkzeug braucht.
type Skriptkonsole interface {
	Auswerten(ctx context.Context, karteID int64, ausdruck string) (Skriptergebnis, error)

	// Namen zählt auf, was ein Ausdruck an dieser Karte nennen darf
	// (#1164).
	Namen(ctx context.Context, karteID int64) ([]string, error)
}

// Skriptergebnis ist, was ein ausgewerteter Ausdruck zurückmeldet.
//
// **Die Marken sind die Zusage** — siehe [Satz].
type Skriptergebnis struct {
	Quelltext string   `json:"Quelltext"`
	Text      string   `json:"Text"`
	Art       string   `json:"Art"`
	Befunde   []string `json:"Befunde"`
}

// skriptwerkzeug gibt das lesende Werkzeug für die Skriptkonsole.
//
// # Warum es lesend ist
//
// **Es ändert nichts.** MSE-Script kann in diesem Programm nicht schreiben —
// `internal/schichten` weist jeden Zugriff auf das Dateisystem aus
// `msescript` ab —, und diese Naht legt nichts dazu. Ein Ausdruck rechnet und
// gibt zurück.
//
// # Wofür ein Sprachmodell es braucht
//
// Um zu **prüfen**, was eine Vorlage rechnet, bevor es ein Feld ändert:
// `card.name`, `length(card.rule_text)`, `to_upper(card.type)`. Ohne die
// Konsole müsste es raten, was ein Skript ergibt — und ein geratenes Ergebnis
// ist genau das, was §15 verhindern soll.
func skriptwerkzeug(konsole Skriptkonsole) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.skript", Art: Lesend,
			Erklaerung: "Wertet einen MSE-Script-Ausdruck im Kontext einer Karte aus und " +
				"gibt das Ergebnis zurück. ÄNDERT NICHTS. In der Umgebung stehen " +
				"dieselben Namen wie beim Zeichnen: »card« mit den Feldern der Karte, " +
				"»set« mit den Setangaben, »stylesheet« mit den Maßen der Vorlage, " +
				"dazu die Funktionen des Spielpakets. Das Ergebnis kommt als " +
				"QUELLTEXT — eine Zeichenkette trägt Anführungszeichen.",
			Schema: objekt(map[string]any{
				"karteID":  zahl("Die Kennung der Karte aus »lesen.karten«."),
				"ausdruck": text("Der Ausdruck, etwa »card.name« oder »1 + 1«."),
			}, "karteID", "ausdruck"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					KarteID  int64  `json:"karteID"`
					Ausdruck string `json:"ausdruck"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
					return "", err
				}

				ergebnis, fehler := konsole.Auswerten(ctx, argumente.KarteID, argumente.Ausdruck)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(ergebnis)
			},
		},
	}
}

// namenswerkzeug gibt das lesende Werkzeug für die Namen der Umgebung.
//
// # Warum es das über MCP gibt
//
// **Weil hier gerade nicht gilt, was sonst gilt.** Bei einer gewöhnlichen
// Sprache kennt ein Sprachmodell die Bibliothek aus seinem Trainingsstand;
// bei MSE-Script kennt es sie nicht — was zur Verfügung steht, hängt am
// **Paket auf diesem Rechner**, und an einer Karte mit eingebundener Vorlage
// sind es über sechzehnhundert Namen.
//
// Ohne diese Liste bleibt Raten, und ein geratener Name ist genau der Fehler,
// den §15 verhindern soll.
func namenswerkzeug(konsole Skriptkonsole) []Werkzeug {
	return []Werkzeug{
		{
			Name: "lesen.skriptnamen", Art: Lesend,
			Erklaerung: "Zählt auf, welche Namen ein MSE-Script-Ausdruck im Kontext " +
				"dieser Karte nennen darf — die Funktionen des Spielpakets und was " +
				"die Vorlage dazulegt. ÄNDERT NICHTS. Die Liste kommt aus derselben " +
				"Umgebung, in der »lesen.skript« auswertet; ein Name, der hier fehlt, " +
				"läuft dort nicht.",
			Schema: objekt(map[string]any{
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			}, "karteID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					KarteID int64 `json:"karteID"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
					return "", err
				}

				namen, fehler := konsole.Namen(ctx, argumente.KarteID)
				if fehler != nil {
					return "", fehler
				}

				// **Eine leere Liste und kein `null`** — dieselbe Überlegung
				// wie am Fenster: wer die Antwort liest, soll auf eines prüfen.
				if namen == nil {
					namen = []string{}
				}

				return alsJSON(map[string]any{"Namen": namen})
			},
		},
	}
}
