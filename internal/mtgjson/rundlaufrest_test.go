package mtgjson_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Ein unbekanntes Feld der zweiten Seite überlebt den Rundlauf** —
// #1039.
//
// # Die Zusage, um die es geht
//
// `Karte.Rest` sagt zu: „die Schlüssel, die sie nicht kennt — **roh
// mitgeführt**, damit ein Rundlauf sie nicht verliert." Beim Zusammenlegen
// zweier Seiten wurde bis zum 2026-08-24 nur der Rest der **ersten** Seite
// übernommen; alles Unbekannte der zweiten verschwand — **ohne Befund**, und
// das ist der eigentliche Bruch (§8: ein Verlust wird benannt).
//
// # Warum die Prüfung über beide Richtungen geht
//
// Eine Prüfung nur am gelesenen Zwischenstand zeigte, dass der Wert **da**
// ist, nicht dass er wieder **hinausgeht** — und der Weg dorthin führt durch
// `TrenneKartenwerte`, das den Zusatz ` 2` wieder abnimmt. Erst beide
// Richtungen zusammen belegen die Zusage.
func TestEinUnbekanntesFeldJederSeiteUeberlebtDenRundlauf(t *testing.T) {
	t.Parallel()

	hinein := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Wear // Tear","faceName":"Wear","side":"a","layout":"split",
		 "manaCost":"{1}{R}","otherFaceIds":["zwei"],"uuid":"eins",
		 "vornEigen":"A"},
		{"name":"Wear // Tear","faceName":"Tear","side":"b","layout":"split",
		 "manaCost":"{W}","otherFaceIds":["eins"],"uuid":"zwei",
		 "hintenEigen":"B"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(hinein), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	roh, _ := mtgjson.ZuordnungMagic.Schreibe(alsDomaene(quelle)[0])
	karten := geschriebeneKarten(t, roh)

	if len(karten) != 2 {
		t.Fatalf("%d Kartenobjekte, erwartet 2", len(karten))
	}

	erste, zweite := karten[0], karten[1]

	if erste["vornEigen"] != "A" {
		t.Errorf("die erste Seite führt vornEigen = %v, erwartet %q",
			erste["vornEigen"], "A")
	}

	// **Der Befund, um den es geht.** Bis zum 2026-08-24 stand hier `<nil>`.
	if zweite["hintenEigen"] != "B" {
		t.Errorf("die zweite Seite führt hintenEigen = %v, erwartet %q",
			zweite["hintenEigen"], "B")
	}

	// **Und nicht an der falschen Seite.** Ein Zusammenlegen, das alles auf
	// die erste Seite schriebe, bestünde die Zeile darüber nicht — aber ein
	// Trennen, das beide Seiten gleich befüllte, bestünde beide.
	if _, da := erste["hintenEigen"]; da {
		t.Errorf("die erste Seite führt das Feld der zweiten: %v", erste["hintenEigen"])
	}

	if _, da := zweite["vornEigen"]; da {
		t.Errorf("die zweite Seite führt das Feld der ersten: %v", zweite["vornEigen"])
	}
}
