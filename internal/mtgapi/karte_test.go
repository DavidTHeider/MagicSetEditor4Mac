package mtgapi_test

// **Die Zuordnung von MTG-API-Karten auf `domain.Card`** — Werkstück 3 von
// Phase 4.8 (#1321).
//
// **Gegen eingecheckte, selbst geschriebene Beispiele** (§16): sie stehen in
// `testdaten/` und tragen erfundene Karten. Ein echter Kartentext gehört
// nicht ins Repositorium (§11), und für die Zuordnung tut er nichts zur
// Sache — sie liest Feldnamen, nicht Bedeutungen.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
)

// beispiel liest eine Datei aus `testdaten/`.
func beispiel(t *testing.T, name string) []byte {
	t.Helper()

	//nolint:gosec // der Name kommt aus dem Lauf, nicht von außen
	roh, err := os.ReadFile(filepath.Join("testdaten", name))
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return roh
}

// **Eine Karte wird gelesen, und jedes Feld kommt an.**
func TestEineKarteWirdGelesen(t *testing.T) {
	t.Parallel()

	karten, befunde, err := mtgapi.LiesKarten(beispiel(t, "karte.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}

	eine := karten[0]

	for _, fall := range []struct {
		feld     string
		gelesen  string
		erwartet string
	}{
		{"Name", eine.Name, "Blitzschlag"},
		{"Kennung", eine.Kennung, "aaaabbbbccccddddeeeeffff0000111122223333"},
		{"Manakosten", eine.Manakosten, "{R}"},
		{"Typzeile", eine.Typzeile, "Instant"},
		{"Seltenheit", eine.Seltenheit, "Common"},
		{"Setcode", eine.Setcode, "PRB"},
		{"Setname", eine.Setname, "Probeset"},
		{"Nummer", eine.Nummer, "141"},
		{"Regeltext", eine.Regeltext, "Blitzschlag fügt einem Ziel 3 Schadenspunkte zu."},
		{"Flavortext", eine.Flavortext, "Kurz und heiß."},
		{"Kuenstler", eine.Kuenstler, "Selbst gezeichnet"},
		{"Bildadresse", eine.Bildadresse, "https://example.invalid/blitz.png"},
	} {
		if fall.gelesen != fall.erwartet {
			t.Errorf("%s ist %q, erwartet %q", fall.feld, fall.gelesen, fall.erwartet)
		}
	}
}

// **Die Legalitäten kommen als Abbildung an** — dieselbe Gestalt wie bei
// Scryfall, damit die Formatprüfung beide Quellen gleich liest.
//
// **Und in Kleinschreibung**: MTG API schreibt „Legal", Scryfall „legal".
// Zwei Schreibweisen für dieselbe Aussage wären zwei Wahrheiten, und die
// Prüfung müsste beide kennen.
func TestDieLegalitaetenKommenKleingeschriebenAn(t *testing.T) {
	t.Parallel()

	karten, _, err := mtgapi.LiesKarten(beispiel(t, "karte.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for format, erwartet := range map[string]string{
		"commander": "legal",
		"vintage":   "restricted",
	} {
		if gelesen := karten[0].Legalitaeten[format]; gelesen != erwartet {
			t.Errorf("%s ist %q, erwartet %q", format, gelesen, erwartet)
		}
	}
}

// **Kraft und Widerstandskraft kommen an, wo es sie gibt.**
func TestKraftUndWiderstandKommenAn(t *testing.T) {
	t.Parallel()

	karten, _, err := mtgapi.LiesKarten(beispiel(t, "kreatur.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	eine := karten[0]

	if eine.Kraft != "3" || eine.Widerstand != "4" {
		t.Errorf("Kraft/Widerstand sind %q/%q, erwartet 3/4", eine.Kraft, eine.Widerstand)
	}

	// **Und bei einer Karte ohne beides bleiben sie leer** — nicht „0".
	blitz, _, blitzfehler := mtgapi.LiesKarten(beispiel(t, "karte.json"))
	if blitzfehler != nil {
		t.Fatalf("lesen: %v", blitzfehler)
	}

	if blitz[0].Kraft != "" || blitz[0].Widerstand != "" {
		t.Errorf("ein Instant hat Kraft %q und Widerstand %q",
			blitz[0].Kraft, blitz[0].Widerstand)
	}
}

// **Ein fehlendes Pflichtfeld bricht laut** (§16: „Pflichtfelder streng").
//
// **Und die Meldung nennt das Feld** — wer sie liest, soll nicht raten.
func TestEinFehlendesPflichtfeldBrichtLaut(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		roh   string
		nennt string
	}{
		{"ohne Name", `{"cards":[{"id":"a","set":"P"}]}`, "name"},
		{"ohne Kennung", `{"cards":[{"name":"X","set":"P"}]}`, "id"},
		{"ohne Set", `{"cards":[{"name":"X","id":"a"}]}`, "set"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := mtgapi.LiesKarten([]byte(fall.roh))

			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}

// **Ein unbekanntes Feld stört nicht** (§16: „Zusatzfelder tolerant").
//
// Ein neues Feld beim Anbieter darf nichts brechen — sonst stünde die
// Anwendung still, sobald jemand dort etwas hinzufügt.
func TestEinUnbekanntesFeldStoertNicht(t *testing.T) {
	t.Parallel()

	roh := `{"cards":[{"name":"X","id":"a","set":"P","brandneu":{"tief":1}}]}`

	karten, befunde, err := mtgapi.LiesKarten([]byte(roh))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	if len(karten) != 1 || karten[0].Name != "X" {
		t.Errorf("gelesen: %+v", karten)
	}
}

// **Ein unbekannter Aufzählungswert wird ein Befund und nicht verschluckt**
// (§16, „Drift wird erkannt, nicht erlitten").
func TestEinUnbekannterLegalitaetswertWirdEinBefund(t *testing.T) {
	t.Parallel()

	roh := `{"cards":[{"name":"X","id":"a","set":"P",
	  "legalities":[{"format":"Commander","legality":"Vielleicht"}]}]}`

	karten, befunde, err := mtgapi.LiesKarten([]byte(roh))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "Vielleicht") {
		t.Errorf("der Befund %q nennt den Wert nicht", befunde[0].Meldung)
	}

	// **Und die Karte kommt trotzdem** — mit dem Format, das sie hat, und
	// ohne das, das niemand deuten kann (§8).
	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}

	if _, gibts := karten[0].Legalitaeten["commander"]; gibts {
		t.Error("der unlesbare Wert steht in den Legalitäten")
	}
}

// **Unbrauchbares JSON ist ein Fehler und keine leere Liste.**
func TestUnbrauchbaresJSONIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, _, err := mtgapi.LiesKarten([]byte("kein json")); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}
