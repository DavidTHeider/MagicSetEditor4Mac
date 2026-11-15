// **Die empfohlene Rundenzahl** — Befund 2 des Zweitblicks über Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Woher die Zahlen kommen
//
// **Appendix E der *Magic Tournament Rules***, Fassung vom 2026-02-27,
// abgerufen am 2026-09-01 von
// <https://media.wizards.com/ContentResources/WPN/MTG_MTR_2026_Feb27_EN.pdf>.
//
// **Ein Vorschlag und keine Vorschrift.** Wörtlich dort: die Staffel „is
// required for Premier tournaments … It may be used at the Tournament
// Organizer's discretion for non-Premier tournaments." Ein Programm, das sie
// erzwänge, behauptete eine Regel, die es nicht gibt.

package domain

// staffel ist Appendix E — obere Grenze der Teilnehmerzahl und ihre Runden.
//
// **Aufsteigend, und die erste passende gilt.** Eine Abbildung wäre hier die
// falsche Form: die Tabelle nennt Bereiche, keine Werte.
var staffel = []struct{ bis, runden int }{
	// **Bis acht steht dort keine Swiss-Runde.**
	//
	// Die Spalte „Swiss Rounds" lautet für vier Spieler wörtlich „2
	// Single-Elimination Rounds (No Swiss)" und für fünf bis acht „3
	// Single-Elimination Rounds (No Swiss)". Das sind K.-o.-Runden und
	// ausdrücklich **kein** Swiss; die Playoff-Spalte sagt dazu „None (Run
	// Single Elimination)".
	//
	// **Dieses Programm baut kein K.-o.-System** (Vermerk 0062, Verworfen).
	// Für bis zu acht Teilnehmer hat Appendix E deshalb **keine** Empfehlung,
	// die hierher passt — und eine Zahl aus ihrer Bedeutung zu lösen wäre
	// schlimmer als keine.
	//
	// **Der erste Anlauf las die 3 als Swiss-Runden**, weil sie in der
	// Spalte „Swiss Rounds" steht (Befund 4 des Nachlaufs über Phase 4.7).
	// Die Zahl stand richtig und meinte etwas anderes.
	{bis: 8, runden: 0},

	// **Neun bis sechzehn nennt zwei Zahlen** — vier „if Limited Format with
	// Booster Draft in Playoff", fünf „All Other Formats". Dieses Programm
	// kennt kein Playoff und nimmt deshalb die zweite.
	{bis: 16, runden: 5},

	{bis: 32, runden: 5},
	{bis: 64, runden: 6},
	{bis: 128, runden: 7},
	{bis: 226, runden: 8},
	{bis: 409, runden: 9},
}

// hoechsteRunden ist, was Appendix E über 409 Teilnehmern nennt.
const hoechsteRunden = 10

// EmpfohleneRunden gibt die Rundenzahl aus Appendix E — **null heißt: keine**.
func EmpfohleneRunden(spieler int) int {
	for _, stufe := range staffel {
		if spieler <= stufe.bis {
			return stufe.runden
		}
	}

	return hoechsteRunden
}

// RundenvorschlagFuer gibt die Empfehlung für ein Turnier.
//
// **Vor der ersten Runde zählen nur die, die spielen; danach alle.**
//
// Das steht so **nicht** in Appendix E — sie spricht von „Players" und kennt
// den Fall nicht, weil die Zahl dort vor Beginn feststeht. MTR §10.2 hält
// dagegen fest, dass die **angekündigte** Rundenzahl steht. Beides zusammen
// ergibt diese Regel: solange niemand gespielt hat, ist die Empfehlung ein
// Vorschlag und folgt der Anmeldung; sobald eine Runde liegt, ist sie
// angekündigt und bleibt.
//
// **Der erste Anlauf rechnete auch mitten im Turnier neu**, und die Anzeige
// sprang, sobald jemand ausstieg (Befund 4 des Nachlaufs über Phase 4.7).
func RundenvorschlagFuer(turnier Turnier) int {
	// **Nach dem Beginn zählt, wer angemeldet war.** Wer aussteigt, ändert
	// die angekündigte Rundenzahl nicht.
	if len(turnier.Runden) > 0 {
		return EmpfohleneRunden(len(turnier.Teilnehmer))
	}

	var spielen int

	for _, teilnehmer := range turnier.Teilnehmer {
		if !teilnehmer.Ausgestiegen {
			spielen++
		}
	}

	if spielen == 0 {
		return 0
	}

	return EmpfohleneRunden(spielen)
}
