package render

import "unicode"

// Schrift misst Text. Zeichnen ist eine andere Aufgabe.
//
// **Getrennt, weil das Messen prüfbar ist und das Zeichnen nicht.** Ein
// Umbruch, der an einer bestimmten Stelle fällt, lässt sich gegen eine Tabelle
// halten; ein Glyph auf einer Fläche nur gegen ein Goldbild (#35). Wer beides
// in eine Schnittstelle legt, kann den Umbruch nur noch mit einer echten
// Schrift prüfen — und dann prüft der Test die Schrift mit.
type Schrift interface {
	// Breite eines Zeichens in Vorlagenpunkten, bei Schriftgröße 1.
	Breite(zeichen rune) float64

	// Zeilenhoehe bei Schriftgröße 1, von Grundlinie zu Grundlinie.
	Zeilenhoehe() float64
}

// Kernschrift kann sagen, was zwischen zwei Zeichen dazukommt —
// #752.
//
// **Eine eigene Schnittstelle und kein Feld an [Schrift]**, aus zwei Gründen:
// eine Schrift ohne Kerningtabelle soll nichts davon wissen müssen, und jede
// Attrappe in den Tests müsste sonst eine Methode mitbringen, die ihr Fall
// nicht braucht (§3).
type Kernschrift interface {
	// Kernung ist die Verschiebung zwischen zwei Zeichen, bei Schriftgröße 1.
	Kernung(vorher, jetzt rune) float64
}

// kernungZwischen ist **die** Stelle, an der Kerning in eine Breite eingeht.
//
// # Warum es eine Funktion ist und keine vier Zeilen an vier Stellen
//
// Weil es vier Summenstellen über dieselbe Sache gibt — der Umbruch
// ([vorsilbe.nimm]), der Vorschub beim Zeichnen ([vorschub]), die Zeilenbreite
// für die Ausrichtung ([breiteVon]) und die breiteste Zeile eines Blocks
// (`textbreite.go`). Der Kommentar an [vorschub] sagt, was daraus wird:
// **zwei Rechnungen über dieselbe Sache laufen auseinander, sobald eine von
// ihnen etwas lernt.**
//
// **Und der erste Anlauf hat genau das vorgeführt**: Kerning ging nur in die
// gemessene Breite und nicht in den Vorschub. Die Zeile wurde schmaler gemessen
// und weiter ungekernt gesetzt — die Ausrichtung verschob sich, die Glyphen
// standen wie vorher, und der Abstand am Kanon **stieg**.
//
// **`vorher == 0` heißt „kein vorheriges Zeichen"** — am Anfang eines Zuges
// gibt es kein Paar.
func kernungZwischen(vorher, jetzt rune, satz Satz, schrift Schrift) float64 {
	if vorher == 0 || schrift == nil {
		return 0
	}

	kernt, kann := schrift.(Kernschrift)
	if !kann {
		return 0
	}

	return kernt.Kernung(vorher, jetzt) * satz.Groesse
}

// Schriftquelle sucht eine Schrift im System.
//
// **Gesucht, nicht mitgeliefert** (`DRITTANBIETER.md` §5). Schriften sind
// lizenziert; eine mitzuliefern hieße, ihre Lizenz zu prüfen und einzuhalten,
// und für `MPlantin` oder `Beleren` ginge das nicht.
//
// Der Suchweg ist plattformabhängig und deshalb hier nur eine Schnittstelle;
// wie er auf macOS aussieht und wie er anderswo aussähe, steht in
// `docs/PORTIERUNG.md`.
type Schriftquelle interface {
	// Suche gibt die Schrift zu einem Namen — im fetten Schnitt, wenn `fett`.
	//
	// **Ein Gewicht und kein zweiter Name.** Kursiv nennt der Stil unter
	// `italic name:`, fett nicht: MSE reicht Familienname und Gewicht an das
	// Betriebssystem weiter (`src/data/font.cpp`, Zeilen 82–108, Fassung
	// f9d9356 vom 2021-07-06). Wie daraus wieder ein Name wird, entscheidet
	// die Suche und nicht der Rasterer
	// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
	Suche(name string, fett bool) (Schrift, error)
}

// Zeile ist eine gesetzte Textzeile.
type Zeile struct {
	Text string

	// Stuecke sind die Abschnitte dieser Zeile, jeder mit eigenem Schnitt und
	// eigener Stelle. [Zeile.Text] ist ihre Verkettung.
	//
	// **Zwei Sichten auf dieselbe Zeile, und eine davon ist abgeleitet.** Wer
	// nur misst, liest `Text`; wer zeichnet, braucht die Stücke — eine kursive
	// Stelle mitten im Satz hat eine andere Schrift und damit andere Glyphen.
	Stuecke []Stueck

	// Links und Oben sind die linke obere Ecke in Vorlagenpunkten.
	Links, Oben float64

	// Hoehe ist die Höhe **dieser** Zeile.
	//
	// **Je Zeile und nicht je Zug**, denn ein Symbol darf sie heben — siehe
	// [zeilenhoehe]. Solange nur Text dasteht, sind alle Zeilen gleich hoch.
	Hoehe float64

	// Breite ist die gemessene Breite samt Laufweite.
	Breite float64
}

// Zeilengruppe ist eine Folge von Zeilen mit gemeinsamem Maß.
//
// **Absätze und Blöcke** — MSE gibt beide an die Vorlage weiter
// (`LineLayout`, `src/data/field/text.hpp`, Zeilen 50–67, Fassung f9d9356 vom
// 2021-07-06). Ein **Absatz** endet an einem harten Umbruch, ein **Block** an
// einem `<line>`.
//
// Die Maße folgen `update_size` (`src/render/text/viewer.cpp`, Zeilen
// 350–353): die Breite ist die **größte** Zeilenbreite, die Höhe reicht von der
// Oberkante der ersten bis zur Unterkante der letzten Zeile.
type Zeilengruppe struct {
	Breite, Oben, Hoehe float64

	// Zeilen sind die Zeilen dieser Gruppe.
	Zeilen []Zeile
}

// Satzergebnis ist der fertige Umbruch.
type Satzergebnis struct {
	Zeilen []Zeile

	// Absaetze und Bloecke gruppieren [Satzergebnis.Zeilen].
	//
	// **Vorlagen rechnen damit.** `card_style.<feld>.layout.blocks[i].middle`
	// setzt bei den Planeswalker-Stilen von `M15-Magic-Pack` jede
	// Fähigkeitszeile; ohne das Objekt sitzt sie bei −30
	// (#338).
	Absaetze []Zeilengruppe
	Bloecke  []Zeilengruppe

	// Trenner sind die Mitten **zwischen** zwei Blöcken — dort zeichnet MSE
	// eine Trennlinie.
	//
	//	layout->separators.push_back((blocks[i]->bottom() + blocks[i+1]->top)/2);
	//
	// `src/render/text/viewer.cpp`, Zeilen 386–388.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Trenner []float64

	// Hoehe ist, was der Text insgesamt braucht. Sie kann größer sein als der
	// Rahmen — dann steht mehr Text da, als hineinpasst, und der Aufrufer
	// entscheidet, ob er kürzt oder verkleinert.
	Hoehe float64

	// Groesse ist die Schriftgröße, mit der gesetzt wurde.
	Groesse float64

	// Unterkante ist, woran MSE misst, ob der Text passt — vom oberen Rand
	// des Felds an, vor Dehnung und Ausrichtung ([unterkante]).
	Unterkante float64

	// Stauchung ist der **waagerechte** Faktor aus [Satz.Stauchen]; 1 heißt
	// „unverändert".
	//
	// **Der Satz gibt ihn her und wendet ihn nicht an.** Die Stellen und
	// Breiten in [Satzergebnis.Zeilen] sind ungestaucht — so misst MSE
	// `content_width`, und daraus rechnet es den Faktor überhaupt erst aus.
	// Angewandt wird er beim Zeichnen
	// (#319).
	Stauchung float64
}

// Setze bricht Text um und richtet ihn im Rahmen aus.
//
// # Umgebrochen wird an Leerzeichen, und ein zu langes Wort steht über
//
// Ein Wort, das allein nicht in die Zeile passt, bekommt seine eigene Zeile und
// ragt hinaus. **Es mitten im Wort zu trennen wäre die andere Wahl**, und sie
// ist schlechter: MSE-Kartentexte enthalten Zeichenfolgen wie `{T}` und
// `<i-flavor>`, und eine Trennung darin ergäbe eine Zeile, die niemand
// geschrieben hat.
func Setze(satz Satz, rahmen Rechteck, text string, schrift Schrift) Satzergebnis {
	// **Eine Maschine und nicht zwei.** Der Umbruch über Abschnitte kann alles,
	// was der über eine Zeichenkette kann; zwei Wege daneben liefen
	// auseinander, sobald einer von ihnen etwas lernt — dieselbe Begründung wie
	// bei [vorschub] und [breiteVon].
	// **Ohne Bildquelle**, denn ein Zug aus einer Zeichenkette trägt keine
	// Symbole — und eine Quelle, die niemand befragt, wäre eine Zusage, die
	// dieser Weg nicht einlöst.
	return SetzeAbschnitte(satz, rahmen, []Abschnitt{{Text: text}},
		Schriftsatz{Aufrecht: schrift}, nil)
}

// vorschub sagt, wie weit der Stift nach einem Zeichen weiterrückt.
//
// **Dieselbe Rechnung wie in [breiteVon]**, und das ist der Punkt: die Messung
// zählt die Laufweite zwischen **Buchstaben**, und die Zeichenschleife addierte
// sie nach jeder **Rune**. Die Zeile war damit richtig breit und ihre Glyphen
// falsch gesetzt — bei `e` plus U+0301 rückte alles hinter der Marke um eine
// Laufweite nach rechts. Gefunden vom Zweitblick durch Codex über Phase 6.
//
// **Zwei Rechnungen über dieselbe Sache laufen auseinander**, sobald eine von
// ihnen etwas lernt. Es gibt jetzt eine.
// **`markeFolgt` sagt, ob das nächste Zeichen zu demselben Buchstaben gehört.**
// Dann fällt die Laufweite weg: sie steht **hinter dem Buchstaben**, nicht
// hinter seinem Grundzeichen. Der erste Anlauf legte sie unmittelbar hinter das
// Grundzeichen — und damit vor die Marke, die danach neben ihrem Buchstaben saß
// statt über ihm. Der Nachlauf zum Zweitblick über Phase 6 hat es gefunden: die
// Korrektur aus #113 hatte die Zeichen **hinter** der Marke gerade gerückt und
// die Marke selbst stehen lassen — der Fall behandelt, die Klasse nicht.
func vorschub(
	vorher, zeichen rune, satz Satz, schrift Schrift,
	vergroesserung float64, markeFolgt bool,
) float64 {
	// **Das Kerningpaar gehört **vor** das Zeichen** (#752): der Stift rückt
	// erst um die Korrektur, dann um den Vorschub. MSE zeichnet ganze
	// Zeichenketten und lässt wx beides tun.
	breite := kernungZwischen(vorher, zeichen, satz, schrift) * vergroesserung
	breite += schrift.Breite(zeichen) * satz.Groesse * vergroesserung

	if istBuchstabe(zeichen) && !markeFolgt {
		breite += satz.Laufweite * vergroesserung
	}

	return breite
}

// istBuchstabe sagt, ob ein Zeichen eine eigene Stelle beansprucht.
//
// **Nicht-abstandhaltende Marken tun das nicht**: `e` gefolgt von U+0301 ist
// ein Buchstabe aus zwei Runen
// ([0016](../../docs/entscheidungen/0016-keine-unicode-segmentierung.md)).
func istBuchstabe(zeichen rune) bool {
	return !unicode.Is(unicode.Mn, zeichen) && !unicode.Is(unicode.Me, zeichen)
}

// buchstabenIn zählt die Buchstaben einer Zeile.
func buchstabenIn(zeile string) int {
	anzahl := 0

	for _, zeichen := range zeile {
		if istBuchstabe(zeichen) {
			anzahl++
		}
	}

	return anzahl
}

// breiteVon misst eine Zeile samt Laufweite.
//
// **Die Laufweite steht zwischen den Zeichen, nicht hinter dem letzten.** Sonst
// säße eine rechtsbündige Zeile um eine Laufweite zu weit links, und bei den
// 0,2 Punkten aus `magic-m15` fiele das erst im Vergleich zweier Karten auf.
func breiteVon(zeile string, satz Satz, schrift Schrift) float64 {
	// **Kein Sonderfall für die leere Zeile.** Sie hat keine Zeichen, also
	// keine Breite, und `max(0-1, 0)` ist null — ein `if` davor prüfte etwas,
	// das die Rechnung schon beantwortet.
	breite := 0.0
	vorher := rune(0)

	for _, einzelnes := range gesetzt(zeile, satz, schrift) {
		breite += kernungZwischen(vorher, einzelnes, satz, schrift)
		breite += schrift.Breite(einzelnes) * satz.Groesse
		vorher = einzelnes
	}

	// **Die Laufweite steht zwischen Buchstaben, nicht zwischen Runen.**
	//
	// `e` gefolgt von U+0301 ist **ein** Buchstabe aus zwei Runen; eine
	// Laufweite dazwischen setzte die Marke neben ihren Buchstaben statt über
	// ihn (#113).
	//
	// **`max` und nicht `-1`.** Eine Zeile aus lauter Marken hat null
	// Buchstaben, und null minus eins ergäbe eine negative Laufweite — also
	// eine Zeile, die schmaler ist als ihre Zeichen.
	return breite + float64(max(buchstabenIn(zeile)-1, 0))*satz.Laufweite
}

// waagerechterVersatz gibt, wie weit ein Zug in seinem Kasten nach rechts
// rückt — `align_delta_x` daneben, dieselbe Gestalt
// (`src/util/alignment.cpp`, Zeilen 14–20).
//
// Fassung f9d9356 vom 2021-07-06.
func waagerechterVersatz(ausrichtung Ausrichtung, rahmen, inhalt float64) float64 {
	switch ausrichtung {
	case MitteOben, MitteMitte, MitteUnten:
		return (rahmen - inhalt) / 2

	case RechtsOben, RechtsMitte, RechtsUnten:
		return rahmen - inhalt

	case LinksOben, LinksMitte, LinksUnten:
	}

	return 0
}

// senkrechterVersatz gibt, wie weit ein Zug in seinem Kasten nach unten rückt.
//
//	double align_delta_y(Alignment align, double box_height, double obj_height) {
//	  if      (align & ALIGN_MIDDLE) return (box_height - obj_height) / 2;
//	  else if (align & ALIGN_BOTTOM) return  box_height - obj_height;
//	  else                           return 0;
//	}
//
// `src/util/alignment.cpp`, Zeilen 22–26, gerufen aus `alignParagraph`
// (`src/render/text/viewer.cpp`, Zeile 821).
//
// **Ohne Klemmen.** Ist der Zug höher als der Kasten, wird der Versatz
// negativ, und er läuft nach **oben** hinaus — MSE kürzt ihn nicht, es meldet
// ihn (#392).
//
// Fassung f9d9356 vom 2021-07-06.
func senkrechterVersatz(ausrichtung Ausrichtung, rahmen, inhalt float64) float64 {
	switch ausrichtung {
	case LinksMitte, MitteMitte, RechtsMitte:
		return (rahmen - inhalt) / 2

	case LinksUnten, MitteUnten, RechtsUnten:
		return rahmen - inhalt

	case LinksOben, MitteOben, RechtsOben:
	}

	return 0
}
