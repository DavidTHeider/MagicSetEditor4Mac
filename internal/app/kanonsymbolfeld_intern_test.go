//go:build fremdpakete

// **Was der Kanonvergleich im Symbolfeld nicht misst** —
// #744.
//
// # Der Befund
//
// **MSE2 zeichnet auf diesem Rechner kein Symbolinneres.** Jedes Setsymbol in
// `mse2-ausgabe/` steht durchweg in seiner **Randfarbe**; die Füllung der
// Variante kommt nirgends vor. Nachgemessen am 2026-08-21 über einen Bisect mit
// `--export-images` an einem auf zwei Karten gekürzten `White.mse-set`:
//
//	Seltenheit      Füllung laut set_fields   Rand laut set_fields   MSE2 zeichnet
//	common          schwarz                   weiß                   weiß
//	basic land      schwarz                   weiß                   weiß
//	uncommon        Grauverlauf               schwarz                schwarz
//	rare            Goldverlauf               schwarz                schwarz
//	mythic rare     Orangeverlauf             schwarz                schwarz
//	special         Lilaverlauf               weiß                   weiß
//
// Sechsmal die Randfarbe, nullmal die Füllung.
//
// # Wo es herkommt
//
//	void combineBuffers(DC& dc, DC* borders, DC* interior) {
//	  if (borders)  dc.Blit(0,0,w,h, borders,  0,0, wxOR);
//	  if (interior) dc.Blit(0,0,w,h, interior, 0,0, wxAND_INVERT);
//	}
//
// `SymbolViewer::draw` und `combineBuffers`, `src/render/symbol/viewer.cpp`,
// Zeilen 85–118, Fassung f9d9356 vom 2021-07-06.
//
// Der Randpuffer trägt das **ganze** Vieleck — `border->SetBrush(…)` steht vor
// `DrawPolygon`, nicht nur ein Strich (Zeilen 265–270). Erst der zweite Blit
// nimmt das Innere wieder heraus. Bleibt er ohne Wirkung, steht das ganze
// Symbol in der Randfarbe, und außen bleibt es durchsichtig — genau das ist zu
// sehen.
//
// **Warum er ohne Wirkung bleibt, ist nicht nachgelesen** (uncertain, §15):
// wxWidgets liegt hier nur als Homebrew-Binärdatei vor. Beobachtet ist die
// Wirkung, nicht die Ursache. Gebaut ist gegen wx **3.3.3** statt der in MSE2s
// Anleitung geprüften 3.0.5 ([0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)).
//
// # Was das für die Zahl bedeutet
//
// Über acht Karten aus `White`, mit der Metrik dieses Projekts gemessen:
//
//	ganze Karte              22,202
//	ohne das Symbolfeld      21,423
//	im Symbolfeld allein    158,557
//
// **Rund 0,78 des Abstands sind dort nicht einzuholen.** Es ist kein Fehler
// dieses Programms, sondern eine Eigenheit des Vergleichsmaßstabs — und der
// Rest der Karte ist davon unberührt: mit der alten wie der neuen
// Symbolauflösung steht „ohne das Symbolfeld" auf **derselben** Zahl
// (#742).
package app

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// Symbolfeldkern ist der Bereich, in dem bei `magic-m15-altered` nur das
// Setsymbol steht — ohne seine geglätteten Ränder.
//
// **Die Zahlen sind gemessen, nicht gerechnet**: die Seltenheitsebene fragt dort
// 44 × 22 Punkte, und das Symbol landet mittig rechts darin.
var Symbolfeldkern = image.Rect(327, 302, 340, 315)

// **Der Wächter zu dem Befund: MSE2s Symbol ist einfarbig.**
//
// # Warum das die richtige Zusicherung ist
//
// Ein richtig gezeichnetes Symbol hat **zwei** Farben — Füllung und Rand, und
// die Varianten des Spiels unterscheiden sie durchweg. Einfarbig kann es nur
// sein, wenn eine der beiden fehlt. Die Prüfung braucht damit weder die
// Seltenheit der Karte noch die Farbtabelle des Spiels.
//
// # Was er nicht prüft
//
// **Nicht, dass unser Symbol zwei Farben hat.** Das steht in
// `internal/render` (`symbolfilter_test.go`) und gehört dorthin; hier geht es
// allein um den Maßstab.
//
// **Und nicht, ob die Zahl oben noch stimmt.** Sie ist eine Messung von acht
// Karten eines Sets; wird MSE2 neu gebaut, wird dieser Wächter rot, und dann
// gehört sie neu erhoben.
func TestMSE2ZeichnetKeinSymbolinneres(t *testing.T) {
	ordner := referenzort(t, "White")
	if _, err := os.Stat(ordner); err != nil {
		t.Skip("keine MSE2-Ausgabe zu White")
	}

	// **Name, erwartete Randfarbe und Seltenheit** — die Farbe steht im
	// Spielpaket (`magic.mse-game/set_fields`, ab Zeile 542), und sie ist die
	// des **Randes**, nicht der Füllung. Genau das ist die Zusicherung.
	karten := []struct {
		name, seltenheit string
		rand             color.RGBA
	}{
		{"Emeria Angel", "rare", color.RGBA{A: 255}},
		{"Reveillark", "rare", color.RGBA{A: 255}},
		{"Steel-Plume Marshal", "rare", color.RGBA{A: 255}},
		{"Bring Down", "uncommon", color.RGBA{A: 255}},
		{"Hotshot Mechanic", "uncommon", color.RGBA{A: 255}},
		{"Melesse Spirit", "uncommon", color.RGBA{A: 255}},
	}

	geprueft := 0

	for _, karte := range karten {
		name := karte.name
		bild, err := ladeBild(filepath.Join(ordner, referenzdatei(name)))
		if err != nil {
			continue
		}

		geprueft++

		farben := map[color.RGBA]int{}

		for y := Symbolfeldkern.Min.Y; y < Symbolfeldkern.Max.Y; y++ {
			for x := Symbolfeldkern.Min.X; x < Symbolfeldkern.Max.X; x++ {
				r, g, b, a := bild.At(bild.Bounds().Min.X+x, bild.Bounds().Min.Y+y).RGBA()
				farben[color.RGBA{
					R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8),
				}]++
			}
		}

		if len(farben) != 1 {
			t.Errorf("%s: %d Farben im Symbolfeld — MSE2 zeichnet wieder ein "+
				"Inneres, und der Vermerk oben gehört neu geschrieben: %v",
				name, len(farben), farben)

			continue
		}

		// **Und es ist die Randfarbe der Seltenheit, nicht irgendeine.**
		//
		// Ohne diese Zeile hielte der Wächter auch dann, wenn das Symbol ganz
		// fehlte oder der Ausschnitt danebenläge — ein einfarbiger Hintergrund
		// ist ebenso einfarbig. Angestrichen vom Zweitblick durch Codex am
		// 2026-08-21: „reine Farbanzahl ohne positive Prüfung von Ort und
		// erwarteter Farbe."
		for farbe := range farben {
			if farbe != karte.rand {
				t.Errorf("%s (%s): das Symbolfeld ist %v, erwartet die Randfarbe %v",
					name, karte.seltenheit, farbe, karte.rand)
			}
		}
	}

	// **Ohne eine einzige geprüfte Karte sagt der Lauf nichts** (§10). Ein
	// Wächter, der stillschweigend null Fälle prüft, ist grün und wertlos.
	if geprueft == 0 {
		t.Skip("keines der genannten Referenzbilder liegt vor")
	}
}
