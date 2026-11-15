package svg

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// text setzt eine Textebene und schreibt sie als `<text>`.
//
// # Dieselbe Umbruchmaschine wie der Rasterer
//
// [render.SetzeAbschnitte] — nicht eine zweite daneben. Der Umbruch entscheidet
// über Zeilen, Ausrichtung, Polster, Verkleinerung und Stauchung; zwei
// Rechnungen darüber laufen auseinander, sobald eine etwas lernt, und dann
// stünde derselbe Text im PNG und im SVG an verschiedenen Stellen.
//
// # Und dieselben zwei Befunde
//
// Eine fehlende Schriftsuche und ein leerer Schriftname werden **benannt**, so
// wie in `render.text`: eine Karte mit falscher Schrift sieht richtig aus und
// ist es nicht.
func (s *schreiber) text(
	ctx context.Context, aus *strings.Builder, plan *render.Plan,
	nummer int, ebene, innen render.Ebene,
) {
	if s.quellen.Schriften == nil {
		s.melde(nummer, ebene, "für Text ist keine Schriftsuche eingerichtet")

		return
	}

	name := strings.TrimSpace(ebene.Satz.Schrift)
	if name == "" {
		s.melde(nummer, ebene, "eine Textebene ohne Schriftnamen")

		return
	}

	schrift, err := s.quellen.Schriften.Suche(name, false)
	if err != nil {
		s.melde(nummer, ebene, "die Schrift "+name+" ist nicht zu finden: "+err.Error())

		return
	}

	kursivname := strings.TrimSpace(ebene.Satz.Kursivschrift)
	schriften := s.schnitte(schrift, name, kursivname)
	gesetzt := render.SetzeAbschnitte(innen.Satz, innen.Rahmen, innen.Abschnitte,
		schriften, s.quellen.Symbole)

	// **Ein Zug mit Symbolen geht durch den Rasterer.** Wohin ein `<sym>`
	// mitten im Regeltext kommt, weiß erst der Umbruch; die Symbolmaschine ein
	// zweites Mal zu bauen wäre die zweite Rechnung über dieselbe Sache.
	if traegtSymbole(gesetzt) {
		s.alsRaster(ctx, aus, plan, nummer, ebene, innen)

		return
	}

	s.zeilen(aus, innen, gesetzt, schnittnamen{
		aufrecht:  name,
		kursiv:    kursivname,
		hatKursiv: schriften.Kursiv != nil,
		hatFett:   schriften.Fett != nil,
	})
}

// schnittnamen sagt, welche Schnitte es **wirklich** gibt.
//
// # Warum das im SVG stehen muss
//
// **Ein Betrachter, der `font-style="italic"` sieht, erfindet eine Kursive.**
// Er neigt die aufrechten Glyphen, und die haben andere Vorschübe als die
// echte Kursive — der Satz hier rechnet aber mit den Maßen der Schrift aus
// `italic name:`. Das Ergebnis waren zusammenlaufende Wörter im Flavortext,
// gesehen bei der Sichtprüfung an vier echten Karten.
//
// **Der Rasterer erfindet ebenso wenig**: [render.Schriftsatz] fällt auf die
// aufrechte Schrift zurück, wenn der Schnitt fehlt (`schnitte` oben, und MSEs
// eigener Weg über `src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom
// 2021-07-06). Wo er aufrecht zeichnet, steht im SVG auch keine Kursive.
type schnittnamen struct {
	aufrecht, kursiv   string
	hatKursiv, hatFett bool
}

// traegtSymbole sagt, ob im gesetzten Zug ein Symbol steht.
func traegtSymbole(gesetzt render.Satzergebnis) bool {
	for _, zeile := range gesetzt.Zeilen {
		for _, stueck := range zeile.Stuecke {
			if len(stueck.Symbole) > 0 {
				return true
			}
		}
	}

	return false
}

// schnitte sucht die drei Nebenschnitte zur aufrechten Schrift.
//
// **Dieselbe Bestenfalls-Arbeit wie in `render.schnitte`**: was nicht zu
// finden ist, bleibt leer, und [render.Schriftsatz] fällt auf das zurück, was
// da ist. Ein Flavortext, der aufrecht steht, ist eine Karte; einer, der
// fehlt, ist keine (§8).
func (s *schreiber) schnitte(aufrecht render.Schrift, name, kursivname string) render.Schriftsatz {
	schriften := render.Schriftsatz{Aufrecht: aufrecht}

	if fett, err := s.quellen.Schriften.Suche(name, true); err == nil {
		schriften.Fett = fett
	}

	schriften.Kursiv, schriften.KursivFett = render.Kursivschnitte(s.quellen.Schriften, name, kursivname)

	return schriften
}

// zeilen schreibt den gesetzten Zug als `<text>` mit einem `<tspan>` je Stück.
//
// # Die Grundlinie entscheidet der Betrachter
//
// **`dominant-baseline="text-before-edge"` und keine gerechnete Grundlinie.**
// [render.SetzeAbschnitte] gibt die **Oberkante** einer Zeile; der Rasterer
// rechnet daraus über `schrift.Grundlinie()` die Grundlinie und rastet sie
// dabei auf Viertelpunkte (MSEs `text_scaling = 4`). Ein Betrachter hat
// dieselbe Schriftdatei und tut es nicht.
//
// Die Abweichung ist damit **höchstens ein Viertelpunkt je Zeile** und
// benannt; die Grundlinie hier ein zweites Mal zu rechnen wäre die Rechnung,
// die auseinanderläuft — und sie ginge nur mit einer [render.Glyphenschrift],
// während gesetzt wird schon mit einer bloßen Breitentabelle.
//
// # Die Stauchung wirkt als Umformung
//
//	factor = (width - padding_left - padding_right) / layout->width
//
// `src/data/field/text.cpp`, Zeilen 73–81, Fassung f9d9356 vom 2021-07-06 —
// **ein Faktor auf die Breite, nicht auf die Größe**, und die linke Kante des
// Feldes bleibt stehen. In SVG ist das `translate(kante,0) scale(f,1)
// translate(-kante,0)`; sie wirkt damit auf die Stellen **und** auf die
// Glyphenbreiten, so wie beim Rastern.
func (s *schreiber) zeilen(
	aus *strings.Builder, ebene render.Ebene, gesetzt render.Satzergebnis,
	namen schnittnamen,
) {
	if schatten := ebene.Satz.Schatten; schatten.Farbe.A > 0 {
		s.schatten(aus, ebene, gesetzt, namen, schatten)
	}

	s.textelement(aus, ebene, gesetzt, namen, ebene.Farbe, true)
}

// schatten schreibt den Schatten der Schrift **vor** den Text — derselbe Text
// in der Farbe des Schattens, um seinen Versatz verschoben und für einen
// Vorleser verborgen ([render.Textschatten]).
//
// **Die Weichheit ist hier eine Näherung** (adapted): MSE weicht `n`-mal mit
// dem Kern (2·Mitte + vier Nachbarn) / 6 an Ort und Stelle
// (`blur_image_alpha`, `src/gfx/resample_text.cpp`, Zeilen 144–162, Fassung
// f9d9356 vom 2021-07-06). Seine Varianz ist je Achse ein Drittel je
// Durchgang; ein Gaußfilter mit σ = √(n/3) kommt ihm am nächsten. Und **alle
// Schatten eines Feldes liegen unter all seinem Text** — MSE legt Schatten und
// Text je Stück abwechselnd.
func (s *schreiber) schatten(
	aus *strings.Builder, ebene render.Ebene, gesetzt render.Satzergebnis,
	namen schnittnamen, schatten render.Textschatten,
) {
	fmt.Fprintf(aus, `<g transform="translate(%s,%s)" aria-hidden="true"`, zahl(schatten.X), zahl(schatten.Y))

	if weich := int(schatten.Weichheit); weich > 0 {
		kennung := s.name("schatten")
		fmt.Fprintf(&s.vorlagen, `<filter id="%s" x="-0.5" y="-0.5" width="2" height="2">`+
			`<feGaussianBlur stdDeviation="%s"/></filter>`, kennung, zahl(math.Sqrt(float64(weich)/3)))
		fmt.Fprintf(aus, ` filter="url(#%s)"`, kennung)
	}

	aus.WriteString(">")
	s.textelement(aus, ebene, gesetzt, namen, schatten.Farbe, false)
	aus.WriteString("</g>")
}

// textelement schreibt ein `<text>` mit allen Stücken in einer Farbe —
// `feld` sagt, ob es das Feld trägt, und das tut nur der Text, nicht sein
// Schatten.
func (s *schreiber) textelement(
	aus *strings.Builder, ebene render.Ebene, gesetzt render.Satzergebnis,
	namen schnittnamen, farbe render.Farbe, feld bool,
) {
	aus.WriteString("<text")

	if feld {
		fmt.Fprintf(aus, ` data-feld="%s"`, entschaerfe(ebene.Feld))
	}

	fmt.Fprintf(aus, ` font-family="%s" font-size="%s" `+
		`fill="%s" dominant-baseline="text-before-edge" xml:space="preserve"`,
		entschaerfe(namen.aufrecht), zahl(gesetzt.Groesse), alsFarbe(farbe))

	// **Die Laufweite steht am `<text>`, nicht am Stück.** MSE trägt sie am
	// Satz und nicht am Abschnitt ([render.Satz.Laufweite]).
	if ebene.Satz.Laufweite != 0 {
		fmt.Fprintf(aus, ` letter-spacing="%s"`, zahl(ebene.Satz.Laufweite))
	}

	if gesetzt.Stauchung != 1 {
		kante := ebene.Rahmen.Links
		fmt.Fprintf(aus, ` transform="translate(%s,0) scale(%s,1) translate(%s,0)"`,
			zahl(kante), zahl(gesetzt.Stauchung), zahl(-kante))
	}

	aus.WriteString(">")

	for _, zeile := range gesetzt.Zeilen {
		for _, stueck := range zeile.Stuecke {
			// **Ein weiches Stück wird gemessen, aber nicht gezeichnet** —
			// `DRAW_ACTIVE` setzt allein der Editor
			// (`src/render/text/element.cpp`, Zeile 234, und
			// `src/data/draw_what.hpp`, Zeile 23, Fassung f9d9356 vom
			// 2021-07-06).
			if stueck.Weich {
				continue
			}

			s.stueck(aus, zeile, stueck, namen)
		}
	}

	aus.WriteString("</text>")
}

// stueck schreibt ein einzelnes `<tspan>` an seiner gerechneten Stelle.
//
// **Die Kursive wird benannt, nicht angesagt** — siehe [schnittnamen]. Ein
// `font-style="italic"` ohne echte Kursive lässt den Betrachter eine erfinden,
// und ihre Vorschübe sind andere als die, mit denen hier gesetzt wurde.
func (s *schreiber) stueck(
	aus *strings.Builder, zeile render.Zeile, stueck render.Stueck, namen schnittnamen,
) {
	fmt.Fprintf(aus, `<tspan x="%s" y="%s"`, zahl(stueck.Links), zahl(zeile.Oben))

	if stueck.Kursiv && namen.hatKursiv {
		fmt.Fprintf(aus, ` font-family="%s"`, entschaerfe(namen.kursiv))
	}

	// **Fett bleibt ein Gewicht.** MSE fragt die Plattform danach
	// (`src/data/font.cpp`, Zeilen 82–108, Fassung f9d9356 vom 2021-07-06),
	// und ein Betrachter tut dasselbe — hier wird nichts erfunden, was es
	// nicht gibt. Fehlt der Schnitt, setzt auch der Rasterer aufrecht.
	if stueck.Fett && namen.hatFett {
		aus.WriteString(` font-weight="bold"`)
	}

	fmt.Fprintf(aus, ">%s</tspan>", entschaerfe(stueck.Text))
}
