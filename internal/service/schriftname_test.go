package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// namensfeld baut ein Textfeld mit einem Schriftnamen.
func namensfeld(name, kursiv string) msepaket.Feldstil {
	feld := flaeche("name")
	feld.Schrift = msepaket.Schrift{
		Name: name, KursivName: kursiv, Groesse: "12", Farbe: "black",
	}

	return feld
}

// **Ein Schriftname darf ein Skript sein.**
//
// Über die 190 Stile von `M15-Magic-Pack` sind **664 von 3.119** `font: name:`
// Skripte, dazu **439 von 581** `italic name:`. Bis
// #393 ging der
// rohe Text an die Schriftsuche — gesucht wurde nach einer Schrift namens
// `{ body_font() }`, und die Karte blieb ohne Text.
func TestEinSchriftnameDarfEinSkriptSein(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(namensfeld(`{ "Helvetica" }`, `{ "Helvetica-Oblique" }`)))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gesetzt := textebene(t, plan)
	if gesetzt.Satz.Schrift != "Helvetica" {
		t.Errorf("Schrift %q, erwartet Helvetica", gesetzt.Satz.Schrift)
	}

	if gesetzt.Satz.Kursivschrift != "Helvetica-Oblique" {
		t.Errorf("Kursivschrift %q, erwartet Helvetica-Oblique", gesetzt.Satz.Kursivschrift)
	}
}

// **Ein Literal bleibt ein Literal** — 2.455 der 3.119 Angaben sind eines.
func TestEinSchriftnameDarfEinLiteralBleiben(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(namensfeld("Beleren Bold", "MPlantin-Italic")))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gesetzt := textebene(t, plan)
	if gesetzt.Satz.Schrift != "Beleren Bold" ||
		gesetzt.Satz.Kursivschrift != "MPlantin-Italic" {
		t.Errorf("Schrift %q, Kursiv %q", gesetzt.Satz.Schrift, gesetzt.Satz.Kursivschrift)
	}
}

// **Eine Rechnung, die scheitert, wird benannt** (§8) — und der rohe Name
// bleibt stehen, damit die Meldung der Schriftsuche sagt, wonach sie sucht.
func TestEinUnausrechenbarerSchriftnameWirdBenannt(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(namensfeld("{ gibtesnicht() }", "")))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl der Name nicht auszurechnen ist")
	}

	if !strings.Contains(befunde[0].Meldung, "ausrechnen") {
		t.Errorf("Befund %+v", befunde[0])
	}

	if gesetzt := textebene(t, plan); gesetzt.Satz.Schrift == "" {
		t.Error("die Ebene führt gar keinen Schriftnamen mehr")
	}
}

// **Ein Fehlerwert im Namen wird benannt.**
//
// `card.gibtsnicht` ist in MSE ein Wert, den `or else` auffangen darf; fängt
// die Vorlage ihn nicht, steht er hier — und die Meldung nennt den Zugriff.
func TestEinSchriftnameAusEinemFehlerwertWirdBenannt(t *testing.T) {
	_, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(namensfeld("{ card.gibtsnicht }", "")))

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl der Zugriff ins Leere ging")
	}

	if !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befund %+v", befunde[0])
	}
}

// **Nichts ist kein Name, aber auch kein Fehler** — dieselbe Regel wie bei der
// Farbe: `{ if … then "X" }` ohne `else` ergibt `nil`, und dann gilt keine
// Schrift.
func TestEinSchriftnameAusNichtsIstKeinBefund(t *testing.T) {
	plan, befunde := gedeutetMit(t, stilkarte("rot"),
		bildstil(namensfeld(`{ if card.farbe == "blau" then "Helvetica" }`, "")))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if gesetzt := textebene(t, plan); gesetzt.Satz.Schrift != "" {
		t.Errorf("Schrift %q, erwartet keine", gesetzt.Satz.Schrift)
	}
}
