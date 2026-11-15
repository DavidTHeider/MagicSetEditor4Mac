package app

// **Der Plan der Zufallsstichprobe** — QA-02.AK-01
// (#1734):
// „ein tatsächlich zufällig gezogener, versionierter und reproduzierbar
// gespeicherter geschichteter Stichprobenplan".
//
// Ohne Build-Tag: was hier geprüft wird, braucht keine fremden Pakete.

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

// **Dieselbe Saat, derselbe Plan** — und aus dem Plan lässt sich die Ziehung
// nachrechnen. Eine andere Saat zieht andere Karten.
func TestDerPlanFolgtAusSaatUndGroessen(t *testing.T) {
	t.Parallel()

	groessen := []planschicht{{Satz: "Vanguard", Karten: 107}, {Satz: "Black", Karten: 3788}}

	erst := ziehePlan([2]uint64{1, 2}, groessen)
	wieder := ziehePlan([2]uint64{1, 2}, groessen)
	anders := ziehePlan([2]uint64{3, 4}, groessen)

	if !slices.Equal(erst.Schichten[1].Stichprobe, wieder.Schichten[1].Stichprobe) {
		t.Error("dieselbe Saat zog andere Karten")
	}

	if slices.Equal(erst.Schichten[1].Stichprobe, anders.Schichten[1].Stichprobe) {
		t.Error("eine andere Saat zog dieselben Karten")
	}

	for _, schicht := range erst.Schichten {
		pruefeSchicht(t, schicht)
	}

	// **Die gemeinsame Sicherheit ist das Produkt der Schichten** (AK-03).
	if erst.Gemeinsam != gemeinsameSicherheit([]float64{
		erst.Schichten[0].Findewahrscheinlichkeit, erst.Schichten[1].Findewahrscheinlichkeit,
	}) {
		t.Errorf("gemeinsam %v", erst.Gemeinsam)
	}
}

// pruefeSchicht hält, was eine gezogene Schicht sein muss.
func pruefeSchicht(t *testing.T, schicht planschicht) {
	t.Helper()

	noetig := noetigeStichprobe(0.999, 0.01, schicht.Karten)
	if len(schicht.Stichprobe) != noetig {
		t.Errorf("%s: %d gezogen, nötig %d", schicht.Satz, len(schicht.Stichprobe), noetig)
	}

	if !slices.IsSorted(schicht.Stichprobe) {
		t.Errorf("%s: die Stellen stehen nicht der Reihe nach", schicht.Satz)
	}

	for i, stelle := range schicht.Stichprobe {
		if stelle < 0 || stelle >= schicht.Karten || (i > 0 && schicht.Stichprobe[i-1] == stelle) {
			t.Errorf("%s: die Stelle %d ist keine oder doppelt", schicht.Satz, stelle)
		}
	}

	if schicht.Betroffen != betroffeneBei(0.01, schicht.Karten) ||
		schicht.Findewahrscheinlichkeit < 0.999 {
		t.Errorf("%s: betroffen %d, Findewahrscheinlichkeit %v", schicht.Satz,
			schicht.Betroffen, schicht.Findewahrscheinlichkeit)
	}
}

// **Geschrieben und gelesen ist derselbe Plan** — und ein fremder wird
// abgewiesen.
func TestDerPlanKommtUnveraendertZurueck(t *testing.T) {
	t.Parallel()

	plan := ziehePlan([2]uint64{7, 8}, []planschicht{{Satz: "Archenemy", Karten: 70, Kartenhash: "sha256:ab"}})

	var puffer bytes.Buffer
	if fehler := plan.schreiben(&puffer); fehler != nil {
		t.Fatal(fehler)
	}

	gelesen, fehler := liesPlan(strings.NewReader(puffer.String()))
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen.Saat != plan.Saat || gelesen.Fassung != 1 ||
		!slices.Equal(gelesen.Schichten[0].Stichprobe, plan.Schichten[0].Stichprobe) ||
		gelesen.Schichten[0].Kartenhash != "sha256:ab" {
		t.Errorf("zurück kam %+v", gelesen)
	}

	for name, fremd := range map[string]string{
		"kein JSON":        "{",
		"eine andere Form": `{"fassung": 2}`,
		"ein fremdes Feld": `{"fassung": 1, "saaat": "1"}`,
	} {
		if _, fehler := liesPlan(strings.NewReader(fremd)); fehler == nil {
			t.Errorf("%s ging als Plan durch", name)
		}
	}
}

// **Der Kartenhash sieht jede Änderung an der Liste** — auch eine andere
// Reihenfolge, denn der Plan nennt Stellen und keine Namen.
func TestDerKartenhashSiehtJedeAenderung(t *testing.T) {
	t.Parallel()

	grund := kartenhash([]string{"Ajani", "Bolas", "Chandra"})

	for name, andere := range map[string][]string{
		"eine andere Reihenfolge": {"Bolas", "Ajani", "Chandra"},
		"eine Karte mehr":         {"Ajani", "Bolas", "Chandra", "Dack"},
		"ein anderer Name":        {"Ajani", "Bolas", "Chandra Nalaar"},
		"zwei Namen zu einem":     {"Ajani\nBolas", "Chandra"},
	} {
		if kartenhash(andere) == grund {
			t.Errorf("%s ließ den Hash gleich", name)
		}
	}

	if !strings.HasPrefix(grund, "sha256:") || kartenhash([]string{"Ajani", "Bolas", "Chandra"}) != grund {
		t.Errorf("der Hash %q ist nicht stabil oder hat nicht die Gestalt sha256:…", grund)
	}
}
