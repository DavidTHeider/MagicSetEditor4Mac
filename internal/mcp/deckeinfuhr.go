// **Eine Deckliste über MCP in ein Deck legen** — Werkstück `T8` von Phase
// 5.6 (#1473).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Zwei Werkzeuge, weil es zwei Schritte sind
//
// `lesen.deckliste` schreibt **nichts** — es sagt, was ins Deck käme, und
// nennt den **Stand**, an den es gebunden ist. `schreiben.deckliste` legt,
// was ihm ausdrücklich genannt wird, und verlangt denselben Stand zurück.
//
// # Warum der Stand hier nicht wahlfrei ist
//
// **Weil niemand hinsieht.** An der Kommandozeile darf `-` die Bindung
// aufheben: wer tippt, sieht seinen Aufruf. Ein Sprachmodell ruft, während
// niemand hinsieht ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)) —
// es muss den Stand aus der Vorschau mitbringen, oder es legt nichts.
//
// # Und warum es keinen Besitz anfasst
//
// **Ein Deck verweist auf Karten; es besitzt keine.** Wer sie in die
// Sammlung tragen will, nimmt `schreiben.liste.einlesen` — ein anderes
// Werkzeug mit einer anderen Wirkung, und die Erklärung sagt es.

package mcp

import (
	"context"
	"encoding/json"
)

// Deckeinfuhr ist, was MCP für eine Deckliste braucht.
type Deckeinfuhr interface {
	Vorschau(ctx context.Context, deckID int64, text string) (Deckvorschau, error)
	Uebernehmen(ctx context.Context, deckID int64, stand string,
		zeilen []Deckeinfuhrzeile) (Deckeinfuhrbericht, error)
}

// Deckvorschauzeile ist eine Zeile, wie sie ins Deck käme.
type Deckvorschauzeile struct {
	Bereich string `json:"bereich"`
	Menge   int    `json:"menge"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Printing ist der Druck — leer, solange die Zeile nicht eindeutig ist.
	Printing string `json:"printing"`

	Treffer []Listentreffer `json:"treffer"`
	Drucke  int             `json:"drucke"`

	Befund       string `json:"befund"`
	Uebernehmbar bool   `json:"uebernehmbar"`
}

// Deckvorschau ist, was aus einer Liste würde.
type Deckvorschau struct {
	DeckID    int64               `json:"deckID"`
	Stand     string              `json:"stand"`
	Zeilen    []Deckvorschauzeile `json:"zeilen"`
	Befunde   []string            `json:"befunde"`
	Hauptdeck int                 `json:"hauptdeck"`
	Sideboard int                 `json:"sideboard"`
}

// Deckeinfuhrzeile ist eine Zeile, für die entschieden ist.
type Deckeinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Bereich  string `json:"bereich"`
}

// Deckeinfuhrbericht sagt, was gelegt wurde.
type Deckeinfuhrbericht struct {
	Gelegt int    `json:"gelegt"`
	Stueck int    `json:"stueck"`
	Stand  string `json:"stand"`
}

// deckeinfuhr gibt die beiden Werkzeuge — lesend zuerst.
func deckeinfuhr(einfuhr Deckeinfuhr) []Werkzeug {
	if einfuhr == nil {
		return nil
	}

	return []Werkzeug{
		decklistewerkzeug(einfuhr), decklegewerkzeug(einfuhr),
	}
}

// decklistewerkzeug gibt das lesende.
func decklistewerkzeug(einfuhr Deckeinfuhr) Werkzeug {
	return Werkzeug{
		Name: "lesen.deckliste", Art: Lesend,
		Erklaerung: "Liest eine Deckliste (»4 Lightning Bolt (M11) 149«, eine " +
			"Zeile je Karte) für ein Deck und sagt, was hineinkäme. SCHREIBT " +
			"NICHTS. Überschriften gliedern: unter »Sideboard« steht das " +
			"Sideboard, unter »Maybeboard«, »Tokens« und »About« steht nichts, " +
			"was ins Deck gehört. »uebernehmbar« sagt, ob eine Zeile ohne " +
			"Rückfrage gelegt werden kann; mehrdeutige Zeilen bekommen KEINEN " +
			"geratenen Druck. »stand« ist der Abdruck des Decks — er gehört in " +
			"»schreiben.deckliste«.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks aus »lesen.decks«."),
			"text": text("Die Liste, eine Karte je Zeile. Leerzeilen und " +
				"Zeilen mit # oder // werden übergangen."),
		}, "deckID", "text"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				DeckID int64  `json:"deckID"`
				Text   string `json:"text"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			vorschau, err := einfuhr.Vorschau(ctx, argumente.DeckID,
				argumente.Text)
			if err != nil {
				return "", err
			}

			return alsJSON(vorschau)
		},
	}
}

// decklegewerkzeug gibt das schreibende.
func decklegewerkzeug(einfuhr Deckeinfuhr) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deckliste", Art: Schreibend,
		Erklaerung: "ÄNDERT ein Deck: legt die genannten Zeilen hinein. Jede " +
			"Zeile nennt ihre Druckkennung — die aus »lesen.deckliste«. »stand« " +
			"ist PFLICHT und stammt aus derselben Vorschau: hat sich das Deck " +
			"seitdem geändert, wird NICHTS gelegt. Es entsteht dabei KEIN " +
			"Besitz — ein Deck verweist auf Karten; wer sie in die Sammlung " +
			"tragen will, nimmt »schreiben.liste.einlesen«. Es wird nichts " +
			"herausgenommen: die Zeilen kommen HINZU.",
		Schema: objekt(map[string]any{
			"deckID": zahl("Die Kennung des Decks aus »lesen.decks«."),
			"stand":  text("Der Abdruck aus »lesen.deckliste«."),
			"zeilen": objektliste("Die Zeilen mit entschiedenem Druck.", objekt(
				map[string]any{
					"printing": text("Die Scryfall-Druckkennung aus »lesen.deckliste«."),
					"menge":    zahl("Die Stückzahl — mindestens 1."),
					"bereich":  text("»hauptdeck« oder »sideboard«."),
				}, "printing", "menge", "bereich")),
		}, "deckID", "stand", "zeilen"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				DeckID int64              `json:"deckID"`
				Stand  string             `json:"stand"`
				Zeilen []Deckeinfuhrzeile `json:"zeilen"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			bericht, err := einfuhr.Uebernehmen(ctx, argumente.DeckID,
				argumente.Stand, argumente.Zeilen)
			if err != nil {
				return "", err
			}

			return alsJSON(bericht)
		},
	}
}
