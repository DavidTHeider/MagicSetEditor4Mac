//go:build fremdpakete

// **Gleiche Karte, anderer Aufwand** — Werkstück Ω.4 aus
// #1029.
//
// Das Werkstück verlangt einen Lauf, der belegt, dass **Layout und
// Zeilenumbruch über alle vier Güten gleich bleiben**.
//
// # Wie weit dieser Lauf das belegt
//
// **Er baut den Plan einmal und rastert ihn viermal.** Er zeigt also, dass der
// **Rasterweg** den Umbruch nicht verschiebt — nicht, dass ein güteabhängiger
// **Planbau** unmöglich wäre.
//
// Dass er unmöglich ist, trägt die Schichtengrenze und nicht dieser Test:
// `Kartenplan` nimmt keine Güte entgegen, sie kommt erst bei
// `Rasterquelle.Zeichne` hinzu. Wer das ändert, ändert eine Signatur — und
// dann steht dieser Test daneben und prüft die andere Hälfte.
//
// Hier stand „belegt ist es hier", und das war breiter als der Test.
// Angestrichen vom Zweitblick über Phase Ω.
package app

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// TestJedeGueteSetztDieselbenZeilen zeigt es an einer echten Karte.
//
// # Zwei Zusagen, und sie sind verschieden stark
//
//  1. **Eine Textebene ohne Symbole ist über alle vier Güten Punkt für Punkt
//     dieselbe.** Glyphen werden aus Umrissen gerastert und nicht abgetastet;
//     die Güte kann dort nichts ändern. Das ist die **starke** Zusage, und sie
//     schlägt fehl, sobald jemand die Güte in den Schriftweg zieht.
//
//  2. **Eine Ebene mit Symbolen hat dieselbe Tuschefläche.** Ihre Bildpunkte
//     unterscheiden sich — die Symbole sind Bilder —, aber der Kasten, in dem
//     Tusche steht, ist derselbe. Ein anderer Umbruch verschöbe ihn.
//
//     **Der Kasten ist das schwächere Maß**, und das gehört dazu: zwei
//     verschiedene Zeilenanordnungen können denselben äußeren Kasten haben.
//     Für Ebenen **ohne** Symbole gilt die stärkere Zusage darüber, und der
//     Lauf zählt, wie viele in welche Klasse fallen.
func TestJedeGueteSetztDieselbenZeilen(t *testing.T) {
	anwendung := gestartetMit(t, func() string { return fremdpaketwurzelOderSkip(t) })

	// **Ein anderer Satz als beim Messen**, mit Absicht: `Planechase` bringt
	// eine andere Vorlage mit, und die Zusage soll nicht an einer hängen.
	geoeffnet := anwendung.SetOeffnen(context.Background(),
		kanonsatzpfad(t, "Planechase"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	zeilen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) == 0 {
		t.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Skip("kein Rasterer")
	}

	plan := planZurKarte(t, anwendung, zeilen.Zeilen[0].ID)

	var ohneSymbole, mitSymbolen int

	for stelle, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if ebeneTraegtSymbole(ebene) {
			mitSymbolen++
		} else {
			ohneSymbole++
		}

		pruefeEineTextebene(t, rasterer, plan, stelle, ebene)
	}

	// **Gezählt und nicht angenommen.** Trüge jede Textebene Symbole, liefe
	// die starke Zusage nie — und der Lauf sähe genauso grün aus
	// (AGENTS.md §10, „keine stillen Deckelungen").
	t.Logf("%d Textebenen ohne Symbole (Punkt für Punkt verglichen), "+
		"%d mit Symbolen (nur die Tuschefläche)", ohneSymbole, mitSymbolen)

	if ohneSymbole == 0 {
		t.Fatal("keine einzige Textebene ohne Symbole — die starke Zusage lief nicht")
	}

	// **Und der Beleg, dass das Messgerät ausschlägt.** Eine Bildebene *muss*
	// sich zwischen zwei Güten unterscheiden; täte sie es nicht, wäre der
	// grüne Lauf oben eine Aussage über den Vergleich und nicht über die
	// Güten.
	pruefeDassBilderSichUnterscheiden(t, rasterer, plan)
}

// pruefeEineTextebene vergleicht eine Ebene über alle vier Güten.
func pruefeEineTextebene(
	t *testing.T, rasterer *render.Rasterer, plan *render.Plan,
	stelle int, ebene render.Ebene,
) {
	t.Helper()

	mitSymbolen := ebeneTraegtSymbole(ebene)

	var (
		ersteTusche image.Rectangle
		erstesBild  *image.NRGBA
	)

	for i, guete := range render.Gueten {
		bild, _, err := rasterer.Zeichne(context.Background(),
			einzelplan(plan, ebene), 1, guete)
		if err != nil {
			t.Fatalf("Ebene %d, Güte %s: %v", stelle, guete.Name(), err)
		}

		tusche := tuschekasten(bild)

		if i == 0 {
			ersteTusche, erstesBild = tusche, bild

			continue
		}

		if tusche != ersteTusche {
			t.Errorf("Ebene %d: die Güte %s setzt anders — Tusche %v statt %v",
				stelle, guete.Name(), tusche, ersteTusche)
		}

		if !mitSymbolen && !gleicheBildpunkte(erstesBild, bild) {
			t.Errorf("Ebene %d ohne Symbole: die Güte %s malt andere Bildpunkte",
				stelle, guete.Name())
		}
	}
}

// einzelplan gibt einen Plan mit genau einer Ebene, im Rahmen des Originals.
func einzelplan(plan *render.Plan, ebene render.Ebene) *render.Plan {
	kopie := *plan
	kopie.Ebenen = []render.Ebene{ebene}

	return &kopie
}

// ebeneTraegtSymbole sagt, ob irgendein Abschnitt in Symbolen steht.
//
// `<sym>T</sym>` mitten im Regeltext ist ein **Bild**, und Bilder werden
// abgetastet — für eine solche Ebene gilt nur die schwächere der beiden
// Zusagen.
func ebeneTraegtSymbole(ebene render.Ebene) bool {
	for _, abschnitt := range ebene.Abschnitte {
		if len(abschnitt.Symbole) > 0 {
			return true
		}
	}

	return false
}

// tuschekasten gibt den Kasten, in dem etwas Sichtbares steht.
//
// **Die Ecken werden getrennt geführt.** `image.Rect` normalisiert seine
// Ecken; ein „leeres" Rechteck aus `Rect(breite, hoehe, 0, 0)` ist deshalb die
// ganze Fläche — eine Falle, die beim Zeilenmessen für Ω.3 schon einmal ein
// Ergebnis vorgetäuscht hat.
func tuschekasten(bild *image.NRGBA) image.Rectangle {
	var (
		links, oben   = bild.Bounds().Dx(), bild.Bounds().Dy()
		rechts, unten = -1, -1
	)

	for y := range bild.Bounds().Dy() {
		for x := range bild.Bounds().Dx() {
			if bild.NRGBAAt(x, y).A == 0 {
				continue
			}

			links, oben = min(links, x), min(oben, y)
			rechts, unten = max(rechts, x), max(unten, y)
		}
	}

	if rechts < 0 {
		return image.Rectangle{}
	}

	return image.Rect(links, oben, rechts+1, unten+1)
}

// gleicheBildpunkte vergleicht zwei Bilder Punkt für Punkt.
func gleicheBildpunkte(eines, anderes *image.NRGBA) bool {
	if eines.Bounds() != anderes.Bounds() {
		return false
	}

	for stelle := range eines.Pix {
		if eines.Pix[stelle] != anderes.Pix[stelle] {
			return false
		}
	}

	return true
}

// TestWasJedeGueteKostet misst den Aufwand — die andere Hälfte von „gleiche
// Karte, anderer Aufwand".
func TestWasJedeGueteKostet(t *testing.T) {
	anwendung := gestartetMit(t, func() string { return fremdpaketwurzelOderSkip(t) })

	geoeffnet := anwendung.SetOeffnen(context.Background(), kanonsatzpfad(t, "Black"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	zeilen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) < ProfilKarten {
		t.Skipf("zu wenige Karten: %+v", zeilen.Fehler)
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Skip("kein Rasterer")
	}

	schritt := len(zeilen.Zeilen) / ProfilKarten
	jeGuete := map[render.Guete][]time.Duration{}

	for i := range ProfilKarten {
		plan := planZurKarte(t, anwendung, zeilen.Zeilen[i*schritt].ID)

		for _, guete := range render.Gueten {
			beginn := time.Now()

			if _, _, err := rasterer.Zeichne(context.Background(), plan, 1, guete); err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			jeGuete[guete] = append(jeGuete[guete], time.Since(beginn))
		}
	}

	for _, guete := range render.Gueten {
		meldeZeiten(t, guete.Name()+" ", jeGuete[guete])
	}
}

// pruefeDassBilderSichUnterscheiden ist die Gegenprobe zum Vergleich selbst.
func pruefeDassBilderSichUnterscheiden(
	t *testing.T, rasterer *render.Rasterer, plan *render.Plan,
) {
	t.Helper()

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		schnell, _, err := rasterer.Zeichne(context.Background(),
			einzelplan(plan, ebene), 1, render.GueteSchnell)
		if err != nil {
			continue
		}

		mittel, _, err := rasterer.Zeichne(context.Background(),
			einzelplan(plan, ebene), 1, render.GueteMittel)
		if err != nil {
			continue
		}

		if !gleicheBildpunkte(schnell, mittel) {
			return
		}
	}

	t.Error("keine einzige Bildebene unterscheidet sich zwischen zwei Güten — " +
		"der Vergleich oben belegt dann nichts")
}
