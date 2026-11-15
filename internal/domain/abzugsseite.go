// **Ein Abzug auf ein Blatt Papier** — PLAY-05.AK-03
// (#1713):
// „Screenshot, Bildexport und PDF besitzen Vorschau … mit Auflösung,
// **Seitenformat** und Zuschnitt."
//
// # Warum das hier steht
//
// **Weil es eine Rechnung ist und keine Datei.** Wohin ein Bild auf einem
// Blatt gehört, hängt an zwei Verhältnissen und sonst nichts — `internal/pdf`
// setzt ein Rechteck, und was ein Seitenverhältnis ist, weiß es ausdrücklich
// nicht (siehe den Kopf von [pdf.Seite.Bild]).
//
// # Und warum nicht neben dem Bogenleger
//
// **Weil der Bogen eine andere Frage beantwortet**: dort liegen *viele*
// Karten mit festem Maß in einem Raster, hier liegt *ein* Bild so groß wie
// möglich. Ein gemeinsamer Rechner hätte in jeder Zeile ein `if` über die
// Frage, welcher Fall gerade gilt.

package domain

// Blattkennung ist der Name eines Papierformats.
type Blattkennung = string

// Die zwei Formate, die zur Wahl stehen.
//
// **Zwei und nicht zehn.** A4 ist das Papier dieses Kontinents, Letter das
// des anderen; wer ein drittes braucht, trägt es hier ein und nicht im
// Fenster.
const (
	BlattA4     Blattkennung = "a4"
	BlattLetter Blattkennung = "letter"
)

// Blatt ist ein Papiermaß in Millimetern.
type Blatt struct {
	Breite float64
	Hoehe  float64
}

// blaetter führt die Maße.
//
// **A4 ist 210 × 297** (ISO 216), **Letter 215,9 × 279,4** — das sind 8,5 ×
// 11 Zoll, und der Zoll ist mit 25,4 mm definiert.
var blaetter = map[Blattkennung]Blatt{
	BlattA4:     {Breite: 210, Hoehe: 297},
	BlattLetter: {Breite: 215.9, Hoehe: 279.4},
}

// BlattZu gibt das Maß zu einer Kennung — der zweite Wert sagt, ob es sie
// gibt.
//
// **Ohne stillschweigende Vorgabe**: wer „A3" tippt, bekäme sonst A4 und
// merkte es am Ausdruck (§8).
func BlattZu(kennung Blattkennung) (Blatt, bool) {
	blatt, gibtEs := blaetter[kennung]

	return blatt, gibtEs
}

// BlattFuer dreht das Blatt nach dem Bild.
//
// **Ein breiter Tisch auf ein stehendes A4 gelegt füllte ein Drittel der
// Seite.** Die Drehung ist damit keine Wahl, sondern die Antwort auf das
// Seitenverhältnis — und wer sie doch treffen will, wählt einen anderen
// Ausschnitt.
//
// **Ein Bild ohne Fläche dreht nichts**: es hat kein Verhältnis.
func BlattFuer(bildbreite, bildhoehe int, blatt Blatt) Blatt {
	if bildbreite <= 0 || bildhoehe <= 0 {
		return blatt
	}

	if bildbreite > bildhoehe == (blatt.Breite > blatt.Hoehe) {
		return blatt
	}

	return Blatt{Breite: blatt.Hoehe, Hoehe: blatt.Breite}
}

// Blattplatz ist das Rechteck, in dem ein Bild auf der Seite steht — in
// Millimetern, von der unteren linken Ecke aus.
//
// **Von unten links**, weil PDF so rechnet; die Umrechnung in Punkte macht
// [pdf.MM].
type Blattplatz struct {
	Links  float64
	Unten  float64
	Breite float64
	Hoehe  float64
}

// PlatzAufSeite legt ein Bild so groß wie möglich mittig auf ein Blatt.
//
// **Verhältnistreu**: ein Abzug, der in die Breite gezogen wird, zeigt eine
// Karte, die es nicht gibt.
//
// **Ein Rechteck ohne Fläche heißt: kein Platz.** Der Rand kann größer sein
// als das Papier, und ein negatives Rechteck wäre ein PDF, das niemand
// öffnet.
func PlatzAufSeite(
	bildbreite, bildhoehe int, blatt Blatt, rand float64,
) Blattplatz {
	raumBreite := blatt.Breite - 2*rand
	raumHoehe := blatt.Hoehe - 2*rand

	if raumBreite <= 0 || raumHoehe <= 0 || bildbreite <= 0 || bildhoehe <= 0 {
		return Blattplatz{}
	}

	// **Der kleinere der beiden Maßstäbe passt.** Der größere ließe das Bild
	// über eine Kante hinauslaufen.
	nachBreite := raumBreite / float64(bildbreite)
	nachHoehe := raumHoehe / float64(bildhoehe)

	massstab := nachBreite
	if nachHoehe < massstab {
		massstab = nachHoehe
	}

	breite := float64(bildbreite) * massstab
	hoehe := float64(bildhoehe) * massstab

	return Blattplatz{
		Links:  (blatt.Breite - breite) / 2,
		Unten:  (blatt.Hoehe - hoehe) / 2,
		Breite: breite,
		Hoehe:  hoehe,
	}
}
