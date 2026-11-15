package app

// **Über fremde Netze erreichbar sein** — Werkstück N3 von Phase 5.4
// (#1456).
//
// **Über eine eingesetzte Naht** (§3): ein Lauf, der wirklich horchte, holte
// eine DERP-Karte aus dem Internet und legte einen Tunnel an.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probefernhorcher gibt eine Einladung und nie eine Verbindung.
type probefernhorcher struct {
	einladung domain.Einladung
	zu        bool
}

func (p *probefernhorcher) Nimm(ctx context.Context) (service.Verbindung, error) {
	<-ctx.Done()

	return nil, domain.MitGrund(
		domain.ErrNotFound, ctx.Err(), "es kam keine Verbindung herein")
}

func (p *probefernhorcher) Schliesse() error {
	p.zu = true

	return nil
}

func (p *probefernhorcher) Adresse() string { return string(p.einladung) }

func (p *probefernhorcher) Einladung() domain.Einladung { return p.einladung }

func (p *probefernhorcher) Wege() []fernleitung.Fernweg { return nil }

func (p *probefernhorcher) Angebunden() int { return 0 }

// eineFernhorchende baut eine Anwendung mit Identität und eingesetzter Naht.
func eineFernhorchende(
	t *testing.T, horcher Fernhorcher, fehler error,
) (*App, context.Context) {
	t.Helper()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		return horcher, fehler
	})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	return anwendung, ctx
}

// **Von selbst ist niemand über fremde Netze erreichbar.**
func TestVonSelbstGibtEsKeineEinladung(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil)

	stand := anwendung.FernHorchStand(ctx).Stand
	if stand.Laeuft || stand.Einladung != "" {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Und danach steht die Einladung da.**
func TestNachDemStartenStehtDieEinladungDa(t *testing.T) {
	t.Parallel()

	horcher := &probefernhorcher{einladung: "mse4mac:1:tc1abc"}
	anwendung, ctx := eineFernhorchende(t, horcher, nil)

	t.Cleanup(anwendung.fernhorchenBeenden)

	antwort := anwendung.FernHorchStarten(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	if !antwort.Stand.Laeuft {
		t.Error("der Stand sagt, es läuft nicht")
	}

	if antwort.Stand.Einladung != "mse4mac:1:tc1abc" {
		t.Errorf("die Einladung kam als %q", antwort.Stand.Einladung)
	}
}

// **Ein zweiter Start ist ein Konflikt mit eigenem Rat.**
func TestZweimalFernhorchenIstEinKonflikt(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil)

	t.Cleanup(anwendung.fernhorchenBeenden)

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("der erste Start: %+v", antwort.Fehler)
	}

	antwort := anwendung.FernHorchStarten(ctx)
	if antwort.Fehler == nil {
		t.Fatal("der zweite Start kam durch")
	}

	if antwort.Fehler.Code != CodeKonflikt {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}

	// **Und die Antwort trägt trotzdem eine leere Einladung statt `null`**
	// (§3).
	if antwort.Stand.Einladung != "" {
		t.Errorf("die Fehlerantwort trägt eine Einladung: %q",
			antwort.Stand.Einladung)
	}
}

// **Beenden entwertet die Einladung** — der Serverschlüssel wird nicht
// aufbewahrt, und das ist Widerruf ohne Zähler.
func TestBeendenEntwertetDieEinladung(t *testing.T) {
	t.Parallel()

	horcher := &probefernhorcher{einladung: "mse4mac:1:tc1abc"}
	anwendung, ctx := eineFernhorchende(t, horcher, nil)

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	antwort := anwendung.FernHorchBeenden(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("FernHorchBeenden: %+v", antwort.Fehler)
	}

	if antwort.Stand.Laeuft || antwort.Stand.Einladung != "" {
		t.Errorf("der Stand ist %+v", antwort.Stand)
	}

	if !horcher.zu {
		t.Error("der Fernhorcher steht noch offen")
	}
}

// **Ein Beenden ohne Start ist kein Fehler.**
func TestBeendenOhneStartIstKeinFehler(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil)

	if antwort := anwendung.FernHorchBeenden(ctx); antwort.Fehler != nil {
		t.Errorf("FernHorchBeenden klagte: %+v", antwort.Fehler)
	}
}

// **Ein Horcher, der nicht aufgeht, wird gemeldet.**
func TestEinFernhorcherDerNichtAufgehtWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, nil, errors.New("kein DERP erreichbar"))

	antwort := anwendung.FernHorchStarten(ctx)
	if antwort.Fehler == nil {
		t.Fatal("der Start kam durch")
	}
}

// **Und ohne Identität gibt es keine Einladung** — dieselbe Linie wie beim
// Horchposten aus N2.
func TestOhneIdentitaetGibtEsKeineEinladung(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		t.Error("es wurde gehorcht, obwohl es keine Identität gibt")

		return nil, nil
	})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler == nil {
		t.Fatal("der Start kam ohne Identität durch")
	}
}

// **Eine Einladung im Adressfeld nimmt den Weg über fremde Netze.**
//
// # Was hier gemessen wird
//
// **Die Weiche, und nur sie.** Im Adressfeld eines Kontakts steht entweder
// eine Adresse im eigenen Netz oder eine Einladung; der Wähler muss sie
// auseinanderhalten.
//
// **Und der Lauf hat dabei eine Naht gefunden, die nicht durchging**: der
// Schreibweg (`PeerkontaktAdresse` → `domain.PruefeAdresse`) wies eine
// Einladung ab, weil sie „keinen Wirt und Port" hat. Die Weiche im Wähler
// nützte nichts, solange nichts sie erreichen konnte — die Naht aus §10, und
// gefunden hat sie ein Lauf und nicht das Lesen.
func TestEineEinladungImAdressfeldNimmtDenFernweg(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil)

	var gewaehlt domain.Einladung

	anwendung.MitFernwaehler(func(
		_ context.Context, einladung domain.Einladung,
		_ peer.Identitaet, _ peer.Kennung,
	) (service.Verbindung, error) {
		gewaehlt = einladung

		return nil, errors.New("hier endet der Lauf")
	})

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(fremd.Kennung()), "Der andere"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	// **Der Schreibweg nimmt sie an** — das ist die Hälfte der Zusage.
	if antwort := anwendung.PeerkontaktAdresse(
		ctx, string(fremd.Kennung()),
		"mse4mac:1:tc1abc"); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	if _, waehlfehler := (adresswaehler{app: anwendung}).Waehle(
		ctx, fremd.Kennung(), ""); waehlfehler == nil {
		t.Fatal("der Wähler kam durch")
	}

	// **Und die andere Hälfte**: es ging wirklich über den Fernweg.
	if gewaehlt != "mse4mac:1:tc1abc" {
		t.Errorf("gewählt wurde %q", gewaehlt)
	}
}

// **Und eine Adresse im eigenen Netz nimmt ihn nicht.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie sagte der Lauf nur, dass
// überhaupt etwas über den Fernweg geht.
func TestEineAdresseImEigenenNetzNimmtDenFernwegNicht(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(
		t, &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil)

	anwendung.MitFernwaehler(func(
		_ context.Context, _ domain.Einladung,
		_ peer.Identitaet, _ peer.Kennung,
	) (service.Verbindung, error) {
		t.Error("eine Adresse im eigenen Netz ging über den Fernweg")

		return nil, nil
	})

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(fremd.Kennung()), "Der andere"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerkontaktAdresse(
		ctx, string(fremd.Kennung()), "127.0.0.1:1"); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	if _, waehlfehler := (adresswaehler{app: anwendung}).Waehle(
		ctx, fremd.Kennung(), ""); waehlfehler == nil {
		t.Fatal("der Wähler kam durch")
	}
}
