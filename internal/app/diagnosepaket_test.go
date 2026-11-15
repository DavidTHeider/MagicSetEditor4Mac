//go:build darwin

// **Das Diagnosepaket** — Werkstück O10 von Phase 5.2
// (#1434).
//
// **Nur darwin**, weil `alsLeserHinein` und `gestarteteAnwendung` eine echte
// Anwendung starten (dieselbe Grenze wie bei `diagnoserolle_test.go`).
package app_test

import (
	"archive/zip"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau nennt jeden Teil mit seiner Größe** — vor jeder Zustimmung.
func TestDieVorschauNenntJedenTeilMitSeinerGroesse(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	if len(vorschau.Teile) != len(domain.Diagnoseteile()) {
		t.Fatalf("%d Teile, erwartet %d", len(vorschau.Teile), len(domain.Diagnoseteile()))
	}

	if vorschau.Grenze != domain.Diagnosepaketgrenze {
		t.Errorf("Grenze = %d", vorschau.Grenze)
	}

	summe := 0
	for _, teil := range vorschau.Teile {
		if teil.Datei == "" {
			t.Errorf("der Teil %q hat keinen Dateinamen", teil.Teil)
		}

		summe += teil.Bytes
	}

	if summe != vorschau.Bytes {
		t.Errorf("die Summe der Teile ist %d, die Antwort sagt %d", summe, vorschau.Bytes)
	}
}

// **Was gezeigt wurde, steht im Archiv** — die Abnahme aus dem Issue.
func TestDasArchivTraegtGenauDieGezeigtenTeile(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	inhalte := archivinhalte(t, ziel)

	for _, teil := range vorschau.Teile {
		roh, da := inhalte[teil.Datei]
		if !da {
			t.Errorf("%q fehlt im Archiv", teil.Datei)

			continue
		}

		if len(roh) != teil.Bytes {
			t.Errorf("%q ist %d Byte, die Vorschau sagte %d", teil.Datei, len(roh), teil.Bytes)
		}
	}

	// **Die Übersicht liegt noch einmal darin — und sie ist die Vorschau.**
	var gelesen app.DiagnosepaketAntwort
	if fehler := json.Unmarshal([]byte(inhalte[app.Uebersichtsname]), &gelesen); fehler != nil {
		t.Fatalf("die Übersicht ist kein JSON: %v", fehler)
	}

	if gelesen.Bytes != vorschau.Bytes || len(gelesen.Teile) != len(vorschau.Teile) {
		t.Errorf("die Übersicht im Archiv weicht ab: %+v", gelesen)
	}

	// **Und das Manifest heißt nach diesem Format**, nicht nach der Sicherung.
	if _, da := inhalte[app.Diagnosemanifest]; !da {
		t.Errorf("kein Manifest %q: %v", app.Diagnosemanifest, namen(inhalte))
	}
}

// **Ohne Vorschau wird nicht geschrieben.**
func TestOhneVorschauGibtEsKeinPaket(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)

	antwort := anwendung.DiagnosepaketNach(t.Context(), ziel)
	if antwort.Fehler == nil {
		t.Fatal("ohne Vorschau entstand ein Paket")
	}
}

// **Ein Leser bekommt das Protokoll nicht — und erfährt, warum** (O9/O10).
func TestEinGastBekommtDasProtokollNichtInsPaket(t *testing.T) {
	anwendung := alsLeserHinein(t)

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	for _, teil := range vorschau.Teile {
		if teil.Teil != string(domain.TeilProtokoll) {
			continue
		}

		switch {
		case teil.Bytes != 0:
			t.Errorf("der Gast bekam %d Byte Protokoll", teil.Bytes)
		case teil.Grund == "":
			t.Error("der ausgelassene Teil nennt keinen Grund")
		}

		return
	}

	t.Error("der Protokollteil fehlt ganz — ausgelassen heißt genannt")
}

// **Das Ziel muss die Endung tragen** — an den Flächen ohne Fenster wird sie
// nicht stillschweigend angehängt.
func TestEinZielOhneEndungWirdAbgewiesen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	antwort := anwendung.DiagnosepaketNach(t.Context(), filepath.Join(t.TempDir(), "probe.zip"))
	if antwort.Fehler == nil || antwort.Fehler.Feld != "ziel" {
		t.Fatalf("das Ziel ohne Endung kam durch: %+v", antwort.Fehler)
	}
}

// archivinhalte liest ein Archiv in eine Tabelle.
func archivinhalte(t *testing.T, pfad string) map[string]string {
	t.Helper()

	archiv, fehler := zip.OpenReader(pfad)
	if fehler != nil {
		t.Fatalf("%s öffnen: %v", pfad, fehler)
	}

	defer func() {
		if schliessfehler := archiv.Close(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	}()

	inhalte := map[string]string{}

	for _, eintrag := range archiv.File {
		offen, oeffnungsfehler := eintrag.Open()
		if oeffnungsfehler != nil {
			t.Fatalf("%s öffnen: %v", eintrag.Name, oeffnungsfehler)
		}

		roh, lesefehler := io.ReadAll(offen)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name, lesefehler)
		}

		if schliessfehler := offen.Close(); schliessfehler != nil {
			t.Errorf("%s schließen: %v", eintrag.Name, schliessfehler)
		}

		inhalte[eintrag.Name] = string(roh)
	}

	return inhalte
}

// namen nennt die Einträge eines Archivs.
func namen(inhalte map[string]string) string {
	liste := make([]string, 0, len(inhalte))
	for name := range inhalte {
		liste = append(liste, name)
	}

	return strings.Join(liste, ", ")
}

// **Und das Diagnosepaket nennt dasselbe** — in seinem Teil `fassung.txt`.
func TestDasDiagnosepaketNenntDenselbenBau(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	teil := archivinhalte(t, ziel)["fassung.txt"]

	for _, zeile := range []string{
		"stand\t" + app.DiesenBau().Stand + "\n",
		"bindungsvertrag\t" + app.Vertragsabdruck() + "\n",
	} {
		if !strings.Contains(teil, zeile) {
			t.Errorf("fassung.txt nennt %q nicht:\n%s", zeile, teil)
		}
	}
}
