// **Die Symbolpalette ohne Fenster**
// (#1391).
//
// # Was die Fähigkeit ohne Fenster ist
//
// Am Fenster ist die Palette eine Reihe von Knöpfen: zu jedem Code ein
// gerastertes Bildchen, damit man **sieht**, was man einfügt. Der Katalog
// nannte sie deshalb „eine Liste zum Anklicken … für jemanden, der hinsieht",
// und ließ sie am Fenster.
//
// **Die Hälfte davon trägt aber gar kein Bild.** Ein Eintrag hat `code`,
// `beschriftung`, `art` und `rueckfrage`; das Bild ist die **Darstellung**
// eines Codes, nicht der Code. Wer über MCP einen Manakosten schreibt, will
// wissen, **welche Codes es gibt** — und das ist eine Auskunft wie jede
// andere.
//
// # Warum ohne die Bilder
//
// Weil sie das Teure sind: `palettenDTO` rastert jeden Eintrag einzeln. Eine
// Kommandozeile, die dreißig PNG als Data-URI ausgibt, ist unbenutzbar, und
// ein Modell liest sie nicht. **Das ist eine Auslassung mit Grund und kein
// Versehen** — wer ein Bild will, zeichnet die Karte (`mse zeichnen`).

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Palettenzeile ist ein Paletteneintrag ohne sein Bild.
type Palettenzeile struct {
	// Art ist `code`, `custom`, `line` oder `submenu` — **unübersetzt** (§9).
	Art string `json:"art"`

	// Code ist, was eingefügt wird. Leer bei Trenner und Untermenü.
	Code string `json:"code"`

	Beschriftung string `json:"beschriftung"`

	// Rueckfrage ist der Satz über dem Eingabefeld eines `custom`-Eintrags.
	Rueckfrage string `json:"rueckfrage"`
}

// Palette gibt, was ein Feld zum Einfügen anbietet — **ohne die Bildchen**.
func (z *Zeichenwerk) Palette(
	ctx context.Context, karteID int64, feld string,
) ([]Palettenzeile, error) {
	karte, err := z.Leser.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, err
	}

	// **Ohne Zweig auf den Fehler**, und das ist kein Versehen: er wäre
	// `if err != nil { return nil, err }` und damit dasselbe wie die Zeile
	// darunter — `alsPalettenzeilen(nil)` gibt eine leere Liste, und der
	// Aufrufer sieht den Fehler ohnehin zuerst.
	//
	// **Er wäre außerdem nicht zu begehen.** `Kartenpalette` scheitert nur,
	// wenn ein Stil **gefunden** wurde und seine Symbolschrift ausfällt; ein
	// Set ohne Stil gibt keinen Fehler, sondern nichts. Ein Zweig ohne Fall
	// ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
	eintraege, _, _, err := z.Planbauer.Kartenpalette(ctx, karte, feld)

	return alsPalettenzeilen(eintraege), err
}

// alsPalettenzeilen wirft das Bild weg und behält die Auskunft.
func alsPalettenzeilen(eintraege []service.Paletteneintrag) []Palettenzeile {
	zeilen := make([]Palettenzeile, 0, len(eintraege))

	for _, eintrag := range eintraege {
		zeilen = append(zeilen, Palettenzeile{
			Art: string(eintrag.Art), Code: eintrag.Code,
			Beschriftung: eintrag.Beschriftung, Rueckfrage: eintrag.Rueckfrage,
		})
	}

	return zeilen
}
