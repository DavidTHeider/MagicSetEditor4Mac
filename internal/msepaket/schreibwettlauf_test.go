// **Zwei Prozesse an derselben Paketdatei** — der Befund des Zweitblicks
// durch Codex über Phase 3.3.
//
// # Der Befund
//
// Prüfung und Austausch waren nicht serialisiert. Zwei Fenster lesen denselben
// Stand, beide bestehen `standPruefen`, und beide schreiben. Der zweite
// schiebt dabei die **neue** Datei des ersten in die Sicherung und setzt
// seinen eigenen Stand ein. Beide Aufrufe melden Erfolg, und die Zusage im
// `CHANGELOG.md` — „ein zweites Fenster wird abgewiesen statt
// überschrieben" — war falsch.
//
// # Was hier gemessen wird
//
// Sechs Prozesse schreiben **gleichzeitig** dieselbe Datei, alle mit der
// Prüfsumme desselben gelesenen Standes. Genau einer darf durchkommen.
//
// **Und das ist unabhängig von der Reihenfolge.** Wer die Sperre bekommt,
// liest den Stand **hinter** ihr; wer nach ihm liest, sieht dessen Inhalt und
// nicht mehr den erwarteten. Ob die Prozesse sich wirklich überlappen, ändert
// das Ergebnis nicht — es entscheidet nur, ob die **Gegenprobe** greift.
package msepaket_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// wettlaufschalter sagt dem zweiten Prozess, welche Datei er schreiben soll.
const wettlaufschalter = "MSE_SCHREIBWETTLAUF"

// startschalter benennt die Datei, auf deren Erscheinen alle warten.
const startschalter = "MSE_SCHREIBSTART"

// TestMain lässt den Testlauf auch der zweite Prozess sein.
func TestMain(m *testing.M) {
	if ort := os.Getenv(wettlaufschalter); ort != "" {
		schreibeUmDieWette(ort, os.Getenv(startschalter))

		return
	}

	os.Exit(m.Run())
}

// schreibeUmDieWette wartet auf das Startzeichen und schreibt dann.
//
// **Der erwartete Stand steht fest** und ist der, den alle vor dem Start
// gelesen haben — `alt`. Genau darum geht es: sechs Schreiber mit demselben
// Ausgangsstand.
func schreibeUmDieWette(ort, startzeichen string) {
	warteAufStart(startzeichen)

	fehler := msepaket.DateiSchreiben(context.Background(), ort, "style",
		[]byte("von "+os.Getenv("MSE_SCHREIBER")+"\n"), summe([]byte("alt\n")))

	if fehler != nil {
		if _, err := os.Stdout.WriteString("abgewiesen\n"); err != nil {
			os.Exit(3)
		}

		return
	}

	if _, err := os.Stdout.WriteString("geschrieben\n"); err != nil {
		os.Exit(3)
	}
}

// warteAufStart hält den Helfer an, bis die Datei da ist.
//
// **Damit sich die Prozesse überlappen.** Ohne das startete jeder, wenn sein
// Vorgänger längst fertig wäre, und die Gegenprobe zeigte nichts.
func warteAufStart(zeichen string) {
	if zeichen == "" {
		return
	}

	for range 2000 {
		if _, err := os.Stat(zeichen); err == nil { // #nosec G703 -- aus t.TempDir()
			return
		}

		time.Sleep(time.Millisecond)
	}
}

// **Sechs Prozesse schreiben zugleich — genau einer kommt durch.**
func TestZweiProzesseSchreibenNichtUebereinander(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte("alt\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	zeichen := filepath.Join(t.TempDir(), "los")

	var welle sync.WaitGroup

	antworten := make([]string, 6)

	for nummer := range antworten {
		welle.Add(1)

		go func() {
			defer welle.Done()

			antworten[nummer] = zweiterSchreibt(t, ort, zeichen, nummer)
		}()
	}

	// **Erst wenn alle warten, fällt das Startzeichen.**
	time.Sleep(150 * time.Millisecond)

	if err := os.WriteFile(zeichen, nil, 0o600); err != nil {
		t.Fatalf("das Startzeichen setzen: %v", err)
	}

	welle.Wait()

	durch := 0

	for _, antwort := range antworten {
		if antwort == "geschrieben" {
			durch++
		}
	}

	if durch != 1 {
		t.Errorf("%d von %d Prozessen kamen durch — genau einer darf es",
			durch, len(antworten))
	}

	// **Und der Stand ist der eines einzigen Schreibers.** Ein gemischter
	// Inhalt käme hier an, ohne dass die Zählung darüber etwas gemerkt hätte.
	da, err := os.ReadFile(filepath.Join(ort, "style")) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.HasPrefix(string(da), "von ") {
		t.Errorf("es steht %q da", da)
	}
}

// zweiterSchreibt startet einen Helfer und gibt seine Meldung.
func zweiterSchreibt(t *testing.T, ort, zeichen string, nummer int) string {
	t.Helper()

	// #nosec G204,G702 -- der Testlauf selbst
	befehl := exec.CommandContext(t.Context(), os.Args[0])
	befehl.Env = append(os.Environ(),
		wettlaufschalter+"="+ort,
		startschalter+"="+zeichen,
		"MSE_SCHREIBER="+strings.Repeat("x", nummer+1))

	roh, err := befehl.Output()
	if err != nil {
		t.Errorf("der Helfer %d scheiterte: %v", nummer, err)

		return ""
	}

	return strings.TrimSpace(string(roh))
}

// **Wer die Sperre nicht bekommt, wird abgewiesen — und wartet nicht ewig.**
//
// Hier hält ein zweiter Griff sie schon. Die Frist wird für diesen Lauf
// heruntergesetzt: was geprüft wird, ist die **Abweisung**, nicht die Dauer.
func TestEineGehalteneSperreWeistDenZweitenAb(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte("alt\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	griff, err := dateisperre.Oeffnen(ort + msepaket.Schreibsperrendung)
	if err != nil {
		t.Fatalf("die Sperre öffnen: %v", err)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("die Sperre nehmen: %v", fehler)
	}

	frist, takt := msepaket.Schreibfrist, msepaket.Schreibtakt
	msepaket.Schreibfrist, msepaket.Schreibtakt = 40*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { msepaket.Schreibfrist, msepaket.Schreibtakt = frist, takt })

	fehler := msepaket.DateiSchreiben(t.Context(), ort, "style",
		[]byte("neu\n"), summe([]byte("alt\n")))

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", fehler)
	}

	// **Und geschrieben wurde nichts.**
	da, lesefehler := os.ReadFile(filepath.Join(ort, "style")) // #nosec G304 -- t.TempDir()
	if lesefehler != nil || string(da) != "alt\n" {
		t.Errorf("es steht %q da (%v)", da, lesefehler)
	}
}

// **Ein Verzeichnis, das erst nach dem ersten Schreiben eng wird.**
//
// # Warum dieser Umweg
//
// Die Sperrdatei liegt **neben** dem Paket. Wird das Elternverzeichnis eng,
// scheitert schon sie — und der Weg dahinter, an dem die Nachbardatei des
// Archivs entsteht, wird nie betreten. Existiert die Sperrdatei aber schon,
// lässt ein enges Verzeichnis sie weiter **öffnen** (das Verzeichnis regelt
// Anlegen und Löschen, nicht das Öffnen) — und dann kommt der Lauf genau
// dorthin.
func TestEinSpaeterEngesVerzeichnisMeldetDieNachbardatei(t *testing.T) {
	ort := probearchiv(t)
	vorher := "mse version: 2.0.0\ngame: probe\n"

	// **Erst ein glückender Schreibvorgang** — er legt die Sperrdatei an.
	if err := msepaket.DateiSchreiben(t.Context(), ort, "style",
		[]byte(vorher), summe([]byte(vorher))); err != nil {
		t.Fatalf("der erste Schreibvorgang: %v", err)
	}

	eng := filepath.Dir(ort)
	if err := os.Chmod(eng, 0o500); err != nil { // #nosec G302 -- Verzeichnis
		t.Fatalf("engen: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(eng, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	fehler := msepaket.DateiSchreiben(t.Context(), ort, "style",
		[]byte("x\n"), summe([]byte(vorher)))

	if fehler == nil {
		t.Fatal("das enge Verzeichnis nahm die Datei an")
	}

	if !strings.Contains(fehler.Error(), "Nachbardatei") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}
