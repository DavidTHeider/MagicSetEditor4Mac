package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Wahlbilder einer Stiloption** —
// #1225.
//
// # Was MSE zeigt
//
// Die Auswahlliste einer Stiloption trägt bei MSE je Wahl ein Vorschaubild:
//
//	if (style().choice_images.empty() && style().image.isScripted()) { … }
//	…
//	ScriptableImage& img = style().choice_images[name];
//
// `DropDownChoiceListBase::generateThumbnailImages`, `src/gui/value/choice.cpp`,
// Zeilen 145–185, Fassung f9d9356 vom 2021-07-06. Wer zwischen `modern`, `old`
// und `diagonal T` wählen soll, sieht dort drei Tapsymbole.
//
// # Wie viel davon es gibt
//
// Gezählt über `full-magic-pack/data` am 2026-08-28: **1.097 `choice
// images:`-Blöcke** im `styling style:`-Abschnitt von **303 der 376**
// Stilpakete, zusammen **2.707 Einträge**, davon **400 Skripte**. Kein Block
// führt mehr als **fünf** Wahlen — 591 führen zwei, 500 drei, fünf führen fünf.
//
// Diese Zahl ist der Grund, warum die Oberfläche eine **Reihe von Kacheln**
// zeichnen darf statt einer nachgebauten Auswahlliste: sie ist nach oben
// beschränkt, und zwar vom Material.

// Optionsbild ist das Bild einer Wahl einer Stiloption.
type Optionsbild struct {
	// Wahl ist der Name der Wahl, wie die Option ihn erklärt.
	//
	// **`choiceName(i)` und nicht der gespeicherte Wert** — MSE schlägt das
	// Bild unter demselben Namen nach (`choice.cpp`, Zeile 175). Bei einer
	// `choice` und einem `boolean` sind beide dasselbe; bei einer Farbe ist
	// der Name `schwarz` und der Wert `rgb(0,0,0)`, und die Oberfläche ordnet
	// über die **Beschriftung** zu.
	Wahl string

	// Baum ist das Bild als Zeichenbaum — noch nicht gerastert.
	//
	// **Die Naht aus §2**: `service` rechnet den Plan, `render` malt. Wer hier
	// ein PNG zurückgäbe, machte aus dem Planbauer einen zweiten Rasterer.
	Baum render.Bildbaum
}

// Optionsbilder gibt zu einer Stiloption je Wahl ihr Bild.
//
// **Je Option und nicht je Stil**: `magic-m15` erklärt dreizehn Optionen und
// gibt vieren ein Bild; über das ganze Pack wären es 1.097. Was gezeichnet
// wird, entscheidet die Oberfläche — sie fragt nur nach dem, was sie zeigt.
//
// # Warum ohne Karte
//
// Weil die Stilseite keine aufgeschlagene hat, und weil keines der 400
// gerechneten Wahlbilder des Packs `card` liest — alle vierhundert lesen
// `set` (gezählt am 2026-08-28). MSE macht es an dieser Stelle ebenso:
// `StylingEditor::showStylesheet` zeigt die Optionen **des Stils**, nicht die
// einer Karte (`src/gui/control/native_look_editor.cpp`, Zeilen 204–207,
// Fassung f9d9356 vom 2021-07-06).
//
// **Läse eine Vorlage doch `card`**, ergäbe der Ausdruck `nil` und die Wahl
// bekäme kein Bild — dasselbe, was MSEs `or else nil` tut.
func (p *Planbauer) Optionsbilder(
	ctx context.Context, setID int64, option string,
) ([]Optionsbild, []domain.Befund, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	if p.stile == nil {
		return nil, befundliste("für die Wahlbilder ist kein Stilverzeichnis eingerichtet"), nil
	}

	vorlage, befunde, da := p.stile.StilFuerSet(ctx, setID)
	if !da {
		return nil, befunde, nil
	}

	feldstil, hatStil := vorlage.Stil.Optionsstil(option)
	if !hatStil {
		return nil, nil, nil
	}

	wahlen := optionswahlen(vorlage.Stil.Stiloptionen, option)
	if len(wahlen) == 0 {
		return nil, nil, nil
	}

	umgebung, _, umgebungsbefunde := umgebungAus(
		ctx, &domain.CardRow{SetID: setID}, vorlage, p.schriften, p.bilder)
	befunde = append(befunde, umgebungsbefunde...)

	symbole := symbolquelle(p.bilder)

	bilder := make([]Optionsbild, 0, len(wahlen))

	for _, wahl := range wahlen {
		baum, weitere := wahlvorschaubild(ctx, &feldstil, wahl, umgebung)
		befunde = append(befunde, weitere...)

		if baum.Knotenart == "" {
			continue
		}

		bilder = append(bilder, Optionsbild{
			Wahl: wahl,
			Baum: alsZeichenbaum(baum, vorlage.Stil.Paket, symbole),
		})
	}

	return bilder, befunde, nil
}

// wahlvorschaubild gibt das Bild einer einzelnen Wahl.
//
// **Zwei Wege, und MSE hat beide** (`choice.cpp`, Zeilen 147–160):
//
//   - stehen `choice images:` da, gilt der Eintrag zu dieser Wahl;
//   - steht stattdessen ein gerechnetes `image:`, wird es je Wahl gerufen,
//     mit dem Wahlnamen als `input`.
//
// Der zweite Weg ist der, den man weglässt — und dann bliebe die Auswahlliste
// genau dort ohne Bilder, wo der Stil eines **rechnet**.
func wahlvorschaubild(
	ctx context.Context, feldstil *msepaket.Feldstil, wahl string,
	umgebung *msescript.Umgebung,
) (msescript.Bildbaum, []domain.Befund) {
	if len(feldstil.Wahlbilder) > 0 {
		return wahlbild(ctx, feldstil, wahl, umgebung)
	}

	// **Nur ein gerechnetes `image:`** — `style().image.isScripted()`
	// (`choice.cpp`, Zeile 149). Ein fester Dateiname stünde für **jede** Wahl
	// gleich da, und eine Reihe aus dreimal demselben Bild sagt nichts.
	if !bildGerechnet(feldstil.Bild) {
		return msescript.Bildbaum{}, nil
	}

	// **`input` trägt den Wahlnamen** — `ctx.setVariable(_("input"), …)`
	// (`choice.cpp`, Zeile 153). Es ist derselbe Name, den `choice images:`
	// intern benutzt, und deshalb rechnet dasselbe Skript beide Male gleich.
	eigene := umgebung.Darueber()
	eigene.Setze("input", msescript.Text(wahl))

	return bildangabe(ctx, feldstil.Feld, "image", feldstil.Bild, eigene)
}

// optionswahlen gibt die erklärten Wahlen einer Stiloption.
//
// **Die Erklärung und nicht die Sichtseite**: welche Wahlen es gibt, sagt
// `styling field:`; welches Bild dazugehört, sagt `styling style:`. Eine
// Wahl ohne Bild wird übergangen, ein Bild ohne Wahl nie gesucht.
func optionswahlen(optionen []domain.StyleField, name string) []string {
	for _, option := range optionen {
		if option.Name == name {
			return option.Choices
		}
	}

	return nil
}
