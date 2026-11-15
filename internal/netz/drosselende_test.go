package netz_test

// **Bis wann eine Drossel gesperrt ist** — für einen Aufrufer, der es
// sagen muss (Nachlauf über Phase 0, Befund 2, #2333).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestDasEndeDerSperreIstDieLaengere(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 13, 14, 0, 0, 0, time.UTC)
	drossel := netz.NeueDrossel(time.Second, func() time.Time { return jetzt }, nil)

	if bis := drossel.GesperrtBis(); !bis.IsZero() {
		t.Errorf("ungesperrt endet nichts, gemeldet %v", bis)
	}

	drossel.Sperre(jetzt.Add(time.Minute))
	drossel.Sperre(jetzt.Add(10 * time.Second))

	if bis := drossel.GesperrtBis(); !bis.Equal(jetzt.Add(time.Minute)) {
		t.Errorf("Ende %v, erwartet die längere Sperre %v", bis, jetzt.Add(time.Minute))
	}
}

func TestEineAbgelaufeneSperreEndetNicht(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 13, 14, 0, 0, 0, time.UTC)
	drossel := netz.NeueDrossel(time.Second, func() time.Time { return jetzt }, nil)

	drossel.Sperre(jetzt.Add(-time.Second))

	if bis := drossel.GesperrtBis(); !bis.IsZero() {
		t.Errorf("eine vergangene Sperre meldet ein Ende: %v", bis)
	}
}

func TestEineSperreDerGegenstelleTrifftDasGanzeWerk(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 13, 14, 0, 0, 0, time.UTC)
	werk := netz.NeuesDrosselwerk(func() time.Time { return jetzt }, func(time.Duration) {})
	eine, andere := werk.Fuer("freigaben", time.Second), werk.Fuer("beigaben", time.Second)

	bis := jetzt.Add(37 * time.Minute)
	eine.SperreGegenstelle(bis)

	if ende := andere.GesperrtBis(); !ende.Equal(bis) {
		t.Errorf("die andere Drossel endet %v, erwartet genau %v", ende, bis)
	}
}

func TestOhneWerkSperrtDieGegenstelleNurDieEine(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 13, 14, 0, 0, 0, time.UTC)
	drossel := netz.NeueDrossel(time.Second, func() time.Time { return jetzt }, nil)

	drossel.SperreGegenstelle(jetzt.Add(time.Minute))

	if ende := drossel.GesperrtBis(); !ende.Equal(jetzt.Add(time.Minute)) {
		t.Errorf("Ende %v", ende)
	}
}
