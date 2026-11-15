package mtgjson_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// Ein Feld, das die Zuordnung nicht kennt, wird gemeldet **und mitgeführt**.
//
// Gemeldet, damit jemand die Zuordnung ergänzen kann; mitgeführt, damit der
// Rundlauf es nicht verliert
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md)). Nur
// eines von beidem wäre halb: eine Meldung ohne den Wert ist eine Nachricht
// über etwas, das schon fort ist.
func TestUnbekanntesFeldWirdGemeldetUndMitgefuehrt(t *testing.T) {
	pfad := filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json")

	quelle, err := mtgjson.Lies(pfad, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	// **Ein Befund je Set, nicht je Karte** — der Grund steht in
	// `buendeln_test.go` und kommt aus dem Lauf gegen echte Daten.
	erwartet := domain.Befund{
		Datei:   pfad,
		Ort:     "PRB → uuid",
		Meldung: "die Zuordnung für »magic« kennt dieses Feld nicht — 2 Karten führen es, zuerst Karte 1",
	}

	if len(quelle.Befunde) != 1 {
		t.Fatalf("Befunde: %v, erwartet genau einen", quelle.Befunde)
	}

	if quelle.Befunde[0] != erwartet {
		t.Errorf("Befund: %+v, erwartet %+v", quelle.Befunde[0], erwartet)
	}

	// Der Wert bleibt an der Karte — als **rohes JSON**, nicht als gedeuteter
	// Text: nur so lässt sich beim Schreiben `"1"` von `1` unterscheiden.
	rest := quelle.Sets[0].Karten[0].Rest
	if len(rest) != 1 {
		t.Fatalf("Rest: %v, erwartet einen Eintrag", rest)
	}

	erwarteterWert := domain.Value{Field: "uuid", Text: `"00000000-0000-0000-0000-000000000001"`}
	if rest[0] != erwarteterWert {
		t.Errorf("Rest: %+v, erwartet %+v", rest[0], erwarteterWert)
	}
}

// Ein übergangenes Feld ist **kein** Befund.
//
// `type` und `number` stehen in der Zuordnung mit einem Grund statt einem Ziel
// (MSE leitet sie ab). Sie zu melden hieße, über eine getroffene Entscheidung
// zu klagen — und der Befundliste jede Aussagekraft zu nehmen: wer bei jedem
// Import zwei Meldungen sieht, die nichts bedeuten, liest die dritte nicht mehr.
func TestUebergangenesFeldGibtKeinenBefund(t *testing.T) {
	quelle, err := mtgjson.Lies(filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	for _, befund := range quelle.Befunde {
		for _, uebergangen := range []string{"type", "number"} {
			if befund.Ort == "PRB → "+uebergangen {
				t.Errorf("»%s« ist gemeldet worden: %s", uebergangen, befund)
			}
		}
	}

	for _, karte := range quelle.Sets[0].Karten {
		for _, wert := range karte.Rest {
			if wert.Field == "type" || wert.Field == "number" {
				t.Errorf("»%s« liegt im Rest, obwohl die Zuordnung es kennt", wert.Field)
			}
		}
	}
}

// Jedes übergangene Feld nennt seinen Grund.
//
// **Auch eine Ablehnung braucht Deckung** (§10). Ein Feldpaar ohne Ziel und ohne
// Grund sähe aus wie eine Entscheidung und wäre ein Vergessen — und niemand
// könnte später nachlesen, warum das Feld nicht ankommt.
func TestJedesUebergangeneFeldNenntSeinenGrund(t *testing.T) {
	for _, paar := range mtgjson.ZuordnungMagic.Felder {
		if paar.Feld == "" && paar.Grund == "" {
			t.Errorf("%v: übergangen, ohne einen Grund zu nennen", paar.JSON)
		}

		if paar.Feld != "" && paar.Grund != "" {
			t.Errorf("%v: nennt einen Grund und ordnet trotzdem zu", paar.JSON)
		}
	}
}
