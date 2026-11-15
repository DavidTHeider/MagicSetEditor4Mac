package service_test

// **Die zweite öffentliche Seite** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass eine zweiseitige Karte zwischen ihren Seiten wechselt** — zwei
// Bilder, zwei Schlüssel, und beide so breit wie jede andere Karte am Tisch.
//
// **Und dass „Rücken" und „zweite Seite" nicht dasselbe sind.** Das Issue
// sagt es wörtlich, und hier ist es zu sehen: der Rücken kennt keine Karte,
// die zweite Seite ist eine.

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineZweiseitigeKarte gibt eine Karte, die `name 2` führt.
//
// **Ohne Kennung als Parameter**: welche Zeile es ist, spielt hier keine
// Rolle — der Ausschnitt hängt an den Werten und am Plan.
//
// **Und einen Plan doppelter Breite dazu** — so setzt MSE sie:
// `magic-m15-mainframe-dfc.mse-style` ist 752 × 523 statt 375 × 523
// (gemessen am 2026-09-15).
func eineZweiseitigeKarte() *domain.CardRow {
	return &domain.CardRow{ID: 1, Values: []domain.Value{
		{Field: "name", Text: "Delver of Secrets"},
		{Field: "name 2", Text: "Insectile Aberration"},
	}}
}

// **Die zwei Seiten sind zwei Bilder — und jede ist so breit wie eine Karte.**
func TestZweiSeitenSindZweiBilder(t *testing.T) {
	t.Parallel()

	werk, plaene, rasterer, _ := eineWerkstatt(t)

	breit := texturplan("rahmen.png", "MPlantin")
	breit.Breite = 752
	plaene.plan = breit

	vorne, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	zweite, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
		service.SeiteZweite, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if vorne.Schluessel == zweite.Schluessel {
		t.Errorf("beide Seiten geben %q", vorne.Schluessel)
	}

	// **Beide sind eine Kartenbreite breit und nicht eine halbe.**
	//
	// **Das ist der Unterschied, den die Vergrößerung macht**: ohne sie
	// zeichnete der Rasterer die ganze Vorlage auf 512 Punkte, und jede
	// Seite käme mit 256 daher — halb so groß wie jede einseitige Karte
	// daneben.
	ziel := service.GueteTisch.Breite()
	if vorne.Breite != ziel || zweite.Breite != ziel {
		t.Errorf("Seiten %d und %d breit, erwartet je %d",
			vorne.Breite, zweite.Breite, ziel)
	}

	// **Und sie sind gleich hoch wie eine gewöhnliche Karte.**
	if vorne.Hoehe != zweite.Hoehe {
		t.Errorf("Höhen %d und %d", vorne.Hoehe, zweite.Hoehe)
	}

	// **Gezeichnet wurde zweimal** — je Seite ein Ausschnitt, je Ausschnitt
	// ein eigenes Bild im Fach.
	if rasterer.gezeichnet != 2 {
		t.Errorf("%dmal gezeichnet, erwartet zweimal", rasterer.gezeichnet)
	}
}

// **Eine einseitige Karte hat keine zweite Seite — und sagt es.**
//
// # Warum abgewiesen und nicht behelfsweise die erste
//
// **Weil ein stillschweigend ausgetauschtes Bild den Fehler versteckte.** Wer
// die zweite Seite eines Waldes verlangt, hat einen Fehler im Aufrufer; er
// fällt sonst erst auf, wenn sich jemand über eine Karte wundert, die sich
// nicht umdrehen lässt.
func TestEineEinseitigeKarteHatKeineZweiteSeite(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	_, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteZweite, service.GueteTisch)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), "keine zweite Seite") {
		t.Errorf("Fehler %q", fehler)
	}
}

// **Eine einseitige Karte wird nicht zugeschnitten.**
//
// **Die Gegenprobe zum Lauf darüber**: läge der Schnitt an der Zahl der
// Seiten und nicht an der Karte, bekäme jede gewöhnliche Karte ihre halbe
// Breite.
func TestEineEinseitigeKarteWirdNichtZugeschnitten(t *testing.T) {
	t.Parallel()

	werk, _, rasterer, _ := eineWerkstatt(t)

	textur, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if textur.Breite != rasterer.breite {
		t.Errorf("Textur %d breit, gezeichnet %d — da wurde geschnitten",
			textur.Breite, rasterer.breite)
	}
}

// **Und die zweite Seite ist nicht der Rücken.**
//
// **Der Satz aus dem Issue, als Lauf**: „Kartenrückseite ist kein Synonym
// für die zweite öffentliche Kartenseite." Sie gehen durch zwei Wege, geben
// zwei Schlüssel — und der eine fragt eine Karte, der andere ein Spiel.
func TestDieZweiteSeiteIstNichtDerRuecken(t *testing.T) {
	t.Parallel()

	werk, plaene, _, _ := eineWerkstatt(t)

	breit := texturplan("rahmen.png", "MPlantin")
	breit.Breite = 752
	plaene.plan = breit

	zweite, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
		service.SeiteZweite, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	ruecken, fehler := werk.Ruecken(t.Context(), "magic", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	if zweite.Schluessel == ruecken.Schluessel {
		t.Errorf("beide geben %q", zweite.Schluessel)
	}
}

// **Ein Wort, das keine Kartenseite ist, wird abgewiesen.**
func TestEinWortDasKeineKartenseiteIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	_, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		"ruecken", service.GueteTisch)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	// **Gerade `ruecken` ist der Fall, der zählt**: er war einmal eine
	// Kartenseite, und wer ihn noch so schickt, soll es merken.
	if !strings.Contains(fehler.Error(), "Kartenseite") {
		t.Errorf("Fehler %q", fehler)
	}
}

// **Und die zweite Seite zeigt wirklich die rechte Hälfte.**
//
// # Warum dieser Lauf neben dem ersten steht
//
// **Weil jener grün blieb, als beide Seiten dieselbe Hälfte zeigten.**
// Gemessen am 2026-09-15: ein Eingriff, der `SeiteZweite` auf die **linke**
// Hälfte legt, ließ `TestZweiSeitenSindZweiBilder` durchgehen — der Lauf
// vergleicht Schlüssel und Maße, und die unterscheiden sich ohnehin, weil
// die Seite im Schlüssel steht. Die eigentliche Aussage ist eine über
// **Bildpunkte**, und sie war ungemessen (§10: eine Gegenprobe, die nichts
// ändert, ist keine).
func TestDieZweiteSeiteZeigtDieRechteHaelfte(t *testing.T) {
	t.Parallel()

	werk, plaene, _, ablage := eineWerkstatt(t)

	breit := texturplan("rahmen.png", "MPlantin")
	breit.Breite = 752
	plaene.plan = breit

	gemalt := map[service.Kartenseite]color.NRGBA{}

	for _, seite := range []service.Kartenseite{
		service.SeiteVorne, service.SeiteZweite,
	} {
		textur, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
			seite, service.GueteTisch)
		if fehler != nil {
			t.Fatalf("Textur %q: %v", seite, fehler)
		}

		gepackt, da := ablage.Bild(textur.Schluessel)
		if !da {
			t.Fatalf("kein Bild zu %q", seite)
		}

		bild, fehler := png.Decode(bytes.NewReader(gepackt))
		if fehler != nil {
			t.Fatalf("das Bild zu %q: %v", seite, fehler)
		}

		grenzen := bild.Bounds()
		r, g, b, a := bild.At(grenzen.Min.X+1, grenzen.Min.Y+1).RGBA()

		gemalt[seite] = color.NRGBA{
			R: uint8(r >> 8), G: uint8(g >> 8),
			B: uint8(b >> 8), A: uint8(a >> 8),
		}
	}

	if gemalt[service.SeiteVorne] != service.LinkeFarbe {
		t.Errorf("die Vorderseite ist %v, erwartet die linke Hälfte %v",
			gemalt[service.SeiteVorne], service.LinkeFarbe)
	}

	if gemalt[service.SeiteZweite] != service.RechteFarbe {
		t.Errorf("die zweite Seite ist %v, erwartet die rechte Hälfte %v",
			gemalt[service.SeiteZweite], service.RechteFarbe)
	}
}

// **Eine Vorlage, die zu schmal für zwei Gesichter ist, wird nicht
// zugeschnitten.**
//
// # Dass es diesen Fall gibt, ist nachgerechnet und nicht vermutet
//
// **Bei einem Bildpunkt Breite liegt die Mitte auf dem linken Rand**:
// `mitte = Min.X + Dx()/2` ist bei `Dx() == 1` gleich `Min.X`, und die linke
// Hälfte wäre damit ein Rechteck ohne Fläche. Wer daraus einen Ausschnitt
// bauen wollte, griffe hinter das Ende der Bildpunkte — die
// Standardbibliothek prüft an derselben Stelle aus demselben Grund
// (`$GOROOT/src/image/image.go`, `(*NRGBA).SubImage`: „the `Pix[i:]`
// expression below can panic").
//
// **Wann das im Betrieb vorkommt**: wenn ein Rasterer fast nichts liefert —
// eine Vorlage ohne Maße, ein Faktor nahe null. Dann steht ein winziges Bild
// da statt eines Sturzes.
func TestEineZuSchmaleVorlageWirdNichtZugeschnitten(t *testing.T) {
	t.Parallel()

	plaene := &planprobe{plan: texturplan("rahmen.png", "MPlantin")}
	rasterer := &winzigerrasterer{}

	werk := service.NeueSpieltexturen(plaene, rasterer,
		ressourcenprobe{}, service.NeueTexturfachprobe())

	// **Beide Seiten, und die linke ist die, die leer wäre**: bei einem
	// Bildpunkt Breite liegt `mitte` auf `Min.X`.
	for _, seite := range []service.Kartenseite{
		service.SeiteVorne, service.SeiteZweite,
	} {
		textur, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
			seite, service.GueteTisch)
		if fehler != nil {
			t.Fatalf("Textur %q: %v", seite, fehler)
		}

		// **Das ganze Bild und nicht die Hälfte von nichts.**
		if textur.Breite != 1 || textur.Hoehe != 1 {
			t.Errorf("%q: Maße %d×%d, erwartet 1×1",
				seite, textur.Breite, textur.Hoehe)
		}
	}
}

// winzigerrasterer gibt ein Bild von einem Bildpunkt.
type winzigerrasterer struct{}

func (winzigerrasterer) Zeichne(
	_ context.Context, _ *render.Plan, _ float64, _ render.Guete,
) (*image.NRGBA, []render.Befund, error) {
	return image.NewNRGBA(image.Rect(0, 0, 1, 1)), nil, nil
}

// **Die Textur sagt, ob die Karte zwei Seiten hat.**
//
// # Wofür die Auskunft da ist
//
// **Damit das Fenster weiß, ob es eine zweite Seite anbieten darf.** Ohne sie
// bliebe nur, es zu **versuchen** — und eine Schaltfläche, die meistens einen
// Fehler gibt, ist keine.
//
// **Und sie verrät nichts**: wer ein Bild bekommt, darf die Karte sehen. Für
// ein verdecktes Objekt gibt es kein Bild, sondern den Rücken, und der kennt
// keine Karte.
func TestDieTexturSagtObDieKarteZweiSeitenHat(t *testing.T) {
	t.Parallel()

	werk, plaene, _, _ := eineWerkstatt(t)

	breit := texturplan("rahmen.png", "MPlantin")
	breit.Breite = 752
	plaene.plan = breit

	zwei, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if !zwei.Zweiseitig {
		t.Error("eine zweiseitige Karte meldet sich als einseitig")
	}

	// **Und beim Treffer steht es auch da** — der zweite Ruf geht durch das
	// Fach, und eine Auskunft, die nur beim ersten Mal stimmt, ist keine.
	nochmal, fehler := werk.Textur(t.Context(), eineZweiseitigeKarte(),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if !nochmal.Zweiseitig {
		t.Error("beim Treffer fällt die Auskunft weg")
	}

	plaene.plan = texturplan("rahmen.png", "MPlantin")

	eine, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if eine.Zweiseitig {
		t.Error("eine einseitige Karte meldet sich als zweiseitig")
	}

	// **Und der Rücken sagt nichts** — er kennt keine Karte.
	ruecken, fehler := werk.Ruecken(t.Context(), "magic", service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Ruecken: %v", fehler)
	}

	if ruecken.Zweiseitig {
		t.Error("der Rücken behauptet eine zweite Seite")
	}
}
