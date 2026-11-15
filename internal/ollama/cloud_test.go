package ollama_test

// **Was über die Auslagerung feststeht** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Die Serverdatei sagt aus.**
func TestDieServerdateiSagtDieCloudAb(t *testing.T) {
	t.Parallel()

	zuhause := einZuhauseMit(t, `{"disable_ollama_cloud": true}`)

	befund := ollama.LiesCloudstand(zuhause)

	if befund.Stand != ollama.CloudAus {
		t.Errorf("der Stand ist %q", befund.Stand)
	}

	// **Der Beleg nennt die Datei** — sonst steht eine Behauptung ohne
	// Fundstelle da (§10).
	if befund.Beleg == "" {
		t.Error("der Befund nennt keinen Beleg")
	}
}

// **Und sie sagt auch das Gegenteil aus.**
//
// **Die Gegenprobe**: ein Leser, der nur auf `true` sieht, gäbe für `false`
// dasselbe zurück wie für eine fehlende Datei — und das sind zwei
// verschiedene Auskünfte.
func TestDieServerdateiSagtDieCloudAuchAn(t *testing.T) {
	t.Parallel()

	if stand := ollama.LiesCloudstand(
		einZuhauseMit(t, `{"disable_ollama_cloud": false}`)).Stand; stand !=
		ollama.CloudAn {
		t.Errorf("der Stand ist %q", stand)
	}
}

// **Ohne Datei ist es ungewiss und nicht „an".**
//
// # Warum das der Kern ist
//
// **Weil die zweite dokumentierte Abschaltung von hier nicht zu lesen ist.**
// `OLLAMA_NO_CLOUD=1` steht in der Umgebung eines Prozesses, den dieses
// Programm nicht gestartet hat. Wer „keine Datei" als „Cloud an" ausgäbe,
// behauptete etwas über eine Umgebung, die er nicht kennt.
func TestOhneDateiIstEsUngewiss(t *testing.T) {
	t.Parallel()

	if stand := ollama.LiesCloudstand(t.TempDir()).Stand; stand !=
		ollama.CloudUngewiss {
		t.Errorf("der Stand ist %q", stand)
	}
}

// **Ein Schlüssel, der fehlt, ist ebenso ungewiss.**
func TestEinFehlenderSchluesselIstUngewiss(t *testing.T) {
	t.Parallel()

	if stand := ollama.LiesCloudstand(
		einZuhauseMit(t, `{"host":"127.0.0.1"}`)).Stand; stand !=
		ollama.CloudUngewiss {
		t.Errorf("der Stand ist %q", stand)
	}
}

// **Eine kaputte Datei ist ungewiss und nennt den Grund.**
//
// **Nicht „an" und nicht „aus"**: was dort steht, ist unlesbar, und daraus
// folgt nichts über die Laufzeit.
func TestEineKaputteDateiIstUngewissUndNenntDenGrund(t *testing.T) {
	t.Parallel()

	befund := ollama.LiesCloudstand(einZuhauseMit(t, `{kaputt`))

	if befund.Stand != ollama.CloudUngewiss {
		t.Errorf("der Stand ist %q", befund.Stand)
	}

	if befund.Beleg == "" {
		t.Error("eine kaputte Datei blieb ohne Beleg")
	}
}

// einZuhauseMit legt eine Serverdatei an und gibt das Zuhause.
func einZuhauseMit(t *testing.T, inhalt string) string {
	t.Helper()

	zuhause := t.TempDir()

	ordner := filepath.Join(zuhause, filepath.Dir(ollama.Serverdatei))
	if fehler := os.MkdirAll(ordner, 0o750); fehler != nil {
		t.Fatalf("den Ordner anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(
		filepath.Join(zuhause, ollama.Serverdatei), []byte(inhalt), 0o600,
	); fehler != nil {
		t.Fatalf("die Serverdatei schreiben: %v", fehler)
	}

	return zuhause
}

// **Kein Beleg nennt den vollen Pfad.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil der Beleg die Oberfläche erreicht und über MCP und REST hinausgeht.**
// Ein `/Users/<name>/…` darin nennt den Menschen, der hier arbeitet, und den
// Ort seines Zuhauses.
//
// **Gefunden hat es die Sichtprüfung an der laufenden Anwendung** (2026-09-11)
// und kein Test: jeder setzt sein eigenes Zuhause, und in einem Zuhause unter
// `/var/folders/…` fällt ein voller Pfad niemandem auf (§3).
func TestKeinCloudbelegNenntDenVollenPfad(t *testing.T) {
	t.Parallel()

	for name, zuhause := range map[string]string{
		"mit Datei":      einZuhauseMit(t, `{"disable_ollama_cloud": true}`),
		"ohne Datei":     t.TempDir(),
		"kaputte Datei":  einZuhauseMit(t, `{kaputt`),
		"ohne Schlüssel": einZuhauseMit(t, `{"host":"127.0.0.1"}`),
		"unlesbarer Wert": einZuhauseMit(t,
			`{"disable_ollama_cloud": "vielleicht"}`),
		"ausdrücklich an": einZuhauseMit(t, `{"disable_ollama_cloud": false}`),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			beleg := ollama.LiesCloudstand(zuhause).Beleg

			if strings.Contains(beleg, zuhause) {
				t.Errorf("der Beleg nennt das Zuhause: %q", beleg)
			}

			// **Und er nennt die Datei trotzdem** — die Gegenprobe: ein Beleg
			// ohne Fundstelle wäre eine Behauptung (§10).
			if !strings.Contains(beleg, ollama.Serverdatei) {
				t.Errorf("der Beleg nennt die Datei nicht: %q", beleg)
			}
		})
	}
}
