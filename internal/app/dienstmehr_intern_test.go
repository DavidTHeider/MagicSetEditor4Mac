package app

// **Der Dienst aus dem Fenster auf mehreren Adressen** — G08
// (#2272).
//
// Jede Adresse der Wahl geht als eigene Fahne an das Programm, und was es
// ansagt — jede Adresse und jede fehlende —, steht danach im Stand des
// Fensters.

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeAdresseDerWahlGehtAlsFahneHinaus(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", `
echo "Argumente $*" >&2
for a in "$@"; do :; done
echo "Dienst auf http://127.0.0.1:8123"
echo "Dienst auf http://[::1]:8123"
echo "Ohne 192.0.2.1: bind: can't assign requested address"
echo "Marke geheim"
echo "Gesagt $*" > "$(dirname "$0")/argumente"
read -r _
`)

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,::1,192.0.2.1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	lauf, fehler := starteDienstprogrammAufAllen(context.Background(), imOrt(ort), domain.DienstREST, wahl.Bindungen(8123))
	if fehler != nil {
		t.Fatalf("starteDienstprogrammAufAllen: %v", fehler)
	}

	t.Cleanup(func() { lauf.beenden(); <-lauf.fertig })

	gesagt := liesDatei(t, ort, "argumente")
	if !strings.Contains(gesagt, "--adresse 127.0.0.1 --adresse ::1 --adresse 192.0.2.1 --port 8123") {
		t.Errorf("die Fahnen waren %q", gesagt)
	}

	if !reflect.DeepEqual(lauf.adressen, []string{"http://127.0.0.1:8123", "http://[::1]:8123"}) ||
		len(lauf.fehlende) != 1 || lauf.fehlende[0].Adresse != "192.0.2.1" {
		t.Errorf("die Ansage kam als %+v an", lauf)
	}
}

func TestDerStandNenntJedenLauscherUndJedeFehlende(t *testing.T) {
	t.Parallel()

	anwendung, _ := einDienst(t)
	anwendung.MitDienststarter(func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {
		return dienstlauf{
			adresse: "http://127.0.0.1:8123", adressen: []string{"http://127.0.0.1:8123", "http://[::1]:8123"},
			fehlende: []domain.Lauschbefund{{Adresse: "192.0.2.1", Grund: "weg"}},
			marke:    "geheim", beenden: func() {}, fertig: geschlossen(),
		}, nil
	})

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,::1,192.0.2.1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	stand, fehler := anwendung.dienstStartenAuf(riegelkontext(t), domain.DienstREST, wahl.Bindungen(8123))
	if fehler != nil {
		t.Fatalf("dienstStartenAuf: %v", fehler)
	}

	t.Cleanup(func() { anwendung.dienstBeenden(domain.DienstREST) })

	if !reflect.DeepEqual(stand.Lauscher, []string{"http://127.0.0.1:8123", "http://[::1]:8123"}) ||
		len(stand.Fehlende) != 1 || stand.Fehlende[0].Adresse != "192.0.2.1" || stand.Fehlende[0].Grund != "weg" {
		t.Errorf("der Stand ist %+v", stand)
	}
}

func liesDatei(t *testing.T, ort, name string) string {
	t.Helper()

	for range 200 {
		roh, fehler := os.ReadFile(filepath.Join(ort, name)) //nolint:gosec // eine Datei dieses Laufs
		if fehler == nil && len(roh) > 0 {
			return string(roh)
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("%s blieb leer", name)

	return ""
}

func geschlossen() chan struct{} {
	fertig := make(chan struct{})
	close(fertig)

	return fertig
}
