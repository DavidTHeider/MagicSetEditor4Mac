package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Ausdruck läuft im Kontext der Karte** —
// #1070.
//
// MSE-Script ist vollständig portiert, und niemand konnte einen Ausdruck
// eingeben. §15 nennt MSEs eigene Kommandozeile als **Orakel** dieses
// Projekts; eine eigene Konsole beantwortet dieselben Fragen über dieses
// Programm.
func TestEinAusdruckLaeuftImKontextDerKarte(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7, Values: []domain.Value{
			{Field: "name", Text: "Prüfstein"},
		}}}, nil)

	ergebnis, fehler := konsole.Auswerten(t.Context(), 7, `1 + 1`)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	if ergebnis.Quelltext != "2" || ergebnis.Text != "2" {
		t.Errorf("Ergebnis %+v", ergebnis)
	}

	// **Ohne Planbauer steht ein Befund da**, statt die Auswertung
	// abzuweisen: `1 + 1` braucht keine Vorlage.
	if len(ergebnis.Befunde) != 1 {
		t.Errorf("%d Befunde, erwartet einer: %v", len(ergebnis.Befunde), ergebnis.Befunde)
	}
}

// **Das Ergebnis kommt als Quelltext.**
//
// `"a"` und nicht `a` — so zeigt MSE es (`src/cli/cli_main.cpp`, Zeile 234,
// Fassung f9d9356 vom 2021-07-06), und wer eine Konsole benutzt, will eine
// Zeichenkette von einem Namen unterscheiden.
func TestDasErgebnisKommtAlsQuelltext(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	ergebnis, fehler := konsole.Auswerten(t.Context(), 7, `"a" + "b"`)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	if ergebnis.Quelltext != `"ab"` || ergebnis.Text != "ab" {
		t.Errorf("Quelltext %q, Text %q", ergebnis.Quelltext, ergebnis.Text)
	}
}

// **Ein Zerteilfehler nennt Zeile und Spalte.**
//
// Wie überall sonst in diesem Programm (§8): eine Meldung ohne Stelle ist
// eine Meldung, die niemand nachschlagen kann.
func TestEinZerteilfehlerNenntSeineStelle(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	_, fehler := konsole.Auswerten(t.Context(), 7, `1 +`)
	if fehler == nil {
		t.Fatal("der unvollständige Ausdruck kam ohne Fehler durch")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v — kein Validierungsbefund", fehler)
	}

	if !strings.Contains(fehler.Error(), "Zeile") || !strings.Contains(fehler.Error(), "Spalte") {
		t.Errorf("Fehler %v — ohne Stelle", fehler)
	}
}

// **Eine Karte, die es nicht gibt, kommt als Fehler durch.**
func TestEineFehlendeKarteKommtDurch(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{fehler: domain.ErrNotFound}, nil)

	if _, fehler := konsole.Auswerten(t.Context(), 7, `1`); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v", fehler)
	}
}

// konsolenkarten gibt eine feste Karte.
type konsolenkarten struct {
	karte  *domain.CardRow
	fehler error
}

func (k konsolenkarten) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return k.karte, k.fehler
}

// **Mit Planbauer stehen `card` und `set` zur Verfügung.**
//
// Das ist der eigentliche Zweck: ein Ausdruck sieht dieselbe Umgebung wie der
// Rasterer. Eine zweite Umgebung daneben liefe auseinander, und dann
// beantwortete die Konsole Fragen über ein Programm, das es nicht gibt.
func TestMitPlanbauerStehtDieKarteZurVerfuegung(t *testing.T) {
	t.Parallel()

	karte := &domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "name", Text: "Prüfstein"},
	}}

	bauer := service.NeuerPlanbauer().
		MitStilen(festerStil{vorlage: service.Vorlage{Stil: saatstil(t)}, da: true})

	konsole := service.NeueSkriptkonsole(konsolenkarten{karte: karte}, bauer)

	ergebnis, fehler := konsole.Auswerten(t.Context(), 7, `card.name`)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	if ergebnis.Quelltext != `"Prüfstein"` {
		t.Errorf("Quelltext %q", ergebnis.Quelltext)
	}
}

// **Ohne Stilpaket bleibt die Bibliothek — samt Befund.**
//
// Ein Ausdruck wie `1 + 1` braucht keine Vorlage; ihn abzuweisen, weil ein
// Paket fehlt, hieße eine Konsole zu bauen, die genau dann schweigt, wenn man
// sie am nötigsten braucht (§8).
func TestOhneStilpaketBleibtDieBibliothek(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]service.Stilquelle{
		"kein Stilverzeichnis": nil,
		"kein Paket": festerStil{
			befunde: []domain.Befund{{Meldung: "kein Paket"}}, da: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bauer := service.NeuerPlanbauer()
			if quelle != nil {
				bauer = bauer.MitStilen(quelle)
			}

			konsole := service.NeueSkriptkonsole(
				konsolenkarten{karte: &domain.CardRow{ID: 7}}, bauer)

			ergebnis, fehler := konsole.Auswerten(t.Context(), 7, `2 * 3`)
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if ergebnis.Quelltext != "6" {
				t.Errorf("Quelltext %q", ergebnis.Quelltext)
			}

			if len(ergebnis.Befunde) == 0 {
				t.Error("der fehlende Stil blieb unbenannt")
			}
		})
	}
}

// **Die Stelle zeigt in den rohen Quelltext** — Werkstück U13 von Phase 5.1
// (#1424).
//
// Das Fenster schickt den Ausdruck so, wie er im Feld steht: mit führenden
// Leerzeilen, Tabs, Zeilenumbrüchen. Eine Stelle, die sich auf eine
// bereinigte Form bezöge, zeigte auf ein Zeichen, das niemand sieht.
func TestDieStelleZeigtInDenRohenQuelltext(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	_, fehler := konsole.Auswerten(t.Context(), 7, "\n\n\t\t1 +")
	if fehler == nil {
		t.Fatal("der unvollständige Ausdruck kam ohne Fehler durch")
	}

	var stelle *msescript.Fehler
	if !errors.As(fehler, &stelle) {
		t.Fatalf("Fehler %v — ohne Stelle", fehler)
	}

	// Zeile 3, denn zwei Leerzeilen gehen voraus; und die Spalte zählt die
	// beiden Tabs mit.
	if stelle.Zeile != 3 || stelle.Spalte < 5 {
		t.Errorf("Stelle Zeile %d, Spalte %d — der rohe Text hat den Ausdruck in Zeile 3 hinter zwei Tabs",
			stelle.Zeile, stelle.Spalte)
	}
}

// **Eine Abbildung kommt mit ihren Einträgen** — rekursiv, damit das Fenster
// einen Baum zeigen kann (#1424). Ein Wert, der keine ist, hat keine.
func TestEineAbbildungKommtMitIhrenEintraegen(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	ergebnis, fehler := konsole.Auswerten(t.Context(), 7, `[a: 1, b: [c: "zwei"]]`)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(ergebnis.Eintraege) != 2 || ergebnis.Eintraege[0].Name != "a" || ergebnis.Eintraege[0].Quelltext != "1" {
		t.Fatalf("Einträge: %+v", ergebnis.Eintraege)
	}

	b := ergebnis.Eintraege[1]
	if b.Name != "b" || b.Art != string(msescript.ArtAbbildung) || len(b.Eintraege) != 1 {
		t.Fatalf("b: %+v", b)
	}

	if c := b.Eintraege[0]; c.Name != "c" || c.Quelltext != `"zwei"` || c.Art != string(msescript.ArtTextwert) || len(c.Eintraege) != 0 {
		t.Errorf("c: %+v", c)
	}

	zahl, fehler := konsole.Auswerten(t.Context(), 7, `1 + 1`)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(zahl.Eintraege) != 0 {
		t.Errorf("eine Zahl hat Einträge: %+v", zahl.Eintraege)
	}
}
