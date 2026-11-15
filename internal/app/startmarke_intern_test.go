package app

// **Eine Sitzung, die nicht sauber endete, ist beim nächsten Start zu
// sehen** — F04 (#2343),
// Befund A-04.
//
// # Die Marke, nicht ihr Fehlen
//
// Eine Sitzung legt beim Start ihre Marke und nimmt sie **nur** beim
// sauberen Schluss weg. **Eine liegengebliebene Marke ist ein Abbruch**,
// eine fehlende ist sauber — umgekehrt wäre jeder erste Start ein Abbruch.
// Die Marke trägt die Prozesskennung: lebt der Prozess noch, ist es ein
// zweites Fenster und kein Abbruch.
//
// # Was die Prüfungen nicht halten
//
// Einen Prozess, der nach einem Abbruch dieselbe Kennung bekam — dann sieht
// die Marke aus wie lebendig, und der Abbruch fällt aus. Das kostet einen
// Hinweis, nicht einen falschen.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// jetzt ist die gestellte Uhr dieser Läufe.
var jetzt = time.Date(2026, 10, 14, 9, 30, 0, 0, time.UTC)

// legeMarke legt eine Marke, wie eine laufende Sitzung sie hinterlässt.
func legeMarke(t *testing.T, ordner string, pid int, spur string) {
	t.Helper()

	roh, err := json.Marshal(sitzungsmarke{PID: pid, Gestartet: jetzt.Add(-time.Hour), Fassung: "0.9.0", Spur: spur})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ordner, "fenster-"+strconv.Itoa(pid)+".json"), roh, 0o600); err != nil {
		t.Fatal(err)
	}
}

// ohneAusgabe ersetzt `debug.SetCrashOutput` — der Testprozess soll seine
// eigene Absturzausgabe behalten.
func ohneAusgabe(*os.File) error { return nil }

func TestEineLiegengebliebeneMarkeIstEinAbbruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarke(t, ordner, 4711, "")

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 || !beginn.abbrueche[0].Gestartet.Equal(jetzt.Add(-time.Hour)) {
		t.Fatalf("Abbrüche %+v, erwartet einer von vor einer Stunde", beginn.abbrueche)
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-4711.json")); !os.IsNotExist(err) {
		t.Error("die alte Marke liegt noch — beim nächsten Start wäre es derselbe Abbruch")
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-815.json")); err != nil {
		t.Errorf("die eigene Marke fehlt: %v", err)
	}
}

func TestEineLebendeMarkeIstKeinAbbruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarke(t, ordner, 4711, "")

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return true }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 0 {
		t.Errorf("ein zweites Fenster galt als Abbruch: %+v", beginn.abbrueche)
	}

	if _, err := os.Stat(filepath.Join(ordner, "fenster-4711.json")); err != nil {
		t.Errorf("die Marke des lebenden Fensters ist weg: %v", err)
	}
}

func TestEinAbbruchMitSpurSagtEs(t *testing.T) {
	ordner := t.TempDir()
	if err := os.WriteFile(filepath.Join(ordner, "absturz-1-4711.txt"), []byte("panic: Probe\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	legeMarke(t, ordner, 4711, "absturz-1-4711.txt")

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 || !beginn.abbrueche[0].MitSpur {
		t.Errorf("Abbrüche %+v, erwartet einer mit Spur", beginn.abbrueche)
	}
}

func TestDerSaubereSchlussRaeumtMarkeUndLeereSpur(t *testing.T) {
	ordner := t.TempDir()

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if schlussfehler := beginn.ende(); schlussfehler != nil {
		t.Fatal(schlussfehler)
	}

	// **Bis auf die Sperre des Ordners** — sie bleibt, denn eine Sperre auf
	// eine Datei, die gerade ein anderer löscht, sperrte nichts.
	eintraege, _ := os.ReadDir(ordner) //nolint:errcheck // leer heißt hier: nichts da
	if len(eintraege) != 1 || eintraege[0].Name() != filepath.Base(ordnersperrpfad(ordner)) {
		t.Errorf("nach dem sauberen Schluss liegt noch %d Datei(en) da: %v", len(eintraege), eintraege)
	}

	// **Und der nächste Start sieht keinen Abbruch** — die Positivkontrolle.
	zweiter, err := sitzungBeginnen(ordner, 816, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(zweiter.abbrueche) != 0 {
		t.Errorf("nach einem sauberen Schluss galt der nächste Start als Abbruch: %+v", zweiter.abbrueche)
	}
}

func TestDieSpurenRotierenMitFesterGrenze(t *testing.T) {
	ordner := t.TempDir()

	for i := range absturzspurenHoechstens + 5 {
		name := filepath.Join(ordner, "absturz-"+strconv.Itoa(1000+i)+"-1.txt")
		if err := os.WriteFile(name, []byte("panic\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := os.Chtimes(name, jetzt, jetzt.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe); err != nil {
		t.Fatal(err)
	}

	spuren, _ := filepath.Glob(filepath.Join(ordner, "absturz-*.txt")) //nolint:errcheck // ein festes Muster
	if len(spuren) > absturzspurenHoechstens+1 {
		t.Errorf("%d Spuren liegen da, erlaubt sind %d alte und die eigene", len(spuren), absturzspurenHoechstens)
	}

	if _, err := os.Stat(filepath.Join(ordner, "absturz-1000-1.txt")); !os.IsNotExist(err) {
		t.Error("die älteste Spur liegt noch — weg muss die älteste, nicht irgendeine")
	}
}
