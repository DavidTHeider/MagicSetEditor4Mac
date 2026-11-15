package app

// **Lesen, Prüfen, Räumen und Veröffentlichen unter einer Sperre** — Befund 9
// des Nachlaufs über Phase 1 von 0.9.0 (F04, #2343).
//
// Bis dahin gab der Prüfgriff frei, bevor Marke und Sperrdatei einer alten
// Sitzung gingen. Startete dazwischen ein Fenster mit derselben — inzwischen
// wiederverwendeten — Prozessnummer, räumte der Prüfende dessen **neue** Marke
// weg; und ein Ende, das zwischen Lesen und Prüfen eines anderen Starts fiel,
// sah dort wie ein Abbruch aus. Jetzt hält jeder Beginn und jedes Ende den
// Ordner, und wer ihn hält, hält beide an.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// ordnerHaltenImTest nimmt die Sperre über den Ordner, wie ein anderer Start
// mitten im Prüfen sie hält.
func ordnerHaltenImTest(t *testing.T, ordner string) *dateisperre.Griff {
	t.Helper()

	griff, err := dateisperre.Oeffnen(ordnersperrpfad(ordner))
	if err != nil {
		t.Fatal(err)
	}

	// **Mit Frist**: dieselbe Datei kann ein Kind eines parallelen Tests kurz
	// mithalten ([TestEineKurzGeerbteSperreIstKeinLeben]).
	if err := griff.AlleinBis(t.Context(), 2*time.Second, 5*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	return griff
}

// nochNicht sagt, ob `fertig` in einer kurzen Weile ausbleibt.
func nochNicht[T any](fertig <-chan T) bool {
	select {
	case <-fertig:
		return false
	case <-time.After(150 * time.Millisecond):
		return true
	}
}

func TestEinBeginnWartetAufDenOrdner(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	griff := ordnerHaltenImTest(t, ordner)

	type ergebnis struct {
		beginn sitzungsbeginn
		err    error
	}

	fertig := make(chan ergebnis, 1)

	go func() {
		beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
		fertig <- ergebnis{beginn, err}
	}()

	if !nochNicht(fertig) {
		t.Fatal("der Beginn wartete nicht auf den Ordner")
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-815.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("die Marke lag schon, bevor der Ordner frei war: %v", err)
	}

	_ = griff.Schliessen() //nolint:errcheck // frei

	beendet := <-fertig
	if beendet.err != nil {
		t.Fatal(beendet.err)
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-815.json")); err != nil {
		t.Fatalf("danach liegt die Marke nicht: %v", err)
	}

	_ = beendet.beginn.ende() //nolint:errcheck // Aufräumen
}

func TestEinEndeWartetAufDenOrdner(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	griff := ordnerHaltenImTest(t, ordner)
	fertig := make(chan error, 1)

	go func() { fertig <- beginn.ende() }()

	if !nochNicht(fertig) {
		t.Fatal("das Ende wartete nicht auf den Ordner")
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-815.json")); err != nil {
		t.Fatalf("die Marke ging, bevor der Ordner frei war: %v", err)
	}

	_ = griff.Schliessen() //nolint:errcheck // frei

	if err := <-fertig; err != nil {
		t.Fatal(err)
	}

	if rest, _ := filepath.Glob(filepath.Join(ordner, "fenster-*")); len(rest) != 0 { //nolint:errcheck // ein festes Muster
		t.Errorf("nach dem Ende liegt noch %v", rest)
	}
}

func TestOhneDenOrdnerBeginntKeineSitzung(t *testing.T) {
	t.Parallel()

	t.Run("gehalten über die Frist", func(t *testing.T) {
		t.Parallel()

		ordner := t.TempDir()
		ordnerHaltenImTest(t, ordner)

		if _, err := sitzungBeginnenMit(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe, 50*time.Millisecond); err == nil {
			t.Fatal("ein Beginn ohne den Ordner wurde angenommen")
		}

		if rest, _ := filepath.Glob(filepath.Join(ordner, "*")); len(rest) != 1 { //nolint:errcheck // ein festes Muster
			t.Errorf("es liegt mehr als die Sperre des Ordners: %v", rest)
		}
	})

	t.Run("die Sperre lässt sich nicht öffnen", func(t *testing.T) {
		t.Parallel()

		ordner := t.TempDir()
		if err := os.Mkdir(ordnersperrpfad(ordner), 0o700); err != nil {
			t.Fatal(err)
		}

		if _, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe); err == nil {
			t.Fatal("ein Beginn ohne den Ordner wurde angenommen")
		}
	})
}

// **Ein Ende ohne den Ordner räumt trotzdem** — ein Schluss bleibt ein
// Schluss; was dann fehlt, ist nur der Schutz vor einem gleichzeitigen
// Prüfen, und der Fehler sagt es.
func TestEinEndeOhneDenOrdnerRaeumtTrotzdem(t *testing.T) {
	t.Parallel()

	for name, sperren := range map[string]func(*testing.T, string){
		"gehalten über die Frist": func(t *testing.T, ordner string) { t.Helper(); ordnerHaltenImTest(t, ordner) },
		"die Sperre lässt sich nicht öffnen": func(t *testing.T, ordner string) {
			t.Helper()

			if err := os.Remove(ordnersperrpfad(ordner)); err != nil {
				t.Fatal(err)
			}

			if err := os.Mkdir(ordnersperrpfad(ordner), 0o700); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ordner := t.TempDir()

			beginn, err := sitzungBeginnenMit(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe, 50*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}

			sperren(t, ordner)

			if err := beginn.ende(); err == nil {
				t.Error("das Ende verschwieg, dass der Ordner fehlte")
			}

			if rest, _ := filepath.Glob(filepath.Join(ordner, "fenster-*")); len(rest) != 0 { //nolint:errcheck // ein festes Muster
				t.Errorf("nach dem Ende liegt noch %v", rest)
			}
		})
	}
}

// **Eine kurz geerbte Sperre ist kein Leben** — gefunden, als dieser Test im
// vollen Paketlauf rot wurde. Ein `flock` hängt an der offenen
// Dateibeschreibung, und ein Kind, das ein anderer Teil des Programms gerade
// startet, hält sie zwischen `fork` und `exec` mit — gemessen am 2026-10-14
// in 672 von 20.000 Versuchen, bei fortwährend startenden Kindern. Eine
// einmalige Probe hielt die tote Sitzung dann für lebend, und ihr Abbruch
// blieb ungemeldet.
func TestEineKurzGeerbteSperreIstKeinLeben(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	legeMarkeMitAnspruch(t, ordner, 4712)

	erbe, err := dateisperre.Oeffnen(anspruchspfad(ordner, 4712))
	if err != nil {
		t.Fatal(err)
	}

	if err := erbe.Allein(); err != nil {
		t.Fatal(err)
	}

	// **Das Kind ruft `exec` nach einer kurzen Weile**, und die Sperre geht.
	go func() {
		time.Sleep(20 * time.Millisecond)

		_ = erbe.Schliessen() //nolint:errcheck // das Erbe endet
	}()

	marke := sitzungsmarke{PID: 4712, Gesperrt: true}
	if sitzungLebt(ordner, marke, 815, immerLebend) {
		t.Error("eine Sperre, die nur ein Kind kurz hielt, galt als lebende Sitzung")
	}
}
