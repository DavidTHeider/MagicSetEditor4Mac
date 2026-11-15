package app

// **Der Horchbetrieb der Anwendung** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass nichts von selbst horcht**, dass ein zweiter Anlauf abgewiesen wird,
// und dass ein **Profilwechsel** es beendet. Das letzte ist die Zusage, um die
// es geht: der Lauscher trägt die Identität des Profils in seinem Zeugnis, und
// wer ihn über den Wechsel hinweg offen ließe, machte das eine Profil unter
// dem Schlüssel des anderen erreichbar.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probelauscher gibt nie eine Verbindung und merkt sich, dass er zu ist.
type probelauscher struct {
	schloss sync.Mutex
	zu      bool

	// wartet hält `Nimm` auf, bis der Kontext endet — so, wie ein echter
	// Lauscher es täte, an dem niemand anklopft.
	klagt error
}

func (p *probelauscher) Nimm(ctx context.Context) (service.Verbindung, error) {
	<-ctx.Done()

	return nil, domain.MitGrund(
		domain.ErrNotFound, ctx.Err(), "es kam keine Verbindung herein")
}

func (p *probelauscher) Schliesse() error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.zu = true

	return p.klagt
}

func (p *probelauscher) geschlossen() bool {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	return p.zu
}

// Adresse gibt eine erfundene, die nicht die gewünschte ist.
//
// **Und das ist der Punkt**: wer `:0` angibt, bekommt einen Port vom
// Betriebssystem, und die Antwort muss den **gebundenen** nennen.
func (p *probelauscher) Adresse() string { return "127.0.0.1:53123" }

// eineHorchende baut eine Anwendung mit Identität und eingesetztem Lauscher.
func eineHorchende(
	t *testing.T, lauscher service.Hereinkommend, fehler error,
) (*App, context.Context) {
	t.Helper()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	anwendung.MitHorcher(func(
		context.Context, string, peer.Identitaet,
	) (service.Hereinkommend, error) {
		return lauscher, fehler
	})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	return anwendung, ctx
}

// **Von selbst horcht nichts.**
func TestVonSelbstHorchtNichts(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	antwort := anwendung.HorchStand(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("HorchStand: %+v", antwort.Fehler)
	}

	if antwort.Stand.Laeuft || antwort.Stand.Adresse != "" {
		t.Fatalf("der Stand ist %+v", antwort.Stand)
	}
}

// **Und danach nennt der Stand die wirkliche Adresse.**
func TestDerStandNenntDieWirklicheAdresse(t *testing.T) {
	t.Parallel()

	lauscher := &probelauscher{}
	anwendung, ctx := eineHorchende(t, lauscher, nil)

	t.Cleanup(anwendung.horchenBeenden)

	antwort := anwendung.HorchStarten(ctx, "127.0.0.1:0")
	if antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if !antwort.Stand.Laeuft || antwort.Stand.Adresse != "127.0.0.1:53123" {
		t.Fatalf("der Stand ist %+v", antwort.Stand)
	}
}

// **Ein zweiter Anlauf wird abgewiesen.**
//
// **Sonst stünden zwei Lauscher da**, und der erste ließe sich nicht mehr
// beenden — sein Abbruch läge in einem Feld, das der zweite überschrieben hat
// (§7, Regel 4).
func TestEinZweiterAnlaufWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(
		ctx, ":41234"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	antwort := anwendung.HorchStarten(ctx, ":41235")
	if antwort.Fehler == nil || antwort.Fehler.Code != "konflikt" {
		t.Fatalf("der zweite Anlauf gab %+v", antwort.Fehler)
	}
}

// **Das Beenden legt den Lauscher wirklich nieder.**
func TestDasBeendenLegtDenLauscherNieder(t *testing.T) {
	t.Parallel()

	lauscher := &probelauscher{}
	anwendung, ctx := eineHorchende(t, lauscher, nil)

	if antwort := anwendung.HorchStarten(
		ctx, ":41234"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	antwort := anwendung.HorchBeenden(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("HorchBeenden: %+v", antwort.Fehler)
	}

	if antwort.Stand.Laeuft {
		t.Fatal("der Stand sagt, es liefe noch")
	}

	if !lauscher.geschlossen() {
		t.Fatal("der Lauscher steht noch offen")
	}
}

// **Ein Beenden ohne Anfang ist kein Fehler.**
//
// **Das Ziel ist erreicht**, und wer zweimal beendet, soll beim zweiten Mal
// keine Meldung bekommen — dieselbe Gestalt wie beim Entziehen eines Zwecks.
func TestEinBeendenOhneAnfangIstKeinFehler(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	if antwort := anwendung.HorchBeenden(ctx); antwort.Fehler != nil {
		t.Fatalf("HorchBeenden: %+v", antwort.Fehler)
	}
}

// **Ein klagender Lauscher hält das Beenden nicht auf.**
func TestEinKlagenderLauscherHaeltDasBeendenNichtAuf(t *testing.T) {
	t.Parallel()

	lauscher := &probelauscher{klagt: errors.New("er klagt")}
	anwendung, ctx := eineHorchende(t, lauscher, nil)

	if antwort := anwendung.HorchStarten(
		ctx, ":41234"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.HorchBeenden(ctx); antwort.Fehler != nil {
		t.Fatalf("HorchBeenden: %+v", antwort.Fehler)
	}

	if anwendung.horchstand().Laeuft {
		t.Fatal("der Stand sagt, es liefe noch")
	}
}

// **Ohne Adresse wird nicht gehorcht, und eine krumme endet.**
func TestOhneOderMitKrummerAdresseWirdNichtGehorcht(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	for name, adresse := range map[string]string{
		"gar keine": "",
		"krumme":    "frag mal David",
	} {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.HorchStarten(ctx, adresse)
			if antwort.Fehler == nil ||
				antwort.Fehler.Code != "validierung" {
				t.Fatalf("es kam %+v", antwort.Fehler)
			}
		})
	}
}

// **Ein Lauscher, der sich nicht öffnen lässt, wird gemeldet.**
func TestEinLauscherDerSichNichtOeffnetWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(
		t, nil, domain.MitGrund(domain.ErrValidation, nil, "der Port ist weg"))

	if antwort := anwendung.HorchStarten(ctx, ":41234"); antwort.Fehler == nil {
		t.Fatal("das Horchen kam ohne Port durch")
	}
}

// **Ein Profilwechsel beendet das Horchen.**
//
// # Die Zusage, für die es diesen Lauf gibt
//
// **Der Lauscher gehört dem Profil, das ihn geöffnet hat.** Er trägt dessen
// Identität in seinem Zeugnis und fragt dessen Kontaktliste; wer ihn über den
// Wechsel hinweg offen ließe, machte das eine Profil unter dem Schlüssel des
// anderen erreichbar.
func TestEinProfilwechselBeendetDasHorchen(t *testing.T) {
	t.Parallel()

	lauscher := &probelauscher{}
	anwendung, ctx := eineHorchende(t, lauscher, nil)

	if antwort := anwendung.HorchStarten(
		ctx, ":41234"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	// **Über denselben Weg wie im Betrieb** — `MitProfilwechsel` setzt die
	// Naht ein, die sonst das Fenster benachrichtigt; ohne sie greift
	// `profilUmstellen` in eine, die es in diesem Lauf nicht gibt.
	anwendung.MitProfilwechsel(wechselprobe{})
	anwendung.profilUmstellen(riegelkontext(t), 2)

	if anwendung.horchstand().Laeuft {
		t.Fatal("nach dem Wechsel horcht es weiter")
	}

	if !lauscher.geschlossen() {
		t.Fatal("der Lauscher steht noch offen")
	}
}
