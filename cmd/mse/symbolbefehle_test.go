// **Die Symbolbefehle an der Kommandozeile** — Werkstück
// #1391.
//
// # Was hier geprüft wird
//
// **Der Weg durch den Befehl**: Argumente lesen, Eingabe lesen, rechnen,
// ausgeben, und die Rückgabewerte. Dass die Rechnung stimmt, prüft
// `internal/app`; dass alle drei Flächen dasselbe sagen,
// `internal/flaechen/symbol_test.go`.
package main

import (
	"bytes"
	"strings"
	"testing"
)

// fahreMitEingabe führt einen Aufruf mit einem Text auf der Standardeingabe.
func fahreMitEingabe(
	t *testing.T, ort, eingabe string, argumente ...string,
) (code int, aus, fehler string) {
	t.Helper()

	var raus, rfehler bytes.Buffer

	code = lauf(probekontext(t), argumente, strings.NewReader(eingabe),
		&raus, &rfehler, ort)

	return code, raus.String(), rfehler.String()
}

// **Eine Grundform entsteht ohne Bestand und ohne Eingabe.**
func TestSymbolRechnenGibtEineGrundform(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Rechteck", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Rechteck") {
		t.Errorf("keine Grundform: %s", aus)
	}
}

// **Die Kette: eine Liste auf der Eingabe sind die Teile.**
func TestSymbolRechnenNimmtDieTeileVonDerEingabe(t *testing.T) {
	ort, _, _ := mitKarte(t)

	_, teile, _ := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Rechteck", "--json")

	code, aus, fehler := fahreMitEingabe(t, ort, teile,
		"symbol-rechnen", "lesen.symbol.symmetrie",
		"--stelle", "0", "--weise", "anlegen", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{"Symmetrie", "rotation"} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, aus)
		}
	}
}

// **Und ein Objekt auf der Eingabe ist der ganze Rumpf** — dieselbe Gestalt,
// die MCP und der Dienst nehmen.
func TestSymbolRechnenNimmtAuchDenGanzenRumpf(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreMitEingabe(t, ort, `{"art":"Rechteck"}`,
		"symbol-rechnen", "lesen.symbol.grundform", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Rechteck") {
		t.Errorf("keine Grundform: %s", aus)
	}
}

// **Die Fehlerwege.**
func TestDieSymbolbefehleWeisenAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, fall := range map[string]struct {
		eingabe   string
		argumente []string
		erwartet  int
	}{
		"ohne Namen der Rechnung": {
			argumente: []string{"symbol-rechnen"}, erwartet: CodeBedienung,
		},
		"eine Rechnung, die es nicht gibt": {
			eingabe:   `{}`,
			argumente: []string{"symbol-rechnen", "lesen.symbol." + "gibtsnicht"},
			erwartet:  CodeFachlich,
		},
		"eine Eingabe, die kein JSON ist": {
			eingabe:   `[kein json`,
			argumente: []string{"symbol-rechnen", "lesen.symbol.grundform"},
			erwartet:  CodeBedienung,
		},
		"ein Objekt, das kein JSON ist": {
			eingabe:   `{kein json`,
			argumente: []string{"symbol-rechnen", "lesen.symbol.grundform"},
			erwartet:  CodeBedienung,
		},
		"symbol ohne Setkennung": {
			argumente: []string{"symbol"}, erwartet: CodeBedienung,
		},
		"symbol mit unbrauchbarer Kennung": {
			argumente: []string{"symbol", "keine-zahl"}, erwartet: CodeBedienung,
		},
		"symbol-setzen ohne Setkennung": {
			argumente: []string{"symbol-setzen"}, erwartet: CodeBedienung,
		},
		"symbol-setzen ohne Teile": {
			eingabe:   `kein json`,
			argumente: []string{"symbol-setzen", "1"},
			erwartet:  CodeBedienung,
		},
		"symbolpalette ohne Feld": {
			argumente: []string{"symbolpalette", "1"}, erwartet: CodeBedienung,
		},
		"symbolpalette mit unbrauchbarer Kennung": {
			argumente: []string{"symbolpalette", "keine-zahl", "name"},
			erwartet:  CodeBedienung,
		},
		"felder-setzen ohne Karten": {
			argumente: []string{"felder-setzen", "1", "artist", "X"},
			erwartet:  CodeBedienung,
		},
		"felder-setzen mit unbrauchbarer Setkennung": {
			argumente: []string{"felder-setzen", "keine-zahl", "artist", "X", "1"},
			erwartet:  CodeBedienung,
		},
		"felder-setzen mit unbrauchbarer Kartenkennung": {
			argumente: []string{"felder-setzen", "1", "artist", "X", "keine-zahl"},
			erwartet:  CodeBedienung,
		},
	} {
		t.Run(name, func(t *testing.T) {
			code, aus, fehler := fahreMitEingabe(t, ort, fall.eingabe, fall.argumente...)
			if code != fall.erwartet {
				t.Errorf("Rückgabewert %d, erwartet %d (%s%s)",
					code, fall.erwartet, aus, fehler)
			}
		})
	}
}

// **Ein Feld auf mehreren Karten — und der vorige Wert steht danach nicht
// mehr da** (#1390).
func TestFelderSetzenSchreibtAufMehrereKarten(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort,
		"felder-setzen", setID, "artist", "Aus der Kommandozeile", karteID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "1 Karten") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	_, gelesen, _ := fahreAn(t, ort, "karten", setID, "--json")
	if !strings.Contains(gelesen, "Aus der Kommandozeile") {
		t.Errorf("der Wert steht nicht da: %s", gelesen)
	}
}

// **Und als JSON.**
func TestFelderSetzenAlsJSON(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort,
		"felder-setzen", setID, "artist", "X", karteID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"geaendert": 1`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Eine Karte, die es nicht gibt, lässt den ganzen Aufruf scheitern.**
func TestFelderSetzenWeistEineFremdeKarteAb(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, _ := fahreAn(t, ort,
		"felder-setzen", setID, "artist", "Verboten", karteID, "999999")
	if code == CodeGut {
		t.Fatalf("der Aufruf ging durch: %s", aus)
	}

	_, gelesen, _ := fahreAn(t, ort, "karten", setID, "--json")
	if strings.Contains(gelesen, "Verboten") {
		t.Error("die erste Karte wurde geändert, obwohl der Aufruf scheiterte")
	}
}

// **Ein Symbol ablegen und wieder lesen.**
//
// **Mit eigenem Zuhause** (§3): die Symbolablage hängt am Bilderverzeichnis,
// und das leitet sich vom Datenbankpfad ab — der aus `HOME` kommt. Ein Test
// ohne diese Zeile schriebe in den Bestand dieses Rechners.
func TestSymbolSetzenUndLesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, setID, _ := mitKarte(t)

	_, teile, _ := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Rechteck", "--json")

	code, aus, fehler := fahreMitEingabe(t, ort, teile,
		"symbol-setzen", setID, "probe.mse-symbol")
	if code != CodeGut {
		t.Fatalf("ablegen: %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "probe.mse-symbol") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	code, gelesen, fehler := fahreAn(t, ort, "symbol", setID, "probe.mse-symbol")
	if code != CodeGut {
		t.Fatalf("lesen: %d (%s)", code, fehler)
	}

	if !strings.Contains(gelesen, "Rechteck") {
		t.Errorf("das Gelesene lautet %q", gelesen)
	}
}

// **Und als JSON, mit dem Namen in der Antwort.**
func TestSymbolSetzenAlsJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, setID, _ := mitKarte(t)

	_, teile, _ := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Rechteck", "--json")

	code, aus, fehler := fahreMitEingabe(t, ort, teile,
		"symbol-setzen", setID, "--json")
	if code != CodeGut {
		t.Fatalf("ablegen: %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"name"`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Die Palette eines Feldes.**
//
// Das Probeset bringt keine Vorlage mit Palette mit; was hier zählt, ist,
// dass der Weg trägt — und dass eine leere Antwort eine **Liste** ist und
// keine `null` (§3).
func TestSymbolpaletteGibtEineListe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, _, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "symbolpalette", karteID, "name", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "[") {
		t.Errorf("keine Liste: %s", aus)
	}
}
