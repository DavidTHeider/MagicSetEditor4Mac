package render

import (
	"encoding/binary"
	"image"
	"slices"
	"strings"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

// **Eine Ligatur geht durch jeden Weg, der misst oder zeichnet** —
// #1805.
//
// Die Schrift ist Go Regular aus `golang.org/x/image` mit einem erfundenen
// `liga`: `fi` wird zur Glyphe von `W`. Erfunden, damit das Ergebnis bekannt
// ist — Gestalt und Breite von `W` —, und Go Regular, weil sie mit der
// Bibliothek kommt, die ohnehin rastert: die Prüfung hängt an keiner
// Schriftdatei dieses Rechners.

func TestEineLigaturWirdEinGlyphenzeichen(t *testing.T) {
	t.Parallel()

	mit, w := goregularMitLigatur(t)

	ohne, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	for _, fall := range []struct {
		name    string
		schrift *OpenTypeSchrift
		text    string
		gesetzt []rune
	}{
		{"fi wird eine Glyphe", mit, "fif", []rune{glyphenzeichen(w), 'f'}},
		{"ohne i bleibt es beim f", mit, "ff", []rune("ff")},
		{"zwei Ligaturen", mit, "fi fi", []rune{glyphenzeichen(w), ' ', glyphenzeichen(w)}},
		{"ein fehlendes Zeichen dazwischen trennt", mit, "fi", []rune("fi")},
		{"eine Schrift ohne GSUB ersetzt nichts", ohne, "fi", []rune("fi")},
		{"leer", mit, "", []rune{}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := fall.schrift.Ligiere([]rune(fall.text)); !slices.Equal(got, fall.gesetzt) {
				t.Errorf("%q → %U, erwartet %U", fall.text, got, fall.gesetzt)
			}
		})
	}
}

func TestEinGlyphenzeichenMisstUndZeichnetWieSeineGlyphe(t *testing.T) {
	t.Parallel()

	mit, w := goregularMitLigatur(t)
	zeichen := glyphenzeichen(w)

	if got, want := mit.Breite(zeichen), mit.Breite('W'); got != want || got == 0 {
		t.Errorf("Breite %v, die von W ist %v", got, want)
	}

	if got, want := mit.Kernung(zeichen, 'a'), mit.Kernung('W', 'a'); got != want {
		t.Errorf("Kernung %v, die von W ist %v", got, want)
	}

	maske, ecke, da := mit.Maske(zeichen, 40)
	wMaske, wEcke, wDa := mit.Maske('W', 40)

	if !da || !wDa || ecke != wEcke || !gleicheDeckung(maske, wMaske) {
		t.Errorf("die Maske des Glyphenzeichens gleicht der von W nicht: da %v/%v, Ecke %v/%v", da, wDa, ecke, wEcke)
	}

	// **Eine Nummer jenseits der Schrift ist kein Zeichen** — sie misst null
	// und zeichnet nichts, wie ein fehlendes.
	jenseits := glyphenzeichen(0xFFFF)
	if _, _, da := mit.Maske(jenseits, 40); da || mit.Breite(jenseits) != 0 || mit.Kernung(jenseits, 'W') != 0 {
		t.Error("eine Glyphennummer jenseits der Schrift misst oder zeichnet")
	}
}

func TestDieLigaturGehtInJedeBreite(t *testing.T) {
	t.Parallel()

	mit, _ := goregularMitLigatur(t)
	satz := Satz{Groesse: 10}

	for name, breiten := range map[string][2]float64{
		"Textbreite": {Textbreite("fi", 10, mit), Textbreite("W", 10, mit)},
		"breiteVon":  {breiteVon("fi", satz, mit), breiteVon("W", satz, mit)},
		"Vorsilbe":   {(&vorsilbe{}).nimm("fi", satz, mit), (&vorsilbe{}).nimm("W", satz, mit)},
	} {
		if breiten[0] != breiten[1] || breiten[0] == 0 {
			t.Errorf("%s: fi misst %v, W %v", name, breiten[0], breiten[1])
		}
	}

	// **Mit Laufweite keine Ligatur** — der Abstand gehört zwischen die
	// Buchstaben, und eine Ligatur hat keinen Zwischenraum.
	gesperrt := Satz{Groesse: 10, Laufweite: 1}
	einzeln := mit.Breite('f')*10 + mit.Kernung('f', 'i')*10 + mit.Breite('i')*10 + 1

	if got := breiteVon("fi", gesperrt, mit); got != einzeln {
		t.Errorf("gesperrt misst fi %v, einzeln wären es %v", got, einzeln)
	}
}

func TestEineLigaturWirdGezeichnet(t *testing.T) {
	t.Parallel()

	mit, _ := goregularMitLigatur(t)

	ohne, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	fi := gezeichneteZeile(t, mit, "fi")

	if w := gezeichneteZeile(t, mit, "W"); !gleicheDeckung(fi, w) {
		t.Error("fi ist nicht als die Glyphe der Ligatur gezeichnet")
	}

	// **Die Gegenprobe steht im Test**: ohne die Ligatur ist es ein anderes
	// Bild — sonst bewiese die Gleichheit oben nichts.
	if gleicheDeckung(gezeichneteZeile(t, ohne, "fi"), fi) {
		t.Error("fi ohne Ligatur sieht aus wie mit — der Test sieht die Ligatur nicht")
	}
}

// **Auch der Text auf einem Symbol** — [eingepasst] misst ihn mit
// [Textbreite], und die setzt die Ligatur; zeichnete der Symbolzug sie nicht,
// stünde ein Text in einer Breite, die nicht seine ist.
func TestDerSymboltextSetztDieLigatur(t *testing.T) {
	t.Parallel()

	mit, _ := goregularMitLigatur(t)

	ohne, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	fi := gezeichnetesSymbol(t, mit, "fi")

	if w := gezeichnetesSymbol(t, mit, "W"); !gleicheDeckung(fi, w) {
		t.Error("fi auf dem Symbol ist nicht als die Glyphe der Ligatur gezeichnet")
	}

	if gleicheDeckung(gezeichnetesSymbol(t, ohne, "fi"), fi) {
		t.Error("fi ohne Ligatur sieht aus wie mit — der Test sieht die Ligatur nicht")
	}
}

// gezeichnetesSymbol zeichnet ein Symbol mit Text auf weißem Grund.
func gezeichnetesSymbol(t *testing.T, schrift Schrift, text string) *image.NRGBA {
	t.Helper()

	const blatt = "/probe.mse-symbol-font/kreis.png"

	weiss := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	for i := range weiss.Pix {
		weiss.Pix[i] = 255
	}

	plan := &Plan{Breite: 60, Hoehe: 60, Ebenen: []Ebene{
		{
			Art: ArtFlaeche, Rahmen: Rechteck{Breite: 60, Hoehe: 60},
			Farbe: Farbe{R: 255, G: 255, B: 255, A: 255}, Deckung: 1,
		},
		{
			Art: ArtSymbolzug, Rahmen: Rechteck{Breite: 60, Hoehe: 60}, Deckung: 1,
			Satz: Satz{Groesse: 40},
			Abschnitte: []Abschnitt{{
				Symbolmasse: Symbolmasse{Groesse: 40},
				Symbole: []Symbol{{
					Bild: Blatt(blatt), Text: text,
					Textsatz: Symboltextsatz{Schrift: "Go", Groesse: 0.6, Dehnung: 0.5, Farbe: Farbe{A: 255}},
				}},
			}},
		},
	}}

	bild, befunde, err := NeuerRasterer(einBild{blatt: weiss}).MitSchriften(eineSchriftquelle{schrift}).
		Zeichne(t.Context(), plan, 1, GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("zeichnen: %v %v", err, befunde)
	}

	return bild
}

// einBild gibt für einen Pfad ein Bild.
type einBild map[string]image.Image

func (b einBild) Bild(pfad string) (image.Image, error) { return b[pfad], nil }

// goregularMitLigatur ist Go Regular mit einem `liga`, das `fi` zur Glyphe
// von `W` zieht. Es gibt die Schrift und die Nummer von `W`.
func goregularMitLigatur(t *testing.T) (mit *OpenTypeSchrift, w uint16) {
	t.Helper()

	ohne, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	nummer := func(zeichen rune) uint16 {
		kennung, fehler := ohne.schrift.GlyphIndex(&ohne.puffer, zeichen)
		if fehler != nil || kennung == 0 {
			t.Fatalf("Go Regular führt %q nicht", zeichen)
		}

		return uint16(kennung)
	}

	f, i := nummer('f'), nummer('i')
	w = nummer('W')

	gsub := gsubMit([]testmerkmal{{"liga", []int{0}}}, []testnachschlag{{4, [][]byte{
		ligaturuntertabelle(false, map[uint16][][]uint16{f: {{w, i}}}),
	}}})

	mit, err = SchriftAusDatei(mitEingesetzterTabelle(goregular.TTF, "GSUB", gsub))
	if err != nil {
		t.Fatalf("Go Regular mit GSUB: %v", err)
	}

	return mit, w
}

// mitEingesetzterTabelle legt eine Tabelle in eine echte Schriftdatei — das
// Verzeichnis neu und nach Kennung sortiert, jede Tabelle auf vier Byte
// ausgerichtet, wie `sfnt` es verlangt.
func mitEingesetzterTabelle(schrift []byte, kennung string, tabelle []byte) []byte {
	type eintrag struct {
		kennung string
		daten   []byte
	}

	anzahl := int(binary.BigEndian.Uint16(schrift[4:]))

	eintraege := make([]eintrag, 0, anzahl+1)
	eintraege = append(eintraege, eintrag{kennung, tabelle})

	for i := range anzahl {
		stelle := 12 + 16*i
		versatz := binary.BigEndian.Uint32(schrift[stelle+8:])
		laenge := binary.BigEndian.Uint32(schrift[stelle+12:])

		eintraege = append(eintraege, eintrag{string(schrift[stelle : stelle+4]), schrift[versatz : versatz+laenge]})
	}

	slices.SortFunc(eintraege, func(a, b eintrag) int { return strings.Compare(a.kennung, b.kennung) })

	aus := make([]byte, 12+16*len(eintraege))
	copy(aus, schrift[:4])
	binary.BigEndian.PutUint16(aus[4:], uint16(len(eintraege))) //nolint:gosec // Testdaten

	for i, e := range eintraege {
		for len(aus)%4 != 0 {
			aus = append(aus, 0)
		}

		stelle := 12 + 16*i
		copy(aus[stelle:], e.kennung)
		binary.BigEndian.PutUint32(aus[stelle+8:], uint32(len(aus)))      //nolint:gosec // Testdaten
		binary.BigEndian.PutUint32(aus[stelle+12:], uint32(len(e.daten))) //nolint:gosec // Testdaten

		aus = append(aus, e.daten...)
	}

	return aus
}

// gezeichneteZeile zeichnet einen Text in einer Zeile, bei 20 Punkt.
func gezeichneteZeile(t *testing.T, schrift Schrift, text string) *image.NRGBA {
	t.Helper()

	plan := &Plan{Breite: 80, Hoehe: 40, Ebenen: []Ebene{{
		Art:        ArtText,
		Rahmen:     Rechteck{Links: 4, Oben: 4, Breite: 72, Hoehe: 32},
		Abschnitte: Aufrecht(text),
		Farbe:      Farbe{A: 255},
		Satz:       Satz{Schrift: "Go", Groesse: 20, Ausrichtung: LinksOben},
		Deckung:    1,
	}}}

	bild, befunde, err := NeuerRasterer(nil).MitSchriften(eineSchriftquelle{schrift}).
		Zeichne(t.Context(), plan, 1, GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("zeichnen: %v %v", err, befunde)
	}

	return bild
}

// eineSchriftquelle findet für jeden Namen dieselbe Schrift.
type eineSchriftquelle struct{ schrift Schrift }

func (q eineSchriftquelle) Suche(string, bool) (Schrift, error) { return q.schrift, nil }

// gleicheDeckung sagt, ob zwei Bilder dieselbe Deckung tragen — Rahmen und
// jeder Bildpunkt.
func gleicheDeckung(a, b image.Image) bool {
	if a == nil || b == nil || a.Bounds() != b.Bounds() {
		return false
	}

	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}
