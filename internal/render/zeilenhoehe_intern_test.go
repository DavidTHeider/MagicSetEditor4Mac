package render

import (
	"image"
	"math"
	"testing"
)

// Die Zeilenhöhe rastet auf **Viertelpunkte**, nicht auf ganze
// (#546).
//
// # Woher die Regel kommt
//
// MSE setzt die Schrift nicht in der Größe, die im Stil steht, sondern in der
// **vierfachen**:
//
//	dc.SetFont(font.toWxFont(trS(scale) * (quality == QUALITY_LOW ? 1 : text_scaling)));
//
// `src/util/rotation.cpp`, Zeile 326, mit `text_scaling = 4`
// (`src/gfx/resample_text.cpp`, Zeile 22). `Font::toWxFont` macht daraus eine
// **ganze** Punktgröße (`to_int`, `src/data/font.cpp`, Zeile 84), misst damit,
// und `RotatedDC::GetTextExtent` teilt die — ebenfalls ganzzahlige — Höhe
// wieder durch `zoomY * text_scaling` (Zeile 349).
//
// Beim Bildexport ist `zoom = 1` (`UnzoomedDataViewer::getRotation`,
// `src/data/format/image.cpp`, Zeilen 37–45). In Vorlagenpunkten bleibt damit
//
//	ceil(verhaeltnis * round(4 * groesse)) / 4
//
// Alles Fassung f9d9356 vom 2021-07-06.
//
// # Warum nicht `ceil(verhaeltnis * groesse)`
//
// **Das war der erste Anlauf, und er ist gemessen widerlegt.** Am 2026-08-19
// umgesetzt und zurückgenommen: er machte den Abstand von 22,542 auf 26,299,
// also um **3,757 schlechter**. Der Grund ist der Faktor vier — gerundet wird
// bei Vierfachgröße. Für `MPlantin` bei Größe 14 sind das 16,25 statt 17; der
// erste Anlauf lag 0,75 Punkte je Zeile daneben.
//
// # Was gemessen ist, und was nicht
//
// Dass wx' Texthöhe `ceil(verhaeltnis * punktgroesse)` ist, ist am 2026-08-20
// gegen **wx 3.3.3** an **zwanzig** Punktgrößen von 9 bis 120 gemessen, für
// `mplantin` (1,149) und `beleren-bold` (1,205566): **40 von 40**. Die
// Vierfachgrößen sind darin enthalten — 56, 96 und 120 sind das Vierfache von
// 14, 24 und 30.
//
// **Nicht gemessen ist die Grundlinie.** Sie folgt nicht derselben Rechnung
// und steht als eigene Frage an
// #446.
func TestZeilenhoeheRastetAufViertelpunkte(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		verhaeltnis float64
		groesse     float64
		erwartet    float64
	}{
		{
			name:        "MPlantin bei 14 — 16,25 und nicht 17",
			verhaeltnis: 1.149,
			groesse:     14,
			erwartet:    16.25,
		},
		{
			name:        "MPlantin bei 9",
			verhaeltnis: 1.149,
			groesse:     9,
			erwartet:    10.5,
		},
		{
			name:        "Beleren bei 12,5 — eine Größe im Viertelraster",
			verhaeltnis: 1.205566,
			groesse:     12.5,
			erwartet:    15.25,
		},
		{
			name:        "die Größe rastet mit — 14,1 zählt wie 14",
			verhaeltnis: 1.149,
			groesse:     14.1,
			erwartet:    16.25,
		},
		{
			name:        "und 14,2 zählt schon wie 14,25",
			verhaeltnis: 1.149,
			groesse:     14.2,
			erwartet:    16.5,
		},
		{
			name:        "ein glattes Verhältnis bleibt, wie es war",
			verhaeltnis: 1.2,
			groesse:     10,
			erwartet:    12,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			satz := Satz{Groesse: fall.groesse}
			schriften := Schriftsatz{Aufrecht: masshalter{fall.verhaeltnis}}

			hoehe := zeilenhoehe(nil, satz, schriften, nil)

			if math.Abs(hoehe-fall.erwartet) > 1e-9 {
				t.Errorf("Zeilenhöhe = %v, erwartet %v", hoehe, fall.erwartet)
			}
		})
	}
}

// masshalter ist eine Schrift, die nur ihr Höhenverhältnis kennt.
type masshalter struct{ verhaeltnis float64 }

func (masshalter) Breite(rune) float64    { return 1 }
func (m masshalter) Zeilenhoehe() float64 { return m.verhaeltnis }

// Die **Grundlinie** rastet auf dasselbe Viertel wie die Zeilenhöhe.
//
// # Gemessen, nicht abgeleitet
//
// MSE zeichnet eine Textzeile bei ihrer Oberkante
// (`FontTextElement::draw` → `dc.DrawTextWithShadow(text, *font, rect.position())`,
// `src/render/text/font.cpp`, Zeilen 15–25, Fassung f9d9356 vom 2021-07-06)
// und überlässt wx die Grundlinie. **Wo wx sie hinlegt, ist am 2026-08-20
// gemessen** — ein „H" bei y = 0 in eine Fläche gezeichnet und die unterste
// Tintenzeile gesucht, für fünf Punktgrößen von 14 bis 120:
//
//	MPlantin  ceil(0,774·p)  5/5
//	Matrix    ceil(0,633·p)  5/5   (`OS/2` typo, denn `hhea` ist leer — #597)
//
// **Erst die Messung hat es entschieden.** Aus `GetTextExtent` abgeleitet
// (`h − descent`) käme bei MPlantin und Größe 14 eine Grundlinie von 13,0
// heraus; gemessen sind **11,0**. Die Ableitung war plausibel und falsch.
func TestGrundlinieRastetWieDieZeilenhoehe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		verhaeltnis float64
		groesse     float64
		erwartet    float64
	}{
		{
			name:        "MPlantin bei 14 — 11,0 und nicht 10,836",
			verhaeltnis: 0.774,
			groesse:     14,
			erwartet:    11,
		},
		{
			name:        "MPlantin bei 30",
			verhaeltnis: 0.774,
			groesse:     30,
			erwartet:    23.25,
		},
		{
			name:        "Matrix bei 14",
			verhaeltnis: 0.633,
			groesse:     14,
			erwartet:    9,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			zeile := Zeile{Oben: 100}
			satz := Satz{Groesse: fall.groesse}

			hoehe := zeilengrundlinie(zeile, satz, grundlinienschrift{fall.verhaeltnis})

			if math.Abs(hoehe-(100+fall.erwartet)) > 1e-9 {
				t.Errorf("Grundlinie = %v, erwartet %v", hoehe-100, fall.erwartet)
			}
		})
	}
}

// grundlinienschrift ist eine Glyphenschrift, die nur ihre Grundlinie kennt.
type grundlinienschrift struct{ verhaeltnis float64 }

func (grundlinienschrift) Breite(rune) float64   { return 1 }
func (grundlinienschrift) Zeilenhoehe() float64  { return 1 }
func (g grundlinienschrift) Grundlinie() float64 { return g.verhaeltnis }
func (grundlinienschrift) Maske(rune, float64) (image.Image, image.Point, bool) {
	return nil, image.Point{}, false
}

// [mseGanzzahl] rundet auch nach unten — die Seite, die dieses Programm nie
// betritt.
//
// # Warum sie trotzdem dasteht und geprüft wird
//
// `to_int` ist symmetrisch:
//
//	return static_cast<int>(d > 0 ? d + 0.4999995 : d - 0.4999995);
//
// `src/util/vector2d.hpp`, Zeilen 20–22, Fassung f9d9356 vom 2021-07-06. Eine
// Schriftgröße ist hier nie negativ; die Hälfte wegzulassen, machte aus einer
// **Portierung** eine Auslegung, und die nächste Verwendung dieser Funktion
// bekäme stillschweigend eine andere Regel.
//
// **Und die 0,4999995 sind Absicht.** Bei genau einem Halben rundet `to_int`
// zur Null hin, `math.Round` von ihr weg — deshalb steht `−2,5 → −2` hier als
// eigener Fall.
func TestMSEGanzzahlRundetZurNullHin(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		wert, erwartet float64
	}{
		{wert: 0, erwartet: 0},
		{wert: 2.4, erwartet: 2},
		{wert: 2.5, erwartet: 2},
		{wert: 2.6, erwartet: 3},
		{wert: -2.4, erwartet: -2},
		{wert: -2.5, erwartet: -2},
		{wert: -2.6, erwartet: -3},
	}

	for _, fall := range faelle {
		if ergebnis := mseGanzzahl(fall.wert); ergebnis != fall.erwartet {
			t.Errorf("mseGanzzahl(%v) = %v, erwartet %v",
				fall.wert, ergebnis, fall.erwartet)
		}
	}
}
