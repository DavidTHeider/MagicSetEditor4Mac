package domain

// **Ein Schritt gilt der Revision, die ein Mensch sah** —
// #2496,
// Rest aus Vermerk 0200, Punkt 12.
//
// Das Fenster liest vor einem Zug den Stand nach und geht nichts, wenn er
// nicht mehr der gesehene ist. Zwischen diesem Lesen und dem Schritt lag
// trotzdem ein Ruf: kam darin ein Gegenangebot an, galt eine Zustimmung
// dessen Revision, die niemand gesehen hatte. **Der Schritt nennt deshalb
// den Abdruck, den er meint**, und der Dienst prüft ihn gegen den Stand, auf
// den er ihn anwendet.
//
// # Was er nicht prüft
//
// **Eine Stimme oder einen Zustand**: der Abdruck geht über die
// Unterschriftsform einer Revision, und die ändert sich mit dem Inhalt, nicht
// mit einer Zustimmung. Entschieden wird über den Inhalt — eine Stimme des
// Gegenübers zur selben Revision verschiebt nichts daran. Das Fenster
// vergleicht vor dem Zug mehr (`handelszug.ts`, `derselbe`).

// PruefeGesehen sagt, ob dieser Stand die Revision ist, die `gesehen` nennt.
// **Ohne Angabe wird nicht gefragt** — Kommandozeile und Dienst nennen sie,
// wenn sie wollen (§16: erweiternd).
func (a Handelsangebot) PruefeGesehen(gesehen string) error {
	if gesehen == "" || gesehen == a.Abdruck() {
		return nil
	}

	return MitGrund(ErrUeberholt, nil,
		"der Handel hat sich geändert, seit er angesehen wurde — Revision %d ist eine andere",
		a.Revision)
}
