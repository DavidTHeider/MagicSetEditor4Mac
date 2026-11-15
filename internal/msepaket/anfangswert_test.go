package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Auswahlfeld nennt seinen Anfangswert — `initial:`.**
//
// MSE vergibt den Wert eines Auswahlfeldes in drei Stufen:
//
//	value( !field->initial.empty() ? field->initial
//	     : initial_first_choice    ? field->choices->choiceName(0)
//	                               : String())
//
// `src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356 vom 2021-07-06.
//
// **Ohne `initial:` bekommt `rarity` die erste Wahl** — und die heißt in
// `magic.mse-game` `basic land`, während dort ausdrücklich `initial: common`
// steht. 17-mal im Pack
// (#341).
func TestEinAuswahlfeldNenntSeinenAnfangswert(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: choice
	name: rarity
	initial: common
	choice: basic land
	choice: common
card field:
	type: choice
	name: card symbol
	choice: none
	choice: tombstone
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

	if len(spiel.Fields) != 2 {
		t.Fatalf("%d Kartenfelder, erwartet zwei", len(spiel.Fields))
	}

	if spiel.Fields[0].Initial != "common" {
		t.Errorf("der Anfangswert ist %q, erwartet %q",
			spiel.Fields[0].Initial, "common")
	}

	// **Ohne die Angabe bleibt er leer**, nicht „die erste Wahl": welche Stufe
	// gilt, entscheidet der Aufrufer, und `msepaket` liest nur, was dasteht.
	if spiel.Fields[1].Initial != "" {
		t.Errorf("das zweite Feld nennt %q, erwartet keinen Anfangswert",
			spiel.Fields[1].Initial)
	}
}

// probespielpaket legt ein `.mse-game` mit dem gegebenen Inhalt an.
func probespielpaket(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-game")
	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pfad, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Ein Wertskript steht als eingerückter Block, nicht als Wert.**
//
// `magic.mse-game` schreibt es so, und `Knoten.Wert` ist dort leer — es ist
// dieselbe Form wie bei `init script:`
// (#345).
func TestEinWertskriptWirdAlsBlockGelesen(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: text
	script:
		combined_editor(field1: card.rule_text, separator: "-", field2: card.flavor_text)
card field:
	type: text
	name: ohne
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

	spiel, _ := paket.Spielmodell()

	if len(spiel.Fields) != 2 {
		t.Fatalf("%d Kartenfelder, erwartet zwei", len(spiel.Fields))
	}

	if !strings.Contains(spiel.Fields[0].Skript, "combined_editor") {
		t.Errorf("das Wertskript ist %q", spiel.Fields[0].Skript)
	}

	// **Ohne `script:` bleibt es leer** — und nicht etwa der Name des Feldes.
	if spiel.Fields[1].Skript != "" {
		t.Errorf("das zweite Feld nennt %q, erwartet kein Wertskript", spiel.Fields[1].Skript)
	}
}

// **Ein Wertskript steht einzeilig oder als Block — und der Kommentar darunter
// gehört zu keinem von beiden.**
//
// `magic.mse-game` schreibt beide Formen, und auf die einzeilige folgen
// regelmäßig Kommentarzeilen (`##### Name line`). Wer den Block bevorzugte,
// bekäme sie als Skript und daraus „hier fehlt ein Ausdruck" — 3.990-mal über
// die 190 Stile von `M15-Magic-Pack`
// (#345).
func TestEinWertskriptStehtEinzeiligOderAlsBlock(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: name
	script: to_upper(value)
	##### Der Name einer Karte
card field:
	type: text
	name: type
	script:
		to_lower(value)
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

	erwartet := []string{"to_upper(value)", "to_lower(value)"}
	for stelle, wollen := range erwartet {
		if haben := strings.TrimSpace(spiel.Fields[stelle].Skript); haben != wollen {
			t.Errorf("Feld %d hat das Skript %q, erwartet %q", stelle, haben, wollen)
		}
	}
}
