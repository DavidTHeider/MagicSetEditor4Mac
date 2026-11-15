package app

// **Die sechs Bindungen der zwei Dienste** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Was hier gemessen wird und was daneben
//
// **Hier die Fläche**: dass ein Fehlschlag als übersetzter Fehler ankommt und
// nicht als Ausnahme, und dass der Stand auch dann mitkommt. Was ein Posten
// tut, steht in `dienstposten_intern_test.go`.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abweisendeWache sagt zu allem Nein.
type abweisendeWache struct{ gefragt domain.Wirkung }

func (w *abweisendeWache) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	w.gefragt = wirkung

	return domain.MitGrund(domain.ErrValidation, errors.New("keine Rolle"),
		"diese Rolle darf das nicht")
}

// **Ohne die Rolle geht kein Port auf** — und die Fläche sagt es.
func TestOhneRolleGehtKeinPortAuf(t *testing.T) {
	t.Parallel()

	wache := &abweisendeWache{}
	anwendung, gestartet, _ := einePostenhalterin(t,
		"https://0.0.0.0:54591", "geheim", "AB:CD", nil)

	anwendung.MitWache(wache)

	antwort := NeueBindungen(anwendung).RESTStarten(t.Context(), domain.JedeSchnittstelle, 0)
	if antwort.Fehler == nil {
		t.Fatal("die Abweisung kam nicht an")
	}

	if *gestartet != 0 {
		t.Errorf("trotzdem gestartet: %dmal", *gestartet)
	}

	// **Und gefragt wurde nach der richtigen Wirkung** (N7, #1568). Ohne
	// diese Zeile stünde nur, dass **irgendetwas** abgewiesen wurde.
	if wache.gefragt != domain.WirkungExternEmpfangen {
		t.Errorf("gefragt wurde nach %q", wache.gefragt)
	}

	// **Der Stand kommt trotzdem mit** — sonst stünde im Fenster neben der
	// Meldung ein leeres Feld statt „läuft nicht".
	if antwort.Stand.Art != string(domain.DienstREST) || antwort.Stand.Laeuft {
		t.Errorf("der Stand fehlt: %+v", antwort.Stand)
	}
}

// **Die sechs Bindungen führen auf die zwei Posten.**
func TestDieSechsBindungenSchaltenBeideDienste(t *testing.T) {
	t.Parallel()

	anwendung, gestartet, _ := einePostenhalterin(t,
		"http://127.0.0.1:8123", "geheim", "", nil)

	gebunden := NeueBindungen(anwendung)

	if antwort := gebunden.RESTStand(t.Context()); antwort.Stand.Laeuft {
		t.Errorf("REST läuft ungefragt: %+v", antwort.Stand)
	}

	if antwort := gebunden.MCPStand(t.Context()); antwort.Stand.Laeuft {
		t.Errorf("MCP läuft ungefragt: %+v", antwort.Stand)
	}

	if antwort := gebunden.RESTStarten(t.Context(),
		domain.Schleifenadresse, 8123); antwort.Fehler != nil {
		t.Fatalf("RESTStarten: %+v", antwort.Fehler)
	}

	if antwort := gebunden.MCPStarten(t.Context(),
		domain.Schleifenadresse, 8124); antwort.Fehler != nil {
		t.Fatalf("MCPStarten: %+v", antwort.Fehler)
	}

	if *gestartet != 2 {
		t.Errorf("gestartet wurde %dmal", *gestartet)
	}

	if antwort := gebunden.RESTStand(t.Context()); !antwort.Stand.Laeuft ||
		antwort.Stand.Marke != "geheim" {
		t.Errorf("der REST-Stand: %+v", antwort.Stand)
	}

	// **Beenden ist erreicht, wenn nichts mehr läuft** — auch beim zweiten
	// Mal, und deshalb gibt es dafür keinen Fehlerweg.
	for _, beenden := range []func(context.Context) DienstAntwort{
		gebunden.RESTBeenden, gebunden.MCPBeenden,
		gebunden.RESTBeenden, gebunden.MCPBeenden,
	} {
		if antwort := beenden(t.Context()); antwort.Fehler != nil || antwort.Stand.Laeuft {
			t.Errorf("Beenden: %+v", antwort)
		}
	}
}
