// **Ein zweiter Prozess hält den Tausch auf** — Befund Z6-02 von Phase 3.2
// (#1163,
// #1174).
//
// # Warum dieser Lauf hier steht und nicht in `internal/storage`
//
// Dort ist die Sperre für sich geprüft: zwei Prozesse, ein Anspruch, und der
// zweite kommt nicht herein. Was das **nicht** sagt, ist, ob der
// Rückspielweg sie überhaupt nimmt.
//
// Genau das war die Lücke: eine Gegenprobe, die `alleinnaht` den Anspruch
// weglassen ließ, machte **keinen** Lauf rot — die Läufe im Fenster setzen
// ihre eigene Naht. Ein Wächter, den man entfernen kann, ohne dass etwas rot
// wird, ist keiner (§10).
package app_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// halteschalter sagt dem zweiten Prozess, welchen Bestand er halten soll.
const halteschalter = "MSE_HALTE_BESTAND"

// TestMain lässt den Testlauf auch der zweite Prozess sein.
//
// **Der Testlauf selbst, noch einmal gestartet.** `go test` baut ein Programm;
// mit einem Umgebungsschalter verhält es sich als Helfer. Das ist der übliche
// Weg für Mehrprozessläufe in Go und braucht kein zweites Ziel im Bau.
func TestMain(m *testing.M) {
	if pfad := os.Getenv(halteschalter); pfad != "" {
		halteDenBestand(pfad)

		return
	}

	// **Ein Zuhause für den ganzen Lauf** — warum, steht in
	// `echteszuhause_test.go`.
	aufraeumen, fehler := eigenesZuhauseFuerDenLauf()
	if fehler != nil {
		//nolint:forbidigo // in TestMain gibt es kein *testing.T
		fmt.Println(fehler)
		os.Exit(1)
	}

	if fehler := dasEchteZuhauseIstUnberuehrt(); fehler != nil {
		//nolint:forbidigo // dito
		fmt.Println(fehler)
		aufraeumen()
		os.Exit(1)
	}

	code := m.Run()

	raeumeHelfer()
	aufraeumen()

	os.Exit(code)
}

// halteDenBestand **öffnet** den Bestand und wartet auf ein Zeichen.
//
// # Geöffnet und nicht gesperrt
//
// Das ist der wirkliche Fall: `mse-mcp` steht offen daneben und tut nichts.
// Gesehen wird er von [storage.AlleinAmBestand] — eine offene Verbindung hält
// Sperren auf `-wal` und `-shm`, und die exklusive Probe scheitert daran.
//
// **Die Nachbardatei deckt den anderen Fall**, den die Probe nicht sehen kann:
// einen Prozess, der genau **während** des Tauschs öffnet. Dafür gibt es den
// Mehrprozesslauf in `internal/storage`; er gehört dorthin, wo die Zusage
// eingelöst wird.
func halteDenBestand(pfad string) {
	db, err := storage.Oeffnen(context.Background(), pfad)
	if err != nil {
		os.Exit(2)
	}

	// **Erst melden, dann warten.** Der Aufrufer darf nicht weitermachen,
	// bevor der Bestand wirklich offen ist — sonst prüfte er ein Fenster, das
	// noch niemand besetzt hat (§10).
	//
	// **Ohne `defer` auf das Schließen**: hier steht `os.Exit`, und danach
	// läuft keiner mehr. Der Prozess endet ohnehin, und mit ihm der Griff.
	if _, schreibfehler := os.Stdout.WriteString("gefasst\n"); schreibfehler != nil {
		os.Exit(3)
	}

	_, _ = os.Stdin.Read(make([]byte, 1)) //nolint:errcheck // das Zeichen zum Loslassen

	_ = db.Close() //nolint:errcheck // gleich ist der Prozess fort
}

// **Während ein zweiter Prozess den Bestand hält, wird nicht zurückgespielt.**
func TestEinZweiterProzessHaeltDasZurueckspielenAuf(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	anwendung := laufendeAnwendung(t, bestand)

	ziel := gesichertMitSprache(t, anwendung, filepath.Join(ort, "gut.msesicherung"), "de")

	loslassen := zweiterHalter(t, bestand)
	defer loslassen()

	antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, "")
	if antwort.Fehler == nil {
		t.Fatal("es wurde getauscht, während ein zweiter Prozess den Bestand hält")
	}

	// **Und die Anwendung läuft weiter** — sie hat ihre Datenbank zurück.
	if antwort := anwendung.Einstellungen(t.Context()); antwort.Fehler != nil {
		t.Fatalf("die Anwendung hat keine Datenbank mehr: %+v", antwort.Fehler)
	}

	// **Die Gegenprobe: nach dem Loslassen geht es.** Ohne sie sagte der Lauf
	// nur, dass etwas fehlschlägt — nicht, dass der zweite Prozess der Grund
	// ist (§10).
	loslassen()

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, ""); antwort.Fehler != nil {
		t.Fatalf("nach dem Loslassen: %+v", antwort.Fehler)
	}
}

// zweiterHalter startet den Helfer und wartet, bis er die Sperre hält.
func zweiterHalter(t *testing.T, pfad string) (loslassen func()) {
	t.Helper()

	// #nosec G204,G702 -- der Testlauf selbst, aus einer eigenen Kopie
	befehl := exec.CommandContext(probekontext(t), helferbinaer(t))
	befehl.Env = append(os.Environ(), halteschalter+"="+pfad)

	hinein, err := befehl.StdinPipe()
	if err != nil {
		t.Fatalf("Eingabe: %v", err)
	}

	heraus, err := befehl.StdoutPipe()
	if err != nil {
		t.Fatalf("Ausgabe: %v", err)
	}

	if fehler := befehl.Start(); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	if _, fehler := heraus.Read(make([]byte, len("gefasst\n"))); fehler != nil {
		t.Fatalf("der Helfer meldete sich nicht: %v", fehler)
	}

	var schon bool

	return func() {
		if schon {
			return
		}

		schon = true

		_, _ = hinein.Write([]byte("x")) //nolint:errcheck // der Helfer endet ohnehin
		_ = befehl.Wait()                //nolint:errcheck // dito
	}
}

// **Ein Zurückspielen wartet nicht auf sich selbst.**
//
// # Der Befund, den nur die Laufzeit gezeigt hat
//
// Der Alleinanspruch wurde im `defer` zurückgegeben — also **nach** dem
// Wiederöffnen. `flock` gilt für die offene Datei und nicht für den Prozess:
// `storage.Oeffnen` nahm die Nachbardatei geteilt, stieß auf den eigenen
// Anspruch und wartete die volle `storage.Sperrfrist` ab.
//
// **Der Lauf war grün und dauerte zehneinhalb Sekunden.** Genau die Sorte
// Befund, die kein Zusicherungssatz findet — deshalb steht hier eine über die
// **Dauer**.
//
// **Die Schwelle ist großzügig**: die halbe Frist. Was der Lauf zeigen soll,
// ist der Unterschied zwischen „wartet gar nicht" und „wartet zehn Sekunden",
// nicht die Geschwindigkeit dieses Rechners.
func TestEinZurueckspielenWartetNichtAufSichSelbst(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	anwendung := laufendeAnwendung(t, bestand)

	ziel := gesichertMitSprache(t, anwendung, filepath.Join(ort, "gut.msesicherung"), "de")

	beginn := time.Now()

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if gedauert := time.Since(beginn); gedauert > storage.Sperrfrist/2 {
		t.Errorf("das Zurückspielen dauerte %v — es wartete auf den eigenen Anspruch",
			gedauert)
	}
}

// helferbinaer gibt einen Pfad auf dieses Testbinärprogramm, der bleibt.
//
// # Warum nicht `os.Args[0]`
//
// **Weil `go test` sein Programm in ein Arbeitsverzeichnis unter `$TMPDIR`
// legt, und das darf verschwinden.** Am 2026-09-15 ist `make ci` daran rot
// geworden:
//
//	starten: fork/exec /var/folders/…/T/go-build1507741411/b735/app.test:
//	no such file or directory
//
// **Der Lauf über `internal/app` dauert mit dem Race-Detektor 1.363
// Sekunden** — dreiundzwanzig Minuten, in denen diese Datei dastehen muss,
// damit ein Test in der Mitte sie noch einmal starten kann. Allein läuft er
// in einer Sekunde und ist grün; dreimal nachgestellt am selben Tag.
//
// **Eine Kopie nimmt die Frage weg**, statt sie kleiner zu machen. Sie liegt
// unter [os.MkdirTemp] dieses Laufs und wird am Ende mit ihm weggeräumt.
//
// # Warum einmal für den ganzen Lauf
//
// **Weil das Programm 77 MB wiegt** (gemessen am 2026-09-15). Eine Kopie je
// Test wäre eine Kopie zu viel; eine je Lauf ist der Preis dafür, dass der
// Helfer sicher startet.
func helferbinaer(t *testing.T) string {
	t.Helper()

	helfereinmal.Do(func() {
		helferort, helferfehler = kopiereBinaer(os.Args[0])
	})

	if helferfehler != nil {
		t.Fatalf("das Testprogramm ließ sich nicht kopieren: %v", helferfehler)
	}

	return helferort
}

var (
	helfereinmal sync.Once
	helferort    string
	helferfehler error
	helferordner string
)

// kopiereBinaer legt eine ausführbare Kopie an und gibt ihren Pfad.
func kopiereBinaer(quellpfad string) (string, error) {
	ordner, fehler := os.MkdirTemp("", "mse-helfer")
	if fehler != nil {
		return "", fehler
	}

	helferordner = ordner

	quelle, fehler := os.Open(quellpfad) //nolint:gosec // das eigene Programm
	if fehler != nil {
		return "", fehler
	}

	defer func() { _ = quelle.Close() }() //nolint:errcheck // gelesen ist gelesen

	ziel := filepath.Join(ordner, "helfer")

	//nolint:gosec // ein Pfad aus MkdirTemp
	geschrieben, fehler := os.OpenFile(ziel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o700)
	if fehler != nil {
		return "", fehler
	}

	if _, fehler := io.Copy(geschrieben, quelle); fehler != nil {
		_ = geschrieben.Close() //nolint:errcheck // der Fehler oben zählt

		return "", fehler
	}

	if fehler := geschrieben.Close(); fehler != nil {
		return "", fehler
	}

	return ziel, nil
}

// raeumeHelfer wirft die Kopie weg — gerufen von [TestMain].
func raeumeHelfer() {
	if helferordner != "" {
		_ = os.RemoveAll(helferordner) //nolint:errcheck // ein Ordner unter TMPDIR
	}
}

// **Die Kopie liegt außerhalb des Bauverzeichnisses und ist ausführbar.**
//
// # Die Zusage, für die es diesen Lauf gibt
//
// **`go test` legt sein Programm unter `$TMPDIR/go-build…` ab, und das darf
// verschwinden.** Am 2026-09-15 ist `make ci` genau daran rot geworden — der
// Lauf über `internal/app` dauert mit dem Race-Detektor dreiundzwanzig
// Minuten, und in der Mitte war die Datei fort.
//
// **Geprüft wird die Eigenschaft, nicht der Fall**: dass der Helferpfad ein
// anderer ist als `os.Args[0]` und nicht unter dessen Verzeichnis liegt.
// Ein Lauf, der nur „der Helfer startet" sagte, wäre wieder von der Laune
// des Aufräumers abhängig.
func TestDerHelferLiegtAusserhalbDesBauverzeichnisses(t *testing.T) {
	ort := helferbinaer(t)

	if ort == os.Args[0] {
		t.Fatal("der Helfer ist das Programm im Bauverzeichnis selbst")
	}

	if strings.HasPrefix(ort, filepath.Dir(os.Args[0])+string(filepath.Separator)) {
		t.Errorf("%q liegt im Bauverzeichnis %q",
			ort, filepath.Dir(os.Args[0]))
	}

	angaben, fehler := os.Stat(ort)
	if fehler != nil {
		t.Fatalf("der Helfer ist nicht da: %v", fehler)
	}

	if angaben.Mode().Perm()&0o100 == 0 {
		t.Errorf("der Helfer ist nicht ausführbar: %v", angaben.Mode())
	}

	// **Und zweimal gefragt ist einmal kopiert** — 77 MB je Test wären eine
	// Kopie zu viel.
	if zweitens := helferbinaer(t); zweitens != ort {
		t.Errorf("zwei Kopien: %q und %q", ort, zweitens)
	}
}
