package app_test

// **Auch eine Ablehnung trägt ihre Befunde als Liste** — C01
// (#2226).
//
// Der Vertragsrundgang am nativen Fenster fand `befunde: null` in der Antwort
// von `Fremdpreisverlauf`, wenn der Zugang fehlte — das Fenster erwartet
// `string[]` (`antwortenmarkt.ts`), und eine Seite, die über die Befunde
// läuft, scheitert an `null`, bevor sie die Meldung zeigt.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestJedeAbgelehnteVerlaufsfrageTraegtEineLeereBefundliste(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		ruf  func(context.Context, *testing.T) app.Fremdpreisantwort
	}{
		{"unbekannte Quelle", func(ctx context.Context, t *testing.T) app.Fremdpreisantwort {
			return app.AnbieterAnwendungUeber(frischerBestand(t)).Fremdpreisverlauf(ctx, "cardmarket", "abc", "", nil)
		}},
		{"unbekannte Ausführung", func(ctx context.Context, t *testing.T) app.Fremdpreisantwort {
			return app.AnbieterAnwendungUeber(frischerBestand(t)).Fremdpreisverlauf(ctx, "scrydex", "abc", "glaenzend", nil)
		}},
		{"ohne Bestand", func(ctx context.Context, _ *testing.T) app.Fremdpreisantwort {
			return app.NeueApp("").Fremdpreisverlauf(ctx, "scrydex", "abc", "", nil)
		}},
		{"ohne Zugang", func(ctx context.Context, t *testing.T) app.Fremdpreisantwort {
			return app.AnbieterAnwendungUeber(frischerBestand(t)).MitMarktbund(marktbund{}).
				Fremdpreisverlauf(ctx, "scrydex", "abc", "", nil)
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := fall.ruf(probekontext(t), t)
			if antwort.Fehler == nil {
				t.Fatalf("keine Ablehnung: %+v", antwort)
			}

			roh, fehler := json.Marshal(antwort)
			if fehler != nil {
				t.Fatal(fehler)
			}

			if !strings.Contains(string(roh), `"befunde":[]`) {
				t.Errorf("die Befunde sind keine leere Liste: %s", roh)
			}
		})
	}
}
