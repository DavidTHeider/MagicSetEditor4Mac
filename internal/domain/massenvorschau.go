package domain

// **Die Vorschau einer Massenänderung** — Werkstück U6 von Phase 5.1
// (#1417).
//
// # Warum es sie gibt
//
// „Setze `rarity` auf 12 Karten" schrieb bis hierher sofort. Wer sehen
// wollte, was das für jede Karte heißt — welche schon so steht, welche das
// Feld gar nicht hat —, musste es hinterher im Verlauf nachlesen. Die
// Vorschau nennt es **vorher**: je Karte Name, Alt und Neu, und einen
// Ausschluss, wo nichts zu tun ist.
//
// # Und warum sie eine Erwartung hergibt
//
// Zwischen Vorschau und Anwenden kann ein anderer schreiben — der
// MCP-Server, ein zweites Fenster. Die Vorschau gibt deshalb je Karte den
// Altwert als [Feldvorwert] mit; wer anwendet, reicht ihn zurück, und der
// Bestand vergleicht, bevor er schreibt. Weicht eine Karte ab, ist das ein
// [ErrConflict] und keine stille Überschreibung.
//
// # Herkunft
//
// **new.** MSE2 kennt keine Massenänderung mit Vorschau: `CardsPanel::
// doReplaceAll` in `src/gui/set/cards_panel.cpp` (Zeile 453, Fassung
// f9d9356 vom 2021-07-06) ist `return false; // TODO`, und der
// Ersetzen-Dialog arbeitet Karte für Karte über den Editor. Gesucht wurde
// dort, in `src/data/action/value.cpp` (die Wertaktionen, die ein Undo
// tragen) und in `src/util/action_stack.hpp`.

// AusschlussGleich sagt: die Karte steht schon auf dem Wert.
const AusschlussGleich = "gleich"

// Feldvorwert ist der Altwert einer Karte, wie ihn die Vorschau gezeigt hat.
//
// **`GabEs` unterscheidet „war leer" von „gab es nicht"** — dieselbe
// Unterscheidung wie im Verlauf ([Verlaufseintrag.GabVorher]): eine leere
// Zeile ist nicht dasselbe wie keine Zeile.
type Feldvorwert struct {
	KarteID int64  `json:"karteID"`
	Wert    string `json:"wert"`
	GabEs   bool   `json:"gabEs"`
}

// Massenvorschauzeile ist, was die Vorschau über eine Karte sagt.
type Massenvorschauzeile struct {
	KarteID int64 `json:"karteID"`

	// Name ist der Wert des Feldes `name` — leer, wenn die Karte keinen hat;
	// die Oberfläche zeigt dann die Kennung.
	Name string `json:"name"`

	Alt    string `json:"alt"`
	GabAlt bool   `json:"gabAlt"`
	Neu    string `json:"neu"`

	// Ausschluss sagt, warum diese Karte **nicht** geschrieben wird — leer,
	// wenn sie geschrieben wird.
	Ausschluss string `json:"ausschluss,omitempty"`
}

// Massenvorschau ist die Vorschau für ein Feld und einen Wert über Karten.
type Massenvorschau struct {
	Feld   string                `json:"feld"`
	Wert   string                `json:"wert"`
	Zeilen []Massenvorschauzeile `json:"zeilen"`
}

// Vorschauzeile sagt, was die Änderung für **eine** Karte hieße.
func Vorschauzeile(karte *CardRow, feld, wert string) Massenvorschauzeile {
	name, _ := karte.Value("name")
	alt, gabAlt := karte.Value(feld)

	zeile := Massenvorschauzeile{
		KarteID: karte.ID, Name: name, Alt: alt, GabAlt: gabAlt, Neu: wert,
	}

	if gabAlt && alt == wert {
		zeile.Ausschluss = AusschlussGleich
	}

	return zeile
}

// Erwartung gibt die Altwerte der Karten, die geschrieben würden — die
// ausgeschlossenen bleiben weg, denn für sie wird nichts geschrieben.
func (v Massenvorschau) Erwartung() []Feldvorwert {
	erwartet := make([]Feldvorwert, 0, len(v.Zeilen))

	for _, zeile := range v.Zeilen {
		if zeile.Ausschluss != "" {
			continue
		}

		erwartet = append(erwartet, Feldvorwert{
			KarteID: zeile.KarteID, Wert: zeile.Alt, GabEs: zeile.GabAlt,
		})
	}

	return erwartet
}
