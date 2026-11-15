package domain

// **Bereiche und Funktionen, die sich abschalten lassen** — OPS-10
// (#1908,
// Vermerk [0158](../../docs/entscheidungen/0158-ein-schalter-verbirgt-er-sichert-nicht.md)).
//
// # Was ein Schalter ist
//
// **Eine Einstellung am Rechner**, die einen Arbeitsbereich oder eine
// Funktion an allen vier Flächen verbirgt. Welche Fähigkeit an welchem
// Schalter hängt, steht in `docs/SCHALTER.md`; welche Wege das sind, in der
// Wegetafel daneben ([EingebauteWegetafel]).
//
// # Was er nicht ist
//
// **Keine Autorisierung.** Ein abgeschalteter Bereich ist verborgen, nicht
// gesichert: die Rollen aus `docs/AUTORISIERUNG.md` gelten unverändert, und
// ein Sicherheitsriegel wird nie ein Schalter — wer einen Riegel hier
// eintrüge, gäbe Menschen einen Knopf, der ihn öffnet.

import "strings"

// Schalterart sagt, ob ein Schalter einen Arbeitsbereich oder eine Funktion
// abschaltet.
type Schalterart string

// Die zwei Arten.
const (
	// SchalterBereich schaltet einen Arbeitsbereich aus der Seitenleiste ab.
	SchalterBereich Schalterart = "bereich"

	// SchalterFunktion schaltet eine Funktion quer durch die Bereiche ab.
	SchalterFunktion Schalterart = "funktion"
)

// Reife sagt, wie weit eine Funktion ist — OPS-10.AK-04.
type Reife string

// Die drei Reifen.
const (
	// ReifeStabil ist der Normalfall.
	ReifeStabil Reife = "stabil"

	// ReifeVorschau ist neu und darf mit Vorgabe „aus“ ausgeliefert werden.
	ReifeVorschau Reife = "vorschau"

	// ReifeAuslaufend verschwindet in einer späteren Fassung.
	ReifeAuslaufend Reife = "auslaufend"
)

// Schalter ist ein Eintrag im Verzeichnis.
type Schalter struct {
	// Kennung ist der Name im Einstellungsschlüssel und in der Abweisung.
	Kennung string

	Art Schalterart

	// Vorgabe ist, ob er ohne eigene Wahl an ist.
	Vorgabe bool

	Reife Reife
}

// Die zwei Werte eines Schalters.
const (
	SchalterAn  = "an"
	SchalterAus = "aus"
)

// schalterverzeichnis ist die eine Aufzählung.
//
// **Gestalten und das Programm stehen nicht darin**: sie sind der Kern, und
// ein Kern, den man abschalten könnte, ließe ein Fenster ohne Weg zurück zu
// den Einstellungen.
var schalterverzeichnis = []Schalter{
	{Kennung: "sammeln", Art: SchalterBereich, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "spielen", Art: SchalterBereich, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "handeln", Art: SchalterBereich, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "ressourcen", Art: SchalterBereich, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "ki", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "netz", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "markt", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "turniere", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "dienst", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
	{Kennung: "werkzeugserver", Art: SchalterFunktion, Vorgabe: true, Reife: ReifeStabil},
}

// Schalterverzeichnis gibt die Schalter in Anzeigereihenfolge — erst die
// Bereiche in der Ordnung der Seitenleiste, dann die Funktionen.
func Schalterverzeichnis() []Schalter {
	return append([]Schalter(nil), schalterverzeichnis...)
}

// schaltervorsatz steht vor jeder Kennung im Einstellungsschlüssel.
const schaltervorsatz = "schalter."

// SchalterEinstellung gibt den Einstellungsschlüssel eines Schalters.
func SchalterEinstellung(kennung string) string { return schaltervorsatz + kennung }

// schaltereinstellungen gibt je Schalter eine Einstellung.
//
// **Sofort wirksam**: ein Bereich, der erst nach einem Neustart verschwindet,
// stünde bis dahin als Weg da, der abweist.
func schaltereinstellungen() []einstellung {
	zeilen := make([]einstellung, 0, len(schalterverzeichnis))

	for _, schalter := range schalterverzeichnis {
		vorgabe := SchalterAus
		if schalter.Vorgabe {
			vorgabe = SchalterAn
		}

		zeilen = append(zeilen, einstellung{
			Schluessel: SchalterEinstellung(schalter.Kennung),
			Bereich:    BereichSchalter,
			Werte:      []string{SchalterAn, SchalterAus},
			Wirkung:    WirkungSofort,
			Vorgabe:    vorgabe,
		})
	}

	return zeilen
}

// Schalterstand sagt, welche Schalter gerade aus sind.
type Schalterstand struct {
	aus map[string]bool
}

// SchalterstandAus liest den Stand aus wirksamen Einstellungen.
//
// **Ausgeschaltet ist nur, was „aus“ sagt** — oder ohne Eintrag mit Vorgabe
// „aus“ ausgeliefert ist. Ein Wert, den es nicht gibt, gilt mit der Vorgabe
// (OPS-10: „ein Schalter, den die Konfiguration nicht kennt, gilt mit seiner
// Vorgabe“).
func SchalterstandAus(werte map[string]string) Schalterstand {
	stand := Schalterstand{aus: map[string]bool{}}

	for _, schalter := range schalterverzeichnis {
		switch werte[SchalterEinstellung(schalter.Kennung)] {
		case SchalterAus:
			stand.aus[schalter.Kennung] = true
		case SchalterAn:
		default:
			if !schalter.Vorgabe {
				stand.aus[schalter.Kennung] = true
			}
		}
	}

	return stand
}

// Aus sagt, ob ein Schalter aus ist; eine unbekannte Kennung ist es nicht.
func (s Schalterstand) Aus(kennung string) bool { return s.aus[kennung] }

// AllesAn sagt, ob kein Schalter aus ist — der Normalfall, und der Weg, auf
// dem ein Tor ohne weitere Arbeit durchlässt.
func (s Schalterstand) AllesAn() bool { return len(s.aus) == 0 }

// IstSchaltereinstellung sagt, ob ein Einstellungsschlüssel ein Schalter ist.
func IstSchaltereinstellung(schluessel string) bool {
	return strings.HasPrefix(schluessel, schaltervorsatz)
}
