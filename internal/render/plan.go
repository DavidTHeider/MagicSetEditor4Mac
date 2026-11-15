// Package render rastert eine Karte aus einem fertigen Zeichenplan.
//
// **Es sieht kein Skript** (AGENTS.md §2). Wer eine Skriptauswertung hier
// findet, hat die fehlende Naht gefunden: die Auswertung gehört nach `service`,
// das Ergebnis ist der Plan.
//
// Der Grund ist prüftechnisch. Ein Renderer, der auswerten kann, lässt sich nur
// noch über ganze Vorlagen prüfen — und ein Goldbildvergleich, der bei jeder
// Skriptänderung anders ausfällt, sagt nichts mehr darüber, ob das Rastern
// stimmt.
package render

import "fmt"

// Plan beschreibt eine fertig berechnete Karte.
//
// # Warum diese Struktur zuerst entsteht
//
// Sie ist die Stelle, an der man einen Fehler verortet: sieht der Plan richtig
// aus und das Bild falsch, liegt es an `render`; sieht der Plan falsch aus, an
// `msescript` oder an der Vorlage. Ohne diese Trennung ist jeder Renderfehler
// eine Suche über drei Pakete.
//
// **Kein Skriptrest.** Ein Feld, dessen Breite `{ length(card.type_1)*17 }`
// heißt, kommt hier mit einer Zahl an. Steht in einem Plan noch eine
// geschweifte Klammer, ist die Naht undicht.
type Plan struct {
	// Groesse in Punkten der Vorlage, nicht in Bildpunkten. Die Auflösung
	// entscheidet erst der Rasterlauf — derselbe Plan trägt Bildschirm und
	// Druck (#34).
	Breite float64
	Hoehe  float64

	// DPI ist die Auflösung, in der die Vorlage ihre Punkte meint — `card
	// dpi` des Stils, bei MSE 150. Mit ihr wird aus der Punktzahl eine
	// Druckgröße; ohne sie (0) gibt es keine. Werkstück U5 von Phase 5.1
	// (#1416): die Vorschau zeigt die Karte auf Wunsch so groß, wie sie
	// gedruckt wäre.
	DPI int

	// Stil nennt das Paket, mit dem der Plan gebaut ist — leer beim
	// Behelfsplan (#2314). **Eine Herkunft und keine Anweisung**: der
	// Rasterer liest ihn nicht; er reist für das Formular mit, das seine
	// Felder je Paket ordnet.
	Stil string

	// Ebenen von hinten nach vorn. Die Reihenfolge ist die Zeichenreihenfolge;
	// `z index` aus der Vorlage ist beim Bauen des Plans schon aufgelöst.
	Ebenen []Ebene

	// Feldflaechen sagen, **wo ein Feld liegt** — ob es dort etwas zeichnet
	// oder nicht.
	//
	// # Warum nicht die Ebenen genügen
	//
	// Ein leeres Textfeld ergibt keine Ebene: es gibt nichts zu malen. Seinen
	// Kasten hat es trotzdem, und MSE entscheidet daran, ob man das Feld
	// bedienen kann:
	//
	//	bool ValueViewer::isVisible() const {
	//	  return getStyle()->visible
	//	    && bounding_box.width > 0 && bounding_box.height > 0 …
	//	}
	//
	// `src/render/value/viewer.cpp`, Zeilen 36–42, gefragt von `is_enabled`
	// (`src/gui/control/card_editor.cpp`, Zeilen 155–157), Fassung f9d9356 vom
	// 2021-07-06. Ein `ValueViewer` bleibt bestehen, ob sein Wert leer ist
	// oder nicht — und deshalb ist eine frisch angelegte Karte in MSE
	// bearbeitbar, in diesem Programm bis
	// #562
	// aber nicht.
	//
	// # Warum im Plan und nicht daneben
	//
	// Der Kasten entsteht beim Planbau aus Skripten, die die Karte lesen —
	// `left: { if card.card_symbol == "none" then 32 else 50 }` steht so in
	// `magic-m15`. Wer ihn außerhalb noch einmal rechnete, rechnete ein
	// zweites Mal dasselbe und liefe beim nächsten Unterschied auseinander.
	//
	// **Der Rasterer liest sie nicht.** Sie sind die Auskunft an die
	// Oberfläche; gezeichnet wird aus [Plan.Ebenen].
	Feldflaechen []Feldflaeche

	// Vorgaben sind die Feldwerte, die **gerechnet** wurden — `default:`,
	// `initial:` oder die erste `choice:`, für Felder, die die Karte nicht
	// selbst führt
	// (#563).
	//
	// **Warum sie hier stehen**: derselbe Grund wie bei [Plan.Feldflaechen] —
	// sie fallen beim Planbau an. 142 der 249 Felder von `magic.mse-game`
	// tragen ein `default:`, und alle 142 sind Skript über `card`; wer sie
	// außerhalb noch einmal rechnete, rechnete ein zweites Mal dasselbe.
	//
	// **Der Rasterer liest sie nicht** — wie die Flächen sind sie die Auskunft
	// an die Oberfläche. Damit sind es zwei nichtzeichnende Glieder an diesem
	// Plan; kommt ein drittes, gehören sie in einen eigenen Verbund neben ihn.
	//
	// **Ein leerer Wert steht nicht darin.** „Hier steht gerechnet nichts" ist
	// keine Herleitung, und die Oberfläche machte daraus einen Hinweis über
	// einem leeren Feld.
	Vorgaben []Vorgabe
}

// Vorgabe ist ein gerechneter Feldwert.
type Vorgabe struct {
	// Feld ist der Feldname **mit Leerzeichen** — `card symbol`, nicht
	// `card_symbol`. In der Skriptumgebung heißt er anders; die Oberfläche
	// fragt mit dem Namen aus dem Paket.
	Feld string

	Wert string
}

// Feldflaeche ist der Kasten eines Feldes auf der Karte.
type Feldflaeche struct {
	Feld   string
	Rahmen Rechteck
}

// Ebene ist ein Zeichenschritt.
//
// Ein Verbund mit einer Art und nicht eine Schnittstelle je Form: der Rasterer
// schaltet ohnehin über die Art, und eine Schnittstelle mit fünf Trägertypen
// verteilte dieselbe Fallunterscheidung auf fünf Dateien.
//
// Vor allem aber muss ein Plan **hinschreibbar** sein — in einem Test, von
// Hand, ohne Konstruktoren. Das ist die Zusage aus #29, und sie trägt die ganze
// Prüfbarkeit der Phase.
type Ebene struct {
	Art Ebenenart

	// Feld ist der Name des Kartenfeldes, aus dem diese Ebene entstand — leer,
	// wo keines dahintersteht.
	//
	// # Wofür
	//
	//  1. **Ein Klick in die Vorschau wählt das Feld darunter.** Der Plan
	//     kennt die Rechtecke; wem sie gehören, wusste er nicht
	//     (#246).
	//  2. **Ein Befund nennt sein Feld** statt nur seiner Nummer.
	//
	// **Leer ist ein Fall, kein Mangel**: der weiße Grund, den der Planbauer
	// selbst hinzufügt, gehört zu keinem Feld.
	Feld string

	// Rahmen ist die Fläche, auf die diese Ebene wirkt.
	Rahmen Rechteck

	// Farbe füllt (`ArtFlaeche`) oder zeichnet Text (`ArtText`).
	Farbe Farbe

	// Bild ist der **Bildbaum** dieser Ebene — eine Datei oder eine
	// Verknüpfung daraus ([Bildbaum]).
	//
	// Ein Blatt trägt einen **Paketpfad**, keinen Dateisystempfad —
	// `/paket.mse-include/x.png`. Aufgelöst wird er beim Rastern, nicht beim
	// Planen: ein fehlendes Bild ist ein Befund und kein Planungsfehler.
	Bild Bildbaum

	// Abschnitte und Satz gelten für `ArtText`.
	//
	// **Abschnitte und keine Zeichenkette**, seit #312: kursiv steht mitten im
	// Regeltext, und eine flache Zeichenkette hat keine Stelle, an der ein
	// Schnitt wechseln könnte. [Ebene.Klartext] gibt den Zug ohne seine
	// Schnitte — für alles, was nur misst.
	Abschnitte []Abschnitt
	Satz       Satz

	// Maske ist ein Paketpfad auf ein Graustufenbild.
	//
	// **Sie wirkt vor dem Verlauf, nicht danach.** Andersherum überdeckte der
	// Verlauf, was die Maske freigelegt hat — und das sieht auf einer einzelnen
	// Karte fast richtig aus.
	// Maske ist der **Bildbaum** der Maske dieser Ebene.
	//
	// **Bis zum 2026-08-16 ein Pfad**, während `Bild` schon ein Baum war. Ergab
	// ein `mask:`-Skript eine Verknüpfung, war die Wurzel pfadlos und die Maske
	// verschwand still — angestrichen vom Zweitblick über die Phasen B bis D.
	// Bei einer Typmigration bleibt ein gleichartiger Nebenkanal gern im alten
	// Skalartyp zurück.
	Maske Bildbaum

	// Verlauf ist gesetzt, wenn statt `Farbe` ein Übergang gezeichnet wird.
	Verlauf *Verlauf

	// Deckung von 0 (unsichtbar) bis 1 (voll).
	Deckung float64

	// Winkel ist `angle:` in Grad, gegen den Uhrzeigersinn.
	//
	// **MSE gibt jedem Feld einen eigenen gedrehten Zeichenkontext**:
	//
	//	Rotation ValueViewer::getRotation() const {
	//	  return Rotation(deg_to_rad(getStyle()->angle), bounding_box, 1.0, getStretch());
	//	}
	//
	// `src/render/value/viewer.cpp`, Zeilen 44–46, Fassung f9d9356 vom
	// 2021-07-06. `bounding_box` ist `RealRect(left, top, width, height)`
	// (`data/field.hpp:124`), und `Rotation::tr` rechnet daraus
	//
	//	(c·x + s·y + links, −s·x + c·y + oben).
	//
	// **Das Feld verlässt dabei seinen eigenen Kasten**, und das ist so
	// gewollt: `name 2` von `magic-m15-aftermath` steht bei links 342, breit
	// 181, hoch 21 mit `angle: 270` — gezeichnet wird die Spalte
	// x 321…342, y 295…476.
	//
	// **Gebaut sind die geraden Winkel.** `angle:` steht in 19 von 190
	// Stilpaketen von `M15-Magic-Pack` (75 Angaben), im vollen Pack in 59 von
	// 376 (213); schräge Winkel kommen dort viermal vor
	// (#486).
	Winkel float64

	// Mischart sagt, **wie** diese Ebene auf das Bisherige gelegt wird —
	// `multiply`, `shadow`, `glow` und die übrigen aus [mischarten].
	//
	// **Leer heißt „normal"**: das Bild ersetzt, was darunter liegt, mit
	// seinem Alphakanal. MSE unterscheidet genau dort:
	//
	//	if (combine <= COMBINE_NORMAL) {
	//	  dc.DrawBitmap(img, x, y);
	//	} else { … combine_image(source, img, combine); … }
	//
	// `src/gfx/combine_image.cpp`, Zeilen 142–157, Fassung f9d9356 vom
	// 2021-07-06.
	//
	// **Sie kommt aus dem Bildbaum** ([Bildbaum.Mischart]) — `set_combine`
	// setzt sie, und die Verknüpfungen darüber reichen sie weiter. Über das
	// volle Pack steht der Aufruf 698-mal in 34 Paketen, 667-mal mit `shadow`
	// (#470).
	Mischart string

	// Passung sagt, dass ein Bild **eingepasst** wird statt gestreckt — nil
	// heißt gestreckt, so wie MSEs `ASPECT_STRETCH` ([Bildpassung]).
	Passung *Bildpassung

	// Ecke ist der Radius der runden Ecken einer [ArtFlaeche], in
	// Vorlagenpunkten — MSEs `style().radius`.
	Ecke float64

	// Rand macht aus einer [ArtFlaeche] einen **Rahmen**: gemalt wird, was
	// außerhalb des inneren Rechtecks liegt. Nil heißt gefüllt.
	Rand *Randbreiten
}

// Bildpassung passt ein Bild in seinen Rahmen ein, ohne es zu verzerren.
//
//	} else if (options.preserve_aspect == ASPECT_FIT) {
//	  if (iw * options.height > ih * options.width) { // too much height requested
//	    w = options.width;  h = options.width * ih / iw;
//	  } else {
//	    w = options.height * iw / ih; h = options.height;
//	  }
//	}
//
// `conform_image`, `src/gfx/generated_image.cpp`, Zeilen 49–59, Fassung f9d9356
// vom 2021-07-06; ausgerichtet wird danach mit `align_in_rect`
// (`src/render/value/choice.cpp`, Zeile 83).
//
// **Das gilt für ein Wahlbild, nicht für ein Bildfeld.** Ein Bildfeld — das
// Kunstfenster — wird bei MSE mit `image.Rescale(w, h)` **gestreckt**
// (`src/render/value/image.cpp`, Zeile 38); ein Wahlbild bekommt
// `ASPECT_FIT`, sofern seine Ausrichtung nicht `stretch` sagt.
type Bildpassung struct {
	Ausrichtung Ausrichtung
}

// Randbreiten sind die vier Breiten eines Farbfeldrahmens.
//
//	bool clip = style().left_width < style().width  && style().right_width  < style().width &&
//	      style().top_width  < style().height && style().bottom_width < style().height;
//
// `ColorValueViewer::draw`, `src/render/value/color.cpp`, Zeilen 47–50, Fassung
// f9d9356 vom 2021-07-06. **Alle vier müssen kleiner sein als das Feld**;
// MSEs Vorgabe ist 100000, und dann wird gefüllt statt gerahmt.
type Randbreiten struct {
	Links, Oben, Rechts, Unten float64
}

// Rahmt sagt, ob diese Breiten einen Rahmen ergeben — MSEs `clip`.
//
// **Ausgeführt, weil der Planbauer sie prüft**: eine einzelne Breite ergibt
// keinen Rahmen, und das ist eine Zusage, die dort geprüft wird, wo sie
// eingelöst wird.
func (r *Randbreiten) Rahmt(breite, hoehe float64) bool {
	if r == nil {
		return false
	}

	return r.Links < breite && r.Rechts < breite && r.Oben < hoehe && r.Unten < hoehe
}

// Ebenenart benennt, was eine Ebene zeichnet.
type Ebenenart string

const (
	ArtFlaeche Ebenenart = "Fläche"
	ArtBild    Ebenenart = "Bild"
	ArtText    Ebenenart = "Text"
	ArtSymbol  Ebenenart = "Symbol"

	// ArtSymbolzug ist ein **ganzes Feld aus Symbolen** — die Manakosten.
	//
	// Anders als [ArtSymbol] trägt sie kein ausgerechnetes Rechteck, sondern
	// die Symbole und die Maße ihrer Schrift; gemessen und angeordnet wird beim
	// Rastern (`docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md`).
	ArtSymbolzug Ebenenart = "Symbolzug"
)

// Rechteck ist eine Fläche in Vorlagenpunkten, Ursprung oben links.
type Rechteck struct {
	Links, Oben, Breite, Hoehe float64
}

// Farbe ist ein Farbwert mit Deckung, wie ihn `rgb`/`rgba` ergeben.
//
// **Eigen und nicht `msescript.Farbe`.** `render` importiert `msescript` nicht
// und soll es nicht: die Naht zwischen beiden ist der Plan, und ein gemeinsamer
// Typ machte aus der Naht eine Verbindung. Was gleich aussieht, ist hier
// absichtlich getrennt.
type Farbe struct {
	R, G, B, A uint8
}

// Ausrichtung sagt, wo der Text im Rahmen sitzt.
type Ausrichtung string

const (
	LinksOben   Ausrichtung = "links oben"
	MitteOben   Ausrichtung = "mitte oben"
	RechtsOben  Ausrichtung = "rechts oben"
	LinksMitte  Ausrichtung = "links mitte"
	MitteMitte  Ausrichtung = "mitte mitte"
	RechtsMitte Ausrichtung = "rechts mitte"
	LinksUnten  Ausrichtung = "links unten"
	MitteUnten  Ausrichtung = "mitte unten"
	RechtsUnten Ausrichtung = "rechts unten"
)

// Verlauf ist ein Übergang zwischen zwei oder mehr Farben.
type Verlauf struct {
	// Winkel in Grad, 0 heißt von links nach rechts.
	Winkel float64

	// Stufen von 0 bis 1, aufsteigend. Weniger als zwei ergeben keinen Verlauf.
	Stufen []Verlaufsstufe
}

// Verlaufsstufe ist eine Farbe an einer Stelle des Übergangs.
type Verlaufsstufe struct {
	Stelle float64
	Farbe  Farbe
}

func (f Farbe) String() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", f.R, f.G, f.B, f.A)
}

func (r Rechteck) String() string {
	return fmt.Sprintf("(%g,%g %gx%g)", r.Links, r.Oben, r.Breite, r.Hoehe)
}
