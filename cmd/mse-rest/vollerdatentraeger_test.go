package main

// **Ein voller Datenträger kommt am Dienst als 507 an** — Befund 5 des
// Zweitblicks über Phase 6.1.
//
// # Warum es diesen Lauf gibt
//
// **Jedes Glied der Kette hatte seinen eigenen, und die Kette riss trotzdem.**
// Der Bestand meldet `ErrVoll` (`internal/storage/vollerbestand_test.go`),
// `app.Uebersetze` macht `CodeVoll` daraus, `internal/rest` bildet `ErrVoll`
// auf 507 ab (`fehlerwege_test.go`) — und dazwischen gab `AlsFehler` einen
// nackten Fehler zurück, weil seine Tabelle den Code nicht kannte. Am Dienst
// stand 500. Hier läuft der Weg **ab der Übersetzung der Anwendung**: über
// beide Rückwege, durch den echten Dienst bis zum Status.
//
// # Was hier nicht läuft
//
// **Der Bestand selbst.** Die Naht unten erzeugt `ENOSPC` und übersetzt es;
// dass SQLite einen vollen Datenträger als `ErrVoll` meldet, hält
// `internal/storage/vollerbestand_test.go`. Hier stand bis zum Nachlauf über
// Phase 6.1 „der ganze Weg" (Befund 5) — der Kopf versprach den Bestand mit.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// vollevorgaben scheitert, wie ein Schreibweg an einer vollen Platte scheitert.
type vollevorgaben struct{ zurueck func(*app.Fehler) error }

func (v vollevorgaben) Vorgaben(ctx context.Context, _ string) ([]rest.Vorgabenzeile, error) {
	return nil, v.zurueck(app.Uebersetze(ctx, fmt.Errorf("lesen: %w", syscall.ENOSPC)))
}

func (v vollevorgaben) VorgabeSetzen(ctx context.Context, _, _ string, _ []string) (
	rest.Vorgabengesetzt, error,
) {
	return rest.Vorgabengesetzt{}, v.zurueck(app.Uebersetze(ctx,
		domain.Platzmangel(fmt.Errorf("schreiben: %w", syscall.ENOSPC))))
}

func TestEinVollerDatentraegerKommtAmDienstAlsVollAn(t *testing.T) {
	for weg, zurueck := range map[string]func(*app.Fehler) error{
		"Methode":  func(befund *app.Fehler) error { return befund.AlsFehler() },
		"Funktion": app.AlsFehler,
	} {
		t.Run(weg, func(t *testing.T) {
			dienst := rest.NeuerDienst("marke", time.Now,
				rest.Naehte{Vorgaben: vollevorgaben{zurueck: zurueck}})

			anfrage := httptest.NewRequestWithContext(probekontext(t), http.MethodPut,
				"/profile/1/vorgaben/lagerort", strings.NewReader(`{"werte":["Regal"]}`))
			anfrage.Header.Set("Authorization", "Bearer marke")

			// **Die Schleife und nicht `example.com`** — der Dienst weist
			// jeden anderen Namen ab (0115).
			anfrage.Host = "127.0.0.1"

			antwort := httptest.NewRecorder()
			dienst.Behandler().ServeHTTP(antwort, anfrage)

			if antwort.Code != http.StatusInsufficientStorage {
				t.Errorf("Status %d, erwartet war 507 (%s)", antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), `"code":"voll"`) {
				t.Errorf("der Code fehlt: %s", antwort.Body.String())
			}
		})
	}
}
