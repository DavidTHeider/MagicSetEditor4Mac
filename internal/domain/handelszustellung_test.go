package domain_test

// **Was mit einer zweiten Zustellung geschieht** — Werkstück `T5` von Phase
// 5.6 (#1470).
//
// # Was hier gemessen wird
//
// **Dass zweimal dieselbe Nachricht einmal wirkt.** Ein Netz verliert,
// verdoppelt und ordnet um; wer jede Zustellung ausführt, verbucht denselben
// Tausch zweimal.
//
// **Dass zweimal dieselbe Operation mit anderem Inhalt ein Konflikt ist** —
// und nicht die zweite, die die erste überschreibt.
//
// **Und dass ein Grabstein länger steht, als eine Nachricht leben darf.**
// Wer ihn zu früh wegräumt, lässt dieselbe Wirkung ein zweites Mal zu.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// jetztFuerZustellung ist die Uhr dieser Läufe.
func jetztFuerZustellung() time.Time {
	return time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
}

// einEintrag baut, was von einer behandelten Nachricht bleibt.
func einEintrag() domain.Inboxeintrag {
	return domain.Inboxeintrag{
		Schluessel:      "1\x00v-1\x00vorschlag\x002",
		Nutzlastabdruck: "abdruck-1",
		Ergebnis:        `{"angenommen":true}`,
		Behandelt:       jetztFuerZustellung().Add(-time.Hour),
	}
}

func TestWasMitEinerZustellungGeschieht(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		vorher   *domain.Inboxeintrag
		abdruck  string
		erwartet domain.Zustellungsart
	}{
		// **Nichts gesehen, also neu** — der gewöhnliche Fall.
		"noch nie gesehen": {nil, "abdruck-1", domain.ZustellungNeu},

		// **Dieselbe Nachricht noch einmal** — dieselbe Antwort, keine
		// zweite Wirkung.
		"dieselbe noch einmal": {
			zeiger(einEintrag()), "abdruck-1", domain.ZustellungBekannt,
		},

		// **Dieselbe Operation, anderer Inhalt** — das ist ein Konflikt und
		// nicht die zweite, die gewinnt.
		"dieselbe Operation, anderer Inhalt": {
			zeiger(einEintrag()), "abdruck-2", domain.ZustellungKonflikt,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gilt := domain.Zustellung(fall.vorher, fall.abdruck); gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}

// zeiger gibt einen Zeiger auf den Wert.
func zeiger[T any](wert T) *T { return &wert }

// **Ein Grabstein steht, solange eine Nachricht leben darf.**
func TestEinGrabsteinStehtLaengerAlsEineNachrichtLebt(t *testing.T) {
	t.Parallel()

	jetzt := jetztFuerZustellung()

	frisch := einEintrag()
	if frisch.Abgelaufen(jetzt) {
		t.Error("ein Eintrag von vor einer Stunde gilt als abgelaufen")
	}

	// **Genau auf der Frist steht er noch** — abgelaufen ist, was **darüber**
	// liegt. Eine Grenze, die ihren eigenen Wert einschließt, räumt eine
	// Nachricht weg, die im selben Augenblick ankommt.
	auf := einEintrag()
	auf.Behandelt = jetzt.Add(-domain.Idempotenzfrist)

	if auf.Abgelaufen(jetzt) {
		t.Error("auf der Frist gilt er schon als abgelaufen")
	}

	alt := einEintrag()
	alt.Behandelt = jetzt.Add(-domain.Idempotenzfrist - time.Second)

	if !alt.Abgelaufen(jetzt) {
		t.Error("ein Eintrag über der Frist gilt noch")
	}
}

// **Die Frist ist genannt und nicht geraten.**
func TestDieFristIstEineGenannteZahl(t *testing.T) {
	t.Parallel()

	if domain.Idempotenzfrist != 30*24*time.Hour {
		t.Errorf("die Frist ist %v", domain.Idempotenzfrist)
	}
}
