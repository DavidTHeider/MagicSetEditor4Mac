package app

// **Wie eine Verbindung über fremde Netze geht** — NET-02.AK-03
// (#1706):
// „Direkter Weg, Relay und begründeter Ausfall sind nachvollziehbar".
//
// Bis hierhin sagte der Kasten nur „läuft". Ob ein Gegenüber direkt oder
// über ein fremdes Relais angebunden war, sah niemand — und damit auch
// nicht, warum eine Partie zäh lief.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
)

// wegehorcher ist ein Probehorcher, der Wege nennt.
type wegehorcher struct {
	probefernhorcher

	wege []fernleitung.Fernweg
}

func (w *wegehorcher) Wege() []fernleitung.Fernweg { return w.wege }

func TestDerFernstandNenntDieWegeMitIhrerArt(t *testing.T) {
	t.Parallel()

	horcher := &wegehorcher{
		probefernhorcher: probefernhorcher{einladung: "mse4mac:1:tc1abc"},
		wege: []fernleitung.Fernweg{
			{Relais: "fra"},
			{},
			{Direkt: "203.0.113.7:41641", Relais: "fra"},
		},
	}

	anwendung, ctx := eineFernhorchende(t, horcher, nil)
	t.Cleanup(anwendung.fernhorchenBeenden)

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	wege := anwendung.FernHorchStand(ctx).Stand.Wege

	// **Direkt zuerst, dann das Relais, dann, was noch entsteht** — eine
	// Liste, die bei jedem Nachladen umspränge, läse niemand.
	erwartet := []FernwegDTO{
		{Art: "direkt", Adresse: "203.0.113.7:41641"},
		{Art: "relais", Adresse: "fra"},
		{Art: "aufbau"},
	}

	if len(wege) != len(erwartet) {
		t.Fatalf("Wege: %+v", wege)
	}

	for i := range erwartet {
		if wege[i] != erwartet[i] {
			t.Errorf("Weg %d ist %+v, erwartet %+v", i, wege[i], erwartet[i])
		}
	}
}

// **Ohne Horchen gibt es keine Wege** — und nach dem Beenden auch nicht mehr.
func TestNachDemBeendenSindDieWegeFort(t *testing.T) {
	t.Parallel()

	horcher := &wegehorcher{
		probefernhorcher: probefernhorcher{einladung: "mse4mac:1:tc1abc"},
		wege:             []fernleitung.Fernweg{{Relais: "fra"}},
	}

	anwendung, ctx := eineFernhorchende(t, horcher, nil)

	if wege := anwendung.FernHorchStand(ctx).Stand.Wege; len(wege) != 0 {
		t.Errorf("vor dem Start: %+v", wege)
	}

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	// **Die Gegenprobe, dass jemand zuhört** (§3): während es läuft, steht
	// der Weg da.
	if wege := anwendung.FernHorchStand(ctx).Stand.Wege; len(wege) != 1 {
		t.Fatalf("während es läuft: %+v", wege)
	}

	if antwort := anwendung.FernHorchBeenden(ctx); antwort.Stand.Wege == nil ||
		len(antwort.Stand.Wege) != 0 {
		t.Errorf("nach dem Beenden: %#v", antwort.Stand.Wege)
	}
}

// zaehlhorcher nennt keine Wege, aber offene Ströme — so wie tailcat v0.6.0,
// dessen `Status()` nie eine Gegenstelle trug (SEC-21, #2162), und wie v0.7.0
// zwischen Anbinden und erstem Weg.
type zaehlhorcher struct {
	probefernhorcher

	angebunden int
}

func (z *zaehlhorcher) Angebunden() int { return z.angebunden }

// **Wer angebunden ist, steht da, auch wenn der Weg nicht zu ermitteln ist** —
// SEC-21 (#2162).
// Mit leerer Wegliste sagte der Kasten „noch hat niemand diese Einladung
// benutzt", während eine Verbindung offen war.
func TestDerFernstandNenntDieOffenenStroemeOhneWeg(t *testing.T) {
	t.Parallel()

	horcher := &zaehlhorcher{
		probefernhorcher: probefernhorcher{einladung: "mse4mac:1:tc1abc"},
		angebunden:       2,
	}

	anwendung, ctx := eineFernhorchende(t, horcher, nil)
	t.Cleanup(anwendung.fernhorchenBeenden)

	if stand := anwendung.FernHorchStand(ctx).Stand; stand.Angebunden != 0 {
		t.Errorf("vor dem Start: %d angebunden", stand.Angebunden)
	}

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	stand := anwendung.FernHorchStand(ctx).Stand
	if stand.Angebunden != 2 || len(stand.Wege) != 0 {
		t.Errorf("während es läuft: %d angebunden, Wege %+v — erwartet 2 und keine", stand.Angebunden, stand.Wege)
	}
}
