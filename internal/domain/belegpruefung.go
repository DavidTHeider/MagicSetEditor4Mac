package domain

// **Was ein Modell belegen darf** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Der Satz, für den es diese Datei gibt
//
// **„Modell darf nur diese IDs referenzieren; unbekannte IDs/erfundene
// Regelnummern/URLs ergeben Validierungsbefund, keine vertrauenswürdige
// Quellenanzeige."**
//
// Damit ist eine erfundene Regelnummer kein Darstellungsproblem, sondern ein
// **Befund** — er fällt auf, bevor jemand den Satz liest.
//
// # Und der zweite, der weiter geht
//
// **„ID-Gültigkeit beweist noch keine inhaltliche Folgerichtigkeit."** Was
// hier geprüft wird, ist, **ob** es den Beleg gibt und ob er genannt wurde.
// Ob er die Aussage trägt, prüft kein Programm — und genau deshalb steht der
// Satz hier und nicht als stille Annahme im Code.

import (
	"regexp"
	"slices"
	"strings"
)

// Belegbefund ist ein Mangel an einer Aussage mit Quellenanspruch.
type Belegbefund struct {
	// Art ist `unbekannt`, `ohnebeleg`, `regelnummer` oder `adresse`.
	Art string

	// Stelle ist, was beanstandet wird — die genannte Kennung, die Nummer,
	// die Adresse.
	Stelle string

	// Meldung sagt, was daran falsch ist.
	Meldung string
}

// Die vier Arten.
const (
	// BefundUnbekannteKennung: das Modell nennt einen Beleg, den es nicht gibt.
	BefundUnbekannteKennung = "unbekannt"

	// BefundOhneBeleg: eine Aussage mit Quellenanspruch nennt keinen.
	BefundOhneBeleg = "ohnebeleg"

	// BefundErfundeneRegelnummer: eine Regelnummer, die in keinem Beleg steht.
	BefundErfundeneRegelnummer = "regelnummer"

	// BefundErfundeneAdresse: eine Adresse, die in keinem Beleg steht.
	BefundErfundeneAdresse = "adresse"
)

// belegmuster trifft eine genannte Belegkennung — `[B1]`.
//
// **In eckigen Klammern**, weil das die Gestalt ist, in der ein Modell eine
// Fußnote setzt, und weil sie in keinem Kartentext vorkommt.
var belegmuster = regexp.MustCompile(`\[(B\d+)\]`)

// regelnummermuster trifft eine Regelnummer — `205.4a`, `509.1`.
//
// **Mit mindestens einem Punkt**: „3 Schaden" ist keine Regelnummer, und ein
// Muster ohne Punkt striche jede Zahl an.
var regelnummermuster = regexp.MustCompile(`\b\d{3}\.\d+[a-z]?\b`)

// adressmuster trifft eine Adresse.
var adressmuster = regexp.MustCompile(`https?://[^\s)\]}"']+`)

// PruefeBelege hält eine Modellantwort gegen die vergebenen Belege.
//
// **`text` ist, was das Modell geschrieben hat.** Was hier herauskommt, sind
// Befunde und kein Urteil: eine Antwort mit Befunden wird **gezeigt**, aber
// nicht als belegt ausgegeben.
func PruefeBelege(text string, belege []Beleg) []Belegbefund {
	befunde := make([]Belegbefund, 0)

	befunde = append(befunde, unbekannteKennungen(text, belege)...)
	befunde = append(befunde, erfundeneRegelnummern(text, belege)...)

	return append(befunde, erfundeneAdressen(text, belege)...)
}

// unbekannteKennungen findet genannte Belege, die es nicht gibt.
func unbekannteKennungen(text string, belege []Beleg) []Belegbefund {
	bekannt := make([]string, 0, len(belege))
	for _, beleg := range belege {
		bekannt = append(bekannt, string(beleg.Kennung))
	}

	befunde := make([]Belegbefund, 0)
	gesehen := map[string]bool{}

	for _, treffer := range belegmuster.FindAllStringSubmatch(text, -1) {
		genannt := treffer[1]
		if gesehen[genannt] || slices.Contains(bekannt, genannt) {
			continue
		}

		gesehen[genannt] = true

		befunde = append(befunde, Belegbefund{
			Art: BefundUnbekannteKennung, Stelle: genannt,
			Meldung: "Diese Antwort beruft sich auf einen Beleg, den es in " +
				"diesem Kontextpaket nicht gibt.",
		})
	}

	return befunde
}

// erfundeneRegelnummern findet Nummern, die in keinem Beleg stehen.
//
// **Geprüft wird gegen den Text der Belege und nicht gegen eine Regelliste.**
// Dieses Programm hat keine; was es hat, ist, was mitgegeben wurde — und
// darüber hinaus zu behaupten, eine Nummer gäbe es nicht, wäre eine Aussage
// über das Regelwerk, die hier niemand treffen kann.
func erfundeneRegelnummern(text string, belege []Beleg) []Belegbefund {
	befunde := make([]Belegbefund, 0)
	gesehen := map[string]bool{}

	for _, nummer := range regelnummermuster.FindAllString(text, -1) {
		if gesehen[nummer] || inBelegen(nummer, belege) {
			continue
		}

		gesehen[nummer] = true

		befunde = append(befunde, Belegbefund{
			Art: BefundErfundeneRegelnummer, Stelle: nummer,
			Meldung: "Diese Regelnummer steht in keinem Beleg dieses " +
				"Kontextpakets. Sie mag es geben; belegt ist sie hier nicht.",
		})
	}

	return befunde
}

// erfundeneAdressen findet Adressen, die in keinem Beleg stehen.
func erfundeneAdressen(text string, belege []Beleg) []Belegbefund {
	befunde := make([]Belegbefund, 0)
	gesehen := map[string]bool{}

	for _, adresse := range adressmuster.FindAllString(text, -1) {
		if gesehen[adresse] || inBelegen(adresse, belege) {
			continue
		}

		gesehen[adresse] = true

		befunde = append(befunde, Belegbefund{
			Art: BefundErfundeneAdresse, Stelle: adresse,
			Meldung: "Diese Adresse steht in keinem Beleg dieses " +
				"Kontextpakets und wird nicht als Quelle angezeigt.",
		})
	}

	return befunde
}

// inBelegen sagt, ob eine Zeichenfolge in einem Beleg vorkommt.
func inBelegen(gesucht string, belege []Beleg) bool {
	for _, beleg := range belege {
		if beleg.Regelnummer == gesucht || beleg.Bezugsadresse == gesucht ||
			strings.Contains(beleg.Ausschnitt, gesucht) ||
			strings.Contains(beleg.Titel, gesucht) {
			return true
		}
	}

	return false
}

// NenntEinenBeleg sagt, ob ein Satz sich überhaupt auf einen beruft.
//
// **Für die Zeile „Jede als quellenbasiert dargestellte Aussage hat einen
// tatsächlichen Beleg".** Was ein Satz ist, entscheidet der Aufrufer; was
// eine Berufung ist, steht hier.
func NenntEinenBeleg(satz string) bool {
	return belegmuster.MatchString(satz)
}
