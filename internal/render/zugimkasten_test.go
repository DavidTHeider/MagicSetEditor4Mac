package render_test

// **Jeder gesetzte Zug liegt in seinem Kasten**
// (#392).
//
// # Warum hier und nicht über das Pack
//
// Der Zeichenplan trägt die **Angaben** eines Feldes, nicht seine gesetzten
// Zeilen — die entstehen erst beim Zeichnen. Ein Wächter über die Pakete
// müsste den Satz nachbauen und bräuchte dafür Schriften, die
// `internal/service` nach §2 nicht kennen darf.
//
// **Die Zusage gehört ohnehin hierher**: dass ein Zug in seinem Kasten liegt,
// verspricht [render.SetzeAbschnitte], nicht der Planbauer.
//
// # Was geprüft wird, und was nicht
//
// **Waagerecht** beide Kanten: keine Zeile beginnt links vom Kasten oder endet
// rechts davon. Das garantiert der Umbruch.
//
// **Senkrecht nur nach oben.** Nach unten darf ein Zug hinauslaufen — MSE
// kürzt ihn nicht, es meldet ihn
// ([TestMehrTextAlsPlatzWirdGemeldetStattGekuerzt]), und diese Zusage steht
// dort schon.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// zugspiel ist die Toleranz in Vorlagenpunkten.
//
// **Ein halber Punkt**, denn die Kanten entstehen aus Fließkommarechnung; ein
// exakter Vergleich meldete Rundungen als Befund.
const zugspiel = 0.5

func TestJederGesetzteZugLiegtInSeinemKasten(t *testing.T) {
	// **Die Ausrichtungen, die das Pack benutzt**, samt Polster und
	// verschiedener Kastenbreiten — ein Wächter über einen einzigen Fall
	// prüfte den einen Fall.
	ausrichtungen := []render.Ausrichtung{
		render.LinksOben, render.MitteOben, render.RechtsOben,
		render.LinksMitte, render.MitteMitte, render.RechtsMitte,
		render.LinksUnten, render.MitteUnten, render.RechtsUnten,
	}

	texte := []string{
		"kurz",
		"ein Text, der über die Breite läuft und deshalb umbricht",
		"eins\nzwei\ndrei",
		strings.Repeat("unteilbarlangeswort", 3),
	}

	for _, ausrichtung := range ausrichtungen {
		for _, text := range texte {
			for _, breite := range []float64{40, 120, 300} {
				satz := render.Satz{
					Groesse: 10, Ausrichtung: ausrichtung,
					Polster: render.Polster{Links: 6, Rechts: 4, Oben: 4, Unten: 2},
				}
				rahmen := render.Rechteck{Links: 29, Oben: 100, Breite: breite, Hoehe: 60}

				gesetzt := render.SetzeAbschnitte(satz, rahmen,
					[]render.Abschnitt{{Text: text}},
					render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

				for nummer, zeile := range gesetzt.Zeilen {
					pruefeZeile(t, zeile, rahmen, satz, nummer, text, breite)
				}

				// **Senkrecht nur, wenn der Zug hineinpasst.** Ein zu hoher
				// Zug läuft bei unterer Ausrichtung nach **oben** hinaus —
				// dieselbe Regel wie waagerecht, und derselbe Denkfehler im
				// ersten Entwurf.
				if gesetzt.Hoehe <= rahmen.Hoehe-satz.Polster.Oben-satz.Polster.Unten+zugspiel &&
					len(gesetzt.Zeilen) > 0 {
					oben := rahmen.Oben + satz.Polster.Oben
					if gesetzt.Zeilen[0].Oben < oben-zugspiel {
						t.Errorf("Ausrichtung %s, Breite %.0f, %q: beginnt oben bei %.2f, Kasten ab %.2f",
							ausrichtung, breite, kurz(text), gesetzt.Zeilen[0].Oben, oben)
					}
				}
			}
		}
	}
}

// pruefeZeile hält die drei Kanten gegen den Kasten.
func pruefeZeile(
	t *testing.T, zeile render.Zeile, rahmen render.Rechteck, satz render.Satz,
	nummer int, text string, breite float64,
) {
	t.Helper()

	wo := func() string {
		return "Ausrichtung " + string(satz.Ausrichtung) + ", Breite " +
			zahl(breite) + ", Zeile " + zahl(float64(nummer)) + " von " + kurz(text)
	}

	links := rahmen.Links + satz.Polster.Links
	rechts := rahmen.Links + rahmen.Breite - satz.Polster.Rechts
	// **Ein Zug, der nicht hineinpasst, läuft hinaus — und zwar auf der
	// Seite, die seine Ausrichtung freilässt.** MSE bricht nicht mitten im
	// Wort um; wo ein einzelnes Wort breiter ist als der Kasten, steht es
	// über — gemeldet, nicht gekürzt
	// ([TestMehrTextAlsPlatzWirdGemeldetStattGekuerzt]).
	//
	// **Die Ausnahme gilt für beide Kanten.** Der erste Entwurf prüfte die
	// linke unbedingt und fand damit jeden rechtsbündigen Überlauf — der
	// nach links hinausläuft, nicht nach rechts.
	if zeile.Breite > rechts-links+zugspiel {
		return
	}

	if zeile.Links < links-zugspiel {
		t.Errorf("%s: beginnt links bei %.2f, Kasten ab %.2f", wo(), zeile.Links, links)
	}

	if zeile.Links+zeile.Breite > rechts+zugspiel {
		t.Errorf("%s: endet bei %.2f, Kasten bis %.2f",
			wo(), zeile.Links+zeile.Breite, rechts)
	}
}

func zahl(f float64) string {
	ganz := int(f)
	if ganz == 0 {
		return "0"
	}

	ziffern := ""
	for ganz > 0 {
		ziffern = string(rune('0'+ganz%10)) + ziffern
		ganz /= 10
	}

	return ziffern
}

func kurz(text string) string {
	text = strings.ReplaceAll(text, "\n", "⏎")
	if len(text) > 24 {
		return text[:24] + "…"
	}

	return text
}
