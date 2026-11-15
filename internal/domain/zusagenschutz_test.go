package domain_test

// **Was eine offene Zusage einem Schreibweg verbietet** — Werkstück `T2` von
// Phase 5.6
// (#1467).
//
// # Warum das eine eigene Frage ist
//
// **Nebenan steht, was in eine Zusage hineinpasst** (`reservierung_test.go`);
// hier steht, was **danach** noch geschrieben werden darf. Das Issue nennt
// beides getrennt: die Zusage prüft, ob genug frei ist — der Schreibweg prüft,
// ob nach ihm noch genug dasteht.
//
// # Und warum nicht jede Änderung verboten ist
//
// **Weil eine Änderung eine Zusage entwerten darf, solange nichts vorbereitet
// ist.** Die Fassung der Zusage sagt, gegen welchen Stand gerechnet wurde;
// wer die Zeile ändert, hebt sie — und der Handel muss neu prüfen lassen. Was
// **nicht** geschehen darf, ist, dass zugesagte Stücke verschwinden.

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zusage baut eine Zusage über ein Exemplar.
func zusage(menge int, uebergeben bool) domain.Reservierung {
	return domain.Reservierung{
		Vorgang: "h1", Revision: 1, ExemplarID: 7, Menge: menge,
		Fassung: 1, Uebergeben: uebergeben,
	}
}

// besessen baut die Besitzzeile, um die es geht.
func besessen(menge int) domain.Exemplar {
	return domain.Exemplar{
		ID: 7, Fassung: 1, Printing: "abc-123", Menge: menge,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}
}

// **Was zugesagt ist, darf nicht wegfallen.**
func TestEineAenderungUnterDieZusageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	zusagen := []domain.Reservierung{zusage(2, false)}

	fehler := domain.AenderungPasst(besessen(1), zusagen, jetzt)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Die Meldung nennt beide Zahlen.** „Geht nicht" schickt jemanden auf
	// die Suche nach dem Grund (§8).
	if satz := fehler.Error(); !strings.Contains(satz, "2") ||
		!strings.Contains(satz, "1") {
		t.Errorf("die Meldung ist %q", satz)
	}
}

// **Und was darüber liegt, darf geschrieben werden.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie hielte auch eine Fassung,
// die jede Änderung an einer zugesagten Zeile abweist — und damit wäre eine
// Sammlung mit einem offenen Handel nicht mehr zu pflegen.
func TestEineAenderungUeberDerZusageGehtDurch(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
	zusagen := []domain.Reservierung{zusage(2, false)}

	for name, menge := range map[string]int{"genau so viele": 2, "mehr": 5} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.AenderungPasst(
				besessen(menge), zusagen, jetzt); fehler != nil {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Eine abgelaufene Zusage hält nichts mehr fest.**
func TestEineAbgelaufeneZusageHaeltNichtsFest(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

	abgelaufen := zusage(2, false)
	abgelaufen.Frist = jetzt.Add(-time.Hour)

	if fehler := domain.AenderungPasst(
		besessen(0), []domain.Reservierung{abgelaufen}, jetzt); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine bestätigte Übergabe läuft nicht ab.**
//
// **Der Unterschied, um den es geht**: eine Frist gibt frei, eine Übergabe
// nicht. Wer beides gleich behandelte, gäbe Karten frei, die längst über den
// Tisch gegangen sind.
func TestEineUebergebeneZusageHaeltTrotzFrist(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

	uebergeben := zusage(2, true)
	uebergeben.Frist = jetzt.Add(-time.Hour)

	if fehler := domain.AenderungPasst(
		besessen(1), []domain.Reservierung{uebergeben}, jetzt,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Zeile mit offener Zusage wird nicht gelöscht.**
//
// **Sonst nähme die Kaskade die Zusage mit** — `reservierungen.exemplar_id`
// hängt mit `ON DELETE CASCADE` daran (0069). Der Handel wüsste dann nichts
// mehr davon, und der Partner wartete auf eine Karte, die es nicht gibt.
func TestEineZeileMitZusageWirdNichtGeloescht(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

	fehler := domain.LoeschenPasst(
		7, []domain.Reservierung{zusage(1, false)}, jetzt)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Ohne Zusage geht es** — und mit einer abgelaufenen auch.
	abgelaufen := zusage(1, false)
	abgelaufen.Frist = jetzt.Add(-time.Hour)

	for name, zusagen := range map[string][]domain.Reservierung{
		"keine":       nil,
		"abgelaufene": {abgelaufen},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.LoeschenPasst(7, zusagen, jetzt); fehler != nil {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}
