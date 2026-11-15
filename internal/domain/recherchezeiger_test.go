package domain_test

// **Der Fortsetzungszeiger** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Der Satz, gegen den hier geprüft wird**: „bei Wechsel konsistent
// fortsetzen oder explizit ungültig, **keine still gemischten Seiten**."

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineZeigeranfrage ist die, aus der ein Abdruck entsteht.
func eineZeigeranfrage() domain.Rechercheanfrage {
	return domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "spielsteine", Sprache: "de",
		Grenze: 10,
	}
}

// **Ein Zeiger kommt zurück, wie er hinausging.**
func TestEinZeigerKommtZurueckWieErHinausging(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	abdruck := domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-1")

	roh := domain.Recherchezeiger{
		Stelle: 10, Abdruck: abdruck, Erstellt: jetzt,
	}.AlsZeichenkette()

	gelesen, fehler := domain.ZeigerAus(roh, abdruck, jetzt.Add(time.Minute))
	if fehler != nil {
		t.Fatalf("ZeigerAus: %v", fehler)
	}

	if gelesen.Stelle != 10 {
		t.Fatalf("die Stelle ist %d", gelesen.Stelle)
	}
}

// **Jede Änderung an der Frage macht den Zeiger fremd.**
//
// # Der Fall, für den es das gibt
//
// **Eine zweite Seite gehört zu einer Liste.** Wer die Frage ändert, das
// Profil wechselt oder den Index neu baut, bekommt eine **andere** Liste —
// und die zweite Seite daraus wäre eine Mischung, die niemand gestellt hat.
func TestJedeAenderungMachtDenZeigerFremd(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	abdruck := domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-1")

	roh := domain.Recherchezeiger{
		Stelle: 10, Abdruck: abdruck, Erstellt: jetzt,
	}.AlsZeichenkette()

	for name, anderer := range map[string]string{
		"andere Frage": domain.Zeigerabdruck(domain.Rechercheanfrage{
			Fall: domain.FallKarten, Frage: "etwas anderes", Sprache: "de",
			Grenze: 10,
		}, 1, "index-1"),
		"anderer Fall": domain.Zeigerabdruck(domain.Rechercheanfrage{
			Fall: domain.FallRegeln, Frage: "spielsteine", Sprache: "de",
			Grenze: 10,
		}, 1, "index-1"),
		"anderes Profil": domain.Zeigerabdruck(eineZeigeranfrage(), 2, "index-1"),
		"anderer Index":  domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-2"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := domain.ZeigerAus(roh, anderer,
				jetzt); !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Ein Zeiger läuft ab.**
func TestEinZeigerLaeuftAb(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	abdruck := domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-1")

	roh := domain.Recherchezeiger{Abdruck: abdruck, Erstellt: jetzt}.
		AlsZeichenkette()

	// **Kurz davor gilt er noch.**
	if _, fehler := domain.ZeigerAus(roh, abdruck,
		jetzt.Add(domain.Zeigerfrist-time.Second)); fehler != nil {
		t.Fatalf("kurz vor der Frist: %v", fehler)
	}

	if _, fehler := domain.ZeigerAus(roh, abdruck,
		jetzt.Add(domain.Zeigerfrist)); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("zur Frist: %v", fehler)
	}
}

// **Ein unlesbarer Zeiger ist ein Fehler der Anfrage.**
//
// **Und kein Konflikt**: „unlesbar" heißt, jemand hat etwas Falsches
// geschickt; „fremd" heißt, die Welt hat sich geändert. Wer beides
// zusammenfasste, ließe jemanden raten, was zu tun ist.
func TestEinUnlesbarerZeigerIstEinFehlerDerAnfrage(t *testing.T) {
	t.Parallel()

	abdruck := domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-1")

	for name, roh := range map[string]string{
		"kein base64": "!!!nicht base64!!!",
		"kein JSON":   "bmljaHQgSlNPTg",
		"leer":        "",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := domain.ZeigerAus(roh, abdruck,
				time.Now()); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Eine negative Stelle ist keine.**
func TestEineNegativeStelleIstKeine(t *testing.T) {
	t.Parallel()

	jetzt := time.Now()
	abdruck := domain.Zeigerabdruck(eineZeigeranfrage(), 1, "index-1")

	roh := domain.Recherchezeiger{
		Stelle: -1, Abdruck: abdruck, Erstellt: jetzt,
	}.AlsZeichenkette()

	if _, fehler := domain.ZeigerAus(roh, abdruck, jetzt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die negative Stelle kam durch: %v", fehler)
	}
}
