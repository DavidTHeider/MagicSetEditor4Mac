// **Eine eigene Kopie eines Pakets** — FMT-02.AK-04 und S10.NAT-01
// (#1691).
//
// Mitgelieferte und fremdverwaltete Pakete werden über eigene Kopien
// angepasst; die Kopie verändert die Quelle nicht.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kopierplatz baut einen Verwalter über der Saat — sie ist mitgeliefert, denn
// sie liegt außerhalb des verwalteten Orts.
func kopierplatz(t *testing.T) (ort string, verwalter *service.Paketverwalter, katalog *ablage) {
	t.Helper()

	ort = t.TempDir()
	katalog = &ablage{}
	verwalter = service.NeuerPaketverwalter(katalog, ort, saatwurzel(), ort)

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	return ort, verwalter, katalog
}

// ohneKurzname nimmt die Zeile `short name:` heraus.
func ohneKurzname(inhalt string) string {
	zeilen := []string{}

	for _, zeile := range strings.Split(inhalt, "\n") {
		if !strings.HasPrefix(zeile, "short name:") {
			zeilen = append(zeilen, zeile)
		}
	}

	return strings.Join(zeilen, "\n")
}

func TestEineKopieIstEigenUndDasOriginalBleibt(t *testing.T) {
	ort, verwalter, katalog := kopierplatz(t)

	vorher := datei(t, filepath.Join(saatwurzel(), "probe-schlicht.mse-style", "style"))

	kopie, err := verwalter.Kopieren(t.Context(), "mse-style", "probe-schlicht", "")
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	if kopie.Art != "mse-style" || kopie.Name != "probe-schlicht-kopie" ||
		kopie.Pfad != filepath.Join(ort, "probe-schlicht-kopie.mse-style") ||
		kopie.HerkunftsartIn(ort) != domain.PaketEigen {
		t.Fatalf("die Kopie: %+v", kopie)
	}

	// **Das Original bleibt, wie es war** (S10.NAT-01).
	if nachher := datei(t, filepath.Join(saatwurzel(), "probe-schlicht.mse-style", "style")); nachher != vorher {
		t.Error("das Original hat sich geändert")
	}

	// **Die Kopie trägt alles** — samt Beiwerk — und einen eigenen Kurznamen,
	// sonst stünde »Schlicht« in jeder Stilwahl zweimal.
	stil := datei(t, filepath.Join(kopie.Pfad, "style"))
	if !strings.Contains(stil, "\nshort name: schlicht-kopie\n") || ohneKurzname(stil) != ohneKurzname(vorher) {
		t.Errorf("der Stil der Kopie:\n%s", stil)
	}

	for _, beiwerk := range []string{"hell.png", "dunkel.png"} {
		if _, fehler := os.Stat(filepath.Join(kopie.Pfad, beiwerk)); fehler != nil {
			t.Errorf("%s fehlt: %v", beiwerk, fehler)
		}
	}

	gefunden := false

	for _, paket := range katalog.pakete {
		gefunden = gefunden || paket.Pfad == kopie.Pfad
	}

	if !gefunden {
		t.Errorf("die Kopie steht nicht im Katalog: %+v", katalog.pakete)
	}
}

func TestDieKopieNimmtDenNaechstenFreienNamen(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)

	for _, erwartet := range []string{"probe-kopie", "probe-kopie-2", "probe-kopie-3"} {
		kopie, err := verwalter.Kopieren(t.Context(), "mse-game", "probe", "")
		if err != nil || kopie.Name != erwartet {
			t.Fatalf("erwartet %s: %+v, %v", erwartet, kopie, err)
		}

		// **Ein Spiel hat kein Präfix**: der Kurzname ist der ganze Name.
		if spiel := datei(t, filepath.Join(kopie.Pfad, "game")); !strings.Contains(spiel, "\nshort name: "+erwartet+"\n") {
			t.Errorf("das Spiel %s:\n%s", erwartet, spiel)
		}
	}
}

func TestEinGewaehlterNameBekommtDasSpielVorangestellt(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)

	kopie, err := verwalter.Kopieren(t.Context(), "mse-style", "probe-schlicht", "meins")
	if err != nil || kopie.Name != "probe-meins" {
		t.Fatalf("die Kopie: %+v, %v", kopie, err)
	}

	if stil := datei(t, filepath.Join(kopie.Pfad, "style")); !strings.Contains(stil, "\nshort name: meins\n") {
		t.Errorf("der Kurzname:\n%s", stil)
	}
}

func TestEineKopieWirdAbgewiesenOhneHalbenStand(t *testing.T) {
	ort, verwalter, _ := kopierplatz(t)

	if _, err := verwalter.Kopieren(t.Context(), "mse-style", "probe-schlicht", "meins"); err != nil {
		t.Fatalf("die erste Kopie: %v", err)
	}

	var feldfehler *domain.FieldError

	for name, fall := range map[string]struct {
		art, paket, neu string
		pruefe          func(error) bool
	}{
		"unbekanntes Paket": {"mse-style", "probe-fehlt", "", func(err error) bool {
			return errors.Is(err, domain.ErrNotFound)
		}},
		"unbekannte Art": {"mse-set", "probe", "", func(err error) bool {
			return errors.Is(err, domain.ErrNotFound)
		}},
		"ungültiger Name": {"mse-style", "probe-schlicht", "mein stil", func(err error) bool {
			return errors.As(err, &feldfehler) && feldfehler.Key == domain.SchluesselPaketnameUngueltig
		}},
		"vergebener Name": {"mse-style", "probe-schlicht", "meins", func(err error) bool {
			return errors.Is(err, domain.ErrConflict)
		}},
	} {
		if _, err := verwalter.Kopieren(t.Context(), fall.art, fall.paket, fall.neu); !fall.pruefe(err) {
			t.Errorf("%s: %v", name, err)
		}
	}

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		namen = append(namen, eintrag.Name())
	}

	if strings.Join(namen, ",") != "probe-meins.mse-style" {
		t.Errorf("im verwalteten Ort: %v", namen)
	}
}

// eigenesSpiel legt ein Spiel im verwalteten Ort an und gibt seinen Ort.
func eigenesSpiel(t *testing.T, verwalter *service.Paketverwalter, name string) string {
	t.Helper()

	angelegt, err := verwalter.Anlegen(t.Context(), service.Neuanlage{Art: "mse-game", Name: name, Vollname: name})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	return angelegt.Pfad
}

// gesperrt nimmt einem Pfad jedes Recht — und gibt es nach dem Test zurück.
func gesperrt(t *testing.T, pfad string, rechte os.FileMode) {
	t.Helper()

	if err := os.Chmod(pfad, 0); err != nil {
		t.Fatalf("sperren: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, rechte) }) //nolint:errcheck // Aufräumen
}

func TestDieKopieReichtFehlerDurch(t *testing.T) {
	ort, verwalter, katalog := kopierplatz(t)

	verschwunden := eigenesSpiel(t, verwalter, "verschwunden")

	mitOrdner := eigenesSpiel(t, verwalter, "mitordner")
	if err := os.Mkdir(filepath.Join(mitOrdner, "tief"), 0o750); err != nil {
		t.Fatalf("den Ordner anlegen: %v", err)
	}

	gesperrt(t, filepath.Join(mitOrdner, "tief"), 0o750)

	mitDatei := eigenesSpiel(t, verwalter, "mitdatei")
	if err := os.WriteFile(filepath.Join(mitDatei, "notiz.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	gesperrt(t, filepath.Join(mitDatei, "notiz.txt"), 0o600)

	// **Erst jetzt fort**: jede Anlage gleicht den Katalog mit der Platte ab,
	// und ein früher entferntes Paket stünde dann gar nicht mehr darin.
	if err := os.RemoveAll(verschwunden); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	for _, name := range []string{"verschwunden", "mitordner", "mitdatei"} {
		if _, err := verwalter.Kopieren(t.Context(), "mse-game", name, ""); err == nil ||
			errors.Is(err, domain.ErrNotFound) && name != "verschwunden" {
			t.Errorf("%s: %v", name, err)
		}

		if _, err := os.Stat(filepath.Join(ort, name+"-kopie.mse-game")); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s: eine halbe Kopie liegt da: %v", name, err)
		}
	}

	// **Der Katalog wird nicht gefragt** — ein klemmender hält die Kopie
	// nicht auf, gelesen wird die Platte.
	katalog.lesefehler = errors.New("der Katalog klemmt")

	if _, err := verwalter.Kopieren(t.Context(), "mse-game", "probe", ""); err != nil {
		t.Errorf("mit klemmendem Katalog: %v", err)
	}
}
