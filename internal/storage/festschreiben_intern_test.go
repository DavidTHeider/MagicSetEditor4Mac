package storage

// **Vor dem Festschreiben sieht niemand die Wanderung** — M02
// (#2346).
// Die Naht [vorDemFestschreiben] ist der Augenblick, in dem die Hüllenprobe
// `wanderung09` den Prozess hart beendet; hier steht, was ein zweiter Leser
// in genau diesem Augenblick sieht: den Stand von vorher.

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestVorDemFestschreibenSiehtNiemandDieWanderung(t *testing.T) { //nolint:paralleltest // die Naht ist eine Paketvariable
	ort := filepath.Join(t.TempDir(), "bestand.db")

	alt, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = alt.Close() }) //nolint:errcheck // Aufräumen

	if fehler := migriereAusMitVorbereitung(probekontext(t), alt, nurGrundlinie, PruefeBestand, nil, grundlinie); fehler != nil {
		t.Fatal(fehler)
	}

	leser, err := Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = leser.Close() }) //nolint:errcheck // Aufräumen

	var gesehen []string

	vorDemFestschreiben = func() {
		gesehen, _ = AngewendeteMigrationen(probekontext(t), leser) //nolint:errcheck // unten verglichen
	}

	t.Cleanup(func() { vorDemFestschreiben = nil })

	if fehler := Migriere(probekontext(t), alt, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(gesehen, grundlinienumfangOder(t)) {
		t.Errorf("vor dem Festschreiben sah ein zweiter Leser %d Wanderungen, erwartet die %d von vorher",
			len(gesehen), len(grundlinienumfangOder(t)))
	}
}
