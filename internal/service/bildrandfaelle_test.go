package service

// **Die Randfälle des Nachzeichnens** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Warum sie eine eigene Datei haben
//
// Die Läufe daneben prüfen, was ein Bild ergibt. Hier stehen die Fälle, die
// ein Bild **nicht** herstellt: ein Feld, das der Sprungtabelle widerspricht,
// ein Nullvektor, ein Zug, der nicht zurückkommt. Sie sind erreichbar, weil
// die Bausteine paketintern sind — und sie stehen hier, weil ein unbetretener
// Fehlerweg keiner ist (§3).

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein widersprüchlicher Stand bleibt stehen.**
//
// MSE wirft dort `InternalError("in the ground/air")` (Zeile 181). Über ein
// Feld ist der Fall **nicht** erreichbar — der Rundgang steht immer auf einer
// Kante, und eine Kante hat nie vier gleiche Nachbarn.
//
// **Der erste Anlauf glaubte etwas anderes**: ein durchgehend volles 3 × 3
// sollte ihn auslösen. Es tut es nicht — der Zug läuft außen um den Block
// herum, und das ist richtig. Aus zwei Ausgängen für denselben unmöglichen
// Stand ist deshalb einer geworden: die Stelle bleibt, und die Schranke
// beendet den Zug.
func TestEinWiderspruechlicherStandBleibtStehen(t *testing.T) {
	t.Parallel()

	leer := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}

	if x, y := naechsterSchritt(leer, 1, 1, 1); x != 1 || y != 1 {
		t.Errorf("alles leer gab (%d,%d), erwartet (1,1)", x, y)
	}

	voll := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}
	for i := range voll.zellen {
		voll.zellen[i] = true
	}

	if x, y := naechsterSchritt(voll, 1, 1, 1); x != 1 || y != 1 {
		t.Errorf("alles voll gab (%d,%d), erwartet (1,1)", x, y)
	}
}

// **Und ein Zug, der nicht zurückkommt, endet an der Schranke.**
//
// MSE hat keine; sie steht hier, weil ein hängendes Fenster die schlechteste
// aller Antworten ist. Nachgestellt über ein Feld, dessen Zellen sich beim
// Ablaufen ändern — was ein echtes Bild nie tut.
func TestEinZugOhneRueckkehrEndetAnDerSchranke(t *testing.T) {
	t.Parallel()

	feld := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}
	feld.zellen[0] = true

	// **Die Schranke auf null**: dann greift sie beim ersten Schritt, und der
	// Fall lässt sich prüfen, ohne ein Feld zu bauen, das es nicht gibt.
	if _, gefunden := umrissLesenMitSchranke(feld, 0); gefunden {
		t.Error("ein Zug jenseits der Schranke gab eine Form")
	}
}

// **Die Zweideutigkeiten der Sprungtabelle.**
//
// `0x0110` und `0x1001` sind Diagonalen: sie lassen sich von zwei Seiten
// betreten, und MSE entscheidet sie über die **letzte** Bewegung (Zeilen
// 175–176). Ohne diese Unterscheidung liefe der Rundgang an einer Sanduhr in
// die falsche Hälfte.
func TestDieDiagonalenFolgenDerLetztenBewegung(t *testing.T) {
	t.Parallel()

	// 0x0110: b und c voll, a und d leer.
	//	. #
	//	 ·
	//	# .
	feld := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}
	feld.zellen[1] = true // (1,0) = b
	feld.zellen[2] = true // (0,1) = c

	for _, fall := range []struct {
		letzte int
		wartet [2]int
	}{
		{1, [2]int{1, 0}},
		{-1, [2]int{1, 2}},
	} {
		x, y := naechsterSchritt(feld, 1, 1, fall.letzte)
		if x != fall.wartet[0] || y != fall.wartet[1] {
			t.Errorf("letzte=%d: (%d,%d), erwartet (%d,%d)",
				fall.letzte, x, y, fall.wartet[0], fall.wartet[1])
		}
	}

	// 0x1001: a und d voll, b und c leer.
	anders := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}
	anders.zellen[0] = true // (0,0) = a
	anders.zellen[3] = true // (1,1) = d

	for _, fall := range []struct {
		letzte int
		wartet [2]int
	}{
		{1, [2]int{2, 1}},
		{-1, [2]int{0, 1}},
	} {
		x, y := naechsterSchritt(anders, 1, 1, fall.letzte)
		if x != fall.wartet[0] || y != fall.wartet[1] {
			t.Errorf("letzte=%d: (%d,%d), erwartet (%d,%d)",
				fall.letzte, x, y, fall.wartet[0], fall.wartet[1])
		}
	}
}

// **Merken und Nachsehen außerhalb des Feldes tun nichts.**
//
// Dieselbe Zusage wie bei `voll`: außerhalb gibt es nichts, und ein Zugriff
// dorthin darf nicht abstürzen. MSEs `ImageData::operator()` gibt dort eine
// Attrappe zurück, in die geschrieben werden darf (Zeilen 116–127).
func TestAusserhalbWirdNichtsGemerkt(t *testing.T) {
	t.Parallel()

	feld := &schwellenfeld{breite: 2, hoehe: 2, zellen: make([]bool, 4), besucht: make([]bool, 4)}

	feld.merken(-1, 0)
	feld.merken(0, -1)
	feld.merken(2, 0)
	feld.merken(0, 2)

	for _, stelle := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 2}} {
		if feld.gemerkt(stelle[0], stelle[1]) {
			t.Errorf("(%d,%d) gilt als gemerkt", stelle[0], stelle[1])
		}
	}

	for _, gemerkt := range feld.besucht {
		if gemerkt {
			t.Error("eine Zelle im Feld wurde von außerhalb markiert")
		}
	}
}

// **Ein einfarbiges Bild gibt die Grenze 255.**
//
// Die Schleife in `gemesseneGrenze` läuft bis 254; ein Bild, dessen Punkte
// alle heller sind, verlässt sie ohne Treffer. Ohne den Rückfall stünde dort
// der Nullwert, und **jeder** Punkt wäre voll.
func TestEinSehrHellesBildBekommtDieGrenze255(t *testing.T) {
	t.Parallel()

	if grenze := gemesseneGrenze([]uint8{255, 255, 255, 255}); grenze != 255 {
		t.Errorf("Grenze %d, erwartet 255", grenze)
	}
}

// **Und ein zweifarbiges die Stufe darüber** — der zweite Ausgang, `below >=
// threshold_pos` (Zeile 68).
func TestZweiKlassenGebenDieStufeDarueber(t *testing.T) {
	t.Parallel()

	// Vier Punkte: zwei bei 10, zwei bei 200. `mitte` ist 2.
	// i=10: darunter+hist/2 = 0+1 = 1, nicht > 2; darunter wird 2 ≥ 2 → 11.
	if grenze := gemesseneGrenze([]uint8{10, 10, 200, 200}); grenze != 11 {
		t.Errorf("Grenze %d, erwartet 11", grenze)
	}
}

// **Der Nullvektor genormt bleibt der Nullvektor.**
//
// MSEs `normalized()` teilt ohne Prüfung und gäbe `NaN` — das breitete sich
// still über die ganze Form aus und fiele erst beim Zeichnen auf.
func TestDerNullvektorGenormtIstNull(t *testing.T) {
	t.Parallel()

	if genormt(msepaket.Vektor{}) != (msepaket.Vektor{}) {
		t.Error("der Nullvektor genormt ist nicht null")
	}
}

// **Eine Form mit drei Punkten wird nicht vereinfacht.**
//
// Weniger als vier Punkte hat keine Nachbarschaft, über die sich mitteln
// ließe — MSEs Gewichte greifen bis vier Punkte weit in jede Richtung.
func TestEineFormMitDreiPunktenBleibtWieSieIst(t *testing.T) {
	t.Parallel()

	teil := msepaket.Symbolteil{
		Art: msepaket.TeilForm,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0, Y: 0}},
			{Stelle: msepaket.Vektor{X: 1, Y: 0}},
			{Stelle: msepaket.Vektor{X: 0, Y: 1}},
		},
	}

	if len(formVereinfachen(teil).Punkte) != 3 {
		t.Error("eine Form mit drei Punkten wurde angefasst")
	}
}

// **Ein Dreieck verschmilzt seine benachbarten Ecken.**
//
// MSEs eigenes Beispiel (Zeilen 258–272): eine Spitze gibt zwei benachbarte
// Ecken, und wo ihre Tangenten sich treffen, liegt die eine.
func TestEinDreieckVerschmilztSeineEcken(t *testing.T) {
	t.Parallel()

	const mass = 100

	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))

	for y := range mass {
		for x := range mass {
			ton := uint8(255)

			// Ein Dreieck mit der Spitze oben.
			if y >= 20 && y < 80 {
				breite := (y - 20) / 2
				if x >= 50-breite && x <= 50+breite {
					ton = 0
				}
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	teile, err := BildNachzeichnen(gemalt)
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	if len(teile) != 1 {
		t.Fatalf("%d Teile, erwartet 1", len(teile))
	}

	// **Wenige Punkte und mindestens drei Geraden** — ein Dreieck.
	if anzahl := len(teile[0].Punkte); anzahl > 20 {
		t.Errorf("%d Punkte für ein Dreieck", anzahl)
	}
}

// **Ein MSE-1-Symbol wird an seiner Größe erkannt.**
//
// 60 × 80, und die rechte Hälfte schwarzweiß — `is_mse1_symbol` (Zeilen
// 97–114). Genommen wird daraus der Ausschnitt (20,0,40,40).
func TestEinMSE1SymbolWirdErkannt(t *testing.T) {
	t.Parallel()

	gemalt := image.NewRGBA(image.Rect(0, 0, 60, 80))
	mitte := 40.0

	for y := range 80 {
		for x := range 60 {
			farbe := color.RGBA{R: 200, G: 40, B: 40, A: 255} // links bunt

			if x >= 20 {
				ton := uint8(255)
				if math.Hypot(float64(x)-mitte, float64(y)-20) <= 12 {
					ton = 0
				}

				farbe = color.RGBA{R: ton, G: ton, B: ton, A: 255}
			}

			gemalt.Set(x, y, farbe)
		}
	}

	if !istMSE1Symbol(gemalt) {
		t.Fatal("das Symbol wurde nicht als MSE-1-Symbol erkannt")
	}

	if _, err := BildNachzeichnen(gemalt); err != nil {
		t.Errorf("nachzeichnen: %v", err)
	}
}

// **Und ein buntes 60 × 80 ist keines** — die Schwelle 5000 misst genau das.
func TestEinBuntes60Mal80IstKeinMSE1Symbol(t *testing.T) {
	t.Parallel()

	gemalt := image.NewRGBA(image.Rect(0, 0, 60, 80))
	for y := range 80 {
		for x := range 60 {
			gemalt.Set(x, y, color.RGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}

	if istMSE1Symbol(gemalt) {
		t.Error("ein buntes Bild gilt als MSE-1-Symbol")
	}
}

// **Ein Bild anderer Größe ebenso.**
func TestEinAnderesMassIstKeinMSE1Symbol(t *testing.T) {
	t.Parallel()

	if istMSE1Symbol(image.NewRGBA(image.Rect(0, 0, 60, 60))) {
		t.Error("60 × 60 gilt als MSE-1-Symbol")
	}
}

// **Der Betrag einer negativen Zahl.**
//
// `istMSE1Symbol` summiert Farbabstände; ohne den Betrag höben sich ein
// zu großes Rot und ein zu kleines Blau gegenseitig auf, und ein buntes Bild
// käme unter die Schwelle.
func TestDerBetragKehrtDasVorzeichenUm(t *testing.T) {
	t.Parallel()

	for _, fall := range [][2]int{{-3, 3}, {3, 3}, {0, 0}} {
		if gefunden := abs(fall[0]); gefunden != fall[1] {
			t.Errorf("abs(%d) = %d, erwartet %d", fall[0], gefunden, fall[1])
		}
	}
}

// **Die Grenze fällt in die Klasse, die über die Mitte reicht.**
//
// MSEs erster Ausgang, `below + hist[i]/2 > threshold_pos` (Zeile 63): die
// **halbe** Klasse zählt mit. Ohne diese Hälfte spränge die Grenze bei einem
// Bild mit einer sehr großen Klasse um eine ganze Stufe daneben.
func TestDieGrenzeFaelltInDieGrosseKlasse(t *testing.T) {
	t.Parallel()

	// Sieben Punkte: einer bei 0, sechs bei 5. `mitte` ist 3.
	// i=0: 0 + 1/2 = 0, nicht > 3; darunter wird 1, nicht ≥ 3.
	// i=5: 1 + 6/2 = 4 > 3 → die Grenze ist 5.
	if grenze := gemesseneGrenze([]uint8{0, 5, 5, 5, 5, 5, 5}); grenze != 5 {
		t.Errorf("Grenze %d, erwartet 5", grenze)
	}
}

// **Zwei Punkte lassen sich nicht verschmelzen.**
//
// `eckenVerschmelzen` greift auf den Vorgänger und vier Nachbarn; unter drei
// Punkten gibt es keine Nachbarschaft, sondern nur denselben Punkt wieder.
func TestZweiPunkteWerdenNichtVerschmolzen(t *testing.T) {
	t.Parallel()

	punkte := []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}, Sperre: sperreFrei},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}, Sperre: sperreFrei},
	}

	if len(eckenVerschmelzen(punkte)) != 2 {
		t.Error("zwei Punkte wurden angefasst")
	}
}

// **Drei Punkte bleiben auch beim Geradenverschmelzen stehen.**
//
// Eine geschlossene Form braucht mindestens drei; das vierte zu entfernen
// hieße, aus einem Dreieck eine Strecke zu machen.
func TestDreiPunkteBleibenBeimGeradenverschmelzen(t *testing.T) {
	t.Parallel()

	punkte := []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}},
		{Stelle: msepaket.Vektor{X: 0, Y: 1}},
	}

	if len(geradenVerschmelzen(punkte)) != 3 {
		t.Error("aus drei Punkten wurden weniger")
	}
}

// **Ein Punkt mitten auf einer Geraden fällt weg.**
//
// Portiert aus `merge_lines` (Zeilen 373–388). Das Maß ist Winkel mal
// Streckenquadrat: bei drei Punkten auf einer Linie ist der Winkel null, und
// damit ist es das Produkt.
func TestEinPunktAufDerGeradenFaelltWeg(t *testing.T) {
	t.Parallel()

	punkte := []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}},
		{Stelle: msepaket.Vektor{X: 0.1, Y: 0}},
		{Stelle: msepaket.Vektor{X: 0.2, Y: 0}},
		{Stelle: msepaket.Vektor{X: 0.2, Y: 0.2}},
		{Stelle: msepaket.Vektor{X: 0, Y: 0.2}},
	}

	verkuerzt := geradenVerschmelzen(punkte)

	if len(verkuerzt) >= len(punkte) {
		t.Errorf("%d Punkte geblieben von %d — der mittlere hätte wegfallen müssen",
			len(verkuerzt), len(punkte))
	}
}
