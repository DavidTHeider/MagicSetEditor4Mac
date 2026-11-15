package service_test

// **Der Horchposten** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass eine hereinkommende Verbindung zu einer Sitzung wird** — und dass ein
// krummes Gegenüber den Posten nicht umwirft. Das zweite ist die eigentliche
// Zusage: wer horcht, hört alles, und ein Fremder, der Unsinn schickt, darf
// nicht dafür sorgen, dass der Nächste nicht mehr durchkommt.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// horchdoppel gibt vorbereitete Verbindungen und dann einen Schlussfehler.
//
// **`haengt` hält `Nimm` fest, bis geschlossen wird** — so, wie ein echter
// `Accept` es tut: er kennt keinen Kontext und kehrt erst zurück, wenn der
// Lauscher zu ist.
type horchdoppel struct {
	schloss sync.Mutex
	folge   []func() (service.Verbindung, error)
	naechst int

	geschlossen bool

	// schliessfehler lässt das Niederlegen klagen.
	schliessfehler error

	// haengt macht aus `Nimm` ein echtes `Accept`: es kehrt erst zurück,
	// wenn der Lauscher geschlossen wird.
	haengt bool
	zu     chan struct{}

	// betreten sagt, dass `Nimm` wirklich wartet.
	betreten bool
}

func (h *horchdoppel) Nimm(context.Context) (service.Verbindung, error) {
	h.schloss.Lock()

	if h.haengt && h.naechst >= len(h.folge) {
		zu := h.zu
		h.betreten = true
		h.schloss.Unlock()

		<-zu

		return nil, domain.MitGrund(
			domain.ErrNotFound, nil, "der Lauscher ist zu")
	}

	defer h.schloss.Unlock()

	if h.naechst >= len(h.folge) {
		return nil, domain.MitGrund(
			domain.ErrNotFound, nil, "es kam keine Verbindung herein")
	}

	naechste := h.folge[h.naechst]
	h.naechst++

	return naechste()
}

// Adresse gibt eine erfundene — dieses Doppel horcht nirgends.
func (h *horchdoppel) Adresse() string { return "doppel:0" }

func (h *horchdoppel) Schliesse() error {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	if !h.geschlossen {
		h.geschlossen = true

		if h.zu != nil {
			close(h.zu)
		}
	}

	return h.schliessfehler
}

// **Eine hereinkommende Verbindung wird eine Sitzung.**
func TestEineHereinkommendeVerbindungWirdEineSitzung(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	posten := service.NeuerHorchposten(paar.dort, nil)

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{
			folge: []func() (service.Verbindung, error){
				func() (service.Verbindung, error) {
					return paar.dortigeLeitung, nil
				},
			},
		})
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	if len(paar.dort.Alle()) != 1 {
		t.Fatalf("es stehen %d Sitzungen darin", len(paar.dort.Alle()))
	}
}

// **Ein krummes Gegenüber wirft den Posten nicht um.**
//
// # Die Zusage, für die es diesen Lauf gibt
//
// **Wer horcht, hört alles.** Ein Fremder, der die Verbindung aufmacht und
// Unsinn schickt, darf nicht dafür sorgen, dass der Nächste nicht mehr
// durchkommt — das wäre eine Sperre, die jeder von außen auslösen kann.
//
// **Die Klasse entscheidet**: `ErrValidation` heißt „diese eine Verbindung ist
// nichts geworden", `ErrNotFound` heißt „der Lauscher gibt nichts mehr her".
// Beim ersten wird weitergehorcht, beim zweiten aufgehört.
func TestEinKrummesGegenueberWirftDenPostenNichtUm(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	// **Unter einem Schloss**, und nicht, weil dieser Lauf es bräuchte:
	// `Bediene` wartet auf jeden Handschlag, bevor es zurückkehrt, und der
	// Test liest erst danach. Aber die Zusage am Feld lautet „aus mehreren
	// Goroutinen", und ein Beispiel, das sie missachtet, ist das falsche.
	var (
		schloss  sync.Mutex
		gemeldet []error
	)

	posten := service.NeuerHorchposten(paar.dort, func(fehler error) {
		schloss.Lock()
		defer schloss.Unlock()

		gemeldet = append(gemeldet, fehler)
	})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{
			folge: []func() (service.Verbindung, error){
				// **Erst einer, der gar nicht erst zustande kam.**
				func() (service.Verbindung, error) {
					return nil, domain.MitGrund(domain.ErrValidation, nil,
						"die gesicherte Verbindung kam nicht zustande")
				},

				// **Dann einer, der durchkommt.**
				func() (service.Verbindung, error) {
					return paar.dortigeLeitung, nil
				},
			},
		})
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	if len(paar.dort.Alle()) != 1 {
		t.Fatalf("es stehen %d Sitzungen darin", len(paar.dort.Alle()))
	}

	if len(gemeldet) != 1 {
		t.Fatalf("gemeldet wurden %d Befunde: %v", len(gemeldet), gemeldet)
	}
}

// **Ein abgelehntes Gegenüber wird gemeldet und beendet nichts.**
func TestEinAbgelehntesGegenueberWirdGemeldet(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, false)

	// **Unter einem Schloss**, und nicht, weil dieser Lauf es bräuchte:
	// `Bediene` wartet auf jeden Handschlag, bevor es zurückkehrt, und der
	// Test liest erst danach. Aber die Zusage am Feld lautet „aus mehreren
	// Goroutinen", und ein Beispiel, das sie missachtet, ist das falsche.
	var (
		schloss  sync.Mutex
		gemeldet []error
	)

	posten := service.NeuerHorchposten(paar.dort, func(fehler error) {
		schloss.Lock()
		defer schloss.Unlock()

		gemeldet = append(gemeldet, fehler)
	})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{
			folge: []func() (service.Verbindung, error){
				func() (service.Verbindung, error) {
					return paar.dortigeLeitung, nil
				},
			},
		})
	}()

	_, _ = paar.hier.Oeffne( //nolint:errcheck // hier zählt die andere Seite
		t.Context(), paar.dortige.Kennung(), domain.ZweckHandel)

	<-fertig

	if len(gemeldet) != 1 || !errors.Is(gemeldet[0], domain.ErrValidation) {
		t.Fatalf("gemeldet wurde %v", gemeldet)
	}

	if len(paar.dort.Alle()) != 0 {
		t.Fatal("die abgelehnte Sitzung steht in der Übersicht")
	}
}

// **Ein abgebrochener Kontext beendet den Posten — und legt ihn nieder.**
func TestEinAbbruchBeendetDenPostenUndLegtIhnNieder(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	doppel := &horchdoppel{}
	posten := service.NeuerHorchposten(paar.dort, nil)

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	posten.Bediene(ctx, doppel)

	if !doppel.geschlossen {
		t.Fatal("der Posten ließ den Lauscher offen")
	}
}

// **Ein Posten ohne Melder verschluckt nichts, er schweigt nur.**
//
// **Der Unterschied zählt**: `nil` als Melder ist eine Entscheidung des
// Aufrufers und kein vergessenes Feld — ein Sturz beim Melden wäre schlimmer
// als ein ungehörter Befund.
func TestEinPostenOhneMelderSchweigt(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, false)

	posten := service.NeuerHorchposten(paar.dort, nil)

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{
			folge: []func() (service.Verbindung, error){
				func() (service.Verbindung, error) {
					return paar.dortigeLeitung, nil
				},
			},
		})
	}()

	_, _ = paar.hier.Oeffne( //nolint:errcheck // hier zählt die andere Seite
		t.Context(), paar.dortige.Kennung(), domain.ZweckHandel)

	<-fertig
}

// **Ein Lauscher, der beim Niederlegen klagt, wird gemeldet.**
//
// **Und er hält nichts auf.** Der Posten ist ohnehin am Ende; eine Klage
// darüber ist ein Befund fürs Protokoll und kein Grund, irgendetwas
// offenzulassen — dieselbe Gestalt wie bei [App.sitzungenPausieren].
func TestEinKlagenderLauscherWirdGemeldet(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var (
		schloss  sync.Mutex
		gemeldet []error
	)

	posten := service.NeuerHorchposten(paar.dort, func(fehler error) {
		schloss.Lock()
		defer schloss.Unlock()

		gemeldet = append(gemeldet, fehler)
	})

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	posten.Bediene(ctx, &horchdoppel{
		schliessfehler: errors.New("der Lauscher klagt"),
	})

	if len(gemeldet) != 1 {
		t.Fatalf("gemeldet wurden %d Befunde: %v", len(gemeldet), gemeldet)
	}
}

// **Ein Abbruch legt den Lauscher nieder, auch wenn `Nimm` darin festhängt.**
//
// # Der Verklemmer, für den es diesen Lauf gibt
//
// **`Accept` kennt keinen Kontext.** Es kehrt zurück, wenn eine Verbindung
// kommt — oder wenn der Lauscher geschlossen wird, und sonst nie. Bis zum
// 2026-09-14 legte [Horchposten.Bediene] ihn erst **nach** der Schleife
// nieder: der Abbruch kam an, die Schleife stand in `Accept`, und wer auf ihr
// Ende wartete, wartete für immer.
//
// **Gefunden hat es der erste Lauf über einen echten Port**
// (`internal/app/peertcp_intern_test.go`) — mit einem Doppel, dessen `Nimm`
// sofort zurückkehrt, war nichts zu sehen.
func TestEinAbbruchLegtDenLauscherNiederAuchWennNimmHaengt(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	doppel := &horchdoppel{haengt: true, zu: make(chan struct{})}
	posten := service.NeuerHorchposten(paar.dort, nil)

	ctx, abbrechen := context.WithCancel(t.Context())

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(ctx, doppel)
	}()

	// **Erst wenn er wirklich wartet** — sonst misst der Lauf einen Abbruch
	// vor dem ersten `Nimm`, und das ist der Fall daneben.
	for !doppel.wartet() {
		time.Sleep(time.Millisecond)
	}

	abbrechen()

	select {
	case <-fertig:
	case <-time.After(5 * time.Second):
		t.Fatal("der Posten kam nicht zurück")
	}

	if !doppel.geschlossen {
		t.Fatal("der Lauscher steht noch offen")
	}
}

// wartet sagt, ob `Nimm` gerade festhängt.
func (h *horchdoppel) wartet() bool {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	return h.betreten
}
