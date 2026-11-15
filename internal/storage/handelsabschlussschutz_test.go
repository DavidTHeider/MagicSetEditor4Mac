package storage_test

// **Was ein Abschluss abweist** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum das Nein eine eigene Datei bekommt
//
// **Weil es die Hälfte des Ergebnisses ist** (§10). Ein Abschluss, der den
// glücklichen Weg geht, ist geprüft, sobald er einmal gelaufen ist; ob er das
// Falsche abweist, sagt kein einziger grüner Lauf — und genau daran hängt,
// dass niemand zweimal abbucht.
//
// **Und jedes Nein hat hier zwei Hälften**: der Fehler kommt, **und** der
// Bestand steht unverändert da. Ein abgewiesener Abschluss, der die Hälfte
// gebucht hat, ist schlimmer als einer, der durchläuft.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// nichtsGeschehen sichert zu, dass der Vorzustand vollständig dasteht.
func (p abschlussprobe) nichtsGeschehen(t *testing.T, exemplar *domain.Exemplar) {
	t.Helper()

	ctx := probekontext(t)

	if menge, da := p.menge(t, exemplar.ID); !da || menge != 4 {
		t.Errorf("es liegen %d Stück da (da: %v) statt 4", menge, da)
	}

	zugegangen, fehler := p.bestand.Fuer(ctx, "shivan-lea")
	if fehler != nil {
		t.Fatalf("für den Druck: %v", fehler)
	}

	if len(zugegangen) != 0 {
		t.Errorf("es kamen %d Zeilen zu — der Abschluss hat halb gebucht",
			len(zugegangen))
	}

	zeilen, fehler := p.buch.Journal(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 0 {
		t.Errorf("das Journal führt %d Zeilen", len(zeilen))
	}

	beleg, fehler := p.buch.Beleg(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if beleg != nil {
		t.Error("es steht ein Beleg da")
	}
}

func TestOhneBestaetigteUebergabeWirdNichtsGebucht(t *testing.T) {
	probe := eineAbschlussprobe(t)
	exemplar := probe.vierStueck(t, false)

	fehler := probe.buch.Abschliessen(probekontext(t), einAbschluss(exemplar))
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Abschluss gab %v statt eines Konflikts", fehler)
	}

	probe.nichtsGeschehen(t, exemplar)
}

// **Eine Zusage rechnet gegen einen Stand** (§7, Regel 2). Hat sich die
// Besitzzeile seither geändert, wird nicht gebucht, sondern gefragt.
func TestEineVeralteteZusageBuchtNicht(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	geaendert := *exemplar
	geaendert.Notiz = "jemand hat daran gearbeitet"

	if fehler := probe.bestand.Aendern(ctx, &geaendert); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	fehler := probe.buch.Abschliessen(ctx, einAbschluss(exemplar))
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Abschluss gab %v statt eines Konflikts", fehler)
	}

	probe.nichtsGeschehen(t, exemplar)
}

// **Mehr als dasteht geht nicht** — auch nicht, wenn eine Zusage es deckte.
func TestMehrAlsDaIstGehtNichtAb(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)
	abschluss.Beleg.Zeilen[0].Menge = 9

	// **Die Deckung wird mitgehoben**, damit hier wirklich der Vorrat
	// entscheidet und nicht die Zusage (§10: eine Gegenprobe, die nichts
	// ändert, ist keine).
	if fehler := probe.zusagen.Festhalten(ctx, domain.Reservierung{
		Vorgang: einHandel, Revision: 1, ExemplarID: exemplar.ID, Menge: 2,
		Fassung: 1, Frist: time.Now().Add(time.Hour), Uebergeben: true,
	}, ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("zweite Zusage: %v", fehler)
	}

	fehler := probe.buch.Abschliessen(ctx, abschluss)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Abschluss gab %v statt eines Konflikts", fehler)
	}

	probe.nichtsGeschehen(t, exemplar)
}

// **Ein Fehler in der zweiten Zeile lässt die erste nicht stehen** — das ist
// die Fehlerinjektion, die das Issue verlangt.
func TestEineHalbeBuchungBleibtNichtStehen(t *testing.T) {
	probe := eineAbschlussprobe(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)

	// **Der Abgang geht durch, der Zugang nicht**: ein Erhaltungszustand,
	// den es nicht gibt. Er kommt durch [domain.Journalzeile.Pruefe] — die
	// fragt, ob die Zeile ins Journal darf — und fällt erst an
	// [domain.Exemplar.Validate], also **innerhalb** der Transaktion und
	// nach dem Abgang. Wer ohne Transaktion buchte, hätte danach zwei Stück
	// weniger und nichts dafür.
	abschluss.Beleg.Zeilen[1].Zustand = "zerknittert"

	fehler := probe.buch.Abschliessen(probekontext(t), abschluss)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Abschluss gab %v statt einer Beanstandung", fehler)
	}

	probe.nichtsGeschehen(t, exemplar)
}

// **Was ohne Datenbank zu entscheiden ist, wird ohne sie entschieden.**
func TestEinUnstimmigerAbschlussKommtNichtInDieTransaktion(t *testing.T) {
	faelle := map[string]struct {
		verderben func(*domain.Handelsabschluss)
		erwartet  error
	}{
		"fremder Vorgang": {
			func(a *domain.Handelsabschluss) { a.Beleg.Vorgang = "v-anders" },
			domain.ErrValidation,
		},
		"anderer Stand": {
			func(a *domain.Handelsabschluss) { a.Beleg.Abdruck = "abdruck-alt" },
			domain.ErrConflict,
		},
		"ohne Unterschrift": {
			func(a *domain.Handelsabschluss) { a.Unterschrift = "" },
			domain.ErrValidation,
		},
		"unverbuchter Zustand": {
			func(a *domain.Handelsabschluss) {
				a.Angebot.Zustand = domain.HandelUebergabeOffen
			},
			domain.ErrConflict,
		},
		"ohne Bewegung": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen = nil },
			domain.ErrValidation,
		},
		"Abgang ohne Besitzzeile": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen[0].ExemplarID = 0 },
			domain.ErrValidation,
		},
		"Zeile ohne Druck": {
			func(a *domain.Handelsabschluss) { a.Beleg.Zeilen[1].Printing = "" },
			domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			probe := eineAbschlussprobe(t)
			exemplar := probe.vierStueck(t, true)

			abschluss := einAbschluss(exemplar)
			fall.verderben(&abschluss)

			fehler := probe.buch.Abschliessen(probekontext(t), abschluss)
			if !errors.Is(fehler, fall.erwartet) {
				t.Fatalf("der Abschluss gab %v statt %v", fehler, fall.erwartet)
			}

			probe.nichtsGeschehen(t, exemplar)
		})
	}
}

// **Ein Abgang auf eine Zeile, die es nicht gibt.**
func TestEinAbgangOhneBesitzzeileSagtEs(t *testing.T) {
	probe := eineAbschlussprobe(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)
	abschluss.Beleg.Zeilen[0].ExemplarID = 999

	fehler := probe.buch.Abschliessen(probekontext(t), abschluss)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Abschluss gab %v statt „gibt es nicht\"", fehler)
	}

	probe.nichtsGeschehen(t, exemplar)
}

// **Ein anderes Profil sieht den Abschluss nicht** ([0066]).
func TestEinAbschlussGehoertZuEinemProfil(t *testing.T) {
	probe := eineAbschlussprobe(t)
	exemplar := probe.vierStueck(t, true)

	if fehler := probe.buch.Abschliessen(
		probekontext(t), einAbschluss(exemplar)); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	fremd := storage.MitProfil(t.Context(), 2)

	beleg, fehler := probe.buch.Beleg(fremd, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if beleg != nil {
		t.Error("das fremde Profil sieht den Beleg")
	}

	zeilen, fehler := probe.buch.Journal(fremd, einHandel)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 0 {
		t.Errorf("das fremde Profil sieht %d Journalzeilen", len(zeilen))
	}
}

// **Ohne Profil geht gar nichts** — jede Wurzelabfrage klemmt auf eines.
func TestOhneProfilBuchtNiemand(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ohne := t.Context()

	if fehler := probe.buch.Abschliessen(
		ohne, domain.Handelsabschluss{}); fehler == nil {
		t.Error("der Abschluss lief ohne Profil")
	}

	if _, fehler := probe.buch.Beleg(ohne, einHandel); fehler == nil {
		t.Error("der Beleg kam ohne Profil")
	}

	if _, fehler := probe.buch.Journal(ohne, einHandel); fehler == nil {
		t.Error("das Journal kam ohne Profil")
	}

	if fehler := probe.buch.Buchen(ohne, domain.Journalzeile{}); fehler == nil {
		t.Error("die Buchung lief ohne Profil")
	}

	if fehler := probe.buch.Gegenzeichnen(ohne,
		domain.Fremdbeleg{Vorgang: einHandel, Von: "bert", Abdruck: "a1"}); fehler == nil {
		t.Error("die Gegenzeichnung lief ohne Profil")
	}
}
