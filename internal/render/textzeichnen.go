package render

// **Vom gesetzten Text zum Bild.**
//
// Getrennt von `satz.go`, weil es eine andere Frage ist: dort steht, **wie** ein
// Text umbricht und sich ausrichtet — hier, **wie** er auf die Fläche kommt.
// Der Umbruch lässt sich gegen eine Tabelle prüfen, ein Glyph nur gegen ein
// Goldbild; dieselbe Trennung, die schon [Schrift] und [Glyphenschrift]
// auseinanderhält.

import (
	"image"
	"slices"
	"strings"
)

// text setzt eine Textebene und zeichnet sie.
//
// # Eine fehlende Schrift wird benannt, nicht ersetzt
//
// **Eine Karte mit falscher Schrift sieht richtig aus und ist es nicht.** Das
// ist der Grund, warum hier kein Rückfall auf irgendeine vorhandene Schrift
// steht: wer eine Karte mit `Helvetica` statt `MPlantin` gedruckt bekommt,
// merkt es an dem Tag, an dem die Karten aus der Druckerei kommen.
//
// # Und eine Schrift ohne Glyphen wird ebenso benannt
//
// Gesetzt wird mit jeder `Schrift` — dafür genügt eine Breitentabelle, und der
// Umbruch aus #31 ist an genau so einer geprüft. Gezeichnet wird nur mit einer
// `Glyphenschrift`. Wer eine Schrift ohne Glyphen anhängt, bekommt einen Befund
// je Textebene, keine stille leere Karte (§8).
func (r *Rasterer) text(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	guete Guete,
) []string {
	if r.schriften == nil {
		return []string{"für Text ist keine Schriftsuche eingerichtet"}
	}

	name := ebene.Satz.Schrift
	if strings.TrimSpace(name) == "" {
		return []string{"eine Textebene ohne Schriftnamen"}
	}

	schrift, err := r.schriften.Suche(name, false)
	if err != nil {
		return []string{"die Schrift " + name + " ist nicht zu finden: " + err.Error()}
	}

	schriften := r.schnitte(schrift, name, strings.TrimSpace(ebene.Satz.Kursivschrift))

	gesetzt := SetzeAbschnitte(ebene.Satz, ebene.Rahmen, ebene.Abschnitte, schriften, r.quelle)

	if _, kann := schrift.(Glyphenschrift); !kann {
		return []string{"Text ist gesetzt (" + zahlwort(len(gesetzt.Zeilen)) +
			"), aber die Schrift " + name + " gibt keine Glyphen her"}
	}

	// **Zwei Mängel, zwei Meldungen.** Ein Textzug kann Symbole tragen (`<sym>`),
	// und ein Symbol, dessen Bild fehlt, ist kein fehlender Glyph. Bis
	// #422
	// fielen beide in einen Satz, und der schickte jeden Leser zur Schrift.
	//
	// **Und zwei Befunde, nicht einer.** Bis QA-03 (#1656) standen beide in
	// einer Meldung, und `ArtAus` ordnete sie nach dem ersten Treffer: die
	// fehlenden Zeichen (Umgebung) verdeckten ein fehlendes Bild (Portierung).
	zeichen, bilder, symbolschriften := r.zeichneZeilen(
		ziel, ebene, maske, vergroesserung, gesetzt, schriften, guete)

	var maengel []string
	if zeichen != "" {
		maengel = append(maengel, "in der Schrift "+name+" fehlen Zeichen: "+zeichen)
	}

	return append(maengel, symbolmaengel(bilder, symbolschriften)...)
}

// schnitte sucht die drei Nebenschnitte zur aufrechten Schrift.
//
// **Alle drei sind Bestenfalls-Arbeit.** Was nicht zu finden ist, bleibt leer,
// und [Schriftsatz.fuer] fällt darauf zurück, was da ist — ein Flavortext, der
// aufrecht steht, ist eine Karte; einer, der fehlt, ist keine (§8).
//
// **Fett wird als Gewicht angefragt, nicht als Name.** Wie die Suche daraus
// eine Datei macht, ist ihre Sache
// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
func (r *Rasterer) schnitte(aufrecht Schrift, name, kursivname string) Schriftsatz {
	schriften := Schriftsatz{Aufrecht: aufrecht, Ersatz: r.ersatz}

	if fett, err := r.schriften.Suche(name, true); err == nil {
		schriften.Fett = fett
	}

	schriften.Kursiv, schriften.KursivFett = Kursivschnitte(r.schriften, name, kursivname)

	return schriften
}

// zeichneZeilen malt jede gesetzte Zeile und nennt die Zeichen, die fehlen.
func (r *Rasterer) zeichneZeilen(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	gesetzt Satzergebnis, schriften Schriftsatz, guete Guete,
) (zeichen string, bilder, symbolschriften []string) {
	// **Die gesetzte Größe gilt überall** — für die Glyphen, für die Grundlinie
	// und für den Vorschub. Der erste Anlauf rasterte mit `gesetzt.Groesse` und
	// rechnete Grundlinie und Vorschub weiter mit `ebene.Satz.Groesse`; die
	// Abstände gehörten dann zu einer anderen Größe als die Glyphen.
	// Angestrichen vom Zweitblick über die Phasen B bis D.
	satz := ebene.Satz
	satz.Groesse = gesetzt.Groesse

	rahmen := skaliert(ebene.Rahmen, vergroesserung)

	// **Die Stauchung wirkt erst hier.** Der Satz misst ungestaucht und gibt
	// den Faktor her; angewandt wird er auf **x** — auf die Stellen und auf die
	// Breite der Glyphen, nicht auf ihre Höhe (#319).
	gestaucht := gestauchtVon(ebene.Rahmen.Links, gesetzt.Stauchung)

	var fehlendeZeichen []string

	for _, zeile := range gesetzt.Zeilen {
		for _, stueck := range zuege(zeile.Stuecke) {
			if len(stueck.Symbole) > 0 {
				fehlend, ohneSchrift := r.zeichneSymbolstueck(ziel, ebene, maske, rahmen, stueck,
					zeile.Oben, zeile.Hoehe, vergroesserung, gestaucht, guete)
				bilder = append(bilder, fehlend...)
				symbolschriften = append(symbolschriften, ohneSchrift...)

				continue
			}

			// **Ein weiches Stück wird gemessen, aber nicht gezeichnet.**
			//
			//	DrawWhat what = soft > 0 ? DRAW_ACTIVE : DRAW_NORMAL;
			//	…
			//	if ((what & draw_as) != draw_as) return; // don't draw
			//
			// `src/render/text/element.cpp`, Zeile 234, und
			// `FontTextElement::draw`, `src/render/text/font.cpp`, Zeilen
			// 15–16, Fassung f9d9356 vom 2021-07-06. `DRAW_ACTIVE` heißt
			// „draw active editor stuff, such as hidden separators"
			// (`src/data/draw_what.hpp`, Zeile 23) und wird allein vom
			// **Editor** gesetzt — auf einer Karte nie.
			//
			// **Erst hier und nicht im Satz.** Das Stück behält seine Breite:
			// `getCharInfo` misst es wie jedes andere (`font.cpp`, Zeilen
			// 27–48). Wer es schon beim Setzen weggelassen hätte, rückte den
			// Text dahinter nach links.
			//
			// **Und der Symbolzweig steht davor** (siehe oben): ein
			// Symbolstück in weicher Klammer wird gezeichnet, denn
			// `SymbolTextElement::draw` fragt nach `DRAW_NORMAL` und kennt
			// kein `draw_as` (`src/render/text/symbol.cpp`, Zeilen 15–20).
			if stueck.Weich {
				continue
			}

			// **Jedes Stück mit seiner eigenen Schrift.** Ein kursives trägt
			// andere Glyphen, andere Breiten und eine andere Grundlinie — mit
			// der aufrechten gezeichnet säße es auf einer Linie, die zu einer
			// anderen Schrift gehört.
			schrift, kann := schriften.fuer(stueck.Kursiv, stueck.Fett).(Glyphenschrift)
			if !kann {
				continue
			}

			// **Die Grundlinie, nicht die Oberkante.** [SetzeAbschnitte] gibt
			// die obere Kante der Zeile; ein Glyph sitzt auf der Grundlinie,
			// und der Abstand dahin ist eine Eigenschaft der Schrift, keine
			// Zahl, die man raten kann.
			grundlinie := zeilengrundlinie(zeile, satz, schrift) * vergroesserung

			fehlendeZeichen = append(fehlendeZeichen,
				r.zeichneStueck(ziel, ebene, maske, rahmen, stueck, satz, schrift,
					schriften.Ersatz,
					vergroesserung, zeile.Oben*vergroesserung, grundlinie, fehlendeZeichen, gestaucht)...)
		}
	}

	return strings.Join(fehlendeZeichen, " "), bilder, symbolschriften
}

// zuege fasst die Stücke einer Zeile zu den Zügen zusammen, in denen MSE sie
// zeichnet — #2302.
//
// **Eine Zeile je Element in einem Zug**: `elements.draw(dc, scale, rect, …,
// l.start, l.end())` (`TextViewer::draw`, `src/render/text/viewer.cpp`,
// Zeilen 93–96), und `FontTextElement::draw` gibt den ganzen Ausschnitt an
// `DrawTextWithShadow` (`src/render/text/font.cpp`, Zeilen 15–25), Fassung
// f9d9356 vom 2021-07-06. Wo die Wörter darin stehen, bestimmen die
// Vorschübe der Schrift; die Stellen, die der Umbruch auf das
// Viertelpunktraster gelegt hat, gelten nur für den Anfang des Zugs. Hier
// stand jedes Wort an seiner gerasterten Stelle, und seit ein gestauchtes
// Feld umbricht (#2308), traf das auch den Namen: die Vanguard-Avatare lagen
// im Namensstreifen bei 38 statt 1,2.
//
// **Den Blocksatz** zeichnet MSE Zeichen für Zeichen (`viewer.cpp`, Zeilen
// 87–92); ihn setzt dieser Rasterer nicht. **Symbole** sind eigene Elemente
// und bleiben eigene Züge — ihre Stücke tragen keinen [Stueck.Lauf].
func zuege(stuecke []Stueck) []Stueck {
	zusammen := make([]Stueck, 0, len(stuecke))

	for _, stueck := range stuecke {
		if letztes := len(zusammen) - 1; letztes >= 0 && stueck.Lauf != 0 &&
			zusammen[letztes].Lauf == stueck.Lauf {
			zusammen[letztes].Text += stueck.Text
			zusammen[letztes].Breite = stueck.Links + stueck.Breite - zusammen[letztes].Links

			continue
		}

		zusammen = append(zusammen, stueck)
	}

	return zusammen
}

// zeichneSymbolstueck malt die Symbole eines Stücks nebeneinander.
//
// # Warum hier und nicht als eigene Ebene
//
// Ein Symbolfeld — die Manakosten — wird in `internal/service` in eigene
// Ebenen zerlegt; dort steht der Zug allein im Feld und seine Stelle ist
// ausgerechnet. **Mitten im Regeltext geht das nicht**: wo ein `<sym>T</sym>`
// landet, weiß erst der Umbruch, und der findet hier statt.
//
// **In der Zeile und nicht auf ihrer Oberkante.** MSE richtet das Bild in
// seinem Kasten aus, und der ist die ganze Zeilenhöhe; ohne Angabe mittig
// (#330).
func (r *Rasterer) zeichneSymbolstueck(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	stueck Stueck, oben, zeilenhoehe, vergroesserung float64, gestaucht stauchflaeche,
	guete Guete,
) (fehlend, schriften []string) {
	return r.zeichneSymbole(ziel, ebene, maske, rahmen,
		stueck.Symbole, stueck.Symbolmasse,
		stueck.Links, oben, zeilenhoehe, vergroesserung, gestaucht, guete)
}

// stauchflaeche rechnet eine Stelle und eine Breite in die gestauchte Fläche.
//
// **Die linke Kante des Feldes bleibt stehen.** MSE staucht über die
// Zeichenfläche, deren Ursprung der Kasten ist; alles rechts davon rückt zu ihr
// hin (`src/render/value/text.cpp`, Zeile 26, Fassung f9d9356 vom 2021-07-06).
type stauchflaeche struct {
	kante  float64
	faktor float64
}

// **Ohne Prüfung auf einen Faktor von null.** [SetzeAbschnitte] setzt ihn
// immer, und [stauchung] gibt nie null zurück; ein Zweig dafür wäre einer, den
// kein Aufruf betreten kann (§3).
func gestauchtVon(kante, faktor float64) stauchflaeche {
	return stauchflaeche{kante: kante, faktor: faktor}
}

// stelle gibt eine waagerechte Stelle in der gestauchten Fläche.
func (s stauchflaeche) stelle(x float64) float64 {
	return s.kante + (x-s.kante)*s.faktor
}

// breite staucht ein waagerechtes Maß.
func (s stauchflaeche) breite(b float64) float64 { return b * s.faktor }

// zeichneStueck malt ein Stück einer Zeile und nennt die Zeichen, die fehlen.
//
// **Vierfach und heruntergerechnet, wie MSE** — der Weg steht bei
// [vierfachlauf]. `oben` und `grundlinie` sind Bildpunkte im Zielbild.
func (r *Rasterer) zeichneStueck(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	stueck Stueck, satz Satz, schrift Glyphenschrift, ersatz Ersatzschrift,
	vergroesserung, oben, grundlinie float64, bisher []string, gestaucht stauchflaeche,
) []string {
	var fehlend []string

	groesse := satz.Groesse * vergroesserung

	// **Das Textmaß ist ungestaucht** — gestaucht wird erst beim
	// Herunterrechnen, über das ganze Stück, wie MSE es tut.
	links := gestaucht.stelle(stueck.Links) * vergroesserung
	hoehe := gerastertesTextmass(schrift.Zeilenhoehe(), satz.Groesse) * vergroesserung

	// **Der Puffer ist so breit wie der gezeichnete Text**, frisch gemessen:
	// `RealRect r(pos, GetTextExtent(text))` (`RotatedDC::DrawText`,
	// `src/util/rotation.cpp`, Zeilen 189–207, Fassung f9d9356 vom
	// 2021-07-06). Die Breite aus dem Umbruch hängt am Text davor — an der
	// Vorsilbe des Abschnitts und an der Kernung über die Bruchkante —, und
	// unter Stauchung rechnet der Puffer aus seiner Breite das Bild
	// (Nachlauf über die Folgen vor 0.8.0, Befund 3).
	pufferbreite := (&vorsilbe{ersatz: ersatz}).nimm(stueck.Text, satz, schrift) * vergroesserung
	lauf := neuerVierfachlauf(links, oben, pufferbreite, hoehe, grundlinie-oben, gestaucht.faktor)

	// **Der Schatten ist derselbe Text, an anderer Stelle** — ein eigener
	// Lauf, denn ein Versatz, der nicht auf ganze Bildpunkte fällt, rastert
	// anders ([schattenlauf]).
	schatten := schattenlauf(satz.Schatten, links, oben, pufferbreite, hoehe,
		grundlinie-oben, gestaucht.faktor, vergroesserung)

	// **Über die Runen mit ihrer Stelle**, denn der Vorschub eines Zeichens
	// hängt davon ab, was **danach** kommt: eine kombinierende Marke gehört
	// noch zu ihm, und dann fällt die Laufweite dazwischen weg.
	runen := gesetzt(stueck.Text, satz, schrift)
	versatz := 0.0

	for stelle, zeichen := range runen {
		markeFolgt := stelle+1 < len(runen) && !istBuchstabe(runen[stelle+1])

		glyph, ecke, da := schrift.Maske(zeichen, groesse*Textraster)

		// **Fehlt das Zeichen, wird ausgewichen** — der Grund steht bei
		// [Schriftsatz.Ersatz]. Auch der Vorschub kommt dann aus der
		// Ersatzschrift: ein Glyph aus der einen und ein Abstand aus der
		// anderen ergäbe einen Satz, den keine der beiden meint.
		verwendet := Schrift(schrift)

		if !da && ersatz != nil {
			if andere, gefunden := ersatz(zeichen); gefunden {
				if glyph, ecke, da = andere.Maske(zeichen, groesse*Textraster); da {
					verwendet = andere
				}
			}
		}

		// **Das vorige Zeichen desselben Zugs** (#752, #2302). Ein Zug ist ein
		// Abschnitt einer Zeile ([zuege]), und in ihm kernt MSE über die
		// Wortgrenzen hinweg wie jedes `DrawText`. Zwischen zwei Zügen nicht:
		// jeder steht an seiner gerechneten Stelle.
		vorher := rune(0)
		if stelle > 0 {
			vorher = runen[stelle-1]
		}

		breite := vorschub(vorher, zeichen, satz, verwendet, vergroesserung, markeFolgt)

		if !da {
			// **Ein unsichtbares Zeichen fehlt nicht** — es hat keine Gestalt
			// ([unsichtbar]).
			if name := zeichenname(zeichen); !unsichtbar(zeichen) &&
				!slices.Contains(bisher, name) &&
				!slices.Contains(fehlend, name) {
				fehlend = append(fehlend, name)
			}

			versatz += breite

			continue
		}

		// **Die Korrektur vor die Glyphe, der Vorschub dahinter** — sonst
		// rückte das Paar `AV` nicht das `V` heran, sondern das Zeichen danach
		// (QA-03).
		stelleImLauf := versatz + kernungZwischen(vorher, zeichen, satz, verwendet)*vergroesserung
		lauf.setze(glyph, ecke, stelleImLauf)

		if schatten != nil {
			schatten.setze(glyph, ecke, stelleImLauf)
		}

		versatz += breite
	}

	// **Erst der Schatten, dann der Text** — je Stück, wie MSE es je
	// Textelement tut (`DrawTextWithShadow`).
	if schatten != nil {
		r.legeSchatten(ziel, ebene, maske, rahmen, schatten, satz.Schatten)
	}

	r.lege(ziel, ebene, maske, rahmen, lauf)

	return fehlend
}

// zeilengrundlinie gibt die Grundlinie einer Zeile in Vorlagenpunkten — die
// Oberkante plus der Abstand der Schrift, gerastert, wie MSE ihn misst. Siehe
// [gerastertesTextmass].
func zeilengrundlinie(zeile Zeile, satz Satz, schrift Glyphenschrift) float64 {
	return zeile.Oben + gerastertesTextmass(schrift.Grundlinie(), satz.Groesse)
}
