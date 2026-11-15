// Kleine Helfer der Abnahme — Zahlen aus JSON, Dateien prüfen.
package flaechen_test

import (
	"encoding/json"
	"errors"
	"image/png"
	"os"
	"strconv"
	"strings"
	"testing"
)

// errorsAs ist errors.As — hier, damit `ablauf_test.go` den Import nicht braucht.
func errorsAs(fehler error, ziel any) bool { return errors.As(fehler, ziel) }

// mussJSON macht aus einem Wert JSON oder bricht ab.
func mussJSON(t *testing.T, wert any) []byte {
	t.Helper()

	roh, err := json.Marshal(wert)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}

	return roh
}

// setkennungAus liest `setID` aus einer JSON-Antwort und gibt sie als Text.
//
// **Als Text**, weil die Kommandozeile Kennungen als Zeichenketten führt und
// MCP als Zahlen — die Abnahme fährt beide Wege mit denselben Helfern, und der
// Unterschied gehört an **eine** Stelle.
func setkennungAus(t *testing.T, roh string) string {
	const feld = "setID"

	t.Helper()

	var antwort map[string]any
	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	switch wert := antwort[feld].(type) {
	case string:
		return wert
	case float64:
		return strconv.FormatInt(int64(wert), 10)
	default:
		t.Fatalf("%q steht nicht als Zahl in %s", feld, roh)

		return ""
	}
}

// zahl macht aus einer Kennung eine Zahl für MCP.
func zahl(t *testing.T, kennung string) int64 {
	t.Helper()

	wert, err := strconv.ParseInt(kennung, 10, 64)
	if err != nil {
		t.Fatalf("die Kennung %q ist keine Zahl", kennung)
	}

	return wert
}

// ersteKarte liest die erste Kartenkennung aus der Antwort von `mse karten`.
func ersteKarte(t *testing.T, roh string) string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID string `json:"id"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(antwort.Karten) == 0 {
		t.Fatalf("keine Karte in %s", roh)
	}

	return antwort.Karten[0].ID
}

// ersteKarteMCP liest sie aus der Antwort von `lesen.karten`.
//
// **Eine eigene Gestalt, und das ist kein Versehen**: die beiden Flächen geben
// verschiedene Formen — das dürfen sie. Gleich sein muss die **Fachlichkeit**,
// nicht die Ausgabe.
func ersteKarteMCP(t *testing.T, roh string) string {
	t.Helper()

	var karten []struct {
		ID int64 `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &karten); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(karten) == 0 {
		t.Fatalf("keine Karte in %s", roh)
	}

	return strconv.FormatInt(karten[0].ID, 10)
}

// istBild prüft, dass eine Datei ein **entzifferbares** Bild mit Fläche ist.
//
// # Warum nicht die Kennung
//
// Hier stand ein Vergleich der ersten acht Bytes gegen die PNG-Signatur, und
// der Zweitblick über Phase T hat angestrichen, was er durchließe: eine Datei
// aus **genau diesen acht Bytes** wäre grün. Das ist die Klasse „geprüft, dass
// etwas geschah, nicht dass das Richtige geschah" — und der Abnahmelauf ist
// der letzte Ort, an dem sie stehen darf.
//
// Entziffert wird deshalb wirklich, und die Fläche muss größer als null sein.
func istBild(t *testing.T, pfad string) {
	t.Helper()

	datei, err := os.Open(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Bilddatei fehlt: %v", err)
	}

	defer schliesseStill(t, datei)

	bild, fehler := png.Decode(datei)
	if fehler != nil {
		t.Fatalf("%s ist kein entzifferbares PNG: %v", pfad, fehler)
	}

	if flaeche := bild.Bounds(); flaeche.Dx() <= 0 || flaeche.Dy() <= 0 {
		t.Errorf("%s hat keine Fläche: %v", pfad, flaeche)
	}
}

// schliesseStill schließt eine Datei, die nur gelesen wurde.
func schliesseStill(t *testing.T, datei *os.File) {
	t.Helper()

	if err := datei.Close(); err != nil {
		t.Errorf("%s schließen: %v", datei.Name(), err)
	}
}

// istWiederLesbaresSet prüft, dass ein geschriebenes Set **wieder aufgeht**.
//
// # Warum nicht os.Stat
//
// Hier stand genau das, und der Zweitblick über Phase T hat gefragt, was es
// durchließe: eine **leere Datei**. Damit hätte der Abnahmelauf „ein
// festgelegter Arbeitsablauf läuft" zugesagt, während sein letzter Schritt
// nichts Brauchbares hinterlässt.
//
// Der Beleg ist deshalb, dass die Datei durch **denselben Weg** wieder
// hereinkommt, über den ein Mensch sie hereinbekäme — und dass die Karte
// darin steht, die vorher geändert wurde.
//
// # Und warum in einem **frischen** Bestand
//
// Im selben ging es nicht: `sichern` schreibt den neuen Pfad an das Set, und
// ein `oeffnen` darauf trifft die Regel „eine Quelle, ein Set" — es gibt die
// schon offene Kennung zurück und **liest die Datei gar nicht**. Der erste
// Anlauf prüfte damit den Bestand gegen sich selbst und wäre auch dann grün
// gewesen, wenn die geschriebene Datei Unsinn enthielte.
//
// Gemessen am 2026-08-23: derselbe Bestand gibt `setID 1` ohne Dateizugriff,
// ein frischer liest 3 Karten aus der Datei.
//
// **Was er nicht prüft**: dass die Datei Byte für Byte der Vorlage entspricht.
// Das ist die Zusage von #48 und wird dort gehalten; hier geht es darum, dass
// das Ergebnis dieses Ablaufs benutzbar ist.
func istWiederLesbaresSet(t *testing.T, _ []string, pfad, erwarteterName string) {
	t.Helper()

	umgebung := eigenerBestand(t)

	aus, fehlerstrom, code := kommandozeile(t, umgebung, "oeffnen", pfad, "--json")
	if code != 0 {
		t.Fatalf("das geschriebene Set geht nicht wieder auf: %d (%s)", code, fehlerstrom)
	}

	karten, _, code := kommandozeile(t, umgebung,
		"karten", setkennungAus(t, aus), "--json")
	if code != 0 {
		t.Fatalf("das wieder eingefahrene Set gibt keine Karten: %d", code)
	}

	if !strings.Contains(karten, erwarteterName) {
		t.Errorf("die Änderung steht nicht im geschriebenen Set — %q fehlt in %s",
			erwarteterName, karten)
	}
}

// setzeile ist ein Set, wie `mse sets --json` es herausgibt.
//
// **Die Kennung ist eine Zeichenkette.** So schreibt die Kommandozeile sie —
// nachgemessen, nicht angenommen: der erste Anlauf las sie als Zahl und bekam
// eine leere Liste.
type setzeile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// setzeilenAus liest die Setliste aus einer Antwort von `mse sets --json`.
func setzeilenAus(t *testing.T, roh string) []setzeile {
	t.Helper()

	var antwort struct {
		Sets []setzeile `json:"sets"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return antwort.Sets
}

// kartenzeilenAus liest die Kartenliste aus `mse karten --json`.
func kartenzeilenAus(t *testing.T, roh string) []struct {
	ID string `json:"id"`
} {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID string `json:"id"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return antwort.Karten
}

// feldwertAus liest den Wert eines Feldes der **ersten** Karte.
func feldwertAus(t *testing.T, roh, feld string) string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(antwort.Karten) == 0 {
		t.Fatalf("keine Karte in %s", roh)
	}

	return antwort.Karten[0].Werte[feld]
}

// zahlen wandelt mehrere Kennungen in Zahlen — für MCP-Argumente (#1390).
func zahlen(t *testing.T, kennungen []string) []int64 {
	t.Helper()

	werte := make([]int64, 0, len(kennungen))
	for _, kennung := range kennungen {
		werte = append(werte, zahl(t, kennung))
	}

	return werte
}

// alsJSONZahlen schreibt Kennungen als JSON-Zahlenliste — für REST-Rümpfe.
//
// **Als Zahlen und nicht als Zeichenketten.** Die Kennungen sind an der
// Kommandozeile Text, weil dort alles Text ist; im JSON-Rumpf sind sie
// Zahlen, und ein Test, der sie als Text schickte, prüfte den Fehlerweg statt
// des Wegs.
func alsJSONZahlen(t *testing.T, kennungen []string) string {
	t.Helper()

	roh, err := json.Marshal(zahlen(t, kennungen))
	if err != nil {
		t.Fatalf("Kennungen als JSON: %v", err)
	}

	return string(roh)
}
