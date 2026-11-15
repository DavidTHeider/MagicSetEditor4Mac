package app

// **Das Auftragsbuch am Fenster** — Werkstück U11 von Phase 5.1 (#1422).

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestDasAuftragsbuchZeigtLaufendeUndBeendete(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	anwendung := NeueApp("")
	anwendung.uhr = func() time.Time { return jetzt }

	beenden := anwendung.auftragBeginnen("preise", "Preise holen")
	jetzt = jetzt.Add(30 * time.Second)

	antwort := anwendung.Auftraege(t.Context())
	if antwort.Laufend != 1 || len(antwort.Auftraege) != 1 || antwort.Auftraege[0].Sekunden != 30 ||
		antwort.Auftraege[0].Begonnen != "2026-09-09T12:00:00Z" || antwort.Auftraege[0].ID != "1" {
		t.Errorf("laufend: %+v", antwort)
	}

	beenden(errors.New("kein Netz"))

	antwort = NeueBindungen(anwendung).Auftraege(t.Context())
	if antwort.Laufend != 0 || antwort.Auftraege[0].Stand != "gescheitert" || antwort.Auftraege[0].Grund != "kein Netz" {
		t.Errorf("beendet: %+v", antwort)
	}

	// **Das Buch entsteht einmal, auch wenn drei Wege es zugleich anlegen.**
	frisch := NeueApp("")

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			frisch.auftragBeginnen("sammeldaten", "laden")(nil)
		}()
	}

	wg.Wait()

	if len(frisch.Auftraege(t.Context()).Auftraege) != 5 {
		t.Errorf("%d Aufträge, erwartet fünf beendete", len(frisch.Auftraege(t.Context()).Auftraege))
	}
}
