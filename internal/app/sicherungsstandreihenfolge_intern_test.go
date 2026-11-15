package app

// **Stand und Plan stimmen zusammen** — gefunden, als `make ci` unter Last
// `TestDerPlanSichertGleichNachDerOrdnerwahl` rot sah: der Plan nannte eine
// letzte Sicherung, die Stufe stand auf „unbekannt". [sicherungsstandUeber]
// las erst den Stand und dann den Planvermerk; sicherte der Plan genau
// dazwischen, sah die zweite Lesung, was die erste nicht sah.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nachziehendesBuch zeigt beim ersten Lesen den Stand davor und vermerkt
// danach — wie ein Planlauf, der zwischen zwei Lesungen fertig wird.
type nachziehendesBuch struct {
	vermerke []domain.Sicherungsvermerk
	spaet    domain.Sicherungsvermerk
	gelesen  bool
}

func (b *nachziehendesBuch) nachziehen() {
	if !b.gelesen {
		b.gelesen = true
		b.vermerke = append([]domain.Sicherungsvermerk{b.spaet}, b.vermerke...)
	}
}

func (b *nachziehendesBuch) Vermerken(context.Context, domain.Sicherungsvermerk) error { return nil }

func (b *nachziehendesBuch) Letzter(context.Context) (domain.Sicherungsvermerk, bool, error) {
	vorher := b.vermerke
	b.nachziehen()

	if len(vorher) == 0 {
		return domain.Sicherungsvermerk{}, false, nil
	}

	return vorher[0], true, nil
}

func (b *nachziehendesBuch) Vermerke(context.Context, domain.Sicherungsanlass) ([]domain.Sicherungsvermerk, error) {
	vorher := b.vermerke
	b.nachziehen()

	return vorher, nil
}

// nurPlanordner nennt den Ordner des Plans und sonst nichts.
type nurPlanordner struct {
	Einstellungsfuehrer
	ordner string
}

func (n nurPlanordner) Alle(context.Context) (map[string]string, error) {
	return map[string]string{domain.SicherungsplanOrdner: n.ordner}, nil
}

func TestStandUndPlanStimmenZusammen(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	jetzt := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	name := service.Sicherungsplanname(jetzt)

	if err := os.WriteFile(filepath.Join(ordner, name), []byte("sicherung"), 0o600); err != nil {
		t.Fatal(err)
	}

	weg := service.NeuerBestandsweg(service.Sicherungsorte{}, nil, nil, nil, nil, nil, "")
	weg.MitSicherungsbuch(&nachziehendesBuch{spaet: domain.Sicherungsvermerk{
		Gesichert: jetzt, Anlass: domain.SicherungNachPlan, Flaeche: domain.FlaecheFenster,
		Datei: name, Groesse: int64(len("sicherung")),
	}})

	stand, err := sicherungsstandUeber(t.Context(), weg, nurPlanordner{ordner: ordner}, jetzt)
	if err != nil {
		t.Fatal(err)
	}

	if stand.Plan.Letzter != "" && stand.Stufe == string(domain.SicherungsstufeUnbekannt) {
		t.Errorf("der Plan nennt %s, die Stufe ist %q", stand.Plan.Letzter, stand.Stufe)
	}
}
