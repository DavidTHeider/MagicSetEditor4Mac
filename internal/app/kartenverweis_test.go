package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verweis auf einen echten Druck an der Fassade** — Werkstück 2 von
// Phase 3.6
// (#1167).

// verweisquelle ist ein Doppel des Verweisdienstes.
type verweisquelle struct {
	verweise map[int64]domain.Kartenverweis
	befunde  []domain.Befund
	fehler   error
	gesetzt  string
}

func neueVerweisquelle() *verweisquelle {
	return &verweisquelle{verweise: map[int64]domain.Kartenverweis{}}
}

func (v *verweisquelle) Setzen(_ context.Context, karteID int64, printing string) (
	*domain.Kartenverweis, []domain.Befund, error,
) {
	v.gesetzt = printing

	if v.fehler != nil {
		return nil, v.befunde, v.fehler
	}

	verweis := domain.Kartenverweis{
		Printing: printing, Name: "Lightning Bolt", Setcode: "lea", Nummer: "161",
		Adresse: "https://scryfall.com/card/lea/161",
		Gesetzt: time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC),
	}
	v.verweise[karteID] = verweis

	return &verweis, v.befunde, nil
}

func (v *verweisquelle) Loesen(_ context.Context, karteID int64) error {
	if v.fehler != nil {
		return v.fehler
	}

	delete(v.verweise, karteID)

	return nil
}

func (v *verweisquelle) Fuer(_ context.Context, karteID int64) (
	*domain.Kartenverweis, error,
) {
	if v.fehler != nil {
		return nil, v.fehler
	}

	verweis, da := v.verweise[karteID]
	if !da {
		return nil, nil //nolint:nilnil // „keiner" ist kein Fehler
	}

	return &verweis, nil
}

func (v *verweisquelle) JeSet(_ context.Context, _ int64) (
	map[int64]domain.Kartenverweis, error,
) {
	return v.verweise, v.fehler
}

func fassadeMitVerweisen(quelle app.Verweisquelle) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitVerweisen(quelle)

	return anwendung
}

// **Setzen gibt den aufgelösten Verweis zurück** — samt Zeitpunkt.
func TestKarteVerweisSetzenGibtDenAufgeloestenVerweis(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()

	antwort := fassadeMitVerweisen(quelle).KarteVerweisSetzen(
		probekontext(t), "7", "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Verweis == nil {
		t.Fatal("es kam kein Verweis zurück")
	}

	if antwort.Verweis.KarteID != "7" || antwort.Verweis.Name != "Lightning Bolt" {
		t.Errorf("Verweis: %+v", *antwort.Verweis)
	}

	if antwort.Verweis.Gesetzt != "2026-08-27T12:00:00Z" {
		t.Errorf("Gesetzt: %q", antwort.Verweis.Gesetzt)
	}

	if quelle.gesetzt != "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1" {
		t.Errorf("weitergereicht wurde %q", quelle.gesetzt)
	}
}

// **Eine Karte ohne Verweis ist kein Fehler.** Der Normalfall darf nicht wie
// ein Ausfall aussehen.
func TestKarteVerweisOhneVerweisIstKeinFehler(t *testing.T) {
	t.Parallel()

	antwort := fassadeMitVerweisen(neueVerweisquelle()).KarteVerweis(probekontext(t), "7")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Verweis != nil {
		t.Errorf("es kam %+v zurück", *antwort.Verweis)
	}
}

// **Eine unbrauchbare Kennung wird an der Fassade abgewiesen** — und geht gar
// nicht erst hinein.
func TestEineUnbrauchbareKennungErreichtDenVerweisdienstNicht(t *testing.T) {
	t.Parallel()

	for name, ruf := range map[string]func(*app.App) *app.Fehler{
		"lesen": func(a *app.App) *app.Fehler {
			return a.KarteVerweis(probekontext(t), "keine Zahl").Fehler
		},
		"setzen": func(a *app.App) *app.Fehler {
			return a.KarteVerweisSetzen(probekontext(t), "keine Zahl", "abc").Fehler
		},
		"lösen": func(a *app.App) *app.Fehler {
			return a.KarteVerweisLoesen(probekontext(t), "keine Zahl").Fehler
		},
		"je Set": func(a *app.App) *app.Fehler {
			return a.SetVerweise(probekontext(t), "keine Zahl").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := neueVerweisquelle()

			if fehler := ruf(fassadeMitVerweisen(quelle)); fehler == nil {
				t.Fatal("die unbrauchbare Kennung kam als Erfolg zurück")
			}

			if quelle.gesetzt != "" {
				t.Errorf("es ging %q hinein", quelle.gesetzt)
			}
		})
	}
}

// **Ohne eingerichteten Zugang sagt jede der vier Bindungen es.**
//
// # Warum alle vier und nicht eine
//
// Weil ein `nil` an dieser Stelle sonst eine Panik im Fenster wäre — und weil
// drei gleichlautende Zweige drei sind, von denen nur einer je einen Test
// sähe (§3).
func TestOhneZugangSagenDieVerweisbindungenEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	for name, fehler := range map[string]*app.Fehler{
		"lesen":  anwendung.KarteVerweis(t.Context(), "7").Fehler,
		"setzen": anwendung.KarteVerweisSetzen(t.Context(), "7", "abc").Fehler,
		"lösen":  anwendung.KarteVerweisLoesen(t.Context(), "7").Fehler,
		"je Set": anwendung.SetVerweise(t.Context(), "1").Fehler,
	} {
		if fehler == nil {
			t.Errorf("%s: der fehlende Zugang kam als Erfolg zurück", name)
		}
	}
}

// **Ein Fehler des Dienstes wird übersetzt** — und die Befunde reisen mit.
func TestEinFehlerBeimSetzenWirdUebersetzt(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	quelle.fehler = domain.ErrNotFound
	quelle.befunde = []domain.Befund{{Meldung: "layout unbekannt"}}

	antwort := fassadeMitVerweisen(quelle).KarteVerweisSetzen(probekontext(t), "7", "abc")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q", antwort.Fehler.Code)
	}

	if len(antwort.Befunde) != 1 {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}
}

// **Auch Lesen, Lösen und die Setliste übersetzen ihren Fehler.**
func TestDieUebrigenVerweisbindungenUebersetzenIhrenFehler(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	quelle.fehler = errors.New("der Bestand ist belegt")

	anwendung := fassadeMitVerweisen(quelle)

	for name, fehler := range map[string]*app.Fehler{
		"lesen":  anwendung.KarteVerweis(t.Context(), "7").Fehler,
		"lösen":  anwendung.KarteVerweisLoesen(t.Context(), "7").Fehler,
		"je Set": anwendung.SetVerweise(t.Context(), "1").Fehler,
	} {
		if fehler == nil {
			t.Errorf("%s: der Fehler kam als Erfolg zurück", name)
		}
	}
}

// **Lösen entfernt ihn.**
func TestKarteVerweisLoesenEntferntIhn(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	anwendung := fassadeMitVerweisen(quelle)

	if antwort := anwendung.KarteVerweisSetzen(t.Context(), "7", "abc"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.KarteVerweisLoesen(t.Context(), "7"); antwort.Fehler != nil {
		t.Fatalf("lösen: %+v", antwort.Fehler)
	}

	if len(quelle.verweise) != 0 {
		t.Errorf("es blieb %v zurück", quelle.verweise)
	}
}

// **Die Verweise eines Sets kommen als Liste, nach Kartenkennung geordnet.**
//
// Numerisch und nicht als Text: als Text käme 10 vor 2, und die Kartenliste
// stünde in einer Reihenfolge, die niemand erwartet.
func TestDieSetverweiseKommenNachKennungGeordnet(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	anwendung := fassadeMitVerweisen(quelle)

	for _, karteID := range []string{"10", "2"} {
		if antwort := anwendung.KarteVerweisSetzen(
			probekontext(t), karteID, "abc"); antwort.Fehler != nil {
			t.Fatalf("setzen: %+v", antwort.Fehler)
		}
	}

	antwort := anwendung.SetVerweise(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Verweise) != 2 {
		t.Fatalf("es kamen %d Verweise", len(antwort.Verweise))
	}

	if antwort.Verweise[0].KarteID != "2" || antwort.Verweise[1].KarteID != "10" {
		t.Errorf("Reihenfolge: %q, %q",
			antwort.Verweise[0].KarteID, antwort.Verweise[1].KarteID)
	}
}

// **Die drei Bindungen reichen durch — mit dem Kontext des Fensters.**
//
// # Warum sie einen eigenen Lauf brauchen
//
// `Bindungen` ist die Naht, die Wails wirklich ruft; die Tests darüber gehen
// an `App` vorbei an ihr vorbei. Ohne diesen Lauf wäre die gebundene
// Oberfläche der einzige Teil, den niemand betritt (§3).
func TestDieVerweisbindungenReichenDurch(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	bindungen := app.NeueBindungen(fassadeMitVerweisen(quelle))

	gesetzt := bindungen.KarteVerweisSetzen(t.Context(), "7", "516e137a")
	if gesetzt.Fehler != nil || gesetzt.Verweis == nil {
		t.Fatalf("setzen: %+v", gesetzt)
	}

	gelesen := bindungen.KarteVerweis(t.Context(), "7")
	if gelesen.Fehler != nil || gelesen.Verweis == nil {
		t.Fatalf("lesen: %+v", gelesen)
	}

	if geloest := bindungen.KarteVerweisLoesen(t.Context(), "7"); geloest.Fehler != nil {
		t.Fatalf("lösen: %+v", geloest)
	}

	if len(quelle.verweise) != 0 {
		t.Errorf("es blieb %v zurück", quelle.verweise)
	}
}
