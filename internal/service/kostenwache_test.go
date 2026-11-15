package service_test

// **Der Kostenwächter** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Was hier geprüft wird und was nicht
//
// **Hier**: dass der Weg eines Auftrags durch Rolle, Nachweis und Journal
// geht — und dass kein Schritt übersprungen wird.
//
// **Nicht hier**: dass zwei Prozesse zusammen unter der Grenze bleiben. Das
// ist eine Aussage über **Transaktionen**, und sie steht in
// `internal/storage` gegen eine echte Datenbank.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var kostenzeit = time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

// kostendoppel spielt das Journal.
type kostendoppel struct {
	stand    domain.Haushaltsstand
	posten   domain.Kostenposten
	liste    []domain.Kostenposten
	fehler   error
	gesehen  domain.Kostenposten
	zeitraum domain.Zeitraum
	grenze   domain.Betrag

	// abschlussprofil ist das Profil, in dem zuletzt abgeschlossen wurde.
	abschlussprofil int64
}

func (k *kostendoppel) BudgetSetzen(
	_ context.Context, _ int64, _ string, zeitraum domain.Zeitraum,
	betrag domain.Betrag, _ time.Time,
) error {
	k.zeitraum, k.grenze = zeitraum, betrag

	return k.fehler
}

func (k *kostendoppel) Reservieren(
	_ context.Context, posten domain.Kostenposten, zeitraum domain.Zeitraum, _ time.Time,
) (domain.Kostenposten, error) {
	k.gesehen, k.zeitraum = posten, zeitraum

	return posten, k.fehler
}

func (k *kostendoppel) Abschliessen(
	_ context.Context, profilID int64, _ string, betrag domain.Betrag, _ time.Time,
) (domain.Kostenposten, error) {
	k.abschlussprofil = profilID
	k.posten.Betrag = betrag
	k.posten.Stand = domain.KostenAbgerechnet

	return k.posten, k.fehler
}

func (k *kostendoppel) Freigeben(_ context.Context, profilID int64, _ string, _ time.Time) (
	domain.Kostenposten, error,
) {
	k.abschlussprofil = profilID
	k.posten.Stand = domain.KostenFreigegeben

	return k.posten, k.fehler
}

func (k *kostendoppel) AlsUnklarMerken(
	_ context.Context, profilID int64, _, grund string, _ time.Time,
) (domain.Kostenposten, error) {
	k.abschlussprofil = profilID
	k.posten.Stand, k.posten.Grund = domain.KostenUnklar, grund

	return k.posten, k.fehler
}

func (k *kostendoppel) Stand(
	_ context.Context, _ int64, _ string, zeitraum domain.Zeitraum, _ time.Time,
) (domain.Haushaltsstand, error) {
	k.zeitraum = zeitraum

	return k.stand, k.fehler
}

func (k *kostendoppel) Posten(
	_ context.Context, _ int64, _ string, _ domain.Zeitraum, _ time.Time, _ int,
) ([]domain.Kostenposten, error) {
	return k.liste, k.fehler
}

// kostenwache baut den Dienst über einem Doppel.
func kostenwache(doppel *kostendoppel, wacht *wachdoppel) *service.Kostenwache {
	return service.NeueKostenwache(doppel, wacht, func() time.Time { return kostenzeit })
}

// **Ein Auftrag geht durch Rolle und Nachweis, bevor er gebucht wird.**
func TestEinAuftragGehtDurchRolleUndNachweis(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	wacht := wache()

	posten, fehler := kostenwache(doppel, wacht).Reservieren(t.Context(),
		service.Kostenauftrag{
			Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
			Zeitraum:      domain.ZeitraumTag,
		}, "nachweis-1")
	if fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if posten.Kennung != "eins" {
		t.Errorf("der Posten ist %+v", posten)
	}

	// **Das Profil kommt vom Akteur** und nicht aus dem Aufruf — wer es
	// mitschicken dürfte, buchte auf fremde Rechnung.
	if doppel.gesehen.ProfilID != 3 {
		t.Errorf("gebucht wurde auf Profil %d", doppel.gesehen.ProfilID)
	}

	// **Die Wirkung ist verbindlich**: hier wird Geld ausgegeben.
	if wacht.wirkung != domain.WirkungVerbindlich {
		t.Errorf("geprüft wurde %q", wacht.wirkung)
	}

	// **Der Nachweis trägt die Kosten** — die Grenze des Menschen wird gegen
	// den Höchstbetrag gehalten (O1).
	if wacht.kennung != "nachweis-1" || wacht.bindung.Kosten != 200 ||
		wacht.bindung.Waehrung != "EUR" || wacht.bindung.Ziel != "anbieter" {
		t.Errorf("die Bindung ist %+v", wacht.bindung)
	}
}

// **Ohne Nachweis wird nichts gebucht** — und die Absage kommt vor dem
// Journal.
func TestOhneNachweisWirdNichtsGebucht(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	wacht := wache()

	_, fehler := kostenwache(doppel, wacht).Reservieren(t.Context(),
		service.Kostenauftrag{
			Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
		}, "")
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("ohne Nachweis: %v", fehler)
	}

	if doppel.gesehen.Kennung != "" {
		t.Error("gebucht wurde trotzdem")
	}
}

// **Eine Absage der Wache hält an** — vor dem Journal.
func TestEineAbsageDerWacheHaeltDenAuftragAn(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	wacht := wache()
	wacht.absage = domain.ErrErlaubnis

	if _, fehler := kostenwache(doppel, wacht).Reservieren(t.Context(),
		service.Kostenauftrag{
			Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
		}, "nachweis-1"); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("mit Absage: %v", fehler)
	}

	if doppel.gesehen.Kennung != "" {
		t.Error("gebucht wurde trotzdem")
	}
}

// **Ohne Zeitraum gilt der Tag** — die engere Grenze, nicht die weitere.
func TestOhneZeitraumGiltDerTag(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}

	if _, fehler := kostenwache(doppel, wache()).Reservieren(t.Context(),
		service.Kostenauftrag{
			Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
		}, "nachweis-1"); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if doppel.zeitraum != domain.ZeitraumTag {
		t.Errorf("gerechnet wurde über %q", doppel.zeitraum)
	}
}

// **Ein erfundener Zeitraum wird abgewiesen** — statt still zum Tag zu werden.
func TestEinErfundenerZeitraumWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}

	if _, fehler := kostenwache(doppel, wache()).Reservieren(t.Context(),
		service.Kostenauftrag{
			Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
			Zeitraum:      "stunde",
		}, "nachweis-1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein erfundener Zeitraum: %v", fehler)
	}
}

// **Die drei Abschlüsse gehen durch, ohne einen zweiten Nachweis zu
// verlangen.**
//
// **Der Nachweis gilt dem Auftrag, nicht der Buchhaltung**: wer ihn beim
// Abschließen noch einmal verlangte, ließe einen Menschen ein zweites Mal
// zustimmen — zu etwas, das schon geschehen ist.
func TestDieAbschluesseVerlangenKeinenZweitenNachweis(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{posten: domain.Kostenposten{Kennung: "eins"}}
	weg := kostenwache(doppel, wache())

	abgerechnet, fehler := weg.Abschliessen(t.Context(), "eins",
		domain.Betrag{Wert: 120, Waehrung: "EUR"})
	if fehler != nil || abgerechnet.Stand != domain.KostenAbgerechnet {
		t.Errorf("abschließen: %+v (%v)", abgerechnet, fehler)
	}

	frei, fehler := weg.Freigeben(t.Context(), "eins")
	if fehler != nil || frei.Stand != domain.KostenFreigegeben {
		t.Errorf("freigeben: %+v (%v)", frei, fehler)
	}

	unklar, fehler := weg.AlsUnklarMerken(t.Context(), "eins", "die Frist lief ab")
	if fehler != nil || unklar.Stand != domain.KostenUnklar || unklar.Grund == "" {
		t.Errorf("unklar: %+v (%v)", unklar, fehler)
	}
}

// **Der Stand nennt vier Zahlen und den Zeitraum, über den sie gelten.**
func TestDerStandNenntVierZahlen(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{stand: domain.Haushaltsstand{
		Konto: "anbieter", Grenze: domain.Betrag{Wert: 500, Waehrung: "EUR"},
	}}

	stand, fehler := kostenwache(doppel, wache()).Stand(t.Context(),
		"anbieter", domain.ZeitraumMonat)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Konto != "anbieter" || doppel.zeitraum != domain.ZeitraumMonat {
		t.Errorf("der Stand ist %+v über %q", stand, doppel.zeitraum)
	}
}

// **Das Budget setzt, wer verwalten darf** — es ist die Freigabe selbst.
func TestDasBudgetSetztWerVerwaltenDarf(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	wacht := wache()

	if fehler := kostenwache(doppel, wacht).BudgetSetzen(t.Context(), "anbieter",
		domain.ZeitraumMonat, domain.Betrag{Wert: 5000, Waehrung: "EUR"}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if doppel.grenze.Wert != 5000 || doppel.zeitraum != domain.ZeitraumMonat {
		t.Errorf("gesetzt wurde %+v über %q", doppel.grenze, doppel.zeitraum)
	}

	// **Eine Grenze zu setzen ist ein Schreiben** — und braucht keinen
	// Nachweis: sie gibt nichts aus, sie begrenzt.
	if wacht.wirkung != domain.WirkungSchreiben || wacht.kennung != "" {
		t.Errorf("geprüft wurde %q mit %q", wacht.wirkung, wacht.kennung)
	}
}

// **Jeder Fehler des Journals kommt durch** — eine stille Null wäre eine
// Freigabe, die niemand gegeben hat.
func TestJederFehlerDesJournalsKommtDurch(t *testing.T) {
	t.Parallel()

	// **Ein eigenes Doppel je Unterlauf, und der Grund ist gemessen.**
	//
	// Hier standen **ein** Doppel und **eine** Wache für sieben parallele
	// Unterläufe. Das Doppel merkt sich, womit es gerufen wurde — also
	// schreibt jeder Unterlauf in dieselben Felder, und der Renn-Detektor hat
	// es am 2026-09-11 angestrichen (drei Rennen in einem Lauf).
	//
	// **Es war immer falsch und fiel nur nicht auf**: keiner der Unterläufe
	// liest, was ein anderer schreibt — der Fehler ist derselbe, gleich wer
	// gefragt hat. Ein Rennen ohne falsches Ergebnis ist trotzdem eines, und
	// unter Last wird daraus eines mit.
	for name, wege := range map[string]func(*service.Kostenwache) error{
		"reservieren": func(weg *service.Kostenwache) error {
			_, fehler := weg.Reservieren(t.Context(), service.Kostenauftrag{
				Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
				Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
			}, "nachweis-1")

			return fehler
		},
		"abschließen": func(weg *service.Kostenwache) error {
			_, fehler := weg.Abschliessen(t.Context(), "eins",
				domain.Betrag{Wert: 120, Waehrung: "EUR"})

			return fehler
		},
		"freigeben": func(weg *service.Kostenwache) error {
			_, fehler := weg.Freigeben(t.Context(), "eins")

			return fehler
		},
		"unklar merken": func(weg *service.Kostenwache) error {
			_, fehler := weg.AlsUnklarMerken(t.Context(), "eins", "grund")

			return fehler
		},
		"der Stand": func(weg *service.Kostenwache) error {
			_, fehler := weg.Stand(t.Context(), "anbieter", domain.ZeitraumTag)

			return fehler
		},
		"die Posten": func(weg *service.Kostenwache) error {
			_, fehler := weg.Posten(t.Context(), "anbieter", domain.ZeitraumTag)

			return fehler
		},
		"das Budget": func(weg *service.Kostenwache) error {
			return weg.BudgetSetzen(t.Context(), "anbieter", domain.ZeitraumTag,
				domain.Betrag{Wert: 500, Waehrung: "EUR"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := &kostendoppel{fehler: errors.New("das Journal ist zu")}

			if fehler := wege(kostenwache(doppel, wache())); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ein Fehler der Wache kommt ebenso durch** — an jedem der sieben Wege.
func TestEinFehlerDerWacheHaeltJedenKostenwegAn(t *testing.T) {
	t.Parallel()

	wacht := wache()
	wacht.fehler = errors.New("kein Akteur")
	weg := kostenwache(&kostendoppel{}, wacht)

	if _, fehler := weg.Stand(t.Context(), "anbieter", domain.ZeitraumTag); fehler == nil {
		t.Error("der Stand kam ohne Akteur durch")
	}

	if _, fehler := weg.Posten(t.Context(), "anbieter", domain.ZeitraumTag); fehler == nil {
		t.Error("die Posten kamen ohne Akteur durch")
	}

	if fehler := weg.BudgetSetzen(t.Context(), "anbieter", domain.ZeitraumTag,
		domain.Betrag{Wert: 500, Waehrung: "EUR"}); fehler == nil {
		t.Error("das Budget kam ohne Akteur durch")
	}

	if _, fehler := weg.Reservieren(t.Context(), service.Kostenauftrag{
		Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
	}, "nachweis-1"); fehler == nil {
		t.Error("reserviert wurde ohne Akteur")
	}
}
