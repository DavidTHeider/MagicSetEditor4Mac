// **Die Gestalt, in der Antworten dieses Servers ankommen**
// (#985).
//
// # Warum es diese Typen gibt
//
// Bis zum 2026-08-23 gaben drei Werkzeuge die Entitäten aus `internal/domain`
// **unmittelbar** heraus. MCP ist nach
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// Regel 5 eine Fläche nach **außen**: dort steht ein Gegenüber, das nicht
// mitgeliefert wird und eine andere Fassung haben kann.
//
// **Ein `Origin`, das jemand in `Quelle` umbenennt, änderte damit
// stillschweigend den Vertrag** — und kein Test hätte es gesehen. Aufgefallen
// ist es beim Schreiben der Flächenabnahme (#886).
//
// # Warum die Namen bleiben, wie sie waren
//
// Die Kommandozeile nennt dieselben Angaben `id`, `name`, `spiel`, `stil`.
// Sie hier anzugleichen wäre nach 0033 Regel 7 eine **brechende** Änderung und
// bräuchte eine Hauptfassung — für den Gewinn, dass zwei Flächen gleich
// aussähen. **Das ist kein Grund, ein Gegenüber zu brechen.**
//
// Die Einordnung nach Regel 6 lautet deshalb: **verträglich.** Für einen
// Klienten ändert sich nichts; geändert hat sich die **Herkunft** der Namen.
// Sie stehen jetzt hier und bewegen sich nicht mehr mit der Fachlichkeit.
//
// Ob die Flächen eines Tages dieselben Namen führen, ist eine Frage für die
// nächste Hauptfassung — und dann eine, die man einordnen kann.

package mcp

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// setzeile ist ein Set, wie dieser Server es herausgibt.
//
// **Die Marken sind die Zusage.** Ohne sie nähme `encoding/json` die
// Go-Feldnamen — und die stünden dann wieder an einer Entität.
type setzeile struct {
	ID     int64  `json:"ID"`
	Name   string `json:"Name"`
	Game   string `json:"Game"`
	Style  string `json:"Style"`
	Origin string `json:"Origin"`
	Cards  int    `json:"Cards"`
}

// kartenzeile ist eine Karte, wie dieser Server sie herausgibt.
type kartenzeile struct {
	ID     int64       `json:"ID"`
	SetID  int64       `json:"SetID"`
	Number int         `json:"Number"`
	Notes  string      `json:"Notes"`
	Values []wertzeile `json:"Values"`
}

// wertzeile ist ein Feldwert.
type wertzeile struct {
	Field string `json:"Field"`
	Text  string `json:"Text"`
}

// treffer ist ein Suchtreffer.
type treffer struct {
	CardID int64  `json:"CardID"`
	SetID  int64  `json:"SetID"`
	Number int    `json:"Number"`
	Field  string `json:"Field"`
	Value  string `json:"Value"`
}

// suchantwort ist das Ergebnis einer Suche.
type suchantwort struct {
	Hits      []treffer `json:"Hits"`
	Truncated bool      `json:"Truncated"`
}

// alsSetzeilen bringt Setköpfe in die zugesagte Gestalt.
func alsSetzeilen(koepfe []domain.SetHead) []setzeile {
	zeilen := make([]setzeile, 0, len(koepfe))
	for _, kopf := range koepfe {
		zeilen = append(zeilen, setzeile{
			ID: kopf.ID, Name: kopf.Name, Game: kopf.Game,
			Style: kopf.Style, Origin: kopf.Origin, Cards: kopf.Cards,
		})
	}

	return zeilen
}

// alsKartenzeilen bringt Kartenzeilen in die zugesagte Gestalt.
func alsKartenzeilen(karten []*domain.CardRow) []kartenzeile {
	zeilen := make([]kartenzeile, 0, len(karten))

	for _, karte := range karten {
		werte := make([]wertzeile, 0, len(karte.Values))
		for _, wert := range karte.Values {
			werte = append(werte, wertzeile{Field: wert.Field, Text: wert.Text})
		}

		zeilen = append(zeilen, kartenzeile{
			ID: karte.ID, SetID: karte.SetID, Number: karte.Number,
			Notes: karte.Notizen(), Values: werte,
		})
	}

	return zeilen
}

// alsSuchantwort bringt ein Suchergebnis in die zugesagte Gestalt.
func alsSuchantwort(ergebnis domain.SearchResult) suchantwort {
	hits := make([]treffer, 0, len(ergebnis.Hits))
	for _, hit := range ergebnis.Hits {
		hits = append(hits, treffer{
			CardID: hit.CardID, SetID: hit.SetID, Number: hit.Number,
			Field: hit.Field, Value: hit.Value,
		})
	}

	return suchantwort{Hits: hits, Truncated: ergebnis.Truncated}
}
