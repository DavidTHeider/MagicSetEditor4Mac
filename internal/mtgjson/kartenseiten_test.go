// **Mehrseitige Karten werden zu einer Karte mit zwei Feldsätzen** —
// Befund 4 des Zweitblicks über Phase V, 2026-08-24.
//
// # Woran es belegt ist
//
// Zwei unabhängige Quellen, beide am 2026-08-24 nachgelesen:
//
//   - `magic.mse-game/card_fields` definiert 249 Kartenfelder, davon **46 mit
//     dem Zusatz ` 2`** und **keines mit ` 3`**.
//   - `magic-set-generator` (MIT, §11) legt in `src/mse.rs` über das Makro
//     `push_alt!` beide Seiten in **dieselbe** Karte und hängt bei der zweiten
//     ` 2` an jeden Schlüssel.
//
// **Eine eigene Datei**, weil `mehrseitig_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, wie eine **Seite**
// gelesen wird, hier, wie die Seiten **zusammenkommen**.
package mtgjson_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

func TestZweiSeitenWerdenEineKarteMitZweitemFeldsatz(t *testing.T) {
	t.Parallel()

	inhalt := `{"meta":{"version":"5.2.2"},"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Wear // Tear","faceName":"Wear","side":"a","layout":"split",
		 "manaCost":"{1}{R}","types":["Instant"],"text":"Zerstöre ein Artefakt.",
		 "otherFaceIds":["zweite-seite"],"uuid":"erste-seite"},
		{"name":"Wear // Tear","faceName":"Tear","side":"b","layout":"split",
		 "manaCost":"{W}","types":["Instant"],"text":"Zerstöre eine Verzauberung.",
		 "otherFaceIds":["erste-seite"],"uuid":"zweite-seite"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	// **Eine Karte, nicht zwei.**
	if len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1 — MSE führt eine mehrseitige Karte als "+
			"eine Karte mit zwei Feldsätzen", len(quelle.Sets[0].Karten))
	}

	werte := quelle.Sets[0].Karten[0].Werte

	erwartet := map[string]string{
		"name":           "Wear",
		"casting cost":   "1R",
		"rule text":      "Zerstöre ein Artefakt.",
		"name 2":         "Tear",
		"casting cost 2": "W",
		"rule text 2":    "Zerstöre eine Verzauberung.",
	}

	for feld, soll := range erwartet {
		ist, da := wertVon(werte, feld)
		if !da || ist != soll {
			t.Errorf("%q = %q (da: %v), erwartet %q", feld, ist, da, soll)
		}
	}
}

// **Die Reihenfolge folgt `side`, nicht der Reihenfolge in der Datei.**
//
// MTGJSON sichert keine Reihenfolge der Kartenobjekte zu. Wer die
// Dateireihenfolge nähme, bekäme bei mancher Datei die Rückseite vorn — und
// die Vorlage zeichnete sie als Vorderseite.
func TestDieSeitenfolgeKommtAusSideUndNichtAusDerDatei(t *testing.T) {
	t.Parallel()

	// **`b` steht zuerst in der Datei.**
	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"A // B","faceName":"Rückseite","side":"b","layout":"transform",
		 "otherFaceIds":["eins"],"uuid":"zwei"},
		{"name":"A // B","faceName":"Vorderseite","side":"a","layout":"transform",
		 "otherFaceIds":["zwei"],"uuid":"eins"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten", len(quelle.Sets[0].Karten))
	}

	werte := quelle.Sets[0].Karten[0].Werte

	if name, _ := wertVon(werte, "name"); name != "Vorderseite" {
		t.Errorf("name = %q, erwartet die Seite mit side=\"a\"", name)
	}

	if name, _ := wertVon(werte, "name 2"); name != "Rückseite" {
		t.Errorf("name 2 = %q", name)
	}
}

// **Eine Verbindung, die kein Vorder-Rückseiten-Paar ist, verliert nichts.**
//
// # Was hier bis zum 2026-08-24 stand
//
// Ein Lauf über drei Objekte, von denen jedes auf die beiden anderen zeigt —
// und die Zusage, die dritte Seite werde gemeldet und verworfen. Die Gestalt
// war **erfunden**: sie kommt in MTGJSON nicht vor. Gemessen an `EMN.json`
// (mtgjson.com/api/v5, 2026-08-24) über alle mehrseitigen Karten:
//
//	layout      side  otherFaceIds  Objekte
//	meld        a     1             6
//	meld        b     2             3
//	transform   a     1             15
//	transform   b     1             15
//
// **Eine Vorderseite nennt genau eine andere Seite.** Wer zwei oder mehr
// nennt, ist eine geteilte Rückseite — und die wird die zweite Seite **jeder**
// Vorderseite, die auf sie zeigt (#1038).
//
// **Was mit einer unbekannten Gestalt geschieht**, steht deshalb hier: sie
// wird nicht gruppiert, und **keine Karte geht verloren**. Das ist der
// vorsichtigere Ausgang — eine erfundene Zuordnung wäre schlechter als keine.
func TestEineUnbekannteVerbindungVerliertNichts(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"A // B // C","faceName":"Eins","side":"a","otherFaceIds":["z","d"],"uuid":"e"},
		{"name":"A // B // C","faceName":"Zwei","side":"b","otherFaceIds":["e","d"],"uuid":"z"},
		{"name":"A // B // C","faceName":"Drei","side":"c","otherFaceIds":["e","z"],"uuid":"d"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if len(quelle.Sets[0].Karten) != 3 {
		t.Fatalf("%d Karten, erwartet drei — keine darf verschwinden",
			len(quelle.Sets[0].Karten))
	}

	gesehen := map[string]bool{}
	for _, karte := range quelle.Sets[0].Karten {
		name, _ := wertVon(karte.Werte, "name")
		gesehen[name] = true
	}

	for _, name := range []string{"Eins", "Zwei", "Drei"} {
		if !gesehen[name] {
			t.Errorf("%q fehlt", name)
		}
	}
}
