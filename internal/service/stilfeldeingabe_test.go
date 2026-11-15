package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Skript eines Feldstils liest `input` — und das ist der Wert dieses
// Felds auf dieser Karte**
// (#240).
//
// # Herkunft
//
// MSE-Skripte kennen keine Stellungsargumente; sie lesen die Kontextvariable
// `input` (`SCRIPT_PARAM_C(String,input)`, `src/script/functions/editor.cpp`
// von twanvl/MagicSetEditor2, Stand 2026-08-16). Für das Skript eines Feldes
// ist der Kontext das Feld, und `input` damit sein Wert.
//
// # Was gemessen daran hing
//
// Am laufenden Fenster, mit einem Set aus der Saat auf `magic-m15`:
//
//	card color → image: die Angabe ließ sich nicht ausrechnen:
//	Zeile 12642, Spalte 4: der Name "input" ist hier nicht bekannt
//
// Über die 190 Stile von `M15-Magic-Pack` waren es **162** Feldstile. Der
// häufigste Aufruf im Pack ist `chosen@(choice: …)` — teilweise angewandt, und
// das fehlende Stellungsargument kommt genau von hier.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinFeldskriptLiestSeinenEigenenWertAlsInput(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = `{ "rahmen_" + input + ".png" }`

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "card color", Text: "blau"},
	}}

	plan, befunde := gedeutetMit(t, karte, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/rahmen_blau.png" {
		t.Errorf("Bild %q", bild.Pfad)
	}
}

// **Ein Feld, für das die Karte nichts führt, bekommt den leeren Text.**
//
// Nicht „unbekannter Name": `indicator` in `magic-m15.mse-style` rechnet sein
// Bild aus `has_identity()` aus, und die Karte führt kein Feld dieses Namens.
// Ein Fehler dort ließe jedes rein gerechnete Bild fallen — und der leere Text
// ist bei MSE ohnehin, was `nil` in einem Text ergibt.
func TestEinFeldOhneKartenwertBekommtDenLeerenText(t *testing.T) {
	feld := flaeche("indicator")
	feld.Darstellung = "image"
	feld.Bild = `{ "rahmen" + input + ".png" }`

	karte := &domain.CardRow{ID: 1, Number: 1}

	plan, befunde := gedeutetMit(t, karte, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/rahmen.png" {
		t.Errorf("Bild %q", bild.Pfad)
	}
}

// **Das `input` gilt für dieses Feld und nicht für das nächste.**
//
// Die Gegenprobe: wer es in die gemeinsame Umgebung schriebe statt in eine
// Ebene darüber, ließe das zweite Feld mit dem Wert des ersten rechnen — und
// das fiele erst an einer Karte auf, bei der beide verschieden sind.
func TestJedesFeldBekommtSeinEigenesInput(t *testing.T) {
	erstes := flaeche("card color")
	erstes.Darstellung = "image"
	erstes.Bild = `{ "eins_" + input + ".png" }`

	zweites := flaeche("indicator")
	zweites.Darstellung = "image"
	zweites.Bild = `{ "zwei_" + input + ".png" }`

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "card color", Text: "rot"},
		{Field: "indicator", Text: "gruen"},
	}}

	plan, befunde := gedeutetMit(t, karte, bildstil(erstes, zweites))
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gebaut := ebenen(t, plan)

	if len(gebaut) != 2 {
		t.Fatalf("%d Ebenen", len(gebaut))
	}

	if gebaut[0].Bild.Pfad != "/probe.mse-style/eins_rot.png" {
		t.Errorf("erste Ebene %q", gebaut[0].Bild.Pfad)
	}

	if gebaut[1].Bild.Pfad != "/probe.mse-style/zwei_gruen.png" {
		t.Errorf("zweite Ebene %q", gebaut[1].Bild.Pfad)
	}
}
