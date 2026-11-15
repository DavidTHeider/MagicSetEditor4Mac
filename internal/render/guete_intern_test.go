package render

import (
	"image"
	"image/color"
	"testing"
)

// **Die vier Güten unterscheiden sich in der Abtastung** — und nur darin.
//
// Gemessen am 2026-08-25 über zwanzig Karten aus `Black.mse-set`
// (#1029):
// Bildebenen tragen 78 Prozent der Rasterzeit, und der nächste Nachbar
// halbiert sie. Das ist der einzige gemessene Hebel; alles andere am Rastern
// ist zu klein, um eine Güte darauf zu gründen.

// verlaufsbild gibt ein Bild, dessen Rotwert je Spalte um [Stufe] wächst.
//
// **Ein Verlauf und keine Stufe.** Der erste Anlauf nahm zwei Farben
// nebeneinander — und traf die Kante nie: beim Verkleinern von 40 auf 30
// liegen die Abtaststellen so, dass keine der vier Nachbarschaften über den
// Sprung reicht. Der Test war rot, und die Vorgabe war in Ordnung.
//
// Bei einem Verlauf trägt **jede** Zwischenstelle einen Wert, den keine
// Quellspalte hat — daran ist Mischen von Nichtmischen zu unterscheiden.
func verlaufsbild(t *testing.T, breite, hoehe int) *image.NRGBA {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, color.NRGBA{R: uint8(x * Stufe), A: 255})
		}
	}

	return bild
}

// Stufe ist der Abstand zweier Quellspalten im Verlauf.
const Stufe = 6

// werteInDerZeile gibt die verschiedenen Rotwerte einer abgetasteten Zeile.
func werteInDerZeile(
	quelle image.Image, rahmen image.Rectangle, guete Guete,
) map[uint8]int {
	werte := map[uint8]int{}

	for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
		werte[abtasten(quelle, rahmen, x, rahmen.Min.Y, guete).R]++
	}

	return werte
}

// **Der nächste Nachbar mischt nicht.** Jeder Zielpunkt trägt eine der beiden
// Quellfarben und keine dazwischen.
//
// **Und das ist die Auslassung, die diese Güte ansagen muss**: der Kommentar
// an [abtasten] führt seit jeher den Grund gegen den nächsten Nachbarn — eine
// Vorschau mit Treppen und ein Druckbild ohne sie sind nicht dasselbe Bild.
// Für den Goldbildvergleich taugt `GueteSchnell` deshalb nicht.
func TestSchnellNimmtDenNaechstenNachbarn(t *testing.T) {
	quelle := verlaufsbild(t, 40, 40)
	rahmen := image.Rect(0, 0, 30, 30)

	werte := werteInDerZeile(quelle, rahmen, GueteSchnell)
	if len(werte) < 2 {
		t.Fatalf("die Zeile trägt nur %d Wert(e) — der Verlauf kam nicht an", len(werte))
	}

	for wert := range werte {
		if wert%Stufe != 0 {
			t.Errorf("der nächste Nachbar hat gemischt: Rotwert %d", wert)
		}
	}
}

// **Die Vorgabe mischt** — und der Nullwert ist die Vorgabe.
//
// Ein vergessener Parameter gibt damit die Parität zu MSE und keine
// Überraschung. Dieselbe Falle wie bei `Deckung 0` in [Rasterer.ebene], nur
// andersherum aufgelöst.
func TestDieVorgabeIstDieParitaet(t *testing.T) {
	var vorgabe Guete

	if vorgabe != GueteMittel {
		t.Fatalf("der Nullwert ist %v, nicht GueteMittel", vorgabe)
	}

	// **Leicht verkleinert, nicht vergrößert.** MSEs Regel rechnet nur beim
	// Verkleinern unter Faktor zwei bilinear; beim Vergrößern mittelt sie über
	// die Fläche und mischt gerade **nicht**. Der erste Anlauf dieses Tests
	// vergrößerte und schlug fehl — an der Vorgabe lag es nicht.
	quelle := verlaufsbild(t, 40, 40)
	rahmen := image.Rect(0, 0, 30, 30)

	gemischt := false

	for wert := range werteInDerZeile(quelle, rahmen, vorgabe) {
		if wert%Stufe != 0 {
			gemischt = true
		}
	}

	if !gemischt {
		t.Error("die Vorgabe hat nicht gemischt")
	}
}

// **Alle vier sind aufgezählt, und jede trägt einen Namen.**
//
// Die Aufzählung ist der Weg der Flächen (§16): eine Güte, die hier fehlt, ist
// über keine von ihnen erreichbar.
func TestJedeGueteIstAufgezaehltUndBenannt(t *testing.T) {
	if len(Gueten) != 4 {
		t.Fatalf("es sind %d Güten aufgezählt, nicht vier", len(Gueten))
	}

	gesehen := map[string]bool{}

	for _, guete := range Gueten {
		name := guete.Name()
		if name == "" {
			t.Errorf("die Güte %d trägt keinen Namen", int(guete))
		}

		if gesehen[name] {
			t.Errorf("der Name %q kommt zweimal vor", name)
		}

		gesehen[name] = true
	}
}

// **Nur die Vorgabe schweigt.** Jede andere Güte sagt an der Ausgabe, worin
// sie von MSE abweicht — das ist die Forderung des Werkstücks Ω.4, und sie
// gilt für die bessere Abweichung genauso wie für die schnellere.
func TestJedeAbweichungSagtSich(t *testing.T) {
	if hinweis := GueteMittel.Hinweis(); hinweis != "" {
		t.Errorf("die Vorgabe weicht nicht ab, sagt aber %q", hinweis)
	}

	for _, guete := range Gueten {
		if guete == GueteMittel {
			continue
		}

		if guete.Hinweis() == "" {
			t.Errorf("die Güte %q sagt ihre Abweichung nicht an", guete.Name())
		}
	}
}

// schachbild gibt ein Bild mit einem Wechsel je Spalte — die feinste Linie,
// die es gibt.
func schachbild(t *testing.T, breite, hoehe int, dunkel, hell uint8) *image.NRGBA {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			wert := dunkel
			if x%2 == 1 {
				wert = hell
			}

			bild.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return bild
}

// unruhe misst, wie stark benachbarte Zielpunkte springen.
func unruhe(quelle image.Image, rahmen image.Rectangle, guete Guete) int {
	var (
		summe  int
		vorher = -1
	)

	for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
		wert := int(abtasten(quelle, rahmen, x, rahmen.Min.Y, guete).R)
		if vorher >= 0 {
			summe += abs(wert - vorher)
		}

		vorher = wert
	}

	return summe
}

// **Was [GueteHoch] verspricht, ist Ruhe** — „feine Linien flimmern nicht
// mehr". Der Test misst genau das und nicht, dass sich zwei Bilder
// unterscheiden.
//
// MSEs Regel rechnet unterhalb von Faktor zwei bilinear, und bilinear lässt
// ein Punktmuster flimmern: das war die Messung aus
// #113, die
// mit #532
// zugunsten der Parität aufgegeben wurde. `GueteHoch` gibt sie als **Wahl**
// zurück, nicht als Vorgabe.
func TestHochBeruhigtFeineLinien(t *testing.T) {
	quelle := schachbild(t, 60, 4, 40, 200)
	rahmen := image.Rect(0, 0, 41, 4)

	mitParitaet := unruhe(quelle, rahmen, GueteMittel)
	mitHoch := unruhe(quelle, rahmen, GueteHoch)

	if mitHoch >= mitParitaet {
		t.Errorf("GueteHoch beruhigt nicht: %d gegen %d bei der Vorgabe",
			mitHoch, mitParitaet)
	}
}

// **[GueteHoechst] rechnet beim Vergrößern bikubisch**, und daran ist sie zu
// erkennen: der Kern von Catmull-Rom überschwingt an einer harten Kante. Weder
// das Flächenmittel noch die bilineare Rechnung tun das — beide bleiben
// zwischen den beiden Quellwerten.
//
// **Die Quellwerte liegen mit Absicht nicht bei 0 und 255**: dort klemmte
// [alsAnteil] den Überschwinger weg, und der Test sähe nichts.
func TestHoechstSchwingtBeimVergroessernUeber(t *testing.T) {
	const (
		dunkel = 60
		hell   = 190
	)

	quelle := stufe(t, 8, 4, dunkel, hell)
	rahmen := image.Rect(0, 0, 64, 4)

	ueberschwingt := func(guete Guete) bool {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			wert := abtasten(quelle, rahmen, x, rahmen.Min.Y, guete).R
			if wert < dunkel || wert > hell {
				return true
			}
		}

		return false
	}

	if !ueberschwingt(GueteHoechst) {
		t.Error("GueteHoechst rechnet nicht bikubisch: kein Überschwinger")
	}

	for _, andere := range []Guete{GueteMittel, GueteHoch, GueteSchnell} {
		if ueberschwingt(andere) {
			t.Errorf("die Güte %q schwingt über — das tut nur die bikubische",
				andere.Name())
		}
	}
}

// stufe gibt ein Bild mit einer harten Kante in der Mitte.
func stufe(t *testing.T, breite, hoehe int, dunkel, hell uint8) *image.NRGBA {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			wert := dunkel
			if x >= breite/2 {
				wert = hell
			}

			bild.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return bild
}

// **Ein Wert, den es nicht gibt, trägt keinen Namen und keinen Hinweis.**
//
// `Guete` ist ein `int`, und `Guete(7)` ist eine Zeichenkette Abstand von
// einem Tippfehler an einer Fläche entfernt. Ein Name, den er trüge, machte
// ihn zu einer fünften Güte.
func TestEineUnbekannteGueteSchweigt(t *testing.T) {
	fremd := Guete(7)

	if name := fremd.Name(); name != "" {
		t.Errorf("Guete(7) trägt den Namen %q", name)
	}

	if hinweis := fremd.Hinweis(); hinweis != "" {
		t.Errorf("Guete(7) sagt %q", hinweis)
	}
}

// **[GueteNach] findet jede aufgezählte Güte und sonst keine.**
//
// Der zweite Rückgabewert ist die halbe Zusage: ohne ihn läge bei jedem
// Tippfehler still die Vorgabe an, und eine Fläche könnte ihn nicht melden.
func TestGueteNachFindetNurWasEsGibt(t *testing.T) {
	for _, gewollt := range Gueten {
		gefunden, da := GueteNach(gewollt.Name())
		if !da || gefunden != gewollt {
			t.Errorf("%q gab (%d, %v)", gewollt.Name(), int(gefunden), da)
		}
	}

	for _, unbekannt := range []string{"", "Fast", "schnell", "highres"} {
		gefunden, da := GueteNach(unbekannt)
		if da {
			t.Errorf("%q gilt als Güte", unbekannt)
		}

		if gefunden != GueteMittel {
			t.Errorf("%q gab %d statt der Vorgabe", unbekannt, int(gefunden))
		}
	}
}

// **Beim Verkleinern rechnet [GueteHoechst] wie [GueteHoch]** — bikubisch ist
// die Antwort auf das Vergrößern, nicht auf das Verkleinern. Wer eine
// Illustration schrumpft, will kein Klingeln an jeder Kante, sondern das
// Mittel über die Fläche, die verschwindet.
func TestHoechstVerkleinertWieHoch(t *testing.T) {
	quelle := schachbild(t, 60, 4, 40, 200)
	rahmen := image.Rect(0, 0, 41, 4)

	mitHoch := unruhe(quelle, rahmen, GueteHoch)
	mitHoechst := unruhe(quelle, rahmen, GueteHoechst)

	if mitHoechst != mitHoch {
		t.Errorf("GueteHoechst verkleinert anders als GueteHoch: %d gegen %d",
			mitHoechst, mitHoch)
	}
}

// **Ohne Größenänderung gibt die bikubische Rechnung die Quelle zurück.**
//
// Das ist die Probe auf den Kern: bei ganzzahliger Abtaststelle trägt der
// mittlere Nachbar das Gewicht eins und die drei anderen genau null. Fällt
// diese Zusage, verschöbe `ultra` ein Bild, das gar nicht skaliert wird.
func TestHoechstLaesstUngeaendertesUngeaendert(t *testing.T) {
	quelle := verlaufsbild(t, 8, 4)
	rahmen := image.Rect(0, 0, 8, 4)

	for x := range 8 {
		gewollt := quelle.NRGBAAt(x, 0).R
		gemessen := abtasten(quelle, rahmen, x, 0, GueteHoechst).R

		if gemessen != gewollt {
			t.Errorf("Spalte %d: %d statt %d", x, gemessen, gewollt)
		}
	}
}
