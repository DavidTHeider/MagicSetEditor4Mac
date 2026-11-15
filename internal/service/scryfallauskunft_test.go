package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// auskunftsprobe ist ein Doppel für [service.Fremdauskuenfte].
type auskunftsprobe struct {
	karte        *scryfall.Karte
	kartenalter  time.Duration
	hinweise     []scryfall.Regelhinweis
	hinweisalter time.Duration
	satz         *scryfall.Setangabe
	symbole      []scryfall.Symbol
	befunde      []domain.Befund

	kartenfehler  error
	hinweisfehler error
	satzfehler    error
	symbolfehler  error

	gefragteKennung string
}

func (a *auskunftsprobe) KarteNachName(context.Context, string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	return a.karte, a.kartenalter, a.befunde, a.kartenfehler
}

func (a *auskunftsprobe) Rulings(_ context.Context, kartenID string) (
	[]scryfall.Regelhinweis, time.Duration, []domain.Befund, error,
) {
	a.gefragteKennung = kartenID

	return a.hinweise, a.hinweisalter, nil, a.hinweisfehler
}

func (a *auskunftsprobe) Setangaben(context.Context, string) (
	*scryfall.Setangabe, time.Duration, []domain.Befund, error,
) {
	return a.satz, 0, a.befunde, a.satzfehler
}

func (a *auskunftsprobe) Symbologie(context.Context) (
	[]scryfall.Symbol, time.Duration, []domain.Befund, error,
) {
	return a.symbole, 0, a.befunde, a.symbolfehler
}

func TestAuskunftRegelhinweise(t *testing.T) {
	t.Parallel()

	t.Run("holt die Karte und dann ihre Hinweise", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{
			karte: &scryfall.Karte{
				Kennung: "abc-123", Name: "Lightning Bolt",
				Adresse: "https://scryfall.invalid/lea/161",
			},
			hinweise: []scryfall.Regelhinweis{
				{Quelle: scryfall.QuelleWotc, Datum: "2004-10-04", Text: "Ein Hinweis."},
			},
		}

		auskunft, fehler := service.NeueAuskunft(probe).Regelhinweise(
			t.Context(), "lightning bolt")
		if fehler != nil {
			t.Fatalf("Regelhinweise: %v", fehler)
		}

		// **Die Kennung kommt aus der Karte**, nicht vom Aufrufer — niemand
		// kennt sie.
		if probe.gefragteKennung != "abc-123" {
			t.Errorf("gefragt wurde %q", probe.gefragteKennung)
		}

		// **Und der Name aus der Antwort, nicht der eingegebene.** Wer sich
		// vertippt hat, soll es sehen.
		if auskunft.Kartenname != "Lightning Bolt" {
			t.Errorf("Kartenname = %q", auskunft.Kartenname)
		}

		if len(auskunft.Hinweise) != 1 || auskunft.Adresse == "" {
			t.Errorf("Auskunft = %+v", auskunft)
		}
	})

	t.Run("nennt das ältere der beiden Alter", func(t *testing.T) {
		t.Parallel()

		// **Eine Auskunft ist so frisch wie ihr ältester Bestandteil** (§16,
		// Regel 5). Das jüngere zu nennen behauptete mehr, als gemessen ist.
		probe := &auskunftsprobe{
			karte:        &scryfall.Karte{Kennung: "x", Name: "X"},
			kartenalter:  9 * time.Hour,
			hinweisalter: 2 * time.Hour,
		}

		auskunft, fehler := service.NeueAuskunft(probe).Regelhinweise(t.Context(), "X")
		if fehler != nil {
			t.Fatalf("Regelhinweise: %v", fehler)
		}

		if auskunft.Alter != 9*time.Hour {
			t.Errorf("Alter = %v, erwartet 9h", auskunft.Alter)
		}
	})

	t.Run("nennt das Alter der Hinweise, wenn es das ältere ist", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{
			karte:        &scryfall.Karte{Kennung: "x", Name: "X"},
			kartenalter:  1 * time.Hour,
			hinweisalter: 5 * time.Hour,
		}

		auskunft, fehler := service.NeueAuskunft(probe).Regelhinweise(t.Context(), "X")
		if fehler != nil {
			t.Fatalf("Regelhinweise: %v", fehler)
		}

		if auskunft.Alter != 5*time.Hour {
			t.Errorf("Alter = %v, erwartet 5h", auskunft.Alter)
		}
	})

	t.Run("reicht einen Fehler der Kartensuche durch", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{kartenfehler: domain.ErrNotFound}

		if _, fehler := service.NeueAuskunft(probe).Regelhinweise(
			t.Context(), "Gibtesnicht"); !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}

		if probe.gefragteKennung != "" {
			t.Error("es wurde trotzdem nach Hinweisen gefragt")
		}
	})

	t.Run("reicht einen Fehler der Hinweise durch", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{
			karte:         &scryfall.Karte{Kennung: "x", Name: "X"},
			hinweisfehler: domain.ErrNotFound,
		}

		if _, fehler := service.NeueAuskunft(probe).Regelhinweise(
			t.Context(), "X"); !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})
}

func TestAuskunftSatz(t *testing.T) {
	t.Parallel()

	t.Run("gibt die Angaben", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{satz: &scryfall.Setangabe{Code: "lea", Name: "Limited Alpha"}}

		auskunft, fehler := service.NeueAuskunft(probe).Satz(t.Context(), "lea")
		if fehler != nil {
			t.Fatalf("Satz: %v", fehler)
		}

		if auskunft.Satz.Code != "lea" {
			t.Errorf("Satz = %+v", auskunft.Satz)
		}
	})

	t.Run("weist ein leeres Kürzel ab", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{}

		if _, fehler := service.NeueAuskunft(probe).Satz(t.Context(), ""); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("reicht einen Fehler durch", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{satzfehler: domain.ErrNotFound}

		if _, fehler := service.NeueAuskunft(probe).Satz(t.Context(), "xyz"); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})
}

// **Die Manakostenprüfung ist der erste Verbraucher der Symbologie.**
//
// Sie beantwortet eine Frage, die vorher niemand beantworten konnte: `{2}{r}`
// sieht aus wie ein Kosten und ist keiner — MSE zeichnet dafür ein leeres
// Kästchen, ohne zu sagen, warum.
func TestAuskunftManakostenpruefung(t *testing.T) {
	t.Parallel()

	eins := 1.0
	zwanzig := 20.0

	symbole := []scryfall.Symbol{
		{Zeichen: "{W}", IstMana: true, Manawert: &eins, InManakosten: true},
		{Zeichen: "{R}", IstMana: true, Manawert: &eins, InManakosten: true},
		{Zeichen: "{2}", IstMana: true, InManakosten: true},

		// **`{20}` ist ein Manasymbol und steht doch in keinem Kosten.** Wer
		// nur auf „ist ein Symbol" prüfte, ließe es durch.
		{Zeichen: "{20}", IstMana: true, Manawert: &zwanzig, InManakosten: false},
	}

	faelle := map[string]struct {
		kosten   string
		erwartet []string
	}{
		"richtig":                   {"{2}{R}", nil},
		"kleingeschrieben":          {"{2}{r}", []string{"{r}"}},
		"erfunden":                  {"{Zauberstab}", []string{"{Zauberstab}"}},
		"nur im Regeltext":          {"{20}", []string{"{20}"}},
		"leer":                      {"", nil},
		"mehrere unbekannte":        {"{q}{2}{z}", []string{"{q}", "{z}"}},
		"ohne Klammern geschrieben": {"2R", nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &auskunftsprobe{symbole: symbole}

			unbekannt, fehler := service.NeueAuskunft(probe).Manakostenpruefung(
				t.Context(), fall.kosten)
			if fehler != nil {
				t.Fatalf("Manakostenpruefung: %v", fehler)
			}

			if !slices.Equal(unbekannt, fall.erwartet) {
				t.Errorf("unbekannt = %v, erwartet %v", unbekannt, fall.erwartet)
			}
		})
	}

	t.Run("reicht einen Fehler der Symbologie durch", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{symbolfehler: domain.ErrNotFound}

		if _, fehler := service.NeueAuskunft(probe).Manakostenpruefung(
			t.Context(), "{W}"); !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("gibt die Symbologie auch roh heraus", func(t *testing.T) {
		t.Parallel()

		probe := &auskunftsprobe{symbole: symbole}

		auskunft, fehler := service.NeueAuskunft(probe).Symbole(t.Context())
		if fehler != nil {
			t.Fatalf("Symbole: %v", fehler)
		}

		if len(auskunft.Symbole) != 4 {
			t.Errorf("%d Symbole", len(auskunft.Symbole))
		}
	})
}
