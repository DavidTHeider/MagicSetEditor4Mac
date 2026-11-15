package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Maß darf ein eingerückter `script:`-Block sein.**
//
// `magic-modules.mse-include/stamps/card_fields` schreibt jede der vier Angaben
// so:
//
//	left:
//		script:
//			width := face_coordinates_map(1).width
//			…
//
// Das ist dieselbe Form wie bei `alignment:` — der Wert des Knotens ist leer,
// und das Skript steht in seinen Zeilen. Wer nur den Wert liest, bekommt
// nichts; die Ebene fiel damit auf die Maße der **ganzen Karte** zurück und
// zeichnete ein 43 Punkte großes Prägesiegel über die volle Karte
// (#295).
func TestEinMassDarfEinSkriptblockSein(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	name:
		left: 5
		top:
			script:
				7 + 1
		width: 40
		height:
			script:
				9
		right:
			script:
				50
		bottom:
			script:
				60
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile, erwartet einen", len(stil.Felder))
	}

	feld := stil.Felder[0]
	for name, haben := range map[string]string{
		"top": feld.Oben, "height": feld.Hoehe, "right": feld.Rechts, "bottom": feld.Unten,
	} {
		if strings.TrimSpace(haben) == "" {
			t.Errorf("%s ist leer — der Skriptblock wurde nicht gelesen", name)
		}
	}
}

// **Ein leerer Block ist kein Maß.**
//
// `left:` ohne Wert und ohne `script:` darunter bleibt leer — und ein
// `script:`, das nichts enthält, ebenso. Erfunden wird nichts; wer hier
// geschweifte Klammern um nichts setzte, machte aus einer fehlenden Angabe
// einen Skriptfehler.
func TestEinLeererBlockIstKeinMass(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	name:
		left:
		top:
			script:
		width: 40
		height: 9
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile, erwartet einen", len(stil.Felder))
	}

	if links := strings.TrimSpace(stil.Felder[0].Links); links != "" {
		t.Errorf("links ist %q, erwartet leer", links)
	}

	if oben := strings.TrimSpace(stil.Felder[0].Oben); oben != "" {
		t.Errorf("oben ist %q, erwartet leer", oben)
	}
}

// **Ein Stil darf eigene Felder mitbringen — `extra card field:`.**
//
//	REFLECT(extra_card_fields);
//	REFLECT_IF_READING {
//	  if (extra_card_style.init(extra_card_fields)) { … }
//	}
//	REFLECT(extra_card_style);
//
// `src/data/stylesheet.cpp`, Zeilen 116–125, Fassung f9d9356 vom 2021-07-06.
//
// **Daran hängt der Kasten für Stärke und Widerstandskraft.** `magic-m15`
// zeichnet ihn nicht über ein Kartenfeld, sondern über ein eigenes:
//
//	extra card field:
//		type: choice
//		name: pt box
//		script: card.card_color
//	extra card style:
//		pt box:
//			left: 273 … image: { module_ptbox() }
//
// Ohne sie fehlt der Kasten auf **jeder** Kreatur, und `extra_card.…` ist ein
// unbekannter Name — 232-mal über die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestEinStilDarfEigeneFelderMitbringen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
extra card field:
	type: choice
	name: pt box
	script: card.card_color
	save value: false
extra card style:
	pt box:
		left: 73
		top: 66
		width: 21
		height: 12
		z index: 1
		render style: image
		image: { "kasten.png" }
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Zusatzfelder) != 1 || stil.Zusatzfelder[0].Name != "pt box" {
		t.Fatalf("Zusatzfelder: %+v", stil.Zusatzfelder)
	}

	if stil.Zusatzfelder[0].Skript != "card.card_color" {
		t.Errorf("das Skript ist %q", stil.Zusatzfelder[0].Skript)
	}

	if len(stil.Zusatzstile) != 1 || stil.Zusatzstile[0].Feld != "pt box" {
		t.Fatalf("Zusatzstile: %+v", stil.Zusatzstile)
	}

	if stil.Zusatzstile[0].Links != "73" {
		t.Errorf("links ist %q", stil.Zusatzstile[0].Links)
	}
}

// **Ein Zusatzfeld ohne Namen ist ein Befund.**
//
// Dieselbe Regel wie bei den Kartenfeldern (§8): was sich nicht zuordnen lässt,
// wird benannt und übergangen — der Rest der Vorlage bleibt lesbar.
func TestEinZusatzfeldOhneNamenIstEinBefund(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
extra card field:
	type: choice
extra card field:
	type: choice
	name: heil
extra card style:
	#### ein Kommentar ist kein Eintrag
	zweiter:
		left: 2
		z index: 1
	heil:
		left: 1
		z index: 5
	dritter:
		image: { 1 }
		z index: keine Zahl
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	// **Zwei Befunde**: das Feld ohne Namen und der Zusatzstil mit einer
	// `z index`-Angabe, die keine Zahl ist.
	stil, befunde := paket.Stilmodell()
	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet zwei: %v", len(befunde), befunde)
	}

	if len(stil.Zusatzfelder) != 1 || stil.Zusatzfelder[0].Name != "heil" {
		t.Errorf("Zusatzfelder: %+v", stil.Zusatzfelder)
	}

	// **Und die Zusatzstile stehen nach `z index`** — derselbe Grund wie bei
	// den Kartenfeldern: der Kasten für Stärke und Widerstandskraft liegt
	// unter dem Rahmen, nicht darüber.
	if len(stil.Zusatzstile) != 3 {
		t.Fatalf("%d Zusatzstile, erwartet drei", len(stil.Zusatzstile))
	}

	if stil.Zusatzstile[0].Feld != "dritter" || stil.Zusatzstile[2].Feld != "heil" {
		t.Errorf("Reihenfolge: %s, %s, %s",
			stil.Zusatzstile[0].Feld, stil.Zusatzstile[1].Feld, stil.Zusatzstile[2].Feld)
	}
}

// **`visible:` wird gelesen — einzeilig wie als Block.**
//
// `magic-m15` schreibt am Kasten für Stärke und Widerstandskraft
// `visible: { card.pt != "" }`; `magic-modules` schreibt es an anderen Feldern
// als eingerückten `script:`-Block. Beide Formen meinen dasselbe
// (#270).
func TestVisibleWirdGelesen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	eins:
		left: 1
		visible: { card.pt != "" }
	zwei:
		left: 2
		visible:
			script:
				false
	drei:
		left: 3
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 3 {
		t.Fatalf("%d Feldstile", len(stil.Felder))
	}

	nach := map[string]string{}
	for _, feld := range stil.Felder {
		nach[feld.Feld] = strings.TrimSpace(feld.Sichtbar)
	}

	if nach["eins"] != `{ card.pt != "" }` {
		t.Errorf("einzeilig: %q", nach["eins"])
	}

	if !strings.Contains(nach["zwei"], "false") {
		t.Errorf("als Block: %q", nach["zwei"])
	}

	// **Ohne Angabe bleibt es leer** — und der Planbauer zeichnet dann.
	if nach["drei"] != "" {
		t.Errorf("ohne Angabe: %q", nach["drei"])
	}
}

// **`paragraph height:` wird gelesen.**
//
//	REFLECT(paragraph_height);
//
// `IMPLEMENT_REFLECTION(TextStyle)`, `src/data/field/text.cpp`, Zeile 164,
// Fassung f9d9356 vom 2021-07-06. Die Vorlagen lesen sie über
// `card_style.<feld>.paragraph_height`
// (#270).
func TestParagraphHeightWirdGelesen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	text:
		left: 1
		paragraph height: 7.5
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile", len(stil.Felder))
	}

	if haben := strings.TrimSpace(stil.Felder[0].Zeilenhoehen.Absatz); haben != "7.5" {
		t.Errorf("die Absatzhöhe ist %q, erwartet 7.5", haben)
	}
}
