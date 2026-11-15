// **Die Symbolbibliothek an der Kommandozeile** — DOM-02
// (#1697).
//
// **Der Weg durch den Befehl**: Argumente, Ausgabe, Rückgabewerte. Dass alle
// Flächen dasselbe sagen, prüft `internal/flaechen`.
package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitSymbolImSet legt ein gezeichnetes Symbol in das Set von [mitKarte].
func mitSymbolImSet(t *testing.T) (ort, setID string) {
	t.Helper()

	ort, setID, _ = mitKarte(t)

	_, teile, _ := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Rechteck", "--json")
	if code, _, fehler := fahreMitEingabe(t, ort, teile, "symbol-setzen", setID, "probe.mse-symbol"); code != CodeGut {
		t.Fatalf("ablegen: %d (%s)", code, fehler)
	}

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // ein Lauf schließt selbst

	kennung, _ := strconv.ParseInt(setID, 10, 64) //nolint:errcheck // aus dem Aufbau
	if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
		probekontext(t), kennung, "symbol", "probe.mse-symbol"); err != nil {
		t.Fatalf("das Symbol setzen: %v", err)
	}

	return ort, setID
}

func TestDieBibliothekGehtDurchDieKommandozeile(t *testing.T) {
	ort, setID := mitSymbolImSet(t)

	if leer := mussLaufen(t, ort, "symbolbibliothek"); !strings.Contains(leer, "leer") {
		t.Errorf("die leere Bibliothek: %q", leer)
	}

	krone := strings.TrimSpace(mussLaufen(t, ort, "symbolbibliothek", "aufnehmen", setID, "Krone"))

	if liste := mussLaufen(t, ort, "symbolbibliothek", "liste", "kro"); !strings.Contains(liste, "Krone") ||
		!strings.Contains(liste, "Probeset") {
		t.Errorf("die Liste: %q", liste)
	}

	if liste := mussLaufen(t, ort, "symbolbibliothek", "liste", "zzz"); !strings.Contains(liste, "Kein Symbol") {
		t.Errorf("ohne Treffer: %q", liste)
	}

	if text := mussLaufen(t, ort, "symbolbibliothek", "ausgeben", krone); !strings.Contains(text, "mse_version") {
		t.Errorf("als Text: %q", text)
	}

	datei := filepath.Join(t.TempDir(), "reif.mse-symbol")
	if aus := mussLaufen(t, ort, "symbolbibliothek", "ausgeben", krone, datei); !strings.Contains(aus, datei) {
		t.Errorf("in die Datei: %q", aus)
	}

	if _, err := os.Stat(datei); err != nil {
		t.Fatalf("die Datei: %v", err)
	}

	reif := strings.TrimSpace(mussLaufen(t, ort, "symbolbibliothek", "einlesen", datei))
	zacke := strings.TrimSpace(mussLaufen(t, ort, "symbolbibliothek", "duplizieren", krone, "Zacke"))
	mussLaufen(t, ort, "symbolbibliothek", "umbenennen", zacke, "Zinne")

	if eingesetzt := mussLaufen(t, ort, "symbolbibliothek", "einsetzen", reif, setID); !strings.Contains(eingesetzt, ".mse-symbol") {
		t.Errorf("einsetzen: %q", eingesetzt)
	}

	// **Verwendet ohne Bestätigung geht nicht** — und die Absage nennt das Set.
	if code, _, fehler := fahreAn(t, ort, "symbolbibliothek", "loeschen", krone); code != CodeFachlich ||
		!strings.Contains(fehler, "Probeset") {
		t.Errorf("ohne Bestätigung: %d %q", code, fehler)
	}

	if folgen := mussLaufen(t, ort, "symbolbibliothek", "loeschen", krone, bestaetigtfahne); !strings.Contains(folgen, "Probeset") {
		t.Errorf("mit Bestätigung: %q", folgen)
	}

	if liste := mussLaufen(t, ort, "symbolbibliothek", "--json"); !strings.Contains(liste, `"name": "Zinne"`) ||
		strings.Contains(liste, `"name": "Krone"`) {
		t.Errorf("danach: %q", liste)
	}
}

// **Was fehlt, ist ein Bedienfehler** — mit der Hilfe, statt eines Rätsels.
func TestDieBibliotheksbefehleSagenWasFehlt(t *testing.T) {
	ort, _ := mitSymbolImSet(t)

	for name, argumente := range map[string][]string{
		"unbekannt":           {"symbolbibliothek", "zaubern"},
		"aufnehmen ohne Name": {"symbolbibliothek", "aufnehmen", "1"},
		"duplizieren":         {"symbolbibliothek", "duplizieren", "1"},
		"umbenennen":          {"symbolbibliothek", "umbenennen", "1"},
		"loeschen":            {"symbolbibliothek", "loeschen"},
		"einsetzen":           {"symbolbibliothek", "einsetzen", "1"},
		"einlesen":            {"symbolbibliothek", "einlesen"},
		"ausgeben":            {"symbolbibliothek", "ausgeben"},
	} {
		if code, _, fehler := fahreAn(t, ort, argumente...); code != CodeBedienung || fehler == "" {
			t.Errorf("%s: %d %q", name, code, fehler)
		}
	}

	// **Und ein fachliches Nein ist keines der Bedienung.**
	for name, argumente := range map[string][]string{
		"aufnehmen":   {"symbolbibliothek", "aufnehmen", "x", "Krone"},
		"duplizieren": {"symbolbibliothek", "duplizieren", "99", "X"},
		"umbenennen":  {"symbolbibliothek", "umbenennen", "99", "X"},
		"einsetzen":   {"symbolbibliothek", "einsetzen", "99", "1"},
		"einlesen":    {"symbolbibliothek", "einlesen", filepath.Join(t.TempDir(), "weg.mse-symbol")},
		"ausgeben":    {"symbolbibliothek", "ausgeben", "99"},
		"in Datei":    {"symbolbibliothek", "ausgeben", "99", filepath.Join(t.TempDir(), "x.mse-symbol")},
	} {
		if code, _, _ := fahreAn(t, ort, argumente...); code != CodeFachlich {
			t.Errorf("%s: %d", name, code)
		}
	}
}

// **Bearbeiten an der Kommandozeile** — die Teile kommen über die
// Standardeingabe, wie bei `symbol-setzen`.
func TestEinSymbolWirdAnDerKommandozeileBearbeitet(t *testing.T) {
	ort, setID := mitSymbolImSet(t)
	krone := strings.TrimSpace(mussLaufen(t, ort, "symbolbibliothek", "aufnehmen", setID, "Krone"))

	teile := mussLaufen(t, ort, "symbolbibliothek", "teile", krone)
	if !strings.Contains(teile, "Rechteck") {
		t.Fatalf("die Teile: %q", teile)
	}

	_, kreis, _ := fahreMitEingabe(t, ort, "",
		"symbol-rechnen", "lesen.symbol.grundform", "--art", "Kreis", "--json")

	if code, aus, fehler := fahreMitEingabe(t, ort, kreis, "symbolbibliothek", "sichern", krone); code != CodeGut ||
		!strings.Contains(aus, "Probeset") {
		t.Fatalf("sichern: %d %q %q", code, aus, fehler)
	}

	if aus := mussLaufen(t, ort, "symbolbibliothek", "nachziehen", krone); !strings.Contains(aus, "Probeset") {
		t.Errorf("nachziehen: %q", aus)
	}

	if aus := mussLaufen(t, ort, "symbolbibliothek", "zuruecknehmen", krone); !strings.Contains(aus, "Probeset") {
		t.Errorf("zurücknehmen: %q", aus)
	}

	for name, argumente := range map[string][]string{
		"teile":         {"symbolbibliothek", "teile"},
		"sichern":       {"symbolbibliothek", "sichern"},
		"zuruecknehmen": {"symbolbibliothek", "zuruecknehmen"},
		"nachziehen":    {"symbolbibliothek", "nachziehen"},
	} {
		if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
			t.Errorf("%s ohne Kennung: %d", name, code)
		}
	}

	if code, _, _ := fahreMitEingabe(t, ort, "kein json", "symbolbibliothek", "sichern", krone); code != CodeBedienung {
		t.Errorf("unlesbare Teile: %d", code)
	}

	for name, argumente := range map[string][]string{
		"teile":         {"symbolbibliothek", "teile", "99"},
		"zuruecknehmen": {"symbolbibliothek", "zuruecknehmen", "99"},
		"nachziehen":    {"symbolbibliothek", "nachziehen", "99"},
	} {
		if code, _, _ := fahreAn(t, ort, argumente...); code != CodeFachlich {
			t.Errorf("%s ohne Symbol: %d", name, code)
		}
	}

	if code, _, _ := fahreMitEingabe(t, ort, "[]", "symbolbibliothek", "sichern", "99"); code != CodeFachlich {
		t.Errorf("sichern ohne Symbol: %d", code)
	}
}
