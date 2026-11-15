package domain_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieLauschwahlLiestUndPrueft(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		text     string
		adressen []string
		nachAus  bool
		fehler   bool
	}{
		{"leer heißt die Schleife", "", []string{"127.0.0.1"}, false, false},
		{"Leerzeichen heißen die Schleife", "  ", []string{"127.0.0.1"}, false, false},
		{"eine Adresse", "192.168.178.74", []string{"192.168.178.74"}, true, false},
		{"mehrere, mit Komma und Leerzeichen", "127.0.0.1, 192.168.178.74 ,fe80::1", []string{"127.0.0.1", "192.168.178.74", "fe80::1"}, true, false},
		{"die Schleife allein verlässt nichts", "127.0.0.1,::1", []string{"127.0.0.1", "::1"}, false, false},
		{"alle, ausdrücklich", "0.0.0.0", []string{"0.0.0.0"}, true, false},
		{"ein Name wird abgewiesen", "localhost", nil, false, true},
		{"eine doppelte Adresse wird abgewiesen", "127.0.0.1,127.0.0.1", nil, false, true},
		{"alle neben einer einzelnen wird abgewiesen", "0.0.0.0,192.168.178.74", nil, false, true},
		{"ein leerer Eintrag wird abgewiesen", "127.0.0.1,,192.168.178.74", nil, false, true},
		{"eine Adresse mit Port wird abgewiesen", "127.0.0.1:8080", nil, false, true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wahl, fehler := domain.LauschwahlAus(fall.text)
			if fall.fehler {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Fatalf("%q wurde angenommen: %v", fall.text, wahl)
				}

				return
			}

			if fehler != nil {
				t.Fatalf("%q: %v", fall.text, fehler)
			}

			if !reflect.DeepEqual(wahl.Adressen(), fall.adressen) {
				t.Errorf("Adressen %v, erwartet %v", wahl.Adressen(), fall.adressen)
			}

			if wahl.VerlaesstDenRechner() != fall.nachAus {
				t.Errorf("verlässt den Rechner: %v, erwartet %v", wahl.VerlaesstDenRechner(), fall.nachAus)
			}

			if domain.PruefeLauschwahl(fall.text) != nil {
				t.Errorf("die Prüfung der Einstellung weist %q ab", fall.text)
			}
		})
	}
}

func TestDieLauschwahlGibtJeAdresseEineBindung(t *testing.T) {
	t.Parallel()

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,fe80::1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	bindungen := wahl.Bindungen(8080)
	if len(bindungen) != 2 || bindungen[0].Adresstext() != "127.0.0.1:8080" || bindungen[1].Adresstext() != "[fe80::1]:8080" {
		t.Errorf("die Bindungen sind %+v", bindungen)
	}

	if wahl.Text() != "127.0.0.1,fe80::1" {
		t.Errorf("der Text ist %q", wahl.Text())
	}
}

func TestDieLauschwahlIstEineEinstellungDesGeraets(t *testing.T) {
	t.Parallel()

	if ebene, bekannt := domain.EbeneVon(domain.NetzLauschen); !bekannt || ebene != domain.EbeneGeraet {
		t.Errorf("%s liegt auf %v (%v)", domain.NetzLauschen, ebene, bekannt)
	}

	if fehler := domain.PruefeEinstellung(domain.NetzLauschen, "localhost"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die Einstellung nahm einen Namen an: %v", fehler)
	}
}
