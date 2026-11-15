package app

// **Die Doppel der KI-Fehlerwege** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Eine eigene Datei**, weil `kifehlerwege_intern_test.go` mit ihnen über
// die Zeilengrenze aus §2 ginge.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// festeLaufzeit gibt immer dieselben Modelle.
type festeLaufzeit struct{ modelle []domain.KIModell }

func (f festeLaufzeit) Fassung(context.Context) (string, error) {
	return "0.12.3", nil
}

func (f festeLaufzeit) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	return f.modelle, nil, nil
}

func (f festeLaufzeit) Vorschlagen(
	context.Context, string, string, string, map[string]any,
) (KIRohantwort, error) {
	return KIRohantwort{}, errors.New("diese Laufzeit redet nicht")
}

// stummerLeser gibt für jede Karte einen Fehler.
type stummerLeser struct{}

func (stummerLeser) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, errors.New("die Karte ist nicht zu lesen")
}

func (stummerLeser) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return map[string]string{}, nil
}

// stummerSchreiber schreibt nichts und wird nie gefragt.
type stummerSchreiber struct{}

func (stummerSchreiber) FeldSetzenMitVorwert(
	context.Context, int64, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

func (stummerSchreiber) FeldSetzenVonAussen(
	context.Context, int64, int64, string, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

func (stummerSchreiber) FelderVonAussen(
	context.Context, int64, int64, []domain.Value, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

func (stummerSchreiber) FelderVonAussenMitHerkunft(
	context.Context, int64, int64, []domain.Value, string,
	[]domain.KIFeldherkunft, map[string][]domain.KIBildparameter, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

// stummesJournal reserviert alles und bucht nichts.
//
// **Nur `Reservieren` tut etwas** — die übrigen Methoden gibt es, weil
// [Kostenfuehrer] sie verlangt, und sie werden hier nie gerufen.
type stummesJournal struct{}

func (stummesJournal) Reservieren(
	_ context.Context, auftrag service.Kostenauftrag, _ string,
) (domain.Kostenposten, error) {
	return domain.Kostenposten{Kennung: auftrag.Kennung}, nil
}

func (stummesJournal) Stand(context.Context, string, domain.Zeitraum) (
	domain.Haushaltsstand, error,
) {
	return domain.Haushaltsstand{}, nil
}

func (stummesJournal) Posten(context.Context, string, domain.Zeitraum) (
	[]domain.Kostenposten, error,
) {
	return nil, nil
}

func (stummesJournal) BudgetSetzen(
	context.Context, string, domain.Zeitraum, domain.Betrag,
) error {
	return nil
}

func (stummesJournal) Abschliessen(context.Context, string, domain.Betrag) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

func (stummesJournal) Freigeben(context.Context, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

func (stummesJournal) AlsUnklarMerken(context.Context, string, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}
