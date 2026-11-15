package service_test

// **Gibt es dieses Set?** —
// #1336.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// setkoepfe gibt feste Setköpfe.
type setkoepfe struct {
	koepfe []domain.SetHead
	fehler error
}

func (s setkoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return s.koepfe, s.fehler
}

// **Ein Set, das es gibt, kommt durch.**
func TestEinVorhandenesSetKommtDurch(t *testing.T) {
	t.Parallel()

	leser := setkoepfe{koepfe: []domain.SetHead{{ID: 1}, {ID: 7}}}

	if fehler := service.SetDaSein(t.Context(), leser, 7); fehler != nil {
		t.Errorf("SetDaSein: %v", fehler)
	}
}

// **Eine Kennung, die auf nichts zeigt, ist `ErrNotFound`.**
//
// **Nicht `ErrValidation`**: die Kennung ist wohlgeformt, sie zeigt nur auf
// nichts (§8). Der Unterschied entscheidet, was eine Fläche daraus macht —
// 404 gegen 400, Rückgabewert 1 gegen 2.
func TestEineFehlendeKennungIstNichtGefunden(t *testing.T) {
	t.Parallel()

	leser := setkoepfe{koepfe: []domain.SetHead{{ID: 1}}}

	fehler := service.SetDaSein(t.Context(), leser, 999)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}

	// **Die Kennung steht in der Meldung** — sonst weiß niemand, welche
	// gemeint war.
	if !strings.Contains(fehler.Error(), "999") {
		t.Errorf("die Meldung lautet %q", fehler)
	}
}

// **Ein leerer Bestand weist jede Kennung ab.**
func TestEinLeererBestandWeistJedeKennungAb(t *testing.T) {
	t.Parallel()

	if fehler := service.SetDaSein(t.Context(), setkoepfe{},
		1); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}
}

// **Ein Lesefehler ist kein „gibt es nicht".**
//
// Der Unterschied zählt: „die Platte klemmt" und „dieses Set gibt es nicht"
// verlangen verschiedene Antworten (§8).
func TestEinLesefehlerIstKeinFehlendesSet(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	fehler := service.SetDaSein(t.Context(), setkoepfe{fehler: gestolpert}, 1)

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}

	if errors.Is(fehler, domain.ErrNotFound) {
		t.Error("ein Lesefehler kam als »gibt es nicht« heraus")
	}
}

// **Die Wache greift an den Wegen, die sie bekommen haben** —
// #1336.
//
// # Warum sie wahlfrei ist
//
// Weil nicht jeder Aufrufer einen Setleser hat. Ein Pflichtargument zwänge
// jeden Test, einen zu bauen, um einen Verweis zu prüfen — und die Wege ohne
// Wache antworten wie bisher. **Beide Zweige stehen hier**, damit die Wahl
// keine Behauptung bleibt.
func TestDieSetwacheGreiftWoSieGesetztIst(t *testing.T) {
	t.Parallel()

	leser := setkoepfe{koepfe: []domain.SetHead{{ID: 1}}}

	t.Run("Verweise mit Wache", func(t *testing.T) {
		t.Parallel()

		verweise := service.NeueKartenverweise(nil, &verweisdoppel{}, probeuhr()).
			MitSetwache(leser)

		if _, fehler := verweise.JeSet(t.Context(),
			999); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
		}

		if _, fehler := verweise.JeSet(t.Context(), 1); fehler != nil {
			t.Errorf("ein vorhandenes Set gab %v", fehler)
		}
	})

	t.Run("Verweise ohne Wache", func(t *testing.T) {
		t.Parallel()

		if _, fehler := service.NeueKartenverweise(nil, &verweisdoppel{}, probeuhr()).
			JeSet(t.Context(), 999); fehler != nil {
			t.Errorf("ohne Wache gab %v", fehler)
		}
	})

	t.Run("Verlauf mit Wache", func(t *testing.T) {
		t.Parallel()

		verlauf := service.NeuerVerlaufsdienst(&verlaufsdouble{}).MitSetwache(leser)

		if _, fehler := verlauf.Zuruecknehmen(t.Context(),
			999, ""); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Zuruecknehmen: %v", fehler)
		}

		if _, fehler := verlauf.Wiederherstellen(t.Context(),
			999, ""); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Wiederherstellen: %v", fehler)
		}

		if _, fehler := verlauf.Zuruecknehmen(t.Context(), 1, ""); fehler != nil {
			t.Errorf("ein vorhandenes Set gab %v", fehler)
		}
	})

	t.Run("Verlauf ohne Wache", func(t *testing.T) {
		t.Parallel()

		verlauf := service.NeuerVerlaufsdienst(&verlaufsdouble{})

		if _, fehler := verlauf.Zuruecknehmen(t.Context(), 999, ""); fehler != nil {
			t.Errorf("ohne Wache gab %v", fehler)
		}

		if _, fehler := verlauf.Wiederherstellen(t.Context(), 999, ""); fehler != nil {
			t.Errorf("ohne Wache gab %v", fehler)
		}
	})
}
