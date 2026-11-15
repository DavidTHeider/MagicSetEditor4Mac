package domain_test

// **Was ein Auftrag draußen kosten darf** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Diese Anwendung rechnet keine Währungen um** — und sagt es, statt eine
// Zahl zu erfinden, für die niemand einsteht.
func TestZweiWaehrungenLassenSichNichtVerrechnen(t *testing.T) {
	t.Parallel()

	euro := domain.Betrag{Wert: 500, Waehrung: "EUR"}
	dollar := domain.Betrag{Wert: 500, Waehrung: "USD"}

	if _, fehler := euro.Plus(dollar); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("addiert: %v", fehler)
	}

	if _, fehler := euro.Minus(dollar); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgezogen: %v", fehler)
	}

	if _, fehler := euro.Uebersteigt(dollar); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("verglichen: %v", fehler)
	}
}

// **Ein Betrag ohne Währung ist null und passt zu jedem** — sonst brauchte
// jede Summe einen Anfangswert, der schon weiß, worin gerechnet wird.
func TestEinLeererBetragPasstZuJedemAnderen(t *testing.T) {
	t.Parallel()

	euro := domain.Betrag{Wert: 500, Waehrung: "EUR"}

	summe, fehler := domain.Betrag{}.Plus(euro)
	if fehler != nil || summe != euro {
		t.Errorf("null + 5 € ist %+v (%v)", summe, fehler)
	}

	summe, fehler = euro.Plus(domain.Betrag{})
	if fehler != nil || summe != euro {
		t.Errorf("5 € + null ist %+v (%v)", summe, fehler)
	}

	rest, fehler := euro.Minus(domain.Betrag{})
	if fehler != nil || rest != euro {
		t.Errorf("5 € − null ist %+v (%v)", rest, fehler)
	}
}

// **Abgezogen wird nie unter null** — ein negatives Guthaben wäre eine Zahl,
// die niemand liest.
func TestEineDifferenzGehtNichtUnterNull(t *testing.T) {
	t.Parallel()

	rest, fehler := domain.Betrag{Wert: 100, Waehrung: "EUR"}.
		Minus(domain.Betrag{Wert: 250, Waehrung: "EUR"})
	if fehler != nil {
		t.Fatalf("abziehen: %v", fehler)
	}

	if rest.Wert != 0 || rest.Waehrung != "EUR" {
		t.Errorf("der Rest ist %+v", rest)
	}
}

// **Zwei Beträge derselben Währung rechnen** — und ein leerer Betrag
// **mit** Währung ist auch einer.
func TestZweiBetraegeDerselbenWaehrungRechnen(t *testing.T) {
	t.Parallel()

	summe, fehler := domain.Betrag{Wert: 250, Waehrung: "EUR"}.
		Plus(domain.Betrag{Wert: 750, Waehrung: "EUR"})
	if fehler != nil || summe.Wert != 1000 || summe.Waehrung != "EUR" {
		t.Errorf("die Summe ist %+v (%v)", summe, fehler)
	}

	rest, fehler := summe.Minus(domain.Betrag{Wert: 400, Waehrung: "EUR"})
	if fehler != nil || rest.Wert != 600 {
		t.Errorf("der Rest ist %+v (%v)", rest, fehler)
	}

	// **Ein Betrag von null in derselben Währung ist nicht „ohne Währung"**
	// — er geht den gewöhnlichen Weg und nicht die Abkürzung oben.
	summe, fehler = domain.Betrag{Waehrung: "EUR"}.Plus(domain.Betrag{Wert: 5, Waehrung: "EUR"})
	if fehler != nil || summe.Wert != 5 {
		t.Errorf("null € + 5 € ist %+v (%v)", summe, fehler)
	}

	ueber, fehler := domain.Betrag{Wert: 1001, Waehrung: "EUR"}.
		Uebersteigt(domain.Betrag{Wert: 1000, Waehrung: "EUR"})
	if fehler != nil || !ueber {
		t.Errorf("10,01 € über 10 €: %v (%v)", ueber, fehler)
	}

	// **Ein Betrag ohne alles übersteigt keine Grenze.**
	ueber, fehler = domain.Betrag{}.Uebersteigt(domain.Betrag{Wert: 1, Waehrung: "EUR"})
	if fehler != nil || ueber {
		t.Errorf("nichts über 0,01 €: %v (%v)", ueber, fehler)
	}
}

// **Eine Grenze von null ist keine Grenze** — dieselbe Lesart wie bei der
// Kostengrenze eines Nachweises.
func TestOhneGrenzeUebersteigtNichts(t *testing.T) {
	t.Parallel()

	ueber, fehler := domain.Betrag{Wert: 999_999, Waehrung: "USD"}.
		Uebersteigt(domain.Betrag{})
	if fehler != nil || ueber {
		t.Errorf("ohne Grenze: %v %v", ueber, fehler)
	}
}

// **Ein krummer Betrag wird abgewiesen** — an jeder der drei Stellen.
func TestEinKrummerBetragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, betrag := range map[string]domain.Betrag{
		"negativ":           {Wert: -1, Waehrung: "EUR"},
		"ohne Währung":      {Wert: 100},
		"zu kurze Währung":  {Wert: 100, Waehrung: "EU"},
		"klein geschrieben": {Wert: 100, Waehrung: "eur"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := betrag.Pruefe(); fehler == nil {
				t.Error("kam durch")
			}
		})
	}

	if fehler := (domain.Betrag{Wert: 100, Waehrung: "EUR"}).Pruefe(); fehler != nil {
		t.Errorf("ein gültiger Betrag: %v", fehler)
	}
}

// **Vier Stände, und nur einer bindet nichts.**
func TestNurEinFreigegebenerPostenBindetNichts(t *testing.T) {
	t.Parallel()

	staende := domain.Kostenstaende()
	if len(staende) != 4 {
		t.Fatalf("es sind %d Stände: %v", len(staende), staende)
	}

	for _, stand := range staende {
		bindet := stand.Bindet()
		erwartet := stand != domain.KostenFreigegeben

		if bindet != erwartet {
			t.Errorf("%q bindet %v, erwartet war %v", stand, bindet, erwartet)
		}
	}

	// **Offen sind die beiden, über die nichts feststeht.**
	for stand, erwartet := range map[domain.Kostenstand]bool{
		domain.KostenReserviert:  true,
		domain.KostenUnklar:      true,
		domain.KostenAbgerechnet: false,
		domain.KostenFreigegeben: false,
	} {
		if offen := stand.Offen(); offen != erwartet {
			t.Errorf("%q ist offen=%v, erwartet war %v", stand, offen, erwartet)
		}
	}
}

// **Was ein Posten bindet, hängt an seinem Stand** — und ein unklarer bindet
// den ganzen Höchstbetrag.
func TestEinUnklarerPostenBindetSeinenHoechstbetrag(t *testing.T) {
	t.Parallel()

	hoechst := domain.Betrag{Wert: 500, Waehrung: "EUR"}
	echt := domain.Betrag{Wert: 120, Waehrung: "EUR"}

	for stand, erwartet := range map[domain.Kostenstand]domain.Betrag{
		domain.KostenReserviert:  hoechst,
		domain.KostenUnklar:      hoechst,
		domain.KostenAbgerechnet: echt,
		domain.KostenFreigegeben: {Waehrung: "EUR"},
	} {
		posten := domain.Kostenposten{Hoechstbetrag: hoechst, Betrag: echt, Stand: stand}

		if gebunden := posten.Gebunden(); gebunden != erwartet {
			t.Errorf("%q bindet %+v, erwartet war %+v", stand, gebunden, erwartet)
		}
	}
}

// **Unbekannte Kosten sind nicht dasselbe wie keine** — ein Auftrag ohne
// Höchstbetrag kommt nicht ins Journal.
func TestEinAuftragOhneHoechstbetragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gut := domain.Kostenposten{
		Kennung: "abc", ProfilID: 1, Konto: "openai", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 200, Waehrung: "EUR"},
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("ein gültiger Auftrag: %v", fehler)
	}

	ohne := gut
	ohne.Hoechstbetrag = domain.Betrag{Waehrung: "EUR"}

	if fehler := ohne.Pruefe(); fehler == nil {
		t.Error("ein Auftrag ohne Höchstbetrag kam durch")
	}

	for name, kaputt := range map[string]func(*domain.Kostenposten){
		"ohne Identität": func(p *domain.Kostenposten) { p.Kennung = "  " },
		"ohne Profil":    func(p *domain.Kostenposten) { p.ProfilID = 0 },
		"ohne Konto":     func(p *domain.Kostenposten) { p.Konto = "" },
		"ohne Aufgabe":   func(p *domain.Kostenposten) { p.Aufgabe = "" },
		"krumme Währung": func(p *domain.Kostenposten) {
			p.Hoechstbetrag = domain.Betrag{Wert: 5, Waehrung: "x"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			posten := gut
			kaputt(&posten)

			if fehler := posten.Pruefe(); fehler == nil {
				t.Error("kam durch")
			}
		})
	}
}

// **Ein Zeitraum fängt in UTC an** — zwei Prozesse in verschiedenen Zonen
// meinen sonst zwei Budgets.
func TestEinZeitraumFaengtInUTCAn(t *testing.T) {
	t.Parallel()

	// **Ein Augenblick, der in Berlin schon der nächste Tag ist.**
	berlin := time.FixedZone("MESZ", 2*60*60)
	jetzt := time.Date(2026, 9, 11, 1, 30, 0, 0, berlin)

	tag := domain.ZeitraumTag.Fenster(jetzt)
	if tag != time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC) {
		t.Errorf("der Tag fängt bei %s an", tag)
	}

	monat := domain.ZeitraumMonat.Fenster(jetzt)
	if monat != time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("der Monat fängt bei %s an", monat)
	}

	if !domain.IstZeitraum("tag") || domain.IstZeitraum("stunde") {
		t.Error("die Zeiträume sind nicht die zwei")
	}

	if len(domain.Zeitraeume()) != 2 {
		t.Errorf("es sind %d Zeiträume", len(domain.Zeitraeume()))
	}
}
