package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// anlegeprobe merkt sich, für welches Set eine Karte angelegt wurde.
type anlegeprobe struct {
	setID   int64
	kennung int64
	fehler  error
}

func (a *anlegeprobe) KarteAnlegen(_ context.Context, setID int64) (int64, error) {
	a.setID = setID

	return a.kennung, a.fehler
}

func (a *anlegeprobe) FeldSetzenMitVorwert(
	context.Context, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

// FeldSetzenVonAussen schreibt wie oben und ignoriert die Herkunft (#888).
func (a *anlegeprobe) FelderVonAussen(
	ctx context.Context, karteID int64, werte []domain.Value, _ string,
) (vorher string, hatte bool, fehler error) {
	for _, wert := range werte {
		if _, _, fehler := a.FeldSetzenMitVorwert(
			ctx, karteID, wert.Field, wert.Text); fehler != nil {
			return "", false, fehler
		}
	}

	return "", false, nil
}

func (a *anlegeprobe) FeldSetzenVonAussen(
	ctx context.Context, karteID int64, feld, wert, _ string,
) (vorher string, hatte bool, fehler error) {
	return a.FeldSetzenMitVorwert(ctx, karteID, feld, wert)
}

// **Eine Karte lässt sich anlegen** —
// #562.
//
// Bis hierher konnte der Bestand Sets **einlesen** und Feldwerte ändern; eine
// neue Karte gab es nirgends.
func TestKarteAnlegenGibtDieNeueKennung(t *testing.T) {
	probe := &anlegeprobe{kennung: 42}

	anwendung := app.NeueApp("")
	anwendung.MitKartenanleger(probe)

	antwort := anwendung.KarteAnlegen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 7 {
		t.Errorf("angelegt im Set %d", probe.setID)
	}

	// **Die Kennung kommt zurück**, denn wer eine Karte anlegt, will sie
	// aufschlagen.
	if antwort.KarteID != "42" {
		t.Errorf("Kennung %q", antwort.KarteID)
	}
}

// **Am Feld `setID` und nicht am gemeinsamen `id`** (#832).
//
// Hier stand `"id"`, und damit war `KarteAnlegen` der einzige Schreibweg, der
// seine Kennung anders benannte als die sechs daneben — aufgefallen ist es
// dem Zweitblick über Phase O an einem Kommentar, der von sieben Wegen sprach,
// während die Aufzählung sechs führte.
func TestKarteAnlegenWeistEineUnlesbareKennungAb(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitKartenanleger(&anlegeprobe{})

	befund := anwendung.KarteAnlegen(t.Context(), "abc").Fehler
	if befund == nil || befund.Feld != "setID" {
		t.Errorf("%+v", befund)
	}
}

func TestKarteAnlegenMeldetEinenFehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitKartenanleger(&anlegeprobe{fehler: domain.ErrNotFound})

	befund := anwendung.KarteAnlegen(t.Context(), "7").Fehler
	if befund == nil || befund.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", befund)
	}
}
