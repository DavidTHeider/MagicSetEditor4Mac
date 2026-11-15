// **Was die Massenänderung nicht schreiben darf** — Befund 1 des Zweitblicks
// über Phase 3.5
// (#1211).
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// massenablage merkt sich, was geschrieben wurde.
type massenablage struct{ geschrieben int }

func (m *massenablage) FelderSetzen(context.Context, []int64, string, string) (
	string, error,
) {
	m.geschrieben++

	return "vorgang7", nil
}

func (m *massenablage) FelderSetzenMitErwartung(
	context.Context, []domain.Feldvorwert, string, string,
) (string, error) {
	m.geschrieben++

	return "vorgang8", nil
}

// feldformen gibt eine feste Vorlage.
type feldformen struct {
	felder []domain.Field
	fehler error
}

func (f feldformen) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return f.felder, f.fehler
}

// massenaufsicht baut die Aufsicht über einer Vorlage.
func massenaufsicht(t *testing.T, felder []domain.Field, fehler error) (
	*service.Kartenaufsicht, *massenablage,
) {
	t.Helper()

	ablage := &massenablage{}

	aufsicht := service.NeueKartenaufsicht(
		&kartensicht{karten: map[int64]int64{1: 7}}, nil).
		MitFeldweg(nil, ablage).
		MitFeldformen(feldformen{felder: felder, fehler: fehler})

	return aufsicht, ablage
}

// **Ein nicht editierbares Feld wird abgewiesen.**
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356 vom
// 2021-07-06. MSE lässt ein solches Feld nicht einmal auswählen; 24 der 249
// Kartenfelder von `magic.mse-game` tragen es.
func TestEinNichtEditierbaresFeldWirdInDerMasseAbgewiesen(t *testing.T) {
	t.Parallel()

	aufsicht, ablage := massenaufsicht(t, []domain.Field{
		{Name: "shape", Editierbar: false, Gespeichert: true},
	}, nil)

	_, err := aufsicht.FelderSetzen(t.Context(), 7, []int64{1}, "shape", "rounded")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", err)
	}

	if ablage.geschrieben != 0 {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// **Ein berechnetes Feld ebenso.**
//
// `save value: false` heißt: der Wert entsteht aus anderen. Wer hineinschreibt,
// schreibt einen, den die Vorlage nicht liest — 46 der 249 Kartenfelder tragen
// es.
//
// **Und anders als beim Einzelweg wird nicht aufgelöst**, sondern abgewiesen.
// Der Einzelweg kann es, weil dort **eine** Karte einen zusammengesetzten Text
// bekommt und ihn in seine Teile zerlegt. Über viele Karten hinweg gibt es
// keinen solchen Text: derselbe Wert müsste für jede Karte anders zerfallen.
func TestEinBerechnetesFeldWirdInDerMasseAbgewiesen(t *testing.T) {
	t.Parallel()

	aufsicht, ablage := massenaufsicht(t, []domain.Field{
		{Name: "text", Editierbar: true, Gespeichert: false},
	}, nil)

	_, err := aufsicht.FelderSetzen(t.Context(), 7, []int64{1}, "text", "irgendwas")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", err)
	}

	if ablage.geschrieben != 0 {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// **Ein gewöhnliches Feld geht durch.**
//
// Ohne diese Zeile wäre auch ein Weg grün, der gar nichts mehr schreibt.
func TestEinGewoehnlichesFeldGehtDurchDieMasse(t *testing.T) {
	t.Parallel()

	aufsicht, ablage := massenaufsicht(t, []domain.Field{
		{Name: "name", Editierbar: true, Gespeichert: true},
	}, nil)

	if _, err := aufsicht.FelderSetzen(
		t.Context(), 7, []int64{1}, "name", "Feuerball"); err != nil {
		t.Fatalf("FelderSetzen: %v", err)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Und eine unlesbare Vorlage hält nichts auf.**
//
// # Dieselbe Duldung wie beim Einzelweg, und derselbe Grund
//
// Ohne die Vorlage weiß niemand, ob ein Feld berechnet ist. Wer deshalb gar
// nicht schriebe, machte jedes Set unbearbeitbar, dessen Spielpaket nicht
// installiert ist — ein Set aus MTG JSON ist genau das. Der Grund steht
// ausführlich bei [service.Feldweg.setzen]; hier steht, dass die Massenänderung
// ihn teilt.
func TestEineUnlesbareVorlageHaeltDieMasseNichtAuf(t *testing.T) {
	t.Parallel()

	aufsicht, ablage := massenaufsicht(t, nil, errKeineVorlage)

	if _, err := aufsicht.FelderSetzen(
		t.Context(), 7, []int64{1}, "name", "Feuerball"); err != nil {
		t.Fatalf("FelderSetzen: %v", err)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Ein Feld, das die Vorlage nicht kennt, geht ebenfalls durch.**
//
// Ein Set aus MTG JSON trägt Felder, für die die gewählte Vorlage keine
// Erklärung hat. Sie zu verweigern hieße, sie unbearbeitbar zu machen.
func TestEinUnbekanntesFeldGehtDurchDieMasse(t *testing.T) {
	t.Parallel()

	aufsicht, ablage := massenaufsicht(t, []domain.Field{
		{Name: "name", Editierbar: true, Gespeichert: true},
	}, nil)

	if _, err := aufsicht.FelderSetzen(
		t.Context(), 7, []int64{1}, "mtgjson id", "abc"); err != nil {
		t.Fatalf("FelderSetzen: %v", err)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// errKeineVorlage ist der Lesefehler dieser Läufe.
var errKeineVorlage = errors.New("das Spielpaket ist nicht installiert")
