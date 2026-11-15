// **Welche Kanäle das Protokoll hat und wie ausführlich es ist** — OPS-04
// (#1721),
// Marke `new`: MSE2 schreibt kein Protokoll.
//
// # Drei Kanäle, unabhängig
//
// **Datei, Systemprotokoll, Standardfehlerausgabe** — je an oder aus. Das
// Systemprotokoll ist auf macOS das Unified Log; journald gibt es hier nicht,
// und `docs/PORTIERUNG.md` beschreibt, was auf Linux an seine Stelle träte.
// Dazu Stufe, Format, Ort, Drehung, Aufbewahrung und ein Filter auf
// Komponenten für die ausführliche Stufe.
//
// # Alle gelten ab dem nächsten Start
//
// **Das Protokoll entsteht vor dem Bestand** (#848): was beim Öffnen
// schiefgeht, soll dort stehen. Die Einstellungen liegen im Bestand; gelesen
// werden sie, sobald er offen ist, und bis dahin gilt die Vorgabe.
//
// # Ausführlich heißt: kurz aufbewahrt (AK-03)
//
// **Höchstens 1 MB je Datei und eine Vorfassung**, gleich was eingestellt
// ist. Ein Protokoll, das jeden Schritt kennt, soll nicht wochenlang liegen;
// die Diagnose sagt, dass es an ist.

package domain

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Die Schlüssel der Protokolleinstellungen.
const (
	ProtokollStufe        = "protokoll.stufe"
	ProtokollDatei        = "protokoll.datei"
	ProtokollSystem       = "protokoll.system"
	ProtokollAusgabe      = "protokoll.ausgabe"
	ProtokollFormat       = "protokoll.format"
	ProtokollGrenze       = "protokoll.grenze"
	ProtokollAufbewahrung = "protokoll.aufbewahrung"
	ProtokollKomponenten  = "protokoll.komponenten"
	ProtokollOrt          = "protokoll.ort"
)

// Die Werte der Protokolleinstellungen.
const (
	ProtokollAn  = "an"
	ProtokollAus = "aus"

	ProtokollstufeFehler       = "fehler"
	ProtokollstufeWarnung      = "warnung"
	ProtokollstufeInfo         = "info"
	ProtokollstufeAusfuehrlich = "ausfuehrlich"

	ProtokollformatText = "text"
	ProtokollformatJSON = "json"
)

// Die Grenzen der Drehung.
const (
	protokollgrenzeVorgabe      = 4
	protokollgrenzeHoechstens   = 64
	protokollgrenzeAusfuehrlich = 1
	aufbewahrungVorgabe         = 1
	aufbewahrungHoechstens      = 9
)

// Protokollschluessel gibt die Schlüssel in Anzeigereihenfolge.
func Protokollschluessel() []string {
	return []string{
		ProtokollStufe, ProtokollDatei, ProtokollSystem, ProtokollAusgabe, ProtokollFormat,
		ProtokollGrenze, ProtokollAufbewahrung, ProtokollKomponenten, ProtokollOrt,
	}
}

// protokolleinstellungen sind die Zeilen für die Reihe der Einstellungen.
func protokolleinstellungen() []einstellung {
	neu := func(schluessel string, werte []string, vorgabe string, pruefung func(string) error) einstellung {
		return einstellung{
			Schluessel: schluessel, Wirkung: WirkungNeustart, Bereich: BereichAnwendung,
			Werte: werte, Vorgabe: vorgabe, Pruefung: pruefung,
		}
	}

	anAus := []string{ProtokollAn, ProtokollAus}

	return []einstellung{
		neu(ProtokollStufe, []string{
			ProtokollstufeFehler, ProtokollstufeWarnung, ProtokollstufeInfo, ProtokollstufeAusfuehrlich,
		}, ProtokollstufeInfo, nil),
		neu(ProtokollDatei, anAus, ProtokollAn, nil),
		neu(ProtokollSystem, anAus, ProtokollAn, nil),
		neu(ProtokollAusgabe, []string{ProtokollAus, ProtokollAn}, ProtokollAus, nil),
		neu(ProtokollFormat, []string{ProtokollformatText, ProtokollformatJSON}, ProtokollformatText, nil),
		neu(ProtokollGrenze, nil, strconv.Itoa(protokollgrenzeVorgabe), zahlZwischen(1, protokollgrenzeHoechstens)),
		neu(ProtokollAufbewahrung, nil, strconv.Itoa(aufbewahrungVorgabe), zahlZwischen(1, aufbewahrungHoechstens)),
		neu(ProtokollKomponenten, nil, "", PruefeProtokollkomponenten),
		neu(ProtokollOrt, nil, "", PruefeOrdnerangabe),
	}
}

// zahlZwischen prüft eine ganze Zahl in Grenzen — leer heißt: die Vorgabe.
func zahlZwischen(kleinste, groesste int) func(string) error {
	return func(wert string) error {
		if wert == "" {
			return nil
		}

		zahl, fehler := strconv.Atoi(wert)
		if fehler != nil || zahl < kleinste || zahl > groesste {
			return NewFieldError("wert", SchluesselValidierung,
				"»%s« ist keine ganze Zahl von %d bis %d.", wert, kleinste, groesste)
		}

		return nil
	}
}

// komponentenliste erkennt Paketnamen, durch Kommas getrennt.
var komponentenliste = regexp.MustCompile(`^[a-z]+(,[a-z]+)*$`)

// PruefeProtokollkomponenten urteilt über den Filter — leer heißt: alle.
func PruefeProtokollkomponenten(wert string) error {
	if wert == "" || komponentenliste.MatchString(wert) {
		return nil
	}

	return NewFieldError("wert", SchluesselValidierung,
		"»%s« ist keine Liste von Komponenten — Namen in Kleinbuchstaben, durch Kommas getrennt, etwa app.", wert)
}

// Protokollregel ist, was aus den Einstellungen für das Protokoll folgt.
type Protokollregel struct {
	Stufe        string
	Datei        bool
	System       bool
	Ausgabe      bool
	Format       string
	GrenzeBytes  int64
	Aufbewahrung int
	Komponenten  []string

	// Ort ist die Angabe eines Ordners — leer heißt: neben dem Bestand.
	Ort string
}

// Ausfuehrlich sagt, ob jede Zeile geschrieben wird.
func (r Protokollregel) Ausfuehrlich() bool { return r.Stufe == ProtokollstufeAusfuehrlich }

// ProtokollregelAus macht aus den wirksamen Einstellungen eine Regel.
//
// **Ein unlesbarer Wert fällt auf die Vorgabe** — ein von Hand geänderter
// Bestand darf das Protokoll nicht abschalten oder ins Unendliche wachsen
// lassen.
func ProtokollregelAus(werte map[string]string) Protokollregel {
	wahl := func(schluessel string, erlaubt []string, vorgabe string) string {
		if slices.Contains(erlaubt, werte[schluessel]) {
			return werte[schluessel]
		}

		return vorgabe
	}

	zahl := func(schluessel string, groesste, vorgabe int) int {
		gelesen, fehler := strconv.Atoi(werte[schluessel])
		if fehler != nil || gelesen < 1 || gelesen > groesste {
			return vorgabe
		}

		return gelesen
	}

	regel := Protokollregel{
		Stufe: wahl(ProtokollStufe, []string{
			ProtokollstufeFehler, ProtokollstufeWarnung, ProtokollstufeInfo, ProtokollstufeAusfuehrlich,
		}, ProtokollstufeInfo),
		Datei:        werte[ProtokollDatei] != ProtokollAus,
		System:       werte[ProtokollSystem] != ProtokollAus,
		Ausgabe:      werte[ProtokollAusgabe] == ProtokollAn,
		Format:       wahl(ProtokollFormat, []string{ProtokollformatText, ProtokollformatJSON}, ProtokollformatText),
		GrenzeBytes:  int64(zahl(ProtokollGrenze, protokollgrenzeHoechstens, protokollgrenzeVorgabe)) << 20,
		Aufbewahrung: zahl(ProtokollAufbewahrung, aufbewahrungHoechstens, aufbewahrungVorgabe),
		Komponenten:  []string{},
		Ort:          werte[ProtokollOrt],
	}

	if PruefeProtokollkomponenten(werte[ProtokollKomponenten]) == nil && werte[ProtokollKomponenten] != "" {
		regel.Komponenten = strings.Split(werte[ProtokollKomponenten], ",")
	}

	if regel.Ausfuehrlich() {
		regel.GrenzeBytes = min(regel.GrenzeBytes, protokollgrenzeAusfuehrlich<<20)
		regel.Aufbewahrung = 1
	}

	return regel
}
