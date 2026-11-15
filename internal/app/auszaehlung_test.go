package app_test

// **Auszählen nach einer erklärten Dimension, an der Bindungsfassade** —
// Werkstück C von Phase 4.11
// (#1338).

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zaehlquelle ist ein Doppel für `app.Auszaehlquelle`.
type zaehlquelle struct {
	dimensionen []app.AuszaehldimensionDTO
	auszaehlung *app.AuszaehlungDTO
	fehler      error

	gefragt  int64
	gewaehlt string
}

func (z *zaehlquelle) Angebot(_ context.Context, setID int64) (
	[]app.AuszaehldimensionDTO, error,
) {
	z.gefragt = setID

	return z.dimensionen, z.fehler
}

func (z *zaehlquelle) Auszaehlen(_ context.Context, setID int64, name string) (
	*app.AuszaehlungDTO, error,
) {
	z.gefragt = setID
	z.gewaehlt = name

	return z.auszaehlung, z.fehler
}

// **Die Dimensionen kommen durch.**
func TestDieAuszaehldimensionenKommenAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	quelle := &zaehlquelle{dimensionen: []app.AuszaehldimensionDTO{
		{Name: "farbe", Beschreibung: "je Farbe", Numerisch: false},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitAuszaehlung(quelle)

	antwort := anwendung.Auszaehldimensionen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if quelle.gefragt != 7 || len(antwort.Dimensionen) != 1 {
		t.Errorf("gefragt wurde nach %d, zurück kamen %+v",
			quelle.gefragt, antwort.Dimensionen)
	}
}

// **Eine Auszählung kommt flach an — mit beiden Listen.**
func TestEineAuszaehlungKommtAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	quelle := &zaehlquelle{auszaehlung: &app.AuszaehlungDTO{
		Dimension: "farbe",
		Gruppen:   []app.AuszaehlgruppeDTO{{Name: "rot", Karten: 2}},
		Befunde:   []string{},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitAuszaehlung(quelle)

	antwort := anwendung.Auszaehlen(t.Context(), "7", "farbe")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if quelle.gewaehlt != "farbe" || len(antwort.Gruppen) != 1 {
		t.Errorf("gewählt war %q, zurück kamen %+v", quelle.gewaehlt, antwort.Gruppen)
	}
}

// **Ohne Naht sagen beide Bindungen es** — und geben keine leere Liste aus.
func TestOhneNahtSagenDieAuszaehlbindungenEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	antwort := anwendung.Auszaehldimensionen(t.Context(), "7")
	if antwort.Fehler == nil || antwort.Dimensionen == nil {
		t.Errorf("ohne Naht kam %+v", antwort)
	}

	gezaehlt := anwendung.Auszaehlen(t.Context(), "7", "farbe")
	if gezaehlt.Fehler == nil || gezaehlt.Gruppen == nil || gezaehlt.Befunde == nil {
		t.Errorf("ohne Naht kam %+v", gezaehlt)
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen.**
func TestDieAuszaehlbindungenPruefenIhreKennung(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAuszaehlung(&zaehlquelle{})

	if antwort := anwendung.Auszaehldimensionen(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("die Zeichenkette ging als Kennung durch")
	}

	if gezaehlt := anwendung.Auszaehlen(t.Context(), "keine Zahl", "farbe"); gezaehlt.Fehler == nil {
		t.Error("dieselbe beim Auszählen")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerAuszaehlquelleKommtDurch(t *testing.T) {
	t.Parallel()

	eigener := fmt.Errorf("%w: kein Spielpaket", domain.ErrNotFound)

	anwendung := app.NeueApp("")
	anwendung.MitAuszaehlung(&zaehlquelle{fehler: eigener})

	antwort := anwendung.Auszaehldimensionen(t.Context(), "7")
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "Spielpaket") {
		t.Errorf("der Fehler lautet %+v", antwort.Fehler)
	}

	if gezaehlt := anwendung.Auszaehlen(t.Context(), "7", "farbe"); gezaehlt.Fehler == nil {
		t.Error("der Fehler kam beim Auszählen nicht durch")
	}
}

// **Und ein leeres Angebot wird zur leeren Liste.**
func TestEinLeeresDimensionsangebotWirdEineLeereListe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitAuszaehlung(&zaehlquelle{})

	if antwort := anwendung.Auszaehldimensionen(t.Context(), "7"); antwort.Dimensionen == nil {
		t.Error("die Liste ist nil")
	}
}
