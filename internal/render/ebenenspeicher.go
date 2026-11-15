package render

import (
	"image"
	"sync"
	"sync/atomic"
)

// **Eine Ebene wird einmal auf ihre Größe gebracht, nicht je Karte**
// (#672).
//
// # Warum es das gibt
//
// Der Kartenrahmen, die Verzierung, der Stempel — sie sind für **jede** Karte
// desselben Stils dasselbe Bild in derselben Größe. Gemessen über 120
// Vorschauen war das Abtasten **26 %** der gesamten Rechenzeit, und dieselbe
// Karte vierzigmal zu zeichnen kostete genauso viel wie vierzig verschiedene.
//
// # Woran der Schlüssel hängt — und woran nicht
//
// [abtasten] rechnet aus `x-rahmen.Min.X` und `rahmen.Dx()`. **Die Lage des
// Rahmens geht nicht ein, nur seine Größe**: derselbe Rahmen zwei Punkte
// weiter rechts gibt dasselbe Bild.
//
// **Und der Zeiger auf die Quelle, nicht ihr Inhalt.** Die Bilder kommen aus
// `msepaket.Bildquelle`, die sie ihrerseits vorhält; derselbe Zeiger heißt
// dort dieselbe Datei. Ein Bild, das jemand während der Sitzung austauscht,
// ist ein anderer Zeiger — der Preis dafür steht dort und gilt hier mit.
//
// # Was **nicht** gemerkt wird
//
// Eine Quelle, die kein Zeiger ist. `ausschnitt` etwa ist ein Wert, und ein
// Wert als Schlüssel einer Abbildung ist entweder unvergleichbar (dann bricht
// es ab) oder er vergleicht seinen Inhalt (dann ist der Schlüssel nicht mehr
// die Identität). Solche Ebenen werden abgetastet wie bisher.
type ebenenschluessel struct {
	quelle image.Image
	breite int
	hoehe  int

	// guete gehört zum Schlüssel, weil dieselbe Quelle in derselben Größe
	// je Güte **anders** abgetastet wird. Ohne sie bekäme die zweite Güte das
	// Bild der ersten — und der Beleg, dass sich die Güten unterscheiden,
	// hinge davon ab, welche zuerst lief.
	guete Guete
}

// GrenzeEbenenspeicher ist der Platz, den die vorgehaltenen Ebenen einnehmen
// dürfen — **in Bytes und nicht in Stück**.
//
// # Warum nicht in Stück
//
// Ein Ebenenbild ist so groß wie sein Rahmen. In der Vorschau sind das
// 375 × 523 und damit 785 KB; auf dem Druckbogen rechnet `bogen.go` aus
// Papier und Auflösung eine Vergrößerung von rund **1,98**, und dasselbe Bild
// wiegt dann 3 MB. Eine Grenze von „32 Bildern" hieße einmal 25 MB und einmal
// 100 MB — dieselbe Zahl, viermal so viel Speicher.
//
// **64 MB**: das trägt die zwölf bis fünfzehn Ebenen einer Vorschau mehrfach
// und rund zwanzig in Druckgröße.
//
// **Sie räumt ganz, nicht klug** — dieselbe Wahl und derselbe Grund wie bei
// `msepaket.Bildquelle` und beim Baumspeicher: wer die Grenze erreicht, hat
// keinen Arbeitssatz mehr, der hineinpasst.
const GrenzeEbenenspeicher = 64 << 20

// **Mit Sperre**, seit #748.
//
// Seit #741
// laufen zwei Ausfuhren verschiedener Sets wirklich nebeneinander, und
// `internal/app` reicht **einen** Rasterer an fünf gleichzeitig erreichbare
// Wege. Zwei schreibende Zugriffe auf eine Go-`map` sind kein falsches Bild,
// sondern ein Abbruch.
//
// **Der Speicher bleibt geteilt, und das ist der Sinn**: derselbe Kartenrahmen
// in derselben Größe ist für jede Karte des Stils dasselbe Bild (#672). Was
// dazukommt, ist die Sperre — nicht ein Speicher je Lauf.
type ebenenspeicher struct {
	// treffer und fehlgriffe zählen die Abfragen — OPS-08 (#1725).
	treffer, fehlgriffe atomic.Int64

	sperre sync.Mutex
	bilder map[ebenenschluessel]*image.NRGBA
	platz  int

	// unterwegs hält je Schlüssel den Lauf, der ihn gerade rechnet.
	//
	// **Damit zwei gleichzeitige Läufe einmal rechnen** —
	// #1133.
	// Nachsehen und Eintragen nahmen die Sperre getrennt, und dazwischen lag
	// die Rechnung: beide verfehlten, beide rechneten, beide trugen ein.
	//
	// **Die Sperre wird über der Rechnung nicht gehalten.** Ein Rasterer ist
	// von fünf Wegen gleichzeitig erreichbar (#741, #748); eine Sperre über
	// dem Abtasten machte aus fünf Wegen einen. Gewartet wird deshalb auf
	// **diesen einen Schlüssel**, und Läufe mit anderen gehen vorbei.
	unterwegs map[ebenenschluessel]*ebenenlauf
}

// ebenenlauf ist eine Rechnung, auf die andere warten können.
type ebenenlauf struct {
	fertig chan struct{}
	bild   *image.NRGBA
}

// merkeEbenenbild legt eine abgetastete Ebene ab und räumt an der Grenze.
func (r *Rasterer) merkeEbenenbild(schluessel ebenenschluessel, bild *image.NRGBA) {
	r.ebenen.sperre.Lock()
	defer r.ebenen.sperre.Unlock()

	if r.ebenen.bilder == nil || r.ebenen.platz+len(bild.Pix) > GrenzeEbenenspeicher {
		r.ebenen.bilder = map[ebenenschluessel]*image.NRGBA{}
		r.ebenen.platz = 0
	}

	// **Ein Schlüssel zählt einmal.** Stünde er schon da, zählte ein zweiter
	// Eintrag dasselbe Bild noch einmal — die Grenze wäre zu früh erreicht,
	// und dann wird der **ganze** Speicher geräumt (#1133).
	if _, schon := r.ebenen.bilder[schluessel]; !schon {
		r.ebenen.platz += len(bild.Pix)
	}

	r.ebenen.bilder[schluessel] = bild
}

// ebenenschluesselVon baut den Schlüssel — und sagt, ob es einen gibt.
func ebenenschluesselVon(
	quelle image.Image, rahmen image.Rectangle, guete Guete,
) (schluessel ebenenschluessel, merkbar bool) {
	switch quelle.(type) {
	case *image.NRGBA, *image.RGBA, *image.Paletted, *image.Gray,
		*image.NRGBA64, *image.RGBA64, *image.YCbCr, *image.CMYK:
		return ebenenschluessel{
			quelle: quelle, breite: rahmen.Dx(), hoehe: rahmen.Dy(), guete: guete,
		}, true
	}

	return ebenenschluessel{}, false
}

// abgetasteteEbene bringt eine Quelle auf die Größe des Rahmens.
//
// **Einmal je Quelle und Größe.** Der erste Aufruf tastet ab wie bisher, jeder
// weitere liest nur noch.
//
// **Nur eine Quelle, die wiederkommt** (E06, #2278): `ausBildquelle` sagt,
// dass sie aus der Bildquelle stammt und beim nächsten Lauf derselbe Zeiger
// ist. Ein erzeugtes Bild — Schachbrett, Setsymbol, Verknüpfung — ist bei
// jedem Lauf ein neues; sein Eintrag würde nie getroffen, hielte die Quelle
// fest und füllte den Speicher, bis er ganz geräumt wird.
func (r *Rasterer) abgetasteteEbene(
	quelle image.Image, rahmen image.Rectangle, guete Guete, ausBildquelle bool,
) *image.NRGBA {
	schluessel, merkbar := ebenenschluesselVon(quelle, rahmen, guete)
	merkbar = merkbar && ausBildquelle

	// **Ohne Schlüssel gibt es nichts zu teilen.** Eine Quelle, die sich nicht
	// merken lässt, rechnet jeder Lauf für sich — wie bisher.
	if merkbar {
		if bild, warte := r.eintragOderLauf(schluessel); bild != nil || warte != nil {
			r.ebenen.treffer.Add(1)

			if warte != nil {
				<-warte.fertig

				return warte.bild
			}

			return bild
		}

		r.ebenen.fehlgriffe.Add(1)

		defer r.laufBeenden(schluessel)
	}

	bild := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			farbe := abtasten(quelle, rahmen, rahmen.Min.X+x, rahmen.Min.Y+y, guete)
			stelle := bild.PixOffset(x, y)
			bild.Pix[stelle] = farbe.R
			bild.Pix[stelle+1] = farbe.G
			bild.Pix[stelle+2] = farbe.B
			bild.Pix[stelle+3] = farbe.A
		}
	}

	if merkbar {
		r.merkeEbenenbild(schluessel, bild)

		r.ebenen.sperre.Lock()

		if lauf := r.ebenen.unterwegs[schluessel]; lauf != nil {
			lauf.bild = bild
		}

		r.ebenen.sperre.Unlock()
	}

	return bild
}

// eintragOderLauf gibt entweder das fertige Bild, oder den Lauf, auf den zu
// warten ist — oder beides `nil`, und dann rechnet der Aufrufer selbst.
//
// **Drei Ausgänge unter einer Sperre**, denn das ist der Punkt: nachsehen und
// sich eintragen müssen ein Schritt sein, sonst tun es zwei Läufe beide.
func (r *Rasterer) eintragOderLauf(schluessel ebenenschluessel) (
	*image.NRGBA, *ebenenlauf,
) {
	r.ebenen.sperre.Lock()
	defer r.ebenen.sperre.Unlock()

	if bild, da := r.ebenen.bilder[schluessel]; da {
		return bild, nil
	}

	if lauf, da := r.ebenen.unterwegs[schluessel]; da {
		return nil, lauf
	}

	if r.ebenen.unterwegs == nil {
		r.ebenen.unterwegs = map[ebenenschluessel]*ebenenlauf{}
	}

	r.ebenen.unterwegs[schluessel] = &ebenenlauf{fertig: make(chan struct{})}

	return nil, nil
}

// laufBeenden weckt die Wartenden und räumt den Eintrag.
//
// **Auch wenn nichts herauskam.** Bricht die Rechnung ab, warteten die
// anderen sonst für immer — und ein `nil` ist eine Antwort, die sie prüfen
// können.
func (r *Rasterer) laufBeenden(schluessel ebenenschluessel) {
	r.ebenen.sperre.Lock()
	lauf := r.ebenen.unterwegs[schluessel]
	delete(r.ebenen.unterwegs, schluessel)
	r.ebenen.sperre.Unlock()

	if lauf != nil {
		close(lauf.fertig)
	}
}

// farbeAusEbene liest einen Punkt aus einer abgetasteten Ebene.
func farbeAusEbene(bild *image.NRGBA, x, y int) Farbe {
	stelle := bild.PixOffset(x, y)
	punkt := bild.Pix[stelle : stelle+4 : stelle+4]

	return Farbe{R: punkt[0], G: punkt[1], B: punkt[2], A: punkt[3]}
}
