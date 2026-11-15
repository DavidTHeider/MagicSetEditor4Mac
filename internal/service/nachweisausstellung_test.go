package service_test

// **Ausstellen und Nachlesen über die Autorisierung** —
// #1539.
//
// # Was hier geprüft wird
//
// **Die zwei Wege, die am Fenster gefehlt haben.** `Nachweise.Ausstellen`
// stand da und wurde nicht gerufen; die Grenze eines Nachweises war gar nicht
// zu lesen, und deshalb buchte der Auftrag einen leeren Betrag.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitNachweisen gibt eine Autorisierung über diesen Bestand samt Ablage.
func mitNachweisen(
	ablage *nachweisdoppel,
) (*service.Autorisierung, *nutzerdoppel) {
	nutzer, freigaben := bestandFuerAkteure("")

	return autorisierung(nutzer, freigaben).
		MitNachweisen(nachweise(ablage)), nutzer
}

// **Die Autorisierung stellt aus — für den Akteur, den sie selbst baut.**
func TestDieAutorisierungStelltFuerIhrenAkteurAus(t *testing.T) {
	t.Parallel()

	ablage := neuesNachweisdoppel()

	nachweis, fehler := mustAutorisierung(t, ablage).Ausstellen(
		context.Background(), antrag())
	if fehler != nil {
		t.Fatalf("Ausstellen: %v", fehler)
	}

	// **Der Eigner von Profil 7 ist Nutzer 1** (`bestandFuerAkteure`) — und
	// genau der steht im Nachweis, nicht wer ihn beantragt hat.
	if nachweis.NutzerID != 1 || nachweis.ProfilID != 7 {
		t.Fatalf("der Nachweis gehört %d/%d", nachweis.NutzerID, nachweis.ProfilID)
	}
}

// **Ohne Nachweisdienst stellt sie keinen aus.**
//
// **Und das ist die Zusage aus 0078**: an einer Fläche ohne Menschen entsteht
// kein Nachweis. Wäre es ein stiller Erfolg, hätte sich eine Maschine selbst
// freigegeben.
func TestOhneNachweisdienstWirdNichtAusgestellt(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")

	_, fehler := autorisierung(nutzer, freigaben).Ausstellen(
		context.Background(), antrag())
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Ausstellen ohne Dienst: %v, erwartet ErrErlaubnis", fehler)
	}

	_, grenzfehler := autorisierung(nutzer, freigaben).Nachweisgrenze(
		context.Background(), "gleich welche")
	if !errors.Is(grenzfehler, domain.ErrErlaubnis) {
		t.Fatalf("Grenze ohne Dienst: %v, erwartet ErrErlaubnis", grenzfehler)
	}
}

// **Und ohne Akteur auch nicht** — beide Wege fragen ihn zuerst.
func TestOhneAkteurWederAusstellenNochLesen(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kein Profil")

	for name, ruf := range map[string]func(*service.Autorisierung) error{
		"Ausstellen": func(a *service.Autorisierung) error {
			_, fehler := a.Ausstellen(context.Background(), antrag())

			return fehler
		},
		"Nachweisgrenze": func(a *service.Autorisierung) error {
			_, fehler := a.Nachweisgrenze(context.Background(), "n-1")

			return fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache, nutzer := mitNachweisen(neuesNachweisdoppel())
			nutzer.arbeitsprofilfehler = kaputt

			if fehler := ruf(wache); !errors.Is(fehler, kaputt) {
				t.Fatalf("%s: %v, erwartet den Fehler der Ablage", name, fehler)
			}
		})
	}
}

// **Die Grenze eines Nachweises ist zu lesen, ohne ihn zu verbrauchen.**
//
// **Darauf beruht die Buchung**: der Auftrag reserviert, was ein Mensch
// freigegeben hat — und löst den Nachweis erst danach ein.
func TestDieGrenzeIstOhneVerbrauchZuLesen(t *testing.T) {
	t.Parallel()

	ablage := neuesNachweisdoppel()
	wache := mustAutorisierung(t, ablage)

	nachweis, fehler := wache.Ausstellen(context.Background(), antrag())
	if fehler != nil {
		t.Fatalf("Ausstellen: %v", fehler)
	}

	grenze, lesefehler := wache.Nachweisgrenze(context.Background(), nachweis.Kennung)
	if lesefehler != nil {
		t.Fatalf("Nachweisgrenze: %v", lesefehler)
	}

	if grenze != (domain.Kostengrenze{Betrag: 1500, Waehrung: "EUR"}) {
		t.Fatalf("die Grenze ist %+v", grenze)
	}

	// **Und danach trägt er noch** — hätte das Lesen ihn verbraucht, ginge
	// kein Auftrag mehr hinaus, für den er gedacht war.
	if einloesefehler := nachweise(ablage).Einloesen(context.Background(),
		nachweis.Kennung, bindungFuer(nachweis)); einloesefehler != nil {
		t.Fatalf("nach dem Lesen einlösen: %v", einloesefehler)
	}
}

// **Einen fremden Nachweis liest niemand.**
//
// **Dieselbe Zuordnung wie beim Einlösen**: wer ihn nicht einlösen kann, soll
// auch nicht erfahren, auf welchen Betrag er lautet.
func TestEinFremderNachweisIstNichtZuLesen(t *testing.T) {
	t.Parallel()

	for name, fremd := range map[string]domain.Nachweis{
		"ein anderer Nutzer": {Kennung: "f-1", NutzerID: 9, ProfilID: 7},
		"ein anderes Profil": {Kennung: "f-1", NutzerID: 1, ProfilID: 9},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := neuesNachweisdoppel()
			ablage.abgelegt["f-1"] = fremd

			_, fehler := mustAutorisierung(t, ablage).Nachweisgrenze(
				context.Background(), "f-1")
			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Fatalf("%s: %v, erwartet ErrErlaubnis", name, fehler)
			}
		})
	}
}

// **Und eine Kennung, die es nicht gibt, gibt keine Grenze.**
func TestEineUnbekannteKennungGibtKeineGrenze(t *testing.T) {
	t.Parallel()

	_, fehler := mustAutorisierung(t, neuesNachweisdoppel()).Nachweisgrenze(
		context.Background(), "gibtesnicht")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("%v, erwartet ErrNotFound", fehler)
	}
}

// mustAutorisierung gibt die Autorisierung und verwirft den Bestand.
func mustAutorisierung(
	t *testing.T, ablage *nachweisdoppel,
) *service.Autorisierung {
	t.Helper()

	wache, _ := mitNachweisen(ablage)

	return wache
}

// **Die Frist einer Freigabe am Fenster liegt im erlaubten Bereich.**
//
// **Hier und nicht erst in `app`**: [service.Nachweise.Ausstellen] weist eine
// Frist über einem Tag ab, und die Zahl in `app.KIFreigabefrist` steht in
// einer anderen Datei. Wer sie dort hochsetzte, bekäme den Fehler erst, wenn
// ein Mensch auf den Knopf drückt.
func TestDieFristEinerFreigabeLiegtImErlaubtenBereich(t *testing.T) {
	t.Parallel()

	ablage := neuesNachweisdoppel()
	gewuenscht := antrag()
	gewuenscht.Frist = 24 * time.Hour

	if _, fehler := mustAutorisierung(t, ablage).Ausstellen(
		context.Background(), gewuenscht); fehler != nil {
		t.Fatalf("ein Tag ist die Obergrenze: %v", fehler)
	}
}
