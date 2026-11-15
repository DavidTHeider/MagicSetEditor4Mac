package app_test

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Stiloption der Saat tut etwas** —
// #770.
//
// # Warum das einen eigenen Wächter braucht
//
// **Die Saat wird ausgeliefert** (`saat/LIESMICH.md`), und sie führte drei
// `styling field:` gegen **0** `styling.`-Nutzungen — gezählt am 2026-08-21
// über alle fünf Saatstile und das Spiel. Wer die Anwendung auf einem frischen
// Rechner öffnet, stellte drei Bedienelemente um, und es passierte nichts.
//
// **Ein Goldbild fängt das nicht.** `TestDieSaatkarteBleibtWieSieIst` prüft die
// Karte bei **Vorgabe**, und bei Vorgabe soll sich ja gerade nichts ändern. Der
// Fall, den niemand sah, ist der **umgestellte** Wert.
//
// # Was hier zugesichert wird
//
// Zwei Hälften, und die zweite ist die neue: bei Vorgabe steht das Bild wie
// bisher — das hält das Goldbild —, und bei umgestellter Option **ändert es
// sich**. Ohne die zweite wäre `color: { styling.titelfarbe }` nur ein totes
// Skript anstelle eines toten Literals.
func TestEineStiloptionDerSaatAendertDasBild(t *testing.T) {
	t.Parallel()

	// **Jede der drei, einzeln umgestellt.** Ein Wächter über eine Option
	// sagte über die anderen nichts — und genau so sind zwei tote Optionen
	// zehn Monate lang mitgeliefert worden (#770).
	faelle := map[string]string{
		"titelfarbe":    "rgb(255,255,255)",
		"hintergrund":   "dunkel",
		"rahmen zeigen": "yes",
	}

	vorgabe := zeichneSaatkarte(t, nil)

	for name, wert := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if bytes.Equal(vorgabe, zeichneSaatkarte(t, map[string]string{name: wert})) {
				t.Errorf("%q auf %q ändert das Bild nicht — die Option tut nichts",
					name, wert)
			}
		})
	}
}

// **Und der Vorgabewert bleibt der, der bisher fest dastand.**
//
// `titelfarbe` trägt `default: "black"`, und die Überschrift stand vorher auf
// `color: black`. Wer die Option auf **schwarz** stellt, muss dasselbe Bild
// bekommen wie ohne jede Angabe — sonst rechnet der Vorgabeweg etwas anderes
// als der gesetzte.
func TestDieVorgabeDerSaatoptionIstIhrBisherigerWert(t *testing.T) {
	t.Parallel()

	vorgabe := zeichneSaatkarte(t, nil)
	schwarz := zeichneSaatkarte(t, map[string]string{"titelfarbe": "rgb(0,0,0)"})

	if !bytes.Equal(vorgabe, schwarz) {
		t.Error("die Vorgabe zeichnet etwas anderes als der ausdrücklich gesetzte schwarze Wert")
	}
}

// zeichneSaatkarte zeichnet die Saatkarte, wahlweise mit gesetzten
// Stiloptionen.
//
// **Über `has styling` und `styling data`**, also denselben Weg, den ein
// echtes Set nimmt (`service/kartenstiloptionen.go`) — eine Vorrichtung, die
// den Wert direkt in die Umgebung schöbe, prüfte den falschen Weg (§3).
func zeichneSaatkarte(t *testing.T, optionen map[string]string) []byte {
	t.Helper()

	karte := saatkarte()

	if len(optionen) > 0 {
		zeilen := ""
		for name, wert := range optionen {
			zeilen += name + ": " + wert + "\n"
		}

		karte.Values = append(karte.Values,
			domain.Value{Field: service.Stilschalterfeld, Text: "true"},
			domain.Value{Field: service.Stilwertefeld, Text: zeilen})
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(probekontext(t), karte, service.Vorlage{Stil: saatstil(t)})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	bild, rasterbefunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften.Neue(schriften.Orte())).
		Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(rasterbefunde) > 0 {
		t.Fatalf("Rasterbefunde: %v", rasterbefunde)
	}

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	return puffer.Bytes()
}
