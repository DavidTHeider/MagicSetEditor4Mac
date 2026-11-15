// **Meld: zwei Vorderseiten, eine geteilte Rückseite** —
// #1038.
package mtgjson_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// meldsatz schreibt eine Datei in der Gestalt, die **echte** MTGJSON-Daten
// haben.
//
// # Woher die Gestalt stammt
//
// **Gemessen am 2026-08-24 an `EMN.json`** (mtgjson.com/api/v5), über alle
// mehrseitigen Karten des Satzes:
//
//	layout      side  otherFaceIds  Objekte
//	meld        a     1             6
//	meld        b     2             3
//	transform   a     1             15
//	transform   b     1             15
//
// **Daraus die Regel**: wer **zwei oder mehr** `otherFaceIds` führt, ist eine
// **geteilte Rückseite**. Sie wird keine eigene Karte, sondern die zweite
// Seite **jeder** Vorderseite, die auf sie zeigt.
//
// Genau so hält es die Referenz: `magic-set-generator` (MIT, §11) macht aus
// `Layout::Meld { top, bottom, .. }` zwei Karten — `src/lib.rs`, Zeilen
// 275–279 — und hängt jeder den `back` als zweite Seite an.
func meldsatz(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "meld.json")

	inhalt := `{
	  "meta": {"date": "2026-08-24", "version": "5.2.2"},
	  "data": {
	    "MLD": {
	      "code": "MLD", "name": "Melder",
	      "cards": [
	        {"uuid": "a1", "side": "a", "layout": "meld", "otherFaceIds": ["b1"],
	         "name": "Erste Hälfte // Das Ganze", "faceName": "Erste Hälfte",
	         "manaCost": "{2}{W}", "type": "Creature — Angel", "rarity": "rare"},
	        {"uuid": "a2", "side": "a", "layout": "meld", "otherFaceIds": ["b1"],
	         "name": "Zweite Hälfte // Das Ganze", "faceName": "Zweite Hälfte",
	         "manaCost": "{3}{W}", "type": "Creature — Angel", "rarity": "rare"},
	        {"uuid": "b1", "side": "b", "layout": "meld", "otherFaceIds": ["a1", "a2"],
	         "name": "Das Ganze", "faceName": "Das Ganze",
	         "type": "Creature — Eldrazi Angel", "rarity": "rare"}
	      ]
	    }
	  }
	}`

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Aus drei Objekten werden zwei Karten, nicht eine.**
//
// Bis hierher wurden alle drei zu **einer** gruppiert und ab dem dritten
// verworfen: eine spielbare Vorderseite verschwand vollständig.
func TestMeldWirdZuZweiKarten(t *testing.T) {
	t.Parallel()

	zuordnung, bekannt := mtgjson.ZuordnungFuer("magic")
	if !bekannt {
		t.Fatal("für magic gibt es keine Zuordnung")
	}

	quelle, fehler := mtgjson.Lies(meldsatz(t), zuordnung)
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if len(quelle.Sets) != 1 {
		t.Fatalf("%d Sets", len(quelle.Sets))
	}

	karten := quelle.Sets[0].Karten
	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei — jede Vorderseite ist eine Karte", len(karten))
	}

	ersterName, _ := wertVon(karten[0].Werte, "name")
	zweiterName, _ := wertVon(karten[1].Werte, "name")
	namen := []string{ersterName, zweiterName}
	sortiere(namen)

	if namen[0] != "Erste Hälfte" || namen[1] != "Zweite Hälfte" {
		t.Errorf("Namen = %v — beide Vorderseiten müssen da sein", namen)
	}
}

// **Und beide tragen dieselbe Rückseite.**
//
// Das ist die zweite Hälfte der Referenz: `vec![top, bottom]`, und jede
// bekommt den `back` als zweite Seite.
func TestBeideMeldkartenTragenDieselbeRueckseite(t *testing.T) {
	t.Parallel()

	zuordnung, _ := mtgjson.ZuordnungFuer("magic")

	quelle, fehler := mtgjson.Lies(meldsatz(t), zuordnung)
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	for _, karte := range quelle.Sets[0].Karten {
		rueckseite, _ := wertVon(karte.Werte, "name 2")
		if rueckseite != "Das Ganze" {
			vorderseite, _ := wertVon(karte.Werte, "name")
			t.Errorf("»name 2« von %q = %q, erwartet die geteilte Rückseite",
				vorderseite, rueckseite)
		}
	}
}

// **Die geteilte Rückseite wird keine eigene Karte.**
//
// Sonst stünde sie dreimal im Set: einmal für sich und einmal an jeder
// Vorderseite.
func TestDieGeteilteRueckseiteStehtNichtFuerSich(t *testing.T) {
	t.Parallel()

	zuordnung, _ := mtgjson.ZuordnungFuer("magic")

	quelle, fehler := mtgjson.Lies(meldsatz(t), zuordnung)
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	for _, karte := range quelle.Sets[0].Karten {
		if name, _ := wertVon(karte.Werte, "name"); name == "Das Ganze" {
			t.Error("die geteilte Rückseite steht als eigene Karte im Set")
		}
	}
}

// sortiere ordnet zwei Namen.
func sortiere(namen []string) {
	if len(namen) == 2 && namen[0] > namen[1] {
		namen[0], namen[1] = namen[1], namen[0]
	}
}

// **Ein Verweis ins Leere wird gemeldet, nicht verschluckt** (§8).
//
// Er kommt vor, wenn eine Datei nur einen Teil eines Satzes führt: die
// Vorderseite ist da, die Rückseite nicht. Die Karte bleibt — aber wer sie
// später als einseitig ansieht, soll wissen, warum.
func TestEinVerweisInsLeereWirdGemeldet(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Halb","faceName":"Halb","side":"a","otherFaceIds":["fehlt"],"uuid":"e"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten, erwartet eine — sie bleibt", len(quelle.Sets[0].Karten))
	}

	gefunden := false

	for _, befund := range quelle.Befunde {
		if strings.Contains(befund.Meldung, "liegt nicht in dieser Datei") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("kein Befund über den offenen Verweis: %+v", quelle.Befunde)
	}
}

// **Auch wenn die geteilte Rückseite zuerst in der Datei steht.**
//
// # Warum dieser Lauf da ist
//
// **Weil eine Gegenprobe grün blieb.** Die Bedingung „eine Vorderseite nennt
// **genau eine** andere Seite" wurde zu „mindestens eine" aufgeweicht — und
// die Läufe darüber blieben grün: in ihrer Datei steht die Rückseite **hinten**,
// und die Gegenseitigkeitsprüfung fängt sie dort ohnehin ab.
//
// **MTGJSON sichert keine Reihenfolge zu** — der Kopf von `nachSeite` sagt es
// selbst. Steht die Rückseite vorn, entschied allein die aufgeweichte
// Bedingung, und sie machte aus der Rückseite eine Karte mit einer
// Vorderseite als zweiter Seite.
//
// Eine Gegenprobe, die nichts ändert, ist keine (§10).
func TestDieGeteilteRueckseiteDarfZuerstStehen(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"MLD","name":"Melder","cards":[
		{"uuid":"b1","side":"b","otherFaceIds":["a1","a2"],
		 "name":"Das Ganze","faceName":"Das Ganze","type":"Creature","rarity":"rare"},
		{"uuid":"a1","side":"a","otherFaceIds":["b1"],
		 "name":"Erste Hälfte // Das Ganze","faceName":"Erste Hälfte",
		 "manaCost":"{2}{W}","type":"Creature","rarity":"rare"},
		{"uuid":"a2","side":"a","otherFaceIds":["b1"],
		 "name":"Zweite Hälfte // Das Ganze","faceName":"Zweite Hälfte",
		 "manaCost":"{3}{W}","type":"Creature","rarity":"rare"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	karten := quelle.Sets[0].Karten
	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei", len(karten))
	}

	for _, karte := range karten {
		name, _ := wertVon(karte.Werte, "name")
		rueckseite, _ := wertVon(karte.Werte, "name 2")

		if name == "Das Ganze" {
			t.Error("die geteilte Rückseite wurde eine eigene Karte")
		}

		if rueckseite != "Das Ganze" {
			t.Errorf("»name 2« von %q = %q", name, rueckseite)
		}
	}
}
