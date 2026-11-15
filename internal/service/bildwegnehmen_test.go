// **Ein abgelegtes Bild wieder wegnehmen** — Befund Z7-02 von Phase 3.2
// (#1163).
//
// Eine eigene Datei, weil sie eine eigene Frage beantwortet: `bildablage_test`
// sagt, **wohin** ein Bild kommt und wie es heißt — hier steht, wie es wieder
// verschwindet, wenn der Verweis darauf nie entsteht.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein abgelegtes Bild lässt sich wieder wegnehmen** — Befund Z7-02 von
// Phase 3.2 (#1163).
//
// # Wofür es das gibt
//
// `Ablegen` schreibt die Datei, **bevor** ein Feld auf sie zeigt. Scheitert
// das Feld, läge eine Datei ohne Verweis da — und das Bilderverzeichnis wüchse
// bei jedem Fehlversuch um eine, die niemand je wieder findet.
func TestEinAbgelegtesBildLaesstSichWegnehmen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	name, err := ablage.Ablegen(7, "kunst", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	pfad := filepath.Join(ablage.Verzeichnis(7), name)
	if _, statfehler := os.Stat(pfad); statfehler != nil {
		t.Fatalf("das Bild fehlt schon vor dem Wegnehmen: %v", statfehler)
	}

	if fehler := ablage.Wegnehmen(7, name); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	if _, statfehler := os.Stat(pfad); !os.IsNotExist(statfehler) {
		t.Errorf("das Bild liegt noch da: %v", statfehler)
	}

	// **Ein zweites Mal ist kein Fehler.** Der Aufrufer räumt auf einem
	// Fehlerweg auf; dass dort schon jemand aufgeräumt hat, ist keine
	// schlechte Nachricht.
	if fehler := ablage.Wegnehmen(7, name); fehler != nil {
		t.Errorf("das zweite Wegnehmen meldete einen Fehler: %v", fehler)
	}
}

// **Ohne Verzeichnis und ohne Namen geschieht nichts.**
func TestWegnehmenOhneOrtOderNamenTutNichts(t *testing.T) {
	t.Parallel()

	if fehler := service.NeueBildablage("").Wegnehmen(7, "kunst1"); fehler != nil {
		t.Errorf("ohne Verzeichnis: %v", fehler)
	}

	if fehler := service.NeueBildablage(t.TempDir()).Wegnehmen(7, ""); fehler != nil {
		t.Errorf("ohne Namen: %v", fehler)
	}
}

// **Ein Name, der ein Weg ist, wird abgewiesen.**
//
// Was hier gelöscht wird, entscheidet nicht der Aufrufer: ein `../` führte aus
// dem Bilderverzeichnis hinaus.
func TestWegnehmenWeistEinenWegAb(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	daneben := filepath.Join(wurzel, "wichtig.txt")
	if err := os.WriteFile(daneben, []byte("nicht löschen"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	fehler := service.NeueBildablage(wurzel).Wegnehmen(7, "../wichtig.txt")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", fehler)
	}

	if _, statfehler := os.Stat(daneben); statfehler != nil {
		t.Errorf("die Nachbardatei ist fort: %v", statfehler)
	}
}

// **Lässt sich das Bild nicht wegnehmen, wird es gemeldet.**
//
// Der Aufrufer räumt auf einem Fehlerweg auf; scheitert auch das, verbindet er
// beide Meldungen. Ein stilles Verschlucken machte aus zwei Befunden einen.
func TestEinUnmoeglichesWegnehmenWirdGemeldet(t *testing.T) {
	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	name, err := ablage.Ablegen(7, "kunst", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	if chmodfehler := os.Chmod(ablage.Verzeichnis(7), 0o500); chmodfehler != nil {
		t.Fatalf("zusperren: %v", chmodfehler)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(ablage.Verzeichnis(7), 0o700) }) //nolint:errcheck // Aufräumen

	if fehler := ablage.Wegnehmen(7, name); fehler == nil {
		t.Error("das unmögliche Wegnehmen meldete Erfolg")
	}
}

// **Zwei gleichzeitige Einsetzungen bekommen zwei Namen.**
//
// # Der Befund des Nachlaufs
//
// `freierName` fragte mit `Stat` und ließ den Namen liegen; geschrieben wurde
// erst danach. Zwei gleichzeitige Einsetzungen im selben Set bekamen beide
// `kunst1` und schrieben in dieselbe Datei — und wenn danach **einer** von
// beiden Datenbankzügen scheiterte, nahm sein Aufräumen dem anderen das Bild
// weg. Übrig blieb ein gültiger Kartenwert ohne Bilddatei.
//
// **Der Lauf stellt Gleichzeitigkeit her und behauptet sie nicht**: alle
// Ableger warten an derselben Schranke.
func TestZweiGleichzeitigeEinsetzungenBekommenZweiNamen(t *testing.T) {
	const gleichzeitig = 8

	ablage := service.NeueBildablage(t.TempDir())

	los := make(chan struct{})
	namen := make(chan string, gleichzeitig)

	var laeuft sync.WaitGroup

	for range gleichzeitig {
		laeuft.Add(1)

		go func() {
			defer laeuft.Done()

			<-los

			name, err := ablage.Ablegen(7, "kunst", probebild())
			if err != nil {
				t.Errorf("ablegen: %v", err)

				return
			}

			namen <- name
		}()
	}

	close(los)
	laeuft.Wait()
	close(namen)

	gesehen := map[string]bool{}

	for name := range namen {
		if gesehen[name] {
			t.Errorf("der Name %q wurde zweimal vergeben — die eine Einsetzung "+
				"überschrieb die andere", name)
		}

		gesehen[name] = true
	}

	if len(gesehen) != gleichzeitig {
		t.Errorf("es entstanden %d Namen statt %d", len(gesehen), gleichzeitig)
	}
}

// **Ist jeder Name belegt, wird es gemeldet.**
//
// Die Schleife in `namenBelegen` gibt auf, wenn ihr jeder Name unter der Nase
// weggeschnappt wird — im Betrieb hieße das ein Verzeichnis, das jeden Namen
// belegt meldet, etwa ein kaputter Einhängepunkt.
func TestOhneFreienNamenWirdEsGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	ort := ablage.Verzeichnis(7)
	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Die Schranke sitzt an dieser Ablage**, nicht am Paket: die Läufe hier
	// sind nebenläufig, und ein geteilter Wert ließe jeden anderen an ihr
	// scheitern.
	ablage.MitHoechsterBildzahl(3)

	// **Jeder Name als Verzeichnis.** `os.OpenFile` mit `O_EXCL` scheitert
	// daran mit `EEXIST`, also sucht die Schleife weiter — bis zur Schranke.
	for zahl := 1; zahl <= 3; zahl++ {
		if err := os.Mkdir(filepath.Join(ort, "kunst"+strconv.Itoa(zahl)), 0o700); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	if _, fehler := ablage.Ablegen(7, "kunst", probebild()); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", fehler)
	}
}

// **Und ein Name, der sich nicht beschreiben lässt, ebenso.**
func TestEineUnbeschreibbareBilddateiWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	ort := ablage.Verzeichnis(7)
	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Das Verzeichnis ist zu**, nachdem es entstanden ist: `os.OpenFile`
	// scheitert mit `EACCES` und nicht mit `EEXIST`.
	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	if err := os.Chmod(ort, 0o500); err != nil {
		t.Fatalf("zusperren: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck // Aufräumen

	if _, fehler := ablage.Ablegen(7, "kunst", probebild()); fehler == nil {
		t.Error("das unbeschreibbare Verzeichnis kam durch")
	}
}
