package mcpnaehte

// **Die Naht der Problemprüfung** — PERF02
// (#2352):
// vor der Prüfung ungeprüft, danach aktuell, ein fremdes Set ein Fehler.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieProblemnahtPrueftUndNenntDenStand(t *testing.T) {
	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-stil",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Eine"}}}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	naht := probleme{db: db}

	vorher, err := naht.Problemstand(probekontext(t), setID)
	if err != nil || vorher.Urteil != "ungeprueft" || vorher.Karten == nil {
		t.Fatalf("vorher %+v, %v", vorher, err)
	}

	nachher, err := naht.ProblemePruefen(probekontext(t), setID)
	if err != nil || nachher.Urteil != "aktuell" || !nachher.Sicher || nachher.Gesamt != 1 {
		t.Errorf("nachher %+v, %v", nachher, err)
	}

	if _, err := naht.ProblemePruefen(probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("fremdes Set: %v", err)
	}
}

func TestDieProblemnahtBrauchtEinZuhause(t *testing.T) {
	db := frischerBestand(t)
	t.Setenv("HOME", "")

	naht := probleme{db: db}

	if _, err := naht.Problemstand(probekontext(t), 1); err == nil {
		t.Error("ohne Zuhause gelesen")
	}

	if _, err := naht.ProblemePruefen(probekontext(t), 1); err == nil {
		t.Error("ohne Zuhause geprüft")
	}
}
