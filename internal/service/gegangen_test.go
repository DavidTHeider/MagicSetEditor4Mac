package service_test

// **Wer gegangen ist, stellt nichts mehr** — CR 800.4a (Fassung 2026-09-25),
// #1979.
//
// Beim Bau von #1971 bemerkt: ein Sitz, der aufgegeben hatte, konnte weiter
// Leben, Mana und Zonen ändern — alles, was keine Priorität braucht. Der
// Ausleger fragte nur, ob es den Sitz gibt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// lebensbefehlVon ist ein Befehl, der keine Priorität braucht.
func lebensbefehlVon(sitz string) domain.Spielbefehl {
	return domain.Spielbefehl{
		PartieID: "p1", Absender: sitz, Kennung: "leben-" + sitz,
		Art: service.BefehlLeben, Nutzlast: map[string]any{"um": float64(1)},
	}
}

// **Der Ausleger weist ihn ab, und die Verbliebenen stellen weiter.**
func TestEinGegangenerSitzStelltKeineBefehle(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	// **Zu dritt**: zu zweit endet die Partie mit dem Gehen (CR 104.2a), und
	// eine beendete nimmt ohnehin nichts mehr an.
	partie := zugpartieIn(t, 3).Scheidetaus("sitz-2", domain.GrundAufgabe)
	if partie.Stand == domain.StandBeendet {
		t.Fatal("zu dritt endete die Partie mit dem ersten, der ging")
	}

	var befund *domain.FieldError

	_, _, fehler := kasten.Auslegen(partie, lebensbefehlVon("sitz-2"))
	if !errors.Is(fehler, domain.ErrConflict) || !errors.As(fehler, &befund) ||
		befund.Key != domain.SchluesselSitzGegangen {
		t.Errorf("der Gegangene stellte: %v", fehler)
	}

	if _, _, fehler := kasten.Auslegen(partie, lebensbefehlVon("sitz-3")); fehler != nil {
		t.Errorf("sitz-3 ist noch im Spiel: %v", fehler)
	}

	// **Und angeboten wird ihm auch nichts mehr** — sonst stünden Knöpfe da,
	// die nur abgewiesen würden.
	if erlaubt := kasten.Erlaubt(partie, "sitz-2"); len(erlaubt) != 0 {
		t.Errorf("dem Gegangenen wird angeboten: %v", erlaubt)
	}

	if erlaubt := kasten.Erlaubt(partie, "sitz-3"); len(erlaubt) == 0 {
		t.Error("sitz-3 wird nichts mehr angeboten")
	}
}
