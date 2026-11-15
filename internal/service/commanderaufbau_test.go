package service_test

// **Woher eine Commanderpartie ihr Profil bekommt** — Werkstück M19
// (#1620).
//
// # Warum es zwei Wege gibt und beide gefahren werden
//
// **`PartieAnlegen` heißt an den drei Flächen ohne Fenster
// `(spiel, wirt, sitze)` und trägt kein Regelwerk.** Es dort nachzureichen
// wäre eine **brechende** Änderung ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md));
// deshalb wird das Commanderprofil **abgeleitet**, sobald ein Sitz einen
// Commander nennt. Das Fenster nennt es unmittelbar.
//
// **Zwei Wege heißen zwei Läufe.** Bis zum 2026-09-18 hatte keiner von
// beiden einen — der Zweig war grün, weil niemand ihn betreten hat (§10).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein genanntes Commanderwerk gibt das Commanderprofil** — CR 903.
func TestEinGenanntesCommanderwerkGibtDasProfil(t *testing.T) {
	t.Parallel()

	werk := domain.Regelwerk{
		Kennung: domain.Commanderkennung,
		Fassung: domain.Commanderfassung,
		Strenge: domain.StrengeSandkasten,
	}

	partie := commanderpartie(t, werk, "")

	if partie.Regelwerk.Kennung != domain.Commanderkennung {
		t.Errorf("Regelwerk %q", partie.Regelwerk.Kennung)
	}

	// **Vierzig Leben** — CR 903.7.
	if partie.Sitze[0].Leben != domain.Commanderleben {
		t.Errorf("Leben %d", partie.Sitze[0].Leben)
	}
}

// **Und ein Sitz, der einen Commander nennt, ebenso** — die Ableitung für
// die drei Flächen ohne Fenster.
func TestEinGenannterCommanderLeitetDasProfilAb(t *testing.T) {
	t.Parallel()

	partie := commanderpartie(t, einRegelwerk(), "wald")

	if partie.Regelwerk.Kennung != domain.Commanderkennung {
		t.Errorf("Regelwerk %q", partie.Regelwerk.Kennung)
	}

	if partie.Sitze[0].Leben != domain.Commanderleben {
		t.Errorf("Leben %d", partie.Sitze[0].Leben)
	}
}

// **Und ohne beides bleibt es der Sandkasten** — die Gegenprobe (§10).
func TestOhneCommanderBleibtEsDerSandkasten(t *testing.T) {
	t.Parallel()

	partie := commanderpartie(t, einRegelwerk(), "")

	if partie.Regelwerk.Kennung != "sandkasten" {
		t.Errorf("Regelwerk %q", partie.Regelwerk.Kennung)
	}

	if partie.Sitze[0].Leben != 20 {
		t.Errorf("Leben %d", partie.Sitze[0].Leben)
	}
}

// commanderpartie baut eine Partie mit dem genannten Werk und Commander.
func commanderpartie(
	t *testing.T, werk domain.Regelwerk, commander string,
) domain.Partie {
	t.Helper()

	partie, fehler := einPartieaufbau().Aufbauen(
		context.Background(), "p1", "magic", werk, "sitz-1",
		[]service.Sitzwunsch{
			{
				Kennung: "sitz-1", Name: "David", DeckID: 1,
				Commander: commander,
			},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	return partie
}
