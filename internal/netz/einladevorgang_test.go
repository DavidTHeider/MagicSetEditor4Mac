// **Einer holt, die anderen warten** — das Werkstück aus
// #1163,
// Befund Z7-03: „Ein Mehrprozesstest belegt **einen** Ladevorgang und einen
// konsistenten Endstand."
//
// # Warum es das braucht
//
// Eindeutige Nachbardateinamen sorgen für den **konsistenten Endstand** — aber
// nicht für **einen** Ladevorgang. Zwei Prozesse mit demselben fehlenden Stand
// luden beide den vollen Rumpf; bei einer Scryfall-Sammeldatei sind das
// zweimal 392 MB.
//
// Angestrichen vom Zweitblick über Phase 3.2: das Werkstück stand im Issue und
// war nicht eingelöst.
package netz_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// ladeschalter sagt dem zweiten Prozess, wo der Zwischenspeicher liegt.
const ladeschalter = "MSE_LADESPEICHER"

// holeUeberDenSpeicher holt einen Eintrag und meldet, ob **er** geladen hat.
//
// # Erst bereit melden, dann auf das Zeichen warten
//
// Ohne diese Schranke wäre der Lauf grün, ohne seinen Fall zu enthalten:
// starten die Prozesse unter Last nacheinander, lädt der erste, und die
// übrigen finden den Stand schon vor. Sie meldeten trotzdem „gewartet" —
// dieses Wort heißt nur, dass ihr Rückruf nicht lief (Nachlauf über
// Phase 3.2, #1163).
//
// **Er wartet danach eine Zehntelsekunde**, bevor er antwortet: sonst wäre der
// Ladevorgang so kurz, dass die anderen ihn kaum je überlappen.
func holeUeberDenSpeicher(ort string) {
	if _, err := os.Stdout.WriteString("bereit\n"); err != nil {
		os.Exit(4)
	}

	if _, err := os.Stdin.Read(make([]byte, 1)); err != nil {
		os.Exit(5)
	}

	speicher := netz.NeuerSpeicher(ort, time.Hour, time.Now)

	var geladen bool

	if _, _, fehler := speicher.Hole(context.Background(), "karte/x", func() ([]byte, error) {
		geladen = true

		time.Sleep(100 * time.Millisecond)

		return []byte("die Antwort"), nil
	}); fehler != nil {
		os.Exit(2)
	}

	antwort := "gewartet\n"
	if geladen {
		antwort = "geladen\n"
	}

	if _, err := os.Stdout.WriteString(antwort); err != nil {
		os.Exit(3)
	}
}

// **Vier Prozesse, ein Ladevorgang.**
func TestVierProzesseLadenNurEinmal(t *testing.T) {
	const gleichzeitig = 4

	ort := filepath.Join(t.TempDir(), "speicher")

	// **Erst alle starten, dann alle loslassen.** Die Schranke ist der
	// Unterschied zwischen „vier Prozesse haben geladen" und „vier Prozesse
	// haben **zugleich** geladen".
	helfer := make([]*holer, 0, gleichzeitig)
	for range gleichzeitig {
		helfer = append(helfer, holerStarten(t, ort))
	}

	antworten := make(chan string, gleichzeitig)

	var laeuft sync.WaitGroup

	for _, einer := range helfer {
		laeuft.Add(1)

		go func() {
			defer laeuft.Done()

			antworten <- einer.losUndAntwort(t)
		}()
	}

	laeuft.Wait()
	close(antworten)

	var geladen, gewartet int

	for antwort := range antworten {
		switch antwort {
		case "geladen":
			geladen++
		case "gewartet":
			gewartet++
		default:
			t.Errorf("unerwartete Antwort %q", antwort)
		}
	}

	if geladen != 1 {
		t.Errorf("es luden %d Prozesse — verlangt ist einer", geladen)
	}

	if gewartet != gleichzeitig-1 {
		t.Errorf("es warteten %d Prozesse — verlangt sind %d", gewartet, gleichzeitig-1)
	}

	// **Und der Endstand ist einer.** Ein Eintrag, keine halben Nachbardateien.
	if eintraege := nurEintraege(t, ort); len(eintraege) != 1 {
		t.Errorf("im Speicher liegen %d Einträge: %v", len(eintraege), eintraege)
	}
}

// holer ist ein gestarteter Helfer, der an der Schranke wartet.
type holer struct {
	befehl *exec.Cmd
	hinein io.WriteCloser
	heraus io.ReadCloser
}

// holerStarten startet einen Helfer und wartet, bis er bereit ist.
func holerStarten(t *testing.T, ort string) *holer {
	t.Helper()

	// #nosec G204,G702 -- der Testlauf selbst
	befehl := exec.CommandContext(t.Context(), os.Args[0])
	befehl.Env = append(os.Environ(), ladeschalter+"="+ort)

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

	if _, fehler := heraus.Read(make([]byte, len("bereit\n"))); fehler != nil {
		t.Fatalf("der Helfer meldete sich nicht bereit: %v", fehler)
	}

	return &holer{befehl: befehl, hinein: hinein, heraus: heraus}
}

// losUndAntwort lässt den Helfer los und gibt seine Antwort.
func (h *holer) losUndAntwort(t *testing.T) string {
	t.Helper()

	if _, err := h.hinein.Write([]byte("x")); err != nil {
		t.Errorf("loslassen: %v", err)

		return ""
	}

	roh, err := io.ReadAll(h.heraus)
	if err != nil {
		t.Errorf("lesen: %v", err)

		return ""
	}

	if fehler := h.befehl.Wait(); fehler != nil {
		t.Errorf("der Helfer scheiterte: %v", fehler)

		return ""
	}

	return strings.TrimSpace(string(roh))
}
