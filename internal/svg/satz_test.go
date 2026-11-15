// **Was der Umbruch entscheidet** —
// #852.
//
// Schnitte, Laufweite, Stauchung und die weichen Stücke. Der erste
// SVG-Schreiber setzte jeden Zug links bei `Rahmen.Links` auf eine Grundlinie
// bei `Oben + Groesse`; Ausrichtung, Polster und Umbruch gingen vollständig
// verloren. Gesetzt wird jetzt mit [render.SetzeAbschnitte] — **derselben**
// Maschine wie beim Rastern.
package svg_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// TestEinFehlenderKursivschnittFaelltZurueck prüft [schreiber.schnitte].
//
// **Alle drei Nebenschnitte sind Bestenfalls-Arbeit** — was nicht zu finden
// ist, bleibt leer, und der Satz fällt auf das zurück, was da ist.
func TestEinFehlenderKursivschnittFaelltZurueck(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Kursivschrift = "Helvetica Oblique"

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), plan,
		svg.Quellen{Baeume: baumprobe{}, Schriften: nurAufrecht{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %v", befunde)
	}

	if !strings.Contains(geschrieben, ">Feuerball<") {
		t.Error("der Text fehlt, obwohl nur ein Nebenschnitt fehlte")
	}
}

// nurAufrecht findet die aufrechte Schrift und sonst keine.
type nurAufrecht struct{}

func (nurAufrecht) Suche(name string, fett bool) (render.Schrift, error) {
	if fett || name != "Helvetica" {
		return nil, errors.New("den Schnitt gibt es nicht")
	}

	return schriftprobe{}, nil
}

// TestLaufweiteUndStauchungStehenAmText prüft die beiden Satzangaben.
func TestLaufweiteUndStauchungStehenAmText(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Laufweite = 1.5

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `letter-spacing="1.5"`) {
		t.Errorf("die Laufweite fehlt:\n%s", geschrieben)
	}
}

// TestEineStauchungWirdEineUmformung prüft `shrink`.
//
//	factor = (width - padding_left - padding_right) / layout->width
//
// `src/data/field/text.cpp`, Zeilen 73–81, Fassung f9d9356 vom 2021-07-06.
// **Ein Faktor auf die Breite**, und
// die linke Kante des Feldes bleibt stehen.
func TestEineStauchungWirdEineUmformung(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	textebene := &plan.Ebenen[2]
	textebene.Satz.Stauchen = true
	// **Einzeilig**, wie ein Namensfeld: ein mehrzeiliges bricht seit #2308
	// erst um und staucht nur, wenn danach eine Zeile noch zu breit ist.
	textebene.Satz.Einzeilig = true
	textebene.Rahmen.Breite = 20
	textebene.Abschnitte = []render.Abschnitt{
		{Text: "ein sehr langer Kartenname, der nicht hineinpasst"},
	}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, "scale(") {
		t.Errorf("die Stauchung fehlt:\n%s", geschrieben)
	}
}

// TestEinWeichesStueckWirdNichtGezeichnet hält MSEs `DRAW_ACTIVE`.
func TestEinWeichesStueckWirdNichtGezeichnet(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Abschnitte = []render.Abschnitt{
		{Text: "Kreatur"},
		{Text: " — ", Weich: true},
	}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "—") {
		t.Errorf("das weiche Stück wurde gezeichnet:\n%s", geschrieben)
	}
}

// TestAlleVierSchnitteWerdenGesucht prüft den vollen [schreiber.schnitte].
//
// **Vier und nicht zwei**, weil MSE das fette Gewicht auch auf die Kursive
// legt: bei `<b><i>` nimmt es die Schrift aus `italic name:` **und** das
// Gewicht (`src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom
// 2021-07-06).
func TestAlleVierSchnitteWerdenGesucht(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Kursivschrift = "Helvetica Oblique"

	gefragt := &gemerkteSuche{}

	if _, _, fehler := svg.AusPlan(t.Context(), plan,
		svg.Quellen{Baeume: baumprobe{}, Schriften: gefragt}); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	erwartet := []string{
		"Helvetica aufrecht", "Helvetica fett",
		"Helvetica Oblique aufrecht", "Helvetica Oblique fett",
	}

	for _, schnitt := range erwartet {
		if !slices.Contains(gefragt.namen, schnitt) {
			t.Errorf("%q wurde nicht gesucht — gefragt wurde %v", schnitt, gefragt.namen)
		}
	}
}

// gemerkteSuche hält fest, welche Schnitte gesucht wurden.
type gemerkteSuche struct{ namen []string }

func (g *gemerkteSuche) Suche(name string, fett bool) (render.Schrift, error) {
	schnitt := "aufrecht"
	if fett {
		schnitt = "fett"
	}

	g.namen = append(g.namen, name+" "+schnitt)

	return schriftprobe{}, nil
}

// TestNichtzeichenFallenWeg prüft die XML-Regel an ihrem letzten Zweig.
//
//	Char ::= #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | …
//
// XML 1.0, fünfte Ausgabe, §2.2. U+FFFE und U+FFFF stehen nicht darin.
func TestNichtzeichenFallenWeg(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Abschnitte = []render.Abschnitt{
		{Text: "vor\ufffemitten\uffffnach\tmit Tabulator"},
	}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.ContainsRune(geschrieben, 0xFFFE) || strings.ContainsRune(geschrieben, 0xFFFF) {
		t.Error("ein Nichtzeichen steht im SVG")
	}

	// **Der Tabulator bleibt** — er steht ausdrücklich in der Regel. Er steht
	// dabei in einem eigenen `<tspan>`: [render.SetzeAbschnitte] trennt an
	// Leerraum, und jedes Stück bekommt seine gerechnete Stelle.
	if !strings.ContainsRune(geschrieben, '\t') {
		t.Errorf("der Tabulator wurde mitgestrichen:\n%s", geschrieben)
	}
}

// TestOhneEchteKursiveWirdNichtsErfunden ist die andere Hälfte.
//
// **Ein `font-style="italic"` ohne Kursivschrift lässt den Betrachter eine
// erfinden**, und ihre Vorschübe sind andere als die, mit denen hier gesetzt
// wurde — der Flavortext lief zusammen, gesehen bei der Sichtprüfung an vier
// echten Karten aus `M15-Magic-Pack`.
//
// Wo der Rasterer aufrecht zeichnet, steht im SVG auch keine Kursive.
func TestOhneEchteKursiveWirdNichtsErfunden(t *testing.T) {
	t.Parallel()

	// **Kein `italic name:` im Stil** — dann findet auch der Rasterer keine.
	geschrieben, _, fehler := svg.AusPlan(t.Context(), probeplan(), mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, `font-style="italic"`) {
		t.Errorf("eine Kursive wurde erfunden:\n%s", geschrieben)
	}

	// **Und mit einem Namen, den die Suche nicht kennt, ebenso wenig.**
	plan := probeplan()
	plan.Ebenen[2].Satz.Kursivschrift = "Gibtesnicht Italic"

	geschrieben, _, fehler = svg.AusPlan(t.Context(), plan,
		svg.Quellen{Baeume: baumprobe{}, Schriften: nurAufrecht{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "Gibtesnicht") {
		t.Errorf("ein Schnitt, den es nicht gibt, steht im SVG:\n%s", geschrieben)
	}

	// **Und fett ebenso**: `nurAufrecht` gibt kein Gewicht her.
	if strings.Contains(geschrieben, `font-weight="bold"`) {
		t.Errorf("ein fetter Schnitt wurde erfunden:\n%s", geschrieben)
	}
}
