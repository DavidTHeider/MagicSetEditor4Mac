package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Einfärbung einer Paketdatei, samt ihrer Skriptwerte** —
// #1181.
//
// # Warum die Zusammensetzung hier steht
//
// `msepaket.Einfaerben` benennt einen Skriptwert, färbt ihn aber nicht: es ist
// ein Blatt und darf `msescript` nicht kennen (§2). `internal/service` kennt
// beide — und die Zusammensetzung ist genau das, was eine Anwendungsschicht
// tut.
//
// **Das ist die richtige Trennung und nicht bloß die erlaubte**: was ein
// Skriptwert **ist**, weiß das Paketformat; **wie ein Skript aussieht**, weiß
// die Sprache.

// Farbmerkmal ist ein Stück Quelltext mit seiner Art und seiner Stelle.
type Farbmerkmal struct {
	Art    string
	Zeile  int
	Spalte int
	Breite int
}

// Paketfarben zerlegt eine Paketdatei in ihre Merkmale.
//
// **Ohne Fehler.** Eine Datei, die sich nicht zerteilen lässt, färbt so weit,
// wie sie lesbar ist — dieselbe Haltung wie bei `msescript.Einfaerben`.
func Paketfarben(quelltext string) []Farbmerkmal {
	roh := msepaket.Einfaerben(quelltext)

	merkmale := make([]Farbmerkmal, 0, len(roh))

	for _, eins := range roh {
		if eins.Art != msepaket.FarbeSkriptwert {
			merkmale = append(merkmale, Farbmerkmal{
				Art: string(wertfarbe(eins.Art)), Zeile: eins.Zeile,
				Spalte: eins.Spalte, Breite: eins.Breite,
			})

			continue
		}

		merkmale = append(merkmale, skriptmerkmale(eins)...)
	}

	return merkmale
}

// skriptmerkmale färbt einen Skriptwert über den Lexer der Sprache.
//
// **Versetzt auf seine Stelle in der Datei.** Der Lexer zählt ab Zeile eins,
// Spalte eins; hier steht der Rumpf mitten in einer Paketdatei.
//
// **Jede Zeile trägt ihren eigenen Versatz** — seit dem 2026-09-02, Werkstück
// B von Phase 4.9
// (#1324):
// ein Rumpf **unter** seinem Schlüssel ist eingerückt, und dort ist Spalte 1
// des Lexers Spalte 3 der Datei.
// wertfarbe färbt einen Textskriptwert wie einen Wert (#1934) — das Fenster
// kennt keine eigene Farbe für ihn, und `quelltextfarben.ts` ließe eine
// unbekannte Art ungefärbt.
func wertfarbe(art msepaket.Farbart) msepaket.Farbart {
	if art == msepaket.FarbeTextskriptwert {
		return msepaket.FarbeWert
	}

	return art
}

func skriptmerkmale(wert msepaket.Farbmerkmal) []Farbmerkmal {
	roh := msescript.Einfaerben(wert.Text)

	merkmale := make([]Farbmerkmal, 0, len(roh))

	for _, merkmal := range roh {
		// **Zeilenenden bekommen keine Farbe.** Sie sind der Umbruch selbst;
		// ein Abschnitt darüber legte sich über den Zeilenwechsel.
		if merkmal.Art == msescript.ArtZeilenende {
			continue
		}

		zeile, spalte := versetzt(wert, merkmal.Zeile, merkmal.Spalte)

		merkmale = append(merkmale, Farbmerkmal{
			Art:    string(merkmal.Art),
			Zeile:  zeile,
			Spalte: spalte,
			Breite: merkmal.Breite,
		})
	}

	return merkmale
}

// versetzt rechnet Zeile und Spalte **im Skriptrumpf** in die der Datei um.
//
// **Über [msepaket.Farbmerkmal.Zeilen] und nicht über eine Rechnung.** Wo
// jede Zeile des Rumpfes in der Datei steht, weiß das Paketformat; es hängt
// an der Einrückung, und die ist je Zeile eine andere. Wer hier rechnete,
// baute die Kenntnis des Formats ein zweites Mal nach.
//
// **Der Rückfall ist der alte Weg.** Ein Merkmal ohne Zeilenliste kann es
// nach dieser Änderung nicht mehr geben — aber ein Lexer, der über den Rumpf
// hinausliest, kann eine Zeile nennen, die die Liste nicht führt. Dann steht
// die Marke am Anfang des Rumpfes statt an einer erfundenen Stelle.
func versetzt(
	wert msepaket.Farbmerkmal, zeile, spalte int,
) (dateizeile, dateispalte int) {
	if zeile < 1 || zeile > len(wert.Zeilen) {
		return wert.Zeile, wert.Spalte
	}

	anfang := wert.Zeilen[zeile-1]

	return anfang.Zeile, anfang.Spalte + spalte - 1
}
