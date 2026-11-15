package domain

// **Was ein Wunsch über sich sagt** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Zwei Angaben und nicht eine
//
// **Weil sie zwei verschiedene Fragen beantworten.** Der **Rang** sagt, wie
// dringend jemand die Karte will — er ordnet eine Liste. Die **Art** sagt, ob
// genau dieser Druck gemeint ist oder die Karte; sie entscheidet, was
// überhaupt ein Treffer ist. Wer beides in eine Angabe legte, könnte „egal
// welcher Druck, aber dringend" nicht ausdrücken.
//
// # Warum die Art ausdrücklich dasteht
//
// **Weil Namensgleichheit keine Absicht ist.** Zwei Zeilen mit demselben
// Kartennamen sind nicht dasselbe Angebot: ein Alpha-Druck und ein
// Neudruck tragen denselben Namen und nicht denselben Wert. Wer aus der
// Namensgleichheit einen Treffer machte, böte jemandem etwas an, das er nie
// gewollt hat — und es fiele erst am Tisch auf.
//
// # Und warum die leere Angabe die engere ist
//
// **Weil jede heutige Wunschzeile einen Druck nennt.** Aus „ohne Angabe" ein
// „jeder Druck" zu machen weitete beim Migrieren jeden bestehenden Wunsch,
// und zwar still. Die Vorgabe ist deshalb `genau`; wer mehr will, sagt es.

// Wunschrang sagt, wie dringend ein Wunsch ist.
type Wunschrang string

// Die vier Ränge.
const (
	RangNiedrig  Wunschrang = "niedrig"
	RangNormal   Wunschrang = "normal"
	RangHoch     Wunschrang = "hoch"
	RangSehrHoch Wunschrang = "sehrhoch"
)

// Wunschraenge gibt die vier in **ihrer** Ordnung — von niedrig nach hoch.
//
// **Eine Aufzählung und keine Menge.** An ihr hängt die Reihenfolge in jeder
// Fläche; eine Map gäbe sie bei jedem Lauf anders zurück.
func Wunschraenge() []Wunschrang {
	return []Wunschrang{RangNiedrig, RangNormal, RangHoch, RangSehrHoch}
}

// Gewicht gibt die Ordnungszahl — **je größer, desto dringender**.
//
// **Eine Zahl und kein Textvergleich.** „hoch" ist alphabetisch kleiner als
// „normal", und eine Sortierung über die Zeichenkette drehte die Reihenfolge
// still um. Ein unbekannter Rang wiegt wie `normal`: er darf nichts nach vorn
// ziehen, was niemand nach vorn gestellt hat.
func (r Wunschrang) Gewicht() int {
	switch r {
	case RangNiedrig:
		return 0
	case RangHoch:
		return 2
	case RangSehrHoch:
		return 3
	case RangNormal:
		return 1
	}

	return 1
}

// IstWunschrang sagt, ob es den Rang gibt.
func IstWunschrang(rang Wunschrang) bool {
	for _, kandidat := range Wunschraenge() {
		if kandidat == rang {
			return true
		}
	}

	return false
}

// Wunschart sagt, woran ein Wunsch gebunden ist.
type Wunschart string

// Die zwei Arten.
const (
	// WunschGenau meint **diesen** Druck — Set, Nummer, Ausführung.
	WunschGenau Wunschart = "genau"

	// WunschJederDruck meint die Karte, gleich in welchem Druck.
	WunschJederDruck Wunschart = "jederdruck"
)

// Wunscharten gibt die zwei — die engere zuerst.
func Wunscharten() []Wunschart { return []Wunschart{WunschGenau, WunschJederDruck} }

// IstWunschart sagt, ob es die Art gibt.
func IstWunschart(art Wunschart) bool {
	for _, kandidat := range Wunscharten() {
		if kandidat == art {
			return true
		}
	}

	return false
}

// Dringlichkeit gibt den Rang dieses Wunsches — **leer heißt `normal`**.
//
// **Und an einem Besitz heißt alles `normal`.** Wer eine gewünschte Karte
// erwirbt, dreht `Wunsch` um; ein stehengebliebener Rang bestimmte danach
// eine Reihenfolge, die niemand gewollt hat. Derselbe Satz wie bei
// [Exemplar.Ausfuehrung]: die Regel gehört an **eine** Stelle.
func (e Exemplar) Dringlichkeit() Wunschrang {
	if e.Besitzt() || e.Wunschrang == "" {
		return RangNormal
	}

	return e.Wunschrang
}

// Druckbindung gibt die Art dieses Wunsches — **leer heißt `genau`**.
//
// **Und an einem Besitz ebenfalls `genau`**, aus demselben Grund wie bei
// [Exemplar.Dringlichkeit]: eine Angabe, die nichts bedeutet, darf nichts
// bewirken.
func (e Exemplar) Druckbindung() Wunschart {
	if e.Besitzt() || e.Wunschart == "" {
		return WunschGenau
	}

	return e.Wunschart
}
