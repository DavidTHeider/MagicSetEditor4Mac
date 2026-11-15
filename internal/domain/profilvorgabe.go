// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7 der Kommentierung.
//
// # Warum das nicht die Einstellungsmechanik ist
//
// **Eine Einstellung hält einen aufzählbaren Wert.** `internal/domain` kennt
// die erlaubte Liste ([Einstellungsreihe]), [PruefeEinstellung] urteilt
// danach, und das Fenster baut aus derselben Liste seine Auswahl. Das trägt
// einen Kartenzustand und eine Turnierbauart.
//
// **Diese vier zeigen auf Zeilen im Bestand.** Ein Format, ein Spielpaket, ein
// Stilpaket, ein Deck, ein Lagerort — ihre erlaubten Werte stehen nicht hier,
// sie ändern sich mit dem Bestand, und zwei davon sind **mehrwertig**. Das in
// [PruefeEinstellung] hineinzuzwingen hieße, zwei Sachen unter einem Namen zu
// führen; die Prüfung der einen könnte die andere nicht mehr halten.
//
// # Was hier steht und was nicht
//
// **Hier steht die Gestalt**: welche Arten es gibt, welche mehrere Werte
// halten, und was ein Wert überhaupt sein darf. **Ob es die Zeile gibt**,
// weiß nur der Bestand — das prüft `internal/service` gegen dieselbe Liste,
// aus der die Fläche ihre Auswahl baut.

package domain

import "strings"

// Vorgabeart sagt, worauf eine Vorgabe zeigt.
type Vorgabeart string

// Die fünf Arten.
const (
	// VorgabeFormat ist das Format, mit dem ein neues Deck anfängt.
	VorgabeFormat Vorgabeart = "format"

	// VorgabeSpiel und VorgabeVorlage sind, was ein neues Set vorwählt.
	//
	// **Zwei und nicht eine**, obwohl ein Stilpaket zu einem Spiel gehört:
	// MSE2 hält beide getrennt (`settings.default_game` und
	// `gameSettingsFor(*game).default_stylesheet`, `src/gui/new_window.cpp`,
	// Fassung f9d9356 vom 2021-07-06), und wer das Spiel wechselt, behält
	// seine Stilwahl für das vorige.
	VorgabeSpiel   Vorgabeart = "spiel"
	VorgabeVorlage Vorgabeart = "vorlage"

	// VorgabeDeck sind die Präferenzdecks — **mehrere**.
	VorgabeDeck Vorgabeart = "deck"

	// VorgabeLagerort sind die zugeordneten Sammlungen — **mehrere**.
	//
	// **Ein Lagerort und keine zweite Sammlung.** Eine Sammlung gehört in
	// diesem Bestand ohnehin einem Profil (Wanderung 0035); was sich zuordnen
	// lässt, sind ihre **Orte** — „Schrank", „Kiste im Keller".
	VorgabeLagerort Vorgabeart = "lagerort"
)

// mehrwertige sind die Arten, die mehrere Werte halten.
var mehrwertige = map[Vorgabeart]bool{VorgabeDeck: true, VorgabeLagerort: true}

// Vorgabearten gibt alle fünf, in Anzeigereihenfolge.
//
// **Als Funktion und nicht als Variable**: eine Aufzählung, die sich ändern
// lässt, ist keine.
func Vorgabearten() []Vorgabeart {
	return []Vorgabeart{
		VorgabeFormat, VorgabeSpiel, VorgabeVorlage,
		VorgabeDeck, VorgabeLagerort,
	}
}

// Mehrwertig sagt, ob diese Art mehrere Werte hält.
func (a Vorgabeart) Mehrwertig() bool { return mehrwertige[a] }

// Pruefe weist eine Art ab, die es nicht gibt.
//
// **Ohne Beschneiden und ohne Kleinschreiben.** Eine Art ist ein Bezeichner
// und keine Eingabe; wer sie mit Rändern schickt, hat sie falsch gebaut, und
// ein stilles Zurechtrücken verdeckte das.
func (a Vorgabeart) Pruefe() error {
	for _, bekannt := range Vorgabearten() {
		if a == bekannt {
			return nil
		}
	}

	return NewFieldError("art", SchluesselValidierung,
		"Die Vorgabeart %q gibt es nicht.", string(a))
}

// Profilvorgabe ist, worauf ein Profil bei einer Art zeigt.
//
// **Eine Liste und kein einzelner Wert, auch bei den einwertigen Arten.** Sonst
// stünden zwei Gestalten nebeneinander, und jede Fläche müsste wissen, welche
// gerade gilt. Was eine einwertige Art unterscheidet, ist die **Prüfung**:
// mehr als einer wird abgewiesen.
type Profilvorgabe struct {
	Art Vorgabeart

	// Werte sind die Kennungen oder Namen, auf die gezeigt wird.
	//
	// **Leer heißt: keine Vorgabe.** Das ist ein Zustand und kein Fehlen —
	// ein frisches Profil zeigt auf nichts.
	Werte []string
}

// Validate prüft Art und Werte — **nicht**, ob es die Zeilen gibt.
//
// **Das kann sie nicht**, und sie soll es nicht vortäuschen: welche Formate
// es gibt, weiß der Bestand. `internal/service` hält die Werte gegen dieselbe
// Liste, aus der die Fläche ihre Auswahl baut.
func (v Profilvorgabe) Validate() error {
	if fehler := v.Art.Pruefe(); fehler != nil {
		return fehler
	}

	if !v.Art.Mehrwertig() && len(v.Werte) > 1 {
		return NewFieldError("werte", SchluesselValidierung,
			"Für %q gilt genau eine Vorgabe, hier stehen %d.",
			string(v.Art), len(v.Werte))
	}

	return v.pruefeWerte()
}

// pruefeWerte weist leere, gerandete und doppelte Werte ab.
func (v Profilvorgabe) pruefeWerte() error {
	gesehen := make(map[string]bool, len(v.Werte))

	for _, wert := range v.Werte {
		if strings.TrimSpace(wert) == "" {
			return NewFieldError("werte", SchluesselValidierung,
				"Eine Vorgabe ohne Wert ist keine — wer nichts vorgeben will, "+
					"gibt eine leere Liste.")
		}

		// **Abgewiesen und nicht beschnitten** — dieselbe Regel wie beim
		// Anzeigenamen: zwei Vorgaben mit demselben sichtbaren Wert ließen
		// sich nicht unterscheiden.
		if wert != strings.TrimSpace(wert) {
			return NewFieldError("werte", SchluesselValidierung,
				"Der Wert %q trägt Leerzeichen am Rand.", wert)
		}

		if gesehen[wert] {
			return NewFieldError("werte", SchluesselValidierung,
				"Der Wert %q steht zweimal darin.", wert)
		}

		gesehen[wert] = true
	}

	return nil
}
