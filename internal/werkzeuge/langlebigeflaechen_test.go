package werkzeuge

// **Die langlebigen Flächen tauschen ihren Bestand nicht** — G09
// (#2273).
//
// `mse-rest` und `mse-mcp` halten eine Verbindung über ihre ganze Laufzeit,
// und jede ihrer Nähte hat sie sich beim Aufbau geholt. Ein Tausch —
// Zurückspielen, Verschlüsselung umschalten, Neuanfang — schlösse sie, und
// jede Anfrage danach träfe eine geschlossene Verbindung. **Bis G09 tat
// `POST /ruecksetzung/anwendung` genau das**, über
// `app.RuecksetzungsAnwendungUeber(…).Zuruecksetzen`, und kein Lauf sah es:
// `internal/rest/bestandssicherung.go` begründete das Nein fürs
// Zurückspielen, und der Neuanfang lief daneben durch.
//
// # Was er hält
//
// Im Quelltext der beiden Wurzeln und ihrer Nähte (ohne Tests) **keinen Ruf
// eines Tauschwegs**: kein `Zurueckspielen`, kein `VerschluesselungUmschalten`,
// kein `Neuanfang` — und kein `Zuruecksetzen` an einer Anwendung aus
// `RuecksetzungsAnwendungUeber`, denn die trägt den Neuanfang. Die Rücksetzung
// des Dienstes baut `RuecksetzungsAnwendungOhneTausch`.
//
// # Was er nicht hält
//
//   - **Einen Tauschweg, der hier nicht genannt ist** — wer einen neuen baut,
//     trägt ihn in [tauschwege] ein.
//   - **Einen Tausch durch ein anderes Programm**, während ein Dienst läuft —
//     den weist `AlleinAmBestand` ab, und das hält
//     `TestGeschaltetWirdNichtWaehrendEinAndererDenBestandHaelt`.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// langlebigeFlaechen sind die Verzeichnisse, deren Prozess seinen Bestand
// über seine ganze Laufzeit hält.
var langlebigeFlaechen = []string{"cmd/mse-rest", "cmd/mse-mcp", "internal/rest", "internal/mcp", "internal/mcpnaehte"}

// tauschwege sind die Methoden, die den Bestand austauschen.
var tauschwege = map[string]bool{
	"Zurueckspielen":             true,
	"SicherungZurueckspielen":    true,
	"VerschluesselungUmschalten": true,
	"Neuanfang":                  true,
}

// tauschrufeIn nennt jeden Ruf eines Tauschwegs in einer Datei.
func tauschrufeIn(dateien *token.FileSet, datei *ast.File) []string {
	var funde []string

	ast.Inspect(datei, func(knoten ast.Node) bool {
		ruf, istRuf := knoten.(*ast.CallExpr)
		if !istRuf {
			return true
		}

		auswahl, istAuswahl := ruf.Fun.(*ast.SelectorExpr)
		if !istAuswahl {
			return true
		}

		stelle := dateien.Position(ruf.Pos())

		if tauschwege[auswahl.Sel.Name] {
			funde = append(funde, stelle.String()+": "+auswahl.Sel.Name)
		}

		// **Zurücksetzen an einer Anwendung, die den Neuanfang trägt.**
		if auswahl.Sel.Name == "Zuruecksetzen" {
			if innen, istInnen := auswahl.X.(*ast.CallExpr); istInnen {
				if bau, istBau := innen.Fun.(*ast.SelectorExpr); istBau &&
					bau.Sel.Name == "RuecksetzungsAnwendungUeber" {
					funde = append(funde, stelle.String()+": RuecksetzungsAnwendungUeber(…).Zuruecksetzen")
				}
			}
		}

		return true
	})

	return funde
}

func TestDieLanglebigenFlaechenTauschenIhrenBestandNicht(t *testing.T) {
	t.Parallel()

	dateien := token.NewFileSet()
	gelesen := 0

	for _, ordner := range langlebigeFlaechen {
		eintraege, fehler := os.ReadDir(filepath.Join("..", "..", ordner))
		if fehler != nil {
			t.Fatalf("%s lesen: %v", ordner, fehler)
		}

		for _, eintrag := range eintraege {
			name := eintrag.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			pfad := filepath.Join("..", "..", ordner, name)

			datei, lesefehler := parser.ParseFile(dateien, pfad, nil, parser.SkipObjectResolution)
			if lesefehler != nil {
				t.Fatalf("%s: %v", pfad, lesefehler)
			}

			gelesen++

			for _, fund := range tauschrufeIn(dateien, datei) {
				t.Errorf("%s — ein langlebiger Dienst tauscht seinen Bestand nicht; "+
					"die Rücksetzung baut `app.RuecksetzungsAnwendungOhneTausch`", fund)
			}
		}
	}

	// **Die Gegenprobe zum Nein**: er hat etwas gelesen.
	if gelesen < 10 {
		t.Fatalf("nur %d Dateien gelesen — der Wächter greift daneben", gelesen)
	}
}

// **Und er erkennt, was er sucht** — an einem Stück Quelltext, das genau den
// Fehler vor G09 trägt.
func TestDerTauschwaechterErkenntDenNeuanfangVorG09(t *testing.T) {
	t.Parallel()

	const quelle = `package main

func (r auskunft) Zuruecksetzen() {
	_ = app.RuecksetzungsAnwendungUeber(r.db, r.ort).Zuruecksetzen(ctx, "anwendung")
	_ = app.RuecksetzungsAnwendungOhneTausch(r.db).Zuruecksetzen(ctx, "layout")
	_ = naht.Weg().Neuanfang(ctx)
}
`

	dateien := token.NewFileSet()

	datei, fehler := parser.ParseFile(dateien, "probe.go", quelle, parser.SkipObjectResolution)
	if fehler != nil {
		t.Fatalf("die Probe zerteilen: %v", fehler)
	}

	funde := tauschrufeIn(dateien, datei)
	if len(funde) != 2 {
		t.Errorf("gefunden %v — erwartet der alte Weg und der Neuanfang, nicht der neue Weg", funde)
	}
}
