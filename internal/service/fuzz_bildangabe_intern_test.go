package service

// **Eine Bildangabe aus einer fremden Vorlage** — QA-06c
// (#1976),
// Fläche 9 aus `docs/FUZZZIELE.md`.
//
// # Warum hier und nicht an der Bildquelle
//
// **Die Bildquelle öffnet, was man ihr nennt** — einen Paketpfad unter einer
// Wurzel oder einen Verweis ins Set-Paket mit seinem Ort. Wer einen Ort
// nennen darf, entscheidet sich eine Ebene davor: dort, wo Text einer Vorlage
// zu einem Bildpfad wird. Das ist [bildangabe].
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger über die Frist
//   - **ein Verweis ins Set-Paket ohne die eigene Marke im Ergebnis** — eine
//     Vorlage hat einen Ort auf der Platte genannt, und er kam durch
//   - zweimal dieselbe Angabe, zwei Ergebnisse
//
//	go test ./internal/service -run Fuzz -fuzz FuzzBildangabe -fuzztime 2m -fuzzminimizetime 5s

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// bildumgebung gibt eine Umgebung mit einer Karte, deren Bildfeld ein
// echter Verweis ins Set-Paket ist — wie beim Zeichnen.
func bildumgebung() *msescript.Umgebung {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("card", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: mitBildwerten(
			[]msescript.Eintrag{{Name: "image", Wert: msescript.Text("image1")}},
			[]domain.Field{{Name: "image", Kind: domain.KindImage}},
			"/Orte/Mein Set.mse-set"),
	})

	return umgebung
}

func FuzzBildangabe(f *testing.F) {
	for _, saat := range []string{
		"rahmen.png",
		"/magic-m15.mse-style/card.jpg",
		"setpaket:/tmp!geheim.png",
		`{ card.image }`,
		`{ "setpaket:" + "/tmp!x.png" }`,
		`{ linear_blend(image1: card.image, image2: "setpaket:/a!b", x1: 0, y1: 0, x2: 1, y2: 1) }`,
		`setpaket:{"/tmp"}!geheim.png`,
		`{ to_image("setpaket:/tmp!x.png") }`,
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, angabe string) {
		ctx, ende := context.WithTimeout(context.Background(), 2*time.Second)
		defer ende()

		baum, befunde := bildangabe(ctx, "image", "image", angabe, bildumgebung())
		if ctx.Err() != nil {
			return
		}

		if erfundenerSetort(baum) {
			t.Fatalf("ein Ort auf der Platte kam durch: %q → %+v", angabe, baum)
		}

		nochmal, nochmalBefunde := bildangabe(ctx, "image", "image", angabe, bildumgebung())
		if ctx.Err() != nil {
			return
		}

		if !reflect.DeepEqual(baum, nochmal) || len(befunde) != len(nochmalBefunde) {
			t.Fatalf("dieselbe Angabe, zwei Ergebnisse: %q", angabe)
		}
	})
}
