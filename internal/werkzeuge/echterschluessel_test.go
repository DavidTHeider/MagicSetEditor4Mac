package werkzeuge

// **Kein Lauf signiert mit dem Schlüsselbund dieses Rechners** — SEC03
// (#2356):
// „echte Schlüssel bleiben außerhalb der automatischen Probe".
//
// Bis SEC03 signierten zwei Testdateien mit `gpg --local-user
// BF81923EA5D38635` — dem Freigabeschlüssel, sobald er im Schlüsselbund lag,
// also in jedem `make ci` auf dem Rechner, der auch freigibt. Seither
// signiert ein Probeschlüssel: im Prozess über [MitFreigabebund], in den
// gebauten Programmen über den Tag `probeanker`.
//
// # Was er prüft
//
// Keine Testdatei — Go, Shell, JavaScript, TypeScript — ruft `gpg` mit
// `--local-user`, `--detach-sign` oder `--clearsign`, oder in einer Zeile
// mit `gpg` mit `--sign`, `-s` oder `-b`.
//
// # Was er nicht prüft
//
//   - **Einen anderen Weg an den Schlüssel**, etwa `security` oder ein
//     Agent über seine Steckdose — dafür gibt es keinen Aufruf, und einer,
//     der dazukäme, wäre ein Befund für die Durchsicht.
//   - **`scripts/releasesignieren.sh`** — das Signieren einer Freigabe ist
//     sein Zweck, und es ist kein Test.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// gpgSignieren trifft die Schalter, mit denen `gpg` signiert — die langen,
// die nur `gpg` kennt, und `--sign`, `-s`, `-b` in einer Zeile mit `gpg`
// (`codesign --sign -` ist ein anderes Programm).
var gpgSignieren = regexp.MustCompile(`--(local-user|detach-sign|clearsign)\b|\bgpg\b.*\s(--sign|-[a-zA-Z]*[sb][a-zA-Z]*)\b`)

func TestKeinLaufSigniertMitDemSchluesselbundDesRechners(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")
	gelesen := 0

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() {
			switch eintrag.Name() {
			case ".go", ".git", "node_modules", "graft", "dist", ".huellenprobe":
				return filepath.SkipDir
			}

			return nil
		}

		name := eintrag.Name()
		if !strings.HasSuffix(name, "_test.go") && !strings.Contains(name, ".test.") {
			return nil
		}

		// **Dieser Wächter nennt die Schalter selbst**, in seinem Kopf.
		if name == "echterschluessel_test.go" {
			return nil
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Datei im Repositorium
		if lesefehler != nil {
			return lesefehler
		}

		gelesen++

		for nummer, zeile := range strings.Split(string(roh), "\n") {
			if gpgSignieren.MatchString(zeile) {
				t.Errorf("%s:%d: %s — ein Test signiert nicht mit dem Schlüsselbund dieses Rechners",
					pfad, nummer+1, strings.TrimSpace(zeile))
			}
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("abgehen: %v", fehler)
	}

	if gelesen < 1000 {
		t.Fatalf("nur %d Testdateien gelesen — der Weg trifft nicht", gelesen)
	}
}

// **Und kein Freigabeweg baut mit dem Probeanker.** Unter dem Tag
// `probeanker` liest ein Programm seinen Anker aus einer Umgebungsvariablen
// (`internal/aktualisierung/probeanker.go`) — in einem ausgelieferten wäre
// das eine Hintertür. Gelesen werden die Bauwege; **und der eine Lauf, der
// es setzen soll, setzt es**, sonst prüfte der Wächter eine Abwesenheit, die
// es auch ohne ihn gäbe.
func TestKeinFreigabewegBautMitProbeanker(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	// **Die Bauwege, nicht die Prüfwege**: `ci-lint.sh` und
	// `vorpruefung.sh` lintet das Tag, und das sollen sie.
	bauwege, fehler := filepath.Glob(filepath.Join(wurzel, "scripts", "release*.sh"))
	if fehler != nil || len(bauwege) < 5 {
		t.Fatalf("die Freigabeskripte: %v (%d)", fehler, len(bauwege))
	}

	bauwege = append(bauwege, filepath.Join(wurzel, "scripts", "bauen.sh"),
		filepath.Join(wurzel, ".github", "workflows", "ci.yml"))

	for _, pfad := range bauwege {
		if strings.HasSuffix(pfad, ".test.sh") {
			continue
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Datei im Repositorium
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		if strings.Contains(string(roh), "probeanker") {
			t.Errorf("%s nennt das Tag probeanker — kein Freigabeweg baut damit", pfad)
		}
	}

	flaechen, fehler := os.ReadFile(filepath.Join(wurzel, "internal", "flaechen", "abnahme_test.go")) //nolint:gosec // Datei im Repositorium
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !strings.Contains(string(flaechen), `"-tags", "probeanker"`) {
		t.Error("internal/flaechen baut seine Programme nicht mit dem Probeanker — dann signierte es wieder mit einem echten Schlüssel oder gar nicht")
	}
}
