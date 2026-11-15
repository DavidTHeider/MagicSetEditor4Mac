// **Der Abbruch** — Werkstück 7 von Phase 4.13
// (#1352).
//
// # Die Frage aus dem Issue
//
// *„Was geschieht bei ⌘Q mitten in einem Schreibvorgang, bei einem Absturz,
// beim Ruhezustand mit offener Datenbank?"*
//
// **Der ⌘Q-Weg ist am 2026-09-07 durch den gepinnten Wails-Quelltext verfolgt**
// (v2.13.0): `applicationShouldTerminate:` gibt `NSTerminateCancel` zurück und
// schickt stattdessen die Nachricht `"Q"`; die führt über `Frontend.Quit` und
// `[NSApp stop:]` aus der Hauptschleife heraus, `RunMainLoop` kehrt zurück,
// `wails.Run` kehrt zurück — und damit das `defer` in `main.go`, das den
// Bestand schließt. **Kein `os.Exit` auf diesem Weg**, und das war die offene
// Frage.
//
// **Hier stand `OnShutdown` als Zwischenschritt, und das war falsch**: Wails
// ruft ihn an dieser Stelle, aber diese Anwendung setzt keinen. Angestrichen
// vom Zweitblick über diese Phase.
//
// **Der Absturz lässt sich nicht verfolgen, nur herstellen** — und genau das
// tut dieser Lauf: SIGKILL an einen Prozess, der den Bestand offen hält.
//
// **Der Ruhezustand steht hier nicht**, und das gehört gesagt: er lässt sich
// in einem Testlauf nicht herstellen. Was an seine Stelle tritt, ist die
// gemessene Einstellung — `journal_mode = wal`, `synchronous = 2` (FULL),
// gemessen am 2026-09-07: jede festgeschriebene Änderung liegt vor der Antwort
// auf der Platte.
package flaechen_test

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// **Ein hart getöteter Prozess sperrt den Bestand nicht — und was er
// festgeschrieben hat, ist da.**
//
// # Zwei Zusagen in einem Lauf, und sie sind nicht dieselbe
//
//  1. **Die Sperre stirbt mit dem Prozess.** Sie ist eine Sperre des
//     Betriebssystems auf `bestand.db.sperre`; eine, die einen Absturz
//     überlebte, machte den Bestand für immer unbenutzbar.
//  2. **Was quittiert war, bleibt.** Der Dienst hat mit `201` geantwortet,
//     bevor der Schlag kam — `synchronous = FULL` heißt, dass die Antwort
//     erst nach dem `fsync` ging.
func TestEinHartGetoeteterProzessLaesstDenBestandBrauchbar(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	// **Erst ein Bestand**, sonst gäbe es nichts zu sperren.
	mussKommandozeile(t, umgebung, "sets")

	an, toeten := dienstZumToeten(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/sets",
		`{"pfad":`+alsJSONText(t, probeset(t))+`}`)
	if status != http.StatusCreated {
		t.Fatalf("einfahren: %d (%s)", status, rumpf)
	}

	setID := ausRest(t, rumpf, "setID")

	toeten()

	// **Und jetzt geht der Bestand wieder auf.** Ginge er nicht, läge es an
	// einer Sperre, die den Prozess überlebt hat — und die machte den Bestand
	// für immer unbenutzbar.
	aus := mussKommandozeile(t, umgebung, "sets")

	if !strings.Contains(aus, setID) {
		t.Errorf("das quittierte Set %s ist nach dem Schlag fort:\n%s", setID, aus)
	}
}

// dienstZumToeten startet einen Dienst und gibt den Weg, ihn hart zu töten.
//
// **Ein eigener Helfer und keine Erweiterung von [starteDienst]**: jener
// beendet freundlich und wartet, und genau das ist hier der falsche Weg — was
// geprüft wird, ist der **unfreundliche**.
func dienstZumToeten(t *testing.T, umgebung []string) (an dienst, toeten func()) {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())
	t.Cleanup(abbrechen)

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.rest, "-port", "0")
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if err := lauf.Start(); err != nil {
		t.Fatalf("den Dienst starten: %v", err)
	}

	geendet := make(chan struct{})

	go func() {
		_ = lauf.Wait() //nolint:errcheck,gosec // was zählt, ist dass er weg ist
		close(geendet)
	}()

	adresse, marke, _ := ansageLesen(t, angesagt, geendet)

	an = dienst{
		adresse: adresse,
		marke:   marke,
		klient:  &http.Client{Timeout: 2 * time.Minute},
	}

	return an, func() {
		if err := lauf.Process.Kill(); err != nil {
			t.Fatalf("den Dienst töten: %v", err)
		}

		<-geendet
	}
}
