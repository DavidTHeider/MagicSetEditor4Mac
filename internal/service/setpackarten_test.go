// **Die `pack type:`-Blöcke eines Sets werden gedeutet**
// (#1081).
//
// # Der Befund
//
// Bis hierher fielen sie in den Rohweg: sie überlebten den Rundlauf Byte für
// Byte und waren im Bestand **unsichtbar**. `Packbereich` kannte nur `pack
// type:` aus dem Spielpaket, und ein Set, das eine eigene Art nennt, zog aus
// der falschen — an allen vier Flächen.
//
// # Was in der Datei steht
//
// Dieselbe Gestalt wie im Spielpaket — `IMPLEMENT_REFLECTION(PackType)`,
// `src/data/pack.cpp`, Fassung f9d9356 vom 2021-07-06. Deshalb liest sie
// dieselbe Funktion (`msepaket.PackAus`); zwei Fassungen derselben Regel
// liefen auseinander.
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein eigener Booster wird gedeutet, samt Posten.**
func TestDiePackartenEinesSetsWerdenGedeutet(t *testing.T) {
	speicher := einfuhrAus(t, `mse version: 2.0.0
game: probe
pack type:
	name: mein booster
	selectable: true
	item:
		name: common
		amount: 10
	item:
		name: rare
		amount: 1
		weight: 0.5
pack type:
	name: nur filter
	filter: card.rarity == "rare"
card:
	name: Probe
`)

	if len(speicher.set.Packs) != 2 {
		t.Fatalf("%d Packarten: %+v", len(speicher.set.Packs), speicher.set.Packs)
	}

	erste := speicher.set.Packs[0]
	if erste.Name != "mein booster" || !erste.Selectable {
		t.Errorf("die erste ist %+v", erste)
	}

	// **`auto` wird aufgelöst**: mit Posten heißt es `all`
	// (`after_reading(PackType&)`, `src/data/pack.cpp`, Zeilen 43–46,
	// Fassung f9d9356 vom 2021-07-06).
	if erste.Select != "all" {
		t.Errorf("die Auswahl ist %q statt »all«", erste.Select)
	}

	if len(erste.Items) != 2 {
		t.Fatalf("%d Posten: %+v", len(erste.Items), erste.Items)
	}

	if erste.Items[0].Name != "common" || erste.Items[0].Amount != 10 ||
		erste.Items[0].Weight != 1 {
		t.Errorf("der erste Posten ist %+v", erste.Items[0])
	}

	if erste.Items[1].Weight != 0.5 {
		t.Errorf("das Gewicht des zweiten ist %v statt 0,5", erste.Items[1].Weight)
	}

	// **Mit Filter und ohne Posten heißt `auto` dagegen `no replace`.**
	if zweite := speicher.set.Packs[1]; zweite.Select != "no replace" {
		t.Errorf("die zweite ist %+v", zweite)
	}
}

// **Und sie liegen danach nicht mehr im Rohweg.**
//
// **Das ist die Hälfte, die man vergisst.** Ein Block, der gedeutet **und**
// roh mitreist, käme beim Schreiben zweimal heraus.
func TestEineGedeutetePackartLiegtNichtMehrRoh(t *testing.T) {
	speicher := einfuhrAus(t, `mse version: 2.0.0
game: probe
pack type:
	name: mein booster
card:
	name: Probe
`)

	for _, block := range speicher.roh {
		if strings.Contains(block.Text, "mein booster") {
			t.Errorf("die Packart liegt noch roh: %+v", block)
		}
	}
}

// **Ein zweiter `pack type:` mit demselben Namen bleibt erhalten.**
//
// Die Vorrangregel entscheidet erst beim Ziehen, welcher gilt
// (`PackartenMitVorrang`); der Bestand führt beide, weil die Datei beide
// führt — und beim Schreiben kommen beide heraus.
func TestZweiGleichnamigePackartenBleibenBeide(t *testing.T) {
	speicher := einfuhrAus(t, `mse version: 2.0.0
game: probe
pack type:
	name: booster
	filter: erste
pack type:
	name: booster
	filter: zweite
card:
	name: Probe
`)

	if len(speicher.set.Packs) != 2 {
		t.Fatalf("%d Packarten statt zwei: %+v", len(speicher.set.Packs), speicher.set.Packs)
	}
}

// **Ein Set mit eigener Packart kommt Byte für Byte zurück.**
//
// # Warum das eigens geprüft wird
//
// Seit `pack type` in `gespeichert` steht, fällt der Block **nicht mehr** in
// `raw_blocks`. Die Byte-Zusage hängt nicht daran — der Schreiber zerteilt die
// **Quelldatei** und nicht den Bestand (0011) —, aber genau das ist eine
// Behauptung, solange niemand sie prüft. Ohne diesen Fall wäre der Verlust
// eines ganzen Blocks bei jedem Speichern grün durchgegangen.
func TestEinSetMitEigenerPackartBleibtByteGleich(t *testing.T) {
	quelle := "\ufeffmse version: 2.0.0\r\ngame: probe\r\npack type:\r\n" +
		"\tname: mein booster\r\n\titem:\r\n\t\tname: common\r\n\t\tamount: 10\r\n" +
		"card:\r\n\tname: Probe"

	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(quelle), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), ort); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// **Die Quelle reist mit** — das ist die Grundlage der Zusage.
	if speicher.set.Quelle == nil {
		t.Fatal("keine Quelle — die Byte-Zusage ist damit nicht zu halten")
	}

	baum, err := msepaket.Zerteile(speicher.set.Quelle.Datei, speicher.set.Quelle.Inhalt)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	if heraus := string(msepaket.Schreibe(baum)); heraus != quelle {
		t.Errorf("heraus kam\n%q\nerwartet\n%q", heraus, quelle)
	}
}
