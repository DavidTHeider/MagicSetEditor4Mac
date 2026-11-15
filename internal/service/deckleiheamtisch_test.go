package service_test

// **Eine Leihe kommt an den Tisch ihres Verleihers zurück** — P9b
// (#2377).
// Der Gast nennt nur die Kennung; der Wirt nimmt den Abzug, den **er**
// ausgestellt hat, und prüft, dass er ihn **diesem** Gast für **diese**
// Partie geliehen hat. Die Kartenkennungen zeigen dann in seinen eigenen
// Bestand, und er zeichnet die Bilder selbst.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineAusgestellteLeihe lässt Anna Bert ein Deck für die Partie p1 leihen.
func eineAusgestellteLeihe(t *testing.T) (*service.Leihstelle, *leihprobe, domain.Deckleihe, peer.Identitaet) {
	t.Helper()

	anna, bert := zweiPeers(t)
	stelle, ablage, _ := eineLeihstelle(t, anna)

	leihe, _, fehler := stelle.Anbieten(t.Context(), service.Leihwunsch{
		Entleiher: string(bert.Kennung()), DeckID: 7,
		Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		Rechte:  []domain.Leihrecht{domain.RechtSpielen},
		Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
	})
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	return stelle, ablage, leihe, bert
}

func TestDerVerleiherNimmtSeinenEigenenAbzug(t *testing.T) {
	t.Parallel()

	stelle, _, leihe, bert := eineAusgestellteLeihe(t)

	abzug, fehler := stelle.VerliehenesDeck(t.Context(), leihe.Kennung, string(bert.Kennung()), "p1")
	if fehler != nil {
		t.Fatalf("VerliehenesDeck: %v", fehler)
	}

	if abzug.Abdruck != leihe.Abdruck || abzug.Quelle != "" || len(abzug.Karten) == 0 {
		t.Errorf("der Abzug: %+v", abzug)
	}
}

func TestEineLeiheGiltNurIhremEntleiherAnIhrerPartie(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		richte    func(*testing.T, *leihprobe, domain.Deckleihe)
		entleiher func(peer.Identitaet) string
		partie    string
		art       error
	}{
		"ein anderer Gast":   {nil, func(peer.Identitaet) string { return "peer-cora" }, "p1", domain.ErrValidation},
		"eine andere Partie": {nil, func(b peer.Identitaet) string { return string(b.Kennung()) }, "p2", domain.ErrValidation},
		"zurückgenommen": {func(t *testing.T, ablage *leihprobe, leihe domain.Deckleihe) {
			t.Helper()

			stand := ablage.leihen[leihe.Kennung]
			stand.Widerrufen = true
			ablage.leihen[leihe.Kennung] = stand
		}, func(b peer.Identitaet) string { return string(b.Kennung()) }, "p1", domain.ErrConflict},
		"von jemand anderem ausgestellt": {func(t *testing.T, ablage *leihprobe, leihe domain.Deckleihe) {
			t.Helper()

			stand := ablage.leihen[leihe.Kennung]
			stand.Verliehen = false
			ablage.leihen[leihe.Kennung] = stand
		}, func(b peer.Identitaet) string { return string(b.Kennung()) }, "p1", domain.ErrConflict},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, ablage, leihe, bert := eineAusgestellteLeihe(t)
			if fall.richte != nil {
				fall.richte(t, ablage, leihe)
			}

			if _, fehler := stelle.VerliehenesDeck(t.Context(), leihe.Kennung, fall.entleiher(bert), fall.partie); !errors.Is(fehler, fall.art) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

func TestEineUnbekannteLeiheGibtEsNicht(t *testing.T) {
	t.Parallel()

	stelle, _, _, bert := eineAusgestellteLeihe(t)

	if _, fehler := stelle.VerliehenesDeck(t.Context(), "weg", string(bert.Kennung()), "p1"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}
