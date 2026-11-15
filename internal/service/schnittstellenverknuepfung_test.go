package service_test

// **Ein Ziel, dessen Regeln anderswo stehen** — OPS03
// (#2349):
// die Verknüpfungen in `~/.local/bin` kennt `internal/app` (OPS-03, #1720);
// hier geht es nur darum, dass Lesen, Einrichten und Widerrufen sie wie jedes
// andere Ziel behandeln — fremd bleibt fremd, eingerichtet bleibt still.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verknuepfungIm gibt ein Ziel mit festem Stand, das zählt, was getan wurde.
func verknuepfungIm(
	stand service.Schnittstellenstand, fehler error,
) (ziel service.Schnittstellenziel, getan *[]string) {
	getan = &[]string{}

	return service.Schnittstellenziel{
		Name: "befehle", Art: service.ArtVerknuepfung, Datei: "/heim/.local/bin", Vorhanden: true,
		Verknuepfung: &service.Verknuepfungsziel{
			Stand:      func() service.Schnittstellenstand { return stand },
			Einrichten: func() error { *getan = append(*getan, "einrichten"); return fehler },
			Widerrufen: func() error { *getan = append(*getan, "widerrufen"); return fehler },
		},
	}, getan
}

func TestEinVerknuepfungszielFolgtDenRegelnJedesZiels(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		stand            service.Schnittstellenstand
		einrichten       string
		widerrufen       string
		einrichtenFehler error
		widerrufenFehler error
	}{
		"offen":        {stand: service.SchnittstelleOffen, einrichten: "einrichten"},
		"veraltet":     {stand: service.SchnittstelleVeraltet, einrichten: "einrichten", widerrufen: "widerrufen"},
		"eingerichtet": {stand: service.SchnittstelleEingerichtet, widerrufen: "widerrufen"},
		"fremd":        {stand: service.SchnittstelleFremd, einrichtenFehler: domain.ErrConflict, widerrufenFehler: domain.ErrConflict},
	} {
		ziel, getan := verknuepfungIm(fall.stand, nil)

		if gelesen := service.SchnittstelleLesen(ziel); gelesen != fall.stand {
			t.Errorf("%s: gelesen %s", name, gelesen)
		}

		if _, err := service.SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(err, fall.einrichtenFehler) &&
			(fall.einrichtenFehler != nil || err != nil) {
			t.Errorf("%s: einrichten %v", name, err)
		}

		if _, err := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(err, fall.widerrufenFehler) &&
			(fall.widerrufenFehler != nil || err != nil) {
			t.Errorf("%s: widerrufen %v", name, err)
		}

		erwartet := []string{}
		for _, schritt := range []string{fall.einrichten, fall.widerrufen} {
			if schritt != "" {
				erwartet = append(erwartet, schritt)
			}
		}

		if len(*getan) != len(erwartet) || (len(erwartet) > 0 && (*getan)[0] != erwartet[0]) ||
			(len(erwartet) > 1 && (*getan)[1] != erwartet[1]) {
			t.Errorf("%s: getan %v, erwartet %v", name, *getan, erwartet)
		}
	}
}

// **Ein Schritt, der scheitert, sagt es** — und der Stand danach wird neu
// gelesen.
func TestEinScheiternderVerknuepfungsschrittWirdGemeldet(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("ln: Zugriff verweigert")

	offen, _ := verknuepfungIm(service.SchnittstelleOffen, kaputt)
	if ausgang, err := service.SchnittstelleEinrichten(t.Context(), offen); !errors.Is(err, kaputt) ||
		ausgang.Stand != service.SchnittstelleOffen {
		t.Errorf("einrichten: %+v, %v", ausgang, err)
	}

	eingerichtet, _ := verknuepfungIm(service.SchnittstelleEingerichtet, kaputt)
	if ausgang, err := service.SchnittstelleWiderrufen(t.Context(), eingerichtet); !errors.Is(err, kaputt) ||
		ausgang.Bytegleich {
		t.Errorf("widerrufen: %+v, %v", ausgang, err)
	}
}
