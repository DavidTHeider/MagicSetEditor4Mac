package werkzeuge_test

// **Jede Bindung, die ein Profil nennt, fragt nach der Rolle dort** —
// #1817.
//
// # Warum es diesen Wächter gibt
//
// Die Rolle eines Menschen setzt `query_only` durch — für das **gezeigte**
// Profil. Eine Bindung, die die Kennung eines anderen Profils oder eines
// Tokens vom Rufer nimmt, wirkt dort, wo diese Sperre nicht greift. Am
// 2026-09-23 entstand so aus dem eigenen Profil heraus ein schreibendes Token
// für das Profil eines anderen.
//
// # Was er hält
//
// **Jede exportierte Methode von `App`, die eine Profil- oder Tokenkennung
// annimmt** — als Parameter `…profilID`, `…tokenID` oder als Verbund mit einem
// Feld `ProfilID` —, erreicht eine der [zielfragen] oder steht mit Grund in
// [ohneZielfrage].
//
// # Was er nicht hält
//
// **Ob die Frage die richtige Rolle verlangt.** Das halten die Läufe in
// `internal/app/tokenrechte_test.go` und `leserschreibwege_test.go`, und die
// Matrix in `docs/AUTORISIERUNG.md` sagt, welche es sein soll. **Und eine
// Kennung unter einem anderen Namen** — ein Parameter `kennung`, der ein
// Profil meint, sähe er nicht.

import (
	"go/ast"
	"maps"
	"slices"
	"strings"
	"testing"
)

// zielfragen sind die Rufe, die nach der Rolle am genannten Profil fragen.
var zielfragen = []string{
	"App.rolleAm", "App.zumSchreibenAm", "App.rolleAmToken", "App.lesbarAm", "lesbarMit",
	"App.alsVerwalter", "App.ueberProfil", "App.profilVorab",
}

// ohneZielfrage sind die Bindungen, die ein Profil nennen und die Rolle dort
// bewusst nicht an dieser Stelle fragen — mit Grund.
var ohneZielfrage = map[string]string{
	"App.ProfilWechseln": "der Wechsel selbst: ihn schützt der Zutritt — der zweite Faktor des " +
		"Eigners oder eine Freigabe —, und danach gilt die Rolle im gezeigten Profil",
	"App.Zuruecksetzen": "die Rücksetzung fragt selbst, ob das Ziel dem Rufer gehört " +
		"(`Ruecksetzung.gebundenesZiel`), und die ganze Anwendung verlangt einen Nachweis",
	"App.Ruecksetzungsvorschau": "dieselbe Frage wie beim Zurücksetzen, und ein fremdes Ziel " +
		"wird nicht einmal gezählt",
}

// **Jede Bindung mit einem Zielprofil fragt nach der Rolle dort.**
func TestJedeBindungMitZielprofilFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	bestand := funktionenAus(t, "app")
	geprueft := 0

	for _, schluessel := range slices.Sorted(maps.Keys(bestand.funktionen)) {
		funktion := bestand.funktionen[schluessel]
		if !strings.HasPrefix(schluessel, "App.") || !funktion.Name.IsExported() ||
			!bestand.nenntEinProfil(funktion) {
			continue
		}

		geprueft++

		if ohneZielfrage[schluessel] == "" && !bestand.erreicht(schluessel, zielfragen, map[string]bool{}) {
			t.Errorf("%s nennt ein Profil und fragt nicht nach der Rolle dort — eine der "+
				"Zielfragen rufen oder mit Grund in ohneZielfrage eintragen", schluessel)
		}
	}

	// **Und er muss etwas gesehen haben**: Token, Vorgaben, Freigaben und
	// Profilpflege sind mehr als zehn.
	if geprueft < 10 {
		t.Errorf("nur %d Bindungen mit Zielprofil gesehen — das Muster greift daneben", geprueft)
	}

	for _, schluessel := range slices.Sorted(maps.Keys(ohneZielfrage)) {
		if bestand.funktionen[schluessel] == nil || len(ohneZielfrage[schluessel]) < 40 {
			t.Errorf("die Ausnahme %q gibt es nicht oder sie ist nicht begründet", schluessel)
		}
	}
}

// nenntEinProfil sagt, ob eine Bindung eine Profil- oder Tokenkennung annimmt.
func (b quellbestand) nenntEinProfil(funktion *ast.FuncDecl) bool {
	for _, parameter := range funktion.Type.Params.List {
		for _, name := range parameter.Names {
			klein := strings.ToLower(name.Name)
			if strings.HasSuffix(klein, "profilid") || strings.HasSuffix(klein, "tokenid") {
				return true
			}
		}

		if typ, istName := parameter.Type.(*ast.Ident); istName && b.mitProfil[typ.Name] {
			return true
		}
	}

	return false
}
