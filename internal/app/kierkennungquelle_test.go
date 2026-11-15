package app_test

// **Wer Bilder lesen kann und wer nicht** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum eine eigene Datei
//
// **Weil sie eine Frage an die Naht beantwortet und keine an das Fenster** —
// und weil `kierkennungfenster_test.go` mit ihr an die Zeilengrenze aus §2
// kam. [app.KISichtquelle] ist eine **wahlfreie** Schnittstelle: eine
// Laufzeit erfüllt sie oder nicht, und geprüft wird das mit einer
// Typzusicherung. Der Zweig dahinter ist der einzige, den das gewöhnliche
// Probedoppel nicht betreten kann — es **kann** sehen.

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// blindelaufzeit ist eine Laufzeit **ohne** [app.KISichtquelle].
//
// # Wofür es sie gibt
//
// **Für den Zweig, den kein anderer Lauf betritt.** `sichtquelleAus` fragt
// mit einer Typzusicherung, ob eine Laufzeit Bilder lesen kann — und das
// gewöhnliche Probedoppel kann es. Der Lauf über „ein Anbieter ohne Sicht"
// scheitert deshalb eine Stufe früher, an der **Fähigkeit** des Modells, und
// sagt über diese Zusicherung nichts.
//
// **Sie ist der Fall aus dem Betrieb**: `internal/ollama` liest keine Bilder
// über den Weg, den diese Naht braucht, und das Fenster muss es sagen, statt
// abzustürzen.
type blindelaufzeit struct{ modelle []domain.KIModell }

func (b *blindelaufzeit) Fassung(context.Context) (string, error) {
	return "0.12.3", nil
}

func (b *blindelaufzeit) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	return b.modelle, nil, nil
}

func (b *blindelaufzeit) Vorschlagen(
	context.Context, string, string, string, map[string]any,
) (app.KIRohantwort, error) {
	return app.KIRohantwort{}, nil
}

// **Eine Laufzeit, die keine Bilder liest, wird gar nicht erst gefragt.**
//
// **Und die Meldung nennt den Anbieter** — damit niemand rät, welcher von
// vieren gemeint ist.
func TestEineLaufzeitOhneSichtWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})
	welt.anwendung.MitKILaufzeit(func(string) app.KILaufzeitquelle {
		return &blindelaufzeit{modelle: welt.laufzeit.modelle}
	})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})

	if antwort.Fehler == nil {
		t.Fatal("die blinde Laufzeit kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "ollama") {
		t.Errorf("die Meldung nennt den Anbieter nicht: %q",
			antwort.Fehler.Meldung)
	}
}
