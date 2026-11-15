package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wortablage ist ein In-Memory-Double der Schreibseite (§3).
type wortablage struct {
	woerter  []*domain.Keyword
	naechste int64

	fehler error

	// gesetzt merkt sich den letzten Schreibvorgang.
	gesetzt string
}

func (a *wortablage) Schluesselwoerter(context.Context, int64) ([]*domain.Keyword, error) {
	return a.woerter, a.fehler
}

func (a *wortablage) SchluesselwortAnlegen(_ context.Context, _ int64, modus string) (int64, error) {
	if a.fehler != nil {
		return 0, a.fehler
	}

	a.naechste++
	a.woerter = append(a.woerter, &domain.Keyword{ID: a.naechste, Mode: modus})

	return a.naechste, nil
}

func (a *wortablage) SchluesselwortSetzen(
	_ context.Context, _, _ int64, feld, wert string,
) error {
	a.gesetzt = feld + "=" + wert

	return a.fehler
}

func (a *wortablage) SchluesselwortLoeschen(context.Context, int64, int64) error {
	return a.fehler
}

// spielwoerter ist ein Stilquellendouble mit Schlüsselwörtern des Spiels.
type spielwoerter struct {
	vorlage service.Vorlage
	da      bool
}

func (s spielwoerter) StilFuerSet(context.Context, int64) (service.Vorlage, []domain.Befund, bool) {
	return s.vorlage, nil, s.da
}

func (s spielwoerter) StilFuerKarte(
	context.Context, *domain.CardRow,
) (service.Vorlage, []domain.Befund, bool) {
	return s.vorlage, nil, s.da
}

// spielvorlage gibt eine Vorlage mit zwei festen Schlüsselwörtern.
func spielvorlage() service.Vorlage {
	return service.Vorlage{
		Rohwoerter: []msepaket.Schluesselwort{
			{Wort: "Fliegend", Muster: "Fliegend", Modus: "core"},
			{Wort: "Kaputt", Muster: "", Modus: "core"},
		},
		Modi: []msepaket.Schluesselwortmodus{
			{Name: "core"}, {Name: "custom", Vorgabe: true},
		},
	}
}

// **Die Liste zeigt beides — die des Sets und die des Spiels** —
// #1062.
//
// MSE tut dasselbe: `KeywordList::getItems` hängt erst die des Sets an
// (`fixed = false`) und dann die des Spiels (`fixed = true`) —
// `src/gui/control/keyword_list.cpp`, Zeilen 153–165, Fassung f9d9356 vom
// 2021-07-06.
func TestDieListeZeigtSetUndSpiel(t *testing.T) {
	ablage := &wortablage{woerter: []*domain.Keyword{
		{ID: 7, Word: "Beben", Match: "Beben", Mode: "custom"},
	}}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{spielvorlage(), true})

	sicht, err := pflege.Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(sicht.Woerter) != 3 {
		t.Fatalf("%d Einträge: %+v", len(sicht.Woerter), sicht.Woerter)
	}

	// **Die des Sets zuerst**, und nur sie tragen eine Kennung.
	if sicht.Woerter[0].Wort != "Beben" || sicht.Woerter[0].Fest ||
		sicht.Woerter[0].ID != 7 {
		t.Errorf("das erste ist %+v", sicht.Woerter[0])
	}

	if !sicht.Woerter[1].Fest || sicht.Woerter[1].ID != 0 {
		t.Errorf("das zweite ist %+v", sicht.Woerter[1])
	}

	// **Ein Schlüsselwort ohne Muster ist ungültig** — MSE schaltet es ab,
	// „that would make MSE hang" (`Keyword::prepare`, Zeilen 205–209).
	if sicht.Woerter[2].Gueltig {
		t.Errorf("das Muster-lose gilt als gültig: %+v", sicht.Woerter[2])
	}

	if !sicht.Woerter[0].Gueltig || !sicht.Woerter[1].Gueltig {
		t.Errorf("die gültigen sind %+v", sicht.Woerter[:2])
	}

	// **Und die Modi kommen mit** — ohne sie gäbe es keine Auswahl.
	if len(sicht.Modi) != 2 || sicht.Modi[0] != "core" {
		t.Errorf("die Modi sind %+v", sicht.Modi)
	}
}

// **Ohne Spielpaket bleiben die des Sets** — und die Liste sagt es nicht
// zweimal.
func TestOhneSpielpaketBleibenDieDesSets(t *testing.T) {
	ablage := &wortablage{woerter: []*domain.Keyword{{ID: 7, Word: "Beben", Match: "B"}}}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{})

	sicht, err := pflege.Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(sicht.Woerter) != 1 || sicht.Woerter[0].Fest {
		t.Errorf("die Liste ist %+v", sicht.Woerter)
	}
}

// **Ohne Stilverzeichnis bleiben ebenfalls die des Sets.**
//
// Der Unterschied zum Lauf darüber: dort **gibt** es eine Stilquelle, sie
// findet nur nichts. Hier gibt es gar keine — das ist der Stand eines frisch
// gebauten Dienstes, und er darf nicht abstürzen.
func TestOhneStilquelleBleibenDieDesSets(t *testing.T) {
	ablage := &wortablage{woerter: []*domain.Keyword{{ID: 7, Word: "Beben", Match: "B"}}}

	pflege := service.NeueSchluesselwortpflege(ablage, nil)

	sicht, err := pflege.Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(sicht.Woerter) != 1 || len(sicht.Modi) != 0 {
		t.Errorf("die Liste ist %+v", sicht)
	}
}

// **Ein neues bekommt den Vorgabemodus des Spiels.**
//
//	FOR_EACH(mode, set.game->keyword_modes) {
//	  if (mode->is_default) { keyword.mode = mode->name; break; }
//	}
//
// `AddKeywordAction::AddKeywordAction`, `src/data/action/keyword.cpp`, Zeilen
// 20–32, Fassung f9d9356 vom 2021-07-06.
func TestEinNeuesSchluesselwortBekommtDenVorgabemodus(t *testing.T) {
	ablage := &wortablage{}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{spielvorlage(), true})

	if _, err := pflege.Anlegen(t.Context(), 1); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	if ablage.woerter[0].Mode != "custom" {
		t.Errorf("der Modus ist %q, erwartet der mit »is default«", ablage.woerter[0].Mode)
	}
}

// **Ohne Vorgabemodus bleibt er leer** — erfunden wird keiner.
func TestOhneVorgabemodusBleibtErLeer(t *testing.T) {
	ablage := &wortablage{}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{})

	if _, err := pflege.Anlegen(t.Context(), 1); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	if ablage.woerter[0].Mode != "" {
		t.Errorf("der Modus ist %q", ablage.woerter[0].Mode)
	}
}

// **Setzen und Löschen gehen durch.**
func TestSetzenUndLoeschenGehenDurch(t *testing.T) {
	ablage := &wortablage{woerter: []*domain.Keyword{{ID: 7}}}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{})

	if err := pflege.Setzen(t.Context(), 1, 7, "muster", "Beben"); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if ablage.gesetzt != "muster=Beben" {
		t.Errorf("geschrieben wurde %q", ablage.gesetzt)
	}

	if err := pflege.Loeschen(t.Context(), 1, 7); err != nil {
		t.Errorf("Löschen: %v", err)
	}
}

// **Jeder Weg meldet den Fehler seiner Ablage.**
func TestDieSchluesselwortpflegeMeldetIhreFehler(t *testing.T) {
	ablage := &wortablage{fehler: errors.New("die Platte ist voll")}

	pflege := service.NeueSchluesselwortpflege(ablage, spielwoerter{})

	if _, err := pflege.Liste(t.Context(), 1); err == nil {
		t.Error("die Liste kam durch")
	}

	if _, err := pflege.Anlegen(t.Context(), 1); err == nil {
		t.Error("das Anlegen kam durch")
	}

	if err := pflege.Setzen(t.Context(), 1, 7, "muster", "X"); err == nil {
		t.Error("das Setzen kam durch")
	}

	if err := pflege.Loeschen(t.Context(), 1, 7); err == nil {
		t.Error("das Löschen kam durch")
	}
}
