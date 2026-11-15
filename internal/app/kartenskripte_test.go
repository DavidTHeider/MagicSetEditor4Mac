package app_test

// **Karten aus einem Skript, an der Bindungsfassade** — Werkstück B von Phase
// 4.11 (#1337).

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// skriptquelle ist ein Doppel für `app.Kartenskriptquelle`.
type skriptquelle struct {
	skripte []app.KartenskriptDTO
	lauf    *app.KartenskriptlaufDTO
	fehler  error

	gefragt  int64
	gefuehrt string
}

func (s *skriptquelle) Angebot(_ context.Context, setID int64) (
	[]app.KartenskriptDTO, error,
) {
	s.gefragt = setID

	return s.skripte, s.fehler
}

func (s *skriptquelle) Ausfuehren(_ context.Context, setID int64, name string) (
	*app.KartenskriptlaufDTO, error,
) {
	s.gefragt = setID
	s.gefuehrt = name

	return s.lauf, s.fehler
}

// **Die Skripte kommen durch.**
func TestDieKartenskripteKommenAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	quelle := &skriptquelle{skripte: []app.KartenskriptDTO{
		{Name: "Zwei Farben", Kuerzel: "Z", Beschreibung: "zwei Karten"},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitKartenskripten(quelle)

	antwort := anwendung.Kartenskripte(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if quelle.gefragt != 7 {
		t.Errorf("gefragt wurde nach Set %d", quelle.gefragt)
	}

	if len(antwort.Skripte) != 1 || antwort.Skripte[0].Name != "Zwei Farben" {
		t.Errorf("die Skripte lauten %+v", antwort.Skripte)
	}
}

// **Ein Lauf gibt seine Kartenkennungen.**
func TestEinKartenskriptlaufKommtAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	quelle := &skriptquelle{lauf: &app.KartenskriptlaufDTO{
		Karten: []string{"2", "3"}, Befunde: []string{},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitKartenskripten(quelle)

	antwort := anwendung.KartenskriptAusfuehren(t.Context(), "7", "Zwei Farben")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if quelle.gefuehrt != "Zwei Farben" || len(antwort.Karten) != 2 {
		t.Errorf("gerufen wurde %q, angelegt %+v", quelle.gefuehrt, antwort.Karten)
	}
}

// **Ohne Naht sagen beide Bindungen es** — und geben keine leere Liste aus.
//
// „Dieses Spiel hat keine Skripte" und „das Fenster ist nicht verdrahtet" sind
// verschiedene Auskünfte, und die zweite gehört gesagt.
func TestOhneNahtSagenDieKartenskriptbindungenEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	antwort := anwendung.Kartenskripte(t.Context(), "7")
	if antwort.Fehler == nil {
		t.Error("ohne Naht kam ein Angebot zurück")
	}

	if antwort.Skripte == nil {
		t.Error("die Liste ist nil")
	}

	if lauf := anwendung.KartenskriptAusfuehren(t.Context(), "7", "x"); lauf.Fehler == nil {
		t.Error("ohne Naht lief ein Skript")
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen.**
func TestDieKartenskriptbindungenPruefenIhreKennung(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenskripten(&skriptquelle{})

	if antwort := anwendung.Kartenskripte(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("die Zeichenkette ging als Kennung durch")
	}

	if lauf := anwendung.KartenskriptAusfuehren(
		probekontext(t), "keine Zahl", "x"); lauf.Fehler == nil {
		t.Error("dieselbe beim Ausführen")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerKartenskriptquelleKommtDurch(t *testing.T) {
	t.Parallel()

	eigener := fmt.Errorf("%w: kein Spielpaket", domain.ErrNotFound)

	anwendung := app.NeueApp("")
	anwendung.MitKartenskripten(&skriptquelle{fehler: eigener})

	antwort := anwendung.Kartenskripte(t.Context(), "7")
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "Spielpaket") {
		t.Errorf("der Fehler lautet %+v", antwort.Fehler)
	}

	if lauf := anwendung.KartenskriptAusfuehren(t.Context(), "7", "x"); lauf.Fehler == nil {
		t.Error("der Fehler kam beim Ausführen nicht durch")
	}
}

// **Und eine leere Antwort der Naht wird zur leeren Liste.**
func TestEinLeeresAngebotWirdEineLeereListe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenskripten(&skriptquelle{})

	if antwort := anwendung.Kartenskripte(t.Context(), "7"); antwort.Skripte == nil {
		t.Error("die Liste ist nil")
	}
}
