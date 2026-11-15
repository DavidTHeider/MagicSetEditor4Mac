package service

// **Eine Vorlage nennt keinen Ort auf der Platte** — QA-06c
// (#1976),
// Fläche 9 aus `docs/FUZZZIELE.md`.
//
// `setpaket:<ort>!<datei>` ist eine Form **dieses Programms**: sie trägt den
// Ort des Set-Pakets, den nur der Bestand kennt ([msepaket.SetpaketVerweis]).
// Gebaut wird sie an einer Stelle — [kartenbild], aus dem Dateinamen der Karte
// und dem Ort ihres Sets. MSE liest ein Kartenbild nur aus dem eigenen Paket
// (`getLocalPackage().openIn`, `src/render/value/image.cpp`, Zeile 35,
// Fassung f9d9356 vom 2021-07-06).
//
// **Eine Stilangabe aber ist Text einer fremden Vorlage.** Schrieb sie selbst
// `setpaket:/Users/…/Pictures!foto.png` — wörtlich oder als Ergebnis eines
// Skripts —, reichte `paketpfad` den Verweis durch, und die Bildquelle öffnete
// den genannten Ort als Paket. Eine Vorlage las damit Bilder von überall.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// enthaeltSetpaket sagt, ob irgendwo im Baum ein Verweis ins Set-Paket steht.
func enthaeltSetpaket(baum msescript.Bildbaum) bool {
	if strings.Contains(baum.Pfad, "setpaket:") {
		return true
	}

	for _, kind := range baum.Kinder {
		if enthaeltSetpaket(kind.Baum) {
			return true
		}
	}

	return false
}

func TestEineStilangabeNenntKeinenOrtAufDerPlatte(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, angabe string }{
		{"wörtlich", "setpaket:/Users/jemand/Pictures!foto.png"},
		{"mit Leerraum davor", "  setpaket:/tmp!geheim.png"},
		{"als Skript", `{ "setpaket:/tmp!geheim.png" }`},
		{"als Textskript", `setpaket:{"/tmp"}!geheim.png`},
		{"in einem Verbund", `{ linear_blend(image1: "setpaket:/tmp!a.png", image2: "rahmen.png", x1: 0, y1: 0, x2: 1, y2: 1) }`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			baum, befunde := bildangabe(t.Context(), "image", "image", fall.angabe,
				msescript.Bibliothek().Darueber())

			if enthaeltSetpaket(baum) {
				t.Errorf("die Angabe führte zu einem Ort auf der Platte: %+v", baum)
			}

			if len(befunde) == 0 {
				t.Error("kein Befund — eine abgewiesene Angabe wird genannt (§8)")
			}
		})
	}
}

// **Und ein Kartenbild aus dem eigenen Set bleibt erlaubt** — auch über ein
// Skript. `mitBildwerten` legt den Wert eines Bildfelds als Verweis ins
// Set-Paket in die Umgebung; `image: { card.image }` gibt ihn zurück, und das
// ist der Normalfall jeder Vorlage mit einer Illustration.
func TestEinKartenbildAusDemSetBleibtErlaubt(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("card", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: mitBildwerten(
			[]msescript.Eintrag{{Name: "image", Wert: msescript.Text("image1")}},
			[]domain.Field{{Name: "image", Kind: domain.KindImage}},
			"/Orte/Mein Set.mse-set"),
	})

	for _, angabe := range []string{
		`{ card.image }`,
		`{ linear_blend(image1: card.image, image2: "rahmen.png", x1: 0, y1: 0, x2: 1, y2: 1) }`,
	} {
		baum, befunde := bildangabe(t.Context(), "image", "image", angabe, umgebung)
		if !enthaeltSetpaket(baum) || len(befunde) != 0 {
			t.Errorf("%s: das eigene Kartenbild wurde abgewiesen: %+v, %v", angabe, baum, befunde)
		}
	}
}
