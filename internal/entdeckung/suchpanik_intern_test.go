package entdeckung

// **Ein Zusammenbruch in der Suche wird zum Fehler** — BK-05c
// (#2114).
//
// Die Suche fragt je Schnittstelle in einer eigenen Goroutine
// (`ueberAlle`). Eine Panik dort — etwa aus der Bibliothek, die eine
// fremde Antwort deutet — fängt niemand: der Prozess endete, mitsamt dem
// Fenster, das nur wissen wollte, wer im Netz ist.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinZusammenbruchJederSchnittstelleWirdZumFehler(t *testing.T) {
	zweiSchnittstellen(t)
	mitRufen(t, func(int, *mdns.QueryParam) error { panic("die Bibliothek bricht zusammen") })

	angefangen := time.Now()

	if _, fehler := Suche(context.Background(), Frage{Frist: time.Second}); !errors.Is(fehler, domain.ErrZusammengebrochen) {
		t.Errorf("die Suche gab %v", fehler)
	}

	if gebraucht := time.Since(angefangen); gebraucht > 5*time.Second {
		t.Errorf("der Fehler kam erst nach %v", gebraucht)
	}

	// **Und die nächste Suche geht wieder** — nichts blieb hängen.
	mitRufen(t, func(int, *mdns.QueryParam) error { return nil })

	if _, fehler := Suche(context.Background(), Frage{Frist: time.Second}); fehler != nil {
		t.Errorf("die Suche danach: %v", fehler)
	}
}

// **Bricht nur eine zusammen, antwortet die andere** — dieselbe Regel wie
// bei einem gewöhnlichen Fehler: gescheitert ist die Suche erst, wenn jede
// Schnittstelle scheitert.
func TestEinZusammenbruchNimmtDieAndereSchnittstelleNichtMit(t *testing.T) {
	zweiSchnittstellen(t)
	mitRufen(t, func(zahl int, _ *mdns.QueryParam) error {
		if zahl == 1 {
			panic("die Bibliothek bricht zusammen")
		}

		return nil
	})

	if _, fehler := Suche(context.Background(), Frage{Frist: time.Second}); fehler != nil {
		t.Errorf("die Suche gab %v", fehler)
	}
}
