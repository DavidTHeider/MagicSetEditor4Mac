package render

import "image"

// **Die Nähte, die nur Tests benutzen** — an einer Stelle.
//
// # Warum es sie gibt
//
// Der Rasterer beantwortet nach außen **eine** Frage: aus einem Plan wird ein
// Bild. Alles, was auf dem Weg dorthin entschieden wird — welche Auflösung ein
// Symbol bekommt, welche Größe ein Bildbaum gefragt hat —, ist am Ergebnis
// nicht abzulesen, weil das Bild danach ohnehin auf seinen Rahmen kommt.
//
// **Eine Zusage, die niemand prüfen kann, ist eine Behauptung** (§10). Diese
// drei Nähte machen sie prüfbar, und sie stehen hier zusammen, damit sichtbar
// bleibt, wie viele es sind.
//
// # Warum sie im Betrieb nichts kosten
//
// [Rasterer.AusBaumFuerTest] und [Rasterer.AusBaumMitMassFuerTest] sind
// Aufrufe, die niemand macht. [Rasterer.MeldeGroesseFuerTest] setzt ein Feld,
// das ungesetzt einen Nil-Vergleich je Bildbaum kostet.

// AusBaumFuerTest gibt [Rasterer.ausBaum] für den Wächter frei.
//
// **Weil die Planprüfung den Fall vorher abfängt.** Ein Baum ohne Bild kommt
// über `Zeichne` nie bei `ausBaum` an; der Zweig dort ist trotzdem richtig und
// muss geprüft werden — ein Weg, den kein Aufruf betritt, ist entweder tot oder
// eine Lücke (§3). Hier ist er die zweite Verteidigungslinie.
func (r *Rasterer) AusBaumFuerTest(baum Bildbaum) (quelle image.Image, befund string) {
	return r.ausBaum(baum, image.Point{})
}

// AusBaumMitMassFuerTest gibt [Rasterer.ausBaum] samt der gefragten Größe frei.
//
// **Weil die Größe eine Zusage ist und keine Rechnung.** Dass
// [Symbolflaeche] stimmt, sagt nichts darüber, ob `symbol_variation` sie
// fragt — bis zum 2026-08-21 tat es das nicht, und keine Prüfung wurde davon
// rot (#735).
//
// Im Betrieb kommt die Größe aus der Ebene (`bild.go`, Zeilen 55 und 77) oder
// aus [Rasterer.BildAusBaum]; beide gehen durch dieselbe Naht.
func (r *Rasterer) AusBaumMitMassFuerTest(
	baum Bildbaum, gefragt image.Point,
) (quelle image.Image, befund string) {
	return r.ausBaum(baum, gefragt)
}

// MeldeGroesseFuerTest lässt jeden Bildbaumlauf seine gefragte Größe melden.
//
// **Weil die Zusage sonst unprüfbar wäre** —
// #751: dass
// [Symbolflaeche] richtig rechnet, sagt nichts darüber, ob die fünf
// produktiven Wege sie überhaupt fragen. Bis zum 2026-08-21 prüfte das nur der
// Testeingang [Rasterer.AusBaumMitMassFuerTest] — also der Weg, den kein
// Betrieb nimmt.
//
// **Was sie nicht ist**: eine Auskunft an das Frontend. Sie steht in `render`,
// wird nur aus Tests gesetzt, und ohne sie kostet sie einen Nil-Vergleich je
// Bildbaum.
func (r *Rasterer) MeldeGroesseFuerTest(melde func(image.Point)) *Rasterer {
	r.gemeldeteGroesse = melde

	return r
}
