package app

// **Die Ansage eines Dienstes unter launchd** — SEC02
// (#2355),
// Befund A-16: launchd hängte jede Ansage samt Marke an `dienst.txt` an, und
// die Fehlerausgabe ging in dieselbe Datei, ohne Grenze.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func markeNummer(nummer int) string {
	return strings.Repeat(string(rune('a'+nummer)), 64)
}

// **Zehn Starts, eine Marke** — die Ansage ersetzt die vorige, mit 0600,
// im Ordner mit 0700, und daneben bleibt nichts liegen.
func TestDieAnsageErsetztDieVorige(t *testing.T) {
	t.Parallel()

	ordner := filepath.Join(t.TempDir(), "schnittstellen")
	datei := filepath.Join(ordner, "dienst.txt")

	for nummer := range 10 {
		ansage := Dienstansage{Adresse: "127.0.0.1:4711", Marke: markeNummer(nummer)}
		if err := AnsageInDatei(datei, ansage); err != nil {
			t.Fatalf("Start %d: %v", nummer, err)
		}
	}

	roh, err := os.ReadFile(datei) //nolint:gosec // G304: der eigene Testordner
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(roh), "Marke ") != 1 || !strings.Contains(string(roh), markeNummer(9)) {
		t.Errorf("Ansage nach zehn Starts:\n%s", roh)
	}

	for pfad, erwartet := range map[string]os.FileMode{datei: 0o600, ordner: 0o700} {
		if angaben, fehler := os.Stat(pfad); fehler != nil || angaben.Mode().Perm() != erwartet {
			t.Errorf("%s: %v, erwartet %v (%v)", pfad, angaben.Mode().Perm(), erwartet, fehler)
		}
	}

	if eintraege, _ := os.ReadDir(ordner); len(eintraege) != 1 { //nolint:errcheck // gezählt
		t.Errorf("neben der Ansage liegt etwas: %v", eintraege)
	}
}

// **Ohne Ordner keine Ansage** — und keine halbe.
func TestOhneOrdnerKeineAnsage(t *testing.T) {
	t.Parallel()

	blockiert := filepath.Join(t.TempDir(), "eine Datei")
	if err := os.WriteFile(blockiert, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AnsageInDatei(filepath.Join(blockiert, "dienst.txt"), Dienstansage{Marke: markeNummer(0)}); err == nil {
		t.Error("eine Ansage unter einer Datei")
	}

	ziel := filepath.Join(t.TempDir(), "dienst.txt")
	if err := os.Mkdir(ziel, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := AnsageInDatei(ziel, Dienstansage{Marke: markeNummer(0)}); err == nil {
		t.Error("eine Ansage über einen Ordner")
	}

	if eintraege, _ := os.ReadDir(filepath.Dir(ziel)); len(eintraege) != 1 { //nolint:errcheck // gezählt
		t.Errorf("eine halbe Ansage blieb liegen: %v", eintraege)
	}

	// **Ein Ordner ohne Schreibrecht** nimmt keine Nachbardatei an.
	gesperrt := filepath.Join(t.TempDir(), "gesperrt")
	if err := os.Mkdir(gesperrt, 0o500); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // nur zum Aufräumen

	if err := AnsageInDatei(filepath.Join(gesperrt, "dienst.txt"), Dienstansage{Marke: markeNummer(0)}); err == nil {
		t.Error("eine Ansage in einem Ordner ohne Schreibrecht")
	}
}

// **Die Fehlerausgabe hat eine Grenze** — über ihr fängt sie neu an und sagt
// es; darunter bleibt sie, wie sie ist.
func TestDieFehlerausgabeWirdBegrenzt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "dienst.log")
	if err := os.WriteFile(pfad, bytes.Repeat([]byte("alte Zeile\n"), 200), 0o600); err != nil {
		t.Fatal(err)
	}

	datei, err := os.OpenFile(pfad, os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // G304: der eigene Testordner
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = datei.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := FehlerausgabeBegrenzen(datei, 4096); fehler != nil {
		t.Fatal(fehler)
	}

	unveraendert, _ := os.ReadFile(pfad) //nolint:errcheck,gosec // gelesen wird, was eben geschrieben wurde
	if len(unveraendert) != 2200 {
		t.Errorf("unter der Grenze gekürzt: %d Bytes", len(unveraendert))
	}

	if fehler := FehlerausgabeBegrenzen(datei, 1024); fehler != nil {
		t.Fatal(fehler)
	}

	gekuerzt, _ := os.ReadFile(pfad) //nolint:errcheck,gosec // dito
	if len(gekuerzt) > 200 || strings.Contains(string(gekuerzt), "alte Zeile") || !strings.Contains(string(gekuerzt), "gekürzt") {
		t.Errorf("über der Grenze:\n%s", gekuerzt)
	}

	// **Eine Fehlerausgabe, die sich nicht kürzen lässt**, sagt es.
	nurLesend, err := os.Open(pfad) //nolint:gosec // G304: der eigene Testordner
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = nurLesend.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := FehlerausgabeBegrenzen(nurLesend, 1); fehler == nil {
		t.Error("eine nur lesbare Fehlerausgabe galt als gekürzt")
	}

	// **Eine Fehlerausgabe, die keine Datei ist** — ein Rohr, ein Endgerät —,
	// bleibt, wie sie ist.
	leser, schreiber, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _, _ = leser.Close(), schreiber.Close() }) //nolint:errcheck // nur zum Aufräumen

	if err := FehlerausgabeBegrenzen(schreiber, 1); err != nil {
		t.Errorf("ein Rohr: %v", err)
	}
}
