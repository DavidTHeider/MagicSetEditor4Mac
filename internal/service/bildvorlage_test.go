// **Ein Bildwert mit `{` in der Mitte ist eine Zeichenkette mit Ausdruck** —
// #1934.
//
// MSE2 liest `mask: border-mask{pt()}.png` im Zeichenkettenmodus
// (`Reader::handle(ScriptableImage&)`, `src/script/image.cpp` Zeilen 67–78,
// Fassung f9d9356 vom 2021-07-06): der Name ist `border-mask` und das
// Ergebnis von `pt()` und `.png`. Gemessen am 2026-09-29 über
// `full-magic-pack` steht die Form zweimal als eigener Schlüssel da, beide
// Male in `magic-crystal`; dieses Programm suchte die Datei
// `border-mask{pt()}.png`.
package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBildnameMitAusdruckWirdAusgerechnet(t *testing.T) {
	for name, fall := range map[string]struct {
		bild, pfad string
	}{
		"ein Ausdruck in der Mitte": {"rahmen_{card.farbe}.png", "/probe.mse-style/rahmen_rot.png"},
		"ein Ausdruck vorn":         {"{card.farbe}_rahmen.png", "/probe.mse-style/rot_rahmen.png"},
		"eine geschützte Klammer":   {`rahmen\{1\}.png`, "/probe.mse-style/rahmen{1}.png"},
		"ein anderes Paket bleibt":  {"/anderes.mse-include/{card.farbe}.png", "/anderes.mse-include/rot.png"},
	} {
		feld := flaeche("card color")
		feld.Darstellung = "image"
		feld.Bild = fall.bild

		plan, befunde := gedeutet(t, bildstil(feld))
		if len(befunde) > 0 {
			t.Errorf("%s: Befunde %v", name, befunde)

			continue
		}

		if bild := ebenen(t, plan)[0].Bild; bild.Pfad != fall.pfad {
			t.Errorf("%s: Bild %q, erwartet %q", name, bild.Pfad, fall.pfad)
		}
	}
}

// **Ein kaputter Ausdruck im Namen ist ein Befund**, kein Dateiname — sonst
// meldete die Bildquelle „nicht gefunden", wo „kaputt" richtig ist.
func TestEinKaputterAusdruckImBildnamenIstEinBefund(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Maske = "maske-{pt(}.png"
	feld.Bild = "rahmen.png"

	_, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "ließ sich nicht lesen") {
		t.Fatalf("Befunde %v, erwartet einer über die Maske", befunde)
	}
}

// **Die Werkstatt prüft ihn streng** — und findet den Fehler an seiner Stelle
// in der Datei.
func TestDiePruefungFindetEinenKaputtenAusdruckImBildnamen(t *testing.T) {
	for name, fall := range map[string]struct {
		datei         string
		zeile, spalte int
	}{
		"gültig":              {datei: "card style:\n\tgrund:\n\t\tmask: border-mask{pt()}.png\n"},
		"in einer Maske":      {"card style:\n\tgrund:\n\t\tmask: a{f(1,}.png\n", 3, 12},
		"in einem Wahlbild":   {"card style:\n\tfarbe:\n\t\tchoice images:\n\t\t\thell: h{(}.png\n", 4, 13},
		"unter refer script":  {"keyword parameter type:\n\trefer script:\n\t\tname: n\n\t\tscript: {f(}\n", 4, 13},
		"eine offene Klammer": {"card style:\n\tgrund:\n\t\timage: rahmen{pt(.png\n", 3, 16},
	} {
		befund := service.PaketSkriptfehler(fall.datei)

		switch {
		case fall.zeile == 0 && befund != nil:
			t.Errorf("%s: %v, erwartet kein Befund", name, befund)
		case fall.zeile != 0 && (befund == nil || befund.Zeile != fall.zeile || befund.Spalte != fall.spalte):
			t.Errorf("%s: %v, erwartet Zeile %d, Spalte %d", name, befund, fall.zeile, fall.spalte)
		}
	}
}

// **Gefärbt wird er wie ein Wert** — das Fenster kennt keine eigene Farbe für
// ihn, und eine unbekannte Art färbte gar nicht.
func TestEinBildnameMitAusdruckWirdWieEinWertGefaerbt(t *testing.T) {
	for _, merkmal := range service.Paketfarben("card style:\n\tgrund:\n\t\tmask: a{b}.png\n") {
		if merkmal.Zeile == 3 && merkmal.Art != string(msepaket.FarbeSchluessel) && merkmal.Art != string(msepaket.FarbeWert) {
			t.Errorf("Art %q in Zeile 3", merkmal.Art)
		}
	}
}
