package service_test

// **Eine Handelsnachricht über eine echte Leitung** — Werkstück `T5` von
// Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Der Weg durch die Schleife**, und zwar über `internal/leitung` und nicht
// über ein nachgebautes Doppel: das Issue verlangt „Loopback … ergeben
// äquivalente Angebote und Zustände", und eine selbstgebaute Röhre prüfte den
// eigenen Nachbau.
//
// **Dass ein unlesbarer Rumpf ein Befund ist und kein Absturz** (§8).
//
// **Und dass erst eingereiht und dann gesendet wird.** Wer umgekehrt vorginge,
// verlöre die Nachricht, wenn der Prozess dazwischen endet.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineNachrichtGehtDurchDieSchleife(t *testing.T) {
	t.Parallel()

	hier, dort := leitung.Schleife(service.Handelskanal)

	ausgang := &ausgangsprobe{}
	briefkasten := &briefkastenprobe{}

	sender := einStrom(&briefkastenprobe{}, ausgang)
	empfaenger := einStrom(briefkasten, &ausgangsprobe{})

	ctx := context.Background()
	nachricht := einePost(einNachfolger(), nil)

	fertig := make(chan error, 1)

	go func() {
		fertig <- sender.SendeUeber(ctx, hier, dasGegenueber, nachricht)
	}()

	antwort, fehler := empfaenger.EmpfangeUeber(ctx, dort, dasGegenueber)
	if fehler != nil {
		t.Fatalf("EmpfangeUeber: %v", fehler)
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("SendeUeber: %v", fehler)
	}

	if antwort.Art != domain.ZustellungNeu {
		t.Errorf("die Zustellung gilt als %q", antwort.Art)
	}

	if len(briefkasten.angebote) != 1 {
		t.Errorf("drüben wurden %d Wirkungen geschrieben", len(briefkasten.angebote))
	}

	// **Und hier liegt sie im Ausgang**, bis die andere Seite bestätigt.
	if len(ausgang.zeilen) != 1 {
		t.Errorf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}
}

// **Ein unlesbarer Rumpf ist ein Befund** (§8).
func TestEinUnlesbarerRumpfIstEinBefund(t *testing.T) {
	t.Parallel()

	hier, dort := leitung.Schleife(service.Handelskanal)
	ctx := context.Background()

	go func() { _ = hier.Sende(ctx, []byte("kein JSON")) }() //nolint:errcheck // der Lauf misst die Gegenseite

	post := einStrom(&briefkastenprobe{}, &ausgangsprobe{})

	if _, fehler := post.EmpfangeUeber(ctx, dort, dasGegenueber); fehler == nil {
		t.Error("der krumme Rumpf kam durch")
	}
}

// **Was verpackt wurde, lässt sich wieder auspacken** — und trägt dabei
// dieselbe Gestalt.
func TestWasVerpacktWurdeKommtGleichAn(t *testing.T) {
	t.Parallel()

	nachricht := einePost(einNachfolger(), nil)

	roh, fehler := json.Marshal(nachricht)
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	var zurueck service.Handelsnachricht
	if fehler := json.Unmarshal(roh, &zurueck); fehler != nil {
		t.Fatalf("Unmarshal: %v", fehler)
	}

	if zurueck.Umschlag.Abdruck != nachricht.Umschlag.Abdruck {
		t.Errorf("der Abdruck kam als %q an", zurueck.Umschlag.Abdruck)
	}

	// **Der Abdruck des Angebots entsteht aus seinen Feldern** — kommt eines
	// nicht mit, stimmt er drüben nicht mehr, und die Nachricht wird zu Recht
	// abgewiesen. Das ist die Probe darauf, dass die Gestalt vollständig ist.
	if zurueck.Angebot.Abdruck() != nachricht.Angebot.Abdruck() {
		t.Errorf("das Angebot kam mit einem anderen Abdruck an")
	}
}

// **Ein Sender ohne Ausgang schickt nichts** — er reiht zuerst ein, und wenn
// das nicht geht, geht auch nichts hinaus.
func TestOhneAusgangGehtNichtsUeberDieStrecke(t *testing.T) {
	t.Parallel()

	hier, _ := leitung.Schleife(service.Handelskanal)

	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if fehler := post.SendeUeber(context.Background(), hier, dasGegenueber,
		einePost(einNachfolger(), nil)); fehler == nil {
		t.Error("das Senden kam ohne Ausgang durch")
	}
}

// **Eine Strecke, die zu ist, meldet sich** — und der Empfänger gibt den
// Fehler weiter, statt eine leere Nachricht zu behandeln.
func TestEineGeschlosseneStreckeMeldetSich(t *testing.T) {
	t.Parallel()

	hier, dort := leitung.Schleife(service.Handelskanal)

	if fehler := hier.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	post := einStrom(&briefkastenprobe{}, &ausgangsprobe{})

	if _, fehler := post.EmpfangeUeber(context.Background(), dort,
		dasGegenueber); fehler == nil {
		t.Error("die geschlossene Strecke kam durch")
	}
}

// **Was im Ausgang liegt, geht über die Leitung hinaus** —
// #1634.
//
// # Warum das ein eigener Weg ist und nicht [Handelspost.SendeUeber]
//
// **Weil jener **einreiht** und dann schickt.** Was schon im Ausgang liegt,
// noch einmal einzureihen wäre bestenfalls wirkungslos — der Ausgang ist
// idempotent — und schlimmstenfalls eine zweite Zeile mit neuer Kennung.
// Hier wird **getragen**, was dasteht.
func TestWasImAusgangLiegtGehtUeberDieLeitungHinaus(t *testing.T) {
	t.Parallel()

	hier, dort := leitung.Schleife(service.Handelskanal)

	ausgang := &ausgangsprobe{}
	sender := einStrom(&briefkastenprobe{}, ausgang)

	briefkasten := &briefkastenprobe{}
	empfaenger := einStrom(briefkasten, &ausgangsprobe{})

	ctx := context.Background()
	nachricht := einePost(einNachfolger(), nil)

	if fehler := sender.Senden(ctx, dasGegenueber, nachricht); fehler != nil {
		t.Fatalf("Senden: %v", fehler)
	}

	gelesen := make(chan service.Zustellungsantwort, 1)

	go func() {
		antwort, _ := empfaenger.EmpfangeUeber( //nolint:errcheck // unten geprüft
			ctx, dort, dasGegenueber)
		gelesen <- antwort
	}()

	getragen, fehler := sender.Nachliefern(ctx, hier, dasGegenueber)
	if fehler != nil {
		t.Fatalf("Nachliefern: %v", fehler)
	}

	if getragen != 1 {
		t.Fatalf("getragen wurden %d Nachrichten", getragen)
	}

	if antwort := <-gelesen; antwort.Art != domain.ZustellungNeu {
		t.Errorf("die Zustellung gilt als %q", antwort.Art)
	}

	// **Der Ausgang bleibt voll** — erst die **Bestätigung** räumt ihn, und
	// die kommt von drüben. Wer hier räumte, verlöre die Nachricht bei einem
	// Abriss zwischen Senden und Ankommen.
	if len(ausgang.zeilen) != 1 {
		t.Errorf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}
}

// **Ein leerer Ausgang trägt nichts und klagt nicht.**
func TestEinLeererAusgangTraegtNichts(t *testing.T) {
	t.Parallel()

	hier, _ := leitung.Schleife(service.Handelskanal)

	post := einStrom(&briefkastenprobe{}, &ausgangsprobe{})

	getragen, fehler := post.Nachliefern(
		context.Background(), hier, dasGegenueber)
	if fehler != nil {
		t.Fatalf("Nachliefern: %v", fehler)
	}

	if getragen != 0 {
		t.Errorf("getragen wurden %d Nachrichten", getragen)
	}
}

// **Und eine Post ohne Ausgang trägt gar nichts.**
//
// **Dieselbe Absage wie bei [Handelspost.Senden]**: eine Post ohne Ausgang
// kann empfangen und nicht senden, und sie sagt es, statt die Nachricht
// verschwinden zu lassen.
func TestEinePostOhneAusgangTraegtNichts(t *testing.T) {
	t.Parallel()

	hier, _ := leitung.Schleife(service.Handelskanal)

	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := post.Nachliefern(
		context.Background(), hier, dasGegenueber); fehler == nil {
		t.Error("die Post ohne Ausgang trug etwas")
	}
}

// stolperstrecke scheitert beim n-ten Senden.
//
// **Mit einer Zahl und nicht mit einem Schalter**: geprüft wird, dass die
// **Zahl der getragenen** stimmt, wenn es mittendrin abreißt — ein Doppel,
// das immer scheitert, hätte dafür nur die Null zu bieten.
type stolperstrecke struct {
	traegt    int
	geschickt int
}

func (s *stolperstrecke) Sende(_ context.Context, _ []byte) error {
	if s.geschickt >= s.traegt {
		return errors.New("die Leitung riss ab")
	}

	s.geschickt++

	return nil
}

func (s *stolperstrecke) Empfange(context.Context) ([]byte, error) {
	return nil, errors.New("hier wird nur gesendet")
}

// **Ein unlesbarer Ausgang trägt nichts und sagt, warum.**
func TestEinUnlesbarerAusgangTraegtNichts(t *testing.T) {
	t.Parallel()

	post := einStrom(&briefkastenprobe{},
		&ausgangsprobe{klagt: errors.New("der Ausgang ist weg")})

	if _, fehler := post.Nachliefern(
		context.Background(), &stolperstrecke{traegt: 9}, dasGegenueber,
	); fehler == nil {
		t.Error("der unlesbare Ausgang klagte nicht")
	}
}

// **Ein Abriss mitten in der Reihe nennt die Zahl, die durchkam.**
//
// **Und nicht null.** Wer bei einem Abriss die Zahl verwürfe, ließe den
// Aufrufer glauben, es sei nichts hinausgegangen — und die Gegenseite hätte
// es trotzdem.
func TestEinAbrissMittendrinNenntDieZahl(t *testing.T) {
	t.Parallel()

	ausgang := &ausgangsprobe{}
	post := einStrom(&briefkastenprobe{}, ausgang)

	ctx := context.Background()

	// **Zwei Nachrichten im Ausgang** — die erste geht, die zweite nicht.
	for _, revision := range []domain.Handelsangebot{
		einNachfolger(), einNachfolger().Fortgeschrieben(
			[]domain.Handelsposition{{
				Von: "bich", Printing: "a1b2", Menge: 3,
				Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
			}}),
	} {
		nachricht := einePost(revision, func(u *domain.Handelsumschlag) {
			u.NachrichtID = "n-" + revision.Abdruck()
			u.Abdruck = revision.Abdruck()
		})

		if fehler := post.Senden(ctx, dasGegenueber, nachricht); fehler != nil {
			t.Fatalf("Senden: %v", fehler)
		}
	}

	getragen, fehler := post.Nachliefern(
		ctx, &stolperstrecke{traegt: 1}, dasGegenueber)

	if fehler == nil {
		t.Fatal("der Abriss blieb ungemeldet")
	}

	if getragen != 1 {
		t.Errorf("gemeldet wurden %d getragene", getragen)
	}
}
