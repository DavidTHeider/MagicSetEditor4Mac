// Package pdf schreibt PDF-Dateien aus Bildern und Linien.
//
// # Warum selbst geschrieben
//
// Ein Druckbogen braucht drei Dinge: eine Seite mit einer Größe, Bilder an
// genauen Stellen und dünne Linien für die Schnittmarken. Kein Text, keine
// Schriften, keine Formulare, keine Verschlüsselung — und genau das ist der
// Teil von PDF, der klein ist.
//
// Eine Bibliothek dafür brächte den ganzen Rest mit, samt seiner Angriffsfläche
// und seiner Pflege, für einen Nutzen, den dreihundert Zeilen hier ebenso
// liefern ([0013](../../docs/entscheidungen/0013-eigener-pdf-schreiber.md)).
//
// **Was das kostet, steht im Vermerk**: dieses Paket kann genau eine Sorte PDF.
// Wer Text, Ebenen oder Farbprofile im PDF braucht, baut das hier ein oder holt
// eine Bibliothek — beides mit offenen Augen.
package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"io"
	"math"
	"strconv"
)

// MM rechnet Millimeter in Punkte um.
//
// **PDF misst in Punkten zu 1/72 Zoll**, und niemand denkt in ihnen: eine Karte
// ist 63 mm breit, ein Rand ist 10 mm. Die Umrechnung steht deshalb hier und
// nicht in jedem Aufrufer.
func MM(millimeter float64) float64 { return millimeter * 72.0 / 25.4 }

// Dokument ist ein PDF im Entstehen.
type Dokument struct {
	seiten []*Seite
}

// NeuesDokument beginnt ein PDF ohne Seiten.
func NeuesDokument() *Dokument { return &Dokument{} }

// Seite ist eine Seite samt ihrem Inhalt.
//
// **Der Ursprung liegt unten links**, wie PDF es vorgibt. Wer von oben rechnet,
// rechnet um — und zwar dort, wo er weiß, wie hoch die Seite ist.
type Seite struct {
	breite, hoehe float64

	inhalt bytes.Buffer
	bilder []*bildeintrag
}

// bildeintrag ist ein Bild samt seiner Nummer auf der Seite.
type bildeintrag struct {
	name  string
	daten []byte

	breite, hoehe int
}

// NeueSeite hängt eine Seite an; das Maß steht in Punkten (siehe `MM`).
func (d *Dokument) NeueSeite(breite, hoehe float64) *Seite {
	seite := &Seite{breite: breite, hoehe: hoehe}
	d.seiten = append(d.seiten, seite)

	return seite
}

// Bild setzt ein Bild in ein Rechteck; `links` und `unten` sind seine Ecke.
//
// **Das Bild wird in das Rechteck gelegt, nicht daran gemessen.** Wer ein
// Seitenverhältnis wahren will, gibt ein passendes Rechteck an — diese Ebene
// weiß nicht, was eine Karte ist.
func (s *Seite) Bild(bild image.Image, links, unten, breite, hoehe float64) {
	eintrag := &bildeintrag{
		name:   "Bild" + strconv.Itoa(len(s.bilder)+1),
		daten:  gepackt(rgbAus(bild)),
		breite: bild.Bounds().Dx(),
		hoehe:  bild.Bounds().Dy(),
	}

	s.bilder = append(s.bilder, eintrag)

	// `cm` setzt die Abbildung: ein Bild ist im PDF immer eine Einheitsfläche,
	// die über die Matrix auf ihre Größe und Stelle gebracht wird.
	fmt.Fprintf(&s.inhalt, "q\n%s 0 0 %s %s %s cm\n/%s Do\nQ\n",
		zahl(breite), zahl(hoehe), zahl(links), zahl(unten), eintrag.name)
}

// Linie zieht eine gerade Linie in Grau — 0 ist schwarz, 1 ist weiß.
func (s *Seite) Linie(x1, y1, x2, y2, staerke, grau float64) {
	fmt.Fprintf(&s.inhalt, "q\n%s w %s G\n%s %s m %s %s l\nS\nQ\n",
		zahl(staerke), zahl(grau), zahl(x1), zahl(y1), zahl(x2), zahl(y2))
}

// zahl schreibt eine Länge so, wie ein PDF sie erwartet.
//
// **Ohne Exponenten und mit Punkt.** `strconv.FormatFloat` mit 'f' gibt beides;
// ein `1e-05` aus `%v` wäre in einem Inhaltsstrom ein Syntaxfehler, den kein
// Leser meldet — er hört einfach auf zu zeichnen.
func zahl(wert float64) string {
	// **Eine Unzahl wird zur Null.**
	//
	// `strconv.FormatFloat` gibt für `math.Inf(1)` die Zeichenkette `+Inf` und
	// für `math.NaN()` das Wort `NaN` — beides ist in einem Inhaltsstrom ein
	// Syntaxfehler, und zwar einer, der **die ganze Seite** unlesbar macht:
	// ein PDF-Leser bricht dort ab, wo er ihn findet.
	//
	// **Woher eine Unzahl kommt**: `card width:` steht in einem fremden
	// Stilpaket, und der Bogenleger rechnet daraus Punkte. Eine Breite von
	// `1e309` ist ein gültiger Eintrag in einer `.mse-style` — was daraus
	// wird, entscheidet sich hier.
	//
	// **Warum Null und nicht ein Ersatzmaß**: eine Karte ohne Ausdehnung
	// zeichnet nichts, und nichts ist die ehrliche Antwort auf ein Maß, das
	// keines ist. Ein erfundener Ersatz sähe wie eine Karte aus.
	//
	// Gefunden vom ersten Fuzzlauf über den Schreiber am 2026-08-29
	// (#1289).
	if math.IsNaN(wert) || math.IsInf(wert, 0) {
		return "0.0000"
	}

	return strconv.FormatFloat(wert, 'f', 4, 64)
}

// rgbAus macht aus einem Bild die rohen RGB-Bytes, zeilenweise von oben.
//
// **Ohne Alphakanal**, und das ist Absicht: ein Druckbogen ist Papier. Wer
// Durchsichtigkeit mitgeben will, legt sie vorher auf einen Grund — sonst
// bräuchte jedes Bild eine zweite Maske im PDF, und das ist eine Sorte
// Dokument, die dieses Paket nicht bauen soll.
func rgbAus(bild image.Image) []byte {
	rahmen := bild.Bounds()
	roh := make([]byte, 0, rahmen.Dx()*rahmen.Dy()*3)

	for y := rahmen.Min.Y; y < rahmen.Max.Y; y++ {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			r, g, b, _ := bild.At(x, y).RGBA()
			roh = append(roh, achtBit(r), achtBit(g), achtBit(b))
		}
	}

	return roh
}

// achtBit macht aus einem 16-Bit-Kanal das obere Byte.
//
// **`color.Color.RGBA` gibt 16 Bit je Kanal**, ein PDF nimmt acht. Die
// Verschiebung um acht Stellen **ist** die Umrechnung: was übrig bleibt, passt
// in ein Byte. Ein Überlauf ist hier keiner, sondern die Rechnung selbst.
func achtBit(kanal uint32) byte {
	return byte(kanal >> 8) // #nosec G115 -- siehe darüber
}

// gepackt legt Bytes in einen Zlib-Strom — das `FlateDecode` des PDFs.
//
// **Verlustfrei und nicht als JPG.** Ein Druckbogen wird gedruckt; eine
// Kartenvorlage besteht aus harten Kanten, und die sind es, an denen JPG
// sichtbar wird. Der Preis ist die Dateigröße, und den zahlt ein Druck gern.
func gepackt(roh []byte) []byte {
	var gebaut bytes.Buffer

	packer := zlib.NewWriter(&gebaut)

	// **Ohne Fehlerweg**: ein `bytes.Buffer` nimmt jeden Schreibvorgang an, und
	// `zlib` gibt nur weiter, was der Schreiber meldet. Ein Zweig, den kein
	// Aufruf betreten kann, wäre eine Behauptung darüber, was im Ernstfall
	// geschieht (AGENTS.md §3).
	_, _ = packer.Write(roh) //nolint:errcheck // siehe darüber
	_ = packer.Close()       //nolint:errcheck // dito

	return gebaut.Bytes()
}

// Schreibe gibt das Dokument aus.
func (d *Dokument) Schreibe(ziel io.Writer) error {
	if _, err := ziel.Write(d.Bytes()); err != nil {
		return fmt.Errorf("das PDF schreiben: %w", err)
	}

	return nil
}

// Bytes baut das Dokument und gibt es als Bytes.
//
// **Getrennt von `Schreibe`, weil ein Aufrufer beides braucht.** Wer in eine
// Datei schreibt, will erst das ganze Dokument und dann einen Schreibvorgang —
// ein PDF, das zur Hälfte dasteht, sieht im Dateibrowser aus wie ein ganzes.
func (d *Dokument) Bytes() []byte {
	gebaut := &aufbau{}

	// 1 ist der Katalog, 2 der Seitenbaum — beide stehen fest, damit die Seiten
	// auf 2 verweisen können, bevor es geschrieben ist.
	gebaut.reserviere(2)

	nummern := make([]int, 0, len(d.seiten))
	for _, seite := range d.seiten {
		nummern = append(nummern, gebaut.seite(seite))
	}

	gebaut.setze(1, []byte("<< /Type /Catalog /Pages 2 0 R >>"))
	gebaut.setze(2, seitenbaum(nummern))

	return gebaut.baue()
}

// seitenbaum baut das Objekt, das alle Seiten führt.
func seitenbaum(nummern []int) []byte {
	var gebaut bytes.Buffer

	gebaut.WriteString("<< /Type /Pages /Kids [")

	for stelle, nummer := range nummern {
		if stelle > 0 {
			gebaut.WriteString(" ")
		}

		fmt.Fprintf(&gebaut, "%d 0 R", nummer)
	}

	fmt.Fprintf(&gebaut, "] /Count %d >>", len(nummern))

	return gebaut.Bytes()
}
