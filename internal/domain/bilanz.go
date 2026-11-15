// **Die Bilanz eines Teilnehmers** — Werkstück 2 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Woher jede Zahl hier kommt
//
// **Aus Appendix C der *Magic Tournament Rules***, Fassung vom 2026-02-27,
// abgerufen am 2026-09-01 von
// <https://media.wizards.com/ContentResources/WPN/MTG_MTR_2026_Feb27_EN.pdf>.
//
// Keine dieser Zahlen ist erschlossen. Wo etwas nicht dort steht, sagt der
// Kommentar es ausdrücklich.
//
// # Warum sie keine Gegner kennt
//
// Weil die beiden Gegnerquoten das **Mittel über andere Bilanzen** sind. Eine
// Bilanz, die ihre Gegner kennte, wäre ein Graph und keine Zahlenreihe — und
// die Rangliste, die ihn auswertet, steht in `rangliste.go`.

package domain

// Punkte je Ausgang — Appendix C, Abschnitte *Match Points* und *Game Points*.
//
// **Match und Spiel haben dieselben Zahlen**, und das ist kein Zufall:
// „Game points are similar to match points in that players earn 3 game points
// for each game they win and 1 point for each game that ends in a draw."
const (
	punkteSieg          = 3
	punkteUnentschieden = 1
	punkteNiederlage    = 0
)

// Quotenuntergrenze ist die Untergrenze beider Quoten — **0,33**.
//
// Wörtlich aus Appendix C: „If this number is lower than 0.33, use 0.33
// instead. The minimum match-win percentage of 0.33 limits the effect low
// performances have when calculating and comparing opponents' match-win
// percentage."
//
// **Sie ist exportiert, weil sie erklärt werden muss.** Eine Quote von 0,33
// bei einer Bilanz von 1–3–0 sieht wie ein Fehler aus, bis jemand diesen Satz
// liest.
const Quotenuntergrenze = 0.33

// Bilanz ist, was ein Teilnehmer in einem Turnier zusammengespielt hat.
//
// **Sie zählt, sie rechnet nicht.** Die Quoten sind abgeleitet, damit sie
// nicht veralten können — dieselbe Überlegung wie beim Turnier selbst
// (Vermerk 0062, Entscheidung 1: ein Turnier ist ein Verlauf, kein Zustand).
type Bilanz struct {
	// Siege, Niederlagen und Unentschieden zählen **Matches**, nicht Spiele.
	Siege         int
	Niederlagen   int
	Unentschieden int

	// Byes werden getrennt gezählt, obwohl sie als Sieg wirken.
	//
	// **Getrennt, weil sie zweimal etwas anderes tun**: für die eigene
	// Matchquote zählen sie mit („Players receiving byes are considered to
	// have won the match"), für die **Gegnerquoten** nicht („A player's byes
	// are ignored when computing their opponents' match-win and opponents'
	// game-win percentages"). Wer sie zu den Siegen addierte, könnte das
	// zweite nicht mehr.
	Byes int

	// spielsiege, spielniederlagen und spielunentschieden zählen **Spiele**.
	spielsiege         int
	spielniederlagen   int
	spielunentschieden int
}

// Spiele trägt den Ausgang **eines Matches** in Spielen ein.
func (b *Bilanz) Spiele(gewonnen, verloren, unentschieden int) {
	b.spielsiege += gewonnen
	b.spielniederlagen += verloren
	b.spielunentschieden += unentschieden

	switch {
	case gewonnen > verloren:
		b.Siege++
	case verloren > gewonnen:
		b.Niederlagen++
	default:
		b.Unentschieden++
	}
}

// Bye trägt eine Runde ohne Gegner ein.
//
// **Als 2–0 und nicht als „gewonnen"**: Appendix C sagt beides, und die
// sechs Spielpunkte kommen nur aus dem Spielstand.
func (b *Bilanz) Bye() {
	b.Byes++
	b.spielsiege += 2
}

// Matchpunkte sind 3 je Sieg, 1 je Unentschieden, 0 je Niederlage.
func (b Bilanz) Matchpunkte() int {
	return (b.Siege+b.Byes)*punkteSieg +
		b.Unentschieden*punkteUnentschieden +
		b.Niederlagen*punkteNiederlage
}

// Runden ist die Zahl der gespielten Runden — **Byes zählen mit**.
//
// Belegt durch das dritte Beispiel aus Appendix C: „3-2-0, including a
// first-round bye, then withdraws | 9 | 5 | 9/(5*3) = 0.60". Fünf Runden bei
// zwei gespielten Siegen, zwei Niederlagen und einem Bye.
func (b Bilanz) Runden() int {
	return b.Siege + b.Niederlagen + b.Unentschieden + b.Byes
}

// Spielpunkte sind 3 je gewonnenes Spiel, 1 je unentschiedenes.
//
// **Unfertige Spiele zählen als unentschieden, ungespielte als null** —
// Appendix C. Beides ist Sache des Aufrufers: was hier ankommt, ist gezählt.
func (b Bilanz) Spielpunkte() int {
	return b.spielsiege*punkteSieg +
		b.spielunentschieden*punkteUnentschieden +
		b.spielniederlagen*punkteNiederlage
}

// Spielzahl ist die Zahl der gespielten Spiele — **das Bye zählt zwei**.
func (b Bilanz) Spielzahl() int {
	return b.spielsiege + b.spielniederlagen + b.spielunentschieden
}

// Matchquote ist die *match-win percentage* aus Appendix C.
func (b Bilanz) Matchquote() float64 {
	return quote(b.Matchpunkte(), b.Runden())
}

// Spielquote ist die *game-win percentage* aus Appendix C.
func (b Bilanz) Spielquote() float64 {
	return quote(b.Spielpunkte(), b.Spielzahl())
}

// quote teilt Punkte durch das Mögliche und hält die Untergrenze.
//
// **Ohne gespielte Einheit gibt es keine Quote — und nicht 0,33.** Die MTR
// kennt den Fall nicht, weil dort jeder spielt; hier kann jemand angemeldet
// sein, bevor die erste Runde läuft. Die Untergrenze auf ein leeres Feld
// anzuwenden hieße, eine Aussage über jemanden zu machen, über den es keine
// gibt. **Das ist eine eigene Entscheidung** und steht so nicht in Appendix C.
func quote(punkte, moeglich int) float64 {
	if moeglich == 0 {
		return 0
	}

	gerechnet := float64(punkte) / float64(moeglich*punkteSieg)
	if gerechnet < Quotenuntergrenze {
		return Quotenuntergrenze
	}

	return gerechnet
}

// Gegnerquote mittelt eine Quote über die Bilanzen der Gegner.
//
// **Die Untergrenze steckt in `welche`**, nicht hier. Das ist der Unterschied,
// den Appendix C macht: sie gehört zur Definition der **einzelnen** Quote, und
// wer erst mittelt und dann begrenzt, bekommt eine andere Zahl. Zwei Gegner
// mit 0,25 und 0,75 ergeben so 0,54 und nicht 0,50.
//
// **Ohne Gegner gibt es keine Quote** — dieselbe eigene Entscheidung wie in
// [quote].
func Gegnerquote(gegner []Bilanz, welche func(Bilanz) float64) float64 {
	if len(gegner) == 0 {
		return 0
	}

	var summe float64

	for _, eine := range gegner {
		summe += welche(eine)
	}

	return summe / float64(len(gegner))
}
