package main

// **MTG JSON an der Kommandozeile** —
// #1291.
//
// # Der Befund, der sie verlangt
//
// `docs/FAEHIGKEITEN.md` führte zwei Zeilen so:
//
//	| MTG JSON lesen    | `MTGJSONOeffnen`, … | — | — | — | ✋ |
//	| MTG JSON schreiben | `MTGJSONSchreiben`  | — | — | — | ✋ |
//
// Drei Gedankenstriche und eine erhobene Hand. Die Hand steht für den
// Dateidialog — und der steckt allein in der **Bindung**: `App.MTGJSONOeffnen`
// und `App.MTGJSONSchreiben` nehmen einen Pfad entgegen.
//
// **Und gerade dieses Format ist das, das ein Skript braucht**: MTG JSON ist
// der Weg zwischen diesem Programm und allem anderen (0009).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mtgjsonprobe gibt den Pfad zu einer eingecheckten MTG-JSON-Datei.
func mtgjsonprobe(t *testing.T, name string) string {
	t.Helper()

	pfad, err := filepath.Abs(filepath.Join("..", "..", "testdaten", "mtgjson", name))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if _, err := os.Stat(pfad); err != nil {
		t.Fatalf("die Probedatei fehlt: %v", err)
	}

	return pfad
}

// **`mtgjson-lesen` fährt eine Datei ein und nennt, was entstanden ist.**
func TestMTGJSONLesenFaehrtEin(t *testing.T) {
	ort := frischerBestand(t)

	code, aus, fehler := fahreAn(t, ort,
		"mtgjson-lesen", mtgjsonprobe(t, "probe.json"), "magic", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var eingang struct {
		Angelegt []string `json:"angelegt"`
		Befunde  []string `json:"befunde"`
	}

	if err := json.Unmarshal([]byte(aus), &eingang); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if len(eingang.Angelegt) == 0 {
		t.Fatalf("es wurde kein Set angelegt: %s", aus)
	}

	// **Und die Sets stehen danach wirklich im Bestand** — nicht nur in der
	// Antwort. Zwei grüne Seiten einer Grenze sind kein Nachweis für die
	// Grenze.
	_, liste, _ := fahreAn(t, ort, "sets")
	if strings.TrimSpace(liste) == "" {
		t.Error("der Bestand ist leer geblieben")
	}
}

// **`mtgjson-schreiben` schreibt ein Set als MTG JSON — und es ist welches.**
func TestMTGJSONSchreibenSchreibtEineDatei(t *testing.T) {
	ort := frischerBestand(t)

	_, aus, _ := fahreAn(t, ort,
		"mtgjson-lesen", mtgjsonprobe(t, "probe.json"), "magic", "--json")

	var eingang struct {
		Angelegt []string `json:"angelegt"`
	}

	if err := json.Unmarshal([]byte(aus), &eingang); err != nil || len(eingang.Angelegt) == 0 {
		t.Fatalf("einfahren: %v (%s)", err, aus)
	}

	ziel := filepath.Join(t.TempDir(), "raus.json")

	code, ausgabe, fehler := fahreAn(t, ort,
		"mtgjson-schreiben", eingang.Angelegt[0], "magic", ziel)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	// **Und es ist MTG JSON und nicht irgendein JSON**: der Umschlag mit
	// `meta` und `data` ist das, woran ein anderes Werkzeug es erkennt.
	var geschrieben struct {
		Meta map[string]any `json:"meta"`
		Data map[string]any `json:"data"`
	}

	if err := json.Unmarshal(roh, &geschrieben); err != nil {
		t.Fatalf("das Geschriebene ist kein JSON: %v", err)
	}

	if len(geschrieben.Meta) == 0 || len(geschrieben.Data) == 0 {
		t.Errorf("der Umschlag fehlt: %s", roh[:min(120, len(roh))])
	}

	if !strings.Contains(ausgabe, ziel) {
		t.Errorf("die Ausgabe nennt den Pfad nicht: %q", ausgabe)
	}
}

// **`mtgjson-spiele` nennt die Spiele, für die es eine Zuordnung gibt.**
func TestMTGJSONSpieleNenntMagic(t *testing.T) {
	ort := frischerBestand(t)

	code, aus, fehler := fahreAn(t, ort, "mtgjson-spiele")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "magic") {
		t.Errorf("magic fehlt: %q", aus)
	}
}

// **Ein Spiel ohne Zuordnung ist ein Bedienfehler.**
func TestMTGJSONMitEinemUnbekanntenSpiel(t *testing.T) {
	ort := frischerBestand(t)

	if code, _, _ := fahreAn(t, ort,
		"mtgjson-lesen", mtgjsonprobe(t, "probe.json"), "kein-spiel"); code != CodeBedienung {
		t.Errorf("lesen: Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Und die Bedienfehler beim Aufruf selbst.**
func TestMTGJSONBedienfehler(t *testing.T) {
	ort := frischerBestand(t)

	faelle := map[string][]string{
		"lesen ohne Spiel":              {"mtgjson-lesen", "/tmp/x.json"},
		"schreiben ohne Ziel":           {"mtgjson-schreiben", "1", "magic"},
		"schreiben mit krummer Kennung": {"mtgjson-schreiben", "keine Zahl", "magic", "/tmp/x.json"},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}
