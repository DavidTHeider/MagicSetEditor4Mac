package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Welchem Profil ein Token gehört** — die Frage, die eine Bindung vor jedem
// Ruf mit einer Tokenkennung stellt
// (#1817):
// die Rolle zählt an **diesem** Profil, nicht am gezeigten.
func TestDasProfilEinesTokens(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung := service.NeueTokenverwaltung(ablage)

	angelegt, _, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if profil, fehler := verwaltung.ProfilDes(context.Background(), angelegt.ID); fehler != nil || profil != 7 {
		t.Errorf("das Profil des Tokens: %d (%v), erwartet 7", profil, fehler)
	}

	if _, fehler := verwaltung.ProfilDes(context.Background(), 999); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein Token, das es nicht gibt: %v", fehler)
	}
}
