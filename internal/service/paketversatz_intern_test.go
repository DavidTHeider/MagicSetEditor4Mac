// **Der Rückfall der Versetzung** — Werkstück B von Phase 4.9
// (#1324).
//
// # Warum ein interner Lauf
//
// `versetzt` rechnet eine Stelle im **Skriptrumpf** in eine der Datei um. Ihr
// Rückfall greift, wenn der Lexer eine Zeile nennt, die die Zeilenliste des
// Rumpfes nicht führt — über den echten Weg kann das nicht geschehen, denn
// beide kommen aus demselben Text.
//
// **Der Zweig steht trotzdem da**, und zwar mit Absicht: ein Lexer, der eines
// Tages über den Rumpf hinausliest, soll eine Marke am Anfang setzen und
// nicht an einer erfundenen Stelle. Ein Zweig, den kein Lauf begeht, ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3) — hier wird sie
// eingelöst.
package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDerVersatzFaelltAufDenAnfangZurueck(t *testing.T) {
	wert := msepaket.Farbmerkmal{
		Zeile: 7, Spalte: 3,
		Zeilen: []msepaket.Skriptzeile{{Text: "1 + 2", Zeile: 7, Spalte: 3}},
	}

	for name, fall := range map[string]struct{ zeile, spalte int }{
		"eine Zeile jenseits des Rumpfes": {zeile: 2, spalte: 1},
		"Zeile null":                      {zeile: 0, spalte: 1},
	} {
		t.Run(name, func(t *testing.T) {
			zeile, spalte := versetzt(wert, fall.zeile, fall.spalte)

			if zeile != 7 || spalte != 3 {
				t.Errorf("Zeile %d, Spalte %d — erwartet der Anfang des Rumpfes (7, 3)",
					zeile, spalte)
			}
		})
	}
}

// **Und innerhalb des Rumpfes zählt die Zeile ihre eigene Einrückung.**
//
// Die Gegenprobe: ohne sie wäre der Lauf darüber auch grün, wenn `versetzt`
// **immer** den Anfang gäbe.
func TestDerVersatzNimmtDieEinrueckungJederZeile(t *testing.T) {
	wert := msepaket.Farbmerkmal{
		Zeile: 7, Spalte: 3,
		Zeilen: []msepaket.Skriptzeile{
			{Text: "f(1,", Zeile: 7, Spalte: 3},
			{Text: "2)", Zeile: 8, Spalte: 5},
		},
	}

	zeile, spalte := versetzt(wert, 2, 2)

	if zeile != 8 || spalte != 6 {
		t.Errorf("Zeile %d, Spalte %d — erwartet 8 und 6", zeile, spalte)
	}
}
