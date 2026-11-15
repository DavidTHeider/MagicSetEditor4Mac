// Der Wächter über den Namen der Bestandsdatei in der Dokumentation.
//
// # Wofür
//
// `docs/RELEASE.md` beschrieb Sichern und Zurückspielen **sechsmal** über
// `mse.db`. Die Datei heißt `bestand.db`; wer den Weg befolgt hätte, hätte
// `sqlite3` mit einer leeren Datei gefüttert und eine Sicherung weggetragen,
// die nichts enthält
// (#872).
//
// Aufgefallen ist es beim **Gehen** des Wegs, nicht beim Lesen — und genau
// deshalb steht der Abnahmekasten von Phase Q so, wie er steht.
//
// # Was er prüft
//
// Der Name aus `internal/app/orte.go` steht in jeder Datei, die den Ort des
// Bestands nennt — und **kein anderer Dateiname mit `.db`** steht dort, außer
// den ausdrücklich erlaubten.
//
// # Was er nicht prüft
//
// **Ob der beschriebene Weg funktioniert.** Er sieht einen Namen, nicht ein
// Ergebnis; dass `VACUUM INTO` und das Zurückspielen tragen, ist am
// 2026-08-22 von Hand an einem Abzug des echten Bestands gegangen worden und
// steht in `docs/RELEASE.md` mit seinen Zahlen.
package werkzeuge

import (
	"os"
	"regexp"
	"testing"
)

// dateiname trifft den Namen in `filepath.Join(…, "bestand.db")`.
var dateiname = regexp.MustCompile(`"([a-z-]+\.db)"`)

// entschuldigung trifft **genau** den Satzteil, der einen alten Namen als
// Fehler benennt — `und nicht \`mse.db\“.
//
// Er darf dastehen: eine Erklärung, warum es den falschen Namen gab, ist keine
// Wiederholung des Fehlers. Alles andere in der Datei wird geprüft.
var entschuldigung = regexp.MustCompile("und nicht `[a-zA-Z0-9-]+\\.db`")

// erlaubteDatenbanknamen sind die `.db`-Namen, die in der Doku stehen dürfen —
// **jeder mit seinem Grund**.
//
// **Was hier steht, ist nicht aus dem Quelltext gelesen** und damit die
// schwächere Sorte Eintrag. `bestand-vor-rueckspielung.db` steht deshalb
// **nicht** hier, sondern wird wie der Bestandsname selbst aus dem Quelltext
// geholt ([rueckspielnameAusQuelltext]) — sonst liefen Doku und Konstante
// auseinander, und genau davor gibt es diesen Wächter.
var erlaubteDatenbanknamen = map[string]string{
	"mse-sicherung.db": "der Beispielname des **Ziels** einer Sicherung; er gehört den Nutzenden und nicht der Anwendung",
}

// dokumenteMitBestandsort sind die Dateien, die den Ort des Bestands nennen.
var dokumenteMitBestandsort = []string{
	"../../docs/RELEASE.md",
	"../../README.md",
}

func TestDieDokumentationNenntDenEchtenBestandsnamen(t *testing.T) {
	echt := bestandsnameAusQuelltext(t)
	beiseite := rueckspielnameAusQuelltext(t)

	for _, pfad := range dokumenteMitBestandsort {
		roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
		if err != nil {
			t.Fatalf("%s lesen: %v", pfad, err)
		}

		// **Der Satz, der den alten Namen als Fehler benennt, wird
		// herausgeschnitten** — und zwar er allein.
		//
		// Der erste Anlauf prüfte stattdessen mit `strings.Contains` über die
		// **ganze** Datei, ob irgendwo „und nicht `x`" steht. Damit war jede
		// Fundstelle von `mse.db` entschuldigt, sobald der Satz einmal
		// dastand — die Gegenprobe blieb grün, und ein Wächter, der bei
		// einem eingebauten Fehler nichts sagt, ist keiner (§10).
		text := entschuldigung.ReplaceAllString(string(roh), "")

		for _, treffer := range regexp.MustCompile(`\b([a-zA-Z0-9-]+\.db)\b`).
			FindAllStringSubmatch(text, -1) {
			name := treffer[1]

			// **Die Nebendateien des WAL-Modus zählen mit** — `bestand.db-wal`
			// trifft das Muster als `bestand.db`, und das ist richtig so.
			if name == echt || name == beiseite {
				continue
			}

			if _, erlaubt := erlaubteDatenbanknamen[name]; erlaubt {
				continue
			}

			t.Errorf("%s nennt %q — der Bestand heißt %q (internal/app/orte.go)",
				pfad, name, echt)
		}
	}
}

// bestandsnameAusQuelltext liest den Namen aus `Datenbankpfad`.
func bestandsnameAusQuelltext(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile("../app/orte.go") //nolint:gosec // ein fester Pfad
	if err != nil {
		t.Fatalf("orte.go lesen: %v", err)
	}

	treffer := dateiname.FindStringSubmatch(string(roh))
	if treffer == nil {
		t.Fatal("in orte.go steht kein Dateiname mit .db — der Leser greift daneben")
	}

	return treffer[1]
}

// rueckspielnameAusQuelltext liest `app.Rueckspielsicherung`.
//
// **Aus dem Quelltext und nicht aus einer Liste hier** (#940): der Name steht
// in `docs/RELEASE.md` und in `CHANGELOG.md`, und eine Umbenennung, die die
// Doku vergisst, ist genau der Fehler, für den es diesen Wächter gibt.
func rueckspielnameAusQuelltext(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile("../service/bestandersetzen.go") //nolint:gosec // ein fester Pfad
	if err != nil {
		t.Fatalf("bestandersetzen.go lesen: %v", err)
	}

	treffer := regexp.MustCompile(`Rueckspielsicherung = "([a-z-]+\.db)"`).
		FindStringSubmatch(string(roh))
	if treffer == nil {
		t.Fatal("in bestandersetzen.go steht keine Rueckspielsicherung — der Leser greift daneben")
	}

	return treffer[1]
}
