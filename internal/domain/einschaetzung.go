package domain

// **Eine KI-Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Der erste Aufrufer von [PruefeBelege]
//
// **A16 hat die Prüfung gebaut und niemand hat sie gerufen.** „Modell darf
// nur diese IDs referenzieren; unbekannte IDs/erfundene Regelnummern/URLs
// ergeben Validierungsbefund" stand als Zusage da, mit Tests und voller
// Abdeckung — und ohne Weg dorthin. Das ist die erste Dimension aus §10, und
// dieses Werkstück ist der Weg.
//
// # Was eine Einschätzung ist und was sie nicht ist
//
// **Sie ist die Meinung eines Modells über eine Frage, mit Belegen daneben.**
// Sie ist **keine amtliche Regelauskunft**, und das steht am Typ und nicht in
// einer Oberfläche: jede der vier Flächen muss es zeigen, und eine Zusage,
// die nur an einer Stelle steht, fehlt beim nächsten Umbau an dreien (§16).
//
// # Warum eine unbelegte Antwort gezeigt und nicht verworfen wird
//
// **Weil ein Satz mit einer falschen Nummer nicht falsch sein muss** — und
// weil ein Programm, das die Nummer stillschweigend striche, den Rest als
// geprüft aussehen ließe
// ([0107](../../docs/entscheidungen/0107-eine-einschaetzung-ist-keine-regelauskunft.md)).

import "fmt"

// Einschaetzungsart sagt, wonach gefragt wurde.
type Einschaetzungsart string

// Die zwei Arten.
const (
	// EinschaetzungRegel fragt, wie eine Karte nach den Regeln wirkt.
	EinschaetzungRegel Einschaetzungsart = "regel"

	// EinschaetzungBalance fragt, wie eine Karte im Vergleich dasteht.
	EinschaetzungBalance Einschaetzungsart = "balance"
)

var einschaetzungsarten = []Einschaetzungsart{
	EinschaetzungRegel, EinschaetzungBalance,
}

// Einschaetzungsarten gibt die zwei in ihrer Reihenfolge.
func Einschaetzungsarten() []Einschaetzungsart {
	return append([]Einschaetzungsart(nil), einschaetzungsarten...)
}

// Pruefe weist eine Art ab, die es nicht gibt.
func (a Einschaetzungsart) Pruefe() error {
	for _, bekannt := range einschaetzungsarten {
		if a == bekannt {
			return nil
		}
	}

	return fmt.Errorf(
		"%w: die Art %q gibt es nicht — es gibt %q und %q",
		ErrValidation, string(a), string(EinschaetzungRegel),
		string(EinschaetzungBalance))
}

// EinschaetzungsVorbehalt steht an jeder Einschätzung.
//
// # Warum als Konstante und nicht als Satz je Fläche
//
// **Weil er sonst an einer Fläche fehlte.** Vier Flächen, vier Gelegenheiten,
// ihn zu vergessen — und die eine, an der er fehlt, ist die, an der jemand
// die Antwort für eine Regelauskunft hält.
const EinschaetzungsVorbehalt = "Das ist die Einschätzung eines Modells und " +
	"keine amtliche Regelauskunft. Die Belege sagen, worauf sie sich " +
	"stützt — nicht, dass sie stimmt."

// KIEinschaetzung ist, was ein Modell zu einer Frage sagt.
type KIEinschaetzung struct {
	Art Einschaetzungsart

	// Satz ist die Einschätzung selbst — mit `[B1]`-Fußnoten darin.
	Satz string

	// Befunde sind die Mängel an ihrem Quellenanspruch.
	Befunde []Belegbefund

	// Vorbehalt ist [EinschaetzungsVorbehalt].
	Vorbehalt string
}

// PruefeEinschaetzung hält eine Modellantwort gegen die vergebenen Belege.
//
// **Sie verwirft nichts.** Was herauskommt, ist die Antwort **und** was an
// ihrem Quellenanspruch nicht stimmt; ob sie als belegt gezeigt werden darf,
// sagt [KIEinschaetzung.Belegt].
func PruefeEinschaetzung(
	art Einschaetzungsart, satz string, belege []Beleg,
) KIEinschaetzung {
	befunde := PruefeBelege(satz, belege)

	// **Eine Antwort ohne jede Fußnote ist kein Fehler des Modells.** „Dazu
	// steht in den Belegen nichts" ist eine richtige Antwort — sie ist nur
	// keine belegte, und genau das muss dastehen (Issue: „führt zu sichtbarer
	// Einschränkung oder fehlendem Beleg, niemals erfundenem Regelzitat").
	if !NenntEinenBeleg(satz) {
		befunde = append(befunde, Belegbefund{
			Art: BefundOhneBeleg, Stelle: "",
			Meldung: "Diese Einschätzung beruft sich auf keinen der " +
				"mitgegebenen Belege. Sie steht als Meinung da und nicht als " +
				"Auskunft aus einer Quelle.",
		})
	}

	return KIEinschaetzung{
		Art: art, Satz: satz, Befunde: befunde,
		Vorbehalt: EinschaetzungsVorbehalt,
	}
}

// Belegt sagt, ob sie als quellengestützt gezeigt werden darf.
func (e KIEinschaetzung) Belegt() bool { return len(e.Befunde) == 0 }
