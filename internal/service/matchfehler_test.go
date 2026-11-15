package service_test

// **Die Fehlerwege der Matchleitung** — Werkstück G7 von Phase 5.5
// (#1464).
//
// **Jeder betreten** (§3): ein Weg, den kein Lauf geht, ist eine Zusage ohne
// Prüfung — und zum Ergebnis gehört das Nein (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Turnier, das es nicht gibt, wird auf beiden Wegen gemeldet.**
func TestEinFehlendesTurnierWirdGemeldet(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 99, 7, "p1", zweiSitze()); fehler == nil {
		t.Error("binden: das fehlende Turnier kam durch")
	}

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2})

	if _, fehler := leitung.Uebernehmen(t.Context(), 99, 7); fehler == nil {
		t.Error("übernehmen: das fehlende Turnier kam durch")
	}
}

// **Eine klagende Ablage wird auf jedem Weg gemeldet.**
func TestEineKlagendeMatchablageWirdGemeldet(t *testing.T) {
	t.Parallel()

	leitung, _, match := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	match.klagt = errProbe

	wege := map[string]func() error{
		"binden": func() error {
			_, fehler := leitung.Binden(t.Context(), 1, 7, "p1", zweiSitze())

			return fehler
		},
		"melden": func() error {
			return leitung.Melden(t.Context(), domain.Matchmeldung{
				PaarungID: 7, Fassung: 3, PartieID: "p1", Von: "anna",
				Kette: "glied", Ergebnis: domain.Ergebnis{SpieleA: 2},
			})
		},
		"lesen": func() error {
			_, fehler := leitung.Bindung(t.Context(), 7)

			return fehler
		},
		"Meldungen lesen": func() error {
			_, fehler := leitung.Meldungen(t.Context(), 7)

			return fehler
		},
		"übernehmen": func() error {
			_, fehler := leitung.Uebernehmen(t.Context(), 1, 7)

			return fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); !errors.Is(fehler, errProbe) {
				t.Fatalf("erwartet die Klage, bekam %v", fehler)
			}
		})
	}
}

// **Eine Meldung ohne Bindung wird abgewiesen** — und eine, die nicht passt,
// auch.
func TestEineMeldungOhneBindungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	if fehler := leitung.Melden(t.Context(), domain.Matchmeldung{
		PaarungID: 7, Fassung: 3, PartieID: "p1", Von: "anna",
		Kette: "glied", Ergebnis: domain.Ergebnis{SpieleA: 2},
	}); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	if fehler := leitung.Melden(t.Context(), domain.Matchmeldung{
		PaarungID: 7, Fassung: 3, PartieID: "p1", Von: "cem",
		Kette: "glied", Ergebnis: domain.Ergebnis{SpieleA: 2},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine Partie, deren Strenge das Regelwerk nicht trägt, wird nicht
// gebunden.**
//
// **Die Unterstützungsmatrix aus G4 bestimmt den Modus** — das Issue verlangt
// es. Geprüft wird sie von [service.Regelfreigabe] und nicht ein zweites Mal
// in der Matchleitung (§10).
func TestEineUngedeckteStrengeWirdNichtGebunden(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtischMitStrenge(t, domain.StrengeStreng)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Und eine Partie mit einem Profil, das sich selbst widerspricht, ist ein
// Fehler und kein Nein.**
func TestEinKrummesProfilWirdGemeldet(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtischOhneProfil(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler == nil {
		t.Fatal("das krumme Profil kam durch")
	}
}

// **Eine inzwischen geänderte Paarung wird abgewiesen.**
func TestEineGeaenderteFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2})

	// **Jemand legt die Runde neu** — die Fassung steigt.
	turniere.turniere[1].Fassung = 4

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}

// **Ein von Hand eingetragenes, abweichendes Ergebnis wird nicht
// überschrieben.**
func TestEinAbweichendesErgebnisWirdNichtUeberschrieben(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2, SpieleB: 1})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2, SpieleB: 1})

	turniere.turniere[1].Runden[0].Paarungen[0].Ergebnis = &domain.Ergebnis{SpieleA: 1, SpieleB: 2}

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}

// **Ein Turnierspeicher, der beim Eintragen klagt, wird gemeldet.**
func TestEinKlagendesEintragenWirdGemeldet(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2})

	turniere.fehler = errProbe

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, errProbe) {
		t.Fatalf("erwartet die Klage, bekam %v", fehler)
	}
}

// **Eine Bindung mit einem Sitz wird abgewiesen** — auch wenn der Sitz und
// der Teilnehmer beide stimmen.
//
// **Das ist der Weg, auf dem [domain.Matchbindung.Pruefe] greift**: `sitzeAus`
// gibt genau so viele Sitze zurück, wie hereinkamen, und eine Paarung hat
// zwei Seiten.
func TestEineBindungMitEinemSitzWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(t.Context(), 1, 7, "p1",
		[]service.Sitzzuordnung{{TeilnehmerID: 10, Sitz: "anna"}}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine klagende Meldungsliste hält die Übernahme auf.**
func TestEineKlagendeMeldungslisteHaeltDieUebernahmeAuf(t *testing.T) {
	t.Parallel()

	leitung, _, match := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	match.meldeklage = errProbe

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, errProbe) {
		t.Fatalf("erwartet die Klage, bekam %v", fehler)
	}
}

// **Eine Paarung, die es nicht mehr gibt, wird bei der Übernahme gemeldet.**
//
// **Das geschieht, wenn jemand die Runde neu legt.** Die Bindung steht dann
// noch da und zeigt auf eine Begegnung, die es so nicht mehr gibt — und das
// ist eine Meldung und keine stille Übernahme.
func TestEineVerschwundenePaarungWirdGemeldet(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2})

	turniere.turniere[1].Runden = nil

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Ein klagendes Eintragen wird gemeldet** — und zwar erst dort.
func TestEinKlagendesEintragenKommtDurch(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	melde(t, leitung, "anna", "glied", domain.Ergebnis{SpieleA: 2})
	melde(t, leitung, "bert", "glied", domain.Ergebnis{SpieleA: 2})

	turniere.setzfehler = errProbe

	if _, fehler := leitung.Uebernehmen(t.Context(), 1, 7); !errors.Is(
		fehler, errProbe) {
		t.Fatalf("erwartet die Klage, bekam %v", fehler)
	}
}
