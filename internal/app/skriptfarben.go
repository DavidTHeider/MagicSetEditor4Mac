// **Die Merkmale eines Ausdrucks, zum Einfärben** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Warum das eine Bindung ist und keine Sache des Fensters
//
// Weil der Lexer in Go steht und geprüft ist. Ein zweiter in TypeScript, der
// „ungefähr dieselben" Regeln kennt, liefe an dem Tag auseinander, an dem
// einer von beiden eine Regel lernt — und die Einfärbung zeigte eine Sprache,
// die es nicht gibt (§10). Dieselbe Überlegung wie bei den Namen der Umgebung:
// aus der wirklichen Quelle, nicht aus einer Liste daneben.
//
// # Warum sie ohne Fenster nicht gebraucht wird
//
// **Ohne Fenster wird nicht getippt.** Die Merkmale sind die Darstellung eines
// Textfelds und keine Aussage über ein Set; wer über MCP oder Kommandozeile
// einen Ausdruck schickt, hat ihn fertig.

package app

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Merkmalstelle ist ein Stück Quelltext mit seiner Art und seiner Stelle.
//
// **Zeile und Spalte zählen ab eins**, wie überall in diesem Programm und wie
// in jeder Fehlermeldung des Zerteilers. Ein Editor, der ab null zählte,
// müsste an jeder Meldung daneben umrechnen.
type Merkmalstelle struct {
	Art    string `json:"art"`
	Zeile  int    `json:"zeile"`
	Spalte int    `json:"spalte"`

	// Breite ist, wie viele Zeichen im **Quelltext** stehen — bei einer
	// Zeichenkette also samt ihren Anführungszeichen.
	Breite int `json:"breite"`
}

// Farbantwort trägt die Merkmale eines Ausdrucks.
type Farbantwort struct {
	Merkmale []Merkmalstelle `json:"merkmale"`
}

// Skriptfarben zerlegt einen Ausdruck in seine Merkmale.
//
// **Ohne Fehler und ohne Umschlag.** Es gibt nichts, was hier schiefgehen
// könnte: ein unlesbarer Ausdruck färbt so weit, wie er lesbar ist, und der
// Rest bleibt ungefärbt. Wo ein Fehler steht, sagt `Skript` — und das Fenster
// fragt ohnehin beides.
func (a *App) Skriptfarben(quelltext string) Farbantwort {
	merkmale := msescript.Einfaerben(quelltext)

	stellen := make([]Merkmalstelle, 0, len(merkmale))

	for _, merkmal := range merkmale {
		// **Zeilenenden bekommen keine Farbe.** Sie sind der Umbruch selbst;
		// ein Abschnitt darüber legte sich über den Zeilenwechsel.
		if merkmal.Art == msescript.ArtZeilenende {
			continue
		}

		stellen = append(stellen, Merkmalstelle{
			Art:    string(merkmal.Art),
			Zeile:  merkmal.Zeile,
			Spalte: merkmal.Spalte,
			Breite: merkmal.Breite,
		})
	}

	return Farbantwort{Merkmale: stellen}
}

// Paketfarben zerlegt eine **Paketdatei** in ihre Merkmale —
// #1181.
//
// # Warum eine zweite Bindung und kein Schalter
//
// **Weil es zwei Sprachen sind.** Eine `.mse-style` ist das Paketformat —
// `game: magic` ist ein Schlüssel mit einem Wert, kein Ausdruck. Der
// Skript-Lexer liest daraus Bezeichner und Doppelpunkte und färbt nichts
// Sinnvolles; deshalb stand im Vorlagenbereich bis heute `keine`.
//
// Ein Parameter „welche Sprache" wäre dieselbe Bindung mit einer Verzweigung
// darin — und die Verzweigung stünde dann in `app`, wo keine fachliche
// Entscheidung hingehört (§2).
//
// # Ohne Fehler und ohne Umschlag
//
// Wie [App.Skriptfarben] und aus demselben Grund: eine Datei, die sich nicht
// zerteilen lässt, färbt so weit, wie sie lesbar ist.
func (a *App) Paketfarben(quelltext string) Farbantwort {
	merkmale := service.Paketfarben(quelltext)

	stellen := make([]Merkmalstelle, 0, len(merkmale))

	for _, merkmal := range merkmale {
		stellen = append(stellen, Merkmalstelle{
			Art:    merkmal.Art,
			Zeile:  merkmal.Zeile,
			Spalte: merkmal.Spalte,
			Breite: merkmal.Breite,
		})
	}

	return Farbantwort{Merkmale: stellen}
}
