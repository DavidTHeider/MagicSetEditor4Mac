package flaechen_test

// **Ein Dienst fragt seinen eigenen Schalter vor jeder Anfrage und vor dem
// Start** — Z7 des Zweitblicks über Phase 6.12
// (#1905).
//
// Der Werkzeugserver über HTTP fragte ihn nur vor einem Werkzeug: `ping`,
// `initialize` und `tools/list` gingen weiter durch, und der Prozess lief.
// Und beide Programme öffneten ihren Port, auch wenn ihr Schalter schon aus
// war.

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerWerkzeugserverFragtSeinenSchalterBeiJederAnfrage(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	angesagt := starteMCPAn(t, umgebung, "127.0.0.1")

	ping := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	if status, rumpf := sendeMCP(t, angesagt.adresse, angesagt.marke, ping); status != http.StatusOK {
		t.Fatalf("vorher: %d %s", status, rumpf)
	}

	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.SchalterEinstellung("werkzeugserver"), domain.SchalterAus)

	if status, rumpf := sendeMCP(t, angesagt.adresse, angesagt.marke, ping); status != http.StatusForbidden ||
		!strings.Contains(rumpf, "abgeschaltet") {
		t.Fatalf("ping nach dem Abschalten: %d %s", status, rumpf)
	}
}

// **Mit seinem Schalter schon aus startet ein Dienst gar nicht** — und sagt,
// warum.
func TestEinAbgeschalteterDienstStartetNicht(t *testing.T) {
	for programm, schalter := range map[string]string{
		gebaut.rest: "dienst", gebaut.mcp: "werkzeugserver",
	} {
		t.Run(filepath.Base(programm), func(t *testing.T) {
			umgebung := mitEigenerSaat(t)
			mussKommandozeile(t, umgebung, "einstellungen", "setzen",
				domain.SchalterEinstellung(schalter), domain.SchalterAus)

			ctx, fertig := context.WithTimeout(t.Context(), 30*time.Second)
			defer fertig()

			// #nosec G204 -- gebautes Programm aus TestMain
			lauf := exec.CommandContext(ctx, programm, "--adresse", "127.0.0.1", "--port", "0")
			lauf.Env = umgebung

			// **Mit Rückgabe 0** — N4 des Nachlaufs über Phase 6.12. Der
			// LaunchAgent startet `mse-rest` neu, wenn es mit einem Fehler
			// endet (`KeepAlive` mit `SuccessfulExit: false`, `man
			// launchd.plist`); ein bewusst abgeschalteter Dienst wäre sonst
			// immer wieder angelaufen.
			aus, fehler := lauf.CombinedOutput()
			if fehler != nil || ctx.Err() != nil || !strings.Contains(string(aus), "abgeschaltet") ||
				strings.Contains(string(aus), "Marke") {
				t.Errorf("gestartet oder ohne Grund beendet: %v\n%s", fehler, aus)
			}
		})
	}
}
