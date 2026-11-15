// **Was über eine Einstellung auszusagen ist** — Bereich, Werte, Vorgabe, und
// die Prüfung darüber.
//
// **Eine eigene Datei**, weil `einstellung.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **welche** Einstellungen es
// gibt und in welcher Reihe; hier, **was man über eine erfährt** und **was
// als Wert durchgeht**.

package domain

import (
	"slices"
	"strings"
)

// finde gibt die Zeile eines Schlüssels — der zweite Wert sagt, ob es sie gibt.
//
// **Ein Lauf über vier Zeilen und keine Abbildung daneben.** Eine zweite
// Datenstruktur wäre wieder eine zweite Liste, und genau die hat gekostet.
func finde(schluessel string) (einstellung, bool) {
	for _, eine := range einstellungen {
		if eine.Schluessel == schluessel {
			return eine, true
		}
	}

	return einstellung{}, false
}

// Einstellungsreihe gibt die Einstellungen in Anzeigereihenfolge.
func Einstellungsreihe() []string {
	reihe := make([]string, 0, len(einstellungen))
	for _, eine := range einstellungen {
		reihe = append(reihe, eine.Schluessel)
	}

	return reihe
}

// EinstellungsBereich gibt den Bereich — leer für einen unbekannten Schlüssel.
func EinstellungsBereich(schluessel string) string {
	eine, _ := finde(schluessel)

	return eine.Bereich
}

// ErlaubteEinstellungswerte gibt die Werte eines Schlüssels, in Anzeigereihenfolge.
//
// **Eine Kopie**: ein Aufrufer, der die Liste änderte, änderte die Regel.
func ErlaubteEinstellungswerte(schluessel string) []string {
	eine, _ := finde(schluessel)

	// **Nie `nil`, sondern eine leere Liste** (§3) — Werkstück 9 (#1351),
	// Punkt 8. Eine Einstellung ohne Werteliste gibt es seither, und
	// `slices.Clone(nil)` ist `nil`: über die JSON-Naht wird daraus `null`,
	// und das Fenster bekäme **keine** Liste statt einer leeren. Gefunden hat
	// es `TestKeineBindungGibtEineListeAlsNull` — genau der Fall, für den es
	// ihn gibt.
	if len(eine.Werte) == 0 {
		return []string{}
	}

	return slices.Clone(eine.Werte)
}

// EinstellungsVorgabe gibt, was ohne Wahl gilt — leer, wenn es keine gibt.
func EinstellungsVorgabe(schluessel string) string {
	eine, _ := finde(schluessel)

	return eine.Vorgabe
}

// PruefeEinstellung urteilt über ein Paar aus Schlüssel und Wert.
func PruefeEinstellung(schluessel, wert string) error {
	eine, bekannt := finde(schluessel)
	if !bekannt {
		return NewFieldError("schluessel", SchluesselValidierung,
			"%q ist keine Einstellung dieser Anwendung.", schluessel)
	}

	if fehler := gegenListe(eine.Werte, schluessel, wert); fehler != nil {
		return fehler
	}

	// **Und zuletzt die eigene Regel** — Werkstück O3 von Phase 5.2
	// (#1427).
	//
	// **Nach der Liste und nicht statt ihr**: eine Einstellung darf beides
	// haben. Wer sie in den Freitextzweig gesetzt hätte, ließe die Regel
	// stillschweigend fallen, sobald jemand der Einstellung eine Werteliste
	// gibt — und das ist die Art Lücke, die niemand sieht.
	if eine.Pruefung != nil {
		return eine.Pruefung(wert)
	}

	return nil
}

// gegenListe urteilt über einen Wert anhand der erlaubten.
//
// **Eine leere Liste nimmt freien Text** — Werkstück 9 von Phase 4.12
// (#1351), Punkt 8.
//
// # Warum ein leeres `Werte` das sagt und kein Schalter daneben
//
// **Weil es dasselbe sagt und nur einmal.** Ein `Freitext bool` neben einer
// leeren Liste wären zwei Angaben über eine Sache, und die zweite liefe beim
// ersten Widerspruch zurück. Was hier steht, ist die Aufzählung — und keine
// Aufzählung ist auch eine Antwort.
//
// **Geprüft wird trotzdem etwas**: ein `\x00` beendet in jedem Systemaufruf
// die Zeichenkette, und was danach steht, verschwindet stillschweigend.
//
// **Eine Stelle für beide Reihen** — die Nutzereinstellungen prüfen dasselbe,
// und zwei gleichlautende Zweige wären zwei Regeln.
func gegenListe(werte []string, schluessel, wert string) error {
	if len(werte) == 0 {
		if strings.ContainsRune(wert, 0) {
			return NewFieldError("wert", SchluesselValidierung,
				"Diese Angabe darf kein Nullbyte enthalten.")
		}

		return nil
	}

	if !slices.Contains(werte, wert) {
		return NewFieldError("wert", SchluesselValidierung,
			"%q ist kein Wert für %q.", wert, schluessel)
	}

	return nil
}

// Die drei Arten, in denen eine Einstellung einen Wert annimmt — OPS-01.AK-01
// (#1718).
const (
	// ArtListe heißt: einer aus der Werteliste.
	ArtListe = "liste"

	// ArtGeprueft heißt: keine Liste, aber eine Regel — eine Adresse, eine
	// Breite, eine Zone.
	ArtGeprueft = "geprueft"

	// ArtFrei heißt: weder Liste noch Regel; geprüft wird nur das Nullbyte.
	ArtFrei = "frei"
)

// EinstellungsArt sagt, was eine Einstellung annimmt — leer für einen
// unbekannten Schlüssel.
//
// **Aus der Zeile gelesen und nicht daneben geführt**: eine Liste ist eine
// Liste, auch wenn eine Regel dazukommt, und wer keine Liste hat, hat eine
// Regel oder ist frei. Ein eigenes Feld dafür wäre die zweite Angabe über
// dieselbe Sache, vor der `gegenListe` warnt.
func EinstellungsArt(schluessel string) string {
	eine, bekannt := finde(schluessel)
	if !bekannt {
		return ""
	}

	return artAus(eine.Werte, eine.Pruefung)
}

// NutzerArt sagt dasselbe für eine Einstellung des Menschen.
func NutzerArt(schluessel string) string {
	eine, bekannt := findeNutzereinstellung(schluessel)
	if !bekannt {
		return ""
	}

	return artAus(eine.Werte, eine.Pruefung)
}

func artAus(werte []string, pruefung func(string) error) string {
	switch {
	case len(werte) > 0:
		return ArtListe
	case pruefung != nil:
		return ArtGeprueft
	default:
		return ArtFrei
	}
}
