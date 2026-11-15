// **Was neben einer Karte stehen muss** —
// #1025.
//
// Eine eigene Datei, weil es eine eigene Frage ist: `karteneditor.go` sagt,
// **wie** eine Karte gezeichnet wird — hier steht, **was daneben gehört**,
// wenn ihre Werte von woanders stammen.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nennungen stellt zusammen, was neben dieser Karte stehen muss.
//
// **Ein Fehler hält die Vorschau nicht auf** (§8): eine Karte, die nicht mehr
// gezeichnet wird, weil eine Nebenangabe fehlt, wäre der schlechtere Tausch —
// die Vorschau ist das Hauptobjekt (§6). Gemeldet wird er trotzdem, als
// Befund neben dem Bild: sonst stünde ein Bildausschnitt ohne seine Nennung
// da, **ohne dass es jemand merkt**.
func (a *App) nennungen(
	ctx context.Context, karte *domain.CardRow, _ *render.Plan,
) (nennungen []NennungDTO, befund string) {
	// **Eine leere Liste und nicht `null`** — auch auf dem Fehlerweg. Der
	// Umschlagwächter (`umschlag_test.go`) hat genau diese Zeile
	// angestrichen: eine nackte Rückgabe schickt `null`, und dann prüft jede
	// Seite zweimal statt einmal.
	herkuenfte, err := a.kartenleser.Herkuenfte(ctx, karte.ID)
	if err != nil {
		return []NennungDTO{},
			"die Herkünfte dieser Karte ließen sich nicht lesen: " + err.Error()
	}

	// **Die Bildfelder kommen aus der Vorlage, nicht aus dem Plan** (#1032).
	//
	// Der erste Anlauf las die Ebenen: „welches Feld **zeichnet** ein Bild?"
	// Das ist eine andere Frage — `rarity`, `card symbol` und `watermark`
	// zeichnen ebenfalls Bilder, und die Illustrationszeile stand dann neben
	// jeder Übernahme, die eine Seltenheit gesetzt hatte. **Gefunden an der
	// laufenden Anwendung**, nachdem alle Tests grün waren.
	//
	// **Ein Fehler beim Lesen der Felder hält nichts auf**: dann gibt es keine
	// Bildfelder, und die Nennung nennt den Rechteinhaber ohne Illustration.
	// Das ist die vorsichtigere Seite — behauptet wird nur, was zutrifft.
	gefunden := service.Nennungen(karte, herkuenfte, a.bildfelderDesSets(ctx, karte.SetID))

	nennungen = make([]NennungDTO, 0, len(gefunden))
	for _, nennung := range gefunden {
		nennungen = append(nennungen, NennungDTO{
			Adresse: nennung.Adresse, Felder: nennung.Felder,
			Kuenstler: nennung.Kuenstler, Text: nennung.Text,
		})
	}

	return nennungen, ""
}

// NennungDTO ist, was neben einem übernommenen Wert stehen muss.
type NennungDTO struct {
	// Adresse ist die Kartenseite bei Scryfall — die „source" der Bedingung.
	Adresse string `json:"adresse"`

	// Felder sind die Felder dieser Karte, deren Wert von dort stammt.
	Felder []string `json:"felder"`

	Kuenstler string `json:"kuenstler"`

	// Text ist der Satz, der dastehen muss.
	Text string `json:"text"`
}

// bildfelderDesSets liest die Bildfelder der Vorlage hinter diesem Set.
//
// **Ohne Stilwahl gibt es keine** — dann nennt die Nennung den Rechteinhaber
// ohne Illustration. Das ist die vorsichtigere Seite: behauptet wird nur, was
// zutrifft, und eine Fassade ohne Stilwahl zeichnet ohnehin den Behelfsplan.
//
// **Ein Lesefehler hält nichts auf** (§8): die Vorschau ist das Hauptobjekt,
// und eine Karte, die nicht mehr gezeichnet wird, weil ein Paket fehlt, wäre
// der schlechtere Tausch.
func (a *App) bildfelderDesSets(ctx context.Context, setID int64) []string {
	if a.stile == nil {
		return nil
	}

	felder, err := a.stile.Kartenfelder(ctx, setID)
	if err != nil {
		return nil
	}

	return bildfelder(felder)
}

// bildfelder sagt, welche Felder dieser Vorlage **Bildfelder** sind.
//
// # Warum die Feldart und nicht die Zeichenebene
//
// **Weil „zeichnet ein Bild" und „ist ein Bildfeld" nicht dasselbe sind.** Der
// erste Anlauf las die Ebenen des Zeichenplans — und `rarity`, `card symbol`
// und `watermark` zeichnen ebenfalls Bilder. Die Illustrationszeile stand
// danach neben jeder Übernahme, die eine Seltenheit gesetzt hatte.
//
// **Gefunden an der laufenden Anwendung** (#1032), nachdem alle Tests grün
// waren: eine Karte trug zwei Nennungen, und beide behaupteten dieselbe
// Illustration — die eine zu Recht, die andere über eine Textübernahme.
//
// `type: image` steht im Spielpaket und meint genau das, was ein Mensch als
// Bild in die Karte legt.
func bildfelder(felder []domain.Field) []string {
	namen := make([]string, 0, len(felder))

	for _, feld := range felder {
		if feld.Kind == domain.KindImage {
			namen = append(namen, feld.Name)
		}
	}

	return namen
}
