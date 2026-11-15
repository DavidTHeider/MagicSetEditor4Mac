// **Kein Schlagschatten unter etwas, das man bedient** — Vermerk
// [0065](../../docs/entscheidungen/0065-wie-diese-anwendung-aussieht.md),
// Festlegung 3 und die Ausnahme in Festlegung 9.
//
// # Der Befund
//
// Festlegung 3 verwirft Schlagschatten, Festlegung 9 erlaubt Tiefe an **zwei**
// benannten Stellen — der Tischfläche samt Sitzen und einer Karte, wo sie als
// körperlicher Gegenstand liegt — und sagt dann ausdrücklich:
//
//	Überall sonst gilt Festlegung 3 unverändert: Kästen, Felder, Drawer,
//	Befehlspalette und Popover tragen ihre Kante und keinen Schatten.
//
// **Die Anwendung trug am 2026-09-20 sechs** — und darunter genau die drei,
// die der Satz benennt: die Seitenleiste als Schublade (`shadow-xl`), die
// Befehlspalette (`shadow-2xl`) und die Menüliste als Popover (`shadow-lg`).
// Dazu Tooltip, Einführung und Kommandozentrale.
//
// # Warum es niemand gesehen hat
//
// **Weil an der falschen Stelle gemessen wurde.** Der Entwurf der
// Zieloberfläche belegte die Einhaltung mit dem Satz *„`frontend/src/stil.css`
// enthält null `box-shadow`"* — und das stimmt. Die Schatten stehen nicht dort:
// sie stehen als **Tailwind-Klassen** in den Bauteilen. Eine Messung über die
// Merkmalsdatei reicht so weit wie die Merkmalsdatei, und §10 sagt dazu: die
// Reichweite einer Messung ist die Reichweite ihrer Aussage.
//
// # Was er hält
//
// Keine `shadow-…`-Klasse und kein `box-shadow` in `frontend/src` — mit
// **einer** Ausnahme seit dem Nachtrag vom 2026-10-13 (Festlegung 11, AP-1,
// #2337):
// die Stufe `shadow-ueberlage`, bestimmt in `stil.css` und benutzt nur an
// den vier Überlagerungen in [ueberlagen]. Eine zweite Stufe, ein Schatten
// an einem Knopf oder dieselbe Stufe an einem Kasten der Seite bleiben
// Funde.
//
// # Was er **nicht** hält
//
//   - **Nicht die erlaubte Tiefe selbst.** Ob der Verlauf auf der Tischfläche
//     die gemessenen Kontraste trägt, sagt `KONTRASTE.md` und kein Muster.
//
//   - **Nicht, was PixiJS zeichnet.** Der Tisch entsteht über eine Leinwand,
//     nicht über Klassen; ein Schatten dort käme aus `tisch/` und nicht aus
//     einem Klassennamen.
//
//   - **Nicht `inset`.** Ein eingesetzter Streifen ist eine Kante und kein
//     Schatten — der Entwurf markiert die gewählte Zeile so. Käme er als
//     Tailwind-Klasse, stünde er hier trotzdem als Fund; heute gibt es keinen,
//     und eine Ausnahme für einen Fall ohne Beispiel wäre eine Einladung.
//
//   - **Nicht `ring-…`.** Tailwind setzt einen Ring als `box-shadow` um, und
//     `Zuschnittrahmen.tsx` braucht ihn: er zeichnet die **zweite** Linie des
//     Zuschnittrahmens, portiert aus `image_slice_window.cpp`. Ein Ring ist
//     eine Kante — genau das, was Festlegung 3 verlangt.
//
//   - **Und Kommentare zählen nicht.** Der erste Anlauf schlug an
//     `Zuschnittrahmen.tsx` an, und dort stand `box-shadow` in einem Satz über
//     eine **Messung** im laufenden Fenster. Ein Wächter mit falschem Alarm
//     ist schlechter als keiner (§10).
//
//     **Nur `//`-Zeilen**, denn dieses Repositorium schreibt keine
//     Blockkommentare — in `.css` wären es `/* … */`, und dort steht heute
//     keiner. Käme einer, meldete dieser Wächter ihn als Fund; das ist die
//     sichere Richtung.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// schattenklassen findet eine Tailwind-Schattenklasse oder rohes CSS.
//
// **`shadow-` und nicht `shadow`**: Tailwind kennt `shadow-md` ebenso wie
// `shadow-none`, und das zweite ist die Abwesenheit. Eine Grenze am Bindestrich
// trifft beide; die Liste unten entscheidet, welche zählen.
var schattenklassen = regexp.MustCompile(`\bshadow-(\w+)|box-shadow`)

// ohneTiefe sind die Werte, die keinen Schatten erzeugen.
var ohneTiefe = map[string]bool{"none": true, "transparent": true}

// ueberlagen sind die Dateien, die die eine Schattenstufe tragen dürfen —
// Vermerk 0065, Festlegung 11 (AP-1,
// #2337):
// Befehlspalette, Popover samt Menüliste und Kontextmenü. **Nur
// `shadow-ueberlage`**, keine zweite Stufe; und `stil.css` darf das Merkmal
// **bestimmen**, nicht benutzen.
//
// **Die Tiefe aus Festlegung 9** — Tischfläche und liegende Karte — zeichnet
// `frontend/src/tisch/` über eine Leinwand und nicht über Klassen; sie steht
// deshalb nicht hier.
var ueberlagen = map[string]bool{
	"frontend/src/rahmen/Palette.tsx":         true,
	"frontend/src/bausteine/Popover.tsx":      true,
	"frontend/src/bausteine/Menueliste.tsx":   true,
	"frontend/src/bausteine/Kontextmenue.tsx": true,
}

// schattenstufe ist das eine Merkmal aus Festlegung 11.
const schattenstufe = "ueberlage"

// schattenfehler nennt die Schatten einer Datei, die nicht erlaubt sind.
func schattenfehler(kurz, inhalt string) []string {
	var funde []string

	for _, treffer := range schattenklassen.FindAllStringSubmatch(inhalt, -1) {
		switch {
		case ohneTiefe[treffer[1]]:
		case treffer[1] == schattenstufe && ueberlagen[kurz]:
		case kurz == "frontend/src/stil.css" && treffer[1] == schattenstufe &&
			strings.Contains(inhalt, "--shadow-"+schattenstufe+":"):
		default:
			funde = append(funde, treffer[0])
		}
	}

	// **Und wer mehr Kontrast verlangt, bekommt die Stufe nicht** — dieselbe
	// Regel wie bei der Faser (Festlegung 7): ein Schatten arbeitet gegen
	// die Kante, an der die Überlagerung zu erkennen ist.
	if ueberlagen[kurz] && strings.Contains(inhalt, "shadow-"+schattenstufe) &&
		!strings.Contains(inhalt, "contrast-more:shadow-none") {
		funde = append(funde, "shadow-"+schattenstufe+" ohne contrast-more:shadow-none")
	}

	return funde
}

// **Die Gegenprobe zuerst**: die eine Stufe an einer Überlagerung geht
// durch, an einem Knopf nicht, und eine zweite Stufe nirgends.
func TestDerSchattenpruefer(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		datei, inhalt string
		funde         int
	}{
		{"frontend/src/bausteine/Popover.tsx", `className="border-2 shadow-ueberlage contrast-more:shadow-none"`, 0},
		{"frontend/src/bausteine/Popover.tsx", `className="border-2 shadow-ueberlage"`, 1},
		{"frontend/src/bausteine/Knopf.tsx", `className="bg-marke shadow-ueberlage"`, 1},
		{"frontend/src/bausteine/Popover.tsx", `className="shadow-lg"`, 1},
		{"frontend/src/stil.css", `--shadow-ueberlage: 0 6px 18px rgb(0 0 0 / 0.16);`, 0},
		{"frontend/src/stil.css", `.knopf { box-shadow: 0 1px 2px black; }`, 1},
		{"frontend/src/bausteine/Kasten.tsx", `className="shadow-none"`, 0},
	} {
		if gefunden := schattenfehler(fall.datei, fall.inhalt); len(gefunden) != fall.funde {
			t.Errorf("%s: %q — %d Funde erwartet, %d gefunden: %v",
				fall.datei, fall.inhalt, fall.funde, len(gefunden), gefunden)
		}
	}
}

// TestKeinSchlagschattenUnterEinerBedienflaeche hält Festlegung 3.
func TestKeinSchlagschattenUnterEinerBedienflaeche(t *testing.T) {
	t.Parallel()

	ordner := filepath.Join(wurzel, "frontend", "src")
	gesehen := 0

	fehler := filepath.WalkDir(ordner, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || !stilträger(pfad) {
			return err
		}

		gesehen++

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum
		if lesefehler != nil {
			return lesefehler
		}

		pruefeAufTiefe(t, pfad, ohneZeilenkommentare(string(roh)))

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", ordner, fehler)
	}

	if gesehen < 100 {
		t.Fatalf("nur %d Dateien gelesen — der Wächter greift daneben", gesehen)
	}
}

// stilträger sagt, ob eine Datei Klassen oder CSS tragen kann.
func stilträger(pfad string) bool {
	name := filepath.Base(pfad)
	if strings.Contains(name, ".test.") {
		return false
	}

	return strings.HasSuffix(name, ".tsx") ||
		strings.HasSuffix(name, ".ts") ||
		strings.HasSuffix(name, ".css")
}

// pruefeAufTiefe meldet jeden Schatten, der nicht erlaubt ist.
func pruefeAufTiefe(t *testing.T, pfad, inhalt string) {
	t.Helper()

	kurz, fehler := filepath.Rel(wurzel, pfad)
	if fehler != nil {
		kurz = pfad
	}

	for _, fund := range schattenfehler(filepath.ToSlash(kurz), inhalt) {
		t.Errorf("%s trägt %q — Vermerk 0065, Festlegungen 3 und 11: Tiefe "+
			"gibt es an Tisch und liegender Karte (Festlegung 9) und als eine "+
			"Stufe `shadow-ueberlage` an Palette, Popover, Menüliste und "+
			"Kontextmenü — nie unter etwas, das man bedient",
			filepath.ToSlash(kurz), fund)
	}
}
