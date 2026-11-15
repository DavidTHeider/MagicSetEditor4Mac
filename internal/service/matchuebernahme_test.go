package service_test

// **Ein Ergebnis übernehmen — einmal** — Werkstück G7 von Phase 5.5
// (#1464).
//
// **Neben `matchleitung_test.go` und nicht darin**: dort wird **gebunden**,
// hier wird **übernommen**. Zwei Fragen, zwei Dateien — und die erste stand
// mit beiden bei 426 Zeilen (§2).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Ergebnis wird einmal übernommen — und die Wiederholung ändert
// nichts.**
func TestEinErgebnisWirdEinmalUebernommen(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "letztesglied", domain.Ergebnis{SpieleA: 2, SpieleB: 1})

	// **Eine Seite genügt nicht.**
	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); fehler == nil {
		t.Fatal("eine Seite hat gereicht")
	}

	melde(t, leitung, "bert", "letztesglied", domain.Ergebnis{SpieleA: 2, SpieleB: 1})

	stand, fehler := leitung.Uebernehmen(t.Context(), 1, 7)
	if fehler != nil {
		t.Fatalf("übernehmen: %v", fehler)
	}

	if stand.SpieleA != 2 || stand.SpieleB != 1 {
		t.Fatalf("der Stand ist %+v", stand)
	}

	if turniere.gesetzt != 1 {
		t.Fatalf("es wurde %dmal eingetragen", turniere.gesetzt)
	}

	// **Und die Wiederholung trägt nichts zweites ein.**
	//
	// **Die Fassung ist inzwischen gestiegen** — das Eintragen erhöht sie.
	// Eine Übernahme, die daran scheiterte, machte aus „schon erledigt" einen
	// Fehler; deshalb wird die Bindung mitgeführt.
	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); fehler != nil {
		t.Fatalf("die zweite Übernahme: %v", fehler)
	}

	if turniere.gesetzt != 1 {
		t.Fatalf("es wurde %dmal eingetragen", turniere.gesetzt)
	}
}

// melde trägt die Meldung einer Seite ein.
func melde(
	t *testing.T, leitung *service.Matchleitung, von, kette string,
	stand domain.Ergebnis,
) {
	t.Helper()

	if fehler := leitung.Melden(t.Context(), domain.Matchmeldung{
		PaarungID: 7, Fassung: 3, PartieID: "p1",
		Ergebnis: stand, Kette: kette, Von: von,
	}); fehler != nil {
		t.Fatalf("melden (%s): %v", von, fehler)
	}
}

// **Abweichende Meldungen erzeugen Konflikt statt Überschreiben.**
func TestAbweichendeMeldungenErzeugenKonflikt(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2, SpieleB: 1})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 1, SpieleB: 2})

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}

	if turniere.gesetzt != 0 {
		t.Fatal("es wurde trotzdem eingetragen")
	}
}

// einMatchabzug gibt einen kleinsten spielbaren Abzug.
func einMatchabzug(abdruck string) domain.Partiedeck {
	return domain.Partiedeck{
		Abdruck: abdruck,
		Karten: []domain.Abzugskarte{{
			Kennung: "k1", Name: "Wald", Menge: 40,
			Seiten: []domain.Abzugsseite{{Name: "Wald"}},
		}},
		Beigaben: []domain.Beigabe{},
	}
}
