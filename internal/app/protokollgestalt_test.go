//go:build darwin

package app_test

// **Die Gestalt eines Nachweises wird in jeder Protokollsenke geschwärzt** —
// SEC-P5 (#2001),
// Befund P5 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **`redigiere` lief nur über das Diagnosepaket.** `mse.log`, das Unified
// Log, die Fehlerausgabe von `mse`, `mse-rest` und `mse-mcp` und die ältere
// Diagnose-Textausgabe schwärzten nur das Benutzerverzeichnis. Ein `grund`
// mit `Authorization: Bearer …` oder einem langen Token blieb stehen — und
// das Diagnosepaket schwärzte ihn erst beim Packen, also nur dort.
//
// # Was hier gemessen wird
//
// **Je Senke ein Lauf**: der Schreiber der Programme ohne Fenster, die
// Protokolldatei über die Weiche — dieselbe, die das Unified Log speist —
// und die Diagnose über einer Datei, die eine ältere Fassung ungeschwärzt
// geschrieben hat.

import (
	"bytes"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Was keine Senke durchlassen darf.
const (
	nachweiswert = "sk-live-abcdef"
	zugangswert  = "glpatA1b2C3d4E5f6G7h8I9j0K1l2M3n4"
)

// verraeterischerGrund trägt beide Gestalten, so wie ein Fehler aus dem
// Unterbau sie tragen könnte.
const verraeterischerGrund = "Authorization: Bearer " + nachweiswert +
	" — https://git.example/x?private_token=" + zugangswert

// pruefeGeschwaerzt sieht nach, dass nichts durchkam und etwas geschwärzt
// wurde — sonst hieße „nichts durchgekommen" nur, dass nichts ankam.
func pruefeGeschwaerzt(t *testing.T, senke, text string) {
	t.Helper()

	for _, verraten := range []string{nachweiswert, zugangswert} {
		if strings.Contains(text, verraten) {
			t.Errorf("%s trägt %q:\n%s", senke, verraten, text)
		}
	}

	if !strings.Contains(text, app.Geschwaerzt) {
		t.Errorf("%s schwärzte nichts — kam die Zeile überhaupt an?\n%s", senke, text)
	}
}

func TestDerSchreiberDerProgrammeSchwaerztDieGestalt(t *testing.T) {
	t.Parallel()

	puffer := &bytes.Buffer{}
	protokoll := app.Serverprotokoll(puffer)

	protokoll.Error("ein Ruf wurde abgewiesen", "grund", verraeterischerGrund)
	protokoll.With("grund", verraeterischerGrund).Error("gebunden")
	protokoll.Error("im Satz: Bearer " + nachweiswert)

	pruefeGeschwaerzt(t, "der Schreiber der Programme", puffer.String())
}

func TestDieProtokolldateiSchwaerztDieGestalt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), app.Protokollname)

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	schreiber.Error("ein Ruf wurde abgewiesen", "grund", verraeterischerGrund)

	if fehler = schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	gelesen, fehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	pruefeGeschwaerzt(t, app.Protokollname, string(gelesen))
}

// **Eine Datei aus einer älteren Fassung** trägt den Nachweis noch — die
// Diagnose liest sie und darf ihn trotzdem nicht zeigen.
func TestDieDiagnoseSchwaerztEinAltesProtokoll(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	// **Am Schreiber vorbei**, wie eine ältere Fassung es tat.
	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatal(fehler)
	}

	alt, fehler := os.OpenFile(pfad, os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- Pfad im Zuhause aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := alt.WriteString("level=ERROR msg=alt grund=\"" + verraeterischerGrund + "\"\n"); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := alt.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	slog.Default().Info("danach")

	antwort := anwendung.Diagnose(t.Context())
	if !strings.Contains(antwort.Text, "msg=alt") {
		t.Fatalf("die alte Zeile steht gar nicht in der Diagnose:\n%s", antwort.Text)
	}

	pruefeGeschwaerzt(t, "die Diagnose", antwort.Text)
}

// **Der Angriff an der Fläche** — ein Zugang in der Adresse eines privaten
// Repositoriums (`?private_token=…`), über die Bindung des Fensters, und der
// Bezug scheitert. Der Grund nennt die Adresse und geht ins Protokoll; der
// Zugang darf dort nicht stehen. Die Gegenstelle ist ein geschlossener Port
// auf der Schleife — es verlässt nichts den Rechner.
func TestEinZugangInDerPaketadresseLandetNichtInDerProtokolldatei(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}

	ort := lauscher.Addr().String()

	if fehler = lauscher.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	antwort := anwendung.PaketBeziehen(t.Context(),
		"http://"+ort+"/gruppe/vorlagen?private_token="+zugangswert,
		"0123456789abcdef0123456789abcdef01234567", "pakete", "geheim")
	if antwort.Fehler == nil {
		t.Fatal("der Bezug von einem geschlossenen Port glückte")
	}

	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatal(fehler)
	}

	gelesen, fehler := os.ReadFile(pfad) // #nosec G304 -- Pfad im Zuhause aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Erst der Beleg, dass der Grund ankam** — sonst hieße „kein Zugang
	// im Protokoll" nur, dass dort nichts steht (§10).
	if !strings.Contains(string(gelesen), "grund=") {
		t.Fatalf("das Protokoll nennt keinen Grund:\n%s", gelesen)
	}

	if strings.Contains(string(gelesen), zugangswert) {
		t.Errorf("der Zugang steht in %s:\n%s", app.Protokollname, gelesen)
	}
}
