//go:build fremdpakete

package app

// **MSE2 zeichnet die Referenz mit denselben Schriften wie wir** — QA-03.AK-02
// und AK-03 (#1656).
//
// # Warum
//
// Unser Rasterer nimmt eine Schrift aus dem Pack (`MSE_SCHRIFTWURZEL`), MSE2
// die installierte (`MSE2_SCHRIFTEN`, das Verzeichnis aus
// `scripts/mse2-referenz.sh`). Fehlt sie dort, weicht MSE2 still auf eine
// Ersatzschrift aus: die Referenz zeigt eine Karte, die MSE2 mit den Schriften
// des Packs nie zeichnen würde, und jeder Abstand an ihr hieße unser Fehler.
// Liegt dort eine andere Fassung, gilt dasselbe, nur leiser. **Hier stand bis
// QA-03 eine Prüfung auf zwei Dateinamen** — und keine, wenn das Verzeichnis
// ganz fehlte.
//
// # Was er hält
//
// Jede Schriftfamilie, die der Plan einer Karte der festen Kohorte anfordert —
// Text, Kursivtext, Text auf Symbolen —, und die wir **aus dem Pack** nehmen,
// liegt in `MSE2_SCHRIFTEN` als **dieselbe Datei**. Eine Systemschrift sehen
// beide gleich und zählt nicht.
//
// # Was er nicht hält
//
// **Den fetten Schnitt**: gefragt wird der aufrechte; ob ein Stück fett ist,
// steht im Satz und nicht im Plan. **Und wie MSE2 einen Namen auflöst** —
// CoreText über wxWidgets, hier die Namenstabelle: dieselbe Frage, zwei Wege.

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/image/font/sfnt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

func TestDieReferenzschriftenReichen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	schriftwurzel := schriftwurzelDesLaufs()
	installiert := os.Getenv("MSE2_SCHRIFTEN")

	if wurzel == "" || schriftwurzel == "" || installiert == "" {
		t.Skip("MSE_FREMDPAKETE, MSE_SCHRIFTWURZEL und MSE2_SCHRIFTEN sind nicht alle gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })
	namen := angeforderteSchriften(t, anwendung)

	if len(namen) == 0 {
		t.Fatal("keine einzige Schrift angefordert — der Leser greift daneben")
	}

	unsere := schriften.Neue(schriften.Orte()).MitPaketen(schriftwurzel)
	ihre := schriften.Neue([]string{installiert})
	ausDemPack := 0

	for _, name := range namen {
		ort, fehler := unsere.Fundort(name, false)
		if fehler != nil {
			t.Logf("%s: auch dieser Rasterer findet sie nicht — das meldet der Kanonlauf als Umgebung", name)

			continue
		}

		if !strings.HasPrefix(ort, filepath.Clean(schriftwurzel)+string(filepath.Separator)) {
			continue
		}

		ausDemPack++

		ihrOrt, fehler := ihre.Fundort(name, false)
		if fehler != nil {
			t.Errorf("%s: MSE2 findet sie nicht in %s — die Referenz zeichnet sie mit einer Ersatzschrift "+
				"(wir nehmen %s)", name, installiert, ort)

			continue
		}

		hash := dateihash(t, ort)
		if hash != dateihash(t, ihrOrt) {
			t.Errorf("%s: zwei Fassungen — wir %s, MSE2 %s", name, ort, ihrOrt)
		}

		// **Datei, Fassung und Hash ins Protokoll** — QA-03.AK-02 verlangt
		// sie je Schrift, und ein Beleg, der nur zählt, nennt sie nicht.
		t.Logf("%s: %s, %s, sha256 %s", name, filepath.Base(ort), schriftfassung(t, ort), hash[:16])
	}

	t.Logf("%d Schriftfamilien angefordert, %d davon aus dem Pack", len(namen), ausDemPack)
}

// schriftfassung gibt die Fassung aus der Namenstabelle einer Schriftdatei.
func schriftfassung(t *testing.T, pfad string) string {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // eine gefundene Schriftdatei
	if fehler != nil {
		t.Fatalf("%s: %v", pfad, fehler)
	}

	schrift, fehler := sfnt.Parse(roh)
	if fehler != nil {
		return "nicht lesbar: " + fehler.Error()
	}

	fassung, fehler := schrift.Name(nil, sfnt.NameIDVersion)
	if fehler != nil {
		return "ohne Fassung"
	}

	return fassung
}

// angeforderteSchriften gibt die Schriftfamilien, die die Pläne der festen
// Kohorte anfordern — dieselben Karten wie im Kanonlauf.
func angeforderteSchriften(t *testing.T, anwendung *App) []string {
	t.Helper()

	eintraege, fehler := os.ReadDir(kanonverzeichnis())
	if fehler != nil {
		t.Fatalf("die kanonischen Sets: %v", fehler)
	}

	jeSet := stichprobengroesse(t, zaehleSets(eintraege))
	gefunden := map[string]bool{}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".mse-set") {
			continue
		}

		pfad := filepath.Join(kanonverzeichnis(), eintrag.Name())

		geoeffnet := anwendung.SetOeffnen(t.Context(), pfad)
		if geoeffnet.Fehler != nil {
			t.Fatalf("%s öffnen: %+v", eintrag.Name(), geoeffnet.Fehler)
		}

		liste := anwendung.Karten(t.Context(), geoeffnet.SetID)
		if liste.Fehler != nil {
			t.Fatalf("%s: Karten: %+v", eintrag.Name(), liste.Fehler)
		}

		for _, karte := range geschichteteStichprobe(liste.Karten, stilschichten(t, pfad), jeSet) {
			schriftenImPlan(planZurKarte(t, anwendung, karte.ID), gefunden)
		}
	}

	namen := make([]string, 0, len(gefunden))
	for name := range gefunden {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	return namen
}

// schriftenImPlan trägt die Schriftfamilien eines Plans ein.
func schriftenImPlan(plan *render.Plan, gefunden map[string]bool) {
	for _, ebene := range plan.Ebenen {
		for _, name := range []string{ebene.Satz.Schrift, ebene.Satz.Kursivschrift} {
			if strings.TrimSpace(name) != "" {
				gefunden[name] = true
			}
		}

		for _, abschnitt := range ebene.Abschnitte {
			for _, symbol := range abschnitt.Symbole {
				if name := symbol.Textsatz.Schrift; strings.TrimSpace(name) != "" {
					gefunden[name] = true
				}
			}
		}
	}
}

// dateihash gibt den SHA-256 einer Datei.
func dateihash(t *testing.T, pfad string) string {
	t.Helper()

	datei, fehler := os.Open(pfad) //nolint:gosec // eine gefundene Schriftdatei
	if fehler != nil {
		t.Fatalf("%s: %v", pfad, fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	summe := sha256.New()
	if _, fehler := io.Copy(summe, datei); fehler != nil {
		t.Fatalf("%s: %v", pfad, fehler)
	}

	return hex.EncodeToString(summe.Sum(nil))
}
