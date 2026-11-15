package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Feld bringt seine Aufschrift mit — und ihre Übersetzungen.**
//
//	set field:
//		type: text
//		name: title
//		identifying: true
//		localized caption:
//			ch-s: 标题
//			ch-t: 標題
//
// `magic.mse-game/set_fields`, Zeilen 7–13. Über `M15-Magic-Pack` stehen
// **zehn** `localized caption:` — alle an Setfeldern, alle nur auf Chinesisch;
// die Kartenfelder tragen keine. Deshalb zeigt **auch MSE2** dort englische
// Wörter, und das ist die Antwort auf „warum sind die Label nicht übersetzt".
//
// MSE hält beides als eine `LocalizedString` (`src/data/field.hpp`, Zeile 47,
// Fassung f9d9356 vom 2021-07-06): `caption:` ist der Vorgabewert,
// `localized caption:` die Einträge je Sprache.
func TestEinFeldNenntSeineAufschrift(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: pt
	caption: P/T
card field:
	type: text
	name: titel
	localized caption:
		de: Titel
		ch-s: 标题
card field:
	type: text
	name: casting cost
card field:
	type: text
	name: leer
	localized caption:
		# hier stand einmal eine Sprache

card field:
	type: text
	name: mit kommentar
	localized caption:
		# die alte Fassung hatte hier ch-t
		de: Mit Kommentar
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

	spiel, befunde := paket.Spielmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(spiel.Fields) != 5 {
		t.Fatalf("%d Felder", len(spiel.Fields))
	}

	if gab := spiel.Fields[0].Beschriftung; gab != "P/T" {
		t.Errorf("`caption:` kam als %q an", gab)
	}

	titel := spiel.Fields[1]
	if gab := titel.Uebersetzungen["de"]; gab != "Titel" {
		t.Errorf("die deutsche Aufschrift kam als %q an", gab)
	}

	if gab := titel.Uebersetzungen["ch-s"]; gab != "标题" {
		t.Errorf("die chinesische Aufschrift kam als %q an", gab)
	}

	// **Ohne Angabe bleibt beides leer** — abgeleitet wird in der Domäne, nicht
	// hier: der Leser sagt, was dasteht.
	ohne := spiel.Fields[2]
	if ohne.Beschriftung != "" || ohne.Uebersetzungen != nil {
		t.Errorf("ein Feld ohne Angabe trägt %q / %v",
			ohne.Beschriftung, ohne.Uebersetzungen)
	}

	if gab := ohne.Beschriftet(""); gab != "Casting cost" {
		t.Errorf("abgeleitet wurde %q", gab)
	}

	// **Ein Block ohne Sprache ist kein Block.** MSE fragt `translations.empty()`
	// vor der Sprache (`src/data/locale.cpp`, Zeilen 138–141, Fassung f9d9356
	// vom 2021-07-06); eine leere Abbildung wäre dasselbe wie keine, sähe im
	// Test aber anders aus — und irgendwann verließe sich jemand darauf.
	if leer := spiel.Fields[3]; leer.Uebersetzungen != nil {
		t.Errorf("ein Block ohne Sprache ergab %v", leer.Uebersetzungen)
	}

	// **Ein Kommentar ist keine Sprache.** Er steht als Kind im Baum, und wer
	// jedes Kind nähme, legte eine Sprache mit leerem Kürzel an — die dann bei
	// keiner Oberflächensprache trifft und trotzdem den Block nichtleer macht.
	kommentiert := spiel.Fields[4]
	if len(kommentiert.Uebersetzungen) != 1 ||
		kommentiert.Uebersetzungen["de"] != "Mit Kommentar" {
		t.Errorf("mit Kommentar gelesen: %v", kommentiert.Uebersetzungen)
	}
}
