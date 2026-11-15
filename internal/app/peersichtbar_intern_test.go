package app

// **In der Nähe gefunden werden** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass die Ankündigung am Horchposten hängt** — ohne Port keine Einladung —
// **und dass sie ihn nicht überlebt**: wer aufhört zu horchen, hört auch auf,
// sich anzukündigen, und zwar in dieser Reihenfolge.
//
// **Über eingesetzte Nähte und nicht über das Netz** (§3): ein Lauf, der
// wirklich ankündigte und suchte, fragte das Netz dieses Rechners — und fände,
// was zufällig dort steht.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
)

// probeankuendigung merkt sich, was angekündigt wurde.
type probeankuendigung struct {
	angebot entdeckung.Angebot
	beendet bool
	klagt   error
}

func (p *probeankuendigung) Beende() error {
	p.beendet = true

	return p.klagt
}

// **Ohne Horchposten gibt es keine Ankündigung.**
func TestOhneHorchpostenGibtEsKeineAnkuendigung(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	angemeldet := &probeankuendigung{}
	anwendung.MitEntdeckung(func(
		angebot entdeckung.Angebot,
	) (Ankuendigung, error) {
		angemeldet.angebot = angebot

		return angemeldet, nil
	}, nil)

	antwort := anwendung.SichtbarWerden(ctx, "Davids Rechner")
	if antwort.Fehler == nil || antwort.Fehler.Code != "konflikt" {
		t.Fatalf("es kam %+v", antwort.Fehler)
	}
}

// **Mit Horchposten trägt die Ankündigung dessen Port.**
func TestDieAnkuendigungTraegtDenPortDesHorchpostens(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	angemeldet := &probeankuendigung{}
	anwendung.MitEntdeckung(func(
		angebot entdeckung.Angebot,
	) (Ankuendigung, error) {
		angemeldet.angebot = angebot

		return angemeldet, nil
	}, nil)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	antwort := anwendung.SichtbarWerden(ctx, "Davids Rechner")
	if antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	if !antwort.Stand.Sichtbar {
		t.Fatal("der Stand sagt, es sei nicht sichtbar")
	}

	// **Der Port kommt vom Lauscher** — `probelauscher` nennt 53123.
	if angemeldet.angebot.Port != 53123 {
		t.Errorf("angekündigt wurde Port %d", angemeldet.angebot.Port)
	}

	if angemeldet.angebot.Name != "Davids Rechner" {
		t.Errorf("angekündigt wurde %q", angemeldet.angebot.Name)
	}
}

// **Zweimal ankündigen ist einmal.**
func TestZweimalAnkuendigenIstEinmal(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		return &probeankuendigung{}, nil
	}, nil)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(ctx, "Erst"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	antwort := anwendung.SichtbarWerden(ctx, "Noch mal")
	if antwort.Fehler == nil || antwort.Fehler.Code != "konflikt" {
		t.Fatalf("der zweite Anlauf gab %+v", antwort.Fehler)
	}
}

// **Wer aufhört zu horchen, hört auch auf, sich anzukündigen.**
//
// **Und in dieser Reihenfolge**: wer zuerst den Port schlösse, ließe eine
// Einladung im Netz stehen, die ins Leere führt.
func TestWerAufhoertZuHorchenHoertAuchAufSichAnzukuendigen(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	angemeldet := &probeankuendigung{}
	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		return angemeldet, nil
	}, nil)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(ctx, "Wer"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	antwort := anwendung.HorchBeenden(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("HorchBeenden: %+v", antwort.Fehler)
	}

	if !angemeldet.beendet {
		t.Fatal("die Ankündigung steht noch")
	}

	if antwort.Stand.Sichtbar || antwort.Stand.Laeuft {
		t.Fatalf("der Stand ist %+v", antwort.Stand)
	}
}

// **Eine Ankündigung, die beim Zurücknehmen klagt, hält nichts auf.**
func TestEineKlagendeAnkuendigungHaeltNichtsAuf(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		return &probeankuendigung{klagt: errors.New("sie klagt")}, nil
	}, nil)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(ctx, "Wer"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	anwendung.sichtbarkeitBeenden()

	if anwendung.horchstand().Sichtbar {
		t.Fatal("der Stand sagt, es sei noch sichtbar")
	}
}

// **Ein Ankündiger, der nicht ankündigt, wird gemeldet.**
func TestEinAnkuendigerDerNichtAnkuendigtWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	anwendung.MitEntdeckung(func(
		entdeckung.Angebot,
	) (Ankuendigung, error) {
		return nil, errors.New("das Netz nimmt niemanden auf")
	}, nil)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SichtbarWerden(ctx, "Wer"); antwort.Fehler == nil {
		t.Fatal("die Ankündigung kam durch")
	}
}

// **Und ohne Identität wird nichts angekündigt.**
func TestOhneIdentitaetWirdNichtsAngekuendigt(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	ctx := anwendung.imProfil(riegelkontext(t))

	antwort := anwendung.SichtbarWerden(ctx, "Wer")
	if antwort.Fehler == nil ||
		antwort.Fehler.Schluessel != "ohnePeeridentitaet" {
		t.Fatalf("es kam %+v", antwort.Fehler)
	}
}
