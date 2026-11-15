package service

// Welche Kartenfelder ein Wertskript **liest**.
//
// # Wozu das gezählt wird
//
// `card` ist in der Skriptumgebung eine **Kopie**: [kartenabbildung] baut die
// Einträge neu, sobald ein Farb- oder Bildfeld dabei ist — und das ist in jedem
// gesichteten Spielpaket der Fall. Wer nach einem Wertskript nicht neu einträgt,
// lässt das nächste den Wert lesen, den das Feld **vor** seinem Skript hatte.
//
// Nach jedem Wertskript neu einzutragen wäre richtig und teuer: `magic.mse-game`
// führt 249 Kartenfelder, 142 davon mit Skript, und jede Abbildung wandelt Farb-
// und Bildwerte um. Gelesen werden über `card.` nur wenige — je Karte entstehen
// daraus so viele Neubauten und nicht 142.
//
// **Was MSE stattdessen tut**: `ScriptManager` ordnet die Aktualisierungen nach
// ihren Abhängigkeiten und rechnet jede genau einmal (`src/data/set.cpp`,
// Fassung f9d9356 vom 2021-07-06). Den Graphen dafür erhebt MSE durch eine
// abstrakte Auswertung jedes Skripts (`src/script/dependency.cpp`).
//
// **Seit #2396 baut [leserJeFeld] aus derselben Suche einen** — für das
// Nachziehen beim Schreiben (`wertnachzug.go`). Er kennt nur, was `card.<feld>`
// im Skript selbst nennt, nicht was eine Funktion der Initskripte liest.

import "regexp"

// kartenzugriff findet `card.<name>` in einem Skript.
//
// **Nur der Punktzugriff.** `card["a" + b]` gibt es im Pack nicht, und wer ihn
// mitfangen wollte, müsste den Ausdruck auswerten — vor dem Lauf, der ihn
// auswertet. Was dadurch durchfällt, bleibt beim alten Verhalten: der Wert vor
// dem Skript.
var kartenzugriff = regexp.MustCompile(`\bcard\.([a-zA-Z_][a-zA-Z_0-9]*)`)

// gelesenevonKarte gibt die Feldnamen, die irgendein Wertskript über `card`
// liest.
//
// **Gezählt wird über die Skripte, nicht über die Felder**: ein Feld, das
// niemand liest, kostet nichts.
func gelesenevonKarte(skripte map[string]string) map[string]bool {
	gelesen := make(map[string]bool, len(skripte))

	for _, quelle := range skripte {
		for _, treffer := range kartenzugriff.FindAllStringSubmatch(quelle, -1) {
			gelesen[treffer[1]] = true
		}
	}

	return gelesen
}
