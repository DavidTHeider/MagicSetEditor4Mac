// **Was ein Betrag ist** — LOC-03.AK-02 und AK-03
// (#1689):
// „Geld erhält einen eindeutigen Wert-/Währungs-/Skalenvertrag" und
// „mehrdeutige Eingaben werden erklärt und nicht still falsch
// interpretiert."
//
// # Der Vertrag in einem Satz
//
// **Ein Betrag ist eine ganze Zahl in der kleinsten Einheit seiner
// Währung** — Cent beim Euro, Yen beim Yen, Fils beim Dinar. Wie viele
// Stellen das sind, sagt die Währung und nicht der Aufrufer.
//
// # Was vorher galt und woran es scheiterte
//
// **Vorher waren es „Hundertstel der Währung".** Das ist für den Euro
// richtig und für zwei Gruppen falsch, und beides ist gemessen (2026-09-22):
//
//	 15.050 Hundertstel JPY  → »151 ¥«    — ¥150,50 gibt es nicht, und
//	                                        gerundet wurde stillschweigend
//	      1 Hundertstel BHD  → »0,010 BHD« — ein Fils lässt sich gar nicht
//	                                        ausdrücken
//
// **Und das Lesen war genauso stumm**: `strconv.ParseFloat` gibt für
// »1,23« einen Fehler, und der Aufrufer machte daraus eine Null. Wer eine
// deutsch formatierte Liste einlas, verlor **jeden** Preis, ohne ein Wort
// zu lesen.
//
// # Warum die kanonische Form einen Punkt hat und kein Komma
//
// **Weil sie keine Sprache ist.** Was hier hineingeht, ist die Form, in der
// gespeichert und über die Flächen geredet wird (AK-03); wie eine Zahl
// **aussieht**, entscheidet die Sprache des Fensters und nicht der Bestand.
// Ein Komma in der kanonischen Form hieße, dass dieselbe Datei in zwei
// Ländern zwei Beträge bedeutet.

package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// nullstellige sind die Währungen ohne Nachkommastellen — ISO 4217.
//
// **Aufgezählt und nicht gerechnet**: es gibt keine Regel, aus der sich das
// herleiten ließe. Die Liste ist die der laufenden Währungen mit
// `minorUnits = 0`.
var nullstellige = map[string]bool{
	"BIF": true, "CLP": true, "DJF": true, "GNF": true, "ISK": true,
	"JPY": true, "KMF": true, "KRW": true, "PYG": true, "RWF": true,
	"UGX": true, "UYI": true, "VND": true, "VUV": true, "XAF": true,
	"XOF": true, "XPF": true,
}

// dreistellige sind die Währungen mit drei Nachkommastellen — ISO 4217.
//
// **Mit zwei benannten Abweichungen**: CLF und UYW haben dort vier Stellen
// und hier zwei, weil die Stellenzahl die Skala gespeicherter Beträge ist und
// ein Umskalieren unterschriebene Handelsrevisionen bräche
// ([0167](../../docs/entscheidungen/0167-clf-und-uyw-bleiben-zweistellig.md)).
var dreistellige = map[string]bool{
	"BHD": true, "IQD": true, "JOD": true, "KWD": true, "LYD": true,
	"OMR": true, "TND": true,
}

// Waehrungsstellen gibt die Nachkommastellen einer Währung.
//
// **Der zweite Wert sagt, ob die Währung bekannt ist.** Eine unbekannte
// bekommt zwei Stellen — das ist die häufigste Wahl —, und der Aufrufer
// erfährt, dass geraten wurde. Ein stilles Zwei wäre eine Zusage, die
// niemand geprüft hat (§10).
func Waehrungsstellen(code string) (stellen int, bekannt bool) {
	gross := strings.ToUpper(strings.TrimSpace(code))

	switch {
	case nullstellige[gross]:
		return 0, true

	case dreistellige[gross]:
		return 3, true

	case istWaehrungscode(gross):
		return 2, true

	default:
		return 2, false
	}
}

// istWaehrungscode sagt, ob etwas überhaupt einer sein kann.
//
// **Drei lateinische Buchstaben** — ISO 4217. **Und `len` zählt Bytes**:
// »€« ist drei davon und keine Währung; der erste Anlauf ließ es durchgehen.
func istWaehrungscode(gross string) bool {
	if len([]rune(gross)) != 3 {
		return false
	}

	for _, zeichen := range gross {
		if zeichen < 'A' || zeichen > 'Z' {
			return false
		}
	}

	return true
}

// GeldAus liest einen Betrag in kanonischer Form.
//
// **Die kanonische Form**: optionales Minus, Ziffern, optional ein Punkt und
// höchstens so viele Nachkommastellen, wie die Währung hat. Kein
// Tausendertrennzeichen, kein Währungszeichen, kein Leerzeichen.
//
// **Ein leerer Text ist eine Null und kein Fehler** — „kein Preis genannt"
// ist eine Antwort, und sie kommt in jeder zweiten Zeile einer Sammlung vor.
//
// **Alles andere wird erklärt** (AK-03): der Satz nennt, was dastand und was
// erwartet wird.
func GeldAus(text, waehrung string) (int64, error) {
	geglaettet := strings.TrimSpace(text)
	if geglaettet == "" {
		return 0, nil
	}

	stellen, _ := Waehrungsstellen(waehrung)

	vorzeichen := int64(1)
	if nach, gefunden := strings.CutPrefix(geglaettet, "-"); gefunden {
		vorzeichen, geglaettet = -1, nach
	}

	ganzes, gebrochenes, hatPunkt := strings.Cut(geglaettet, ".")

	if fehler := zifferngruppen(text, ganzes, gebrochenes, hatPunkt); fehler != nil {
		return 0, fehler
	}

	// **Überzählige Nullen gehen verlustfrei auf** (Zweitblick über Phase
	// 6.3, M12): 0.5.x schrieb 150 ¥ als »150.00«, und jeder Yen-Preis fiel
	// beim Einlesen weg. Eine andere Ziffer dort bleibt abgewiesen — sie
	// wäre ein Betrag, den die Währung nicht kennt.
	if len(gebrochenes) > stellen && strings.Trim(gebrochenes[stellen:], "0") == "" {
		gebrochenes = gebrochenes[:stellen]
	}

	if len(gebrochenes) > stellen {
		return 0, NewFieldError("erwerbspreis", SchluesselValidierung,
			"»%s« hat %d Nachkommastellen; %s kennt %d.",
			text, len(gebrochenes), waehrungsname(waehrung), stellen)
	}

	// **Aufgefüllt und nicht gerundet**: »1.2« in Euro sind 120 Cent, und
	// wer hier rundete, erfände eine Stelle, die niemand geschrieben hat.
	roh := ganzes + gebrochenes + strings.Repeat("0", stellen-len(gebrochenes))

	betrag, fehler := strconv.ParseInt(roh, 10, 64)
	if fehler != nil {
		return 0, NewFieldError("erwerbspreis", SchluesselValidierung,
			"»%s« ist zu groß für einen Betrag.", text)
	}

	return vorzeichen * betrag, nil
}

// zifferngruppen prüft, dass vor und nach dem Punkt nur Ziffern stehen.
func zifferngruppen(roh, ganzes, gebrochenes string, hatPunkt bool) error {
	// **Vor dem Punkt steht mindestens eine Ziffer.** »․17« ist zwar zu
	// erraten, und genau das ist der Grund: LOC-03.AK-03 verlangt, dass eine
	// mehrdeutige Eingabe **erklärt** wird und nicht gedeutet. Die
	// kanonische Form schreibt 0.17.
	if ganzes == "" {
		return NewFieldError("erwerbspreis", SchluesselValidierung,
			"»%s« ist kein Betrag — vor dem Punkt fehlt die Zahl; "+
				"erwartet wird eine wie 0.17.", roh)
	}

	for _, teil := range []string{ganzes, gebrochenes} {
		for _, zeichen := range teil {
			if zeichen < '0' || zeichen > '9' {
				return NewFieldError("erwerbspreis", SchluesselValidierung,
					"»%s« ist kein Betrag — der Punkt trennt die "+
						"Nachkommastellen, ein Komma gilt nicht, und ein "+
						"Währungszeichen gehört nicht hinein.", roh)
			}
		}
	}

	if hatPunkt && gebrochenes == "" {
		return NewFieldError("erwerbspreis", SchluesselValidierung,
			"»%s« endet auf einem Punkt.", roh)
	}

	return nil
}

// GeldText schreibt einen Betrag in kanonischer Form.
//
// **Mit genau den Stellen der Währung** — »19.99« in Euro, »150« in Yen,
// »1.234« in Dinar. Wer alle mit zwei Stellen schriebe, behauptete beim Yen
// eine Genauigkeit, die es nicht gibt.
func GeldText(betrag int64, waehrung string) string {
	stellen, _ := Waehrungsstellen(waehrung)

	vorzeichen := ""
	if betrag < 0 {
		vorzeichen, betrag = "-", -betrag
	}

	if stellen == 0 {
		return vorzeichen + strconv.FormatInt(betrag, 10)
	}

	teiler := int64(1)
	for range stellen {
		teiler *= 10
	}

	return fmt.Sprintf("%s%d.%0*d",
		vorzeichen, betrag/teiler, stellen, betrag%teiler)
}

// waehrungsname gibt den Code für eine Meldung — oder ein Wort dafür, dass
// keiner dasteht.
//
// **Weil ein Satz „…; kennt 2" ohne Gegenstand keiner ist** (§8).
func waehrungsname(waehrung string) string {
	if geglaettet := strings.TrimSpace(waehrung); geglaettet != "" {
		return geglaettet
	}

	return "eine Währung ohne Angabe"
}
