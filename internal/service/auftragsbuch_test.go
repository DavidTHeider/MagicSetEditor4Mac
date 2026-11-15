package service_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDasAuftragsbuchFuehrtLaufendeUndBeendete(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	buch := service.NeuesAuftragsbuch(func() time.Time { return jetzt })

	erster := buch.Beginnen("preise", "Preise für 120 Drucke")
	zweiter := buch.Beginnen("sammeldaten", "Sammeldaten laden")
	buch.Fortschreiben(erster, "40 von 120")

	alle := buch.Auftraege()
	if len(alle) != 2 || alle[0].Fortschritt != "40 von 120" || alle[1].Stand != domain.AuftragLaeuft {
		t.Fatalf("laufend: %+v", alle)
	}

	jetzt = jetzt.Add(time.Minute)

	// **Eine fremde Kennung geht an beiden Laufenden vorbei** — und der zweite
	// wird vor dem ersten beendet, damit die Suche einen Eintrag überspringt.
	buch.Beenden(99, nil)
	buch.Beenden(zweiter, errors.New("kein Netz"))
	buch.Beenden(erster, nil)

	// Die zuletzt beendeten zuerst: der erste wurde als Letzter geschlossen.
	alle = buch.Auftraege()
	if len(alle) != 2 || alle[0].ID != erster || alle[0].Stand != domain.AuftragFertig ||
		alle[0].Beendet != jetzt || alle[1].ID != zweiter ||
		alle[1].Stand != domain.AuftragGescheitert || alle[1].Grund != "kein Netz" {
		t.Errorf("beendet: %+v", alle)
	}

	// **Unbekannte Kennungen tun nichts** — und ein zweites Beenden auch nicht.
	buch.Fortschreiben(99, "x")
	buch.Beenden(99, nil)
	buch.Beenden(erster, nil)

	if len(buch.Auftraege()) != 2 {
		t.Errorf("nach unbekannten Kennungen: %+v", buch.Auftraege())
	}
}

func TestDasAuftragsbuchBehaeltNurDieJuengstenBeendeten(t *testing.T) {
	t.Parallel()

	buch := service.NeuesAuftragsbuch(time.Now)

	for i := range service.Behalten + 3 {
		id := buch.Beginnen("preise", "Lauf")
		if i%2 == 0 {
			buch.Beenden(id, nil)
		} else {
			buch.Beenden(id, errors.New("nein"))
		}
	}

	alle := buch.Auftraege()
	if len(alle) != service.Behalten || alle[0].ID != int64(service.Behalten+3) {
		t.Errorf("behalten wurden %d, der jüngste ist %d", len(alle), alle[0].ID)
	}
}

// **Aus vielen Goroutinen zugleich** — das Buch ist die Naht, an der sich
// mehrere Aufträge treffen.
func TestDasAuftragsbuchVertraegtNebenlaeufigkeit(t *testing.T) {
	t.Parallel()

	buch := service.NeuesAuftragsbuch(time.Now)

	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			id := buch.Beginnen("preise", "Lauf")
			buch.Fortschreiben(id, "halb")
			buch.Beenden(id, nil)
		}()
	}

	wg.Wait()

	if len(buch.Auftraege()) != service.Behalten {
		t.Errorf("%d Aufträge, erwartet %d beendete", len(buch.Auftraege()), service.Behalten)
	}
}
