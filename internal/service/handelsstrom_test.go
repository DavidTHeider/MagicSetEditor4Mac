package service_test

// **Verlust, Duplikate und Umordnung** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Die Abnahme des Issues an einem Strom, der sich schlecht benimmt.** Ein
// Netz verliert, verdoppelt und ordnet um; jeder dieser Fälle muss einen
// **bestimmten** Zustand ergeben und keine zweite Wirkung.
//
// **Und die Wiederaufnahme**: nach einem Abriss wird ausgetauscht, wie weit
// beide sind, und nachgeliefert, was fehlt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ausgangsprobe führt den Ausgang im Arbeitsspeicher.
type ausgangsprobe struct {
	zeilen []service.Ausgangsnachricht
	klagt  error
}

func (a *ausgangsprobe) Einreihen(
	_ context.Context, zeile service.Ausgangsnachricht,
) error {
	if a.klagt != nil {
		return a.klagt
	}

	for _, dalag := range a.zeilen {
		if dalag.NachrichtID == zeile.NachrichtID {
			return nil
		}
	}

	a.zeilen = append(a.zeilen, zeile)

	return nil
}

func (a *ausgangsprobe) Offene(
	_ context.Context, _ string,
) ([]service.Ausgangsnachricht, error) {
	if a.klagt != nil {
		return nil, a.klagt
	}

	return a.zeilen, nil
}

func (a *ausgangsprobe) Bestaetigen(_ context.Context, nachrichtID string) error {
	if a.klagt != nil {
		return a.klagt
	}

	for stelle, zeile := range a.zeilen {
		if zeile.NachrichtID == nachrichtID {
			a.zeilen = append(a.zeilen[:stelle], a.zeilen[stelle+1:]...)

			return nil
		}
	}

	return domain.ErrNotFound
}

// einStrom stellt eine Post samt Ausgang auf.
func einStrom(briefkasten *briefkastenprobe, ausgang *ausgangsprobe) *service.Handelspost {
	return service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusgang(ausgang)
}

// **Ein Strom, der sich schlecht benimmt, ergibt denselben Zustand.**
func TestVerlustDuplikateUndUmordnung(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}
	post := einStrom(briefkasten, &ausgangsprobe{})

	nachricht := einePost(einNachfolger(), nil)

	// **Dreimal dieselbe Zustellung** — das Netz hat verdoppelt.
	for lauf := range 3 {
		antwort, fehler := post.Empfangen(context.Background(), dasGegenueber, nachricht)
		if fehler != nil {
			t.Fatalf("der %d. Lauf: %v", lauf+1, fehler)
		}

		erwartet := domain.ZustellungBekannt
		if lauf == 0 {
			erwartet = domain.ZustellungNeu
		}

		if antwort.Art != erwartet {
			t.Errorf("der %d. Lauf gilt als %q", lauf+1, antwort.Art)
		}
	}

	if len(briefkasten.angebote) != 1 {
		t.Errorf("es wurden %d Wirkungen geschrieben", len(briefkasten.angebote))
	}
}

// **Was hinausgeht, liegt im Ausgang, bis es bestätigt ist.**
func TestWasHinausgehtLiegtImAusgang(t *testing.T) {
	t.Parallel()

	ausgang := &ausgangsprobe{}
	post := einStrom(&briefkastenprobe{}, ausgang)

	nachricht := einePost(einNachfolger(), nil)

	if fehler := post.Senden(context.Background(), dasGegenueber,
		nachricht); fehler != nil {
		t.Fatalf("Senden: %v", fehler)
	}

	// **Zweimal dasselbe ergibt eine Zeile** — der Ausgang ist idempotent.
	if fehler := post.Senden(context.Background(), dasGegenueber,
		nachricht); fehler != nil {
		t.Fatalf("das zweite Senden: %v", fehler)
	}

	if len(ausgang.zeilen) != 1 {
		t.Fatalf("im Ausgang liegen %d Zeilen", len(ausgang.zeilen))
	}

	if fehler := post.Bestaetigt(context.Background(),
		nachricht.Umschlag.NachrichtID); fehler != nil {
		t.Fatalf("Bestaetigt: %v", fehler)
	}

	if len(ausgang.zeilen) != 0 {
		t.Errorf("nach der Bestätigung liegt noch %+v", ausgang.zeilen)
	}
}

// **Die Wiederaufnahme liefert nach, was fehlt** — und nichts darüber hinaus.
func TestDieWiederaufnahmeLiefertNach(t *testing.T) {
	t.Parallel()

	ausgang := &ausgangsprobe{}
	post := einStrom(&briefkastenprobe{}, ausgang)
	ctx := context.Background()

	for folge := int64(1); folge <= 3; folge++ {
		nachricht := einePost(einNachfolger(), func(u *domain.Handelsumschlag) {
			u.NachrichtID = "n-" + string(rune('0'+folge))
			u.Folge = folge
		})

		if fehler := post.Senden(ctx, dasGegenueber, nachricht); fehler != nil {
			t.Fatalf("Senden: %v", fehler)
		}
	}

	meiner := domain.Handelsfortschritt{
		Vorgang: derPostvorgang, Revision: 2, Abdruck: "gleich", Folge: 3,
	}

	seiner := domain.Handelsfortschritt{
		Vorgang: derPostvorgang, Revision: 2, Abdruck: "gleich", Folge: 1,
	}

	nachzuliefern, klaerung, fehler := post.Wiederaufnehmen(
		ctx, dasGegenueber, meiner, seiner)
	if fehler != nil {
		t.Fatalf("Wiederaufnehmen: %v", fehler)
	}

	if klaerung {
		t.Fatal("es kam eine Klärung")
	}

	if len(nachzuliefern) != 2 {
		t.Fatalf("nachzuliefern sind %d Nachrichten", len(nachzuliefern))
	}

	for _, zeile := range nachzuliefern {
		if zeile.Folge < 2 {
			t.Errorf("geliefert wird auch Folge %d", zeile.Folge)
		}
	}
}

// **Ein widersprüchlicher Verlauf liefert nichts nach.**
func TestEinWiderspruchLiefertNichtsNach(t *testing.T) {
	t.Parallel()

	ausgang := &ausgangsprobe{}
	post := einStrom(&briefkastenprobe{}, ausgang)
	ctx := context.Background()

	if fehler := post.Senden(ctx, dasGegenueber,
		einePost(einNachfolger(), nil)); fehler != nil {
		t.Fatalf("Senden: %v", fehler)
	}

	meiner := domain.Handelsfortschritt{
		Vorgang: derPostvorgang, Revision: 2, Abdruck: "meiner", Folge: 1,
	}

	seiner := domain.Handelsfortschritt{
		Vorgang: derPostvorgang, Revision: 2, Abdruck: "seiner", Folge: 0,
	}

	nachzuliefern, klaerung, fehler := post.Wiederaufnehmen(
		ctx, dasGegenueber, meiner, seiner)
	if fehler != nil {
		t.Fatalf("Wiederaufnehmen: %v", fehler)
	}

	if !klaerung {
		t.Fatal("der Widerspruch kam nicht als Klärung")
	}

	if len(nachzuliefern) != 0 {
		t.Errorf("es wurde trotzdem geliefert: %+v", nachzuliefern)
	}
}

// **Ohne Ausgang geht nichts hinaus** — und das sagt sich, statt still zu
// verschwinden.
func TestOhnAusgangGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if fehler := post.Senden(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler == nil {
		t.Error("das Senden kam ohne Ausgang durch")
	}
}

// **Ohne Ausgang geht auch kein Bestätigen und keine Wiederaufnahme** — aus
// demselben Grund: was nie hinausging, lässt sich nicht bestätigen.
func TestOhneAusgangGehtAuchSonstNichts(t *testing.T) {
	t.Parallel()

	post := service.NeueHandelspost(&briefkastenprobe{},
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if fehler := post.Bestaetigt(context.Background(), "n-1"); fehler == nil {
		t.Error("das Bestätigen kam ohne Ausgang durch")
	}

	if _, _, fehler := post.Wiederaufnehmen(context.Background(), dasGegenueber,
		domain.Handelsfortschritt{}, domain.Handelsfortschritt{}); fehler == nil {
		t.Error("die Wiederaufnahme kam ohne Ausgang durch")
	}
}

// **Ein Ausgang, der klagt, klagt durch** — bei der Wiederaufnahme.
func TestEinKlagenderAusgangKlagtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("der Ausgang ist zu")
	post := einStrom(&briefkastenprobe{}, &ausgangsprobe{klagt: gestolpert})

	gleich := domain.Handelsfortschritt{
		Vorgang: derPostvorgang, Revision: 1, Abdruck: "gleich", Folge: 0,
	}

	if _, _, fehler := post.Wiederaufnehmen(context.Background(), dasGegenueber,
		gleich, gleich); !errors.Is(fehler, gestolpert) {
		t.Errorf("es kam %v", fehler)
	}
}
