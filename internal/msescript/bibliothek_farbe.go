package msescript

import (
	"context"
	"fmt"
)

// `rgb` und `rgba` — die einzige Funktion, die in **drei** Spielen vorkommt.
//
// 95 Aufrufe in `magic-m15`, `magic.mse-game`, `yugioh-series10` und
// `kamen-decade`. Sie ist damit als einzige sicher weder aus einem Spielpaket
// noch aus einem Modul, sondern gehört der Sprache.
func init() {
	eintragen("rgb", farbeRGB)
	eintragen("rgba", farbeRGBA)
}

// Farbe ist ein Farbwert mit Deckung.
//
// **Ein eigener Wert und keine Abbildung.** Was `rgb` zurückgibt, ist aus den
// Dateien nicht zu sehen — sichtbar ist nur, dass es mit drei Zahlen gerufen
// und einem Farbfeld zugewiesen wird. Eine Abbildung mit `r`, `g`, `b` wäre
// dieselbe Behauptung in unauffälligerer Form: dann könnte eine Vorlage
// `farbe.r` schreiben, und niemand hätte je gesehen, dass das geht.
type Farbe struct {
	R, G, B, A uint8
}

// ArtFarbe ist die Art eines Farbwerts.
const ArtFarbe WertArt = "Farbe"

// AlsFarbe gibt den Farbwert, wenn es einer ist.
func (w Wert) AlsFarbe() (Farbe, bool) {
	if w.Art != ArtFarbe {
		return Farbe{}, false
	}

	return Farbe{
		R: w.Eintraege[0].Wert.anteil(), G: w.Eintraege[1].Wert.anteil(),
		B: w.Eintraege[2].Wert.anteil(), A: w.Eintraege[3].Wert.anteil(),
	}, true
}

func (w Wert) anteil() uint8 { return uint8(w.Zahl) } //nolint:gosec // durch anteilWert begrenzt

func farbwert(rot, gruen, blau, deckung Wert) Wert {
	return Wert{Art: ArtFarbe, Eintraege: []Eintrag{
		{Name: "r", Wert: rot},
		{Name: "g", Wert: gruen},
		{Name: "b", Wert: blau},
		{Name: "a", Wert: deckung},
	}}
}

func farbeRGB(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.genauStellung("rgb", 3); err != nil {
		return Nichts, err
	}

	return baueFarbe(ruf, "rgb", ruf.Stellung[0], ruf.Stellung[1], ruf.Stellung[2], Zahl(255))
}

func farbeRGBA(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.genauStellung("rgba", 4); err != nil {
		return Nichts, err
	}

	return baueFarbe(ruf, "rgba",
		ruf.Stellung[0], ruf.Stellung[1], ruf.Stellung[2], ruf.Stellung[3])
}

// baueFarbe prüft die vier Anteile und begrenzt sie auf 0…255.
//
// **Begrenzt, nicht abgewiesen.** In `yugioh` steht `rgba(0,0,0,0)` an Dutzenden
// Stellen; ein Wert daneben, der aus einer Rechnung kommt, kann über 255
// hinauslaufen, ohne dass die Vorlage kaputt ist. Ein Fehler machte dort die
// ganze Karte unbrauchbar, wo eine Farbe reicht.
func baueFarbe(ruf Ruf, name string, rot, gruen, blau, deckung Wert) (Wert, error) {
	anteile := make([]Wert, 0, 4)

	for _, wert := range []Wert{rot, gruen, blau, deckung} {
		if wert.Art != ArtZahlwert {
			return Nichts, ruf.fehlerAn(fmt.Sprintf(
				"%s rechnet mit Zahlen, hier steht %s", name, wert.Art))
		}

		anteile = append(anteile, Zahl(begrenzt(wert.Zahl)))
	}

	return farbwert(anteile[0], anteile[1], anteile[2], anteile[3]), nil
}

func begrenzt(zahl float64) float64 {
	switch {
	case zahl < 0:
		return 0

	case zahl > 255:
		return 255
	}

	return float64(int(zahl))
}

// Farbwert macht aus einer Farbe einen Wert.
//
// **Ausgeführt, weil `internal/service` Farben in einen Bildbaum setzt** —
// `symbol_variation` trägt die Farben seiner Fassung
// (#323).
func Farbwert(farbe Farbe) Wert {
	return farbwert(Zahl(float64(farbe.R)), Zahl(float64(farbe.G)),
		Zahl(float64(farbe.B)), Zahl(float64(farbe.A)))
}
