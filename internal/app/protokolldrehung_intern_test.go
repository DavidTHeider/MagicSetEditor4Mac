package app

// **Gedreht wird auch während der Sitzung** — PERF03
// (#2353),
// Befund A-20 des Audits 0.9.0: die Grenze galt nur beim Einrichten, und ein
// Dienst, der eine Woche lief, schrieb eine Datei ohne Ende.
//
// **Zwei Weichen auf denselben Ort** stehen für zwei Prozesse: Fenster und
// Dienst schreiben in dieselbe `mse.log`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kleineRegel ist eine Regel mit Datei, kleiner Grenze und zwei Vorfassungen.
func kleineRegel() domain.Protokollregel {
	regel := domain.ProtokollregelAus(nil)
	regel.Datei, regel.System, regel.Ausgabe = true, false, false
	regel.GrenzeBytes, regel.Aufbewahrung = 200, 2

	return regel
}

// protokollgroessen gibt `mse.log` und ihre Vorfassungen samt Größen.
func protokollgroessen(t *testing.T, ordner string) map[string]int64 {
	t.Helper()

	eintraege, err := os.ReadDir(ordner)
	if err != nil {
		t.Fatal(err)
	}

	groessen := map[string]int64{}

	for _, eintrag := range eintraege {
		// **Nur die Datei und ihre Vorfassungen** — die Sperren daneben sind
		// leer und keine Protokolle.
		nummer, ist := strings.CutPrefix(eintrag.Name(), Protokollname+".")
		if eintrag.Name() != Protokollname && (!ist || strings.Trim(nummer, "0123456789") != "") {
			continue
		}

		angabe, err := eintrag.Info()
		if err != nil {
			t.Fatal(err)
		}

		groessen[eintrag.Name()] = angabe.Size()
	}

	return groessen
}

func TestEinProtokollDrehtWaehrendDerSitzung(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	zeile := []byte(strings.Repeat("z", 49) + "\n")

	weiche := neueWeicheMit(ordner, nil, nil, os.UserHomeDir)
	weiche.Einrichten(kleineRegel())

	t.Cleanup(func() { _ = weiche.Schliessen() }) //nolint:errcheck // Aufräumen

	datei := weiche.kanaele[0].ziel

	for range 40 {
		if _, err := datei.Write(zeile); err != nil {
			t.Fatal(err)
		}
	}

	dateien := protokollgroessen(t, ordner)

	// **Höchstens Aufbewahrung und eine** — `mse.log`, `.1`, `.2`.
	if len(dateien) > 3 {
		t.Errorf("%d Dateien: %v", len(dateien), dateien)
	}

	for name, groesse := range dateien {
		if groesse > 200+int64(len(zeile)) {
			t.Errorf("%s wiegt %d Byte, Grenze 200 und eine Zeile", name, groesse)
		}
	}
}

// **Ein Erbe hält das Drehen nicht auf** — #2401: hält ein Kind zwischen
// `fork` und `exec` die Drehsperre mit, fiel jedes Drehen in dieser Zeit aus,
// und die Datei wuchs über ihre Grenze. Nachgestellt mit einem Griff, der
// die Sperre 20 ms hält.
func TestEinKurzerHalterHaeltDasDrehenNichtAuf(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	zeile := []byte(strings.Repeat("z", 49) + "\n")

	weiche := neueWeicheMit(ordner, nil, nil, os.UserHomeDir)
	weiche.Einrichten(kleineRegel())

	t.Cleanup(func() { _ = weiche.Schliessen() }) //nolint:errcheck // Aufräumen

	erbe, err := dateisperre.Oeffnen(filepath.Join(ordner, Drehsperre))
	if err != nil {
		t.Fatal(err)
	}

	if err := erbe.Allein(); err != nil {
		t.Fatal(err)
	}

	uhr := time.AfterFunc(20*time.Millisecond, func() { _ = erbe.Schliessen() }) //nolint:errcheck // losgelassen ist losgelassen
	t.Cleanup(func() { uhr.Stop() })

	for range 40 {
		if _, err := weiche.kanaele[0].ziel.Write(zeile); err != nil {
			t.Fatal(err)
		}
	}

	for name, groesse := range protokollgroessen(t, ordner) {
		if groesse > 200+int64(len(zeile)) {
			t.Errorf("%s wiegt %d Byte, Grenze 200 und eine Zeile", name, groesse)
		}
	}
}

// **Ein zweiter Schreiber folgt der Drehung** — er schrieb sonst weiter in die
// beiseitegelegte Datei, und die neue bliebe ihm fremd.
func TestEinZweiterSchreiberFolgtDerDrehung(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	zeile := []byte(strings.Repeat("a", 99) + "\n")

	fenster := neueWeicheMit(ordner, nil, nil, os.UserHomeDir)
	fenster.Einrichten(kleineRegel())

	dienst := neueWeicheMit(ordner, nil, nil, os.UserHomeDir)
	dienst.Einrichten(kleineRegel())

	t.Cleanup(func() { _ = fenster.Schliessen(); _ = dienst.Schliessen() }) //nolint:errcheck // Aufräumen

	for range 3 {
		if _, err := fenster.kanaele[0].ziel.Write(zeile); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := dienst.kanaele[0].ziel.Write([]byte("vom Dienst\n")); err != nil {
		t.Fatal(err)
	}

	jetzt, err := os.ReadFile(filepath.Join(ordner, Protokollname)) //nolint:gosec // Pfad aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(jetzt), "vom Dienst") {
		t.Errorf("die Zeile des zweiten Schreibers fehlt in der neuen Datei: %q", jetzt)
	}

	// **Und eine gelöschte Datei entsteht neu.**
	if err := os.Remove(filepath.Join(ordner, Protokollname)); err != nil {
		t.Fatal(err)
	}

	if _, err := dienst.kanaele[0].ziel.Write([]byte("nach dem Löschen\n")); err != nil {
		t.Fatal(err)
	}

	if neu, err := os.ReadFile(filepath.Join(ordner, Protokollname)); err != nil || !strings.Contains(string(neu), "nach dem Löschen") { //nolint:gosec // Pfad aus t.TempDir()
		t.Errorf("nach dem Löschen: %q, %v", neu, err)
	}
}

// ueberDerGrenze gibt die Datei einer Weiche, die schon über der Grenze liegt.
func ueberDerGrenze(t *testing.T, ordner string) *protokolldatei {
	t.Helper()

	weiche := neueWeicheMit(ordner, nil, nil, os.UserHomeDir)
	weiche.Einrichten(kleineRegel())

	t.Cleanup(func() { _ = weiche.Schliessen() }) //nolint:errcheck // Aufräumen

	datei, ist := weiche.kanaele[0].ziel.(*protokolldatei)
	if !ist {
		t.Fatalf("kein Dateikanal: %T", weiche.kanaele[0].ziel)
	}

	if _, err := datei.File.WriteString(strings.Repeat("x", 250) + "\n"); err != nil {
		t.Fatal(err)
	}

	return datei
}

// **Jeder Fehlschlag heißt „weiter wie bisher"** — ein Protokoll, das beim
// Drehen scheitert, schreibt weiter, wohin es kann.
func TestEinScheiterndesDrehenSchreibtWeiter(t *testing.T) {
	t.Parallel()

	t.Run("die Drehsperre ist kein Ort", func(t *testing.T) {
		t.Parallel()

		ordner := t.TempDir()
		if err := os.Mkdir(filepath.Join(ordner, Drehsperre), 0o700); err != nil {
			t.Fatal(err)
		}

		datei := ueberDerGrenze(t, ordner)

		if _, err := datei.WriteString("weiter\n"); err != nil {
			t.Fatal(err)
		}

		if dateien := protokollgroessen(t, ordner); len(dateien) != 1 {
			t.Errorf("gedreht ohne Drehsperre: %v", dateien)
		}
	})

	t.Run("ein anderer hält die Drehsperre", func(t *testing.T) {
		t.Parallel()

		ordner := t.TempDir()
		datei := ueberDerGrenze(t, ordner)

		halter, err := dateisperre.Oeffnen(filepath.Join(ordner, Drehsperre))
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = halter.Schliessen() }) //nolint:errcheck // Aufräumen

		if err := halter.Allein(); err != nil {
			t.Fatal(err)
		}

		if _, err := datei.WriteString("weiter\n"); err != nil {
			t.Fatal(err)
		}

		if dateien := protokollgroessen(t, ordner); len(dateien) != 1 {
			t.Errorf("gedreht, obwohl ein anderer dreht: %v", dateien)
		}
	})

	t.Run("der Ordner ist fort", func(t *testing.T) {
		t.Parallel()

		ordner := filepath.Join(t.TempDir(), "protokoll")
		if err := os.Mkdir(ordner, 0o700); err != nil {
			t.Fatal(err)
		}

		datei := ueberDerGrenze(t, ordner)

		if err := os.RemoveAll(ordner); err != nil {
			t.Fatal(err)
		}

		if _, err := datei.WriteString("in die alte\n"); err != nil {
			t.Errorf("ohne Ordner: %v", err)
		}
	})

	t.Run("die Datei ist zu", func(t *testing.T) {
		t.Parallel()

		datei := ueberDerGrenze(t, t.TempDir())

		if err := datei.File.Close(); err != nil {
			t.Fatal(err)
		}

		if _, err := datei.WriteString("zu\n"); err == nil {
			t.Error("in eine geschlossene Datei geschrieben")
		}
	})
}

// **Ein anderer war schneller** — er hat gedreht, nachdem dieser Schreiber
// die Grenze sah und bevor er die Drehsperre bekam. Dann wird nicht noch
// einmal gedreht, sondern gefolgt. **Nicht parallel**: die Naht ist eine
// Paketvariable.
func TestEinSchnellererDreherWirdNichtUeberholt(t *testing.T) { //nolint:paralleltest // siehe darüber
	ordner := t.TempDir()
	datei := ueberDerGrenze(t, ordner)
	pfad := filepath.Join(ordner, Protokollname)

	vorher := vorDemDrehen
	vorDemDrehen = func() {
		vorfassungenRuecken(pfad, 2)

		if err := os.WriteFile(pfad, []byte("vom anderen\n"), 0o600); err != nil {
			t.Error(err)
		}
	}

	t.Cleanup(func() { vorDemDrehen = vorher })

	if _, err := datei.WriteString("gefolgt\n"); err != nil {
		t.Fatal(err)
	}

	jetzt, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus t.TempDir()
	if err != nil || string(jetzt) != "vom anderen\ngefolgt\n" {
		t.Errorf("am Ort: %q, %v", jetzt, err)
	}

	if dateien := protokollgroessen(t, ordner); len(dateien) != 2 {
		t.Errorf("noch einmal gedreht: %v", dateien)
	}
}
