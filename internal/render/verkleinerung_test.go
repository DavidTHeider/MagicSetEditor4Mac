package render_test

// **`scale down to:` — ein zu langer Text wird kleiner gesetzt**
// (#498).
//
//	double FontTextElement::minScale() const {
//	  return min(font->size(), font->scale_down_to) / max(0.01, font->size());
//	}
//
// `src/render/text/font.cpp`, Zeilen 51–53. MSE sucht dann den größten
// Maßstab zwischen dieser Schranke und 1, bei dem der Text in den Kasten
// passt — `prepareLinesTryScales`, `src/render/text/viewer.cpp`, Zeilen
// 444–500. Fassung f9d9356 vom 2021-07-06.
//
// # Woran es aufgefallen ist
//
// `Ward of Lights` aus dem kanonischen Set `White` lag bei **48,342** — der
// größte Abstand der 510 verglichenen Karten. Ihr Regeltext lief nach oben aus
// dem Kasten heraus und überschrieb die Typzeile; MSE2 setzt denselben Text
// kleiner und bringt ihn unter.
//
// **571 Angaben in 174 der 190 Stilpakete** von `M15-Magic-Pack`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinZuLangerTextWirdKleinerGesetzt(t *testing.T) {
	// Sieben Zeilen Text in einem Kasten, der bei voller Größe drei fasst.
	abschnitte := []render.Abschnitt{{
		Text: "Eins zwei drei vier fünf sechs sieben acht neun zehn elf " +
			"zwölf dreizehn vierzehn fünfzehn sechzehn siebzehn achtzehn",
	}}

	rahmen := render.Rechteck{Breite: 60, Hoehe: 30}

	t.Run("ohne scale down to läuft er über", func(t *testing.T) {
		gesetzt := render.SetzeAbschnitte(
			render.Satz{Schrift: "Probe", Groesse: 10},
			rahmen, abschnitte, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

		if gesetzt.Hoehe <= rahmen.Hoehe {
			t.Fatalf("Höhe %v passt in %v — der Fall greift nicht",
				gesetzt.Hoehe, rahmen.Hoehe)
		}

		if gesetzt.Groesse != 10 {
			t.Errorf("Größe %v, erwartet 10 — ohne Angabe wird nicht verkleinert",
				gesetzt.Groesse)
		}
	})

	t.Run("mit scale down to passt er", func(t *testing.T) {
		gesetzt := render.SetzeAbschnitte(
			render.Satz{Schrift: "Probe", Groesse: 10, Kleinstgroesse: 4},
			rahmen, abschnitte, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

		if gesetzt.Hoehe > rahmen.Hoehe {
			t.Errorf("Höhe %v passt nicht in %v", gesetzt.Hoehe, rahmen.Hoehe)
		}

		if gesetzt.Groesse >= 10 || gesetzt.Groesse < 4 {
			t.Errorf("Größe %v, erwartet zwischen 4 und 10", gesetzt.Groesse)
		}
	})

	t.Run("die Schranke wird nicht unterschritten", func(t *testing.T) {
		// **Passt es auch bei der kleinsten Größe nicht, bleibt es dabei.**
		// MSE nimmt dann den kleinsten Maßstab und lässt überlaufen — ein
		// Text, der bei Größe 9 nicht in einen Kasten von 30 passt, passt bei
		// 9 auch nicht.
		gesetzt := render.SetzeAbschnitte(
			render.Satz{Schrift: "Probe", Groesse: 10, Kleinstgroesse: 9},
			rahmen, abschnitte, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

		if gesetzt.Groesse < 9 {
			t.Errorf("Größe %v, erwartet mindestens 9", gesetzt.Groesse)
		}
	})

	t.Run("ein kurzer Text bleibt groß", func(t *testing.T) {
		gesetzt := render.SetzeAbschnitte(
			render.Satz{Schrift: "Probe", Groesse: 10, Kleinstgroesse: 4},
			rahmen, []render.Abschnitt{{Text: "kurz"}},
			render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

		if gesetzt.Groesse != 10 {
			t.Errorf("Größe %v, erwartet 10 — er passt schon", gesetzt.Groesse)
		}
	})
}
