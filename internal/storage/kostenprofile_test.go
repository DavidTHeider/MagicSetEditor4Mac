package storage_test

// **Dieselbe Auftragskennung in zwei Profilen sind zwei Aufträge** — Befund 4
// des Nachlaufs über Phase 6.1.
//
// Die Kennung eines Postens gibt der Aufrufer vor, und nicht jede ist
// gewürfelt: die Probe eines Modells heißt `probe/<anbieter>/<alias>` — in
// jedem Profil gleich. Mit der Kennung allein als Schlüssel hielt das zweite
// Profil die Probe für einen fremden Auftrag und wies sie ab, und ein
// Abschluss ging über die Kennung allein, ohne nach dem Profil zu fragen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieselbeKennungInZweiProfilenSindZweiAuftraege(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Zweites')`); fehler != nil {
		t.Fatalf("das zweite Profil: %v", fehler)
	}

	journal := storage.NeuesKostenjournal(db)

	for _, profil := range []int64{1, 2} {
		if fehler := journal.BudgetSetzen(ctx, profil, "anbieter",
			domain.ZeitraumTag, euro(1000), probezeit); fehler != nil {
			t.Fatalf("das Budget von %d: %v", profil, fehler)
		}
	}

	probe := auftrag("probe/anbieter/modell", 100)

	if _, fehler := journal.Reservieren(ctx, probe, domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("im ersten Profil: %v", fehler)
	}

	probe.ProfilID = 2

	if _, fehler := journal.Reservieren(ctx, probe, domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("im zweiten Profil: %v — die Probe desselben Modells gilt dort als fremder Auftrag", fehler)
	}

	if _, fehler := journal.Abschliessen(ctx, 2, probe.Kennung, euro(40), probezeit); fehler != nil {
		t.Fatalf("abschließen im zweiten: %v", fehler)
	}

	// **Das erste Profil bleibt, wie es war** — sein Posten ist reserviert.
	posten, fehler := journal.Posten(ctx, 1, "anbieter", domain.ZeitraumTag, probezeit, 10)
	if fehler != nil {
		t.Fatal(err(fehler))
	}

	if len(posten) != 1 || posten[0].Stand != domain.KostenReserviert {
		t.Errorf("im ersten Profil: %+v — der Abschluss des zweiten traf ihn", posten)
	}

	// **Und ein Abschluss aus dem falschen Profil trifft nichts.**
	if _, fehler := journal.Freigeben(ctx, 3, probe.Kennung, probezeit); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("aus einem dritten Profil: %v, erwartet nicht gefunden", fehler)
	}
}

func err(fehler error) string { return fehler.Error() }
