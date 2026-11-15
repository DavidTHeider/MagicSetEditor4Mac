package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Paket, das es nicht gibt, ist „nicht gefunden"**
// (#997).
//
// # Woran das aufgefallen ist
//
// Am Abnahmelauf über die vierte Fläche. Kommandozeile und MCP zeigen den
// rohen Fehler des Betriebssystems und nennen dabei den Pfad; der Dienst darf
// das nicht (§8) und gab deshalb `{"code":"unbekannt","meldung":"der Aufruf
// ist gescheitert"}` — ein 500 ohne Grund für etwas, das der Aufrufer sofort
// beheben könnte.
//
// **Der Fehler lag nicht am Dienst.** Er lag daran, dass eine fehlende Datei
// im Fachkern **gar nicht eingeordnet** war: kein `ErrNotFound`, kein
// `ErrValidation`, nur ein `*fs.PathError`. Die drei anderen Flächen haben das
// jahrelang verdeckt, weil sie den rohen Text zeigen.
//
// # Warum die Einordnung hierher gehört
//
// `internal/msepaket` kennt `domain` (§2) und ist die Stelle, die den
// Unterschied kennt: es hat gerade versucht, die Datei zu öffnen.
func TestEinFehlendesPaketIstNichtGefunden(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "gibtsnicht.mse-set")

	_, fehler := msepaket.Oeffne(pfad)
	if fehler == nil {
		t.Fatal("ein Paket, das es nicht gibt, ging auf")
	}

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v ist kein %v", fehler, domain.ErrNotFound)
	}

	// **Der Pfad bleibt in der Meldung.** Wer sie sieht, soll wissen, welche
	// Datei fehlt — und die vier Flächen entscheiden je für sich, wie viel
	// davon sie zeigen.
	if !strings.Contains(fehler.Error(), pfad) {
		t.Errorf("die Meldung nennt den Pfad nicht: %v", fehler)
	}
}

// **Und ein Paket, das es gibt, bleibt, was es ist.**
//
// Ohne diesen Fall wäre auch eine Einordnung grün, die **jeden** Fehler zu
// „nicht gefunden" machte.
func TestEinKaputtesPaketIstNichtNichtGefunden(t *testing.T) {
	t.Parallel()

	// Eine Datei, die es gibt und die kein Archiv ist.
	pfad := filepath.Join(t.TempDir(), "kaputt.mse-set")
	if err := os.WriteFile(pfad, []byte("das ist kein Zip"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, fehler := msepaket.Oeffne(pfad)
	if fehler == nil {
		t.Fatal("eine kaputte Datei ging auf")
	}

	if errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("eine kaputte Datei gilt als nicht gefunden: %v", fehler)
	}
}

// **Und ein Paket, das dasteht und nicht gelesen werden darf, ist etwas
// anderes als eines, das fehlt.**
//
// # Warum dieser Fall gebaut werden muss
//
// Er ist der einzige, der den **anderen** Zweig betritt: `os.Stat` scheitert,
// aber nicht mit `ErrNotExist`. Ohne ihn stünde dort eine Zeile, die kein Test
// je durchläuft — und §3 nimmt eine unbequeme Verzweigung ausdrücklich nicht
// als Ausnahme an.
//
// **Er wird als `root` übersprungen.** Dort greift keine Rechteprüfung, und
// ein Test, der dann grün ist, sagt nichts.
func TestEinUnlesbaresPaketIstNichtNichtGefunden(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("als root greift keine Rechteprüfung")
	}

	verschlossen := filepath.Join(t.TempDir(), "verschlossen")
	if err := os.Mkdir(verschlossen, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	pfad := filepath.Join(verschlossen, "paket.mse-set")
	if err := os.WriteFile(pfad, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// Ohne Ausführrecht lässt sich der Inhalt des Verzeichnisses nicht statten.
	if err := os.Chmod(verschlossen, 0o600); err != nil {
		t.Fatalf("Rechte setzen: %v", err)
	}

	t.Cleanup(func() {
		// **Zurückgeben, sonst kann `t.TempDir()` nicht aufräumen.**
		// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht, sonst räumt t.TempDir() nicht auf
		if err := os.Chmod(verschlossen, 0o700); err != nil {
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	_, fehler := msepaket.Oeffne(pfad)
	if fehler == nil {
		t.Fatal("ein unlesbares Paket ging auf")
	}

	if errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein unlesbares Paket gilt als nicht gefunden: %v", fehler)
	}
}

// **Ein Paketverzeichnis, das sich nicht öffnen lässt, ist ebenfalls nicht
// „nicht gefunden"** — FMT-06.AK-01 öffnet es als `os.Root`, und das scheitert
// erst **nach** dem `os.Stat`: das Verzeichnis ist da, nur ohne Leserecht.
func TestEinVerschlossenesPaketverzeichnisGehtNichtAuf(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("als root greift keine Rechteprüfung")
	}

	paket := filepath.Join(t.TempDir(), "zu.mse-set")
	if err := os.Mkdir(paket, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.Chmod(paket, 0o000); err != nil {
		t.Fatalf("Rechte nehmen: %v", err)
	}

	t.Cleanup(func() {
		// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht, sonst räumt t.TempDir() nicht auf
		if err := os.Chmod(paket, 0o700); err != nil {
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	_, fehler := msepaket.Oeffne(paket)
	if fehler == nil {
		t.Fatal("ein verschlossenes Paketverzeichnis ging auf")
	}

	if errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein verschlossenes Paketverzeichnis gilt als nicht gefunden: %v", fehler)
	}
}
