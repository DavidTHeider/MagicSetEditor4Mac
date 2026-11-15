package service_test

// **Ein Abschluss nimmt seine Urkunde mit** —
// #1650.
//
// # Wogegen hier geprüft wird
//
// **Gegen eine Nachricht, die einen Abschluss meldet und nichts belegt.** Die
// Art `beleg` ging schon hinaus, sobald ein Vorgang auf `beidseitigVerbucht`
// stand — nur trug sie den Stand und keine Unterschrift. Drüben blieb
// `beidseitig` damit falsch, und das war richtig: was nicht belegt ist, wird
// nicht behauptet.
//
// **Und gegen eine Urkunde, die Lokales mitnimmt.** Bestandskennung,
// Lagerort und Notiz gehen einen Fremden nichts an.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einAbgeschlossenerStand ist ein Vorgang, aus dem ein Beleg hinausgeht.
func einAbgeschlossenerStand() domain.Handelsangebot {
	stand := derHeutigeStand()
	stand.Zustand = domain.HandelBeidseitigVerbucht

	return stand
}

// einBelegversand stellt den Versand samt Belegquelle auf.
func einBelegversand(t *testing.T, quelle *gegenbelegprobe) (
	*service.Handelsversand, *ausgangsprobe,
) {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	stand := einAbgeschlossenerStand()

	ausgang := &ausgangsprobe{}
	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: stand},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusgang(ausgang)

	versand := service.NeuerHandelsversand(post, standprobe{stand: stand},
		&folgenprobe{}, ausweisprobe{ich: ich}).
		MitUhr(func() time.Time { return dieVersandzeit }).
		MitKennungen(func() string { return "n-fest" }).
		MitBelegen(quelle)

	return versand, ausgang
}

// eineBelegquelle hält einen Beleg samt einer Zeile mit Lokalem darin.
func eineBelegquelle() *gegenbelegprobe {
	return &gegenbelegprobe{
		stand: &domain.Belegstand{
			Vorgang:      derPostvorgang,
			Abdruck:      einAbgeschlossenerStand().Abdruck(),
			Unterschrift: "deadbeef",
		},
		zeilen: []domain.Journalzeile{{
			Art: domain.JournalAbgang, Printing: "p-1", Menge: 1,
			Zustand: domain.ZustandNearMint, Sprache: "de",
			Finish: domain.FinishNonfoil,

			// **Das bleibt hier.**
			ExemplarID: 7, Lagerort: "Kiste 4", Notiz: "zurückgelegt",
		}},
	}
}

// **Ein Beleg geht mit seiner Urkunde hinaus — und ohne das Lokale.**
func TestEinBelegNimmtSeineUrkundeMit(t *testing.T) {
	t.Parallel()

	versand, ausgang := einBelegversand(t, eineBelegquelle())

	if _, fehler := versand.Schicken(
		context.Background(), derPostvorgang, dasGegenueber); fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	if len(ausgang.zeilen) != 1 {
		t.Fatalf("eingereiht wurden %d", len(ausgang.zeilen))
	}

	nachricht := ausgang.zeilen[0].Nachricht

	if nachricht.Umschlag.Art != domain.NachrichtBeleg {
		t.Fatalf("hinaus ging die Art %q", nachricht.Umschlag.Art)
	}

	if nachricht.Beleg == nil {
		t.Fatal("die Urkunde fehlt")
	}

	if nachricht.Beleg.Unterschrift != "deadbeef" {
		t.Errorf("die Unterschrift ist %q", nachricht.Beleg.Unterschrift)
	}

	zeile := nachricht.Beleg.Beleg.Zeilen[0]
	if zeile.ExemplarID != 0 || zeile.Lagerort != "" || zeile.Notiz != "" {
		t.Errorf("hinaus ginge %+v", zeile)
	}
}

// **Ohne Beleg geht die Nachricht trotzdem** — sie sagt den Zustand, und
// der steht auch ohne Urkunde.
func TestEinBelegOhneUrkundeGehtTrotzdemHinaus(t *testing.T) {
	t.Parallel()

	quelle := eineBelegquelle()
	quelle.stand = nil

	versand, ausgang := einBelegversand(t, quelle)

	if _, fehler := versand.Schicken(
		context.Background(), derPostvorgang, dasGegenueber); fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	if ausgang.zeilen[0].Nachricht.Beleg != nil {
		t.Error("es ging eine Urkunde hinaus, die es nicht gibt")
	}
}

// **Und ohne Belegquelle ebenso** — ein Versand ohne sie ist vollständig.
func TestEinVersandOhneBelegquelleSchicktDenStand(t *testing.T) {
	t.Parallel()

	versand, ausgang, _ := einVersand(t, einAbgeschlossenerStand())

	if _, fehler := versand.Schicken(
		context.Background(), derPostvorgang, dasGegenueber); fehler != nil {
		t.Fatalf("Schicken: %v", fehler)
	}

	if ausgang.zeilen[0].Nachricht.Beleg != nil {
		t.Error("es ging eine Urkunde hinaus, die niemand hergab")
	}
}

// **Ein klemmender Bestand hält den Beleg an.**
//
// **Und nicht nur den Beleg**: eine Nachricht, die behauptete, sie sei
// vollständig, während ihre Urkunde an einem Lesefehler hängenblieb, wäre
// die halbe Zusage aus §10.
func TestEinKlemmenderBestandHaeltDenBelegAn(t *testing.T) {
	t.Parallel()

	quelle := eineBelegquelle()
	quelle.klagt = errors.New("der Bestand klemmt")

	versand, ausgang := einBelegversand(t, quelle)

	if _, fehler := versand.Schicken(
		context.Background(), derPostvorgang, dasGegenueber); fehler == nil {
		t.Fatal("der Lesefehler kam nicht an")
	}

	if len(ausgang.zeilen) != 0 {
		t.Errorf("eingereiht wurden %d", len(ausgang.zeilen))
	}
}

// **Ein Journal, das klemmt, hält den Beleg an** — der zweite der beiden
// Wege, auf denen die Urkunde den Bestand fragt.
func TestEinKlemmendesJournalHaeltDenBelegAn(t *testing.T) {
	t.Parallel()

	quelle := eineBelegquelle()
	quelle.journalklage = errors.New("das Journal klemmt")

	versand, ausgang := einBelegversand(t, quelle)

	if _, fehler := versand.Schicken(
		context.Background(), derPostvorgang, dasGegenueber); fehler == nil {
		t.Fatal("der Lesefehler kam nicht an")
	}

	if len(ausgang.zeilen) != 0 {
		t.Errorf("eingereiht wurden %d", len(ausgang.zeilen))
	}
}
