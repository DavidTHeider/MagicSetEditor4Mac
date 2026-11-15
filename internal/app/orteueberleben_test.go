// **Die Daten überleben eine neue Fassung**
// (#1100,
// Werkstück 6).
//
// # Was hier geprüft wird
//
// **Dass kein Ort, an dem diese Anwendung schreibt, im Bündel liegt.** Ein
// Austausch ersetzt `MagicSetEditor4Mac.app` — alles darin ist danach weg. Was
// daneben liegt, bleibt.
//
// Die Orte entstehen aus [os.UserConfigDir] und [os.UserHomeDir], also aus dem
// **Benutzerverzeichnis**; keiner aus [os.Executable]. Das ist der Grund,
// warum die Daten bleiben, und dieser Wächter hält ihn fest.
//
// # Was er nicht prüft
//
//   - **Den Austausch selbst.** Ein Lauf, der eine Fassung durch eine andere
//     ersetzt und danach denselben Bestand vorfindet, braucht zwei gebaute
//     Bündel; er steht im Freigabelauf (`docs/RELEASE.md`) und nicht hier.
//   - **Ob eine Migration den Bestand heil lässt.** Dafür gibt es die
//     Sicherung vor der Migration
//     (#939)
//     und ihre eigenen Tests.
//   - **Was jemand selbst ins Bündel legt.** Wer dorthin schreibt, verliert es
//     — dagegen hilft kein Wächter, sondern die Dokumentation.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestKeinSchreibortLiegtImBuendel(t *testing.T) {
	t.Parallel()

	zuhause, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Heimatverzeichnis: %v", err)
	}

	// **Die Form eines Bündelpfads, nicht der echte.** Dieser Rechner hat
	// vielleicht keins installiert; geprüft wird, dass keiner der Orte
	// darunter fällt.
	buendel := "/Applications/MagicSetEditor4Mac.app"

	for name, holen := range map[string]func() (string, error){
		"Datenbank":           app.Datenbankpfad,
		"Bilder":              app.Bilderverzeichnis,
		"Downloadverzeichnis": app.Downloadverzeichnis,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort, fehler := holen()
			if fehler != nil {
				t.Fatalf("%s: %v", name, fehler)
			}

			if strings.HasPrefix(ort, buendel) {
				t.Errorf("%s liegt im Bündel: %s", name, ort)
			}

			// **Und er hängt am Benutzerverzeichnis.** Ein Ort, der es nicht
			// tut, hängt an irgendetwas anderem — und was das ist, wüsste
			// niemand, bis eine Aktualisierung es wegnimmt.
			if !strings.HasPrefix(ort, zuhause+string(filepath.Separator)) {
				t.Errorf("%s liegt nicht unter %s: %s", name, zuhause, ort)
			}
		})
	}
}
