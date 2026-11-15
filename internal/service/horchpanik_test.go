package service_test

// **Ein Handschlag, der zusammenbricht, wirft den Posten nicht um** — BK-05a
// (#2111).
//
// Jeder Handschlag läuft in einer eigenen Goroutine (`bedieneEine`). Eine
// Panik dort — im Handschlag selbst oder in der Aufnahme danach — fing
// niemand: der Prozess endete, und mit ihm jeder Tisch und jeder Handel.
// Wer sie fängt, muss außerdem freigeben, was der Handschlag hielt: den
// Platz unter den [service.HoechstensHandschlaege] und die Sitzung.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zerbrechendeLeitung bricht beim ersten Lesen zusammen.
type zerbrechendeLeitung struct {
	schloss     sync.Mutex
	geschlossen bool
}

func (*zerbrechendeLeitung) Kanal() string             { return "kanal-zerbrochen" }
func (*zerbrechendeLeitung) Gegenstelle() peer.Kennung { return "" }

func (*zerbrechendeLeitung) Sende(context.Context, []byte) error {
	panic("die Leitung bricht beim Senden zusammen")
}

func (*zerbrechendeLeitung) Empfange(context.Context) ([]byte, error) {
	panic("die Leitung bricht beim Lesen zusammen")
}

func (z *zerbrechendeLeitung) Schliesse() error {
	z.schloss.Lock()
	defer z.schloss.Unlock()

	z.geschlossen = true

	return nil
}

func (z *zerbrechendeLeitung) zu() bool {
	z.schloss.Lock()
	defer z.schloss.Unlock()

	return z.geschlossen
}

// meldebuch sammelt, was der Posten meldet — aus mehreren Goroutinen.
type meldebuch struct {
	schloss sync.Mutex
	fehler  []error
}

func (m *meldebuch) melde(fehler error) {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	m.fehler = append(m.fehler, fehler)
}

func (m *meldebuch) zusammenbrueche() int {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	gezaehlt := 0

	for _, fehler := range m.fehler {
		if errors.Is(fehler, domain.ErrZusammengebrochen) {
			gezaehlt++
		}
	}

	return gezaehlt
}

// **Vor der Übergabe**: mehr Handschläge brechen zusammen, als es Plätze gibt
// — und danach kommt eine gewöhnliche Sitzung trotzdem zustande.
func TestZusammenbrechendeHandschlaegeGebenIhrenPlatzFrei(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	buch := &meldebuch{}

	var (
		schloss      sync.Mutex
		aufgenommene int
	)

	posten := service.NeuerHorchposten(paar.dort, buch.melde).MitAufnahme(
		func(context.Context, service.Angenommene) {
			schloss.Lock()
			defer schloss.Unlock()

			aufgenommene++
		})

	zerbrochene := make([]*zerbrechendeLeitung, service.HoechstensHandschlaege+1)
	folge := []func() (service.Verbindung, error){}

	// **Eine nach der anderen**: die nächste kommt erst, wenn die vorige
	// gemeldet ist. Gleichzeitig wiese der Posten die dreiunddreißigste ab,
	// weil die anderen ihren Platz noch hielten — das prüfte die Grenze und
	// nicht das Freigeben. Ohne Freigabe kommt die Meldung nie, und nach der
	// Frist geht die nächste trotzdem hinaus — in einen vollen Posten.
	nachDer := func(zahl int) {
		frist := time.Now().Add(5 * time.Second)
		for buch.zusammenbrueche() < zahl && time.Now().Before(frist) {
			time.Sleep(time.Millisecond)
		}
	}

	for stelle := range zerbrochene {
		zerbrochene[stelle] = &zerbrechendeLeitung{}
		leitung := zerbrochene[stelle]
		folge = append(folge, func() (service.Verbindung, error) {
			nachDer(stelle)

			return leitung, nil
		})
	}

	folge = append(folge, func() (service.Verbindung, error) {
		nachDer(len(zerbrochene))

		return paar.dortigeLeitung, nil
	})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{folge: folge})
	}()

	if _, fehler := paar.hier.Oeffne(t.Context(), paar.dortige.Kennung(), domain.ZweckPartie); fehler != nil {
		t.Fatalf("Oeffne nach den Zusammenbrüchen: %v", fehler)
	}

	<-fertig

	if gezaehlt := buch.zusammenbrueche(); gezaehlt != len(zerbrochene) {
		t.Errorf("%d Zusammenbrüche gemeldet, erwartet %d", gezaehlt, len(zerbrochene))
	}

	for stelle, leitung := range zerbrochene {
		if !leitung.zu() {
			t.Errorf("die zerbrochene Leitung %d blieb offen", stelle)
		}
	}

	schloss.Lock()
	defer schloss.Unlock()

	if aufgenommene != 1 {
		t.Errorf("%d Aufnahmen, erwartet die eine gewöhnliche", aufgenommene)
	}
}

// **Nach der Übergabe**: die Aufnahme bricht zusammen — die Sitzung geht zu,
// statt als Leiche in der Übersicht zu stehen.
func TestEineZusammenbrechendeAufnahmeSchliesstIhreSitzung(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	buch := &meldebuch{}

	posten := service.NeuerHorchposten(paar.dort, buch.melde).MitAufnahme(
		func(context.Context, service.Angenommene) {
			panic("die Aufnahme bricht zusammen")
		})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), einmaligerLauscher(paar.dortigeLeitung))
	}()

	if _, fehler := paar.hier.Oeffne(t.Context(), paar.dortige.Kennung(), domain.ZweckPartie); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	if buch.zusammenbrueche() != 1 {
		t.Errorf("gemeldet: %v", buch.fehler)
	}

	if offen := paar.dort.Alle(); len(offen) != 0 {
		t.Errorf("nach dem Zusammenbruch stehen noch %d Sitzungen da: %+v", len(offen), offen)
	}
}
