package schriften_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// systemschrift gibt eine echte Schriftdatei — oder überspringt.
//
// **Keine im Repository** (§11): Schriften sind lizenziert. Der Preis ist, dass
// diese Prüfungen an dieser Maschine hängen; sie sagen das, statt es zu
// verschweigen.
func systemschrift(t *testing.T) []byte {
	t.Helper()

	const pfad = "/System/Library/Fonts/Supplemental/Andale Mono.ttf"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	return roh
}

// verzeichnisMitSchrift legt eine Schriftdatei unter dem gewünschten Namen an.
func verzeichnisMitSchrift(t *testing.T, dateiname string) string {
	t.Helper()

	verzeichnis := t.TempDir()

	if err := os.WriteFile(filepath.Join(verzeichnis, dateiname),
		systemschrift(t), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return verzeichnis
}

func TestEineSchriftWirdUeberIhrenDateinamenGefunden(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "MPlantin.ttf")})

	schrift, err := suche.Suche("MPlantin", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if schrift.Zeilenhoehe() <= 0 {
		t.Errorf("Zeilenhöhe %g", schrift.Zeilenhoehe())
	}
}

// **Groß- und Kleinschreibung zählt nicht.** Eine Vorlage nennt `MPlantin`, die
// Datei heißt `mplantin.ttf` — beides ist dieselbe Schrift.
func TestDieSchreibweiseDesNamensZaehltNicht(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "mplantin.OTF")})

	if _, err := suche.Suche("MPlantin", false); err != nil {
		t.Errorf("suchen: %v", err)
	}
}

// **Eine Schrift wird einmal gelesen.** Eine Karte fragt sie je Textebene.
func TestEineSchriftWirdNurEinmalGelesen(t *testing.T) {
	verzeichnis := verzeichnisMitSchrift(t, "MPlantin.ttf")
	suche := schriften.Neue([]string{verzeichnis})

	erste, err := suche.Suche("MPlantin", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	// Die Datei verschwindet — und die Schrift ist trotzdem noch da.
	if fehler := os.Remove(filepath.Join(verzeichnis, "MPlantin.ttf")); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	zweite, err := suche.Suche("MPlantin", false)
	if err != nil {
		t.Fatalf("die Schrift wurde nicht behalten: %v", err)
	}

	if erste != zweite {
		t.Error("es wurde ein zweites Mal gelesen")
	}
}

// **Ein Fehlschlag ist ein Befund, keine falsche Schrift** (§8).
func TestEineFehlendeSchriftWirdBenannt(t *testing.T) {
	suche := schriften.Neue([]string{t.TempDir()})

	_, err := suche.Suche("GibtsNicht", false)
	if !errors.Is(err, schriften.ErrNichtGefunden) {
		t.Errorf("Fehler %v", err)
	}

	// Die Meldung nennt die Orte — sonst rät jemand, wo gesucht wurde.
	if _, leer := suche.Suche("   ", false); !errors.Is(leer, schriften.ErrNichtGefunden) {
		t.Errorf("Fehler %v", leer)
	}
}

// Ein Ort, den es nicht gibt, hält die Suche nicht auf.
func TestEinFehlenderOrtHaeltDieSucheNichtAuf(t *testing.T) {
	suche := schriften.Neue([]string{
		"/gibt/es/nicht",
		verzeichnisMitSchrift(t, "MPlantin.ttf"),
	})

	if _, err := suche.Suche("MPlantin", false); err != nil {
		t.Errorf("suchen: %v", err)
	}
}

// Was keine Schriftdatei ist, wird nicht angeboten.
//
// `.ttc` steht hier **nicht** mehr: eine Sammlung wird gelesen, und auf macOS
// liegt der halbe Grundbestand so.
func TestNurSchriftendungenWerdenAngeboten(t *testing.T) {
	verzeichnis := t.TempDir()

	for _, name := range []string{"MPlantin.txt", "MPlantin.dfont", "unterverzeichnis"} {
		if strings.HasSuffix(name, "verzeichnis") {
			if err := os.Mkdir(filepath.Join(verzeichnis, name), 0o700); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			continue
		}

		if err := os.WriteFile(filepath.Join(verzeichnis, name),
			systemschrift(t), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	if _, err := schriften.Neue([]string{verzeichnis}).Suche("MPlantin", false); err == nil {
		t.Error("eine .txt- oder .dfont-Datei wurde als Schrift angeboten")
	}
}

// Eine Datei mit richtiger Endung und falschem Inhalt wird gemeldet.
func TestEineUnlesbareSchriftdateiWirdGemeldet(t *testing.T) {
	verzeichnis := t.TempDir()

	if err := os.WriteFile(filepath.Join(verzeichnis, "Kaputt.ttf"),
		[]byte("das ist keine Schrift"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := schriften.Neue([]string{verzeichnis}).Suche("Kaputt", false); err == nil {
		t.Error("eine unlesbare Datei wurde als Schrift angenommen")
	}
}

// **Die Orte der Plattform sind nicht leer.**
func TestDiePlattformNenntOrte(t *testing.T) {
	if len(schriften.Orte()) == 0 {
		t.Error("keine Orte — die Suche griffe ins Leere")
	}
}

// **Eine Datei, die zwischen Finden und Lesen verschwindet**, wird gemeldet.
//
// Der Zweig sieht wie eine Formalie aus und ist keiner: zwischen der
// Verzeichnisliste und dem Lesen liegt Zeit, und in ihr räumt jemand auf.
func TestEineVerschwundeneDateiWirdGemeldet(t *testing.T) {
	verzeichnis := t.TempDir()
	pfad := filepath.Join(verzeichnis, "MPlantin.ttf")

	if err := os.WriteFile(pfad, systemschrift(t), 0o000); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Cleanup(func() {
		if fehler := os.Chmod(pfad, 0o600); fehler != nil {
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	// Ohne Leserechte findet die Verzeichnisliste die Datei, das Lesen scheitert.
	if _, err := schriften.Neue([]string{verzeichnis}).Suche("MPlantin", false); err == nil {
		t.Error("eine unlesbare Datei wurde gelesen")
	}
}

// **Eine Sammlung wird auf ihre erste Schrift gelesen.**
//
// Auf macOS liegt der halbe Grundbestand als `.ttc`; ohne das fände die Suche
// dort fast nichts.
func TestEineSammlungWirdGelesen(t *testing.T) {
	// **Helvetica und nicht Helvetica Neue**: `x/image` weist deren Sammlung
	// mit „unsupported cmap encodings" ab. Nachgesehen, nicht vermutet — und
	// deshalb steht die Grenze auch im Vermerk 0008.
	const pfad = "/System/Library/Fonts/Helvetica.ttc"

	roh, err := os.ReadFile(pfad) //nolint:gosec // fester Pfad, kein Zulauf von außen
	if err != nil {
		t.Skipf("keine Sammlung unter %s: %v", pfad, err)
	}

	verzeichnis := t.TempDir()
	ziel := filepath.Join(verzeichnis, "Helvetica.ttc")

	//nolint:gosec // das Ziel liegt in t.TempDir(), nicht in einer Eingabe
	if fehler := os.WriteFile(ziel, roh, 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	schrift, err := schriften.Neue([]string{verzeichnis}).Suche("Helvetica", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if schrift.Breite('A') <= 0 {
		t.Errorf("Breite(A) %g", schrift.Breite('A'))
	}
}

// **Nebenläufig gefragt darf sie nicht abstürzen.**
//
// Der erste Anlauf hatte den Zwischenspeicher ohne Schloss. Kein Test sah es —
// sie fragen nacheinander —, und die **laufende Anwendung** stürzte mit
// „fatal error: concurrent map writes" ab, als zwei Vorschauen gleichzeitig
// dieselbe Schrift wollten.
//
// Mit `-race` ist dieser Test der Wächter darüber.
func TestDieSucheHaeltNebenlaeufigenFragenStand(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "MPlantin.ttf")})

	var warten sync.WaitGroup

	for range 16 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			schrift, err := suche.Suche("MPlantin", false)
			if err != nil {
				t.Errorf("suchen: %v", err)

				return
			}

			// Und die Schrift selbst wird ebenso nebenläufig benutzt: Breite
			// und Maske greifen beide in denselben Puffer von `sfnt`.
			if schrift.Breite('A') <= 0 {
				t.Errorf("Breite %g", schrift.Breite('A'))
			}
		}()
	}

	warten.Wait()
}
