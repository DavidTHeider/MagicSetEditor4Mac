package service_test

// **Die Ratengrenze der Einlösung** — NET-01b.AK-02
// (#1954):
// „Eine Ratengrenze je Einladung und je Absender hält Fluten ab, und ihr
// Greifen ist ein benannter Fehler."

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineGrenze gibt eine Grenze mit einer Uhr, die nur auf Zuruf geht.
func eineGrenze() (grenze *service.Einloesegrenze, weiter func(time.Duration)) {
	jetzt := einladungsuhr

	return service.NeueEinloesegrenze(func() time.Time { return jetzt }),
		func(dauer time.Duration) { jetzt = jetzt.Add(dauer) }
}

func TestJeAbsenderGreiftDieGrenze(t *testing.T) {
	t.Parallel()

	grenze, _ := eineGrenze()

	for versuch := range service.EinloesungenJeAbsender {
		if fehler := grenze.Versuch("bert", fmt.Sprintf("n%d", versuch)); fehler != nil {
			t.Fatalf("Versuch %d: %v", versuch, fehler)
		}
	}

	if fehler := grenze.Versuch("bert", "noch-eine"); !errors.Is(fehler, domain.ErrEinloesenGedrosselt) {
		t.Errorf("über der Grenze: %v", fehler)
	}

	// **Ein anderer Absender ist davon unberührt.**
	if fehler := grenze.Versuch("carla", "n0-carla"); fehler != nil {
		t.Errorf("ein anderer: %v", fehler)
	}
}

func TestJeEinladungGreiftDieGrenze(t *testing.T) {
	t.Parallel()

	grenze, _ := eineGrenze()

	// **Viele Kennungen, eine Einladung** — wer Schlüssel würfelt, soll an
	// einer bekannten Einladung nicht beliebig oft rütteln.
	for versuch := range service.EinloesungenJeEinladung {
		if fehler := grenze.Versuch(fmt.Sprintf("wuerfel-%d", versuch), "n1"); fehler != nil {
			t.Fatalf("Versuch %d: %v", versuch, fehler)
		}
	}

	if fehler := grenze.Versuch("wuerfel-neu", "n1"); !errors.Is(fehler, domain.ErrEinloesenGedrosselt) {
		t.Errorf("über der Grenze: %v", fehler)
	}
}

// **Und insgesamt** — sonst wüchse die Liste mit jeder gewürfelten Kennung
// und jeder erfundenen Nonce, und ein Fremder bestimmte den Speicher.
func TestInsgesamtGreiftDieGrenze(t *testing.T) {
	t.Parallel()

	grenze, _ := eineGrenze()

	for versuch := range service.EinloesungenInsgesamt {
		if fehler := grenze.Versuch(fmt.Sprintf("k%d", versuch), fmt.Sprintf("n%d", versuch)); fehler != nil {
			t.Fatalf("Versuch %d: %v", versuch, fehler)
		}
	}

	if fehler := grenze.Versuch("k-neu", "n-neu"); !errors.Is(fehler, domain.ErrEinloesenGedrosselt) {
		t.Errorf("über der Grenze: %v", fehler)
	}
}

// **Die Grenze vergeht mit ihrem Fenster.**
func TestNachDemFensterGehtEsWieder(t *testing.T) {
	t.Parallel()

	grenze, weiter := eineGrenze()

	for versuch := range service.EinloesungenJeAbsender + 1 {
		_ = grenze.Versuch("bert", fmt.Sprintf("n%d", versuch)) //nolint:errcheck // gemessen wird danach
	}

	weiter(service.Einloesefenster)

	if fehler := grenze.Versuch("bert", "spaeter"); fehler != nil {
		t.Errorf("nach dem Fenster: %v", fehler)
	}
}
