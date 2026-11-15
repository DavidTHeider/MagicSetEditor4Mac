package app

// **Ein Datei-Lauf des Fensters zählt und lässt sich anhalten** — FMT-03.AK-02
// (#1692).

import (
	"context"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinDateilaufZaehltMeldetUndLaesstSichAbbrechen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	var (
		sperre   sync.Mutex
		gemeldet []AuftragDTO
	)

	anwendung.MitMeldung(func(_ context.Context, name string, daten ...any) {
		sperre.Lock()
		defer sperre.Unlock()

		if dto, ist := daten[0].(AuftragDTO); name == EreignisAuftrag && ist {
			gemeldet = append(gemeldet, dto)
		}
	})

	lauf, ende := anwendung.dateilauf(t.Context(), "bogen", "Mein Set")
	service.Melde(lauf, domain.Fortschritt{Erledigt: 3, Gesamt: 9, Einheit: domain.EinheitKarte})

	laufend := anwendung.Auftraege(t.Context()).Auftraege[0]
	if !laufend.Abbrechbar || laufend.Erledigt != 3 || laufend.Gesamt != 9 || laufend.Einheit != "karte" ||
		laufend.Art != "bogen" || laufend.Was != "Mein Set" {
		t.Fatalf("laufend: %+v", laufend)
	}

	bindungen := NeueBindungen(anwendung)

	if antwort := bindungen.AuftragAbbrechen(t.Context(), laufend.ID); antwort.Fehler != nil || lauf.Err() == nil {
		t.Fatalf("abbrechen: %+v, Lauf %v", antwort.Fehler, lauf.Err())
	}

	ende(lauf.Err())

	sperre.Lock()
	defer sperre.Unlock()

	if len(gemeldet) != 3 || gemeldet[0].Stand != "laeuft" || gemeldet[1].Erledigt != 3 ||
		gemeldet[2].Stand != "abgebrochen" || gemeldet[2].Abbrechbar {
		t.Errorf("gemeldet: %+v", gemeldet)
	}

	for kennung, code := range map[string]Fehlercode{laufend.ID: CodeNichtGefunden, "x": CodeValidierung} {
		if antwort := bindungen.AuftragAbbrechen(t.Context(), kennung); antwort.Fehler == nil || antwort.Fehler.Code != code {
			t.Errorf("%q: %+v", kennung, antwort.Fehler)
		}
	}
}
