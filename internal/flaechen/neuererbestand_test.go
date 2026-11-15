package flaechen_test

// **Ein Dienst schreibt nicht in einen Bestand, den eine neuere Fassung
// gewandert hat** — M02
// (#2346),
// Befund A-11 des Audits 0.9.0.
//
// Ein `mse-rest` unter launchd überlebt den Tausch des Bündels: der Prozess
// läuft mit dem alten Code weiter, und die neue Fassung wandert den Bestand
// unter ihm. Bis M02 schrieb er danach weiter — in ein Schema, das er nicht
// kennt. **Zwei echte Prozesse**: der Dienst, und dieser Test als die neuere
// Fassung, die eine Wanderung mit der nächsten freien Nummer geht.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dienstMitAusgang startet einen Dienst und gibt seinen Rückgabewert, sobald
// er von selbst endet.
func dienstMitAusgang(t *testing.T, umgebung []string) (an dienst, ausgang <-chan int) {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.rest, "-port", "0")
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if err := lauf.Start(); err != nil {
		abbrechen()
		t.Fatalf("den Dienst starten: %v", err)
	}

	geendet := make(chan struct{})
	code := make(chan int, 1)

	go func() {
		fehler := lauf.Wait()

		var beendet *exec.ExitError
		if errors.As(fehler, &beendet) {
			code <- beendet.ExitCode()
		} else {
			code <- 0
		}

		close(geendet)
	}()

	t.Cleanup(func() {
		abbrechen()
		<-geendet
	})

	adresse, marke, _ := ansageLesen(t, angesagt, geendet)

	return dienst{adresse: adresse, marke: marke, klient: &http.Client{Timeout: 2 * time.Minute}}, code
}

// neuereFassungWandert geht eine Wanderung, die dieser Bau nicht kennt — mit
// der nächsten freien Nummer, wie eine Fassung nach dieser.
func neuereFassungWandert(t *testing.T, bestand string) {
	t.Helper()

	db, err := storage.Oeffnen(t.Context(), bestand)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gewandert

	verbucht, err := storage.AngewendeteMigrationen(t.Context(), db)
	if err != nil {
		t.Fatal(err)
	}

	name := fmt.Sprintf("%04d_kuenftig.sql", len(verbucht)+1)
	weitere := fstest.MapFS{"migrationen/" + name: {Data: []byte("CREATE TABLE kuenftig (wert INTEGER) STRICT;")}}

	if fehler := storage.MigriereMit(t.Context(), db, weitere, nil); fehler != nil {
		t.Fatalf("die neuere Fassung wandert: %v", fehler)
	}
}

func TestEinDienstEndetVorEinemNeuerenBestand(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	mussKommandozeile(t, umgebung, "sets")

	var bestand string

	for _, eintrag := range umgebung {
		if wert, da := strings.CutPrefix(eintrag, "MSE_BESTAND="); da {
			bestand = wert
		}
	}

	an, ausgang := dienstMitAusgang(t, umgebung)
	einfahren := `{"pfad":` + alsJSONText(t, probeset(t)) + `}`

	if status, rumpf := an.ruf(t, http.MethodPost, "/sets", einfahren); status != http.StatusCreated {
		t.Fatalf("vorher schreibt er nicht: %d (%s)", status, rumpf)
	}

	neuereFassungWandert(t, bestand)

	// **Eine andere Quelle**: dieselbe noch einmal ergäbe dasselbe Set, und
	// ein Schreiben wäre an der Zahl nicht zu sehen.
	zweite := filepath.Join(t.TempDir(), "zweites.mse-set")
	if err := os.CopyFS(zweite, os.DirFS(probeset(t))); err != nil {
		t.Fatal(err)
	}

	status, rumpf := an.ruf(t, http.MethodPost, "/sets", `{"pfad":`+alsJSONText(t, zweite)+`}`)
	if status != http.StatusServiceUnavailable || ausRest(t, rumpf, "code") != "belegt" {
		t.Errorf("nach der Wanderung: %d (%s), erwartet 503 belegt", status, rumpf)
	}

	// **Und nicht mit 0**: unter launchd heißt das „nicht wieder starten"
	// (`SuccessfulExit` ist `false`, `schnittstellenziele.go`), und der Dienst
	// bliebe nach einer Wanderung über die Kommandozeile aus. Mit einem
	// anderen Code startet launchd ihn neu — mit dem Programm aus dem neuen
	// Bündel.
	select {
	case code := <-ausgang:
		if code == 0 {
			t.Error("der Dienst endete mit 0 — launchd startete ihn nicht wieder")
		}
	case <-time.After(15 * time.Second):
		t.Error("der Dienst lief weiter — er hätte gehen sollen")
	}

	// **Und geschrieben hat er nichts mehr** — gezählt am Bestand, den diese
	// Fassung nicht mehr öffnet, also lesend unter ihr.
	db, err := storage.Oeffnen(t.Context(), bestand)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gezählt

	var saetze int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM sets").Scan(&saetze); err != nil {
		t.Fatal(err)
	}

	if saetze != 1 {
		t.Errorf("im Bestand stehen %d Sets, erwartet das eine von vorher", saetze)
	}
}

// **Und über MCP dasselbe** — eine Sitzung bleibt offen wie die eines
// Sprachmodells, und unter ihr wandert eine neuere Fassung (§16: dieselbe
// Fähigkeit, überall dieselbe Antwort).
func TestEineMCPSitzungEndetVorEinemNeuerenBestand(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	mussKommandozeile(t, umgebung, "sets")

	var bestand string

	for _, eintrag := range umgebung {
		if wert, da := strings.CutPrefix(eintrag, "MSE_BESTAND="); da {
			bestand = wert
		}
	}

	sitzung := starteMCPProzess(t, umgebung)

	if text, istFehler := sitzung.ruf("schreiben.set.einfahren", map[string]any{"pfad": probeset(t)}); istFehler {
		t.Fatalf("vorher schreibt sie nicht: %s", text)
	}

	neuereFassungWandert(t, bestand)

	zweite := filepath.Join(t.TempDir(), "zweites.mse-set")
	if err := os.CopyFS(zweite, os.DirFS(probeset(t))); err != nil {
		t.Fatal(err)
	}

	text, istFehler := sitzung.ruf("schreiben.set.einfahren", map[string]any{"pfad": zweite})
	if !istFehler || !strings.Contains(text, "älter als der Bestand") {
		t.Errorf("nach der Wanderung: %v %s", istFehler, text)
	}

	// **Sie endet von selbst**: die Ausgabe schließt, ohne dass die Eingabe
	// zuging.
	geschlossen := make(chan error, 1)

	go func() {
		_, fehler := sitzung.heraus.ReadString('\n')
		geschlossen <- fehler
	}()

	select {
	case fehler := <-geschlossen:
		if !errors.Is(fehler, io.EOF) {
			t.Errorf("statt des Endes: %v", fehler)
		}
	case <-time.After(15 * time.Second):
		t.Error("die Sitzung lief weiter — sie hätte gehen sollen")
	}
}

// **Ein alter Dienst, der über einem neueren Bestand anläuft, endet mit 0** —
// Zweitblick über Phase 2 von 0.9.0, Befund 5 (#2323): zeigt die plist auf ein
// Programm einer älteren Fassung, startete launchd es nach jedem Ende mit 1
// wieder, alle zehn Sekunden, ohne Ende. Mit 0 bleibt es aus, und der Grund
// steht in der Ausgabe.
func TestEinAlterDienstStartetUeberEinemNeuerenBestandNichtImmerWieder(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	mussKommandozeile(t, umgebung, "sets")

	var bestand string

	for _, eintrag := range umgebung {
		if wert, da := strings.CutPrefix(eintrag, "MSE_BESTAND="); da {
			bestand = wert
		}
	}

	neuereFassungWandert(t, bestand)

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.rest, "-port", "0")
	lauf.Env = umgebung

	ausgabe, fehler := lauf.CombinedOutput()
	if fehler != nil {
		t.Errorf("der Dienst endete mit %v, erwartet 0 — launchd liefe sonst im Kreis\n%s", fehler, ausgabe)
	}

	if !strings.Contains(string(ausgabe), "neueren Fassung") {
		t.Errorf("ohne Grund:\n%s", ausgabe)
	}
}
