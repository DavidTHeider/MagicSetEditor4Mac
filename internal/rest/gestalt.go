// **Die Gestalt, in der Antworten dieses Dienstes ankommen**
// (#997).
//
// # Warum eigene Typen und nicht die Entitäten
//
// Derselbe Grund wie bei MCP (`internal/mcp/gestalt.go`): REST ist nach
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// Regel 5 eine Fläche nach **außen**. Gäbe dieses Paket `domain.SetHead`
// heraus, änderte ein Umbenennen dort stillschweigend den Vertrag hier — und
// kein Test sähe es.
//
// # Warum die Namen die der Kommandozeile sind
//
// `id`, `name`, `spiel`, `stil`, `werte` — genau wie `mse --json`. Ein Skript,
// das beide Flächen benutzt, liest damit **eine** Gestalt.
//
// **MCP hat andere**, und das bleibt so: dort sind sie seit 0.1.0 im Umlauf,
// und sie anzugleichen wäre nach 0033 Regel 7 eine brechende Änderung für den
// Gewinn, dass zwei Flächen gleich aussähen. Hier ist nichts im Umlauf, also
// wird die Wahl getroffen, solange sie frei ist.
//
// **Auch die Kennungen sind Zeichenketten**, wie auf der Kommandozeile.
// Einheitlichkeit schlägt hier die Gewohnheit, in REST Zahlen zu schicken:
// zwei Flächen, die dieselbe Kennung einmal als `1` und einmal als `"1"`
// geben, kosten jeden Aufrufer eine Fallunterscheidung.

package rest

import (
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// setzeile ist ein Set, wie dieser Dienst es herausgibt.
type setzeile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Spiel string `json:"spiel"`
	Stil  string `json:"stil"`
}

// kartenzeile ist eine Karte, wie dieser Dienst sie herausgibt.
type kartenzeile struct {
	ID     string            `json:"id"`
	Nummer int               `json:"nummer"`
	Name   string            `json:"name"`
	Werte  map[string]string `json:"werte"`
}

// alsSetzeilen übersetzt die Entitäten auf die Gestalt dieser Fläche.
//
// **Die leere Liste bleibt eine leere Liste.** `make(…, 0, …)` und nicht
// `var`: ein `nil`-Slice würde zu `null`, und ein Aufrufer, der über die
// Antwort schleift, bekäme einen Fehler statt keiner Runde. Dieselbe Falle
// wie das `omitempty` in AGENTS.md §3.
func alsSetzeilen(koepfe []domain.SetHead) []setzeile {
	zeilen := make([]setzeile, 0, len(koepfe))

	for _, kopf := range koepfe {
		zeilen = append(zeilen, setzeile{
			ID:    strconv.FormatInt(kopf.ID, 10),
			Name:  kopf.Name,
			Spiel: kopf.Game,
			Stil:  kopf.Style,
		})
	}

	return zeilen
}

// alsKartenzeilen übersetzt die Kartenzeilen.
func alsKartenzeilen(karten []*domain.CardRow) []kartenzeile {
	zeilen := make([]kartenzeile, 0, len(karten))

	for _, karte := range karten {
		werte := make(map[string]string, len(karte.Values))
		for _, wert := range karte.Values {
			werte[wert.Field] = wert.Text
		}

		zeilen = append(zeilen, kartenzeile{
			ID:     strconv.FormatInt(karte.ID, 10),
			Nummer: karte.Number,
			Name:   werte["name"],
			Werte:  werte,
		})
	}

	return zeilen
}
