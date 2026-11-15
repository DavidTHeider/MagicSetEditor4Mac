package service_test

// **Ein Gegenbeleg, der nicht passt, bleibt nicht still** —
// #2500.
//
// # Was sich geändert hat
//
// **Bis hierher endete die Zustellung mit einem Fehler**, und die Sitzung
// mit ihr: beim Absender stand weiter „hinausgegangen", beim Empfänger
// nichts. Jetzt hält der Empfänger einen benannten Befund fest und stellt
// weiter zu. **Ein Fehler bleibt, was wiederholbar ist**: ohne eigenen Beleg
// (es wird noch verbucht) und wenn der Befund sich nicht festhalten lässt.

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// derBefund verlangt genau einen Befund mit diesem Mangel — und keine
// Gegenzeichnung.
func derBefund(
	t *testing.T, probe *gegenbelegprobe, von string, mangel domain.Gegenbelegmangel,
) domain.Gegenbelegbefund {
	t.Helper()

	if len(probe.bekomen) != 0 {
		t.Errorf("gegengezeichnet wurde %dmal", len(probe.bekomen))
	}

	if len(probe.abgewiesen) != 1 {
		t.Fatalf("%d Befunde, erwartet einer", len(probe.abgewiesen))
	}

	befund := probe.abgewiesen[0]
	if befund.Mangel != mangel || befund.Von != von || befund.Vorgang != derPostvorgang ||
		befund.Grund == "" {
		t.Errorf("der Befund ist %+v, erwartet Mangel %q von %q", befund, mangel, von)
	}

	return befund
}

// TestEineUrkundeDieNichtSpiegeltWirdFestgehalten ist das Rot aus #2500:
// echt unterschrieben und trotzdem ein anderer Handel.
func TestEineUrkundeDieNichtSpiegeltWirdFestgehalten(t *testing.T) {
	t.Parallel()

	probe, _, drueben := einGegenbelegpaar(t)

	// **Dieselben Zeilen in derselben Richtung** — beide gäben dasselbe weg.
	beleg := domain.Handelsbeleg{
		Vorgang: derPostvorgang, Abdruck: einBelegstand(drueben).Abdruck(),
		Parteien: einBelegstand(drueben).Parteien, Zeilen: probe.zeilen,
	}

	urkunde := domain.Belegurkunde{
		Beleg:        beleg,
		Unterschrift: hex.EncodeToString(drueben.Signiere(beleg.Unterschriftsform())),
	}

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); fehler != nil {
		t.Fatalf("die Zustellung hielt an: %v", fehler)
	}

	befund := derBefund(t, probe, string(drueben.Kennung()), domain.MangelZeilen)
	if befund.Abdruck != beleg.Abdruck {
		t.Errorf("der Befund nennt den Abdruck %q, erwartet %q", befund.Abdruck, beleg.Abdruck)
	}
}

// TestEineFremdeUnterschriftWirdFestgehalten — der Sprecher ist echt, die
// Unterschrift stammt von einem anderen.
func TestEineFremdeUnterschriftWirdFestgehalten(t *testing.T) {
	t.Parallel()

	probe, urkunde, _ := einGegenbelegpaar(t)

	jemand, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	post := einePostMitBelegen(t, probe, jemand)

	if _, fehler := post.Empfangen(context.Background(),
		string(jemand.Kennung()), eineBelegnachricht(urkunde, jemand)); fehler != nil {
		t.Fatalf("die Zustellung hielt an: %v", fehler)
	}

	derBefund(t, probe, string(jemand.Kennung()), domain.MangelUnterschrift)
}

// TestEineKrummeUnterschriftWirdFestgehalten — kein Hex ist keine
// Unterschrift.
func TestEineKrummeUnterschriftWirdFestgehalten(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	urkunde.Unterschrift = "keine Zahl"

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben)); fehler != nil {
		t.Fatalf("die Zustellung hielt an: %v", fehler)
	}

	derBefund(t, probe, string(drueben.Kennung()), domain.MangelUnterschrift)
}

// TestEinBefundDerNichtBleibtHaeltDieZustellungAn — wer ihn nicht festhalten
// kann, stellt nicht zu: der nächste Anlauf versucht es noch einmal.
func TestEinBefundDerNichtBleibtHaeltDieZustellungAn(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	urkunde.Unterschrift = "keine Zahl"
	probe.weistab = domain.ErrValidation

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// TestOhneEigenenBelegGibtEsKeinenBefund — dann wird noch verbucht, und der
// Fehler sorgt dafür, dass es noch einmal versucht wird (TRADE-03).
func TestOhneEigenenBelegGibtEsKeinenBefund(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	probe.stand = nil
	urkunde.Unterschrift = "keine Zahl"

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}

	if len(probe.abgewiesen) != 0 {
		t.Errorf("%d Befunde ohne eigenen Beleg", len(probe.abgewiesen))
	}
}
