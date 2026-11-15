package app

// **Der Schluss räumt die eigenen Läufe, bevor der Bestand zugeht** — BK-07a
// (#2115).
//
// `App.Beenden` beendete Dienste und Gerätemodell und schloss dann den
// Bestand. Was sonst lief — der Horchposten, die Sitzungen der Gäste, der
// Tisch beim Wirt —, lief weiter, über einem Bestand, den es nicht mehr gab.

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitgeschrieben ist ein Protokoll, das aus mehreren Goroutinen beschrieben
// wird.
type mitgeschrieben struct {
	schloss sync.Mutex
	puffer  bytes.Buffer
}

func (m *mitgeschrieben) Write(roh []byte) (int, error) {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	return m.puffer.Write(roh) //nolint:wrapcheck // ein Puffer scheitert nicht
}

func (m *mitgeschrieben) String() string {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	return m.puffer.String()
}

// **Nicht parallel**, denn er liest das Protokoll des Prozesses.
func TestDerSchlussBeendetDieSitzungenDesWirtsVorDemBestand(t *testing.T) {
	protokoll := &mitgeschrieben{}
	vorher := slog.Default()
	t.Cleanup(func() { slog.SetDefault(vorher) })
	slog.SetDefault(slog.New(slog.NewTextHandler(protokoll, nil)))

	wirt, gast := zweiAnwendungen(t)
	amTischSetzen(t, wirt, gast)

	if stand := gast.anwendung.PartieBeitrittStand(gast.ctx); !stand.Laeuft {
		t.Fatalf("vor dem Schluss sitzt der Gast nicht: %+v", stand)
	}

	angefangen := time.Now()

	if fehler := wirt.anwendung.Beenden(); fehler != nil {
		t.Fatalf("Beenden: %v", fehler)
	}

	if dauer := time.Since(angefangen); dauer > Schlussfrist+time.Second {
		t.Errorf("der Schluss brauchte %v", dauer)
	}

	// **Der Gast merkt es** — seine Leitung endet, sein Tisch pausiert. Ein
	// Wirt, der nur den Bestand schlösse, ließe ihn sitzen.
	if stand := wartetBisDerGastNichtMehrSitzt(t, gast); stand.Laeuft {
		t.Errorf("nach dem Schluss des Wirts sitzt der Gast noch: %+v", stand)
	}

	// **Und kein Lauf griff auf den geschlossenen Bestand** — gemessen vor
	// der Änderung: „eine Gastsitzung wurde abgewiesen grund=die Partie
	// lesen: sql: database is closed".
	if strings.Contains(protokoll.String(), "database is closed") {
		t.Errorf("ein Lauf griff nach dem Schließen auf den Bestand:\n%s", protokoll)
	}

	// **Danach nimmt das Tor nichts mehr an** — mit dem Abbruch und nicht mit
	// einem Fehler aus einem Bestand, der zu ist.
	if code := torcodeNachDemSchluss(t, wirt.anwendung); code != CodeAbgebrochen {
		t.Errorf("nach dem Schluss: %q", code)
	}

	// **Und ein zweiter Schluss ist keiner.**
	if fehler := wirt.anwendung.Beenden(); fehler != nil {
		t.Errorf("zweites Beenden: %v", fehler)
	}
}

// torcodeNachDemSchluss ruft eine Bindung und gibt den Code ihres Tors.
func torcodeNachDemSchluss(t *testing.T, anwendung *App) (code Fehlercode) {
	t.Helper()

	defer func() {
		var tor torfehler
		if grund, ist := recover().(error); ist && errors.As(grund, &tor) {
			code = tor.fehler.Code
		}
	}()

	NeueBindungen(anwendung).Sets(t.Context())

	return ""
}

// **Ein Start, der vor dem Schluss durchs Tor kam, endet trotzdem** —
// Zweitblick über A und B, Befund 2: `HorchStarten` wartete im Lauscher,
// der Schluss lief durch, und danach stand ein Horchposten, den niemand mehr
// beenden würde.
func TestEinStartNachDemSchlussBeendetSichSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	drin, weiter := make(chan struct{}), make(chan struct{})
	anwendung.MitHorcher(func(context.Context, string, peer.Identitaet) (service.Hereinkommend, error) {
		close(drin)
		<-weiter

		return &probelauscher{}, nil
	})

	antwort := make(chan Horchantwort, 1)

	go func() { antwort <- NeueBindungen(anwendung).HorchStarten(ctx, "127.0.0.1:0") }()

	<-drin

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("Beenden: %v", fehler)
	}

	close(weiter)

	if stand := (<-antwort).Stand; stand.Laeuft {
		t.Errorf("nach dem Schluss läuft ein Horchposten: %+v", stand)
	}

	if anwendung.horchstand().Laeuft {
		t.Error("der Stand hält einen Horchposten nach dem Schluss")
	}
}
