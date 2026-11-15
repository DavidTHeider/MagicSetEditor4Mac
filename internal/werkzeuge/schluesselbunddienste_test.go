// **Jeder Schlüsselbunddienst steht in der Administration** — Sichtprüfung
// über die ganze Anwendung, 2026-09-07.
//
// # Warum es diesen Wächter gibt
//
// **`docs/ADMINISTRATION.md` sagte, wo alles liegt — und nirgends, wie man es
// wieder loswird.** Der Abschnitt „Außerbetriebnahme" ist deshalb neu, und er
// nennt vier Dienste im Schlüsselbund, die das Löschen des Anwendungsordners
// **überleben**: Zugangsdaten fremder Kartendienste, die `client_secret` der
// Anmeldeanbieter, die TOTP-Geheimnisse und der Schlüssel des verschlüsselten
// Bestands.
//
// **Eine Liste von Hand veraltet still.** Ein fünfter Dienst im Quelltext ohne
// Zeile dort hieße: jemand entfernt die Anwendung nach Anleitung und lässt ein
// Geheimnis zurück, von dem er nichts weiß.
//
// # Was er hält und was nicht
//
//   - **Er hält beide Richtungen**: jeder Dienstname aus dem Quelltext steht
//     in der Datei, und die Datei erfindet keinen.
//   - **Er hält nicht**, ob der Löschbefehl daneben stimmt. Dass `security
//     delete-generic-password` so heißt und so wirkt, sieht nur ein Mensch
//     (`AGENTS.md` §10).
//   - **Und nicht, ob die Beschreibung zutrifft.** „die TOTP-Geheimnisse" ist
//     eine Aussage über den Inhalt, keine über den Namen.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// dienstImQuelltext findet `= "MagicSetEditor4Mac…"` in einer Konstante.
//
// **Nur Konstanten**: derselbe Text steht als Aussteller in `internal/totp`
// und als Kenner im `User-Agent`, und beides ist kein Schlüsselbunddienst.
var dienstImQuelltext = regexp.MustCompile(
	`(?m)^const\s+\w*[Dd]ienst\s*=\s*"(MagicSetEditor4Mac[A-Za-z-]*)"`)

// dienstInDerTabelle findet `| \x60MagicSetEditor4Mac…\x60 |` in der Datei.
var dienstInDerTabelle = regexp.MustCompile(
	"\\|\\s*`(MagicSetEditor4Mac[A-Za-z-]*)`\\s*\\|")

func TestJederSchluesselbunddienstStehtInDerAdministration(t *testing.T) {
	t.Parallel()

	imCode := map[string]bool{}

	for _, wurzel := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join("..", "..", wurzel),
			func(pfad string, eintrag os.DirEntry, gehfehler error) error {
				if gehfehler != nil {
					return gehfehler //nolint:wrapcheck // der Aufrufer meldet ihn
				}

				if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
					strings.HasSuffix(pfad, "_test.go") {
					return nil
				}

				roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus dem Repositorium
				if lesefehler != nil {
					return lesefehler //nolint:wrapcheck // der Aufrufer meldet ihn
				}

				for _, treffer := range dienstImQuelltext.FindAllStringSubmatch(
					string(roh), -1) {
					imCode[treffer[1]] = true
				}

				return nil
			})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", wurzel, fehler)
		}
	}

	// **Die Gegenprobe zur Messung**: findet sie gar nichts, ist sie grün und
	// sagt nichts (§10). Am 2026-09-07 waren es vier.
	if len(imCode) < 3 {
		t.Fatalf("nur %d Dienste gefunden — heißen die Konstanten noch "+
			"`…dienst`?", len(imCode))
	}

	inDerDatei := map[string]bool{}
	for _, treffer := range dienstInDerTabelle.FindAllStringSubmatch(
		lies(t, filepath.Join("docs", "ADMINISTRATION.md")), -1) {
		inDerDatei[treffer[1]] = true
	}

	var fehlt, erfunden []string

	for name := range imCode {
		if !inDerDatei[name] {
			fehlt = append(fehlt, name)
		}
	}

	for name := range inDerDatei {
		if !imCode[name] {
			erfunden = append(erfunden, name)
		}
	}

	if len(fehlt) > 0 {
		sort.Strings(fehlt)
		t.Errorf("diese Schlüsselbunddienste stehen nicht in der Tabelle der "+
			"Außerbetriebnahme: %s — wer die Anwendung nach Anleitung "+
			"entfernt, lässt ihre Geheimnisse zurück", strings.Join(fehlt, ", "))
	}

	if len(erfunden) > 0 {
		sort.Strings(erfunden)
		t.Errorf("die Tabelle nennt Dienste, die es im Quelltext nicht gibt: "+
			"%s", strings.Join(erfunden, ", "))
	}
}
