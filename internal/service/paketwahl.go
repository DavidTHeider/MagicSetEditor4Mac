package service

import (
	"context"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Wahl einer `package choice` — #632.
//
// **Eine eigene Datei**, weil `stiloptionen.go` sonst zwei Fragen beantwortete:
// dort steht, welchen Wert eine Option hat, hier, welche Werte ihr offenstehen
// (§2).

// Musterablage beantwortet, welche Pakete auf ein `match:`-Muster passen.
//
// **Eine eigene Naht neben [Paketablage]**, obwohl derselbe Katalog beide
// erfüllt: wer Stiloptionen liest, braucht nicht den ganzen Katalog, und eine
// Schnittstelle, die mehr verlangt als sie benutzt, macht jeden Test größer,
// als er sein müsste.
type Musterablage interface {
	PaketeNachMuster(ctx context.Context, muster string) ([]domain.Paket, error)
}

// moeglichkeitenVon baut die Wahlen einer Stiloption — je nach Art.
//
// **Eine Stelle für alle vier**, weil es eine Frage ist: was steht dieser
// Option offen. Vier Stellen liefen auseinander, sobald eine fünfte Art
// dazukäme.
func moeglichkeitenVon(
	ctx context.Context, ablage Musterablage, feld domain.StyleField,
) ([]domain.Wahlmoeglichkeit, error) {
	switch feld.Kind {
	case domain.KindPackageChoice:
		return Paketwahlen(ctx, ablage, feld)

	case domain.KindColor:
		return Farbwahlen(feld), nil

	case domain.KindChoice, domain.KindBoolean, domain.KindMultipleChoice:
		return gleichnamigeWahlen(feld.Choices), nil

	default:
		return nil, nil
	}
}

// gleichnamigeWahlen gibt Wahlen, deren Wert ihre Beschriftung ist.
//
// **`choice`, `boolean` und `multiple choice`**: die Wahl heißt, wie sie
// heißt. `yes` und `no` stehen so in jeder Vorlage, und ein Feldskript
// vergleicht damit.
//
// **Die Reihenfolge ist die der Erklärung**, und bei einer Mehrfachwahl hängt
// daran mehr als die Anzeige: die Normalform ordnet den gespeicherten Wert
// genauso (#637).
func gleichnamigeWahlen(namen []string) []domain.Wahlmoeglichkeit {
	wahlen := make([]domain.Wahlmoeglichkeit, 0, len(namen))

	for _, name := range namen {
		wahlen = append(wahlen, domain.Wahlmoeglichkeit{Wert: name, Beschriftung: name})
	}

	return wahlen
}

// Farbwahlen gibt die benannten Farben einer [domain.KindColor].
//
// **ported.** MSE zeigt zu jeder `choice:` ihren Namen mit einer Farbkachel
// (`src/gui/value/color.cpp`, Zeilen 60–72, Fassung f9d9356 vom 2021-07-06).
// Gespeichert wird die **Farbe** — `rgb(r,g,b)`, die Form, die
// `format_color` erzeugt (`src/gfx/color.cpp`, Zeilen 76–83) —, angezeigt der
// Name.
//
// **Eine Option ohne `choice:` gibt keine Liste**, und das ist kein Mangel: 16
// der 105 Farbstiloptionen des Packs nennen keine, und für die ist der
// Farbwähler das Einzige, was es gibt (#630).
//
// # Was hier fehlt
//
// **Der Vorgabeeintrag.** MSE stellt einer Farboption mit `default:` einen
// Eintrag voran, der sagt „nimm, was das Skript rechnet". Bei uns fehlt dafür
// die Unterscheidung: ein leerer gespeicherter Wert und „nichts gesetzt" sind
// dasselbe. Das sind 10 der 105 und ein eigenes Werkstück (#634).
func Farbwahlen(feld domain.StyleField) []domain.Wahlmoeglichkeit {
	wahlen := make([]domain.Wahlmoeglichkeit, 0, len(feld.Choices))

	for stelle, name := range feld.Choices {
		if stelle >= len(feld.Wahlfarben) {
			break
		}

		wahlen = append(wahlen, domain.Wahlmoeglichkeit{
			Wert: feld.Wahlfarben[stelle], Beschriftung: name,
		})
	}

	return wahlen
}

// Paketwahlen baut die Liste einer [domain.KindPackageChoice].
//
// # Was MSE dabei tut
//
// **ported** (`src/render/value/package_choice.cpp`, Zeilen 28–43, Fassung
// f9d9356 vom 2021-07-06):
//
//   - Gefragt wird der Paketverwalter nach `match:`; die Ordnung — Rang, dann
//     Name — kommt aus dem Katalog.
//   - **Gespeichert wird `relativeFilename()`**, also der Dateiname samt
//     Endung. Ein Skript vergleicht damit, und eine andere Form machte jedes
//     vorhandene Set unlesbar.
//   - **Angezeigt wird `capitalize_sentence(short_name)`** — der erste
//     Buchstabe groß, sonst nichts (`src/util/string.cpp`, Zeilen 115–119).
//
// Ist keine Wahl verlangt, steht der leere Eintrag **vorn**
// (`src/gui/value/package_choice.cpp`, Zeilen 44–59).
//
// # Was hier fehlt
//
// **Das Symbol.** MSE zeigt neben jedem Eintrag das 16×16-Bild des Pakets. Ein
// `<select>` trägt keine Bilder; eine eigene Aufklappliste wäre ein eigenes
// Werkstück, und bis dahin steht der Name allein da.
func Paketwahlen(
	ctx context.Context, ablage Musterablage, feld domain.StyleField,
) ([]domain.Wahlmoeglichkeit, error) {
	pakete, err := ablage.PaketeNachMuster(ctx, feld.Match)
	if err != nil {
		return nil, err
	}

	wahlen := make([]domain.Wahlmoeglichkeit, 0, len(pakete)+1)

	if !feld.Required {
		wahlen = append(wahlen, domain.Wahlmoeglichkeit{Beschriftung: grossAmAnfang(feld.EmptyName)})
	}

	for _, paket := range pakete {
		wahlen = append(wahlen, domain.Wahlmoeglichkeit{
			Wert: paket.Name + "." + paket.Art,
			// **Dieselbe Regel wie in der Skriptumgebung** (#636): `short name:`,
			// sonst der Dateiname. `paket.Name` trägt ihn schon ohne Endung.
			Beschriftung: grossAmAnfang(Kurzname(paket.Kurzname, paket.Name)),
		})
	}

	return wahlen, nil
}

// grossAmAnfang macht den ersten Buchstaben groß und sonst nichts.
//
// **`capitalize_sentence`, nicht `capitalize`**: MSE hat beides, und für die
// Paketwahl nimmt es das erste (`src/util/string.cpp`, Zeilen 115–119, Fassung
// f9d9356 vom 2021-07-06). Aus `folie glänzend` wird `Folie glänzend`, nicht
// `Folie Glänzend`.
func grossAmAnfang(text string) string {
	zeichen := []rune(text)
	if len(zeichen) == 0 {
		return text
	}

	zeichen[0] = unicode.ToUpper(zeichen[0])

	return string(zeichen)
}
