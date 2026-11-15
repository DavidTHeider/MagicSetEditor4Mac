// **Wie der MCP-Prozess sich ausweist** — Werkstück O2 von Phase 5.2
// (#1426),
// nach [0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// # Datei oder Umgebung, nie ein Argument
//
// **Ein Argument steht in der Prozessliste** — `ps` zeigt es jedem Nutzer des
// Rechners, und die Einrichtung eines MCP-Clients schreibt es in eine
// Konfiguration, die mitgeteilt wird. Eine Datei mit den Rechten `0600` sieht
// nur, wer sie angelegt hat; eine Umgebungsvariable sieht nur der Prozess.
//
// # Warum der Altbetrieb bleibt
//
// **Weil dieser Prozess über stdio auf demselben Rechner läuft** und bis zu
// diesem Werkstück ohne Ausweis lief. Wer keine Datei und keine Variable
// setzt, bekommt, was er hatte: das gezeigte Profil, ohne Wache. Der
// Vertrag steht in `docs/ADMINISTRATION.md`; das Token macht daraus einen
// Zugang, der begrenzt, widerrufbar und je Ruf geprüft ist.

package mcpnaehte

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Die zwei Wege, auf denen ein Token hereinkommt.
//
// **`#nosec G101`**: gosec sieht das Wort `Token` neben einer Zeichenkette und
// hält sie für ein Geheimnis. Es ist der **Name** der Umgebungsvariablen; das
// Geheimnis steht in ihr, nie hier.
const (
	Tokendateivariable = "MSE_TOKEN_DATEI" // #nosec G101 -- der Name der Variablen
	Tokenvariable      = "MSE_TOKEN"       // #nosec G101 -- der Name der Variablen
)

// TokenAusUmgebung liest das Token des Prozesses — leer heißt Altbetrieb.
//
// **Die Datei geht vor**: wer beides setzt, meint die Datei, denn sie ist der
// sicherere Weg. Und eine Datei, die andere lesen dürfen, wird **abgewiesen**,
// nicht gelesen — ein Geheimnis, das schon auf der Platte für jeden lesbar
// war, ist keines mehr.
func TokenAusUmgebung(
	getenv func(string) string, lies func(string) ([]byte, error),
	rechte func(string) (fs.FileMode, error),
) (string, error) {
	if pfad := strings.TrimSpace(getenv(Tokendateivariable)); pfad != "" {
		modus, fehler := rechte(pfad)
		if fehler != nil {
			return "", fmt.Errorf("die Tokendatei %s ansehen: %w", pfad, fehler)
		}

		if modus.Perm()&0o077 != 0 {
			return "", fmt.Errorf("die Tokendatei %s ist für andere lesbar (%s) — "+
				"sie braucht die Rechte 0600", pfad, modus.Perm())
		}

		roh, fehler := lies(pfad)
		if fehler != nil {
			return "", fmt.Errorf("die Tokendatei %s lesen: %w", pfad, fehler)
		}

		if klartext := strings.TrimSpace(string(roh)); klartext != "" {
			return klartext, nil
		}

		return "", errors.New("die Tokendatei " + pfad + " ist leer")
	}

	return strings.TrimSpace(getenv(Tokenvariable)), nil
}

// Dateirechte gibt die Rechte einer Datei.
func Dateirechte(pfad string) (fs.FileMode, error) {
	angaben, fehler := os.Stat(pfad)
	if fehler != nil {
		return 0, fehler
	}

	return angaben.Mode(), nil
}
