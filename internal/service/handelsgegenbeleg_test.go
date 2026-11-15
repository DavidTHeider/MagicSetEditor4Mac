package service_test

// **Ein angekommener Beleg wird geprüft, bevor er einer ist** —
// #1650.
//
// # Wogegen hier geprüft wird
//
// **Gegen `beidseitig` als Behauptung.** Bis zum 2026-09-19 setzte die
// Spalte, sobald an **diesem** Fenster jemand für beide Parteien bestätigt
// hatte. Jetzt setzt sie nur, wer eine Unterschrift vorweisen kann, die zur
// Peerkennung des Absenders **und** zum eigenen Beleg passt.

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// gegenbelegprobe hält den eigenen Beleg und nimmt den fremden entgegen.
type gegenbelegprobe struct {
	stand  *domain.Belegstand
	zeilen []domain.Journalzeile

	// klagt und journalklage sind **zwei** Schalter, damit jeder der beiden
	// Lesewege für sich geprüft werden kann. Ein Doppel, das auf beiden
	// klagt, klagt schon beim ersten — und der zweite bliebe ungemessen
	// (§10).
	klagt        error
	journalklage error

	nimmt   error
	bekomen []domain.Fremdbeleg

	// weistab ist die Antwort auf [Abweisen], abgewiesen was ankam (#2500).
	weistab    error
	abgewiesen []domain.Gegenbelegbefund
}

func (g *gegenbelegprobe) Beleg(
	context.Context, string,
) (*domain.Belegstand, error) {
	if g.klagt != nil {
		return nil, g.klagt
	}

	return g.stand, nil
}

func (g *gegenbelegprobe) Journal(
	context.Context, string,
) ([]domain.Journalzeile, error) {
	if g.journalklage != nil {
		return nil, g.journalklage
	}

	return g.zeilen, nil
}

func (g *gegenbelegprobe) Abweisen(
	_ context.Context, befund domain.Gegenbelegbefund,
) error {
	g.abgewiesen = append(g.abgewiesen, befund)

	return g.weistab
}

func (g *gegenbelegprobe) Gegenzeichnen(
	_ context.Context, fremd domain.Fremdbeleg,
) error {
	g.bekomen = append(g.bekomen, fremd)

	return g.nimmt
}

// einBelegstand ist der Vorgang, zu dem die Belege gehören — mit der
// **echten** Peerkennung als Partei.
//
// **Das ist keine Kosmetik**: die Post lässt nur eine Partei durch (Befund 3
// aus #1634), und `peer.Pruefe` verlangt eine Kennung, die einen Schlüssel
// trägt. Ein erfundener Name erfüllt genau eines von beidem.
func einBelegstand(drueben peer.Identitaet) domain.Handelsangebot {
	stand := derHeutigeStand()
	stand.Parteien = []string{"bich", string(drueben.Kennung())}

	return stand
}

// einePostMitBelegen stellt die Post samt Gegenbelegstelle auf.
func einePostMitBelegen(
	t *testing.T, probe *gegenbelegprobe, drueben peer.Identitaet,
) *service.Handelspost {
	t.Helper()

	return service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: einBelegstand(drueben)},
		kontaktnachschlagprobe{kontakt: domain.Peerkontakt{
			Kennung: string(drueben.Kennung()),
			Stand:   domain.KontaktBestaetigt,
			Rechte:  []domain.Zweck{domain.ZweckHandel},
		}}).MitGegenbelegen(probe)
}

// eineBelegnachricht ist, was mit einem Abschluss hinausgeht.
func eineBelegnachricht(
	urkunde domain.Belegurkunde, drueben peer.Identitaet,
) service.Handelsnachricht {
	stand := einBelegstand(drueben)

	nachricht := einePost(stand, func(umschlag *domain.Handelsumschlag) {
		umschlag.Art = domain.NachrichtBeleg
		umschlag.Absender = string(drueben.Kennung())
	})
	nachricht.Beleg = &urkunde

	return nachricht
}

// einGegenbelegpaar baut, was hier liegt, und was von drüben kommt.
func einGegenbelegpaar(t *testing.T) (
	*gegenbelegprobe, domain.Belegurkunde, peer.Identitaet,
) {
	t.Helper()

	drueben, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	meine := []domain.Journalzeile{{
		Art: domain.JournalAbgang, Printing: "p-1", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil,
	}}

	seine := []domain.Journalzeile{{
		Art: domain.JournalZugang, Printing: "p-1", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil,
	}}

	stand := einBelegstand(drueben)

	probe := &gegenbelegprobe{
		stand: &domain.Belegstand{
			Vorgang: derPostvorgang, Abdruck: stand.Abdruck(),
		},
		zeilen: meine,
	}

	beleg := domain.Handelsbeleg{
		Vorgang: derPostvorgang, Abdruck: stand.Abdruck(),
		Parteien: stand.Parteien, Zeilen: seine,
	}

	return probe, domain.Belegurkunde{
		Beleg: beleg,
		Unterschrift: hex.EncodeToString(
			drueben.Signiere(beleg.Unterschriftsform())),
	}, drueben
}

// TestEinGeprueterBelegZeichnetGegen ist die Zusage aus #1650.
func TestEinGeprueterBelegZeichnetGegen(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if len(probe.bekomen) != 1 {
		t.Fatalf("gegengezeichnet wurde %dmal", len(probe.bekomen))
	}

	if probe.bekomen[0].Von != string(drueben.Kennung()) {
		t.Errorf("der Beleg kam von %q", probe.bekomen[0].Von)
	}
}

// TestEinBelegOhneEigenenWirdAbgewiesen — hier ist nichts verbucht.
func TestEinBelegOhneEigenenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	probe.stand = nil

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// TestEinKlemmenderBestandHaeltDenGegenbelegAn — jeder Weg, auf dem der
// eigene Beleg gelesen wird, reicht seinen Fehler durch.
//
// **Drei Wege und nicht einer**: der Beleg, sein Journal und der Stand. Wer
// nur den ersten prüfte, hätte zwei Zweige, die kein Lauf betritt (§10).
func TestEinKlemmenderBestandHaeltDenGegenbelegAn(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	probe.klagt = errors.New("der Bestand klemmt")

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); fehler == nil {
		t.Error("der Lesefehler kam nicht an")
	}

	if len(probe.bekomen) != 0 {
		t.Errorf("gegengezeichnet wurde %dmal", len(probe.bekomen))
	}
}

// TestEineAbgewieseneGegenzeichnungKommtZurueck — der Bestand darf Nein
// sagen, und das Nein reist.
func TestEineAbgewieseneGegenzeichnungKommtZurueck(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	probe.nimmt = errors.New("der Bestand nimmt ihn nicht")

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); fehler == nil {
		t.Error("die Absage kam nicht an")
	}
}

// TestEineNachrichtOhneUrkundeZeichnetNichtGegen — die Art allein belegt
// nichts.
func TestEineNachrichtOhneUrkundeZeichnetNichtGegen(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)

	nachricht := eineBelegnachricht(urkunde, drueben)
	nachricht.Beleg = nil

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), nachricht); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if len(probe.bekomen) != 0 {
		t.Errorf("gegengezeichnet wurde %dmal", len(probe.bekomen))
	}
}

// TestEinKlemmendesJournalHaeltDenGegenbelegAn — der zweite Leseweg.
func TestEinKlemmendesJournalHaeltDenGegenbelegAn(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	probe.journalklage = errors.New("das Journal klemmt")

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben),
	); fehler == nil {
		t.Error("der Lesefehler kam nicht an")
	}
}

// TestEineUnterschriftLiegtInIhrerKanonischenSchreibung — Nebenbefund zu
// QA-FZ01 (#2091).
//
// **Dieselbe Unterschrift in Großschreibung ist dieselbe Unterschrift** —
// geprüft wird sie dekodiert. Abgelegt wurde sie bis zum 2026-10-06 so, wie
// sie ankam, und das Journal vergleicht die Zeichenkette: derselbe Beleg
// zweimal, einmal groß, einmal klein, stünde als Konflikt in der Klärung.
func TestEineUnterschriftLiegtInIhrerKanonischenSchreibung(t *testing.T) {
	t.Parallel()

	probe, urkunde, drueben := einGegenbelegpaar(t)
	klein := urkunde.Unterschrift
	urkunde.Unterschrift = strings.ToUpper(klein)

	post := einePostMitBelegen(t, probe, drueben)

	if _, fehler := post.Empfangen(context.Background(),
		string(drueben.Kennung()), eineBelegnachricht(urkunde, drueben)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if len(probe.bekomen) != 1 || probe.bekomen[0].Unterschrift != klein {
		t.Errorf("abgelegt wurde %+v, erwartet die Unterschrift %q", probe.bekomen, klein)
	}
}
