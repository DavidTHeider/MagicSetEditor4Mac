package app_test

// **Die Wurzel des Kostenjournals geht wirklich durch den Bestand** —
// Werkstück O12 von Phase 5.2
// (#1439).
//
// **Über eine echte Datenbank**, denn die Frage ist, ob die Naht verdrahtet
// ist (§10) — und ob der Weg **durch** Wache, Dienst und Journal trägt.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieKostenwurzelGehtDurchDenBestand(t *testing.T) {
	db := frischerBestand(t)
	anwendung := app.KostenAnwendungUeber(db)
	ctx := probekontext(t)

	// **Ohne Budget ist alles null** — und das ist eine Auskunft.
	leer := anwendung.Kostenstand(ctx, "anbieter", "tag")
	if leer.Fehler != nil {
		t.Fatalf("der leere Stand: %+v", leer.Fehler)
	}

	if leer.Grenze.Wert != 0 || leer.Verfuegbar.Wert != 0 {
		t.Errorf("der leere Stand ist %+v", leer)
	}

	if antwort := anwendung.KostenbudgetSetzen(
		ctx, "anbieter", "tag", 50_000, "EUR"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	gesetzt := anwendung.Kostenstand(ctx, "anbieter", "tag")
	if gesetzt.Fehler != nil || gesetzt.Verfuegbar.Wert != 50_000 {
		t.Errorf("nach dem Setzen: %+v (%+v)", gesetzt, gesetzt.Fehler)
	}

	// **Und die Liste ist leer, nicht null.**
	posten := anwendung.Kostenposten(ctx, "anbieter", "tag")
	if posten.Fehler != nil || len(posten.Posten) != 0 {
		t.Errorf("die Posten sind %+v (%+v)", posten.Posten, posten.Fehler)
	}
}

// **Ein Auftrag geht nur mit Freigabe hinaus** — und steht danach im Journal.
//
// **Das ist der ganze Weg**: Rolle, Nachweis, Budget, Buchung — durch eine
// echte Datenbank, mit einem Nachweis, den die Ablage kennt.
func TestEinAuftragGehtNurMitFreigabeHinaus(t *testing.T) {
	db := frischerBestand(t)
	anwendung := app.KostenAnwendungUeber(db)
	ctx := probekontext(t)

	if antwort := anwendung.KostenbudgetSetzen(
		ctx, "anbieter", "tag", 50_000, "EUR"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	jetzt := time.Now()

	if err := storage.NeueNachweisablage(db).NachweisAnlegen(ctx, domain.Nachweis{
		Kennung: "freigabe", NutzerID: 1, ProfilID: 1,
		Zweck: service.KostenjournalZweck, Ziel: "anbieter",
		Grenze:      domain.Kostengrenze{Betrag: 30_000, Waehrung: "EUR"},
		Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(time.Hour),
	}); err != nil {
		t.Fatalf("den Nachweis legen: %v", err)
	}

	wache := app.KostenwacheUeber(func() *sql.DB { return db },
		app.AutorisierungAus(db), time.Now)

	// **Ohne Nachweis geht nichts** — vor jeder Buchung.
	if _, err := wache.Reservieren(ctx, service.Kostenauftrag{
		Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 20_000, Waehrung: "EUR"},
	}, ""); err == nil {
		t.Error("ohne Freigabe wurde gebucht")
	}

	posten, err := wache.Reservieren(ctx, service.Kostenauftrag{
		Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 20_000, Waehrung: "EUR"},
	}, "freigabe")
	if err != nil {
		t.Fatalf("reservieren: %v", err)
	}

	if posten.Stand != domain.KostenReserviert {
		t.Errorf("der Posten ist %+v", posten)
	}

	// **Und der Stand zeigt es an allen vier Zahlen.**
	stand := anwendung.Kostenstand(ctx, "anbieter", "tag")
	if stand.Fehler != nil || stand.Reserviert.Wert != 20_000 ||
		stand.Verfuegbar.Wert != 30_000 {
		t.Errorf("der Stand ist %+v (%+v)", stand, stand.Fehler)
	}

	// **Der Nachweis ist verbraucht** — ein zweiter Auftrag mit derselben
	// Freigabe geht nicht.
	if _, zweiter := wache.Reservieren(ctx, service.Kostenauftrag{
		Kennung: "zwei", Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 1_000, Waehrung: "EUR"},
	}, "freigabe"); zweiter == nil {
		t.Error("dieselbe Freigabe trug zwei Aufträge")
	}

	// **Abgeschlossen wird ohne zweite Freigabe.**
	abgerechnet, abschlussfehler := wache.Abschliessen(ctx, "eins",
		domain.Betrag{Wert: 12_000, Waehrung: "EUR"})
	if abschlussfehler != nil || abgerechnet.Stand != domain.KostenAbgerechnet {
		t.Errorf("abschließen: %+v (%v)", abgerechnet, abschlussfehler)
	}

	nachher := anwendung.Kostenposten(ctx, "anbieter", "tag")
	if nachher.Fehler != nil || len(nachher.Posten) != 1 ||
		nachher.Posten[0].Betrag.Wert != 12_000 {
		t.Errorf("die Posten sind %+v (%+v)", nachher.Posten, nachher.Fehler)
	}

	pruefeDieBeidenAnderenEnden(t, db, wache, ctx)
}

// **Die beiden anderen Enden gehen denselben Weg** — freigeben und unklar
// merken.
//
// **Sie stehen hier und nicht in einem eigenen Lauf**, weil sie denselben
// Aufbau brauchen: ein Budget, ein Nachweis, ein reservierter Posten.
func pruefeDieBeidenAnderenEnden(
	t *testing.T, db *sql.DB, wache *service.Kostenwache, ctx context.Context,
) {
	t.Helper()

	jetzt := time.Now()

	for nummer, kennung := range []string{"frei", "unklar"} {
		if err := storage.NeueNachweisablage(db).NachweisAnlegen(ctx, domain.Nachweis{
			Kennung: "freigabe-" + kennung, NutzerID: 1, ProfilID: 1,
			Zweck: service.KostenjournalZweck, Ziel: "anbieter",
			Ausgestellt: jetzt.Add(-time.Minute), Ablauf: jetzt.Add(time.Hour),
		}); err != nil {
			t.Fatalf("den Nachweis %d legen: %v", nummer, err)
		}

		if _, err := wache.Reservieren(ctx, service.Kostenauftrag{
			Kennung: kennung, Konto: "anbieter", Aufgabe: "karte.text",
			Hoechstbetrag: domain.Betrag{Wert: 1_000, Waehrung: "EUR"},
		}, "freigabe-"+kennung); err != nil {
			t.Fatalf("%s reservieren: %v", kennung, err)
		}
	}

	frei, err := wache.Freigeben(ctx, "frei")
	if err != nil || frei.Stand != domain.KostenFreigegeben {
		t.Errorf("freigeben: %+v (%v)", frei, err)
	}

	offen, err := wache.AlsUnklarMerken(ctx, "unklar", "die Frist lief ab")
	if err != nil || offen.Stand != domain.KostenUnklar {
		t.Errorf("unklar merken: %+v (%v)", offen, err)
	}
}
