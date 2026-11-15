package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Form wird als Vieleck gefüllt — gerade und ungerade.**
//
// MSE zeichnet jede Form mit `DrawPolygon`, und dessen Vorgabe ist
// `wxODDEVEN_RULE` — ein Punkt liegt innen, wenn ein Strahl von ihm aus die
// Kanten **ungerade** oft schneidet. Für ein Symbol mit einem Loch heißt das,
// dass der innere Zug das Loch macht, gleich in welcher Richtung er läuft
// (#323).
//
// **Gemessen wird in der Mitte eines Bildpunkts.** Ein Vieleck von 2 bis 8
// deckt damit die Punkte 2 bis 7; der Punkt 8 liegt schon draußen.
func TestEineFormWirdGeradeUndUngeradeGefuellt(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(10, 10)

	puffer.Fuelle([]render.Punkt{{X: 2, Y: 2}, {X: 8, Y: 2}, {X: 8, Y: 8}, {X: 2, Y: 8}},
		255, render.PufferSetzen)

	faelle := []struct {
		x, y int
		wert uint8
	}{
		{x: 2, y: 2, wert: 255},
		{x: 7, y: 7, wert: 255},
		{x: 8, y: 4, wert: 0},
		{x: 4, y: 8, wert: 0},
		{x: 1, y: 4, wert: 0},
		{x: 0, y: 0, wert: 0},
	}

	for _, fall := range faelle {
		if haben := puffer.Bei(fall.x, fall.y); haben != fall.wert {
			t.Errorf("(%d,%d) ist %d, erwartet %d", fall.x, fall.y, haben, fall.wert)
		}
	}
}

// **Ein Loch entsteht aus einem zweiten Zug im ersten.**
func TestEinZweiterZugMachtEinLoch(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(12, 12)

	puffer.Fuelle([]render.Punkt{
		{X: 1, Y: 1},
		{X: 11, Y: 1},
		{X: 11, Y: 11},
		{X: 1, Y: 11},
		{X: 1, Y: 1},
		{X: 4, Y: 4},
		{X: 8, Y: 4},
		{X: 8, Y: 8},
		{X: 4, Y: 8},
	}, 255, render.PufferSetzen)

	if puffer.Bei(2, 2) != 255 {
		t.Error("der Rand des äußeren Zugs ist nicht gefüllt")
	}

	if puffer.Bei(6, 6) != 0 {
		t.Error("das Loch ist gefüllt")
	}
}

// **Die drei Rechenarten von MSE stehen zur Verfügung.**
//
// `SetLogicalFunction(wxAND)` beim Abziehen, `wxXOR` beim Unterschied,
// `wxCOPY` sonst; die Puffer werden mit `wxOR` und `wxAND_INVERT` auf den
// Schirm gebracht (`combineBuffers`, `src/render/symbol/viewer.cpp`, Zeilen
// 85–89, Fassung f9d9356 vom 2021-07-06).
func TestDieRechenartenDesPuffers(t *testing.T) {
	faelle := []struct {
		name     string
		art      render.Pufferart
		erwartet uint8
	}{
		{name: "setzen", art: render.PufferSetzen, erwartet: 255},
		{name: "und", art: render.PufferUnd, erwartet: 255},
		{name: "entweder", art: render.PufferEntweder, erwartet: 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			puffer := render.NeuerSymbolpuffer(4, 4)
			viereck := []render.Punkt{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}}

			puffer.Fuelle(viereck, 255, render.PufferSetzen)
			puffer.Fuelle(viereck, 255, fall.art)

			if haben := puffer.Bei(1, 1); haben != fall.erwartet {
				t.Errorf("zweimal 255 mit %s ergibt %d, erwartet %d",
					fall.name, haben, fall.erwartet)
			}
		})
	}
}

// **Der Rand ist ein Strich auf der Kante, keine zweite Form.**
//
//	border->SetPen(wxPen(*wxWHITE, (int) rotation.trS(border_radius)));
//	border->DrawPolygon((int)points.size(), &points[0]);
//
// `drawSymbolShape`, `src/render/symbol/viewer.cpp`, Zeilen 265–269, Fassung
// f9d9356 vom 2021-07-06. Der Strich sitzt **mittig** auf der Kante — die
// Hälfte liegt innen und wird später vom Inneren überschrieben, die andere
// Hälfte ist der sichtbare Rand.
//
// **Runde Ecken**: die Vorgabe von wxWidgets ist `wxJOIN_ROUND` und
// `wxCAP_ROUND`, und ein Symbol wird mit ihr gezeichnet
// (#323).
func TestDerRandSitztMittigAufDerKante(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(20, 20)

	puffer.Striche([]render.Punkt{
		{X: 5, Y: 5}, {X: 15, Y: 5}, {X: 15, Y: 15}, {X: 5, Y: 15},
	}, 4, 255, render.PufferSetzen)

	faelle := []struct {
		name string
		x, y int
		wert uint8
	}{
		{name: "innen neben der Kante", x: 6, y: 10, wert: 255},
		{name: "außen neben der Kante", x: 4, y: 10, wert: 255},
		{name: "in der Mitte der Form", x: 10, y: 10, wert: 0},
		{name: "weit außerhalb", x: 0, y: 10, wert: 0},

		// **Die Ecke ist mitgezeichnet** — ohne runde Ecken bliebe hier ein
		// Loch zwischen den beiden Strichen.
		{name: "in der Ecke", x: 5, y: 5, wert: 255},
	}

	for _, fall := range faelle {
		if haben := puffer.Bei(fall.x, fall.y); haben != fall.wert {
			t.Errorf("%s (%d,%d): %d, erwartet %d",
				fall.name, fall.x, fall.y, haben, fall.wert)
		}
	}
}

// **Was außerhalb liegt, ist null und wird nicht geschrieben.**
//
// Ein Symbol darf über seine Fläche hinausragen — MSE schneidet am Rand des
// `wxDC` ab, und ein Zugriff daneben ist hier kein Fehler, sondern nichts.
func TestAusserhalbDerFlaecheIstNichts(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(4, 4)

	puffer.Fuelle([]render.Punkt{
		{X: -10, Y: -10}, {X: 10, Y: -10}, {X: 10, Y: 10}, {X: -10, Y: 10},
	}, 255, render.PufferSetzen)

	if puffer.Bei(2, 2) != 255 {
		t.Error("die Fläche selbst blieb leer")
	}

	for _, daneben := range [][2]int{{-1, 0}, {0, -1}, {4, 0}, {0, 4}} {
		if puffer.Bei(daneben[0], daneben[1]) != 0 {
			t.Errorf("(%d,%d) außerhalb ist nicht null", daneben[0], daneben[1])
		}
	}
}

// **Eine Strecke der Länge null hat keine Querrichtung — die runde Ecke bleibt.**
//
// Das ist kein erfundener Fall: jeder Punkt einer Zeichnung wird auf einen
// Bildpunkt gerundet ([render.Zerlege]), und zwei benachbarte Punkte einer
// engen Kurve landen dabei regelmäßig auf demselben.
func TestEineStreckeOhneLaengeBleibtEinePunktscheibe(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(9, 9)

	puffer.Striche([]render.Punkt{{X: 4, Y: 4}, {X: 4, Y: 4}}, 4, 255, render.PufferSetzen)

	if puffer.Bei(4, 4) != 255 {
		t.Error("die Scheibe fehlt")
	}

	if puffer.Bei(0, 0) != 0 {
		t.Error("die Ecke ist gefüllt")
	}
}

// **Zu wenige Punkte sind keine Form, und ein Strich ohne Breite ist keiner.**
func TestWasSichNichtZeichnenLaesstBleibtLeer(t *testing.T) {
	puffer := render.NeuerSymbolpuffer(4, 4)

	puffer.Fuelle([]render.Punkt{{X: 0, Y: 0}, {X: 4, Y: 4}}, 255, render.PufferSetzen)
	puffer.Striche([]render.Punkt{{X: 0, Y: 0}, {X: 4, Y: 4}}, 0, 255, render.PufferSetzen)

	for y := range 4 {
		for x := range 4 {
			if puffer.Bei(x, y) != 0 {
				t.Fatalf("(%d,%d) ist gefüllt", x, y)
			}
		}
	}
}
