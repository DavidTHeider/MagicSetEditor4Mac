package service

import (
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Die Kartenblöcke der Quelldatei gegen die Karten des Bestands.
//
// **Eine eigene Datei**, weil `beiwerk.go` an der Zeilengrenze steht (§2) — und
// weil der Schnitt einer Frage folgt: dort steht, was neben der Hauptdatei
// mitkommt, hier, welcher `card:`-Block zu welcher Karte gehört.

// angleichen bringt den Baum auf den Stand des Bestands.
//
// **Nur die Zeilen, die sich geändert haben.** Ein Knoten, dessen Wert derselbe
// ist, behält seine Rohzeile — und damit Einrückung, Abstände und Zeilenende.
// Wer stattdessen alles neu setzte, hätte die Datei angefasst, ohne dass jemand
// es angeordnet hat.
//
// # Zugeordnet wird über die Reihenfolge, nicht übers Abzählen
//
// `cards.reihenfolge` **ist** die Stelle, an der die Karte in der Datei steht:
// der Import vergibt sie beim Einlesen fortlaufend, und jeder `card:`-Block
// wird eine Karte — auch ein leerer
// ([`import.go`](import.go), „eine leere Karte bleibt eine Karte").
//
// Bis #621
// zählte diese Stelle stattdessen mit und paarte der Reihe nach. Das trug,
// solange es Blöcke und Karten gleich oft gab. Seit eine Karte gelöscht werden
// kann, gibt es sie das nicht mehr — und das Abzählen schrieb die Werte in die
// **falschen** Blöcke: nach dem Löschen der mittleren von drei Karten stand die
// dritte zweimal in der Datei und die zweite gar nicht.
func angleichen(baum *msepaket.Baum, karten []*domain.CardRow) {
	nachStelle := make(map[int]*domain.CardRow, len(karten))
	for _, karte := range karten {
		nachStelle[karte.Number] = karte
	}

	bloeckeAngleichen(baum, istKartenblock, nachStelle, karteAngleichen,
		func(ab int, zeilenende string) []*msepaket.Knoten {
			return neueBloecke(karten, ab, zeilenende)
		})
}

// istKartenblock sagt, ob ein Wurzelknoten eine Karte ist.
func istKartenblock(knoten *msepaket.Knoten) bool {
	return knoten.Zeilenart == msepaket.ArtEintrag && knoten.Schluessel == "card"
}

// neueBloecke gibt die Blöcke der Karten, für die es in der Datei keinen gibt.
//
// **Sie sind seit dem letzten Schreiben dazugekommen**
// (#623). Wer
// eine Karte anlegte und speicherte, bekam bis dahin eine Datei ohne sie: der
// Schreiber lief nur die vorhandenen Blöcke durch. Das Anlegen war damit
// stiller Datenverlust beim Speichern.
//
// **Der Rumpf bleibt leer und `karteAngleichen` füllt ihn** — über denselben
// Weg, über den ein neues Feld in einer vorhandenen Karte seine Zeile bekommt
// (#191). Zwei Wege, dieselbe Frage, liefen sonst auseinander.
func neueBloecke(karten []*domain.CardRow, ab int, zeilenende string) []*msepaket.Knoten {
	// **Nach der Stelle sortiert**, nicht in der Reihenfolge des Bestands: der
	// Leser gibt sie zwar geordnet, aber das ist seine Zusage und nicht die
	// dieser Stelle.
	nachStelle := slices.SortedFunc(slices.Values(karten),
		func(eine, andere *domain.CardRow) int { return eine.Number - andere.Number })

	var neue []*msepaket.Knoten

	for _, karte := range nachStelle {
		if karte.Number < ab {
			continue
		}

		knoten := &msepaket.Knoten{
			Zeilenart:  msepaket.ArtEintrag,
			Schluessel: "card",
			Zeilenende: zeilenende,
		}

		karteAngleichen(knoten, karte)
		neue = append(neue, knoten)
	}

	return neue
}

// karteAngleichen bringt den Kartenblock auf den Stand des Bestands.
//
// Drei Dinge, und sie sind nicht dasselbe: **Werte setzen**, **Zeilen
// entfernen**, **Zeilen ergänzen**. Bis
// #191 gab es
// nur das erste — wer ein Feld entfernte und speicherte, bekam eine Datei, in
// der es weiter stand.
func karteAngleichen(knoten *msepaket.Knoten, karte *domain.CardRow) {
	werteSetzen(knoten, karte)

	knoten.Kinder = zeilenEntfernen(knoten.Kinder, karte)
	knoten.Kinder = zeilenErgaenzen(knoten.Kinder, karte)
}
