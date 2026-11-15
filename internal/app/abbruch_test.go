package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein abgebrochener Lauf ist kein Fehler** — er ist die Absicht (§7).
//
// # Woran es aufgefallen ist
//
// Die Vorschau bricht den vorigen Lauf ab, sobald ein neuer beginnt. Über der
// gezeichneten Karte stand dabei „Es ist ein unerwarteter Fehler aufgetreten" —
// die Meldung des abgebrochenen Laufs, der genau so abgebrochen wurde, wie es
// vorgesehen ist.
//
// **Der Mangel ist älter als die Vorschau beim Tippen**
// (#288): wer
// schnell zwischen zwei Karten wechselte, sah ihn ebenso. Nur tippt niemand so
// schnell zwischen Karten, wie er Buchstaben tippt.
func TestEinAbbruchIstKeinUnerwarteterFehler(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		fehler error
	}{
		{"abgebrochen", context.Canceled},
		{"abgelaufen", context.DeadlineExceeded},
		{"umschlossen", fmt.Errorf("die Karte zeichnen: %w", context.Canceled)},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befund := app.Uebersetze(t.Context(), fall.fehler)
			if befund == nil {
				t.Fatal("kein Befund")
			}

			if befund.Code != app.CodeAbgebrochen {
				t.Errorf("Code %q, erwartet %q", befund.Code, app.CodeAbgebrochen)
			}
		})
	}
}

// **Was dieser Wächter nicht prüft, und warum das in Ordnung ist.**
//
// Er sieht auf den **Code** und nicht auf Schlüssel und Meldung. Die tragen
// weiterhin `unbekannt` und „Es ist ein unerwarteter Fehler aufgetreten." —
// also genau den Satz, gegen den dieser Test antritt.
//
// **Der Grund ist, dass die Bedeutung im Code reist.** `useLaden` fragt
// `ursache.code === 'abgebrochen'` (`frontend/src/haken/useLaden.ts`) und zeigt
// den Befund gar nicht; die Meldung erreicht niemanden. Sie hier zu übersetzen
// hieße, einen Text zu pflegen, den kein Fenster je darstellt.
//
// **Aufgefallen ist das am 2026-08-21** an der Gegenprobe zu
// [TestEineZweiteVorschauBeendetDieErste]: dort stand der Satz in der Ausgabe,
// und der Name dieses Tests versprach, dass er das nicht mehr tut.

// **Und alles andere bleibt, was es war.**
func TestEinGewoehnlicherFehlerBleibtUnbekannt(t *testing.T) {
	t.Parallel()

	befund := app.Uebersetze(t.Context(), errors.New("die Platte ist voll"))
	if befund == nil || befund.Code != app.CodeUnbekannt {
		t.Errorf("%+v", befund)
	}
}
