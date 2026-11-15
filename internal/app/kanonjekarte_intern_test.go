package app

// **Der Kanon Karte für Karte gegen einen früheren Lauf** — F02
// (#2244).
//
// F02 verlangt, dass **kein einzelner Fall schlechter** wird als eine
// bestätigte Toleranz. Mittel, Median und P99 sagen darüber nichts: eine
// Änderung, die tausend Karten bessert und zehn verschlechtert, senkt alle
// drei. Die Stichprobe ist ohne Zufall gezogen (MESSMETHODIK §4), deshalb
// ist dieselbe Stelle im nächsten Lauf dieselbe Karte.
//
// **Ohne Build-Tag**: die Rechnung braucht keine fremden Pakete, und so läuft
// sie in jeder Pipeline.

import (
	"bytes"
	"strings"
	"testing"
)

func TestDieKartenwerteGehenHinUndZurueck(t *testing.T) {
	t.Parallel()

	werte := []kartenwert{
		{Set: "Archenemy", Stelle: 0, Name: "All in Good Time", Abstand: 13.5},
		{Set: "Black", Stelle: 17, Name: "Name\tmit Tab", Abstand: 21.125},
	}

	var puffer bytes.Buffer
	if err := schreibeKartenwerte(&puffer, werte); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := liesKartenwerte(&puffer)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 2 || gelesen[werte[0].schluessel()] != 13.5 || gelesen[werte[1].schluessel()] != 21.125 {
		t.Errorf("gelesen: %v", gelesen)
	}
}

func TestEineVerdorbeneZeileIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, zeile := range []string{"Black\t1\tName\n", "Black\teins\tName\t2.0\n", "Black\t1\tName\tviel\n"} {
		if _, err := liesKartenwerte(strings.NewReader(zeile)); err == nil {
			t.Errorf("%q wurde ohne Fehler gelesen", zeile)
		}
	}
}

func TestDerVergleichNenntJedeVerschlechterung(t *testing.T) {
	t.Parallel()

	vorher := map[string]float64{
		"A\t0\teins":  10,
		"A\t1\tzwei":  10,
		"A\t2\tdrei":  10,
		"A\t3\tvier":  10,
		"A\t4\tweg":   10,
		"B\t0\tfünf":  20,
		"B\t1\tsechs": 20,
	}

	nachher := []kartenwert{
		{Set: "A", Stelle: 0, Name: "eins", Abstand: 9},       // besser
		{Set: "A", Stelle: 1, Name: "zwei", Abstand: 10.004},  // in der Toleranz
		{Set: "A", Stelle: 2, Name: "drei", Abstand: 10.5},    // schlechter
		{Set: "A", Stelle: 3, Name: "vier", Abstand: 10},      // gleich
		{Set: "B", Stelle: 0, Name: "fünf", Abstand: 23},      // schlechter, am meisten
		{Set: "B", Stelle: 1, Name: "sechs", Abstand: 19.995}, // in der Toleranz
		{Set: "C", Stelle: 0, Name: "neu", Abstand: 5},        // ohne Gegenstück
	}

	ergebnis := vergleicheKartenwerte(vorher, nachher, 0.01)

	if ergebnis.verglichen != 6 || ergebnis.besser != 1 || ergebnis.gleich != 3 || ergebnis.ohneVorher != 1 || ergebnis.ohneNachher != 1 {
		t.Errorf("gezählt: %+v", ergebnis)
	}

	if len(ergebnis.schlechter) != 2 || ergebnis.schlechter[0].Name != "fünf" || ergebnis.schlechter[1].Name != "drei" {
		t.Fatalf("schlechter: %+v — erwartet fünf vor drei", ergebnis.schlechter)
	}

	if ergebnis.schlechter[0].Vorher != 20 || ergebnis.schlechter[0].Abstand != 23 {
		t.Errorf("die schlechteste trägt %+v", ergebnis.schlechter[0])
	}
}
