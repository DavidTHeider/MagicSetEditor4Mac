package service

import (
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Nennung ist, was neben einem übernommenen Wert stehen muss.
//
// # Warum es sie gibt
//
// Scryfall verlangt für den Bildausschnitt, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`:
//
//	When using the art_crop, list the artist name and copyright elsewhere in
//	the same interface presenting the art crop, or use the full card image
//	elsewhere in the same interface. Users should be able to identify the
//	artist and source of the image somehow.
//
// Erfüllt war das bis #1025
// **nur, solange das Zuschnittfenster offen war** — danach stand das Bild auf
// der Karte, und der Kasten mit der Nennung war fort. Eine Bedingung, die das
// Schließen eines Kastens nicht überlebt, ist keine.
type Nennung struct {
	// Adresse ist die Kartenseite bei Scryfall — die „source" der Bedingung.
	Adresse string

	// Felder sind die Felder dieser Karte, deren Wert von dort stammt.
	Felder []string

	// Kuenstler ist der Name aus dem Kartenfeld `illustrator`, falls gesetzt.
	Kuenstler string

	// Text ist der Satz, der dastehen muss.
	Text string
}

// feldIllustrator ist das Kartenfeld, in dem MSE den Künstler führt.
//
// **Belegt am 2026-08-24** an `magic.mse-game/card_fields` aus dem
// `M15-Magic-Pack`: unter 249 Feldern steht `illustrator` in Zeile 2738, und
// jede M15-Vorlage zeichnet es auf die Karte.
const feldIllustrator = "illustrator"

// Nennungen gibt zu einer Karte, was neben ihr stehen muss.
//
// # Gruppiert nach Herkunft, nicht nach Feld
//
// Wer Namen, Regeltext und Bild von **einer** Scryfall-Karte übernimmt, soll
// nicht dreimal denselben Satz lesen. Genannt wird, woher es kam — und das ist
// einmal dieselbe Seite. Die Felder stehen dabei, damit erkennbar bleibt,
// worauf sich die Nennung bezieht.
//
// # Die Illustrationszeile hängt am Bild, nicht am Illustrator
//
// **`bildfelder` sagt, welche Felder dieser Karte ein Bild zeichnen.** Nur
// wenn eines davon von dort stammt, wird eine Illustration behauptet; sonst
// steht der Rechtevermerk allein. Genannt wird die Herkunft trotzdem — auch
// ein übernommener Regeltext stammt von dort.
//
// **Gefunden bei der Sichtprüfung am 2026-08-24**, nachdem der Rest schon
// grün war: Bild und Künstler kamen zusammen von Scryfall — zwei Zeilen mit
// derselben Herkunft. Danach wurde über das Bildfeld ein eigener Wert
// geschrieben; seine Herkunftszeile fiel, die des Künstlers blieb, und die
// Karte trug weiter „Illustration: Chris Rahn" neben einem fremden Bild.
//
// Kein Test konnte das zeigen: jeder von ihnen setzte eine **einzige**
// Herkunft. Erst die laufende Anwendung hatte zwei.
//
// # Was diese Funktion **nicht** prüft
//
// **Ob die Herkunft noch zum Wert gehört** — das sagt die Tabelle
// `wert_herkunft` zu (Migration 0014): sie wird in derselben Transaktion
// geführt wie der Wert selbst.
//
// **Ob ein Bildfeld ohne zeichnende Ebene eines ist.** `bildfelder` kommt aus
// dem Zeichenplan, und ein Feld, das nichts malt, steht nicht darin. Für die
// Bedingung reicht das genau: wo kein Bild gezeigt wird, ist nichts zu nennen.
func Nennungen(
	karte *domain.CardRow, herkuenfte map[string]string, bildfelder []string,
) []Nennung {
	nach := map[string][]string{}

	for feld, herkunft := range herkuenfte {
		if !scryfall.IstKartenseite(herkunft) {
			continue
		}

		nach[herkunft] = append(nach[herkunft], feld)
	}

	if len(nach) == 0 {
		return nil
	}

	bilder := map[string]bool{}
	for _, feld := range bildfelder {
		bilder[feld] = true
	}

	kuenstler, _ := karte.Value(feldIllustrator)

	nennungen := make([]Nennung, 0, len(nach))

	for adresse, felder := range nach {
		// **In fester Reihenfolge.** Eine Map gibt sie in wechselnder heraus,
		// und die Anzeige tanzte dann bei jedem Zeichnen.
		sort.Strings(felder)

		// **Behauptet wird nur, was zutrifft.** Ohne Bildfeld aus dieser
		// Quelle steht der Rechtevermerk allein.
		genannt := ""
		if traegtBild(felder, bilder) {
			genannt = kuenstler
		}

		nennungen = append(nennungen, Nennung{
			Adresse: adresse, Felder: felder,
			Kuenstler: genannt, Text: scryfall.NennungFuer(genannt),
		})
	}

	sort.Slice(nennungen, func(a, b int) bool {
		return nennungen[a].Adresse < nennungen[b].Adresse
	})

	return nennungen
}

// traegtBild sagt, ob unter diesen Feldern ein Bildfeld ist.
func traegtBild(felder []string, bilder map[string]bool) bool {
	for _, feld := range felder {
		if bilder[feld] {
			return true
		}
	}

	return false
}
