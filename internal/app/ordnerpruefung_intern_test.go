package app

// **Die Frage an die Platte, Zweig für Zweig** — OPS-02.AK-01
// (#1719):
// Existenz, Ordner, Schreibbarkeit und Platz, je mit eigenem Satz.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDiePlatteSagtWarumEinOrdnerNichtTaugt(t *testing.T) {
	wurzel := t.TempDir()
	datei := filepath.Join(wurzel, "eine-datei")
	gesperrt := filepath.Join(wurzel, "gesperrt")

	if fehler := os.WriteFile(datei, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Mkdir(gesperrt, 0o500); fehler != nil {
		t.Fatal(fehler)
	}

	heim := func() (string, error) { return wurzel, nil }
	viel := func(string) (int64, error) { return 1 << 30, nil }
	wenig := func(string) (int64, error) { return 1 << 10, nil }
	unbekannt := func(string) (int64, error) { return 0, errors.New("kein Statfs") }

	faelle := []struct {
		name   string
		angabe string
		platz  func(string) (int64, error)
		satz   string
	}{
		{"eine Datei statt eines Ordners", datei, viel, "kein Ordner"},
		{"ein Ordner ohne Schreibrecht", gesperrt, viel, "nicht schreiben"},
		{"ein volles Laufwerk", "~", wenig, "nur noch"},
		{"ein Platz, der sich nicht bestimmen lässt, hält nicht an", "~", unbekannt, ""},
		{"ein Ordner mit Platz", "~", viel, ""},
	}

	for _, fall := range faelle {
		_, fehler := pruefeOrdnerMit(fall.angabe, heim, fall.platz)

		if fall.satz == "" {
			if fehler != nil {
				t.Errorf("%s: %v", fall.name, fehler)
			}

			continue
		}

		if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), fall.satz) {
			t.Errorf("%s: %v", fall.name, fehler)
		}
	}

	if _, fehler := pruefeOrdnerMit("~", func() (string, error) { return "", errors.New("kein Zuhause") }, viel); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Zuhause: %v", fehler)
	}

	// **Ein Ordner, in den niemand hineinsehen darf**, gibt keinen
	// Existenzfehler, sondern den der Platte.
	innen := filepath.Join(gesperrt, "innen")
	if fehler := os.Chmod(gesperrt, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := os.Chmod(gesperrt, 0o700); fehler != nil { //nolint:gosec // ein Ordner dieses Tests
			t.Error(fehler)
		}
	})

	if _, fehler := pruefeOrdnerMit(innen, heim, viel); !errors.Is(fehler, domain.ErrValidation) || strings.Contains(fehler.Error(), "nicht erreichbar") {
		t.Errorf("ein Ordner hinter einem gesperrten: %v", fehler)
	}
}

func TestGroesseSchreibtMegabyte(t *testing.T) {
	if text := groesse(3 << 19); text != "1.5 MB" {
		t.Errorf("%q", text)
	}
}

// **Ein Ordner, der taugt, kommt gedehnt zurück** — `~` zum Zuhause.
func TestEinTauglicherOrdnerKommtGedehntZurueck(t *testing.T) {
	heim := t.TempDir()

	ordner, fehler := pruefeOrdnerMit("~", func() (string, error) { return heim, nil },
		func(string) (int64, error) { return 1 << 30, nil })
	if fehler != nil || ordner != heim {
		t.Errorf("%q, %v", ordner, fehler)
	}
}

// **Jeder Befund der Platte trägt seinen Schlüssel und den Ordner** — damit
// das Fenster den Satz in seiner Sprache zeigt statt „nicht gültig".
func TestJederBefundDerPlatteTraegtSeinenSchluessel(t *testing.T) {
	wurzel := t.TempDir()
	datei := filepath.Join(wurzel, "eine-datei")

	if fehler := os.WriteFile(datei, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	heim := func() (string, error) { return wurzel, nil }
	viel := func(string) (int64, error) { return 1 << 30, nil }
	wenig := func(string) (int64, error) { return 1 << 10, nil }

	for angabe, schluessel := range map[string]domain.Schluessel{
		filepath.Join(wurzel, "fehlt"): domain.SchluesselOrdnerFehlt,
		datei:                          domain.SchluesselOrdnerIstDatei,
	} {
		_, fehler := pruefeOrdnerMit(angabe, heim, viel)

		var befund *domain.FieldError
		if !errors.As(fehler, &befund) || befund.Key != schluessel || befund.Werte["ordner"] != angabe {
			t.Errorf("%s: %+v", angabe, befund)
		}
	}

	_, fehler := pruefeOrdnerMit("~", heim, wenig)

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Key != domain.SchluesselOrdnerVoll ||
		befund.Werte["ordner"] != wurzel || befund.Werte["frei"] == "" {
		t.Errorf("voll: %+v", befund)
	}
}
