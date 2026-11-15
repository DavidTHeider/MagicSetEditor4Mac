package render_test

// **Der Tiefenvertrag gilt für jeden, der den Baum abläuft**
// (#318).
//
// `ausBaum` zählt seit jeher mit; `Beschreibung`, `Gebaut` und `Blaetter`
// liefen an demselben Ring weiter bis zum Stapelüberlauf. Der „rote" Zustand
// war dort kein Fehlschlag, sondern ein **Absturz des ganzen Laufs** — deshalb
// prüft dieser Test in erster Linie, dass die drei überhaupt zurückkehren.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ring baut einen Baum, dessen **erstes** Kind er selbst ist.
//
// **Über die geteilte Slice.** Ein [render.Bildbaum] ist ein Werttyp, und ein
// Selbstkind ließe sich als Wert nicht bauen — der Kopiervorgang endete. Die
// Kinder liegen aber in einer Slice, und ihr Kopf zeigt auf **dasselbe** Feld:
// `b.Kinder[0].Baum.Kinder` ist danach `b.Kinder`.
func ring() render.Bildbaum {
	baum := render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Bildbaum{}}},
	}
	baum.Kinder[0].Baum = baum

	return baum
}

// ringMitBlatt hat ein Blatt **vor** dem Selbstkind.
//
// **Für [render.Bildbaum.Blaetter] gebraucht**: ein Ring ohne Blatt gäbe zu
// Recht nichts zurück, und dann sagte der Test nichts darüber, ob sie
// zurückkehrt.
func ringMitBlatt() render.Bildbaum {
	baum := render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder: []render.Bildkind{
			{Name: "image", Baum: render.Blatt("/a.png")},
			{Name: "mask", Baum: render.Bildbaum{}},
		},
	}
	baum.Kinder[1].Baum = baum

	return baum
}

func TestDerTiefenvertragGiltFuerAlleAblaeufer(t *testing.T) {
	t.Run("Blaetter", func(t *testing.T) {
		if blaetter := ringMitBlatt().Blaetter(); len(blaetter) == 0 {
			t.Error("keine Blätter")
		}
	})

	t.Run("Gebaut", func(t *testing.T) {
		if !ring().Gebaut() {
			t.Error("der Baum gilt als nicht gebaut")
		}
	})

	t.Run("Beschreibung", func(t *testing.T) {
		beschreibung := ring().Beschreibung()

		if !strings.HasPrefix(beschreibung, "recolor_image(") {
			t.Errorf("Beschreibung %q", beschreibung)
		}

		// **Sie bricht ab und sagt es.** Eine Beschreibung, die still auf
		// vierzig Ebenen endet, sähe aus wie ein echter Baum.
		if !strings.Contains(beschreibung, "…") {
			t.Errorf("die Beschreibung sagt nicht, dass sie abbricht: %q", beschreibung)
		}
	})
}

// **Ein gewöhnlicher Baum bleibt unberührt** — die Gegenprobe. Ohne sie
// bestünde der Test auch, wenn die drei bei der ersten Verknüpfung aufgäben.
func TestEinFlacherBaumBleibtUnberuehrt(t *testing.T) {
	baum := render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/a.png")}},
	}

	if beschreibung := baum.Beschreibung(); beschreibung != "recolor_image(/a.png)" {
		t.Errorf("Beschreibung %q", beschreibung)
	}

	if blaetter := baum.Blaetter(); len(blaetter) != 1 || blaetter[0] != "/a.png" {
		t.Errorf("Blätter %v", blaetter)
	}

	if !baum.Gebaut() {
		t.Error("der Baum gilt als nicht gebaut")
	}
}
