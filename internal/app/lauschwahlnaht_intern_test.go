package app

import (
	"flag"
	"io"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieLauschwahlflaggeSammeltJedeNennung(t *testing.T) {
	t.Parallel()

	fahne := flag.NewFlagSet("probe", flag.ContinueOnError)
	fahne.SetOutput(io.Discard)

	genannt := Lauschwahlflagge(fahne, "eine Adresse")
	if fehler := fahne.Parse([]string{"-adresse", "127.0.0.1", "-adresse", "::1,192.0.2.1"}); fehler != nil {
		t.Fatal(fehler)
	}

	wahl, fehler := LauschwahlAusGenannten(genannt())
	if fehler != nil || !reflect.DeepEqual(wahl.Adressen(), []string{"127.0.0.1", "::1", "192.0.2.1"}) {
		t.Errorf("gelesen %v (%v)", wahl.Adressen(), fehler)
	}
}

func TestDieLauschwahlDesDienstes(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)

	for _, fall := range []struct {
		name, einstellung string
		genannt           []string
		erwartet          []string
	}{
		{"ohne alles die Schleife", "", nil, []string{"127.0.0.1"}},
		{"ohne Nennung die Einstellung", "::1,127.0.0.1", nil, []string{"::1", "127.0.0.1"}},
		{"die Nennung schlägt die Einstellung", "::1", []string{"127.0.0.1"}, []string{"127.0.0.1"}},
	} {
		setze(domain.NetzLauschen, fall.einstellung)

		wahl, fehler := anwendung.LauschwahlFuerDienst(riegelkontext(t), fall.genannt)
		if fehler != nil || !reflect.DeepEqual(wahl.Adressen(), fall.erwartet) {
			t.Errorf("%s: %v (%v)", fall.name, wahl.Adressen(), fehler)
		}
	}
}

func TestOhneLesbareEinstellungenKeineLauschwahl(t *testing.T) {
	t.Parallel()

	anwendung, _ := einDienst(t)
	anwendung.MitEinstellungen(klagendeEinstellungen{})

	if _, fehler := anwendung.LauschwahlFuerDienst(riegelkontext(t), nil); fehler == nil {
		t.Error("ohne lesbare Einstellungen kam eine Wahl zustande")
	}
}

func TestRechnerteileUndNachAussen(t *testing.T) {
	t.Parallel()

	teile := Rechnerteile([]string{"127.0.0.1:8080", "[::1]:8080", "[fe80::1%en0]:8080", "kein-paar"})
	if !reflect.DeepEqual(teile, []string{"127.0.0.1", "::1", "fe80::1", "kein-paar"}) {
		t.Errorf("die Teile sind %v", teile)
	}

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,192.0.2.1,::1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	if draussen := NachAussen(wahl.Bindungen(0)); !reflect.DeepEqual(draussen, []string{"192.0.2.1"}) {
		t.Errorf("nach außen gehen %v", draussen)
	}
}
