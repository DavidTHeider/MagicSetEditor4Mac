//go:build darwin

// **Sensible Teile lassen sich vor dem Sichern herausnehmen** — OPS-04.AK-04
// (#1721):
// „Nutzer sehen Vorschau und können sensible Bestandteile vor Export
// entfernen." Die Vorschau stand seit O10; das Herausnehmen fehlte.
package app_test

import (
	"archive/zip"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinHerausgenommenerTeilFehltImPaket(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	ziel := filepath.Join(t.TempDir(), "bericht"+domain.Diagnosepaketendung)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return ziel, nil })

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("%+v", vorschau.Fehler)
	}

	if antwort := anwendung.DiagnosepaketSichern(t.Context(), []string{string(domain.TeilProtokoll)}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	archiv, fehler := zip.OpenReader(ziel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := archiv.Close(); fehler != nil {
			t.Error(fehler)
		}
	})

	namen := make([]string, 0, len(archiv.File))
	for _, eintrag := range archiv.File {
		namen = append(namen, eintrag.Name)
	}

	if strings.Contains(strings.Join(namen, " "), "protokoll") || len(namen) < 2 {
		t.Errorf("im Paket: %v", namen)
	}
}

// **Ein Paket ohne jeden Teil entsteht nicht**, und ein Teil, den es nicht
// gibt, ist ein Fehler und kein stilles Nichts.
func TestEinPaketBrauchtMindestensEinenBekanntenTeil(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return filepath.Join(t.TempDir(), "bericht"+domain.Diagnosepaketendung), nil
	})

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("%+v", vorschau.Fehler)
	}

	alle := make([]string, 0, len(domain.Diagnoseteile()))
	for _, teil := range domain.Diagnoseteile() {
		alle = append(alle, string(teil))
	}

	for name, ohne := range map[string][]string{"alle heraus": alle, "unbekannt": {"geheimnisse"}} {
		if antwort := anwendung.DiagnosepaketSichern(t.Context(), ohne); antwort.Fehler == nil {
			t.Errorf("%s: angenommen", name)
		}
	}
}
