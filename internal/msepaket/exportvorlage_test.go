package msepaket_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Exportvorlage ist selbst skriptgesteuert** (#51).
//
// # Was der Leser herausgibt und was nicht
//
// Er liest den **Kopf**: Name, Spiel, Dateityp, die Wahlmöglichkeiten und den
// Skripttext. Was das Skript **tut**, weiß `msescript`, und wer es auswertet,
// ist `service` — dieselbe Naht wie beim Stil (§2).
//
// Gezählt über die 19 Exportvorlagen von `M15-Magic-Pack`: alle 19 führen
// `script:`, `game:` und `depends on:`; 18 führen `file type:`, 16 `create
// directory:`, und 78 `option field:`-Blöcke verteilen sich auf sie.
func TestEineExportvorlageWirdGelesen(t *testing.T) {
	vorlage := exportvorlage(t)

	if vorlage.KurzerName != "Probe-Export" || vorlage.VollerName != "Ausfuhr zur Probe" {
		t.Errorf("Namen %q / %q", vorlage.KurzerName, vorlage.VollerName)
	}

	if vorlage.Spiel != "probe" {
		t.Errorf("Spiel %q", vorlage.Spiel)
	}

	if vorlage.Dateityp != "*.txt|*.txt" {
		t.Errorf("Dateityp %q", vorlage.Dateityp)
	}

	// **`create directory: false` heißt eine Datei.** Vier der Vorlagen des
	// Packs sagen das, zehn das Gegenteil — wer die Angabe übergeht, legt für
	// eine Textdatei ein Verzeichnis an.
	if vorlage.Verzeichnis {
		t.Error("Verzeichnis, erwartet war eine Datei")
	}

	if !strings.Contains(vorlage.Skript, "for each card in cards") {
		t.Errorf("Skript:\n%s", vorlage.Skript)
	}
}

// **Die Wahlmöglichkeiten kommen mit** — sie sind die Eingabe des Skripts.
//
// `options.nur_namen` steht so im Skript; ohne die Blöcke wüsste niemand,
// welche Namen es gibt, welche Werte erlaubt sind und was gilt, wenn niemand
// wählt.
func TestEineExportvorlageNenntIhreWahlmoeglichkeiten(t *testing.T) {
	optionen := exportvorlage(t).Optionen

	if len(optionen) != 2 {
		t.Fatalf("%d Wahlmöglichkeiten, erwartet zwei: %+v", len(optionen), optionen)
	}

	if optionen[0].Name != "kopfzeile" || optionen[0].Art != "text" ||
		optionen[0].Vorgabe != "Meine Karten" {
		t.Errorf("erste: %+v", optionen[0])
	}

	if optionen[0].Beschreibung != "Was über der Liste steht" {
		t.Errorf("Beschreibung %q", optionen[0].Beschreibung)
	}

	// **`initial:` bei einer Auswahl, `default:` bei einem Text.** Beide
	// stehen in den Vorlagen des Packs, und beide meinen dasselbe: was gilt,
	// wenn niemand wählt.
	if optionen[1].Art != "choice" || optionen[1].Vorgabe != "nein" {
		t.Errorf("zweite: %+v", optionen[1])
	}

	if len(optionen[1].Auswahl) != 2 || optionen[1].Auswahl[0] != "nein" {
		t.Errorf("Auswahl %v", optionen[1].Auswahl)
	}
}

// exportvorlage liest das Prüfstück.
func exportvorlage(t *testing.T) msepaket.Exportvorlage {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(testdaten, "probe-export.mse-export-template"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return paket.Exportmodell()
}
