package domain

// **Die portable Referenz und ihre Auflösung** — Werkstück A2 von Phase 5.3
// (#1441), 0086.
//
// # Was einen Bestand verlässt
//
// Eine Ausfuhr trägt **nie** die lokale Zeilennummer eines Drucks: in einem
// anders nummerierten Bestand träfe sie eine zufällige Zeile. Sie trägt eine
// [Druckverweis]: Fassung, System, Semantik, Wert — und Name, Setcode, Nummer,
// Sprache als **Rückfall** für ein Ziel, das das System nicht kennt.
//
// # Wie sie ankommt
//
// [LoeseAuf] sucht unter Kandidaten zuerst die Kennung. Trifft sie genau
// einen Druck, ist das Ergebnis eindeutig; trifft sie mehrere, ist das ein
// **Konflikt** — auch bei einem Konzept, das per Definition mehrere trifft —
// und kein Automerge. Trifft sie keinen, entscheidet der Rückfall, aber nur
// **vollständig**: Name, Setcode, Nummer und Sprache zusammen. Der Name allein
// gibt es in dreißig Sets.

import "strings"

// Verweisfassung ist die Fassung der portablen Referenz.
//
// **Eine Zahl, die steigt, wenn sich die Gestalt ändert** — wie beim
// Einstellungsaustausch (0079). Ein Ziel, das eine höhere Fassung liest,
// weist ab und nennt beide Zahlen.
const Verweisfassung = 1

// Druckverweis ist die portable Referenz auf einen Druck.
type Druckverweis struct {
	Fassung int

	System   Kennungssystem
	Semantik Semantik
	Wert     string

	Name    string
	Setcode string
	Nummer  string
	Sprache string
}

// AlsVerweis gibt die portable Referenz eines Drucks.
//
// **Die eindeutige Kennung zuerst** — Scryfall vor MTGJSON, weil Scryfall
// die verbreitetere ist; ohne eine bleibt der Rückfall.
func (d Druck) AlsVerweis() Druckverweis {
	verweis := Druckverweis{
		Fassung: Verweisfassung,
		Name:    d.Name, Setcode: d.Setcode, Nummer: d.Nummer, Sprache: d.Sprache,
	}

	for _, system := range []Kennungssystem{SystemScryfall, SystemMTGJSON} {
		if wert, gibts := d.Kennung(system, SemantikDruck); gibts {
			verweis.System, verweis.Semantik, verweis.Wert = system, SemantikDruck, wert

			break
		}
	}

	return verweis
}

// Validate prüft die Referenz: Fassung, und dass sie etwas nennt.
func (v Druckverweis) Validate() []*FieldError {
	var befunde []*FieldError

	if v.Fassung > Verweisfassung {
		befunde = append(befunde, NewFieldError("fassung", SchluesselValidierung,
			"Die Referenz hat Fassung %d, dieses Programm liest bis %d.", v.Fassung, Verweisfassung))
	}

	if strings.TrimSpace(v.Wert) == "" && strings.TrimSpace(v.Name) == "" {
		befunde = append(befunde, NewFieldError("wert", SchluesselPflichtfeld,
			"Eine Referenz nennt eine Kennung oder einen Namen."))
	}

	return befunde
}

// Kennung gibt die Kennung der Referenz — oder false, wenn sie keine trägt.
//
// **Exportiert, weil der Dienst sie braucht**: er fragt die Ablage nach den
// Kandidaten, bevor [LoeseAuf] entscheidet.
func (v Druckverweis) Kennung() (Druckkennung, bool) {
	if strings.TrimSpace(v.Wert) == "" || v.System == "" {
		return Druckkennung{}, false
	}

	return Druckkennung{System: v.System, Semantik: v.Semantik, Wert: v.Wert}, true
}

// Die drei Ausgänge einer Auflösung.
const (
	// AufloesungEindeutig — genau ein Druck.
	AufloesungEindeutig = "eindeutig"
	// AufloesungMehrdeutig — mehrere Drucke: ein Konflikt, der gezeigt wird.
	AufloesungMehrdeutig = "mehrdeutig"
	// AufloesungUnbekannt — keiner: die Identität bleibt vorläufig.
	AufloesungUnbekannt = "unbekannt"
)

// Aufloesung ist das Ergebnis von [LoeseAuf].
type Aufloesung struct {
	Art      string
	DruckIDs []int64
}

// LoeseAuf ordnet eine Referenz den Kandidaten zu — Kennung zuerst, dann
// der vollständige Rückfall.
func LoeseAuf(kandidaten []Druck, verweis Druckverweis) Aufloesung {
	if kennung, hatKennung := verweis.Kennung(); hatKennung {
		var treffer []int64

		for _, druck := range kandidaten {
			for _, eigene := range druck.Kennungen {
				if eigene.gleich(kennung) {
					treffer = append(treffer, druck.ID)

					break
				}
			}
		}

		if ergebnis, entschieden := urteil(treffer); entschieden {
			return ergebnis
		}
	}

	return urteilOderUnbekannt(rueckfall(kandidaten, verweis))
}

// rueckfall trifft über Name, Setcode, Nummer und Sprache — alle vier.
func rueckfall(kandidaten []Druck, verweis Druckverweis) []int64 {
	if strings.TrimSpace(verweis.Name) == "" || strings.TrimSpace(verweis.Setcode) == "" ||
		strings.TrimSpace(verweis.Nummer) == "" {
		return nil
	}

	var treffer []int64

	for _, druck := range kandidaten {
		if strings.EqualFold(strings.TrimSpace(druck.Name), strings.TrimSpace(verweis.Name)) &&
			strings.EqualFold(strings.TrimSpace(druck.Setcode), strings.TrimSpace(verweis.Setcode)) &&
			strings.EqualFold(strings.TrimSpace(druck.Nummer), strings.TrimSpace(verweis.Nummer)) &&
			strings.EqualFold(strings.TrimSpace(druck.Sprache), strings.TrimSpace(verweis.Sprache)) {
			treffer = append(treffer, druck.ID)
		}
	}

	return treffer
}

// urteil macht aus Treffern ein Ergebnis — oder sagt, dass es keines gibt.
func urteil(treffer []int64) (Aufloesung, bool) {
	switch len(treffer) {
	case 0:
		return Aufloesung{}, false
	case 1:
		return Aufloesung{Art: AufloesungEindeutig, DruckIDs: treffer}, true
	default:
		return Aufloesung{Art: AufloesungMehrdeutig, DruckIDs: treffer}, true
	}
}

func urteilOderUnbekannt(treffer []int64) Aufloesung {
	if ergebnis, entschieden := urteil(treffer); entschieden {
		return ergebnis
	}

	return Aufloesung{Art: AufloesungUnbekannt}
}
