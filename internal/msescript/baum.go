package msescript

import "strings"

// Knoten ist ein Ausdruck. Jeder trägt die Stelle, an der er beginnt.
//
// Ein Verbund statt einer Schnittstelle je Form: der Auswerter schaltet ohnehin
// über die Art, und eine Schnittstelle mit zwölf Trägertypen verteilte die
// Fallunterscheidung auf zwölf Dateien, ohne eine Entscheidung zu ersparen.
type Knoten struct {
	Form Form

	// Text trägt, was die Form braucht: den Namen eines Bezeichners, den Inhalt
	// einer Zeichenkette, das Zeichen eines Operators.
	Text string

	// Kinder in der Reihenfolge, in der die Form sie erwartet.
	Kinder []*Knoten

	// Namen benennt die Kinder eines Aufrufs, soweit sie benannt sind. Ein
	// leerer Eintrag heißt: an dieser Stelle stand ein Stellungsargument.
	Namen []string

	Zeile  int
	Spalte int
}

// Form benennt die Art eines Knotens.
type Form string

const (
	FormZahl         Form = "Zahl"
	FormZeichenkette Form = "Zeichenkette"
	FormName         Form = "Name"
	FormNichts       Form = "Nichts"
	FormZuweisung    Form = "Zuweisung"
	FormBedingung    Form = "Bedingung"
	FormFall         Form = "Fall"
	FormBinaer       Form = "Binär"
	FormUnaer        Form = "Unär"
	FormZugriff      Form = "Zugriff"
	FormAufruf       Form = "Aufruf"
	FormBindung      Form = "Bindung"
	FormKlammer      Form = "Klammer"
	FormFolge        Form = "Folge"
	FormAbbildung    Form = "Abbildung"
	FormIndex        Form = "Index"
	FormSchleife     Form = "Schleife"
	FormZaehlung     Form = "Zählung"
)

func (k *Knoten) String() string {
	if len(k.Kinder) == 0 {
		if k.Text == "" {
			return string(k.Form)
		}

		return string(k.Form) + "(" + k.Text + ")"
	}

	teile := make([]string, 0, len(k.Kinder))
	for _, kind := range k.Kinder {
		teile = append(teile, kind.String())
	}

	kopf := string(k.Form)
	if k.Text != "" {
		kopf += " " + k.Text
	}

	return "(" + kopf + " " + strings.Join(teile, " ") + ")"
}
