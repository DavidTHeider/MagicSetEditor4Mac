// **Was in einem neu angelegten Paket steht** — die zweite der drei Fragen
// aus #1350.
//
// # Warum überhaupt etwas darin steht
//
// **MSE2 verlangt nichts.** `Game::validate` (`src/data/game.cpp`, Zeilen
// 66–88, Fassung f9d9356 vom 2021-07-06) baut Statistikdimensionen und einen
// Pack selbst und kommt mit null Kartenfeldern durch; `Packaged::validate`
// (`src/util/io/package.cpp`, Zeilen 621–631) setzt den Kurznamen notfalls aus
// dem Dateinamen. Ein leeres Paket wäre also **gültig**.
//
// **Und nutzlos.** Eine Karte auf einem Spiel ohne Felder hat nichts, was man
// eingeben könnte, und ein Stil ohne `card style` zeichnet eine leere Fläche.
// Wer eine Vorlage anlegt und ein weißes Rechteck bekommt, hat keine Vorlage
// bekommen — das ist die erste Dimension aus §10.
//
// # Die Saat ist eine Entscheidung und keine Übernahme
//
// Drei Kartenfelder (`name`, `text`, `kunst`) und ein Setfeld (`title`). Die
// **Gestalt** steht so in jedem Paket — `card field:` mit `type:` und `name:`
// darunter; der **Inhalt** ist eigener (§11). Aus `saat/probe.mse-game`
// stammt der Zuschnitt, nicht der Text.

package service

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// saatfeld ist ein Kartenfeld der Saat.
type saatfeld struct {
	name string
	art  string

	// hoehe ist die Höhe im Stil, in Stellen bei 375 × 523 und 150 dpi.
	hoehe int
}

// saatfelder sind die drei, die ein Spiel braucht, um benutzbar zu sein.
//
// **Drei und nicht eins**: ein Name allein macht keine Karte, und ein Bild
// ohne Text auch nicht. Drei zeigen die drei Feldarten, die eine Vorlage
// überhaupt kennt — Zeile, Absatz, Bild.
var saatfelder = []saatfeld{
	{name: "name", art: "text", hoehe: 28},
	{name: "text", art: "text", hoehe: 120},
	{name: "kunst", art: "image", hoehe: 240},
}

// spielbaum baut die `game`-Datei eines neuen Spiels.
func spielbaum(wunsch Neuanlage) *msepaket.Baum {
	wurzel := []*msepaket.Knoten{
		saatzeile("mse version", "2.0.0"),
		saatzeile("short name", wunsch.Name),
		saatzeile("full name", wunsch.Vollname),
		saatzeile("version", heute()),
		leerzeile(),
	}

	for _, feld := range saatfelder {
		wurzel = append(wurzel, mitKindern("card field",
			saatzeile("type", feld.art),
			saatzeile("name", feld.name),
		))
	}

	// **Ein Setfeld, damit ein Set einen Titel tragen kann.** Ohne eines
	// stünde in der Setverwaltung nichts als der Dateiname.
	return &msepaket.Baum{
		Datei: "game",
		Wurzel: append(wurzel, leerzeile(), mitKindern("set field",
			saatzeile("type", "text"),
			saatzeile("name", "title"),
		)),
		MitBOM: true,
	}
}

// stilbaum baut die `style`-Datei eines neuen Stils.
//
// **Er legt jedes Feld des Spiels aus**, das er bekommt — auch die, die diese
// Saat nicht kennt. Ein Stil auf einem bezogenen Spiel ist der Normalfall, und
// dort heißen die Felder anders.
func stilbaum(wunsch Neuanlage, felder []saatfeld) *msepaket.Baum {
	stile := make([]*msepaket.Knoten, 0, len(felder))
	oben := randOben

	for _, feld := range felder {
		zeilen := []*msepaket.Knoten{
			saatzeile("left", fmt.Sprint(randLinks)),
			saatzeile("top", fmt.Sprint(oben)),
			saatzeile("width", fmt.Sprint(kartenbreite-2*randLinks)),
			saatzeile("height", fmt.Sprint(feld.hoehe)),
		}

		// **Eine Schrift nur an einem Textfeld.** MSE2 liest `font:` an einem
		// `TextStyle` (`src/data/field/text.cpp`, Zeilen 143–159, Fassung
		// f9d9356 vom 2021-07-06); ein `ImageStyle` kennt nur `default`
		// (`src/data/field/image.cpp`, Zeilen 24–27) und meldet beim Öffnen
		// „Unexpected key: 'font'". Der erste Anlauf schrieb sie überall —
		// gefunden hat es nicht ein Test, sondern MSE2 selbst (§15).
		if feld.art == "text" {
			zeilen = append(zeilen, mitKindern("font",
				saatzeile("name", "Helvetica"),
				saatzeile("size", "12"),
			))
		}

		stile = append(stile, mitKindern(feld.name, zeilen...))

		oben += feld.hoehe + randOben
	}

	return &msepaket.Baum{
		Datei: "style",
		Wurzel: []*msepaket.Knoten{
			saatzeile("mse version", "2.0.0"),
			saatzeile("game", wunsch.Spiel),
			saatzeile("short name", wunsch.Name),
			saatzeile("full name", wunsch.Vollname),
			saatzeile("version", heute()),
			mitKindern("depends on",
				saatzeile("package", wunsch.Spiel+".mse-game"),
			),
			leerzeile(),
			saatzeile("card width", fmt.Sprint(kartenbreite)),
			saatzeile("card height", fmt.Sprint(kartenhoehe)),
			saatzeile("card dpi", "150"),
			leerzeile(),
			mitKindern("card style", stile...),
		},
		MitBOM: true,
	}
}

// Die Maße der Saat — dieselben wie in `saat/probe-schlicht.mse-style`.
//
// **Ein Kartenformat und keine Erfindung**: 375 × 523 bei 150 dpi sind
// 63,5 × 88,9 mm, das Maß einer Spielkarte.
const (
	kartenbreite = 375
	kartenhoehe  = 523
	randLinks    = 24
	randOben     = 24
)

// hoeheFuer gibt die Höhe eines Feldes im Stil.
//
// **Bekannte Felder bekommen ihre Höhe, unbekannte eine Zeile.** Ein Feld, von
// dem niemand weiß, was darin steht, als Absatz auszulegen wäre eine
// Behauptung; eine Zeile ist das Wenigste, das etwas zeigt.
func hoeheFuer(feld string) int {
	for _, saat := range saatfelder {
		if saat.name == feld {
			return saat.hoehe
		}
	}

	return 28
}

// saatzeile baut einen Knoten mit Schlüssel und Wert.
//
// **Ohne `Roh`**, und das ist der Punkt: [msepaket.Schreibe] setzt nur
// zusammen, was niemand gelesen hat. Eine Rohzeile hier hieße, die
// Einrückung selbst zu rechnen — und die Rundlaufzusage stünde auf zwei
// Rechnungen statt auf einer.
func saatzeile(schluessel, wert string) *msepaket.Knoten {
	return &msepaket.Knoten{
		Zeilenart: msepaket.ArtEintrag, Schluessel: schluessel, Wert: wert,
		Zeilenende: "\n",
	}
}

// mitKindern baut einen Knoten mit Kindern und ohne eigenen Wert.
func mitKindern(schluessel string, kinder ...*msepaket.Knoten) *msepaket.Knoten {
	return &msepaket.Knoten{
		Zeilenart: msepaket.ArtEintrag, Schluessel: schluessel,
		Kinder: kinder, Zeilenende: "\n",
	}
}

// leerzeile trennt zwei Abschnitte.
func leerzeile() *msepaket.Knoten {
	return &msepaket.Knoten{Zeilenart: msepaket.ArtLeer, Zeilenende: "\n"}
}
