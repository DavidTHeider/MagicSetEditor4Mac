package leitung_test

// **Nachrichten über einen Strom** — Werkstück N2 von Phase 5.4
// (#1455),
// erster Schnitt.
//
// # Wogegen hier geprüft wird
//
// **Dass eine Nachricht eine Nachricht bleibt.** TCP ist ein Strom: wer drei
// Nachrichten hineinschreibt, bekommt am anderen Ende irgendeine Aufteilung
// derselben Bytes. Eine Hülle, die in zwei Stücken ankommt, ist kein JSON
// mehr — und zwei Hüllen, die in einem Stück ankommen, auch nicht.
//
// **Und dass eine behauptete Länge keinen Speicher stellt.** Der Rahmen sagt,
// wie viel folgt. Wer ihm glaubt und zuteilt, hat einem Fremden erlaubt, den
// Arbeitsspeicher dieses Rechners zu bestimmen — mit vier Bytes.

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// einPaar gibt zwei verbundene Leitungen, die mit dem Lauf enden.
func einPaar(t *testing.T) (hier, dort *leitung.Leitung) {
	t.Helper()

	hier, dort = leitung.Schleife("kanal-probe")

	t.Cleanup(func() {
		_ = hier.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dort.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	return hier, dort
}

// **Eine Nachricht geht hin und kommt an.**
func TestEineNachrichtGehtHinUndKommtAn(t *testing.T) {
	t.Parallel()

	hier, dort := einPaar(t)

	fertig := make(chan []byte, 1)

	go func() {
		gelesen, fehler := dort.Empfange(t.Context())
		if fehler != nil {
			t.Errorf("Empfange: %v", fehler)
		}

		fertig <- gelesen
	}()

	if fehler := hier.Sende(t.Context(), []byte("hallo")); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	if angekommen := <-fertig; string(angekommen) != "hallo" {
		t.Fatalf("angekommen ist %q", angekommen)
	}
}

// **Drei Nachrichten bleiben drei — und in ihrer Reihenfolge.**
//
// # Der Satz, für den es diese Schicht gibt
//
// **Ein Strom kennt keine Nachrichten.** Ohne Rahmen bekäme der Empfänger
// „einszweidrei" in irgendeiner Aufteilung, und keine davon wäre eine Hülle.
func TestDreiNachrichtenBleibenDrei(t *testing.T) {
	t.Parallel()

	hier, dort := einPaar(t)

	gesendet := []string{"eins", "zweiundzwanzig", "drei"}
	empfangen := make(chan string, len(gesendet))

	go func() {
		for range gesendet {
			gelesen, fehler := dort.Empfange(t.Context())
			if fehler != nil {
				t.Errorf("Empfange: %v", fehler)

				return
			}

			empfangen <- string(gelesen)
		}
	}()

	for _, was := range gesendet {
		if fehler := hier.Sende(t.Context(), []byte(was)); fehler != nil {
			t.Fatalf("Sende %q: %v", was, fehler)
		}
	}

	for _, erwartet := range gesendet {
		if angekommen := <-empfangen; angekommen != erwartet {
			t.Fatalf("angekommen ist %q, erwartet war %q",
				angekommen, erwartet)
		}
	}
}

// **Eine zu große Nachricht geht nicht hinaus.**
func TestEineZuGrosseNachrichtGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	hier, _ := einPaar(t)

	if fehler := hier.Sende(
		t.Context(), make([]byte, peer.Hoechstlaenge+1),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("sie kam durch: %v", fehler)
	}

	// **Und eine leere auch nicht.** Sie trägt nichts und wäre ein Weg, ein
	// Gegenüber beschäftigt zu halten.
	if fehler := hier.Sende(t.Context(), nil); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die leere kam durch: %v", fehler)
	}
}

// **Ein behaupteter Rahmen stellt keinen Speicher.**
//
// # Wie das überhaupt zu messen ist
//
// **Am Warten.** Geschickt werden **nur die vier Bytes des Rahmens**, die eine
// riesige Länge behaupten — und sonst nichts. Wer der Zahl glaubt, teilt zu
// und **wartet** dann auf Bytes, die nie kommen; der Lauf liefe in seine
// Zeitgrenze. Wer sie prüft, ist sofort fertig.
//
// **Das ist der Unterschied zwischen „wird abgewiesen" und „wird abgewiesen,
// bevor es kostet"** — und nur der zweite ist die Zusage (§10).
func TestEinBehaupteterRahmenStelltKeinenSpeicher(t *testing.T) {
	t.Parallel()

	hier, dort := net.Pipe()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
		_ = dort.Close() //nolint:errcheck // nur ein Lauf
	})

	empfaenger := leitung.Ueber(dort, "kanal-probe")

	go func() {
		var kopf [4]byte

		binary.BigEndian.PutUint32(kopf[:], uint32(peer.Hoechstlaenge)+1)

		if _, fehler := hier.Write(kopf[:]); fehler != nil {
			t.Errorf("Write: %v", fehler)
		}
	}()

	// **Eine großzügige Frist**, damit der Lauf nicht die Maschine misst:
	// wer die Länge prüft, ist in Mikrosekunden fertig.
	ctx, abbrechen := context.WithTimeout(t.Context(), 2*time.Second)
	defer abbrechen()

	if _, fehler := empfaenger.Empfange(ctx); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der behauptete Rahmen kam durch: %v", fehler)
	}
}

// **Eine Frist, die schon um ist, lässt nichts warten.**
func TestEineAbgelaufeneFristLaesstNichtsWarten(t *testing.T) {
	t.Parallel()

	hier, dort := einPaar(t)

	ctx, abbrechen := context.WithDeadline(
		t.Context(), time.Now().Add(-time.Second))
	defer abbrechen()

	if fehler := hier.Sende(ctx, []byte("hallo")); fehler == nil {
		t.Error("das Senden wartete nicht auf die Frist")
	}

	if _, fehler := dort.Empfange(ctx); fehler == nil {
		t.Error("das Empfangen wartete nicht auf die Frist")
	}
}

// **Ein Abbruch unterbricht ein wartendes Empfangen.**
//
// **Sonst hinge ein Ruf an einem Gegenüber, das nichts mehr schickt** — und
// §16 verlangt für jede Anfrage eine Frist **und** einen Abbruch.
func TestEinAbbruchUnterbrichtEinWartendesEmpfangen(t *testing.T) {
	t.Parallel()

	_, dort := einPaar(t)

	ctx, abbrechen := context.WithCancel(t.Context())

	fertig := make(chan error, 1)

	go func() {
		_, fehler := dort.Empfange(ctx)
		fertig <- fehler
	}()

	abbrechen()

	select {
	case fehler := <-fertig:
		if fehler == nil {
			t.Fatal("das Empfangen kam ohne Fehler zurück")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("das Empfangen hing trotz Abbruch")
	}
}

// **Eine geschlossene Leitung sendet nicht mehr — und zweimal schließen ist
// einmal.**
func TestEineGeschlosseneLeitungSendetNichtMehr(t *testing.T) {
	t.Parallel()

	hier, _ := einPaar(t)

	if fehler := hier.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if fehler := hier.Schliesse(); fehler != nil {
		t.Fatalf("das zweite Schließen klagte: %v", fehler)
	}

	if fehler := hier.Sende(t.Context(), []byte("hallo")); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("die geschlossene Leitung sendete: %v", fehler)
	}

	if _, fehler := hier.Empfange(t.Context()); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("die geschlossene Leitung empfing: %v", fehler)
	}
}

// **Zwei gleichzeitige Sender verschränken ihre Rahmen nicht.**
//
// # Der Fund
//
// **Ohne Schloss stünde die Länge des einen vor dem Rumpf des anderen.** Zwei
// Schreibvorgänge auf denselben Strom dürfen sich nicht verzahnen; der
// Empfänger läse sonst eine Länge und danach fremde Bytes — und was er daraus
// macht, wäre Zufall.
func TestZweiGleichzeitigeSenderVerschraenkenNicht(t *testing.T) {
	t.Parallel()

	hier, dort := einPaar(t)

	const anzahl = 20

	nachrichten := []string{
		strings.Repeat("a", 500), strings.Repeat("b", 17),
	}

	empfangen := make(chan string, 2*anzahl)

	go func() {
		for range 2 * anzahl {
			gelesen, fehler := dort.Empfange(t.Context())
			if fehler != nil {
				t.Errorf("Empfange: %v", fehler)

				return
			}

			empfangen <- string(gelesen)
		}
	}()

	var warten sync.WaitGroup

	for _, was := range nachrichten {
		warten.Add(1)

		go func() {
			defer warten.Done()

			for range anzahl {
				if fehler := hier.Sende(
					t.Context(), []byte(was)); fehler != nil {
					t.Errorf("Sende: %v", fehler)

					return
				}
			}
		}()
	}

	warten.Wait()

	// **Jede empfangene Nachricht ist eine der beiden, ganz.** Eine
	// verschränkte wäre weder die eine noch die andere.
	for range 2 * anzahl {
		gelesen := <-empfangen

		if gelesen != nachrichten[0] && gelesen != nachrichten[1] {
			t.Fatalf("verschränkt angekommen: %d Bytes", len(gelesen))
		}
	}
}

// **Ein abgeschnittener Rahmen endet mit einem Fehler und nicht mit Warten.**
func TestEinAbgeschnittenerRahmenEndetMitEinemFehler(t *testing.T) {
	t.Parallel()

	hier, dort := net.Pipe()

	t.Cleanup(func() { _ = dort.Close() }) //nolint:errcheck // nur ein Lauf

	empfaenger := leitung.Ueber(dort, "kanal-probe")

	go func() {
		var kopf [4]byte

		binary.BigEndian.PutUint32(kopf[:], 100)

		if _, fehler := hier.Write(
			append(kopf[:], bytes.Repeat([]byte("x"), 10)...)); fehler != nil {
			t.Errorf("Write: %v", fehler)
		}

		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	}()

	ctx, abbrechen := context.WithTimeout(t.Context(), 2*time.Second)
	defer abbrechen()

	if _, fehler := empfaenger.Empfange(ctx); fehler == nil {
		t.Fatal("der abgeschnittene Rahmen kam durch")
	}
}

// **Der Kanal kommt durch — und beide Enden nennen denselben.**
//
// # Warum hier kein Gegenüber mehr steht
//
// **Hier stand „Gegenüber und Kanal kommen durch"**, und der Satz daneben
// lautete: „sie sind das, was [service.Verbindung] von einer Leitung wissen
// will". Die Kennung war es nie — kein produktiver Leser hat sie je gelesen,
// und wer da ist, sagt der Handschlag. Am 2026-09-14 ist sie gefallen.
//
// **Dass beide Enden denselben Kanal nennen, ist dagegen tragend.** Darauf
// ruht der Schutz gegen das Wiedereinspielen über Verbindungen hinweg: der
// Handschlag unterschreibt ihn, und zwei verschiedene Zeichenketten ergäben
// zwei verschiedene Unterschriften.
func TestBeideEndenNennenDenselbenKanal(t *testing.T) {
	t.Parallel()

	hier, dort := einPaar(t)

	if hier.Kanal() != "kanal-probe" || dort.Kanal() != "kanal-probe" {
		t.Errorf("die Kanäle sind %q und %q", hier.Kanal(), dort.Kanal())
	}
}
