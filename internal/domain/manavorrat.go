package domain

// **Der Manavorrat** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Warum er nicht dieselbe Sache ist wie eine Farbe
//
// **Weil farblos eine Manaart ist und keine Farbe** — CR 106.1b nennt
// **sechs** Arten, CR 106.1a **fünf** Farben. [Farbe] beantwortet, welche
// Farbe eine Karte hat (CR 202.2); hier geht es darum, was in einem Vorrat
// liegt und womit sich etwas bezahlen lässt. Wer beides in einen Typ legte,
// müsste bei jeder Stelle nachdenken, welche der beiden Fragen gerade gemeint
// ist.
//
// # Und warum Schnee keine Art ist
//
// **Weil CR 107.4h es ausdrücklich sagt**: `{S}` ist weder Farbe noch
// Manaart. Schneemana ist Mana aus einer Schneequelle — eine Eigenschaft der
// **Herkunft**, nicht des Manas. Dieses Programm führt die Herkunft nicht,
// und deshalb steht `{S}` unter den Symbolen, die es nicht zahlen kann
// ([Unbezahlbar]).

// Manaart ist eine der sechs Arten von Mana — CR 106.1b (Fassung 2026-08-19).
type Manaart string

const (
	ManaWeiss   Manaart = "W"
	ManaBlau    Manaart = "U"
	ManaSchwarz Manaart = "B"
	ManaRot     Manaart = "R"
	ManaGruen   Manaart = "G"

	// ManaFarblos ist farbloses Mana — CR 107.4c.
	//
	// **Keine Farbe, aber eine Art**: `{C}` verlangt genau dieses Mana, und
	// ein weißes zahlt es nicht.
	ManaFarblos Manaart = "C"
)

// alleManaarten steht als Aufzählung und nicht als Ableitung aus [Farben].
//
// **Weil farblos dort nicht vorkommt** und eine Ableitung mit einem Anhängsel
// so aussähe, als wäre sie eine — genau die Vermischung, die der Kopf dieser
// Datei verhindert.
var alleManaarten = []Manaart{
	ManaWeiss, ManaBlau, ManaSchwarz, ManaRot, ManaGruen, ManaFarblos,
}

// Manaarten gibt alle sechs — **in der Reihenfolge des Regelwerks**, farblos
// zuletzt (CR 106.1b).
func Manaarten() []Manaart {
	arten := make([]Manaart, len(alleManaarten))
	copy(arten, alleManaarten)

	return arten
}

// ManaartAus deutet einen Buchstaben — `false`, wenn er keine Art benennt.
func ManaartAus(zeichen string) (Manaart, bool) {
	for _, art := range alleManaarten {
		if string(art) == zeichen {
			return art, true
		}
	}

	return "", false
}

// manaartZuFarbe gibt die Art zu einer Farbe.
//
// **Es ist eine Abbildung und keine Gleichsetzung**: jede Farbe hat eine Art,
// aber nicht jede Art hat eine Farbe.
func manaartZuFarbe(farbe Farbe) Manaart {
	return Manaart(farbe)
}

// Manavorrat ist das ungenutzte Mana eines Spielers — CR 106.4.
//
// # Warum eine Tabelle und keine sechs Felder
//
// **Weil jede Stelle sonst sechsmal dasteht.** Sechs Felder hießen sechs
// Zweige beim Zählen, sechs beim Bezahlen, sechs beim Leeren — und beim
// nächsten Symbol, das eine siebte Art einführte, sechs Stellen zum
// Nachziehen.
//
// # Und warum jede Methode eine Kopie gibt
//
// **Weil eine Tabelle ein Zeiger ist.** [Partie] ist ein Wert und wird
// kopiert; läge darin eine geteilte Tabelle, änderte eine Zustandsprüfung den
// Vorrat der Partie, aus der sie kam. Genau dieselbe Regel wie überall im
// Kern, nur hier leicht zu übersehen.
type Manavorrat map[Manaart]int

// Zahl gibt, wie viel von einer Art dasteht — null, wo nichts steht.
func (v Manavorrat) Zahl(art Manaart) int {
	return v[art]
}

// Gesamt gibt, wie viel Mana insgesamt dasteht.
func (v Manavorrat) Gesamt() int {
	var summe int
	for _, art := range alleManaarten {
		summe += v[art]
	}

	return summe
}

// Leer sagt, ob nichts mehr dasteht.
func (v Manavorrat) Leer() bool {
	return v.Gesamt() == 0
}

// Mehr legt Mana hinein und gibt den **neuen** Vorrat.
//
// **Eine negative Menge nimmt weg**, und unter null geht es nicht: ein Vorrat
// mit minus einem Mana wäre ein Zustand, den das Regelwerk nicht kennt.
func (v Manavorrat) Mehr(art Manaart, um int) Manavorrat {
	danach := v.kopie()

	danach[art] += um
	if danach[art] <= 0 {
		delete(danach, art)
	}

	return danach
}

// Kopie gibt eine eigene Tabelle.
//
// **Exportiert, weil `internal/service` sie braucht**: wer eine Partie
// kopiert, muss auch die Tabellen darin kopieren, und das kann er von außen
// nur mit einer Methode. **Ein leerer Vorrat gibt eine leere Tabelle** und
// nicht `nil` — ein Aufrufer soll nicht zwei Fälle unterscheiden müssen.
func (v Manavorrat) Kopie() Manavorrat {
	return v.kopie()
}

// kopie gibt eine eigene Tabelle.
func (v Manavorrat) kopie() Manavorrat {
	danach := make(Manavorrat, len(v))
	for art, zahl := range v {
		if zahl > 0 {
			danach[art] = zahl
		}
	}

	return danach
}
