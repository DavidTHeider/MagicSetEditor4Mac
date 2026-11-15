package domain_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieDauerEinesAuftragsHaengtAnSeinemStand(t *testing.T) {
	t.Parallel()

	anfang := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	laufend := domain.Auftrag{Stand: domain.AuftragLaeuft, Begonnen: anfang}

	if dauer := laufend.Dauer(anfang.Add(90 * time.Second)); dauer != 90*time.Second {
		t.Errorf("ein laufender Auftrag dauert %s", dauer)
	}

	fertig := domain.Auftrag{Stand: domain.AuftragFertig, Begonnen: anfang, Beendet: anfang.Add(time.Minute)}
	if dauer := fertig.Dauer(anfang.Add(time.Hour)); dauer != time.Minute {
		t.Errorf("ein fertiger Auftrag dauert %s — die Uhr danach zählt nicht", dauer)
	}
}
