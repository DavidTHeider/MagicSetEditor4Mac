package werkzeuge_test

// **Identität bleibt lokal** — ID-01
// (#1700,
// [0128](../../docs/entscheidungen/0128-identitaet-bleibt-lokal.md)).
//
// # Was er hält
//
//   - **Kein Dienst verschickt Post** (AK-02): kein `net/smtp`, keine
//     Versandbibliothek in `go.mod`. Eine Adresse ist eine Kennung und kein
//     Kanal, den dieses Programm bedient.
//   - **Ein Nutzer trägt keinen Nachweis über sich** (AK-02): kein Feld, das
//     „verifiziert" oder „bestätigt" heißt. Bestätigt wird eine
//     **Gegenstelle**, über ihren Schlüssel — nicht eine Adresse.
//   - **Kein Systemname und keine Kontaktkarte fließt still ein** (AK-04):
//     weder `os/user` noch das Kontakte-Framework. ID-04 bringt die
//     Kontakte mit Zustimmung; dann bekommt dieser Wächter eine benannte
//     Ausnahme und nicht ein Loch.
//
// # Was er nicht hält
//
// **Ob die Oberfläche eine Adresse als geprüft bezeichnet** — das ist ein
// Satz in einem Katalog, und den liest ein Mensch. Und nicht, dass eine
// Gegenstelle wirklich über ihren Schlüssel bestätigt wird; das halten die
// Läufe in `internal/peer` und `internal/service`.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// produktivdateien gibt jede Go-Datei ohne Test unter den Wurzeln.
func produktivdateien(t *testing.T) []string {
	t.Helper()

	var dateien []string

	for _, wurzel := range []string{"../../internal", "../../cmd"} {
		fehler := filepath.WalkDir(wurzel,
			func(pfad string, eintrag os.DirEntry, fehler error) error {
				if fehler == nil && !eintrag.IsDir() &&
					strings.HasSuffix(pfad, ".go") &&
					!strings.HasSuffix(pfad, "_test.go") {
					dateien = append(dateien, pfad)
				}

				return fehler
			})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", wurzel, fehler)
		}
	}

	return append(dateien, "../../main.go")
}

// importiertIrgendwo nennt die Dateien, die eines der Pakete importieren.
func importiertIrgendwo(t *testing.T, pakete ...string) []string {
	t.Helper()

	var funde []string

	for _, datei := range produktivdateien(t) {
		gelesen, fehler := parser.ParseFile(token.NewFileSet(), datei, nil,
			parser.ImportsOnly)
		if fehler != nil {
			t.Fatalf("%s zerlegen: %v", datei, fehler)
		}

		for _, eingebunden := range gelesen.Imports {
			pfad := strings.Trim(eingebunden.Path.Value, `"`)

			for _, paket := range pakete {
				if pfad == paket || strings.HasPrefix(pfad, paket+"/") {
					funde = append(funde, datei+": "+pfad)
				}
			}
		}
	}

	return funde
}

// **Kein Dienst verschickt Post** (AK-02).
func TestKeinDienstVerschicktPost(t *testing.T) {
	t.Parallel()

	if funde := importiertIrgendwo(t, "net/smtp", "net/mail"); len(funde) > 0 {
		t.Errorf("hier würde Post verschickt oder gelesen — ID-01.AK-02 "+
			"schließt einen Versand aus:\n%s", strings.Join(funde, "\n"))
	}

	roh, fehler := os.ReadFile("../../go.mod")
	if fehler != nil {
		t.Fatalf("go.mod lesen: %v", fehler)
	}

	versand := regexp.MustCompile(`(?i)gomail|sendgrid|mailgun|mailjet|postmark|go-mail`)
	if treffer := versand.FindString(string(roh)); treffer != "" {
		t.Errorf("go.mod führt eine Versandbibliothek: %s", treffer)
	}
}

// nachweisname trifft einen Feldnamen, der einen Nachweis behauptet.
var nachweisname = regexp.MustCompile(`(?i)verifi|bestaetig|geprueft|nachgewiesen`)

// **Ein Nutzer trägt keinen Nachweis über sich** (AK-02).
func TestEinNutzerTraegtKeinenNachweis(t *testing.T) {
	t.Parallel()

	gelesen, fehler := parser.ParseFile(token.NewFileSet(),
		"../../internal/domain/nutzer.go", nil, 0)
	if fehler != nil {
		t.Fatalf("nutzer.go zerlegen: %v", fehler)
	}

	gefunden := false

	ast.Inspect(gelesen, func(knoten ast.Node) bool {
		typ, ist := knoten.(*ast.TypeSpec)
		if !ist || typ.Name.Name != "Nutzer" {
			return true
		}

		gefunden = true

		struktur, istStruktur := typ.Type.(*ast.StructType)
		if !istStruktur {
			t.Fatalf("Nutzer ist kein Verbund, sondern %T", typ.Type)
		}

		for _, feld := range struktur.Fields.List {
			for _, name := range feld.Names {
				if nachweisname.MatchString(name.Name) {
					t.Errorf("Nutzer.%s behauptet einen Nachweis — eine "+
						"Adresse ist eine Kennung, kein geprüfter Kanal "+
						"(ID-01.AK-02)", name.Name)
				}
			}
		}

		return false
	})

	// **Die Gegenprobe gegen einen umbenannten Typ** — ohne sie stimmte der
	// Lauf jeder Datei zu, in der es keinen Nutzer mehr gibt.
	if !gefunden {
		t.Fatal("internal/domain/nutzer.go führt keinen Typ Nutzer")
	}
}

// **Kein Systemname und keine Kontaktkarte fließt still ein** (AK-04).
func TestKeinSystemnameFliesstStillEin(t *testing.T) {
	t.Parallel()

	funde := importiertIrgendwo(t, "os/user", "github.com/tmc/apple/contacts")

	for _, datei := range produktivdateien(t) {
		roh, fehler := os.ReadFile(datei) //nolint:gosec // feste Wurzeln im Baum
		if fehler != nil {
			t.Fatalf("%s lesen: %v", datei, fehler)
		}

		if strings.Contains(string(roh), "CNContactStore") {
			funde = append(funde, datei+": CNContactStore")
		}
	}

	if len(funde) > 0 {
		t.Errorf("hier fließt ein Systemname oder eine Kontaktkarte ein — "+
			"nach ID-01.AK-04 nur mit Zustimmung und nie als Nachweis; "+
			"ID-04 bringt den Weg dafür:\n%s", strings.Join(funde, "\n"))
	}
}
