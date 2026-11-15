// Die Doubles für den Scryfall-Zugang
// (#1005).
//
// **Eine eigene Datei**, weil `scryfall_test.go` sonst über die Zeilengrenze
// liefe (§2) — und weil sie eine eigene Frage beantworten: was der Fassade
// vorgesetzt wird, nicht was sie tut.
package app_test

import (
	"context"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Doubles.

func quelleMitVorschlag() app.Scryfallquelle {
	return &feste{vorschlag: vorschlagMitAenderung()}
}

// vorschlagMitAenderung gibt einen Vorschlag mit genau einer Änderung.
func vorschlagMitAenderung() *service.Vorschlag {
	return &service.Vorschlag{
		Aenderungen: []service.Feldvorschlag{{Feld: "rule text", Alt: "Alt", Neu: "Neu"}},
		Herkunft:    "https://beispiel.invalid/karte",
		Alter:       3 * time.Hour,
	}
}

func quelleDieBeimSchreibenScheitert() app.Scryfallquelle {
	return &feste{vorschlag: vorschlagMitAenderung(), schreibfehler: domain.ErrConflict}
}

type feste struct {
	vorschlag     *service.Vorschlag
	schreibfehler error
}

func (f *feste) VorschlagFuerKarte(
	context.Context, string, map[string]string, int64, string,
) (*service.Vorschlag, error) {
	return f.vorschlag, nil
}

func (f *feste) Uebernehmen(
	context.Context, int64, int64, *service.Vorschlag, []string,
) ([]service.Feldvorschlag, error) {
	return nil, f.schreibfehler
}

type zaehlendequelle struct {
	feste

	rufe int
}

func (z *zaehlendequelle) VorschlagFuerKarte(
	ctx context.Context, spiel string, jetzige map[string]string,
	karteID int64, name string,
) (*service.Vorschlag, error) {
	z.rufe++

	return z.feste.VorschlagFuerKarte(ctx, spiel, jetzige, karteID, name)
}

type quelleScheitert struct{}

func (quelleScheitert) VorschlagFuerKarte(
	context.Context, string, map[string]string, int64, string,
) (*service.Vorschlag, error) {
	return nil, errors.New("die Gegenstelle antwortet nicht")
}

func (quelleScheitert) Uebernehmen(
	context.Context, int64, int64, *service.Vorschlag, []string,
) ([]service.Feldvorschlag, error) {
	return nil, nil
}

// bestandMitSet gibt einen Leser mit genau einem Set.
func bestandMitSet() app.Bestandsleser { return bestandfest{} }

type bestandfest struct{}

func (bestandfest) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Name: "Probe", Game: "magic"}}, nil
}

func (bestandfest) Karten(context.Context, int64) ([]*domain.CardRow, error) { return nil, nil }

func (bestandfest) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) {
	return nil, nil
}

// bestandOhneSets kennt keines.
func bestandOhneSets() app.Bestandsleser { return bestandleer{} }

type bestandleer struct{ bestandfest }

func (bestandleer) Setkoepfe(context.Context) ([]domain.SetHead, error) { return nil, nil }

// bestandScheitert kann die Köpfe nicht lesen.
type bestandScheitert struct{ bestandfest }

func (bestandScheitert) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand antwortet nicht")
}

// karteMitFeldern gibt eine Karte aus Set 1.
func karteMitFeldern() app.Kartenleser { return kartefest{setID: 1} }

// karteAusAnderemSet gibt eine Karte, die zu Set 9 gehört.
func karteAusAnderemSet() app.Kartenleser { return kartefest{setID: 9} }

type kartefest struct{ setID int64 }

func (k kartefest) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return &domain.CardRow{
		ID: 2, SetID: k.setID,
		Values: []domain.Value{{Field: "rule text", Text: "Alt"}},
	}, nil
}

type karteleserScheitert struct{}

func (karteleserScheitert) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, errors.New("die Karte lässt sich nicht lesen")
}

func (kartefest) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return nil, nil
}
