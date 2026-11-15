// **Die Angaben eines Kontos** — Werkstück 8 von Phase 4.12
// (#1351),
// nach der Kommentierung vom 2026-09-05.
//
// # Fünf Namensteile und ein Anzeigename
//
// **Sie sind nicht dasselbe.** Der Anzeigename ist eindeutig und das Einzige,
// was ein Nutzer haben **muss**; die Teile sind freiwillig und beschreiben
// einen Menschen. Die Anwendung leitet den einen **nicht** aus den anderen ab
// — das hieße, ihn beim ersten Tippfehler zu ändern.
//
// # Warum die Anrede freier Text ist
//
// **Weil jede Liste jemanden auslässt.** „Herr/Frau" ist keine Aufzählung
// aller Menschen, und eine Anwendung, die jemanden zwingt, sich für einen
// falschen Eintrag zu entscheiden, hat ihn falsch angeredet — genau das, was
// das Feld verhindern soll. Geprüft wird die **Länge** und der Rand, nicht der
// Inhalt.

package domain

import "strings"

// Die Grenzen der Angaben.
//
// **Sie sind Riegel und keine Aussagen darüber, wie lang ein Name ist.** Ein
// Feld ohne Grenze nimmt eine ganze Datei entgegen; was hier steht, ist die
// Stelle, an der eine Eingabe aufhört, ein Name zu sein.
const (
	// Hoechstnamensteil gilt für Anrede, Zusatz, Vor- und Nachnamen.
	Hoechstnamensteil = 100

	// Hoechstslogan ist die Zeile unter dem Namen.
	Hoechstslogan = 200

	// Hoechstkurzebio ist der Absatz, der auf eine Karte passt.
	Hoechstkurzebio = 500

	// Hoechstlangebio ist der Text, den niemand mehr überfliegt.
	Hoechstlangebio = 5000
)

// pruefeAngabe weist eine zu lange Angabe ab — und eine mit Rändern.
//
// **In Zeichen und nicht in Byte**: sonst wäre ein Name mit Umlauten kürzer
// erlaubt als einer ohne, und niemand sähe, warum.
//
// **Leer ist erlaubt.** Jede dieser Angaben ist freiwillig; die Pflicht trägt
// allein der Anzeigename ([Nutzer.Validate]).
func pruefeAngabe(feld, wert string, hoechstens int) error {
	if wert == "" {
		return nil
	}

	if wert != strings.TrimSpace(wert) {
		return NewFieldError(feld, SchluesselValidierung,
			"Diese Angabe trägt keine Leerzeichen am Rand; %q geht nicht.", wert)
	}

	if len([]rune(wert)) > hoechstens {
		return NewFieldError(feld, SchluesselValidierung,
			"Diese Angabe hat höchstens %d Zeichen.", hoechstens)
	}

	return nil
}

// pruefeAngaben prüft die zehn freiwilligen Angaben eines Nutzers.
//
// **Eine Tabelle und keine zehn `if`**: wer eine Angabe hinzufügt, trägt eine
// Zeile ein — und die Meldung nennt das Feld, weil die Tabelle es trägt (§6).
func (n Nutzer) pruefeAngaben() error {
	for _, zeile := range []struct {
		feld       string
		wert       string
		hoechstens int
	}{
		{"anrede", n.Anrede, Hoechstnamensteil},
		{"namenszusatz", n.Namenszusatz, Hoechstnamensteil},
		{"vorname", n.Vorname, Hoechstnamensteil},
		{"weiterenamen", n.Weiterenamen, Hoechstnamensteil},
		{"nachname", n.Nachname, Hoechstnamensteil},
		{"slogan", n.Slogan, Hoechstslogan},
		{"kurzebio", n.Kurzebio, Hoechstkurzebio},
		{"langebio", n.Langebio, Hoechstlangebio},
	} {
		if fehler := pruefeAngabe(zeile.feld, zeile.wert, zeile.hoechstens); fehler != nil {
			return fehler
		}
	}

	return nil
}
