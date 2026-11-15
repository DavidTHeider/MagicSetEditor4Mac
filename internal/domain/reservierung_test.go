package domain_test

// **Die Reservierung** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Was hier gemessen wird
//
// **Dass niemand zweimal dasselbe Stück Papier zusagt.** Verfügbar ist der
// Besitz **minus** dem, was schon festgehalten ist — und wer mehr
// festhalten will, als frei ist, bekommt einen Konflikt.
//
// # Und die zwei Fälle, die gleich aussehen und es nicht sind
//
// **Eine abgelaufene Frist gibt frei, eine bestätigte Übergabe nicht.** Wer
// beides gleich behandelte, gäbe Karten frei, die längst über den Tisch
// gegangen sind — und der Bestand sagte, man habe sie noch.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// reservierungsuhr ist der Zeitpunkt, gegen den diese Läufe rechnen.
//
// **Nicht `reservierungsuhr`**: den Namen trägt schon der Nachweislauf, und zwei
// Zeitpunkte unter einem Namen in einem Paket gehen nicht.
var reservierungsuhr = time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

// einBesitz gibt eine Besitzzeile mit drei Stücken.
func einBesitz() domain.Exemplar {
	return domain.Exemplar{
		ID: 7, Fassung: 2, Printing: "abc-123", Menge: 3,
	}
}

// eineReservierung gibt eine gültige mit den genannten Abweichungen.
func eineReservierung(
	aendere ...func(*domain.Reservierung),
) domain.Reservierung {
	reservierung := domain.Reservierung{
		Vorgang:    "h1",
		Revision:   1,
		ExemplarID: 7,
		Menge:      2,
		Fassung:    2,
		Frist:      reservierungsuhr.Add(time.Hour),
	}

	for _, aenderung := range aendere {
		aenderung(&reservierung)
	}

	return reservierung
}

// **Verfügbar ist der Besitz minus dem, was festgehalten ist.**
func TestVerfuegbarIstBesitzMinusReservierung(t *testing.T) {
	t.Parallel()

	frei := domain.Verfuegbar(einBesitz(),
		[]domain.Reservierung{eineReservierung()}, reservierungsuhr)

	if frei != 1 {
		t.Errorf("frei sind %d von drei bei zwei festgehaltenen", frei)
	}
}

// **Ohne Reservierung ist alles frei.**
func TestOhneReservierungIstAllesFrei(t *testing.T) {
	t.Parallel()

	if frei := domain.Verfuegbar(einBesitz(), nil, reservierungsuhr); frei != 3 {
		t.Errorf("frei sind %d von drei", frei)
	}
}

// **Ein Wunsch ist kein Besitz und deshalb nichts Freies.**
//
// **Die Regel steht schon da** ([domain.Exemplar.Besitzt], 0057) — hier wird
// sie eingelöst: wer eine Wunschzeile reservierte, sagte etwas zu, das er
// nicht hat.
func TestEinWunschIstNichtsFreies(t *testing.T) {
	t.Parallel()

	wunsch := einBesitz()
	wunsch.Wunsch = true

	if frei := domain.Verfuegbar(wunsch, nil, reservierungsuhr); frei != 0 {
		t.Errorf("an einer Wunschzeile sind %d frei", frei)
	}
}

// **Eine abgelaufene Frist gibt frei.**
func TestEineAbgelaufeneFristGibtFrei(t *testing.T) {
	t.Parallel()

	abgelaufen := eineReservierung(func(r *domain.Reservierung) {
		r.Frist = reservierungsuhr.Add(-time.Minute)
	})

	if frei := domain.Verfuegbar(einBesitz(),
		[]domain.Reservierung{abgelaufen}, reservierungsuhr); frei != 3 {
		t.Errorf("frei sind %d von drei nach abgelaufener Frist", frei)
	}
}

// **Eine bestätigte Übergabe gibt nicht frei — auch nicht nach der Frist.**
//
// **Das ist die Zusage, an der am meisten hängt**: sobald eine Übergabe
// bestätigt ist, darf keine Uhr, kein Neustart und kein Verbindungsverlust
// die Stücke wieder freigeben. Was dann noch klärt, klärt ein Mensch.
func TestEineBestaetigteUebergabeGibtNichtFrei(t *testing.T) {
	t.Parallel()

	uebergeben := eineReservierung(func(r *domain.Reservierung) {
		r.Frist = reservierungsuhr.Add(-time.Hour)
		r.Uebergeben = true
	})

	if frei := domain.Verfuegbar(einBesitz(),
		[]domain.Reservierung{uebergeben}, reservierungsuhr); frei != 1 {
		t.Errorf("frei sind %d von drei trotz bestätigter Übergabe", frei)
	}
}

// **Zwei Vorgänge können zusammen nie mehr festhalten als den Besitz.**
//
// **Die erste beobachtbare Abnahme des Issues**, und sie steht hier in ihrer
// fachlichen Hälfte: dass auch ein **gleichzeitiger** Schreibweg sie nicht
// bricht, hält der Lauf gegen den echten Bestand.
func TestZweiVorgaengeKoennenNichtMehrAlsDenBesitzHalten(t *testing.T) {
	t.Parallel()

	erste := eineReservierung()

	zweite := eineReservierung(func(r *domain.Reservierung) {
		r.Vorgang = "h2"
		r.Menge = 2
	})

	fehler := domain.ReservierungPasst(
		einBesitz(), []domain.Reservierung{erste}, zweite, reservierungsuhr)

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}

	// **Eine, die passt, passt** — sonst prüfte der Lauf darüber nur, dass
	// überhaupt etwas abgewiesen wird.
	passend := eineReservierung(func(r *domain.Reservierung) {
		r.Vorgang = "h2"
		r.Menge = 1
	})

	if fehler := domain.ReservierungPasst(
		einBesitz(), []domain.Reservierung{erste}, passend, reservierungsuhr,
	); fehler != nil {
		t.Errorf("die passende wird abgewiesen: %v", fehler)
	}
}

// **Was eine Reservierung scheitern lässt.**
func TestWasEineReservierungScheiternLaesst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		verbiegen func(*domain.Reservierung)
		art       error
	}{
		"ohne Vorgang": {
			verbiegen: func(r *domain.Reservierung) { r.Vorgang = "" },
			art:       domain.ErrValidation,
		},
		"ohne Revision": {
			verbiegen: func(r *domain.Reservierung) { r.Revision = 0 },
			art:       domain.ErrValidation,
		},
		"ohne Exemplar": {
			verbiegen: func(r *domain.Reservierung) { r.ExemplarID = 0 },
			art:       domain.ErrValidation,
		},
		"ohne Menge": {
			verbiegen: func(r *domain.Reservierung) { r.Menge = 0 },
			art:       domain.ErrValidation,
		},
		"für ein anderes Exemplar": {
			verbiegen: func(r *domain.Reservierung) { r.ExemplarID = 9 },
			art:       domain.ErrValidation,
		},
		"gegen eine alte Fassung": {
			verbiegen: func(r *domain.Reservierung) { r.Fassung = 1 },
			art:       domain.ErrConflict,
		},
		"über den ganzen Besitz hinaus": {
			verbiegen: func(r *domain.Reservierung) { r.Menge = 4 },
			art:       domain.ErrConflict,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reservierung := eineReservierung()
			fall.verbiegen(&reservierung)

			if fehler := domain.ReservierungPasst(
				einBesitz(), nil, reservierung, reservierungsuhr,
			); !errors.Is(fehler, fall.art) {
				t.Errorf("%q ergibt %v", name, fehler)
			}
		})
	}
}

// **Und ein Wunsch lässt sich gar nicht reservieren.**
func TestEinWunschLaesstSichNichtReservieren(t *testing.T) {
	t.Parallel()

	wunsch := einBesitz()
	wunsch.Wunsch = true

	if fehler := domain.ReservierungPasst(
		wunsch, nil, eineReservierung(), reservierungsuhr,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}
