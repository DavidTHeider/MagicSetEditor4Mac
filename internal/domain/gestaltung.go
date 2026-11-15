package domain

// **Eigene Merkmale** — UX-09.AK-04
// (#1681):
// *„Eigene CSS-Erweiterungen sind auf deklarative, eingeschränkte Gestaltung
// begrenzt; keine Skriptausführung, keine unbemerkten Netzbezüge und kein
// automatisches Verteilen nicht lizenzierter Schriftdateien."*
//
// # Warum es Farben sind und kein Stylesheet
//
// **Weil sich nur eine Farbe prüfen lässt.** Ein Stylesheet mit Selektoren
// wäre eine Adresse in einen Baum, der sich mit jeder Fassung ändert — und
// eine Regel `display: none` am Knopf „Zurücksetzen" nähme den Weg zurück
// gleich mit. Eine Zeile `--mse-grund: #1e1b16;` kann nichts davon: sie
// setzt eines der zehn Merkmale, aus denen die Oberfläche ihre Farben
// ohnehin nimmt (`frontend/src/stil.css`), und sonst nichts.
//
// **Damit ist AK-04 eine Eigenschaft der Form und keine Prüfung, die man
// umgehen könnte**: kein `url(` (kein Netz, keine Datei), kein `@import`,
// kein `@font-face` (keine Schrift), keine Maskierung, kein Ausdruck — die
// Zeile besteht aus einem bekannten Namen und einer Hexzahl, oder sie wird
// abgewiesen. Das Fenster setzt sie zudem einzeln über
// `style.setProperty` und nie als Text in ein `<style>`.
//
// # Warum `new` und wo gesucht wurde
//
// **MSE2 hat keine Oberflächengestaltung**: gesucht am 2026-09-25 in
// `/Users/david/mse2-quelle/src` (Fassung f9d9356 vom 2021-07-06) nach
// `theme`, `colour_scheme` und `stylesheet` im Sinn einer Oberfläche —
// `stylesheet` ist dort die Kartenvorlage, keine Fensterfarbe.

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Gestaltungsgrenze ist die größte Länge der eigenen Merkmale, in Zeichen.
//
// **Zehn Zeilen zu je etwa dreißig Zeichen** sind alles, was es gibt; die
// Grenze lässt Platz für Leerzeilen und nimmt einem Feld von Megabytes den
// Weg, bevor irgendetwas gelesen wird.
const Gestaltungsgrenze = 1000

// Gestaltungsmerkmale sind die zehn Merkmale, die sich ändern lassen.
//
// **Die Tischfarben gehören nicht dazu**: ihre Kanten halten 3 : 1 gegen
// **jeden** Untergrund, auf dem sie liegen (`docs/KONTRASTE.md`), und das
// lässt sich für eine eigene Farbe nicht zusagen.
func Gestaltungsmerkmale() []string {
	return []string{
		"--mse-grund", "--mse-flaeche", "--mse-text", "--mse-leise",
		"--mse-kante", "--mse-marke", "--mse-betont", "--mse-ring",
		"--mse-marketext", "--mse-warnung",
	}
}

// gestaltungszeile ist die ganze Gestalt einer Zeile: ein Name, ein
// Doppelpunkt, eine Hexzahl mit drei oder sechs Stellen, ein Semikolon.
var gestaltungszeile = regexp.MustCompile(
	`^(--mse-[a-z]+)\s*:\s*#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\s*;?$`)

// PruefeGestaltung urteilt über die eigenen Merkmale — leer heißt: keine.
//
// **Der Befund nennt die Zeile** (§8): wer zehn Zeilen geschrieben hat,
// sucht sonst die eine, die nicht passt.
func PruefeGestaltung(wert string) error {
	if utf8.RuneCountInString(wert) > Gestaltungsgrenze {
		return NewFieldError("wert", SchluesselValidierung,
			"Mehr als %d Zeichen sind keine Gestaltung mehr.", Gestaltungsgrenze)
	}

	gesehen := map[string]int{}

	for index, roh := range strings.Split(wert, "\n") {
		zeile := strings.TrimSpace(roh)
		if zeile == "" {
			continue
		}

		nummer := index + 1

		teile := gestaltungszeile.FindStringSubmatch(zeile)
		if teile == nil {
			return NewFieldError("wert", SchluesselValidierung,
				"Zeile %d: erwartet wird eine Angabe wie »--mse-grund: #1e1b16;« — ein Merkmal und eine Farbe als Hexzahl, eine je Zeile.",
				nummer)
		}

		name := teile[1]
		if !slices.Contains(Gestaltungsmerkmale(), name) {
			return NewFieldError("wert", SchluesselValidierung,
				"Zeile %d: »%s« ist kein Merkmal, das sich ändern lässt.", nummer, name)
		}

		if vorher, doppelt := gesehen[name]; doppelt {
			return NewFieldError("wert", SchluesselValidierung,
				"Zeile %d: »%s« steht schon in Zeile %d.", nummer, name, vorher)
		}

		gesehen[name] = nummer
	}

	return nil
}
