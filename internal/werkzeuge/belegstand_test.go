package werkzeuge_test

// **Die Zahlen im README stammen aus einem Beleg** — QA-01.AK-04
// (#1733):
// „README, Abnahme und Berichte beziehen Kennzahlen auf denselben aktuellen
// Beleg."
//
// # Der Anlass
//
// **Der Einstieg des README sagte »null Befunde der Art Portierung«**, und die
// Abnahme zu 0.5.0 nannte 679 — zwei Läufe, zwei Umfänge, und keiner der
// beiden Texte sagte, welcher gemeint war. Die Tabelle »Der volle Kanon« kam
// aus einem Lauf von Hand, den `make fremdpakete` gar nicht fuhr.
//
// # Was er hält
//
//  1. **Jeder Beleg steht im Belegstand des README** — Lauf, Stand, Commit,
//     Tag der Messung, gemessene und erwartete Fälle —, und was dort steht,
//     ist der Beleg.
//  2. **Die Befundzahlen des Kanonlaufs und die Tabellen »Der volle Kanon«**
//     gleichen den Kennzahlen des Belegs, auf die Stellen gerundet, die das
//     README zeigt.
//
// # Was er nicht hält
//
// **Ob der Beleg aktuell ist.** Das sagt `make belege`: ein Wächter, der mit
// der Uhr rot würde, wäre rot, ohne dass sich im Baum etwas geändert hätte.
// **Und keinen Absatz, der sein Datum nennt.** Den Freigabeblock hält seit
// dem Abschluss von 0.6.0 `freigabezahlen_test.go` — bis dahin hielt ihn
// niemand, und er nannte noch die Zahlen von 0.5.0.

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// kanonbeleg liest den Beleg von `make fremdpakete` aus docs/belege — der
// einzige, gegen den README und Freigabeblock Zahlen halten.
func kanonbeleg(t *testing.T) laufbeleg.Beleg {
	t.Helper()

	const lauf = "fremdpakete"

	datei, fehler := os.Open(filepath.Join("..", "..", "docs", "belege", lauf+".json")) //nolint:gosec // fester Ort im Baum
	if fehler != nil {
		t.Fatalf("kein Beleg für %s: %v — `make %s` schreibt ihn", lauf, fehler, lauf)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	beleg, fehler := laufbeleg.Lesen(datei)
	if fehler != nil {
		t.Fatalf("der Beleg für %s: %v", lauf, fehler)
	}

	return beleg
}

// kennzahl liest eine Kennzahl als Zahl — und fällt, wenn der Beleg sie
// nicht trägt: dann lief der Lauf nicht, der sie liefert.
func kennzahl(t *testing.T, beleg laufbeleg.Beleg, name string) float64 {
	t.Helper()

	roh, gibt := beleg.Kennzahlen[name]
	if !gibt {
		t.Fatalf("der Beleg für %s trägt keine Kennzahl %s — lief der Lauf, der sie liefert?",
			beleg.Lauf, name)
	}

	wert, fehler := strconv.ParseFloat(roh, 64)
	if fehler != nil {
		t.Fatalf("die Kennzahl %s ist keine Zahl: %q", name, roh)
	}

	return wert
}

// deutsch schreibt eine Zahl, wie das README sie zeigt: Komma vor den
// Nachkommastellen, Punkt zwischen den Tausendern.
func deutsch(wert float64, stellen int) string {
	text := strconv.FormatFloat(wert, 'f', stellen, 64)
	ganz, rest, hatRest := strings.Cut(text, ".")

	var gruppiert strings.Builder

	for i, ziffer := range ganz {
		if i > 0 && (len(ganz)-i)%3 == 0 {
			gruppiert.WriteByte('.')
		}

		gruppiert.WriteRune(ziffer)
	}

	if hatRest {
		return gruppiert.String() + "," + rest
	}

	return gruppiert.String()
}

// **Der Belegstand im README ist der Beleg.**
func TestDerBelegstandImReadmeIstDerBeleg(t *testing.T) {
	t.Parallel()

	liesmich := lies(t, "README.md")
	beleg := kanonbeleg(t)

	// **Der Tag steht im Zeitversatz des Belegs** und nicht in dem des
	// Rechners, der prüft — sonst wäre ein Lauf um halb eins auf zwei
	// Rechnern an zwei Tagen gemessen.
	zeile := fmt.Sprintf("| %#q | **%s** | %#q | %s | %d von %d |",
		beleg.Lauf, beleg.Status, beleg.Commit[:min(7, len(beleg.Commit))],
		beleg.Ende.Format(time.DateOnly), beleg.Gemessen, beleg.Erwartet)

	if !strings.Contains(liesmich, zeile) {
		t.Errorf("das README nennt den Beleg nicht so, wie er ist — erwartet die Zeile\n%s", zeile)
	}

	// **Zwei Spalten, zwei Kohorten** (QA-02.AK-04): die feste des
	// Befundlaufs und die Zufallsstichprobe, nebeneinander und nicht vermischt.
	befunde := []struct{ bezeichnung, fest, zufall string }{
		{"Karten gezeichnet", "kanon.karten", "zufall.karten"},
		{"Befunde der Art `Portierung`", "kanon.portierung", "zufall.portierung"},
		{"Befunde der Art `Vorlage`", "kanon.vorlage", "zufall.vorlage"},
		{"Befunde der Art `Umgebung`", "kanon.umgebung", "zufall.umgebung"},
	}

	for _, einer := range befunde {
		erwartet := fmt.Sprintf("| %s | **%s** | **%s** |", einer.bezeichnung,
			deutsch(kennzahl(t, beleg, einer.fest), 0), deutsch(kennzahl(t, beleg, einer.zufall), 0))
		if !strings.Contains(liesmich, erwartet) {
			t.Errorf("das README nennt %s nicht wie der Beleg — erwartet\n%s",
				einer.bezeichnung, erwartet)
		}
	}
}

// **Die feste Kohorte ist der Beleg** — beschreibend, jede Zahl auf drei
// Stellen. **Ohne Band**: ein Vertrauensband setzt eine Zufallsauswahl voraus,
// und die feste Kohorte ist systematisch gezogen (QA-02.AK-03).
func TestDieFesteKohorteImReadmeIstDerBeleg(t *testing.T) {
	t.Parallel()

	liesmich := lies(t, "README.md")
	beleg := kanonbeleg(t)

	wert := func(name string) string { return deutsch(kennzahl(t, beleg, "mse2."+name), 3) }
	ganz := kennzahl(t, beleg, "mse2.grundgesamtheit")
	gezogen := kennzahl(t, beleg, "mse2.stichprobe")

	for _, zeile := range []string{
		fmt.Sprintf("| Grundgesamtheit | **%s** Karten in zehn Sets |", deutsch(ganz, 0)),
		fmt.Sprintf("| Feste Kohorte | **%s** Karten (%s %%) |", deutsch(gezogen, 0), anteilVon(gezogen, ganz)),
		fmt.Sprintf("| **Kanonmittel, gewichtet** | **%s** |", wert("kanonmittel")),
		fmt.Sprintf("| Median | %s | Standardabweichung | %s |", wert("median"), wert("standardabweichung")),
		fmt.Sprintf("| Minimum | %s | Maximum | %s |", wert("minimum"), wert("maximum")),
		fmt.Sprintf("| Q1 (25 %%) | %s | Q3 (75 %%) | %s |", wert("q1"), wert("q3")),
		fmt.Sprintf("| Interquartilsabstand | %s | Streuung | %s |", wert("iqa"), wert("streuung")),
		fmt.Sprintf("| P90 | %s | P95 | %s |", wert("p90"), wert("p95")),
		fmt.Sprintf("| **P99** | **%s** | | |", wert("p99")),
	} {
		if !strings.Contains(liesmich, zeile) {
			t.Errorf("das README zeigt nicht, was der Beleg sagt — erwartet die Zeile\n%s", zeile)
		}
	}
}

// **Die Zufallsstichprobe ist der Beleg** — mit Standardfehler und Band, denn
// ihre Ziehung erfüllt, was beide voraussetzen, und mit der Sicherheit je
// Satz und über alle zugleich (QA-02.AK-03).
func TestDieZufallsstichprobeImReadmeIstDerBeleg(t *testing.T) {
	t.Parallel()

	liesmich := lies(t, "README.md")
	beleg := kanonbeleg(t)

	wert := func(name string) string { return deutsch(kennzahl(t, beleg, "zufall."+name), 3) }
	ganz := kennzahl(t, beleg, "zufall.grundgesamtheit")
	gezogen := kennzahl(t, beleg, "zufall.stichprobe")

	for _, zeile := range []string{
		fmt.Sprintf("| Zufallsstichprobe | **%s** Karten (%s %%) |", deutsch(gezogen, 0), anteilVon(gezogen, ganz)),
		fmt.Sprintf("| **Kanonmittel** | **%s** |", wert("kanonmittel")),
		fmt.Sprintf("| Standardfehler | %s |", wert("standardfehler")),
		fmt.Sprintf("| **95-%%-Band** | **[%s · %s]** |", wert("band.unten"), wert("band.oben")),
		fmt.Sprintf("| Sicherheit je Satz, mindestens | %s %% |",
			deutsch(100*kennzahl(t, beleg, "zufall.sicherheit.je"), 2)),
		fmt.Sprintf("| Sicherheit über alle zehn zugleich | %s %% |",
			deutsch(100*kennzahl(t, beleg, "zufall.sicherheit.gemeinsam"), 2)),
	} {
		if !strings.Contains(liesmich, zeile) {
			t.Errorf("das README zeigt nicht, was der Beleg sagt — erwartet die Zeile\n%s", zeile)
		}
	}
}

// anteilVon gibt einen Anteil in Prozent, auf zwei Stellen.
func anteilVon(teil, ganz float64) string {
	return deutsch(math.Round(10000*teil/ganz)/100, 2)
}

// **Die Zahlen werden geschrieben, wie das README sie schreibt** — die
// Gegenprobe, ohne die jede Zeile oben an der Schreibweise statt an der Zahl
// scheitern könnte.
func TestDeutscheZahlenWerdenWieImReadmeGeschrieben(t *testing.T) {
	t.Parallel()

	for wert, erwartet := range map[string]string{
		deutsch(25148, 0):    "25.148",
		deutsch(679, 0):      "679",
		deutsch(20.0556, 3):  "20,056",
		deutsch(0.0581, 3):   "0,058",
		deutsch(1234567, 0):  "1.234.567",
		deutsch(16.9596, 2):  "16,96",
		deutsch(6286, 0):     "6.286",
		deutsch(39.38849, 3): "39,388",
	} {
		if wert != erwartet {
			t.Errorf("geschrieben %q, erwartet %q", wert, erwartet)
		}
	}
}
