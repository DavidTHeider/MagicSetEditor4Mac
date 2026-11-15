package storage_test

// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Was diese Läufe halten
//
// Die Abnahme verlangt es wörtlich: „Tests verwenden dieselbe echte
// Storage-Naht." Hier gibt es kein Doppel — jeder Fall geht durch eine echte
// Datenbank, weil die Zusage eine über **Transaktionen** ist.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probezeit ist ein fester Augenblick — mitten im Monat, damit Tag und Monat
// verschiedene Fenster haben.
var probezeit = time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

func euro(wert int64) domain.Betrag { return domain.Betrag{Wert: wert, Waehrung: "EUR"} }

// auftrag baut einen gültigen Posten.
func auftrag(kennung string, hoechst int64) domain.Kostenposten {
	return domain.Kostenposten{
		Kennung: kennung, ProfilID: 1, Konto: "anbieter",
		Aufgabe: "karte.text", Hoechstbetrag: euro(hoechst),
	}
}

// journalMitBudget gibt ein Journal, dessen Konto eine Grenze hat.
func journalMitBudget(t *testing.T, grenze int64) *storage.Kostenjournal {
	t.Helper()

	journal := storage.NeuesKostenjournal(frischeDatenbank(t))

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(grenze), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	return journal
}

// **Ohne freigegebenes Budget geht kein kostenpflichtiger Auftrag hinaus.**
//
// Das ist die sichere Vorgabe, und sie steht in der Abnahme: „Widerruf oder
// Local-only sperrt neue externe Aufträge vor Dial."
func TestOhneBudgetGehtKeinAuftragHinaus(t *testing.T) {
	journal := storage.NeuesKostenjournal(frischeDatenbank(t))

	_, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 100), domain.ZeitraumTag, probezeit)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("ohne Budget: %v", fehler)
	}
}

// **Zwei Aufrufer können zusammen nicht mehr reservieren als die Grenze.**
//
// **Das ist die Zusage, für die es das Journal gibt** — und sie geht durch
// dieselbe Datenbank, nicht durch zwei Zähler.
func TestZweiAuftraegeSprengenDieGrenzeNicht(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("der erste: %v", fehler)
	}

	_, fehler := journal.Reservieren(t.Context(),
		auftrag("zwei", 300), domain.ZeitraumTag, probezeit)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der zweite: %v", fehler)
	}

	// **Und einer, der passt, geht** — die Grenze sperrt nicht alles.
	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("drei", 200), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Errorf("der dritte: %v", fehler)
	}
}

// **Dieselbe Identität bucht kein zweites Mal.**
//
// Das ist der Schutz gegen den zweiten kostenpflichtigen Auftrag: ein Abbruch
// nach dem Ruf, ein Neustart, ein Wiederholungsversuch — alle drei zeigen
// dieselbe Kennung vor.
func TestDieselbeIdentitaetBuchtNichtZweimal(t *testing.T) {
	journal := journalMitBudget(t, 500)

	erster, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der erste: %v", fehler)
	}

	// **Der zweite Ruf gibt den ersten zurück** — obwohl 300 + 300 die
	// Grenze sprengte.
	zweiter, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der zweite: %v", fehler)
	}

	if zweiter.Kennung != erster.Kennung || zweiter.Stand != domain.KostenReserviert {
		t.Errorf("der zweite ist %+v", zweiter)
	}

	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Reserviert.Wert != 300 {
		t.Errorf("reserviert sind %d — zweimal gebucht", stand.Reserviert.Wert)
	}
}

// **Abgerechnet wird höchstens, was reserviert war.**
func TestAbgerechnetWirdHoechstensDasReservierte(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "eins", euro(400), probezeit); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("über dem Höchstbetrag: %v", fehler)
	}

	posten, fehler := journal.Abschliessen(t.Context(), 1, "eins", euro(120), probezeit)
	if fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	if posten.Stand != domain.KostenAbgerechnet || posten.Betrag.Wert != 120 {
		t.Errorf("der Posten ist %+v", posten)
	}

	// **Der Rest ist wieder frei** — 300 reserviert, 120 verbraucht.
	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Verbraucht.Wert != 120 || stand.Verfuegbar.Wert != 380 {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Ein zweiter Abschluss ist ein Konflikt** — wer zweimal abrechnet, bucht
// zweimal.
func TestEinZweiterAbschlussIstEinKonflikt(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "eins", euro(120), probezeit); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	if _, fehler := journal.Abschliessen(
		t.Context(), 1, "eins", euro(120), probezeit); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der zweite Abschluss: %v", fehler)
	}

	// **Und einen Auftrag, den es nicht gibt, gibt es nicht.**
	if _, fehler := journal.Freigeben(
		t.Context(), 1, "gibtsnicht", probezeit); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein erfundener Auftrag: %v", fehler)
	}
}

// **Unklar bindet weiter** — und lässt sich danach klären.
//
// „Timeout nach möglicher Anbieterannahme hält Reservierung bis belastbarer
// Klärung", sagt die Abnahme.
func TestEinUnklarerAuftragBindetWeiter(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	posten, fehler := journal.AlsUnklarMerken(t.Context(), 1, "eins",
		"die Frist lief nach dem Absenden ab", probezeit)
	if fehler != nil {
		t.Fatalf("als unklar merken: %v", fehler)
	}

	if posten.Stand != domain.KostenUnklar || posten.Grund == "" {
		t.Errorf("der Posten ist %+v", posten)
	}

	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	// **Der ganze Höchstbetrag bleibt gebunden** — nicht null, und nicht
	// „vielleicht".
	if stand.Unklar.Wert != 300 || stand.Verfuegbar.Wert != 200 {
		t.Errorf("der Stand ist %+v", stand)
	}

	// **Ein unklarer Posten ist offen und lässt sich klären** — das ist der
	// belegwahrende Weg heraus.
	geklaert, fehler := journal.Abschliessen(t.Context(), 1, "eins", euro(90), probezeit)
	if fehler != nil {
		t.Fatalf("klären: %v", fehler)
	}

	if geklaert.Stand != domain.KostenAbgerechnet {
		t.Errorf("nach dem Klären: %+v", geklaert)
	}
}

// **Freigegeben bindet nichts** — der Ruf scheiterte vor jeder Annahme.
func TestEinFreigegebenerAuftragBindetNichts(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if _, fehler := journal.Freigeben(t.Context(), 1, "eins", probezeit); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Verfuegbar.Wert != 500 {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Wer die Grenze unter das Gebundene senkt, hat null verfügbar** — und
// keine negative Zahl.
func TestEineGesenkteGrenzeGibtNullVerfuegbar(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 400), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(100), probezeit); fehler != nil {
		t.Fatalf("die Grenze senken: %v", fehler)
	}

	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumTag, probezeit)
	if fehler != nil {
		t.Fatalf("der Stand: %v", fehler)
	}

	if stand.Verfuegbar.Wert != 0 || stand.Verfuegbar.Waehrung != "EUR" {
		t.Errorf("verfügbar sind %+v", stand.Verfuegbar)
	}

	// **Und ein neuer Auftrag geht nicht mehr durch.**
	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("zwei", 50), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("gebucht wurde über die gesenkte Grenze hinaus")
	}
}

// **Ein neuer Tag ist ein neues Fenster** — und der Posten von gestern bindet
// dort nichts.
func TestEinNeuerTagIstEinNeuesFenster(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("gestern", 400), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("gestern: %v", fehler)
	}

	morgen := probezeit.Add(24 * time.Hour)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("heute", 400), domain.ZeitraumTag, morgen); fehler != nil {
		t.Fatalf("heute: %v", fehler)
	}

	// **Im Monatsfenster stehen beide** — dieselben Posten, andere Frage.
	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumMonat, euro(1000), probezeit); fehler != nil {
		t.Fatalf("das Monatsbudget: %v", fehler)
	}

	stand, fehler := journal.Stand(t.Context(), 1, "anbieter", domain.ZeitraumMonat, morgen)
	if fehler != nil {
		t.Fatalf("der Monatsstand: %v", fehler)
	}

	if stand.Reserviert.Wert != 800 {
		t.Errorf("im Monat sind %d gebunden", stand.Reserviert.Wert)
	}
}

// **Die Posten kommen mit ihrer Geschichte** — jüngste zuerst.
func TestDiePostenKommenMitIhrerGeschichte(t *testing.T) {
	journal := journalMitBudget(t, 900)

	for nummer, kennung := range []string{"eins", "zwei", "drei"} {
		wann := probezeit.Add(time.Duration(nummer) * time.Minute)

		if _, fehler := journal.Reservieren(t.Context(),
			auftrag(kennung, 100), domain.ZeitraumTag, wann); fehler != nil {
			t.Fatalf("%s: %v", kennung, fehler)
		}
	}

	posten, fehler := journal.Posten(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit, 2)
	if fehler != nil {
		t.Fatalf("die Posten: %v", fehler)
	}

	if len(posten) != 2 || posten[0].Kennung != "drei" {
		t.Errorf("die Posten sind %+v", posten)
	}

	if posten[0].Aufgabe != "karte.text" || posten[0].Begonnen.IsZero() {
		t.Errorf("der jüngste ist %+v", posten[0])
	}
}

// **Ein Budget lässt sich lesen und ändern** — und ohne eines ist es leer.
func TestEinBudgetLaesstSichLesenUndAendern(t *testing.T) {
	journal := storage.NeuesKostenjournal(frischeDatenbank(t))

	leer, lesefehler := journal.Budget(t.Context(), 1, "anbieter", domain.ZeitraumTag)
	if lesefehler != nil || !leer.Leer() {
		t.Fatalf("ohne Budget: %+v (%v)", leer, lesefehler)
	}

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(700), probezeit); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	gesetzt, zweitfehler := journal.Budget(t.Context(), 1, "anbieter", domain.ZeitraumTag)
	if zweitfehler != nil || gesetzt.Wert != 700 {
		t.Errorf("das Budget ist %+v (%v)", gesetzt, zweitfehler)
	}

	// **Ein krummer Betrag kommt nicht durch.**
	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, domain.Betrag{Wert: 5, Waehrung: "x"}, probezeit); fehler == nil {
		t.Error("eine krumme Währung kam durch")
	}
}

// **Ein krummer Auftrag kommt nicht ins Journal.**
func TestEinKrummerAuftragKommtNichtInsJournal(t *testing.T) {
	journal := journalMitBudget(t, 500)

	ohne := auftrag("eins", 100)
	ohne.Aufgabe = ""

	if _, fehler := journal.Reservieren(
		t.Context(), ohne, domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("ein Auftrag ohne Aufgabe kam durch")
	}
}

// **Eine andere Währung als die Grenze wird abgewiesen** — und nicht
// umgerechnet.
func TestEineAndereWaehrungWirdAbgewiesen(t *testing.T) {
	journal := journalMitBudget(t, 500)

	inDollar := auftrag("eins", 100)
	inDollar.Hoechstbetrag = domain.Betrag{Wert: 100, Waehrung: "USD"}

	if _, fehler := journal.Reservieren(t.Context(), inDollar,
		domain.ZeitraumTag, probezeit); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("in einer anderen Währung: %v", fehler)
	}
}
