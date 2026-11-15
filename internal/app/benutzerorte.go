package app

// **Die schreibenden Orte im Zuhause, an einer Stelle** — A01
// (#2215).
//
// Warum ein Testlauf hier abgewiesen wird, steht in
// `echteszuhausegesperrt_intern_test.go`.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// errEchtesZuhauseImTestlauf weist einen Testlauf ab, der einen schreibenden
// Ort im echten Zuhause verlangt.
var errEchtesZuhauseImTestlauf = errors.New("ein Testlauf verlangt einen Ort im echten Zuhause " +
	"— HOME umlenken (t.Setenv oder TestMain), wie scripts/ci.sh es tut")

// konfigurationsort ist `os.UserConfigDir` mit dem Riegel.
func konfigurationsort() (string, error) { return imZuhause(os.UserConfigDir) }

// zwischenspeicherort ist `os.UserCacheDir` mit dem Riegel.
func zwischenspeicherort() (string, error) { return imZuhause(os.UserCacheDir) }

// heimort ist `os.UserHomeDir` mit dem Riegel.
func heimort() (string, error) { return imZuhause(os.UserHomeDir) }

// imZuhause fragt nach einem Ort — und weist einen Testlauf ab, dessen
// Zuhause nicht unter der Zwischenablage liegt.
//
// **Nicht „ist es das echte?"**, sondern „ist es eines der Läufe?" — die
// erste Frage bräuchte `os/user` und damit den Systemnamen, den
// `TestKeinSystemnameFliesstStillEin` aus dem Produktivcode heraushält;
// die zweite ist strenger und braucht nichts.
func imZuhause(frage func() (string, error)) (string, error) {
	if testlauf && !zuhauseFreiFuerTestlauf(os.Getenv("HOME"), os.TempDir()) {
		return "", errEchtesZuhauseImTestlauf
	}

	return frage()
}

// zuhauseFreiFuerTestlauf sagt, ob `heim` **unter** der Zwischenablage liegt.
//
// **Ein leeres Zuhause ist frei**: dann scheitert das Betriebssystem selbst,
// ohne irgendwo zu schreiben — und der Fehlerweg „ohne Benutzerverzeichnis"
// bleibt der, den ein Lauf prüft.
func zuhauseFreiFuerTestlauf(heim, ablage string) bool {
	if heim == "" {
		return true
	}

	if ablage == "" {
		return false
	}

	innen, fehler := filepath.Rel(filepath.Clean(ablage), filepath.Clean(heim))

	return fehler == nil && innen != "." && !strings.HasPrefix(innen, "..")
}
