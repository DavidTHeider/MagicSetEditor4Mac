package service

// **Zwei Seiten auf einer Vorlage** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Was hier nachgesehen wurde
//
// **Eine zweiseitige Karte ist in MSE *eine* Karte auf einer doppelt breiten
// Vorlage.** Nicht zwei Kartenobjekte, nicht zwei Pläne: derselbe `card
// style` erklärt `name` **und** `name 2`, `image` **und** `image 2`, und er
// setzt sie nebeneinander. Gemessen am 2026-09-15 über `M15-Magic-Pack`:
//
//   - `magic-m15-mainframe-dfc.mse-style` ist **752 × 523** statt 375 × 523,
//     `name` beginnt bei 58 bis 79, `name 2` bei 435 bis 456.
//   - `magic-m15-showcase-kaldheim-dfc.mse-style` ebenso, `name 2` bei 422.
//   - Es sind **die einzigen zwei** Stile im Pack mit dieser Vorlagengröße.
//
// # Warum daraus ein Ausschnitt wird und kein zweiter Plan
//
// **Weil das Issue neue Kartenrenderlogik ausschließt** — „Kein Kartenlayout
// in Pixi und keine Bild-Base64-Rückkehr", und der Planbauer bleibt, wie er
// ist. Ein Plan je Seite hieße, dem Planbauer eine Seite beizubringen; ein
// Ausschnitt heißt, aus dem fertigen Bild die Hälfte zu nehmen, die gemeint
// ist. Das ist eine Auswahl und kein Zeichnen.
//
// # Warum genau in der Mitte
//
// **Weil die Vorlage aus zwei Karten besteht und die Naht dazwischen liegt.**
// 752 = 375 + 2 + 375: zwei Kartenbreiten und zwei Punkte Fuge. Geteilt wird
// bei 376 — jede Hälfte bekommt damit einen Punkt der Fuge mit, und das ist
// **0,27 %** der Breite an einer Stelle, an der Hintergrund steht.
//
// **Die Alternative wäre, die Fuge auszurechnen**, und dafür müsste dieses
// Paket die Einzelbreite einer Karte kennen — die steht in keinem Plan. Eine
// gerechnete Zahl, die aus einer Annahme kommt, wäre schlechter als eine
// gemessene Näherung, die ihren Fehler nennt (§10).
//
// # Und warum die Breite nicht aus der Vorlage geschlossen wird
//
// **Weil sie es nicht verrät.** Im Pack sind `744 × 1039` **eine** Karte bei
// 300 dpi (40 Stile), `523 × 375` eine **quer** liegende (7 Stile, darunter
// `split` und `battles`) und `752 × 523` zwei Seiten (2 Stile). Das
// Seitenverhältnis von quer (1,39) und zweiseitig (1,44) liegt so dicht
// beieinander, dass keine Schwelle sie trennt. Gefragt wird deshalb die
// **Karte** — [domain.CardRow.HatZweiteSeite] —, und die weiß es.

import "image"

// ausschnittswunsch sagt, welches Gesicht einer Vorlage gemeint ist.
type ausschnittswunsch struct {
	seite Kartenseite

	// gesichter ist, wie viele Seiten nebeneinander auf der Vorlage stehen.
	gesichter int
}

// ausschnittAus nimmt aus dem gezeichneten Bild das gewünschte Gesicht.
//
// **Bei einem Gesicht ist es das ganze Bild** — und zwar dasselbe und keine
// Kopie: ein Zuschnitt, der nichts zuschneidet, soll nichts kosten.
func ausschnittAus(bild *image.NRGBA, wunsch ausschnittswunsch) *image.NRGBA {
	if bild == nil || wunsch.gesichter < 2 {
		return bild
	}

	grenzen := bild.Bounds()
	mitte := grenzen.Min.X + grenzen.Dx()/2

	feld := image.Rect(grenzen.Min.X, grenzen.Min.Y, mitte, grenzen.Max.Y)
	if wunsch.seite == SeiteZweite {
		feld = image.Rect(mitte, grenzen.Min.Y, grenzen.Max.X, grenzen.Max.Y)
	}

	// **Von Hand und nicht über `SubImage`.**
	//
	// # Warum nicht das Verfahren aus der Standardbibliothek
	//
	// **Weil es `image.Image` gibt und dieser Typ enger ist.** Eine
	// Typzusicherung darauf hätte einen Zweig für den Fall, dass sie
	// misslingt — und der ist **unerreichbar**: `(*image.NRGBA).SubImage`
	// gibt immer wieder ein `*image.NRGBA` zurück. Ein unerreichbarer Zweig
	// ist eine unbetretene Anweisung, und §3 kennt dafür keine Ausnahme.
	//
	// **Was hier steht, ist dasselbe.** Nachgelesen in
	// `$GOROOT/src/image/image.go`, `(*NRGBA).SubImage`: Schnittmenge mit
	// dem eigenen Rechteck, dann derselbe `Pix` ab dem Versatz, derselbe
	// `Stride`, das neue Rechteck.
	//
	// **Und der leere Fall tritt ein**, sobald die Vorlage einen Bildpunkt
	// breit ist: `mitte` ist dann `Min.X`, und die **linke** Hälfte ist ein
	// Rechteck ohne Fläche. Ohne diese Zeile griffe der Ausdruck darunter
	// hinter das Ende von `Pix` — die Standardbibliothek prüft an derselben
	// Stelle aus demselben Grund („the `Pix[i:]` expression below can
	// panic"). Dann steht ein winziges Bild da statt eines Sturzes.
	feld = feld.Intersect(grenzen)
	if feld.Empty() {
		return bild
	}

	return &image.NRGBA{
		Pix:    bild.Pix[bild.PixOffset(feld.Min.X, feld.Min.Y):],
		Stride: bild.Stride,
		Rect:   feld,
	}
}
