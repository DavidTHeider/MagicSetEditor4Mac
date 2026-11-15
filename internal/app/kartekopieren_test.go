package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kopierprobe merkt sich, welche Karte kopiert wurde.
type kopierprobe struct {
	gesehen int64
	kennung int64
	fehler  error
}

func (k *kopierprobe) KarteKopieren(_ context.Context, _, karteID int64) (int64, error) {
	k.gesehen = karteID

	return k.kennung, k.fehler
}

func mitKopierer(probe *kopierprobe) *app.App {
	anwendung := app.NeueApp(":memory:")
	anwendung.MitKartenkopierer(probe)

	return anwendung
}

// **Die Kopie kommt mit ihrer Kennung zurück** — wer kopiert, will sie
// aufschlagen (#625).
func TestKarteKopierenGibtDieNeueKennung(t *testing.T) {
	probe := &kopierprobe{kennung: 9}

	antwort := mitKopierer(probe).KarteKopieren(probekontext(t), "1", "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.KarteID != "9" {
		t.Errorf("KarteID %q, erwartet `9`", antwort.KarteID)
	}

	if probe.gesehen != 7 {
		t.Errorf("beim Speicher kam %d an, erwartet 7", probe.gesehen)
	}
}

func TestKarteKopierenWeistEineUnlesbareKennungAb(t *testing.T) {
	befund := mitKopierer(&kopierprobe{}).KarteKopieren(probekontext(t), "1", "abc").Fehler
	if befund == nil {
		t.Fatal("kein Befund auf eine unlesbare Kennung")
	}
}

func TestKarteKopierenMeldetEinenFehler(t *testing.T) {
	befund := mitKopierer(&kopierprobe{fehler: domain.ErrNotFound}).
		KarteKopieren(probekontext(t), "1", "7").Fehler
	if befund == nil || befund.Code != "nichtGefunden" {
		t.Errorf("%+v, erwartet nichtGefunden", befund)
	}
}
