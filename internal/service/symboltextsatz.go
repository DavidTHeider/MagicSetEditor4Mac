package service

// Der **Text auf einem Symbol** — die Ziffer im grauen Kreis.
//
// **Eigene Datei, seit die Farbe gerechnet wird**
// (#382):
// `stilsymbole.go` beantwortet, welche Symbole ein Zug führt und wie groß sie
// sind; hier steht, was auf ihnen geschrieben steht und in welcher Farbe. Der
// Anlass war die 400-Zeilen-Grenze (§2), der Schnitt folgt der Frage.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// faerbeSymboltexte rechnet die Textfarbe jedes Symbols aus.
//
// # Warum hier und nicht in [textsatzAus]
//
// Weil [symbolbilder] **zweimal** läuft: einmal für den Plan und einmal für die
// Messung von `content_width` ([symbolmass]). Die Messung braucht keine Farbe —
// und ein Befund, der dabei entstünde, stünde zweimal an derselben Karte.
//
// **Eine Farbe, die ein Skript ist**, ist auch hier der Normalfall (#382):
// `magic-mana-large` schreibt die Ziffernfarbe seiner Kreise so.
func faerbeSymboltexte(
	ctx context.Context, symbole []render.Symbol, treffer []Symboltreffer,
	feld string, umgebung *msescript.Umgebung,
) []domain.Befund {
	var befunde []domain.Befund

	for i := range symbole {
		farbe, weitere := farbangabe(ctx, feld+" → "+treffer[i].Symbol.Code, "text font color",
			treffer[i].Symbol.Textschrift.Farbe, umgebung)

		symbole[i].Textsatz.Farbe = farbe
		befunde = append(befunde, weitere...)
	}

	return befunde
}

// textsatzAus überträgt die Textangaben eines Symbols in den Plan.
//
// **Ohne die Farbe** — die rechnet [faerbeSymboltexte], und dort steht auch,
// warum sie nicht hierher gehört.
func textsatzAus(symbol msepaket.Symbol) render.Symboltextsatz {
	return render.Symboltextsatz{
		Schrift: symbol.Textschrift.Name,
		Groesse: symbol.Textschrift.Groesse,
		Dehnung: symbol.Textschrift.Dehnung,
		// **Ohne Angabe bleibt es leer**, und der Rasterer setzt `middle
		// center` ein — MSEs Vorgabe für den Symboltext. [ausrichtungAus]
		// gäbe hier links oben, denn das ist die Vorgabe eines **Feldes**.
		Ausrichtung: symboltextausrichtung(symbol.Textausrichtung),
		RandLinks:   symbol.TextrandLinks,
		RandRechts:  symbol.TextrandRechts,
		RandOben:    symbol.TextrandOben,
		RandUnten:   symbol.TextrandUnten,
	}
}
