package service_test

// **Der Kartenrücken** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass zwei verdeckte Karten ununterscheidbar sind** — in Bild **und** in
// Größe. Unterschieden sie sich in einem von beidem, wäre die Zusage
// „verdeckt" hinfällig, ohne dass ein einziges Feld verraten würde.
//
// **Dass der Rücken aus dem Spielpaket kommt, wenn es einen mitbringt** — er
// liegt auf diesem Rechner, weil jemand das Paket bezogen hat, und geht
// denselben Weg wie jeder Rahmen (§11).
//
// **Und dass sonst einer gezeichnet wird**, der keine fremde Datei nennt.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Rücken ist für jede Karte derselbe — und er fragt keine Karte.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Ein Tisch, auf dem Rückseiten anders aussähen, verriete die Karte.** Zwei
// verdeckte Karten müssen ununterscheidbar sein; unterschieden sie sich in
// Größe oder Bild, wäre die Zusage „verdeckt" hinfällig, ohne dass ein
// einziges Feld verraten würde.
func TestDerRueckenFragtKeineKarte(t *testing.T) {
	t.Parallel()

	werk, plaene, rasterer, _ := eineWerkstatt(t)

	erster, fehler := werk.Ruecken(t.Context(), "magic", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	zweiter, fehler := werk.Ruecken(t.Context(), "magic", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	switch {
	case erster.Schluessel != zweiter.Schluessel:
		t.Errorf("zwei Rücken: %q und %q",
			erster.Schluessel, zweiter.Schluessel)
	case rasterer.gezeichnet != 1:
		t.Errorf("%dmal gezeichnet, erwartet einmal", rasterer.gezeichnet)
	case plaene.gefragt != 0:
		t.Errorf("der Planbauer wurde %dmal gefragt, erwartet keinmal",
			plaene.gefragt)
	case erster.Adresse == "":
		t.Error("keine Adresse")
	}
}

// **Zwei Spiele haben zwei Rücken.**
//
// **Sonst zeigte eine Partie in einem anderen Spiel den Rücken von Magic** —
// und der Schlüssel, der beide gleich machte, verhinderte auch noch, dass es
// jemand merkt.
func TestZweiSpieleHabenZweiRuecken(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	magic, fehler := werk.Ruecken(t.Context(), "magic", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	// **`saat` bringt keinen mit** — die Vorrichtung kennt nur `magic`.
	saat, fehler := werk.Ruecken(t.Context(), "saat", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	if magic.Schluessel == saat.Schluessel {
		t.Errorf("beide Spiele geben %q", magic.Schluessel)
	}
}

// **Der Rücken aus dem Paket wird genommen, wo einer liegt.**
func TestDerRueckenAusDemPaketWirdGenommen(t *testing.T) {
	t.Parallel()

	plan := service.Rueckseitenplan("/magic.mse-game/card-back.png")

	pfade := render.Bildpfade(plan)
	if len(pfade) != 1 || pfade[0] != "/magic.mse-game/card-back.png" {
		t.Errorf("Bildpfade %v", pfade)
	}

	if plan.Breite != service.Rueckenbreite ||
		plan.Hoehe != service.Rueckenhoehe {
		t.Errorf("%v×%v", plan.Breite, plan.Hoehe)
	}
}

// **Und wo keiner liegt, trägt der gezeichnete kein fremdes Bild.**
//
// **Die Aussage von §11 im Lauf**: der Plan nennt keine einzige Bilddatei.
// Ein Rahmen aus einem Paket darin wäre fremde Grafik, die dieses Programm
// mitlieferte — und genau das darf es nicht.
func TestDerGezeichneteRueckenNenntKeineDatei(t *testing.T) {
	t.Parallel()

	plan := service.Rueckseitenplan("")

	if pfade := render.Bildpfade(plan); len(pfade) != 0 {
		t.Errorf("der gezeichnete Rücken nennt %v", pfade)
	}

	if len(plan.Ebenen) == 0 {
		t.Error("er zeichnet gar nichts")
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtFlaeche {
			t.Errorf("eine Ebene der Art %v", ebene.Art)
		}
	}
}

// **Eine Güte, die es nicht gibt, wird auch am Rücken abgewiesen.**
//
// **Der Fehlerweg gehört dazu** (§10): wer nur den Erfolgsweg prüft, hat die
// Hälfte ausgelassen — und dieser Weg hat seine eigene Prüfung, weil er
// nicht durch [service.Spieltexturen.Textur] geht.
func TestEineUnbekannteGueteWirdAmRueckenAbgewiesen(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	_, fehler := werk.Ruecken(t.Context(), "magic", "riesig")
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), "Texturgüte") {
		t.Errorf("Fehler %q", fehler)
	}
}

// **Jede Ebene des Rückens ist sichtbar.**
//
// # Der Lauf, der gefehlt hat
//
// **Eine Ebene ohne `Deckung` malt nichts.** [render.Ebene.Deckung] geht von
// 0 bis 1, der Nullwert von `float64` ist die Null, und der Rasterer steigt
// bei `anteil <= 0` aus. Der erste Anlauf setzte sie nirgends: am Spieltisch
// standen achtzehn Rücken, deren PNG an **jeder** Stelle `(0,0,0,0)` war —
// zu sehen war der CSS-Rahmen des Kästchens darum.
//
// **Kein Lauf sah es.** Sie prüften den Plan, den Schlüssel, die Maße und die
// Bildpfade — alle vier waren richtig. Gefunden hat es die Sichtprüfung
// (§3), und zwar erst, als die Bilder überhaupt ankamen.
//
// **Deshalb steht hier eine Zusicherung über Bildpunkte** und nicht über den
// Plan: dass eine Ebene **da** ist, sagt nicht, dass man sie sieht.
func TestJedeEbeneDesRueckensIstSichtbar(t *testing.T) {
	t.Parallel()

	for name, plan := range map[string]render.Plan{
		"gezeichnet": service.Rueckseitenplan(""),
		"aus dem Paket": service.Rueckseitenplan(
			"/magic.mse-game/card-back.png"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if len(plan.Ebenen) == 0 {
				t.Fatal("der Plan hat keine Ebene")
			}

			for i, ebene := range plan.Ebenen {
				if ebene.Deckung <= 0 {
					t.Errorf("Ebene %d hat die Deckung %v — sie malt nichts",
						i+1, ebene.Deckung)
				}
			}
		})
	}
}

// **Und der gezeichnete Rücken kommt wirklich gefärbt heraus.**
//
// **Die Zusicherung eine Stufe weiter**: nicht, dass die Ebenen eine Deckung
// tragen, sondern dass am Ende Farbe dasteht. Gerastert wird mit dem
// **echten** Rasterer — eine Vorrichtung könnte hier jede Zahl behaupten.
func TestDerGezeichneteRueckenIstNichtDurchsichtig(t *testing.T) {
	t.Parallel()

	plan := service.Rueckseitenplan("")

	bild, _, fehler := render.NeuerRasterer(nil).Zeichne(
		t.Context(), &plan, 1, render.GueteSchnell)
	if fehler != nil {
		t.Fatalf("Zeichne: %v", fehler)
	}

	grenzen := bild.Bounds()

	mitte := bild.NRGBAAt(grenzen.Dx()/2, grenzen.Dy()/2)
	if mitte.A == 0 {
		t.Errorf("die Mitte ist durchsichtig: %v", mitte)
	}

	// **Und der Rand ebenso** — er ist die zweite Ebene.
	rand := bild.NRGBAAt(grenzen.Dx()/2, 4)
	if rand.A == 0 {
		t.Errorf("der Rand ist durchsichtig: %v", rand)
	}

	// **Beide sind verschieden**, sonst zeigte der Rücken eine Fläche statt
	// eines Musters.
	if mitte == rand {
		t.Errorf("Rand und Mitte sind beide %v", mitte)
	}
}
