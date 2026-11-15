package app

import "context"

// **Am Tisch reden und zeigen, am Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Zwei Bindungen für beide Seiten
//
// **Das Fenster ruft dasselbe, gleich an welchem Ende es sitzt.** Ob dieser
// Rechner freigegeben hat oder beigetreten ist, weiß die Anwendung aus ihrem
// eigenen Stand; eine Oberfläche, die es wissen müsste, entschiede bei jedem
// Wort neu, welche Bindung sie nimmt — und irgendwann falsch (§16).
//
// # Und warum sie hier stehen und nicht bei den Beitrittsbindungen
//
// **Weil sie nicht zum Beitritt gehören.** Jene Datei beantwortet, wie dieses
// Programm sich an einen **fremden** Tisch setzt; diese beiden gelten auch am
// eigenen (§2: der Schnitt folgt der Frage).

// PartieSagen schickt einen Satz an den Tisch.
func (b *Bindungen) PartieSagen(ctx context.Context, text string) PartienebenbeiAntwort {
	return b.app(ctx).PartieSagen(b.ruf(ctx), text)
}

// PartieZeigen schickt einen Zeiger, einen Griff, ein Schweben oder einen Ping.
//
// **Vier Zahlen und kein Verbund.** Wails bildet einen Verbund als Abbildung
// ab, und dann stünde an dieser Naht `map[string]any` statt vier benannter
// Werte — die Stelle, an der ein Tippfehler im Feldnamen still durchginge.
func (b *Bindungen) PartieZeigen(
	ctx context.Context,
	art, objekt string, x, y float64,
) PartienebenbeiAntwort {
	return b.app(ctx).PartieZeigen(b.ruf(ctx), art, objekt, x, y)
}
