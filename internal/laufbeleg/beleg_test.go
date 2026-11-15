package laufbeleg_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// einKopf gibt den Kopf eines Laufs an einem sauberen Stand.
func einKopf() laufbeleg.Kopf {
	anfang := time.Date(2026, 9, 22, 23, 0, 0, 0, time.UTC)

	return laufbeleg.Kopf{
		Lauf: "fremdpakete", Commit: "7a1c9152", Inhalt: "sha256:inhalt", BaumSauber: true,
		Werkzeuge: map[string]string{"go": "go1.27.1", "plattform": "darwin/arm64"},
		Anfang:    anfang, Ende: anfang.Add(40 * time.Minute),
		Befehl: []string{"go", "test", "-json"}, Rueckgabe: 0,
	}
}

// eineAuswertung gibt eine grüne Auswertung über zwei Fälle, einer davon
// übersprungen und erlaubt.
func eineAuswertung() laufbeleg.Auswertung {
	return laufbeleg.Auswertung{
		Status: laufbeleg.Bestanden,
		Faelle: []laufbeleg.Fallergebnis{
			{
				Fall: laufbeleg.Fall{Paket: "p", Test: "TestA"}, Art: laufbeleg.Pflicht,
				Ausfuehrung: laufbeleg.Bestanden, Ergebnis: laufbeleg.Bestanden,
			},
			{
				Fall: laufbeleg.Fall{Paket: "p", Test: "TestB"}, Art: laufbeleg.Optional,
				Ausfuehrung: laufbeleg.Uebersprungen, Ergebnis: laufbeleg.Uebersprungen,
				Grund: "Vorrichtung",
			},
		},
		Kennzahlen: map[string]string{"kanon.portierung": "0"},
	}
}

// eineEingabe gibt eine vorhandene Eingabe.
func eineEingabe() laufbeleg.Eingabe {
	return laufbeleg.Eingabe{Name: "MSE_KANON", Pfad: "~/kanon", Hash: "sha256:ab", Dateien: 10, Passend: 10}
}

// **Was der Beleg aus seinen Teilen macht** (AK-01, AK-02).
func TestDerBelegFasstZusammen(t *testing.T) {
	t.Parallel()

	fehlend := eineEingabe()
	fehlend.Fehlt = "leer"

	ohneCommit := einKopf()
	ohneCommit.Commit = ""

	ohneInhalt := einKopf()
	ohneInhalt.Inhalt = ""

	faelle := map[string]struct {
		kopf      laufbeleg.Kopf
		eingaben  []laufbeleg.Eingabe
		artefakte []laufbeleg.Eingabe
		status    laufbeleg.Status
		grund     string
	}{
		"alles da": {
			einKopf(),
			[]laufbeleg.Eingabe{eineEingabe()},
			nil,
			laufbeleg.Bestanden, "",
		},
		// **Der Fall aus #1656**: das Verzeichnis lag da, und es war leer.
		"eine Pflichteingabe fehlt": {
			einKopf(),
			[]laufbeleg.Eingabe{fehlend},
			nil,
			laufbeleg.Blockiert, "MSE_KANON: leer",
		},
		"ein Artefakt fehlt": {
			einKopf(), nil,
			[]laufbeleg.Eingabe{{Name: "Rohprotokoll", Fehlt: "liegt nicht vor"}},
			laufbeleg.Gescheitert, "das Artefakt Rohprotokoll fehlt: liegt nicht vor",
		},
		// **Ein Beleg ohne Stand belegt nichts**: welcher Bau gemessen wurde,
		// wüsste niemand.
		"kein Commit": {
			ohneCommit, nil, nil,
			laufbeleg.Gescheitert, "ohne Commit ist der Stand nicht benannt",
		},
		// **Ohne Inhaltshash wüsste niemand, ob er noch passt** — der Commit
		// allein überlebt keinen Squash-Merge.
		"kein Inhalt": {
			ohneInhalt, nil, nil,
			laufbeleg.Gescheitert, "ohne Inhaltshash",
		},
		// **FAIL schlägt BLOCKED**: fehlt etwas und ist zugleich etwas rot,
		// ist der Lauf zuerst gescheitert.
		"beides": {
			einKopf(),
			[]laufbeleg.Eingabe{fehlend},
			[]laufbeleg.Eingabe{{Name: "Rohprotokoll", Fehlt: "leer"}},
			laufbeleg.Gescheitert, "das Artefakt Rohprotokoll fehlt",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			beleg := laufbeleg.Belegen(einer.kopf, einer.eingaben, eineAuswertung(), einer.artefakte)

			if beleg.Status != einer.status || !strings.Contains(beleg.Grund, einer.grund) {
				t.Errorf("Status %s, Grund %q — erwartet %s und %q",
					beleg.Status, beleg.Grund, einer.status, einer.grund)
			}

			if beleg.Erwartet != 2 || beleg.Gemessen != 1 ||
				len(beleg.Uebersprungen) != 1 || beleg.Uebersprungen[0] != "p.TestB: Vorrichtung" {
				t.Errorf("erwartet %d, gemessen %d, übersprungen %q",
					beleg.Erwartet, beleg.Gemessen, beleg.Uebersprungen)
			}
		})
	}
}

// **Ein leerer Beleg schreibt leere Listen und kein `null`** — sonst sähe
// ein Leser »nichts gelaufen« und »nichts gesagt« gleich.
func TestEinLeererBelegSchreibtLeereListen(t *testing.T) {
	t.Parallel()

	// **Ein Kopf ohne Werkzeuge und ohne Befehl** — auch dann `{}` und `[]`.
	beleg := laufbeleg.Belegen(laufbeleg.Kopf{Lauf: "leer", Commit: "abc", Inhalt: "sha256:x", BaumSauber: true},
		nil, laufbeleg.Auswertung{}, nil)

	// **Eine leere Auswertung ist kein grüner Lauf** (AK-02).
	if beleg.Status != laufbeleg.NichtGelaufen || beleg.Grund != "keine Auswertung" {
		t.Errorf("Status %s, Grund %q — erwartet NOT_RUN", beleg.Status, beleg.Grund)
	}

	var puffer bytes.Buffer
	if fehler := beleg.Schreiben(&puffer); fehler != nil {
		t.Fatal(fehler)
	}

	// **Die Listen müssen dastehen** — ein Beleg, der gar nichts schreibt,
	// enthielte auch kein `null`, und die Prüfung wäre umsonst grün.
	for _, leer := range []string{
		`"faelle": []`, `"eingaben": []`, `"uebersprungen": []`,
		`"artefakte": []`, `"fehler": []`, `"kennzahlen": {}`, `"werkzeuge": {}`, `"befehl": []`,
	} {
		if !strings.Contains(puffer.String(), leer) {
			t.Errorf("der Beleg nennt %s nicht:\n%s", leer, puffer.String())
		}
	}

	if strings.Contains(puffer.String(), "null") {
		t.Errorf("der Beleg enthält null:\n%s", puffer.String())
	}
}

// **Geschrieben und gelesen ist derselbe Beleg.**
func TestEinBelegKommtUnveraendertZurueck(t *testing.T) {
	t.Parallel()

	geschrieben := laufbeleg.Belegen(einKopf(), []laufbeleg.Eingabe{eineEingabe()}, eineAuswertung(), nil)

	var puffer bytes.Buffer
	if fehler := geschrieben.Schreiben(&puffer); fehler != nil {
		t.Fatal(fehler)
	}

	gelesen, fehler := laufbeleg.Lesen(&puffer)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen.Commit != "7a1c9152" || gelesen.Status != laufbeleg.Bestanden ||
		!gelesen.Ende.Equal(geschrieben.Ende) || gelesen.Kennzahlen["kanon.portierung"] != "0" ||
		len(gelesen.Faelle) != 2 || gelesen.Eingaben[0].Hash != "sha256:ab" {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Was kein Beleg dieser Gestalt ist, wird abgewiesen** — ein Tippfehler in
// einem Feldnamen hieße sonst, dass eine Angabe still fehlt.
func TestEinFremderBelegWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein JSON":            "Beleg",
		"ein anderes Format":   `{"format": 2, "status": "PASS"}`,
		"ein fremder Stand":    `{"format": 1, "status": "GRUEN"}`,
		"ein unbekanntes Feld": `{"format": 1, "status": "PASS", "stauts": "PASS"}`,
	}

	for name, text := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := laufbeleg.Lesen(strings.NewReader(text)); fehler == nil {
				t.Errorf("%q ging als Beleg durch", text)
			}
		})
	}
}

// **Ein weiterer Fall zählt nicht zu den erwarteten** — er steht im Beleg,
// weil er rot war, und nicht, weil jemand ihn bestellt hat.
func TestEinWeitererFallZaehltNichtMit(t *testing.T) {
	t.Parallel()

	auswertung := eineAuswertung()
	auswertung.Faelle = append(auswertung.Faelle, laufbeleg.Fallergebnis{
		Fall: laufbeleg.Fall{Paket: "p", Test: "TestFremd"}, Art: laufbeleg.Weitere,
		Ausfuehrung: laufbeleg.Bestanden, Ergebnis: laufbeleg.Gescheitert,
	})

	if beleg := laufbeleg.Belegen(einKopf(), nil, auswertung, nil); beleg.Erwartet != 2 ||
		beleg.Gemessen != 1 || len(beleg.Faelle) != 3 {
		t.Errorf("erwartet %d, gemessen %d, Fälle %d", beleg.Erwartet, beleg.Gemessen, len(beleg.Faelle))
	}
}

// **Was JSON nicht fassen kann, wird ein Fehler** — und kein halber Beleg.
func TestEinUnfassbarerWertIstEinFehler(t *testing.T) {
	t.Parallel()

	auswertung := eineAuswertung()
	auswertung.Faelle[0].Dauer = math.NaN()

	var puffer bytes.Buffer
	if fehler := laufbeleg.Belegen(einKopf(), nil, auswertung, nil).Schreiben(&puffer); fehler == nil ||
		puffer.Len() != 0 {
		t.Errorf("Fehler %v, geschrieben %d Bytes — erwartet einen Fehler und nichts", fehler, puffer.Len())
	}
}

// **Der Beleg wird ganz geschrieben oder gar nicht** — er wird eingecheckt,
// und ein halber sähe aus wie ein ganzer.
func TestDerBelegWirdAtomarGeschrieben(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ziel := filepath.Join(wurzel, "docs", "belege", "lauf.json")
	beleg := laufbeleg.Belegen(einKopf(), nil, eineAuswertung(), nil)

	if fehler := beleg.SchreibeNach(ziel); fehler != nil {
		t.Fatal(fehler)
	}

	datei, fehler := os.Open(ziel) //nolint:gosec // ein Pfad unter TempDir
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	if gelesen, lesefehler := laufbeleg.Lesen(datei); lesefehler != nil || gelesen.Commit != "7a1c9152" {
		t.Errorf("gelesen %+v, %v", gelesen, lesefehler)
	}

	// **Und neben ihm bleibt nichts liegen.**
	eintraege, fehler := os.ReadDir(filepath.Dir(ziel))
	if fehler != nil || len(eintraege) != 1 {
		t.Errorf("im Verzeichnis liegen %d Dateien, erwartet nur den Beleg (%v)", len(eintraege), fehler)
	}
}

// **Was nicht zu schreiben ist, lässt den alten Beleg stehen** — und nichts
// Halbes daneben.
func TestEinUnschreibbarerBelegLaesstNichtsZurueck(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	schreibe(t, wurzel, "datei", "kein Verzeichnis")

	unfassbar := eineAuswertung()
	unfassbar.Faelle[0].Dauer = math.NaN()

	nurlesen := filepath.Join(wurzel, "nurlesen")
	if fehler := os.MkdirAll(nurlesen, 0o500); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(nurlesen, 0o750) }) //nolint:errcheck,gosec // nur zum Aufräumen; ein Verzeichnis braucht das Ausführungsrecht

	faelle := map[string]struct {
		ziel  string
		beleg laufbeleg.Beleg
	}{
		"unter einer Datei":                      {filepath.Join(wurzel, "datei", "lauf.json"), laufbeleg.Belegen(einKopf(), nil, eineAuswertung(), nil)},
		"über einem Verzeichnis":                 {wurzel, laufbeleg.Belegen(einKopf(), nil, eineAuswertung(), nil)},
		"mit einem unfassbaren Wert":             {filepath.Join(wurzel, "neu", "lauf.json"), laufbeleg.Belegen(einKopf(), nil, unfassbar, nil)},
		"in einem Verzeichnis ohne Schreibrecht": {filepath.Join(nurlesen, "lauf.json"), laufbeleg.Belegen(einKopf(), nil, eineAuswertung(), nil)},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := einer.beleg.SchreibeNach(einer.ziel); fehler == nil ||
				!strings.Contains(fehler.Error(), einer.ziel) {
				t.Errorf("Fehler %v, erwartet einen, der %s nennt", fehler, einer.ziel)
			}
		})
	}

	reste, fehler := filepath.Glob(filepath.Join(wurzel, "*", ".beleg-*"))
	if fehler != nil || len(reste) > 0 {
		t.Errorf("liegen geblieben: %v (%v)", reste, fehler)
	}
}
