//go:build darwin

// **Die Nähte der Verträglichkeitsprüfung** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Gegen echte Dinge**: einen frisch migrierten Bestand, ein echtes
// Volumen, ein echtes Auftragsbuch. Ein Doppel sagte hier nur, dass die Naht
// eine Naht ist.
package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Ledger nennt die Wanderungen dieses Bestands — mit Prüfsummen.**
func TestDerLedgerNenntDieVerbuchtenWanderungen(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	stand, fehler := app.LedgerUeber(db).Ledger(t.Context())
	if fehler != nil {
		t.Fatalf("Ledger: %v", fehler)
	}

	if len(stand) == 0 {
		t.Fatal("ein migrierter Bestand hat keine Wanderung verbucht")
	}

	for name, summe := range stand {
		switch {
		case !strings.HasSuffix(name, ".sql"):
			t.Errorf("%q sieht nicht wie eine Wanderung aus", name)
		case len(summe) != 64:
			t.Errorf("%q trägt die Summe %q — keine SHA-256", name, summe)
		}
	}
}

// **Der freie Platz ist eine Zahl über null** — und die des Volumens, auf dem
// der Bestand liegt, nicht die des Systemvolumens.
func TestDerFreiePlatzWirdGemessen(t *testing.T) {
	t.Parallel()

	frei, fehler := app.PlatzNeben(t.TempDir() + "/bestand.db").FreierPlatz()
	if fehler != nil {
		t.Fatalf("messen: %v", fehler)
	}

	if frei <= 0 {
		t.Errorf("freier Platz = %d", frei)
	}
}

// **Ein Ort, den es nicht gibt, meldet sich** — eine erfundene Zahl ließe
// eine Vorprüfung durchgehen, die nichts geprüft hat.
func TestEinOrtOhneVolumenMeldetSich(t *testing.T) {
	t.Parallel()

	if _, fehler := app.PlatzNeben(
		"/gibtesnicht/und/wird/es/nie/bestand.db").FreierPlatz(); fehler == nil {
		t.Error("ein Ort ohne Volumen gab eine Zahl")
	}
}

// **Und der Weg für die Flächen ohne Fenster steht.**
func TestDieNahtFuerDieFlaechenAntwortet(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	// **Ohne eingeschaltete Aktualisierung gibt es kein Urteil** — und
	// keinen Netzverkehr. Das ist die Vorgabe, und ein Test darf ohnehin
	// nicht nach draußen (§3).
	// **Mit gezeigtem Profil im Kontext**, wie es die drei Wurzeln ohne
	// Fenster tun (`storage.MitGezeigtemProfil`): ohne es fragt der
	// Einstellungsspeicher zurück, wessen Einstellungen gemeint sind.
	antwort := app.VertraeglichkeitUeber(db).Vertraeglichkeit(probekontext(t))

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Urteil != "":
		t.Errorf("Urteil = %q — es hätte gefragt werden müssen", antwort.Urteil)
	case antwort.Hindernisse == nil:
		t.Error("die Hindernisse kamen als nil")
	}

	_ = domain.UrteilUngeklaert
}
