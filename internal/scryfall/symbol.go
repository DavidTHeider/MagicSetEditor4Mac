package scryfall

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Symbol ist ein Kartensymbol, wie Scryfall es führt.
//
// # Wofür dieses Programm es braucht
//
// **Nicht zum Zeichnen.** Eine MSE-Vorlage bringt ihre eigene Symbolschrift
// mit, und die zu ersetzen wäre eine Verhaltensänderung (§15, Regel 10). Die
// Symbologie beantwortet eine andere Frage: **welche Symbole es gibt** und was
// sie bedeuten — für die Symbolpalette, für die Prüfung eines geschriebenen
// Manakostens und für den Alternativtext.
//
// Gelesen am 2026-08-24 unter `https://scryfall.com/docs/api/card-symbols`.
type Symbol struct {
	// Zeichen ist die Schreibweise mit Klammern — `{W}`, `{2/U}`, `{∞}`.
	//
	// **Nicht jedes ist ASCII**, das sagt die Dokumentation ausdrücklich. Wer
	// hier eine Prüfung auf Buchstaben einbaute, verlöre `{∞}`.
	Zeichen string

	// Beschreibung ist der englische Satz dazu — für den Alternativtext.
	//
	// **Er bleibt englisch** (§9): er kommt von außen, wie ein Feldname aus
	// einem MSE-Paket. Eine Übersetzung wäre eine zweite Quelle.
	Beschreibung string

	// IstMana sagt, ob es ein Manasymbol ist.
	IstMana bool

	// Manawert ist der Beitrag zu den Manakosten — `nil`, wenn keiner.
	//
	// **Ein Zeiger, weil 0 ein gültiger Wert ist**: `{0}` hat den Manawert
	// null, und `{T}` hat gar keinen. Wer beides als 0 führte, könnte sie nie
	// unterscheiden.
	Manawert *float64

	// InManakosten sagt, ob das Symbol in einem Manakosten vorkommt.
	//
	// **`{20}` steht auf false** — es kommt nur im Regeltext vor. Das ist der
	// Unterschied, an dem eine Prüfung von Manakosten hängt.
	InManakosten bool

	// Farben sind die Farben, die es darstellt.
	Farben []string

	// Hybrid und Phyrexianisch sagen, welcher Art es ist.
	Hybrid        bool
	Phyrexianisch bool
	Spassvariante bool
	Vektoradresse string
}

// rohsymbol ist die Gestalt, die Scryfall schickt.
type rohsymbol struct {
	Objekt        string   `json:"object"`
	Zeichen       string   `json:"symbol"`
	Beschreibung  string   `json:"english"`
	IstMana       bool     `json:"represents_mana"`
	Manawert      *float64 `json:"mana_value"`
	InManakosten  bool     `json:"appears_in_mana_costs"`
	Farben        []string `json:"colors"`
	Hybrid        bool     `json:"hybrid"`
	Phyrexianisch bool     `json:"phyrexian"`
	Spassvariante bool     `json:"funny"`
	Vektoradresse string   `json:"svg_uri"`
}

// LiesSymbole liest eine Antwort von `/symbology`.
//
// **`symbol` ist Pflicht** — es ist die Kennung. Alles Übrige darf fehlen:
// `mana_value` und `svg_uri` sind ausdrücklich als nullable dokumentiert.
func LiesSymbole(roh []byte) ([]Symbol, []domain.Befund, error) {
	var gelesen rohliste[rohsymbol]

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return nil, nil, fmt.Errorf("%w: die Symbologie ließ sich nicht lesen: %s",
			domain.ErrValidation, fehler.Error())
	}

	symbole := make([]Symbol, 0, len(gelesen.Daten))

	var befunde []domain.Befund

	for stelle, eines := range gelesen.Daten {
		if eines.Zeichen == "" {
			return nil, nil, fmt.Errorf(
				"%w: das Symbol an Stelle %d hat keine Schreibweise",
				domain.ErrValidation, stelle)
		}

		befunde = append(befunde, farbbefunde(eines)...)

		symbole = append(symbole, Symbol{
			Zeichen: eines.Zeichen, Beschreibung: eines.Beschreibung,
			IstMana: eines.IstMana, Manawert: eines.Manawert,
			InManakosten: eines.InManakosten, Farben: eines.Farben,
			Hybrid: eines.Hybrid, Phyrexianisch: eines.Phyrexianisch,
			Spassvariante: eines.Spassvariante, Vektoradresse: eines.Vektoradresse,
		})
	}

	return symbole, befunde, nil
}

// bekannteFarben sind die fünf Farben von Magic, wie Scryfall sie schreibt.
var bekannteFarben = []string{"W", "U", "B", "R", "G"}

// farbbefunde meldet eine Farbe, die dieses Programm nicht kennt.
//
// **Warum das kein Fehler ist**: eine sechste Farbe wäre eine Neuerung des
// Spiels, und die Anwendung soll sie anzeigen können, statt sich zu weigern.
// Sie ist aber auch nicht nichts — ein Manakosten mit ihr rechnet anders.
func farbbefunde(eines rohsymbol) []domain.Befund {
	var befunde []domain.Befund

	for _, farbe := range eines.Farben {
		if istBekannt(bekannteFarben, farbe) {
			continue
		}

		befunde = append(befunde, domain.Befund{
			Datei:   eines.Zeichen,
			Ort:     "colors",
			Art:     domain.ArtUmgebung,
			Meldung: fmt.Sprintf("Scryfall nennt für %s die unbekannte Farbe %q.", eines.Zeichen, farbe),
		})
	}

	return befunde
}
