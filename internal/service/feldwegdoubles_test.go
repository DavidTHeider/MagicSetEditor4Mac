package service_test

// **Die Doubles des Feldwegs** — Lesen, Felder und Ablage. Aus
// `feldweg_test.go` herausgelöst, als die Ablage für das Gesamtreview über
// 0.9.0 den Stand und den Nachzug mitschreiben musste und die Datei an die
// Zeilengrenze aus §2 kam.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wegkarten gibt eine Karte und zählt mit, wonach gefragt wurde.
type wegkarten struct {
	karte  *domain.CardRow
	fehler error
}

func (w wegkarten) KarteLesen(_ context.Context, _ int64) (*domain.CardRow, error) {
	return w.karte, w.fehler
}

// wegfelder gibt die Kartenfelder eines Spiels.
type wegfelder struct {
	felder []domain.Field
	fehler error
}

func (w wegfelder) Kartenfelder(_ context.Context, _ int64) ([]domain.Field, error) {
	return w.felder, w.fehler
}

// wegablage merkt sich, was geschrieben wurde.
type wegablage struct {
	geschrieben []string
	fehler      error

	// reihen zählt die Aufrufe von [wegablage.FelderSetzenMitVorwert] — eine
	// Reihe ist **eine** Transaktion (#714).
	reihen int

	// herkuenfte hält fest, was als Herkunft ankam (#888).
	herkuenfte []string

	// feldherkuenfte hält fest, was eine Übernahme mitschickte (A11, #1450).
	feldherkuenfte []domain.KIFeldherkunft

	// erwartet hält fest, welcher Stand als Bedingung ankam (Befund 2 des
	// Zweitblicks über Phase 5.3).
	erwartet []string

	// nachgezogen hält fest, welche Felder als nachgezogen ankamen
	// (Gesamtreview über 0.9.0, Befund 1).
	nachgezogen []string

	// massenfeld und massenerwartung hält, was eine Massenänderung schickte
	// (#2516).
	massenfeld      string
	massenerwartung []domain.Feldvorwert
}

// FeldSetzenVonAussen schreibt wie oben und merkt sich die Herkunft.
func (w *wegablage) FeldSetzenVonAussen(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	w.herkuenfte = append(w.herkuenfte, herkunft)

	return w.FeldSetzenMitVorwert(ctx, karteID, feld, wert)
}

// FelderSetzenVonAussen schreibt eine Reihe und merkt sich die Herkunft.
func (w *wegablage) FelderSetzenVonAussen(
	ctx context.Context, karteID int64, werte []domain.Value, herkunft, erwartet string,
) (vorher string, hatte bool, fehler error) {
	w.herkuenfte = append(w.herkuenfte, herkunft)
	w.erwartet = append(w.erwartet, erwartet)

	return w.FelderSetzenMitVorwert(ctx, karteID, werte)
}

// FelderSetzenMitHerkunft schreibt Werte samt Feldherkunft (A11, #1450).
func (w *wegablage) FelderSetzenMitHerkunft(
	ctx context.Context, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	_ map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	w.herkuenfte = append(w.herkuenfte, herkunft)
	w.feldherkuenfte = append(w.feldherkuenfte, herkuenfte...)
	w.erwartet = append(w.erwartet, erwartet)

	return w.FelderSetzenMitVorwert(ctx, karteID, werte)
}

func (w *wegablage) FeldSetzenMitVorwert(
	_ context.Context, _ int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	if w.fehler != nil {
		return "", false, w.fehler
	}

	w.geschrieben = append(w.geschrieben, feld+"="+wert)

	return "alt", true, nil
}

// **Die Reihe geht in **einem** Aufruf hinaus** (#714) — die Mitschrift hält
// deshalb fest, wie viele Aufrufe es waren, nicht nur was geschrieben wurde.
func (w *wegablage) FelderSetzenMitVorwert(
	_ context.Context, _ int64, werte []domain.Value,
) (vorher string, hatte bool, fehler error) {
	if w.fehler != nil {
		return "", false, w.fehler
	}

	w.reihen++

	for _, wert := range werte {
		w.geschrieben = append(w.geschrieben, wert.Field+"="+wert.Text)

		if wert.Nachgezogen {
			w.nachgezogen = append(w.nachgezogen, wert.Field)
		}
	}

	return "alt", true, nil
}
