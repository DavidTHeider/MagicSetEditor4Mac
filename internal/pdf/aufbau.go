package pdf

import (
	"bytes"
	"fmt"
)

// aufbau sammelt die Objekte eines Dokuments und schreibt sie mit ihrer Tabelle.
//
// # Warum die Objekte gesammelt werden statt gleich zu fließen
//
// Ein PDF endet mit einer **Querverweistabelle**: für jedes Objekt die Stelle,
// an der es beginnt, in Bytes vom Dateianfang. Wer beim Schreiben mitzählt,
// zählt bei jeder Änderung neu; wer die Objekte erst sammelt, schreibt sie
// einmal und weiß die Stellen dabei.
type aufbau struct {
	objekte [][]byte
}

// reserviere legt leere Plätze an, deren Nummern schon vergeben sind.
func (a *aufbau) reserviere(anzahl int) {
	for range anzahl {
		a.objekte = append(a.objekte, nil)
	}
}

// nimm hängt ein Objekt an und gibt seine Nummer.
func (a *aufbau) nimm(inhalt []byte) int {
	a.objekte = append(a.objekte, inhalt)

	return len(a.objekte)
}

// setze füllt einen reservierten Platz.
func (a *aufbau) setze(nummer int, inhalt []byte) { a.objekte[nummer-1] = inhalt }

// seite schreibt eine Seite samt ihren Bildern und gibt ihre Nummer.
func (a *aufbau) seite(seite *Seite) int {
	quellen := &bytes.Buffer{}

	quellen.WriteString("<< /XObject << ")

	for _, bild := range seite.bilder {
		nummer := a.nimm(bildobjekt(bild))
		fmt.Fprintf(quellen, "/%s %d 0 R ", bild.name, nummer)
	}

	quellen.WriteString(">> >>")

	inhalt := a.nimm(strom(nil, seite.inhalt.Bytes()))

	return a.nimm([]byte(fmt.Sprintf(
		"<< /Type /Page\n /Parent 2 0 R /MediaBox [0 0 %s %s] /Contents %d 0 R /Resources %s >>",
		zahl(seite.breite), zahl(seite.hoehe), inhalt, quellen.String())))
}

// bildobjekt baut das XObject eines Bilds.
func bildobjekt(bild *bildeintrag) []byte {
	kopf := fmt.Sprintf(
		"/Type /XObject /Subtype /Image /Width %d /Height %d"+
			" /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode",
		bild.breite, bild.hoehe)

	return strom([]byte(kopf), bild.daten)
}

// strom baut ein Objekt mit Datenstrom; `kopf` sind zusätzliche Einträge.
func strom(kopf, daten []byte) []byte {
	var gebaut bytes.Buffer

	gebaut.WriteString("<< ")
	gebaut.Write(kopf)
	fmt.Fprintf(&gebaut, " /Length %d >>\nstream\n", len(daten))
	gebaut.Write(daten)
	gebaut.WriteString("\nendstream")

	return gebaut.Bytes()
}

// baue setzt Kopf, Objekte, Tabelle und Nachspann zusammen.
func (a *aufbau) baue() []byte {
	var gebaut bytes.Buffer

	// **Die zweite Zeile trägt hohe Bytes**, und das ist keine Zierde: sie sagt
	// jedem Werkzeug auf dem Weg, dass diese Datei binär ist. Ein Programm, das
	// sie für Text hält, ersetzt Zeilenenden — und danach stimmt jede Stelle in
	// der Tabelle nicht mehr.
	gebaut.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")

	stellen := make([]int, 0, len(a.objekte))

	for nummer, objekt := range a.objekte {
		stellen = append(stellen, gebaut.Len())
		fmt.Fprintf(&gebaut, "%d 0 obj\n", nummer+1)
		gebaut.Write(objekt)
		gebaut.WriteString("\nendobj\n")
	}

	tabelle := gebaut.Len()

	fmt.Fprintf(&gebaut, "xref\n0 %d\n0000000000 65535 f \n", len(a.objekte)+1)

	for _, stelle := range stellen {
		// Zehn Stellen, fünf Stellen, ein Leerzeichen: das Format ist auf das
		// Byte festgelegt, weil ein Leser die Einträge über ihre Länge findet.
		fmt.Fprintf(&gebaut, "%010d 00000 n \n", stelle)
	}

	fmt.Fprintf(&gebaut, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(a.objekte)+1, tabelle)

	return gebaut.Bytes()
}
