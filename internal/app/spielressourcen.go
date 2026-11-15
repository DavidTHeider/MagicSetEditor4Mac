package app

// **Woher die Abdrücke der Ressourcen kommen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum diese Naht hier steht
//
// **Weil nur der Composition Root beides kennt.** Der Schlüssel einer Textur
// braucht, was in den **Dateien** steht, die ein Zeichenplan nur nennt: das
// Bild hinter einem Paketpfad und die Schrift hinter einem Namen. Das eine
// weiß `internal/msepaket`, das andere `internal/schriften` — und
// `internal/service` darf keines von beiden unmittelbar fragen (§2).
//
// # Und warum beide Schnitte gefragt werden
//
// **Weil beide zeichnen.** Ein Stil nennt `italic name:` für den kursiven
// Schnitt, und MSE setzt fett über ein **Gewicht**
// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
// Welche Datei das Betriebssystem für „MPlantin" fett findet, ist eine andere
// als aufrecht — und ein Schlüssel, der nur die aufrechte kennte, ließe einen
// ausgetauschten fetten Schnitt durch.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// spielressourcen erfüllt `service.Ressourcenquelle` über den echten Quellen.
type spielressourcen struct {
	bilder    *msepaket.Bildabdruecke
	schriften *schriften.Suche
}

// Bildabdruck gibt den Abdruck des Bildes unter einem Paketpfad.
func (s spielressourcen) Bildabdruck(pfad string) string {
	return s.bilder.Abdruck(pfad)
}

// Schriftabdruck gibt den Abdruck der Schrift zu einem Namen.
//
// **Beide Schnitte in einem Wert**, getrennt durch ein Trennzeichen: der
// Plan nennt einen Namen, gezeichnet wird mit zweien.
func (s spielressourcen) Schriftabdruck(name string) string {
	return s.schriften.Abdruck(name, false) + "\x1f" +
		s.schriften.Abdruck(name, true)
}

// Rueckenbilddatei ist, wie ein Spielpaket seinen Kartenrücken nennt.
//
// # Warum am Dateinamen und nicht am `icon:` des Pakets
//
// **Weil `icon:` in fünf von sieben Spielpaketen eine Beispielkarte ist.**
// Gemessen am 2026-09-15 über `Full-Magic-Pack`: `archenemy`, `planechase`,
// `saga`, `vanguard` und `herospath` nennen dort `card-sample` — eine
// **Vorderseite**. Wer das Symbol nähme, legte auf jede verdeckte
// Archenemy-Karte eine offene.
//
// **`card-back.png` dagegen ist einer**: `magic.mse-game` und
// `space.mse-game` führen ihn, und beide zeigen wirklich einen Rücken —
// angesehen, nicht am Namen erschlossen.
const Rueckenbilddatei = "card-back.png"

// Rueckenbild gibt den Paketpfad des Kartenrückens eines Spiels.
//
// **Leer, wo es keinen gibt** — dann zeichnet `internal/service` einen. Ein
// Fehler wäre hier falsch: ein Spiel ohne Rücken ist kein Fehler, sondern ein
// Spiel ohne Rücken.
func (s spielressourcen) Rueckenbild(spiel string) string {
	if spiel == "" {
		return ""
	}

	// **Mit führendem Schrägstrich**, denn so sieht ein Paketverweis aus:
	// `/magic.mse-game/card-back.png`. [msepaket.ZerlegePaketpfad] weist
	// jeden ohne ab — der erste Anlauf baute ihn ohne, und damit fand er
	// **nie** einen Rücken; auffallen konnte das nicht, denn „kein Rücken"
	// ist eine gültige Antwort, und dann wird gezeichnet.
	verweis := "/" + spiel + ".mse-game/" + Rueckenbilddatei

	if !s.bilder.Hat(verweis) {
		return ""
	}

	return verweis
}
