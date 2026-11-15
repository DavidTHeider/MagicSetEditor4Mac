package render_test

// **Was der Baum über sich selbst sagt** — und warum eine Messung ihn fragen
// muss, statt selbst zu raten.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`Gebaut` gilt für den Knoten und alles darunter.**
//
// # Wozu eine Messung das braucht
//
// Eine nicht gebaute Verknüpfung zeichnet ihr Ausgangsbild — sichtbar falsch
// statt unsichtbar fehlend. Damit trägt der Plan eine Bildebene, und dass sie
// nur zur Hälfte stimmt, stünde erst im Rasterbefund.
//
// Der Zweitblick über die Phasen B bis D hat das angestrichen: die 965
// `recolor_image` **verschwanden aus dem Zähler**, obwohl sich an der Karte
// nichts geändert hatte. Eine Messgröße darf ihre Bedeutung nicht
// stillschweigend wechseln.
// nochNichtGebaut ist eine erfundene Knotenart.
//
// **Sie muss erfunden sein.** Seit dem 2026-09-03 ist jede Bildfunktion, die
// `internal/msescript` registriert, auch gebaut — es gibt keine echte
// Knotenart mehr, die hier einspringen könnte. Geprüft wird der **Mechanismus**
// (eine unbekannte Art zeichnet ihr Ausgangsbild und meldet das), nicht das
// Fehlen einer bestimmten MSE-Funktion.
const nochNichtGebaut = "eine_kuenftige_verknuepfung"

func TestGebautGiltFuerDenGanzenBaum(t *testing.T) {
	faelle := map[string]struct {
		baum     render.Bildbaum
		erwartet bool
	}{
		"ein Blatt":              {render.Blatt("/a.png"), true},
		"eine gebaute Blende":    {mitKind("masked_blend", render.Blatt("/a.png")), true},
		"eine gebaute Umfärbung": {mitKind("recolor_image", render.Blatt("/a.png")), true},
		"eine fremde Knotenart":  {mitKind(nochNichtGebaut, render.Blatt("/a.png")), false},
		"fremd unter gebaut": {
			mitKind("linear_blend", mitKind(nochNichtGebaut, render.Blatt("/a.png"))),
			false,
		},
		"der leere Baum": {render.Bildbaum{}, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if got := fall.baum.Gebaut(); got != fall.erwartet {
				t.Errorf("Gebaut() ist %v, erwartet %v", got, fall.erwartet)
			}
		})
	}
}

// **`Blaetter` gibt alle Pfade, nicht den ersten.**
//
// Der Messwächter prüfte bis zum 2026-08-16 nur den Pfad am **Wurzelknoten** —
// bei einem Baum ist der leer, und damit prüfte er nichts.
func TestBlaetterGibtAllePfade(t *testing.T) {
	baum := render.Bildbaum{Knotenart: "masked_blend", Kinder: []render.Bildkind{
		{Name: "light", Baum: render.Blatt("/hell.png")},
		{Name: "dark", Baum: mitKind("linear_blend", render.Blatt("/tief.png"))},
		{Name: "mask", Baum: render.Blatt("/maske.png")},
	}}

	got := baum.Blaetter()
	erwartet := []string{"/hell.png", "/tief.png", "/maske.png"}

	if !slices.Equal(got, erwartet) {
		t.Errorf("die Blätter sind %v, erwartet %v", got, erwartet)
	}
}

func mitKind(art string, kind render.Bildbaum) render.Bildbaum {
	return render.Bildbaum{
		Knotenart: art, Kinder: []render.Bildkind{{Name: "image", Baum: kind}},
	}
}

// **Ein Ring wird benannt, nicht gelaufen.**
//
// # Der Befund des Zweitblicks
//
// Ein [render.Bildbaum] ist ein Werttyp, und trotzdem lässt sich über eine
// rückverweisende Slice ein Selbstkind bauen. Die Werte aus `msescript` sind
// heute azyklisch — **der Zeichenplan erzwingt das aber weder technisch noch
// vertraglich**, und ein Plan kommt aus einem fremden Paket.
//
// Ohne Grenze rekursierte die Ausführung bis zum Stapelüberlauf: kein Befund,
// keine Karte, kein Fenster.
func TestEinRingWirdBenannt(t *testing.T) {
	kinder := make([]render.Bildkind, 1)
	ring := render.Bildbaum{Knotenart: "recolor_image", Kinder: kinder}
	kinder[0] = render.Bildkind{Name: "image", Baum: ring}

	// **Und noch eine Windung**, damit der Ring wirklich einer ist.
	kinder[0].Baum.Kinder = kinder

	rasterer := render.NeuerRasterer(bildquelle{})

	_, befund := rasterer.AusBaumFuerTest(ring)
	if !strings.Contains(befund, "tiefer als") {
		t.Errorf("der Befund lautet %q und nennt die Tiefe nicht", befund)
	}
}
