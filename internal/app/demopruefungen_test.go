//go:build fremdpakete

package app_test

// Was auf einer Demokarte stehen muss — die Zusagen, nicht die Karten.
//
// **Eigene Datei**, weil `demokarten_test.go` die Frage „welche Karten" schon
// beantwortet; hier steht „was muss ankommen". Der Anlass war die
// 400-Zeilen-Grenze (§2), der Schnitt folgt der Frage.

import (
	"image"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// pruefeBemalung prüft, dass die Karte nicht leer und nicht schwarz ist.
//
// **Die stillste Art zu scheitern.** Ein Plan mit zwanzig Ebenen, von denen
// keine ein Bild findet, ergibt eine weiße Karte **ohne** Befund — genau der
// Fall aus #248,
// den ein Review über das Handwerk nicht sieht.
func pruefeBemalung(t *testing.T, bild *image.NRGBA, karte demokarte) {
	t.Helper()

	rahmen := bild.Bounds()

	var bunt, dunkel int

	for y := rahmen.Min.Y; y < rahmen.Max.Y; y++ {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			farbe := bild.NRGBAAt(x, y)

			hell := int(farbe.R) + int(farbe.G) + int(farbe.B)
			switch {
			case hell < 60:
				dunkel++
			case farbe.R != farbe.G || farbe.G != farbe.B:
				bunt++
			}
		}
	}

	gesamt := rahmen.Dx() * rahmen.Dy()

	// **Ein Fünftel Farbe** — der Kartenrahmen allein bringt mehr; weniger
	// heißt, dass die Rahmenbilder fehlen.
	//
	// **Die Schwelle darf je Karte anders sein**, siehe
	// [demokarte.buntMindestens]: eine textlose Vorlage hat keinen Textkasten,
	// und ein weißer Rahmen ist grau.
	bunthoehe := karte.buntMindestens
	if bunthoehe == 0 {
		bunthoehe = 20
	}

	if bunt*100 < gesamt*bunthoehe {
		t.Errorf("nur %d von %d Bildpunkten sind farbig (%d %% verlangt) — "+
			"die Karte ist fast leer", bunt, gesamt, bunthoehe)
	}

	// **Und höchstens sieben Zehntel schwarz.**
	//
	// Die Schwelle liegt nicht bei der Hälfte, und der Grund gehört dazu: ein
	// **leeres Bildfenster ist schwarz**, weil der Kartenrahmen dort ein Loch
	// hat — und bei `magic-m15-saga` nimmt es die halbe Karte ein. Eine Karte
	// ohne Illustration ist damit legitim zu über der Hälfte schwarz; was
	// diese Zusage fängt, ist der Rahmen, der ganz ausfällt.
	dunkelhoehe := karte.dunkelHoechstens
	if dunkelhoehe == 0 {
		dunkelhoehe = 70
	}

	if dunkel*100 > gesamt*dunkelhoehe {
		t.Errorf("%d von %d Bildpunkten sind schwarz (höchstens %d %% erlaubt) — "+
			"der Rahmen fehlt", dunkel, gesamt, dunkelhoehe)
	}
}

// pruefeTexte prüft, dass Name, Typzeile und Regeltext auf der Karte stehen.
//
// **Im Plan und nicht im Bild.** Ob ein Zug gesetzt wurde, sagt der Plan
// genau; aus Bildpunkten zurückzulesen, welcher Text dort steht, hieße eine
// Texterkennung zu bauen (§10 — die Reichweite einer Messung).
func pruefeTexte(t *testing.T, plan *render.Plan, karte demokarte) {
	t.Helper()

	var gesetzt []string

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtText {
			gesetzt = append(gesetzt, plan.Ebenen[i].Klartext())
		}
	}

	ganzerText := strings.Join(gesetzt, "\n")

	for _, feld := range []string{"name", "super type"} {
		wert := karte.felder[feld]
		if wert == "" || slices.Contains(karte.ohneFelder, feld) {
			continue
		}

		if !strings.Contains(ganzerText, wert) {
			t.Errorf("%q steht nicht auf der Karte; gesetzt wurde:\n%s", wert, ganzerText)
		}
	}

	// **Der Regeltext steht mit seinem ersten Wort da.** Ganz zu vergleichen
	// ginge nicht: die Vorlage setzt Symbole und Blockgrenzen hinein, und
	// genau das soll sie.
	//
	// **Ohne Auszeichnungen**: `<sym>T</sym>: Add …` beginnt mit einem Symbol,
	// und im gesetzten Zug steht davon nur das `T`.
	if roh := karte.felder["rule text"]; roh != "" &&
		!slices.Contains(karte.ohneFelder, "rule text") {
		worte := strings.Fields(ohneAuszeichnungen(roh))
		if len(worte) > 0 && !strings.Contains(ganzerText, strings.Trim(worte[0], ",.:")) {
			t.Errorf("der Regeltext fehlt; gesetzt wurde:\n%s", ganzerText)
		}
	}
}

// ohneAuszeichnungen entfernt `<…>`-Marken aus einem Kartentext.
func ohneAuszeichnungen(text string) string {
	var gebaut strings.Builder

	tief := 0

	for _, zeichen := range text {
		switch {
		case zeichen == '<':
			tief++
		case zeichen == '>':
			tief--
		case tief == 0:
			gebaut.WriteRune(zeichen)
		}
	}

	return gebaut.String()
}

// pruefeSeitenrand verlangt einen durchgehend schwarzen Rand an beiden Seiten.
//
// # Was er fängt
//
// Bis #388
// standen dort zwei weiße Kerben: das Feld `corners` zeichnet ein 5 × 5 großes
// Bild über die ganze Karte und schneidet es mit einer 375 × 523 großen Maske.
// Eingepasst statt gestreckt wurde daraus ein 375 × 375 großes Bild — und die
// Maske mit ihm gestaucht, ihre untere Rundung 148 Zeilen zu hoch.
//
// **Gemessen wird zwischen y = 30 und 492**: darüber und darunter liegen die
// abgerundeten Ecken, und dort ist außerhalb der Rundung nichts (weiß).
//
// **Die vier äußersten Spalten**, denn so weit reicht der Rand jeder
// M15-Vorlage dieses Packs (14 bis 15 Bildpunkte, nachgemessen).
func pruefeSeitenrand(t *testing.T, bild *image.NRGBA) {
	t.Helper()

	var hell int

	for y := 30; y < 493; y++ {
		for _, x := range []int{0, 1, 2, 3, 371, 372, 373, 374} {
			farbe := bild.NRGBAAt(x, y)
			if int(farbe.R)+int(farbe.G)+int(farbe.B) > 150 {
				hell++
			}
		}
	}

	if hell > 0 {
		t.Errorf("%d helle Bildpunkte im Seitenrand — dort gehört der Kartenrand hin", hell)
	}
}

// pruefeIndikator verlangt den Farbpunkt vor der Typzeile.
//
// **Er hängt an der Wahl der Nutzenden**, nicht an der Vorlage:
// `has_identity_general` (`magic.mse-game/script`, Zeilen 444–471) zeigt ihn
// nur, wenn die Identität der Rahmenfarbe widerspricht. Sein Vorgabewert ist
// die Rahmenfarbe selbst — eine gewöhnliche weiße Kreatur führt ihn deshalb
// **nicht**, und das ist richtig so.
func pruefeIndikator(t *testing.T, plan *render.Plan) {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtBild &&
			strings.Contains(plan.Ebenen[i].Bild.Pfad, "indicators/") {
			return
		}
	}

	t.Error("keine Ebene aus `indicators/` — der Farbpunkt vor der Typzeile fehlt")
}

// pruefeAbsatzband prüft, dass jeder Absatz in seinem Band sitzt.
//
// **An der Zeilenstelle und nicht am Augenschein**
// (#479). Die
// drei Fähigkeiten einer Planeswalkerkarte stehen in **einem** Textfeld; was
// sie neben ihre Loyalitätssymbole setzt, ist `paragraph height: 49`.
//
// **Der Satz wird hier noch einmal gefahren.** Der Plan führt die Abschnitte,
// nicht die gesetzten Zeilen — die entstehen beim Zeichnen. Ihn hier zu
// wiederholen ist billiger, als die Stellen aus Bildpunkten zurückzulesen
// (§10).
func pruefeAbsatzband(t *testing.T, plan *render.Plan, suche render.Schriftquelle) {
	t.Helper()

	for i := range plan.Ebenen {
		ebene := plan.Ebenen[i]
		if ebene.Art != render.ArtText || ebene.Satz.Absatzhoehe <= 0 {
			continue
		}

		schrift, err := suche.Suche(ebene.Satz.Schrift, false)
		if err != nil {
			t.Fatalf("Schrift %q: %v", ebene.Satz.Schrift, err)
		}

		gesetzt := render.SetzeAbschnitte(ebene.Satz, ebene.Rahmen, ebene.Abschnitte,
			render.Schriftsatz{Aufrecht: schrift}, nil)

		// **Der Abstand ist die Absatzhöhe, mit benannter Toleranz.**
		//
		// Genau wäre er es nur, wenn alle Absätze gleich hoch **und** ohne
		// eigenen oberen Rand wären. Auf dieser Karte trägt der vierte
		// Absatz — der leere Flavourtext hinter dem Blocktrenner — einen Rand
		// von 1,6 Punkten und sitzt dadurch 0,8 tiefer in seinem Band. MSE
		// rechnet genauso: `alignParagraph` bekommt das Band und richtet
		// darin aus, was die Zeilen samt ihren Rändern einnehmen.
		//
		// **Zwei Punkte Toleranz, und sie reicht weit genug**: vor dieser
		// Änderung standen die Absätze 16,09 auseinander — eine Zeilenhöhe,
		// nicht ein Band.
		//
		// **Die Absätze, nicht die Zeilen.** Eine umgebrochene Fähigkeit
		// bleibt ein Absatz.
		const toleranz = 2

		if len(gesetzt.Absaetze) < 2 {
			t.Error("nur ein Absatz — die Zusage prüft nichts")

			return
		}

		// **Ein leerer Absatz hat keine Lage, die man zusichern könnte.**
		//
		// Der Abstand zweier **Oberkanten** ist nur dann die Bandhöhe, wenn
		// beide Absätze gleich hoch sind: bei `links mitte` sitzt jeder in
		// seinem Band mittig, also `(Band − Inhalt) / 2` unter dessen Kante.
		// Ein leerer Absatz hat Inhalt 0 und sitzt damit tiefer als ein
		// gefüllter — ohne dass irgendetwas gezeichnet würde.
		//
		// **MSE rechnet genauso**: `alignParagraph` sucht die Höhe und zählt
		// die letzten Zeilen nicht mit, wenn sie leer sind
		// (`src/render/text/viewer.cpp`, Zeilen 771–775, Fassung f9d9356 vom
		// 2021-07-06), und richtet den Rest im Band aus.
		//
		// **Gemessen am 2026-08-22** an `Planeswalker`: die Absätze 1 bis 3
		// tragen je 7,00 und stehen bei 350,65 / 399,65 / 448,65 — **exakt
		// 49,00** auseinander. Der vierte trägt **0,00** und sitzt bei 501,50.
		// Der Wächter meldete 52,85 und traf damit nichts Gezeichnetes.
		for n := 1; n < len(gesetzt.Absaetze); n++ {
			if gesetzt.Absaetze[n].Hoehe == 0 || gesetzt.Absaetze[n-1].Hoehe == 0 {
				t.Logf("  · Absatz %d oder sein Vorgänger ist leer — keine Lage zu prüfen", n+1)

				continue
			}

			abstand := gesetzt.Absaetze[n].Oben - gesetzt.Absaetze[n-1].Oben
			if math.Abs(abstand-ebene.Satz.Absatzhoehe) > toleranz {
				t.Errorf("Absatz %d beginnt %.2f unter dem vorigen, erwartet "+
					"%.2f ± %d (`paragraph height`)",
					n+1, abstand, ebene.Satz.Absatzhoehe, toleranz)
			}
		}

		return
	}

	t.Error("keine Textebene mit `paragraph height` — die Zusage prüft nichts")
}

// pruefeMischart prüft, dass eine Ebene mit dieser Mischart im Plan steht.
//
// **Der Beleg, dass `set_combine` ankommt** (#470). Der Feldstil
// `lvlup multiply 0` von `magic-m15-leveler` schreibt
// `image: {set_combine(input: "lvlup_multiply_0.png", combine: "multiply")}`
// und ist nur sichtbar, wenn `card.level_0` gesetzt ist — deshalb trägt die
// Demokarte ihre Stufen.
func pruefeMischart(t *testing.T, plan *render.Plan, erwartet string) {
	t.Helper()

	gefunden := 0

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtBild &&
			plan.Ebenen[i].Mischart == erwartet {
			gefunden++
		}
	}

	if gefunden == 0 {
		t.Errorf("keine Bildebene mit der Mischart %q — `set_combine` bringt "+
			"sie nicht an die Ebene", erwartet)
	}
}
