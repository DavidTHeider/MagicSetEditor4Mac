package domain

// Das **Schlüsselwort eines Sets**
// (#1062).
//
// # Warum es in der Domäne steht und nicht in `msepaket`
//
// `msepaket.Schluesselwort` ist, was in einer **Paketdatei** steht — gelesen
// und geschrieben. Dies hier ist, was der **Bestand** führt: es hat eine
// Kennung, es lässt sich anlegen und löschen, und es gehört zu einem Set.
//
// Zwei Typen für dieselben fünf Felder sind kein Versehen, sondern dieselbe
// Trennung wie zwischen [Card] und einem `card:`-Block. Wer sie
// zusammenlegte, hätte eine Kennung in einer Datei oder eine Zeilennummer im
// Bestand.

// Keyword ist ein Schlüsselwort, das ein **Set** mitbringt.
//
// Nachgelesen an `IMPLEMENT_REFLECTION(Keyword)`, `src/data/keyword.cpp`,
// Zeilen 86–93, Fassung f9d9356 vom 2021-07-06: mehr als diese fünf Felder
// steht nicht in der Datei.
type Keyword struct {
	// ID ist die Kennung im Bestand — 0 für eines, das es dort noch nicht gibt.
	ID int64

	// Word ist der Name, „only for human use" (`keyword.hpp:93`).
	Word string

	// Match ist der Text, auf den es passt, mit `<atom-param>art</atom-param>`
	// an den Stellen der Parameter.
	Match string

	// Reminder ist der Erinnerungstext — **ein Skript**, kein fester Text
	// (`StringScript reminder`, `keyword.hpp:95`).
	Reminder string

	// Rules ist eine Erläuterung für Menschen, die nicht auf der Karte
	// erscheint.
	Rules string

	// Mode ist der Name eines Modus des Spiels — `core`, `expert`, `custom`.
	Mode string

	// Reihenfolge ist die Stelle, an der dieses Schlüsselwort in der **Datei**
	// stand — und damit der Block, in den es zurückgehört.
	//
	// **Nicht die Stelle in der Liste.** Nach einer Löschung ist die Liste
	// kürzer als die Datei, und wer mitzählt, schreibt in den falschen Block:
	// blieb von Alpha und Beta nur Beta, bekam **Alphas** Block Betas Felder
	// — samt aller Zeilen, die Alpha trug und die dieses Programm nicht kennt.
	//
	// Der Kartenweg trägt dieselbe Zahl als `CardRow.Number` und aus
	// demselben Grund. Angestrichen vom Zweitblick durch Codex über Phase W.
	//
	// **Sie bleibt gültig, weil die Quelle bleibt**: `Setschreiber` liest die
	// Datei, aus der eingelesen wurde, und schreibt sie nie zurück in den
	// Bestand. Die Blöcke der Datei stehen deshalb dauerhaft so, wie sie
	// gezählt wurden.
	Reihenfolge int
}
