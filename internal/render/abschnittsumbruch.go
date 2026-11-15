package render

import "strings"

// zeilenAus bricht einen Zug aus Abschnitten in Zeilen.
//
// **breite gibt die Breite je Zeile**, gezählt über den ganzen Zug — mit der
// Maske eines Felds ist jede Zeile anders breit ([zeilenraum]).
func zeilenAus(
	abschnitte []Abschnitt, breite func(zeile int) float64, satz Satz,
	schriften Schriftsatz, bilder Bildquelle,
) []rohzeile {
	zeilen := make([]rohzeile, 0, len(abschnitte))

	for nummer, absatz := range absaetze(abschnitte) {
		// **Der linke Rand verkürzt die Zeile, der rechte nicht.** Bei MSE
		// beginnt `line_size` bei `line.margin_left` (`fitLineWidth`), und
		// gebrochen wird, sobald eine Zeile über `lineRight` hinausreicht —
		// den Rand der Maske abzüglich `padding_right`, **ohne**
		// `margin_right` (`src/render/text/viewer.cpp`, Zeilen 548 und
		// 614–616, `lineRight` Zeilen 739–741). Der rechte Rand wirkt beim
		// Abrücken ([zeilenraum.abgerueckt]) und beim Ausrichten
		// (`alignHorizontal`, Zeile 835). Fassung f9d9356 vom 2021-07-06.
		//
		// Hier stand bis zum Nachlauf über E/F, auch der rechte verkürze die
		// Breite; ein Absatz mit `<margin:…:r:…>` brach früher um als bei MSE
		// und konnte dadurch verkleinert werden (Befund 2).
		erste := len(zeilen)
		frei := func(zeile int) float64 {
			return max(breite(erste+zeile)-absatz.margen.Links, 0)
		}

		gebrochen, einzug := absatzInZeilen(absatz.abschnitte, frei, satz, schriften, bilder)

		for stelle, stuecke := range gebrochen {
			letzte := stelle == len(gebrochen)-1

			zeilen = append(zeilen, rohzeile{
				stuecke: stuecke,
				margen:  absatz.margen,
				// **Die erste Zeile trägt den Punkt, die anderen den Einzug.**
				einzug: map[bool]float64{true: 0, false: einzug}[stelle == 0],
				// **Der obere Rand steht nur an der ersten Zeile.** Er kommt
				// bei MSE auf `line.top` beim Absatz**wechsel**, nicht bei
				// jedem Umbruch (`viewer.cpp:694`).
				//
				// **Und ein weicher Umbruch ist kein Absatzwechsel**:
				//
				//	void addParagraphs(const String& text, size_t start, size_t end) {
				//	  if (line == 0 && soft_line > 0) return;
				//	  …
				//	}
				//
				// `src/render/text/element.cpp`, Zeilen 272–274, Fassung
				// f9d9356 vom 2021-07-06. Steht der Umbruch in einem
				// `<soft-line>` und **nicht** in einem `<line>`, entsteht gar
				// kein Absatz — also auch kein neuer oberer Rand.
				//
				// **Was das gekostet hat**: `Ranger Class` trägt zwei weiche
				// Umbrüche innerhalb eines `<margin:0:0:30>`. Wir legten den
				// Rand zweimal neu auf — 60 Punkte, die MSE nicht setzt
				// (#584).
				//
				// **Und nie am ersten Absatz**: MSE legt die erste Zeile auf
				// `padding_top` und liest `margin_top` nur beim Wechsel
				// (`viewer.cpp`, Zeilen 566–569 und 694). Hier stand bis zum
				// 2026-10-11 das Gegenteil, und ein Text, der mit
				// `<margin:…:…:t>` beginnt, stand um den Rand zu tief.
				margeOben: stelle == 0 && !absatz.folgtWeich && nummer > 0,
				// **Ein Block endet mit dem Absatz, nicht mitten darin.** Die
				// Auszeichnung `<line>` steht bei MSE um einen Zeilenumbruch
				// herum (`<line>\n</line>`); geschnitten wird, wo der Umbruch
				// fällt.
				blockende: letzte && absatz.blockende,
				// **Ein weicher Umbruch schließt keinen Absatz.** Er bekommt
				// damit `line height soft` und zählt nicht als Absatzgrenze —
				// beides folgt bei MSE aus derselben Auszeichnung.
				weich: letzte && absatz.weich,
				// **Nur die letzte Zeile eines Absatzes schließt ihn.** Die
				// anderen sind Umbrüche, keine Absatzenden — MSE unterscheidet
				// `LineBreak::SOFT` von `LineBreak::HARD`.
				absatzende: letzte,
			})
		}
	}

	return zeilen
}

// rohzeile ist eine umgebrochene Zeile samt der Auskunft, ob sie einen Absatz
// schließt.
type rohzeile struct {
	stuecke    []Stueck
	absatzende bool
	blockende  bool

	// weich sagt, dass diese Zeile mit einem `<soft-line>`-Umbruch endet.
	weich bool

	// margen sind die Ränder ihres Absatzes; margeOben sagt, ob sie die
	// **erste** Zeile darin ist und den oberen Rand mitbringt.
	margen    Margen
	margeOben bool

	// einzug ist der hängende Einzug aus `<bullet>` — siehe
	// [Abschnitt.ImRand]. Auf der ersten Zeile eines Absatzes ist er null.
	einzug float64
}

// margenOben gibt den oberen Absatzrand, sofern diese Zeile ihn mitbringt.
func (z rohzeile) margenOben() float64 {
	if !z.margeOben {
		return 0
	}

	return z.margen.Oben
}

// absatz ist eine Folge von Abschnitten zwischen zwei harten Umbrüchen.
type absatz struct {
	abschnitte []Abschnitt

	// blockende sagt, dass hinter diesem Absatz ein **Block** endet — die
	// Auszeichnung `<line>` (siehe [Abschnitt.Blockende]).
	blockende bool

	// weich sagt, dass dieser „Absatz" von einem `<soft-line>`-Umbruch
	// abgeschlossen wird und damit gar keiner ist (siehe
	// [Abschnitt.WeicheZeile]).
	weich bool

	// folgtWeich sagt, dass er **auf** einen solchen Umbruch folgt — dann ist
	// er die Fortsetzung des vorigen und bringt dessen oberen Rand nicht noch
	// einmal mit.
	folgtWeich bool

	// margen sind seine Ränder — siehe [Abschnitt.Margen].
	margen Margen
}

// absaetze teilt die Abschnitte an harten Zeilenumbrüchen.
//
// **Der Umbruch kann mitten in einem Abschnitt stehen** — `<i-flavor>` beginnt
// in `magic-m15` regelmäßig hinter einem `\n`, und der Abschnitt danach trägt
// ihn mit. Ein Teiler, der nur zwischen Abschnitten schnitte, verlöre ihn.
//
// **Und die Blockgrenze hängt am Umbruch, nicht am Text.** `<line>\n</line>`
// ist ein Abschnitt, dessen Text **nur** aus dem Umbruch besteht; er ergibt
// keinen sichtbaren Teil, und ohne diese Zuordnung ginge die Grenze verloren
// (#338).
func absaetze(abschnitte []Abschnitt) []absatz {
	laufend := []Abschnitt{}
	alle := []absatz{}
	blockende := false
	folgtWeich := false

	// **Der Rand gilt für den Absatz, in dem die Auszeichnung steht** — auch
	// wenn sie erst in seiner Mitte auftaucht. MSE schreibt sie an
	// `paragraphs.back()`, also den laufenden.
	var margen Margen

	for _, abschnitt := range abschnitte {
		teile := strings.Split(abschnitt.Text, "\n")

		for nummer, teil := range teile {
			if nummer > 0 {
				// **`<line>` schlägt `<soft-line>`.** MSE fragt
				// `line == 0 && soft_line > 0`; innerhalb eines `<line>` gilt
				// der Umbruch also weiter als Absatzende
				// (`element.cpp:274`).
				weich := abschnitt.WeicheZeile && !abschnitt.Blockende

				alle = append(alle, absatz{
					abschnitte: laufend, blockende: blockende, margen: margen,
					weich: weich, folgtWeich: folgtWeich,
				})
				laufend = []Abschnitt{}
				blockende = false
				margen = Margen{}
				folgtWeich = weich
			}

			if abschnitt.Margen != nil {
				margen = *abschnitt.Margen
			}

			// **Nur beim ersten Teil.** Ein Abschnitt mit `<line>` schließt den
			// Absatz, in dem er **beginnt** — bei `<line>\n</line>` ist das der
			// davor. Stünde die Zuweisung für jeden Teil, trüge auch der Absatz
			// **danach** die Grenze, und der Zug rückte zweimal weit.
			if abschnitt.Blockende && nummer == 0 {
				blockende = true
			}

			if teil != "" || abschnitt.istSymbolisch() {
				laufend = append(laufend, Abschnitt{
					Text: teil, Kursiv: abschnitt.Kursiv, Fett: abschnitt.Fett,
					Weich:   abschnitt.Weich,
					Symbole: abschnitt.Symbole, Symbolmasse: abschnitt.Symbolmasse,
					Blockende:   abschnitt.Blockende,
					WeicheZeile: abschnitt.WeicheZeile,
					Margen:      abschnitt.Margen,
					ImRand:      abschnitt.ImRand,
				})
			}
		}
	}

	return append(alle, absatz{
		abschnitte: laufend, blockende: blockende, margen: margen,
		folgtWeich: folgtWeich,
	})
}

// worteAus zerlegt einen Absatz in Wörter samt ihrer Trennlücken.
//
// Ein Wort ist eine Folge von Stücken ohne Leerzeichen dazwischen; die Lücke
// **davor** gehört dazu, damit sie beim Zeilenwechsel mit ihr wegfällt und
// sonst in dem Schnitt gemessen wird, in dem sie stand.
func worteAus(
	absatz []Abschnitt, satz Satz, schriften Schriftsatz, bilder Bildquelle,
) [][]Stueck {
	worte := [][]Stueck{}
	laufend := make([]Stueck, 0, 2)

	for _, teil := range teileAus(absatz, satz, schriften, bilder) {
		// **Eine Lücke beginnt ein neues Wort — außer es besteht selbst noch
		// nur aus Lücken.** Zwei Abschnitte, von denen der eine auf ein
		// Leerzeichen endet und der nächste mit einem beginnt, ergeben zwei
		// Lückenstücke nebeneinander; ohne diese Bedingung entstünde daraus ein
		// „Wort" aus lauter Leerzeichen, das der Umbruch wie ein Wort behandelte.
		if istLuecke(teil.Text) && len(laufend) > 0 && !nurLuecken(laufend) {
			worte = append(worte, laufend)
			laufend = nil
		}

		laufend = append(laufend, teil)
	}

	if len(laufend) > 0 && !nurLuecken(laufend) {
		worte = append(worte, laufend)
	}

	return worte
}

// teileAus zerlegt die Abschnitte in Stücke aus Lücke oder Nichtlücke.
//
// **Und misst sie gleich mit** (#598). Die Breite eines Stücks ist die
// Differenz zweier gerasterter Vorsilben seines **Abschnitts** — nicht die
// gerasterte Breite des Stücks. Der Grund steht in `vorsilbenbreite.go`; der
// erste Anlauf hat je Stück gerastert und den Abstand um 4,05 verschlechtert.
//
// **Gemessen wird hier und nicht beim Verteilen**, weil die Vorsilbe den
// Abschnitt kennen muss und die späteren Wege nur noch Stücke sehen.
func teileAus(
	absatz []Abschnitt, satz Satz, schriften Schriftsatz, bilder Bildquelle,
) []Stueck {
	var teile []Stueck

	for nummer, abschnitt := range absatz {
		// **Ein symbolischer Abschnitt ist ein Stück und bleibt eines.** Seine
		// Zeichenkette ist der rohe Code — `2/R` enthält keinen Leerraum, aber
		// darauf ist kein Verlass, und ein an einer Lücke zerschnittener
		// Symbolzug verlöre die Zuordnung zu seinen Bildern.
		if abschnitt.istSymbolisch() {
			stueck := Stueck{
				Text: abschnitt.Text, Kursiv: abschnitt.Kursiv, Fett: abschnitt.Fett,
				Symbole: abschnitt.Symbole, Symbolmasse: abschnitt.Symbolmasse,
				Weich: abschnitt.Weich, ImRand: abschnitt.ImRand,
			}
			stueck.Breite = symbolbreite(stueck, bilder)

			teile = append(teile, stueck)

			continue
		}

		// **Ein Lauf, eine Vorsilbe.** Sie beginnt mit jedem Abschnitt neu —
		// bei MSE ist jeder Schriftlauf ein eigenes Element, und `prev_width`
		// startet darin bei null (`font.cpp`, Zeile 31).
		lauf := vorsilbe{ersatz: schriften.Ersatz}
		schrift := schriften.fuer(abschnitt.Kursiv, abschnitt.Fett)

		for _, teil := range zerlegeAnLuecken(abschnitt.Text) {
			teile = append(teile, Stueck{
				Text: teil, Kursiv: abschnitt.Kursiv, Fett: abschnitt.Fett,
				Weich:     abschnitt.Weich,
				Blockende: abschnitt.Blockende,
				ImRand:    abschnitt.ImRand,
				Breite:    lauf.nimm(teil, satz, schrift),
				Lauf:      nummer + 1,
			})
		}
	}

	return teile
}

// zerlegeAnLuecken teilt einen Text in Läufe aus Leerzeichen und Nichtleerzeichen.
func zerlegeAnLuecken(text string) []string {
	var (
		teile    []string
		laufend  strings.Builder
		inLuecke bool
	)

	for nummer, zeichen := range text {
		luecke := zeichen == ' ' || zeichen == '\t'

		if nummer > 0 && luecke != inLuecke {
			teile = append(teile, laufend.String())
			laufend.Reset()
		}

		inLuecke = luecke

		laufend.WriteRune(zeichen)
	}

	if laufend.Len() > 0 {
		teile = append(teile, laufend.String())
	}

	return teile
}

// istLuecke sagt, ob ein Stück nur aus Leerzeichen besteht.
func istLuecke(text string) bool {
	return text != "" && strings.TrimLeft(text, " \t") == ""
}

// nurLuecken sagt, ob eine Folge von Stücken nichts als Lücken enthält.
func nurLuecken(stuecke []Stueck) bool {
	for _, stueck := range stuecke {
		if !istLuecke(stueck.Text) {
			return false
		}
	}

	return true
}

// ohneLuecke wirft die führenden Lücken eines Worts weg.
func ohneLuecke(wort []Stueck) []Stueck {
	schnitt := 0
	for schnitt < len(wort) && istLuecke(wort[schnitt].Text) {
		schnitt++
	}

	return append([]Stueck{}, wort[schnitt:]...)
}
