package domain

// **Wann der Beginn einer Partie läuft** — CR 103.8 (Fassung 2026-09-25),
// Befund 2 des Zweitblicks über Phase 6.9 und
// #1851.
//
// **Eine eigene Datei**, weil `eroeffnung.go` an der Zeilengrenze steht
// (§2). Der Schnitt folgt der Frage: dort steht, was die Schritte des
// Beginns *tun*, hier, *wann* er läuft.

// BeginnLaeuft sagt, ob der Beginn noch aussteht: ein Sitz, der mitspielt,
// hat seine Starthand noch nicht gezogen oder noch nicht behalten.
// **Solange nimmt der Sandkasten nur ihn an**, wo die Zugstruktur gilt — und
// dazu, was ihn nicht berührt: Aufgeben (CR 104.3a) und die Würfe, mit denen
// CR 103.1 den Startspieler bestimmen lässt. Der erste Zug kommt danach
// (CR 103.8).
//
// # Vom Aufbau an und nicht erst ab der ersten Starthand
//
// **Der Aufbau lässt den Beginn für jeden Sitz ausstehen, und von da an
// läuft er.** Zuerst galt er nur, wenn schon einer angefangen hatte — Befund
// 2 des Zweitblicks über Phase 6.9 war damit zur Hälfte behoben: eine
// Partie, in der niemand eine Starthand zog, spielte ohne Beginn aus der
// ungemischten Bibliothek. Der Grund war kein fachlicher, sondern die
// Vorrichtungen der Läufe über Sandkasten, Naht und Flächen: ihre Decks
// reichten für keine Starthand, und sie spielten gleich nach dem Anlegen
// (#1851). Sie gehen seither durch den Beginn wie jeder andere.
//
// **Eine Partie von vorher** trägt keinen Stand und beginnt nichts — dieselbe
// Lesart wie beim Aufbau.
//
// **Wer gegangen ist, zählt nicht mehr** — Befund 2 des Nachlaufs über
// Phase 6.9. Gibt zu dritt einer mitten im Beginn auf, spielen die zwei
// anderen weiter (CR 104.3a); sein ausstehender Beginn hielt sie sonst für
// immer auf.
func (p Partie) BeginnLaeuft() bool {
	for _, sitz := range p.Sitze {
		if p.istDraussen(sitz.Kennung) {
			continue
		}

		if sitz.Eroeffnung == EroeffnungOffen ||
			sitz.Eroeffnung == EroeffnungAusstehend {
			return true
		}
	}

	return false
}
