package app

import "context"

// **Die Bindungen der Marken an Karten** —
// #1208.
//
// Eine eigene Datei, weil `bindungenkarten.go` mit ihnen über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort stehen die Bindungen über die
// **Karte** — ihre Werte, ihre Bilder, ihre Symbole —, hier die über das, was
// **neben** ihr liegt.

// **Marken an Karten**
// (#1208).

// MarkeSetzen hängt eine Marke an eine Karte.
func (b *Bindungen) MarkeSetzen(ctx context.Context, setID, karteID, marke string) SchreibAntwort {
	return b.app(ctx).MarkeSetzen(b.ruf(ctx), setID, karteID, marke)
}

// MarkeEntfernen nimmt eine Marke wieder ab.
func (b *Bindungen) MarkeEntfernen(ctx context.Context, setID, karteID, marke string) SchreibAntwort {
	return b.app(ctx).MarkeEntfernen(b.ruf(ctx), setID, karteID, marke)
}
