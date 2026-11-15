package service_test

// **Ein Druckbogen, vermessen wie von einem fremden Werkzeug** — I04
// (#2262).
//
// `bogen_test.go` liest die Bildgröße mit einem regulären Ausdruck aus dem
// rohen Inhalt — richtig für eine Zahl, blind für alles daneben: welche Seite,
// welche Matrix davor, welches Bild. Hier wird das PDF **gelesen, wie ein
// Leser es liest**: Querverweistabelle, Seiten, ihre `MediaBox`, die Matrix
// jedes Bildes über `q`/`Q`/`cm` und die Linien der Schnittmarken. Alles in
// Millimetern.
//
// # Was er nicht ist
//
// **Kein allgemeiner PDF-Leser.** Er versteht, was `internal/pdf` schreibt —
// ungepackte Inhaltsströme, Bilder als XObject, gerade Linien. Ein gepackter
// Inhaltsstrom oder eine Pfadoperation, die er nicht kennt, hält ihn mit einem
// Fehler an, statt sie zu übergehen.

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// punktInMM rechnet einen PDF-Punkt in Millimeter.
const punktInMM = 25.4 / 72

// vermessenesBild ist ein gesetztes Bild in Millimetern samt seinen Bildpunkten.
type vermessenesBild struct {
	links, unten, breite, hoehe float64
	punkteBreit, punkteHoch     int
}

// vermesseneLinie ist eine gezogene Linie in Millimetern.
type vermesseneLinie struct{ x1, y1, x2, y2 float64 }

// vermesseneSeite ist eine Seite in Millimetern.
type vermesseneSeite struct {
	breite, hoehe float64
	bilder        []vermessenesBild
	linien        []vermesseneLinie
}

var (
	objektkopf  = regexp.MustCompile(`^(\d+) 0 obj\n`)
	xrefzeile   = regexp.MustCompile(`(?m)^(\d{10}) 0{5} n $`)
	startxref   = regexp.MustCompile(`startxref\n(\d+)\n%%EOF\n$`)
	mediabox    = regexp.MustCompile(`/MediaBox \[0 0 ([0-9.]+) ([0-9.]+)\]`)
	inhaltsref  = regexp.MustCompile(`/Contents (\d+) 0 R`)
	xobjektref  = regexp.MustCompile(`/(\w+) (\d+) 0 R`)
	bildmass    = regexp.MustCompile(`/Width (\d+) /Height (\d+)`)
	stromgrenze = regexp.MustCompile(`(?s)stream\n(.*)\nendstream`)
)

// vermiss liest einen Bogen. Jede Stelle der Querverweistabelle muss auf ihr
// Objekt zeigen — eine verrutschte Tabelle ist ein Bogen, den ein Leser
// verwirft, und das ist ein Befund wie jeder andere.
func vermiss(roh []byte) ([]vermesseneSeite, error) {
	ende := startxref.FindSubmatch(roh)
	if ende == nil {
		return nil, fmt.Errorf("kein startxref am Ende")
	}

	tabelle, _ := strconv.Atoi(string(ende[1])) //nolint:errcheck // das Muster lässt nur Ziffern zu
	if tabelle >= len(roh) || !bytes.HasPrefix(roh[tabelle:], []byte("xref\n")) {
		return nil, fmt.Errorf("startxref %d zeigt nicht auf die Tabelle", tabelle)
	}

	objekte := map[int]string{}

	for nummer, stelle := range xrefzeile.FindAllSubmatch(roh[tabelle:], -1) {
		ab, _ := strconv.Atoi(string(stelle[1])) //nolint:errcheck // dito

		kopf := objektkopf.FindSubmatch(roh[ab:])
		if len(kopf) < 2 || string(kopf[1]) != strconv.Itoa(nummer+1) {
			return nil, fmt.Errorf("die Tabelle zeigt für Objekt %d auf Byte %d, dort steht keines", nummer+1, ab)
		}

		rumpf := roh[ab+len(kopf[0]):]

		schluss := bytes.Index(rumpf, []byte("\nendobj\n"))
		if schluss < 0 {
			return nil, fmt.Errorf("Objekt %d ohne endobj", nummer+1)
		}

		objekte[nummer+1] = string(rumpf[:schluss])
	}

	var seiten []vermesseneSeite

	for nummer := 1; nummer <= len(objekte); nummer++ {
		if !strings.HasPrefix(objekte[nummer], "<< /Type /Page\n") {
			continue
		}

		seite, fehler := seiteVermessen(objekte, objekte[nummer])
		if fehler != nil {
			return nil, fmt.Errorf("Seite in Objekt %d: %w", nummer, fehler)
		}

		seiten = append(seiten, seite)
	}

	return seiten, nil
}

// seiteVermessen liest MediaBox, Bilder und Linien einer Seite.
func seiteVermessen(objekte map[int]string, seite string) (vermesseneSeite, error) {
	box := mediabox.FindStringSubmatch(seite)
	inhalt := inhaltsref.FindStringSubmatch(seite)

	if box == nil || inhalt == nil {
		return vermesseneSeite{}, fmt.Errorf("ohne MediaBox oder Inhalt")
	}

	vermessen := vermesseneSeite{breite: mm(box[1]), hoehe: mm(box[2])}

	bilder := map[string]string{}

	for _, eintrag := range xobjektref.FindAllStringSubmatch(seite, -1) {
		if eintrag[1] != "Contents" && eintrag[1] != "Parent" {
			nummer, _ := strconv.Atoi(eintrag[2]) //nolint:errcheck // das Muster lässt nur Ziffern zu
			bilder[eintrag[1]] = objekte[nummer]
		}
	}

	nummer, _ := strconv.Atoi(inhalt[1]) //nolint:errcheck // dito

	strom := stromgrenze.FindStringSubmatch(objekte[nummer])
	if strom == nil {
		return vermesseneSeite{}, fmt.Errorf("Inhalt ohne Strom")
	}

	return inhaltLesen(vermessen, strom[1], bilder)
}

// inhaltLesen geht einen Inhaltsstrom Wort für Wort durch.
//
// **Mit dem Stapel der Grafikzustände**: eine Matrix gilt bis zum `Q`, das ihr
// `q` schließt. Ein Leser, der `cm` ohne Stapel verkettete, sähe die zweite
// Karte an der Summe beider Stellen.
func inhaltLesen(seite vermesseneSeite, strom string, bilder map[string]string) (vermesseneSeite, error) {
	matrix := [6]float64{1, 0, 0, 1, 0, 0}

	var (
		stapel [][6]float64
		zahlen []float64
		name   string
		anfang [2]float64
	)

	for _, wort := range strings.Fields(strom) {
		if zahl, fehler := strconv.ParseFloat(wort, 64); fehler == nil {
			zahlen = append(zahlen, zahl)

			continue
		}

		if strings.HasPrefix(wort, "/") {
			name = wort[1:]

			continue
		}

		switch wort {
		case "q":
			stapel = append(stapel, matrix)
		case "Q":
			if len(stapel) == 0 {
				return seite, fmt.Errorf("ein Q ohne q")
			}

			matrix, stapel = stapel[len(stapel)-1], stapel[:len(stapel)-1]
		case "cm":
			if len(zahlen) != 6 {
				return seite, fmt.Errorf("cm mit %d Zahlen", len(zahlen))
			}

			matrix = verkette([6]float64(zahlen), matrix)
		case "m":
			anfang = [2]float64{zahlen[0], zahlen[1]}
		case "l":
			x1, y1 := abbilden(matrix, anfang[0], anfang[1])
			x2, y2 := abbilden(matrix, zahlen[0], zahlen[1])
			seite.linien = append(seite.linien, vermesseneLinie{
				x1 * punktInMM, y1 * punktInMM, x2 * punktInMM, y2 * punktInMM,
			})
		case "Do":
			bild, fehler := bildLesen(matrix, bilder[name])
			if fehler != nil {
				return seite, fmt.Errorf("Bild %s: %w", name, fehler)
			}

			seite.bilder = append(seite.bilder, bild)
		case "w", "G", "S":
		default:
			return seite, fmt.Errorf("der Operator %q ist hier unbekannt", wort)
		}

		zahlen = zahlen[:0]
	}

	if len(stapel) != 0 {
		return seite, fmt.Errorf("%d q ohne Q", len(stapel))
	}

	return seite, nil
}

// bildLesen bildet das Einheitsquadrat eines Bildes über die Matrix ab.
//
// **Ein gedrehtes oder geschertes Bild hält an**: der Bogen setzt keines, und
// eine Breite aus einer gedrehten Matrix wäre eine Zahl, die etwas anderes
// misst, als sie behauptet.
func bildLesen(matrix [6]float64, objekt string) (vermessenesBild, error) {
	mass := bildmass.FindStringSubmatch(objekt)
	if mass == nil {
		return vermessenesBild{}, fmt.Errorf("kein Bild mit Breite und Höhe")
	}

	if matrix[1] != 0 || matrix[2] != 0 {
		return vermessenesBild{}, fmt.Errorf("gedreht oder geschert: %v", matrix)
	}

	links, unten := abbilden(matrix, 0, 0)
	rechts, oben := abbilden(matrix, 1, 1)
	breit, _ := strconv.Atoi(mass[1]) //nolint:errcheck // das Muster lässt nur Ziffern zu
	hoch, _ := strconv.Atoi(mass[2])  //nolint:errcheck // dito

	return vermessenesBild{
		links: links * punktInMM, unten: unten * punktInMM,
		breite: (rechts - links) * punktInMM, hoehe: (oben - unten) * punktInMM,
		punkteBreit: breit, punkteHoch: hoch,
	}, nil
}

// verkette setzt `m` vor die laufende Matrix — die Reihenfolge, in der ein
// PDF-Leser `cm` anwendet.
func verkette(m, laufend [6]float64) [6]float64 {
	return [6]float64{
		m[0]*laufend[0] + m[1]*laufend[2],
		m[0]*laufend[1] + m[1]*laufend[3],
		m[2]*laufend[0] + m[3]*laufend[2],
		m[2]*laufend[1] + m[3]*laufend[3],
		m[4]*laufend[0] + m[5]*laufend[2] + laufend[4],
		m[4]*laufend[1] + m[5]*laufend[3] + laufend[5],
	}
}

// abbilden bringt einen Punkt über eine Matrix auf die Seite.
func abbilden(m [6]float64, x, y float64) (aufX, aufY float64) {
	return m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]
}

// mm liest eine Zahl in Punkten und gibt sie in Millimetern.
func mm(punkte string) float64 {
	zahl, _ := strconv.ParseFloat(punkte, 64) //nolint:errcheck // das Muster lässt nur Zahlen zu

	return zahl * punktInMM
}
