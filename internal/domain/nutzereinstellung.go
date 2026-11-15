// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von Phase
// 4.12 (#1351),
// Punkte 9 und 11 der Kommentierung vom 2026-09-05.
//
// # Zwei Sorten, und beide sind Vorgaben
//
// **Was hier steht, gilt für die Person; was im Profil steht, sticht es.**
// „Meine Sprache" ist eine Aussage über einen Menschen und nicht über eine
// Arbeitsumgebung; sie in jedem Profil noch einmal zu setzen hieße, dieselbe
// Antwort so oft zu geben, wie es Profile gibt.
//
// # Warum ein Override nur strenger machen kann
//
// **Weil das System recht hat, wenn es weniger verlangt.** Wer im
// Betriebssystem „weniger Bewegung" eingestellt hat, hat einen Grund; eine
// Anwendung, die das mit einer eigenen Einstellung **aufhebt**, arbeitet
// gegen ihn. Die Übersteuerung geht deshalb nur in eine Richtung: sie schaltet
// zu, nie ab.

package domain

import "slices"

// Die Schlüssel der Nutzereinstellungen.
//
// **Mit Bereichspräfix**, weil sie neben den Profileinstellungen stehen und
// beide in derselben Antwort landen: `barrierefreiheit.bewegung` sagt, wohin
// es gehört, `bewegung` allein nicht.
const (
	// NutzerBewegung übersteuert `prefers-reduced-motion`.
	NutzerBewegung = "barrierefreiheit.bewegung"

	// NutzerKontrast übersteuert `prefers-contrast`.
	NutzerKontrast = "barrierefreiheit.kontrast"

	// NutzerFaktorbeimWechsel sagt, ob der Profilwechsel den zweiten Faktor
	// verlangt.
	//
	// **Eine Übersteuerung, die lockert — und deshalb die einzige ihrer
	// Art.** Auf einem Rechner mit einem Menschen ist ein Einmalkennwort bei
	// jedem Profilwechsel Reibung ohne Gegenwert; wer das anders sieht, lässt
	// es an. **Am Anmelden ändert sie nichts**: dort steht der Faktor für das
	// ein, wofür es ihn gibt.
	NutzerFaktorbeimWechsel = "sicherheit.faktorbeimwechsel"
)

// Die Werte der drei Schlüssel.
//
// **Der erste ist die Vorgabe**, und er heißt jeweils „so wie bisher": ein
// frischer Bestand verhält sich, wie er sich immer verhalten hat.
const (
	// NachSystem heißt: was das Betriebssystem sagt.
	NachSystem = "system"

	// BewegungReduziert schaltet Bewegung ab — auch wenn das System sie
	// erlaubt.
	BewegungReduziert = "reduziert"

	// KontrastHoch schaltet die Faserstruktur ab und hebt die Farben, bis
	// jeder Text 7 : 1 hält — auch wenn das System es nicht verlangt
	// (UX-10.AK-03, #1682; gemessen in `docs/KONTRASTE.md`).
	KontrastHoch = "hoch"

	// FaktorVerlangen heißt: auch beim Profilwechsel.
	FaktorVerlangen = "verlangen"

	// FaktorNurAnmelden heißt: nur beim Anmelden.
	FaktorNurAnmelden = "nurAnmelden"
)

// nutzereinstellung ist eine Zeile der Tabelle darunter.
type nutzereinstellung struct {
	Schluessel string
	Werte      []string
	Vorgabe    string

	// Pruefung ist die Regel für einen Wert ohne Werteliste — `nil`, wo es
	// keine gibt. Dieselbe Begründung wie bei [einstellung.Pruefung].
	Pruefung func(wert string) error
}

// nutzereinstellungen ist die Aufzählung — **die** Wahrheit über sie.
//
// **Von Hand geführt und an einer Stelle**, wie `einstellungen` daneben: eine
// zweite Liste liefe zurück, und niemand sähe es.
var nutzereinstellungen = []nutzereinstellung{
	{
		Schluessel: NutzerBewegung,
		Werte:      []string{NachSystem, BewegungReduziert},
		Vorgabe:    NachSystem,
	},
	{
		Schluessel: NutzerKontrast,
		Werte:      []string{NachSystem, KontrastHoch},
		Vorgabe:    NachSystem,
	},
	{
		Schluessel: NutzerFaktorbeimWechsel,
		Werte:      []string{FaktorVerlangen, FaktorNurAnmelden},
		Vorgabe:    FaktorVerlangen,
	},

	// **Die vier, die ein Profil erben kann** — Werkstück 9 (#1351), Punkt 9.
	//
	// **Dieselben Schlüssel wie bei den Profileinstellungen**, und das ist
	// Absicht: es ist dieselbe Frage auf einer anderen Ebene. Zwei Namen für
	// „Sprache" wären zwei Fragen, und niemand wüsste, welche gemeint ist.
	//
	// **Ohne `erben`**: hier ist der Boden. Die **Vorgabewerte** stehen
	// deshalb hier und nicht mehr bei den Profileinstellungen — dort gilt
	// `erben`.
	{
		Schluessel: EinstellungSprache,
		Werte:      fenstersprachen,
	},
	{
		Schluessel: EinstellungErscheinungsbild,
		Werte:      []string{"system", "hell", "dunkel"},
		Vorgabe:    "system",
	},
	{
		Schluessel: EinstellungSchriftgroesse,
		Werte:      []string{"klein", "normal", "gross", "sehrgross"},
		Vorgabe:    "normal",
	},
	{
		Schluessel: EinstellungSchriftart,
		Werte:      schriftarten(),
		Vorgabe:    SchriftartPergament,
	},
}

// Nutzereinstellungsreihe gibt die Schlüssel in Anzeigereihenfolge.
func Nutzereinstellungsreihe() []string {
	reihe := make([]string, 0, len(nutzereinstellungen))
	for _, eine := range nutzereinstellungen {
		reihe = append(reihe, eine.Schluessel)
	}

	return reihe
}

// ErlaubteNutzerwerte gibt die Werte eines Schlüssels, in Anzeigereihenfolge.
//
// **Eine Kopie**: ein Aufrufer, der die Liste änderte, änderte die Regel.
//
// **Nie `nil`, sondern eine leere Liste** (§3) — Werkstück O3 von Phase 5.2
// (#1427):
// seit den Spaltenbreiten gibt es auch hier Einstellungen ohne Werteliste, und
// `slices.Clone(nil)` ist `nil`. Über die JSON-Naht würde daraus `null`, und
// das Fenster bekäme **keine** Liste statt einer leeren — genau der Fall, für
// den es `TestKeineBindungGibtEineListeAlsNull` gibt.
func ErlaubteNutzerwerte(schluessel string) []string {
	eine, _ := findeNutzereinstellung(schluessel)

	if len(eine.Werte) == 0 {
		return []string{}
	}

	return slices.Clone(eine.Werte)
}

// NutzerVorgabe gibt, was ohne Wahl gilt — leer für einen unbekannten
// Schlüssel.
func NutzerVorgabe(schluessel string) string {
	eine, _ := findeNutzereinstellung(schluessel)

	return eine.Vorgabe
}

// PruefeNutzereinstellung urteilt über ein Paar aus Schlüssel und Wert.
func PruefeNutzereinstellung(schluessel, wert string) error {
	eine, bekannt := findeNutzereinstellung(schluessel)
	if !bekannt {
		return NewFieldError("schluessel", SchluesselValidierung,
			"%q ist keine Einstellung eines Nutzers.", schluessel)
	}

	if fehler := gegenListe(eine.Werte, schluessel, wert); fehler != nil {
		return fehler
	}

	// **Und zuletzt die eigene Regel** — dieselbe Reihenfolge wie bei
	// [PruefeEinstellung], und aus demselben Grund.
	if eine.Pruefung != nil {
		return eine.Pruefung(wert)
	}

	return nil
}

// findeNutzereinstellung gibt die Zeile eines Schlüssels.
func findeNutzereinstellung(schluessel string) (nutzereinstellung, bool) {
	for _, eine := range nutzereinstellungen {
		if eine.Schluessel == schluessel {
			return eine, true
		}
	}

	return nutzereinstellung{}, false
}
