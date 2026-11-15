package storage_test

// **Wo zwei Währungen aufeinandertreffen** — Werkstück O12 von Phase 5.2
// (#1439).
//
// **Eine eigene Datei**, weil `kostenfehler_test.go` über die Zeilengrenze aus
// §2 ging. Der Schnitt folgt der Frage: dort steht, was geschieht, wenn der
// **Bestand** nicht antwortet — hier, was geschieht, wenn die **Zahlen** nicht
// zueinander passen.
//
// **Es sind vier Stellen, an denen summiert wird**, und jede hat ihre eigene
// Absage: die Summe der Posten, die Summe plus dem neuen Auftrag, die Summe
// aus Reserviertem und Verbrauchtem, und die unklare Menge obendrauf. Wer nur
// eine prüft, hat drei Zweige ungeprüft gelassen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei Währungen in einem Konto sperren jede Rechnung** — und werden nicht
// gegeneinander verrechnet.
//
// **Der Zustand kommt aus einer Fassung, die es einmal zuließ** — oder aus
// einem Anbieter, der die Währung wechselte. Beides ist ein Befund und keine
// Summe.
func TestZweiWaehrungenInEinemKontoSperrenJedeRechnung(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	for _, zeile := range []struct{ kennung, waehrung, stand string }{
		{"eins", "EUR", "reserviert"},
		{"zwei", "USD", "abgerechnet"},
	} {
		if _, fehler := db.ExecContext(t.Context(), `
			INSERT INTO kostenposten (
				kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
				waehrung, stand, begonnen
			) VALUES (?, 1, 'anbieter', 'karte.text', 100, 100, ?, ?,
				'2026-09-15T00:00:00.000Z')`,
			zeile.kennung, zeile.waehrung, zeile.stand); fehler != nil {
			t.Fatalf("%s legen: %v", zeile.kennung, fehler)
		}
	}

	// **Die Summe selbst geht nicht auf.**
	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("drei", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("zwei Währungen wurden summiert")
	}

	// **Und der Stand ebenso** — reserviert in Euro, verbraucht in Dollar.
	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der Stand summierte zwei Währungen")
	}
}

// **Und die unklare Menge trägt ihre eigene Rechnung** — sie kommt nach den
// beiden anderen, und wer nur jene prüft, hat den dritten Zweig nie betreten.
func TestEineUnklareMengeInAndererWaehrung(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	// **Reserviert und verbraucht in Euro** — die erste Summe geht auf.
	for _, zeile := range []struct{ kennung, stand, waehrung string }{
		{"eins", "reserviert", "EUR"},
		{"zwei", "abgerechnet", "EUR"},
		{"drei", "unklar", "USD"},
	} {
		if _, fehler := db.ExecContext(t.Context(), `
			INSERT INTO kostenposten (
				kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
				waehrung, stand, begonnen
			) VALUES (?, 1, 'anbieter', 'karte.text', 100, 100, ?, ?,
				'2026-09-15T00:00:00.000Z')`,
			zeile.kennung, zeile.waehrung, zeile.stand); fehler != nil {
			t.Fatalf("%s legen: %v", zeile.kennung, fehler)
		}
	}

	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("die unklare Menge wurde umgerechnet")
	}
}

// **Eine Summe in einer Währung, ein Auftrag in einer anderen** — das ist der
// Weg **nach** der Summe, und er hat seinen eigenen Zweig.
func TestEineSummeInAndererWaehrungAlsDerAuftrag(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, domain.Betrag{Wert: 500, Waehrung: "USD"},
		probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), `
		INSERT INTO kostenposten (
			kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
			waehrung, stand, begonnen
		) VALUES ('alt', 1, 'anbieter', 'karte.text', 100, 0, 'EUR',
			'reserviert', '2026-09-15T00:00:00.000Z')`); fehler != nil {
		t.Fatalf("den alten Posten legen: %v", fehler)
	}

	// **Die Summe steht in Euro, der neue Auftrag rechnet in Dollar.**
	inDollar := auftrag("neu", 100)
	inDollar.Hoechstbetrag = domain.Betrag{Wert: 100, Waehrung: "USD"}

	if _, fehler := journal.Reservieren(t.Context(), inDollar,
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("die Summe wurde umgerechnet")
	}
}

// **Ein krummer Abschlussbetrag kommt nicht durch** — und ein Abschluss in
// einer anderen Währung als der Reservierung ebenso wenig.
func TestEinKrummerAbschlussbetragKommtNichtDurch(t *testing.T) {
	journal := journalMitBudget(t, 500)

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 300), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "eins",
		domain.Betrag{Wert: 5, Waehrung: "x"}, probezeit); fehler == nil {
		t.Error("eine krumme Währung kam durch")
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "eins",
		domain.Betrag{Wert: 5, Waehrung: "USD"}, probezeit); fehler == nil {
		t.Error("ein Abschluss in einer anderen Währung kam durch")
	}
}

// **Ein unlesbarer Zeitpunkt wird zur Nullzeit und nicht zum Fehler.**
//
// Er ist die Auskunft „wann", nicht die Buchung; wer daran das Lesen eines
// Postens scheitern ließe, machte aus einer krummen Zeile eine Sperre für das
// ganze Journal.
func TestEinUnlesbarerZeitpunktWirdZurNullzeit(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if _, fehler := db.ExecContext(t.Context(), `
		INSERT INTO kostenposten (
			kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
			waehrung, stand, begonnen, beendet
		) VALUES ('krumm', 1, 'anbieter', 'karte.text', 100, 0, 'EUR',
			'reserviert', 'irgendwann', 'auch irgendwann')`); fehler != nil {
		t.Fatalf("die krumme Zeile legen: %v", fehler)
	}

	posten, fehler := journal.Posten(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit, 5)
	if fehler != nil {
		t.Fatalf("die Posten: %v", fehler)
	}

	if len(posten) != 1 || !posten[0].Begonnen.IsZero() || !posten[0].Beendet.IsZero() {
		t.Errorf("der Posten ist %+v", posten)
	}
}
