package werkzeuge_test

// **Jede Bindung, die einen Menschen nennt, fragt, ob sie über ihn schreiben
// darf** — ID-03 (#1702).
//
// # Was er hält
//
// **Jede exportierte Methode von `app.App` mit einem Parameter `nutzerID`**
// ruft `ueberNutzer` oder `nutzerVorab` — oder steht unten mit Grund als
// Ausnahme. Dazu die Bindungen, die ein **Profil** beschreiben: sie rufen
// `ueberProfil` oder `profilVorab`. Eine neue Bindung über einen Menschen
// erzwingt damit die Entscheidung, statt still offen zu stehen — so fehlte
// die Frage bis ID-03 an **allen** siebzehn Stellen, und die Matrix
// versprach sie seit 0078.
//
// # Was er nicht hält
//
// **Ob die Frage vor der ersten Wirkung steht.** Er sieht einen Ruf im Rumpf,
// nicht seine Stelle; dass eine Absage nichts hinterlässt, halten die Läufe
// `TestUeberEinenGeschuetztenNutzerSchreibtNurEr` und
// `TestOhneRolleWirdImFremdenProfilNichtsGeschrieben` an der Wirkung. Und
// nicht, ob die Regel selbst stimmt — die steht in
// `service.Nutzerverwaltung.UeberNutzer` und hat dort ihre Fälle.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ohneFrage sind die Bindungen mit `nutzerID`, die nicht fragen — mit Grund.
var ohneFrage = map[string]string{
	"Faktoren":                 "liest nur, welche Arten es gibt — wer zu jemandem wechseln will, muss wissen, wonach er gefragt wird",
	"Anbieterkonten":           "liest die Konten des Eigners des gezeigten Profils; die Nutzerseite zeigt sie jedem, der dort ist — steht im Bedrohungsmodell",
	"Nutzereinstellungen":      "liest; keine Einstellung trägt ein Geheimnis",
	"Tischbild":                "liest, ob es ein eigenes Tischbild gibt",
	"Fensterbild":              "liest, ob es ein eigenes Fensterbild gibt",
	"NutzerWechseln":           "fragt selbst: der Wechsel verlangt den Faktor des Ziels (0074)",
	"LokalAnmelden":            "fragt selbst: das Kennwort des Ziels und über den Nutzerwechsel dessen zweiten Faktor (ID-02, #1701)",
	"KennwortWiederherstellen": "fragt selbst: der Code vom Zettel des Ziels ist der Nachweis, und gesperrt stellt nur der Gesperrte wieder her (ID-02)",
	"Ruecksetzungsvorschau":    "liest nur, was eine Rücksetzung träfe",
	"Zuruecksetzen":            "prüft im Dienst, ob das Ziel dem Arbeitenden gehört (RuecksetzungFremdesZiel)",
	"PersonendatenSchreiben":   "holt sie über Personendaten, und das fragt",
	"FreigabeLoeschen":         "nennt den, dessen Freigabe fällt, und nicht den, über den geschrieben wird — gefragt wird, ob der Arbeitende das Profil verwaltet (alsVerwalter)",
}

// profilbeschreibend sind die Bindungen, die ein Profil ändern oder löschen.
var profilbeschreibend = []string{
	"ProfilAendern", "ProfilLoeschen",
	"ProfilbadgeWaehlen", "ProfilbadgeSetzen", "ProfilbadgeEntfernen",
}

func TestJedeBindungUeberEinenMenschenFragt(t *testing.T) {
	t.Parallel()

	dateien, fehler := filepath.Glob("../../internal/app/*.go")
	if fehler != nil {
		t.Fatalf("suchen: %v", fehler)
	}

	gesehen := map[string]bool{}

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		gelesen, lesefehler := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
		if lesefehler != nil {
			t.Fatalf("%s zerlegen: %v", datei, lesefehler)
		}

		for _, deklaration := range gelesen.Decls {
			funktion, ist := deklaration.(*ast.FuncDecl)
			if !ist || !vonApp(funktion) || !funktion.Name.IsExported() {
				continue
			}

			name := funktion.Name.Name

			switch {
			case slices.Contains(profilbeschreibend, name):
				gesehen[name] = true

				if !ruft(funktion, "ueberProfil", "profilVorab") {
					t.Errorf("%s (%s) beschreibt ein Profil und fragt nicht, ob sie darf",
						name, filepath.Base(datei))
				}

			case hatParameter(funktion, "nutzerID"):
				gesehen[name] = true

				if _, ausgenommen := ohneFrage[name]; ausgenommen {
					continue
				}

				if !ruft(funktion, "ueberNutzer", "nutzerVorab") {
					t.Errorf("%s (%s) nennt einen Menschen und fragt nicht, ob sie über ihn "+
						"schreiben darf — ruf `a.ueberNutzer` oder trag sie mit Grund in `ohneFrage` ein",
						name, filepath.Base(datei))
				}
			}
		}
	}

	// **Und keine Ausnahme ohne Bindung** — ein Eintrag, der niemanden mehr
	// trifft, hielte eine Tür offen, die es nicht mehr gibt.
	for name := range ohneFrage {
		if !gesehen[name] {
			t.Errorf("die Ausnahme %s trifft keine Bindung mehr", name)
		}
	}

	for _, name := range profilbeschreibend {
		if !gesehen[name] {
			t.Errorf("die profilbeschreibende Bindung %s gibt es nicht mehr", name)
		}
	}
}

// vonApp sagt, ob eine Funktion eine Methode von `*App` ist.
func vonApp(funktion *ast.FuncDecl) bool {
	if funktion.Recv == nil || len(funktion.Recv.List) != 1 {
		return false
	}

	zeiger, ist := funktion.Recv.List[0].Type.(*ast.StarExpr)
	if !ist {
		return false
	}

	name, benannt := zeiger.X.(*ast.Ident)

	return benannt && name.Name == "App"
}

// hatParameter sagt, ob die Funktion einen Parameter dieses Namens hat.
func hatParameter(funktion *ast.FuncDecl, gesucht string) bool {
	for _, feld := range funktion.Type.Params.List {
		for _, name := range feld.Names {
			if name.Name == gesucht {
				return true
			}
		}
	}

	return false
}

// ruft sagt, ob der Rumpf eine der genannten Methoden von `a` ruft.
func ruft(funktion *ast.FuncDecl, methoden ...string) bool {
	gefunden := false

	ast.Inspect(funktion.Body, func(knoten ast.Node) bool {
		aufruf, ist := knoten.(*ast.CallExpr)
		if !ist {
			return true
		}

		auswahl, gewaehlt := aufruf.Fun.(*ast.SelectorExpr)
		if gewaehlt && slices.Contains(methoden, auswahl.Sel.Name) {
			gefunden = true
		}

		return !gefunden
	})

	return gefunden
}
