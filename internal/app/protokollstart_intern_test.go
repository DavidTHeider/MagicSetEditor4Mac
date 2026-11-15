package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestDieStartweicheDrehtNicht hält Nachlauf über Phase 6.4, N2: die Weiche
// entsteht vor dem Bestand mit der Vorgaberegel — und drehte schon damit, mit
// 4 MB und **einer** Vorfassung. Eine eingestellte Grenze über 4 MB wirkte
// deshalb nie, und eine Aufbewahrung über 1 auch nicht: die Vorgabe
// überschrieb `mse.log.1`, bevor die Einstellung galt.
func TestDieStartweicheDrehtNicht(t *testing.T) {
	ordner := t.TempDir()
	pfad := filepath.Join(ordner, Protokollname)

	for name, groesse := range map[string]int{Protokollname: 5 << 20, Protokollname + ".1": 10} {
		if fehler := os.WriteFile(filepath.Join(ordner, name), make([]byte, groesse), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	weiche, _, _ := weicheMitPuffern(t, ordner)

	if angabe, fehler := os.Stat(pfad); fehler != nil || angabe.Size() < 5<<20 {
		t.Fatalf("die Startweiche hat gedreht: %v", fehler)
	}

	// **Eine Grenze über der Vorgabe gilt**: 5 MB bleiben unter 8.
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollGrenze: "8"}))

	if angabe, fehler := os.Stat(pfad); fehler != nil || angabe.Size() < 5<<20 {
		t.Fatalf("gedreht unter der eingestellten Grenze: %v", fehler)
	}

	// **Und die Aufbewahrung**: die alte Vorfassung rückt weiter und wird
	// nicht überschrieben.
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
		domain.ProtokollGrenze: "1", domain.ProtokollAufbewahrung: "3",
	}))

	if angabe, fehler := os.Stat(vorfassung(pfad, 2)); fehler != nil || angabe.Size() != 10 {
		t.Errorf("die alte Vorfassung ist nicht weitergerückt: %v", fehler)
	}
}
