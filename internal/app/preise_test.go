// **Die Preise am Fenster** — Werkstück 3 von Phase 3.4
// (#1165).
package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// preisnahtdoppel beantwortet die vier Wege.
type preisnahtdoppel struct {
	staende    []domain.Preisstand
	alter      time.Duration
	befunde    []domain.Befund
	uebersicht service.Portfolio
	fehler     error
}

func (p preisnahtdoppel) Holen(context.Context, string) (
	[]domain.Preisstand, time.Duration, []domain.Befund, error,
) {
	return p.staende, p.alter, p.befunde, p.fehler
}

func (p preisnahtdoppel) Neueste(context.Context, string) ([]domain.Preisstand, error) {
	return p.staende, p.fehler
}

func (p preisnahtdoppel) Verlauf(
	context.Context, string, domain.Preisstand,
) ([]domain.Preisstand, error) {
	return p.staende, p.fehler
}

func (p preisnahtdoppel) Portfolio(context.Context) (service.Portfolio, error) {
	return p.uebersicht, p.fehler
}

// festeUhr gibt eine Uhr, die immer denselben Augenblick zeigt.
func festeUhr() func() time.Time {
	fest := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	return func() time.Time { return fest }
}

// probestand baut einen Preisstand.
func probestand(wert int64, stunden int) domain.Preisstand {
	return domain.Preisstand{
		Printing: "abc",
		Quelle:   domain.PreisquelleScryfall,

		// **Markt, Art und Lage gehören seit A3 dazu** (#1442).
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Lage:     domain.PreisGenannt,
		Waehrung: "EUR",
		Finish:   domain.FinishNonfoil,
		Wert:     wert,
		Erhoben:  festeUhr()().Add(-time.Duration(stunden) * time.Hour),
	}
}

// **Ein Preis kommt mit Quelle, Zeitpunkt und Alter an.**
//
// # Warum alle drei
//
// Die Quelle sagt über ihre Preise ausdrücklich, dass sie für nichts einsteht.
// Eine Zahl ohne diese drei Angaben wäre eine Behauptung, die niemand deckt.
func TestEinPreisKommtMitSeinenAngabenAn(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{probestand(417, 0)},
		alter:   50 * time.Hour,
	})

	antwort := anwendung.PreiseHolen(t.Context(), "abc")
	if antwort.Fehler != nil {
		t.Fatalf("PreiseHolen: %+v", antwort.Fehler)
	}

	stand := antwort.Staende[0]

	if stand.Wert != "4.17" {
		t.Errorf("der Wert ist %q", stand.Wert)
	}

	if stand.Quelle != "scryfall" {
		t.Errorf("die Quelle ist %q", stand.Quelle)
	}

	if stand.AlterStunden != 50 {
		t.Errorf("das Alter ist %d Stunden", stand.AlterStunden)
	}

	// **Der Zeitpunkt trägt seine Zone.** Eine Ortszeit ohne sie liest sich
	// auf einem anderen Rechner anders.
	if stand.Erhoben != "2026-08-27T12:00:00Z" {
		t.Errorf("der Zeitpunkt ist %q", stand.Erhoben)
	}
}

// **Ein abgelegter Stand trägt sein eigenes Alter.**
//
// Bei `Preise` gibt es keine Antwort aus dem Netz — das Alter kommt aus dem
// Zeitpunkt, an dem er erhoben wurde.
func TestEinAbgelegterStandTraegtSeinAlter(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{probestand(417, 30)},
	})

	antwort := anwendung.Preise(t.Context(), "abc")
	if antwort.Fehler != nil {
		t.Fatalf("Preise: %+v", antwort.Fehler)
	}

	if antwort.Staende[0].AlterStunden != 30 {
		t.Errorf("das Alter ist %d Stunden", antwort.Staende[0].AlterStunden)
	}
}

// **Das Alter einer Liste ist das des ältesten Standes.**
//
// Eine Auskunft ist so frisch wie ihr ältester Bestandteil; wer den jüngsten
// nennte, gäbe sie frischer aus, als sie ist.
func TestDasAlterEinerListeIstDasDesAeltesten(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{staende: []domain.Preisstand{
		probestand(417, 2), probestand(500, 40),
	}})

	antwort := anwendung.Preise(t.Context(), "abc")

	for i, stand := range antwort.Staende {
		if stand.AlterStunden != 40 {
			t.Errorf("Stand %d ist %d Stunden alt", i, stand.AlterStunden)
		}
	}
}

// **Befunde reisen mit — auch wenn der Abruf scheitert.**
//
// Sonst verschwände „drei Preise waren nicht zu deuten" hinter dem einen Satz
// „es ging nicht".
func TestBefundeReisenAuchImFehlerfallMit(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitPreisen(preisnahtdoppel{
		befunde: []domain.Befund{{Meldung: "einer ist nicht zu deuten"}},
		fehler:  errors.New("die Gegenstelle schweigt"),
	})

	antwort := anwendung.PreiseHolen(t.Context(), "abc")

	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam nicht an")
	}

	if len(antwort.Befunde) != 1 {
		t.Errorf("es kamen %d Befunde", len(antwort.Befunde))
	}
}

// **Die Übersicht trägt Summe, Zahl, Alter und Quelle.**
func TestDieUebersichtTraegtIhreAngaben(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitSammlung(&sammlungsnahtdoppel{})
	anwendung.MitPreisen(preisnahtdoppel{uebersicht: service.Portfolio{
		Zeilen: []service.Portfoliozeile{{
			Waehrung:       "EUR",
			Wert:           123456,
			Exemplare:      42,
			AeltesterPreis: festeUhr()().Add(-25 * time.Hour),
			Quelle:         domain.PreisquelleScryfall,
		}},
		OhnePreis: 7,
	}})

	antwort := anwendung.Portfolio(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Portfolio: %+v", antwort.Fehler)
	}

	zeile := antwort.Zeilen[0]

	if zeile.Wert != "1234.56" || zeile.Exemplare != 42 {
		t.Errorf("die Zeile ist %+v", zeile)
	}

	if zeile.AlterStunden != 25 || zeile.Quelle != "scryfall" {
		t.Errorf("die Angaben sind %+v", zeile)
	}

	// **Und die Exemplare ohne Preis stehen eigens da.** Ohne diese Zahl sähe
	// die Summe aus wie eine über alle.
	if antwort.OhnePreis != 7 {
		t.Errorf("ohne Preis sind %d", antwort.OhnePreis)
	}
}

// **Ohne Naht sagen alle vier Wege es — und geben leere Listen** (§3).
func TestOhnePreisnahtSagenAlleWegeEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	for name, antwort := range map[string]app.Preisantwort{
		"PreiseHolen":  anwendung.PreiseHolen(t.Context(), "abc"),
		"Preise":       anwendung.Preise(t.Context(), "abc"),
		"Preisverlauf": anwendung.Preisverlauf(t.Context(), probereihe()),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if antwort.Fehler == nil || antwort.Staende == nil || antwort.Befunde == nil {
				t.Errorf("die Antwort ist %+v", antwort)
			}
		})
	}

	uebersicht := anwendung.Portfolio(t.Context())
	if uebersicht.Fehler == nil || uebersicht.Zeilen == nil {
		t.Errorf("Portfolio: %+v", uebersicht)
	}
}

// **Ein Fehler kommt auf jedem Weg übersetzt an.**
func TestEinFehlerDerPreisnahtKommtAn(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(&sammlungsnahtdoppel{})
	anwendung.MitPreisen(preisnahtdoppel{fehler: errors.New("die Naht stolpert")})

	for name, ruf := range map[string]func() *app.Fehler{
		"PreiseHolen": func() *app.Fehler {
			return anwendung.PreiseHolen(t.Context(), "abc").Fehler
		},
		"Preise": func() *app.Fehler { return anwendung.Preise(t.Context(), "abc").Fehler },
		"Preisverlauf": func() *app.Fehler {
			return anwendung.Preisverlauf(t.Context(), probereihe()).Fehler
		},
		"Portfolio": func() *app.Fehler { return anwendung.Portfolio(t.Context()).Fehler },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if ruf() == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Ohne gesetzte Uhr gilt die echte.**
//
// Ein Lauf, der ein `App{}` ohne Composition Root baut, soll eine Antwort
// bekommen und keinen Absturz.
func TestOhneGesetzteUhrGiltDieEchte(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{{
			Printing: "abc", Quelle: domain.PreisquelleScryfall,
			Markt: domain.MarktPapier, Art: domain.PreisartVerkauf,
			Lage:     domain.PreisGenannt,
			Waehrung: "EUR", Finish: domain.FinishNonfoil,
			Wert: 100, Erhoben: time.Now(),
		}},
	})

	antwort := anwendung.Preise(t.Context(), "abc")
	if antwort.Fehler != nil {
		t.Fatalf("Preise: %+v", antwort.Fehler)
	}

	if antwort.Staende[0].AlterStunden != 0 {
		t.Errorf("das Alter ist %d Stunden", antwort.Staende[0].AlterStunden)
	}
}

// **Ein Stand aus der Zukunft ist null Stunden alt und nicht negativ.**
func TestEinStandAusDerZukunftIstNullStundenAlt(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitSammlung(&sammlungsnahtdoppel{})
	anwendung.MitPreisen(preisnahtdoppel{uebersicht: service.Portfolio{
		Zeilen: []service.Portfoliozeile{{
			Waehrung:       "EUR",
			AeltesterPreis: festeUhr()().Add(3 * time.Hour),
		}},
	}})

	if alter := anwendung.Portfolio(t.Context()).Zeilen[0].AlterStunden; alter != 0 {
		t.Errorf("das Alter ist %d", alter)
	}
}

// **Der Verlauf kommt durch — mit Werten und nicht nur ohne Fehler.**
//
// Der Lauf darüber prüft den Fehlerweg; dieser den, den jemand benutzt.
func TestDerPreisverlaufKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{probestand(100, 5), probestand(200, 1)},
	})

	antwort := anwendung.Preisverlauf(t.Context(), probereihe())
	if antwort.Fehler != nil {
		t.Fatalf("Preisverlauf: %+v", antwort.Fehler)
	}

	if len(antwort.Staende) != 2 || antwort.Staende[0].Wert != "1.00" {
		t.Errorf("es kam %+v", antwort.Staende)
	}
}

// probereihe benennt die Reihe, um die es in diesen Läufen geht (A3, #1442).
//
// **Als Helfer, weil sie sieben Teile hat**: sieben Zeilen je Aufruf sagten
// nichts über den Lauf, in dem sie stehen.
func probereihe() app.PreisreiheDTO {
	return app.PreisreiheDTO{
		Printing: "abc",
		Quelle:   string(domain.PreisquelleScryfall),
		Markt:    string(domain.MarktPapier),
		Art:      string(domain.PreisartVerkauf),
		Waehrung: "EUR",
		Finish:   string(domain.FinishNonfoil),
	}
}
