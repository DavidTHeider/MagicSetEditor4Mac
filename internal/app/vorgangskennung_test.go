package app_test

// **Die Vorgangskennung an der Fenstergrenze** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineBindungWuerfeltIhreKennungSelbst(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	erste := anwendung.Arbeitskontext(t.Context()).Fehler
	zweite := anwendung.Arbeitskontext(t.Context()).Fehler

	switch {
	case erste == nil || zweite == nil:
		t.Fatal("ohne Autorisierung muss die Bindung einen Fehler geben")
	case erste.Vorgang == "":
		t.Fatal("der Fehler trägt keine Vorgangskennung")
	case len(erste.Vorgang) != 8:
		t.Fatalf("die Kennung %q ist nicht acht Zeichen lang", erste.Vorgang)
	case erste.Vorgang == zweite.Vorgang:
		t.Fatalf("zwei Rufe teilen die Kennung %q — dann ist sie keine", erste.Vorgang)
	}
}

// **Eine mitgebrachte Kennung bleibt stehen.** Die drei Flächen ohne Fenster
// setzen sie in ihrer Wurzel; eine zweite mitten im Vorgang gäbe demselben
// Ruf zwei Namen.
func TestEineVorhandeneKennungBleibtStehen(t *testing.T) {
	t.Parallel()

	ctx := app.MitVorgang(t.Context(), domain.Vorgangskennung("abcdef01"))

	if kennung := app.VorgangAus(app.MitNeuemVorgang(ctx)); kennung != "abcdef01" {
		t.Fatalf("die Kennung wurde zu %q", kennung)
	}

	antwort := app.NeueApp("").Arbeitskontext(ctx)
	if antwort.Fehler == nil || antwort.Fehler.Vorgang != "abcdef01" {
		t.Fatalf("die Antwort trägt %+v", antwort.Fehler)
	}
}

// **Ein Nutzerfehler trägt keine.** Er hat keinen verdeckten Grund und keine
// Protokollzeile — eine Kennung wäre ein Griff ins Leere.
func TestEinNutzerfehlerTraegtKeineKennung(t *testing.T) {
	t.Parallel()

	befund := app.Nutzerfehler(errors.New("nur ein Satz"))
	if befund.Vorgang != "" {
		t.Fatalf("der Nutzerfehler trägt die Kennung %q", befund.Vorgang)
	}
}

// **Und `Uebersetze(nil)` bleibt `nil`** — auch mit dem Stempel davor.
func TestOhneFehlerGibtEsKeinenBefund(t *testing.T) {
	t.Parallel()

	if befund := app.Uebersetze(app.MitNeuemVorgang(t.Context()), nil); befund != nil {
		t.Fatalf("aus nil wurde %+v", befund)
	}
}
