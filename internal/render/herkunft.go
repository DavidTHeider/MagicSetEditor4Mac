package render

import (
	"context"
	"image"
)

// Herkunft sagt für jeden Bildpunkt, welche Ebene ihn **zuletzt geändert** hat.
//
// # Warum nicht das Rechteck der Ebene
//
// Der erste Anlauf von
// #521 hat
// die Fläche jeder Ebene aus ihrem [Ebene.Rahmen] genommen und den Bildpunkt
// der obersten Ebene zugeschlagen, deren Rahmen ihn enthält. Am echten Kanon
// kam dabei **100 % Bild** heraus — und die Zerlegung war wertlos.
//
// Der Grund steht im Plan einer gewöhnlichen M15-Karte: **Ebene 25 von 26 ist
// ein Bild über die ganze Karte** (0,0 375×523), die Überlagerung. Sie bedeckt
// jeden Bildpunkt und bemalt fast keinen.
//
// **Rechteckdeckung ist nicht Bemalung.** Wer sie verwechselt, misst die
// Geometrie des Plans statt das Bild.
//
// # Was hier steht, ist keine Schätzung
//
// Der Rasterer legt die Ebenen nacheinander auf **ein** Ziel. Wird vor und
// nach jeder Ebene verglichen, steht fest, welche Bildpunkte sie geändert hat
// — ohne Regel, ohne Annahme über Deckung, Maske oder Mischart.
//
// # Was sie nicht sagt
//
// Sie sagt, **wer zuletzt gemalt hat** — nicht, wer schuld ist. Das schneidet
// in **beide** Richtungen:
//
//   - Ein Rahmenbild, das falsch gezeichnet und danach von einem Textkasten
//     überschrieben wird, erscheint als Text.
//   - **Und eine Ebene, die gar nicht malt, erscheint überhaupt nicht.** Ihr
//     Abstand fällt der Ebene darunter zu oder [ArtUnbedeckt]. Fehlender Text
//     zählt damit als Bild.
//
// **Der zweite Fall ist gemessen, nicht gedacht**: die Sagaseite einer
// doppelseitigen Karte zeichnet ihre Kapitelkästen nicht
// (#529), und
// dieser Abstand steht in der Zerlegung unter `Bild`.
//
// **Eine Zerlegung nach dieser Karte ist deshalb eine Aufteilung und keine
// Ursachenmessung.** Wer aus „Text trägt 50,94 %" schließt, der Satz sei der
// größte **Hebel**, überdehnt sie — angestrichen vom Zweitblick über Phase I.
type Herkunft struct {
	Breite, Hoehe int

	// arten ist zeilenweise abgelegt, [ArtUnbedeckt] für jeden Bildpunkt, den
	// keine Ebene je geändert hat.
	arten []Ebenenart
}

// NeueHerkunft legt eine leere Karte an — jeder Bildpunkt unbedeckt.
func NeueHerkunft(breite, hoehe int) *Herkunft {
	arten := make([]Ebenenart, max(0, breite*hoehe))
	for i := range arten {
		arten[i] = ArtUnbedeckt
	}

	return &Herkunft{Breite: breite, Hoehe: hoehe, arten: arten}
}

// Art gibt die Ebenenart, die diesen Bildpunkt zuletzt geändert hat.
//
// Außerhalb der Karte gilt [ArtUnbedeckt] — ein Zugriff daneben ist kein
// Absturz, sondern eine Auskunft.
func (h *Herkunft) Art(x, y int) Ebenenart {
	if h == nil || x < 0 || y < 0 || x >= h.Breite || y >= h.Hoehe {
		return ArtUnbedeckt
	}

	return h.arten[y*h.Breite+x]
}

// bytesJeBildpunkt ist die Schrittweite in `image.NRGBA.Pix`.
const bytesJeBildpunkt = 4

// vermerke schreibt jedem geänderten Bildpunkt diese Art zu.
func (h *Herkunft) vermerke(vorher, nachher []uint8, art Ebenenart) {
	for i := range h.arten {
		stelle := i * bytesJeBildpunkt

		if vorher[stelle] != nachher[stelle] ||
			vorher[stelle+1] != nachher[stelle+1] ||
			vorher[stelle+2] != nachher[stelle+2] ||
			vorher[stelle+3] != nachher[stelle+3] {
			h.arten[i] = art
		}
	}
}

// ZeichneMitHerkunft rastert wie [Rasterer.Zeichne] und sagt dabei, **welche
// Ebene welchen Bildpunkt zuletzt geändert hat**.
//
// **Dasselbe Bild wie [Rasterer.Zeichne].** Die Aufzeichnung ändert nichts am
// Ergebnis; sie vergleicht nur zwischendurch. Ein Test hält das fest — sonst
// wäre die Messung an einem anderen Bild erhoben als dem, das die Anwendung
// zeigt (§10).
//
// **Sie kostet.** Je Ebene wird das halbfertige Bild kopiert und verglichen;
// bei 26 Ebenen und 375 × 523 sind das rund 20 MB Kopieren je Karte. Deshalb
// ist es eine eigene Methode und keine Voreinstellung.
func (r *Rasterer) ZeichneMitHerkunft(
	ctx context.Context, plan *Plan, vergroesserung float64,
) (*image.NRGBA, *Herkunft, []Befund, error) {
	// **Parität, und das ist hier keine Nebensache**: die Herkunftsaufzeichnung
	// dient dem Vergleich mit MSE2. Eine andere Güte verglich ein Bild, das MSE
	// so nicht zeichnet.
	return r.zeichne(ctx, plan, vergroesserung, GueteMittel, true)
}
