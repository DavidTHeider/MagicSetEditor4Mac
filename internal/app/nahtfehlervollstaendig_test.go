// **Jeder Fehlercode findet seinen Weg zurück** — Werkstück 17 von Phase 4.12.
//
// # Der Befund, für den es diese Datei gibt
//
// **`AlsFehler` bildete nur drei von acht Codes ab.** Alles andere fiel auf
// einen nackten Fehler zurück, und damit verloren `ErrErlaubnis`,
// `ErrZweiterFaktor`, `ErrBelegt`, `ErrNurLesend` und `ErrSchluesselFehlt`
// beim Übergang nach `cmd/mse`, MCP und Dienst **ihre Art**.
//
// **Gemessen am 2026-09-06**: `PUT /profile/2/freigaben/1` gab
// `500 {"code":"unbekannt","meldung":"der Aufruf ist gescheitert"}`, obwohl
// dieselbe Absage an der Kommandozeile „dafür braucht es mindestens die Rolle
// verwalter" lautete. `internal/rest/antwort.go` bildet `ErrErlaubnis` seit
// Werkstück 11 auf 403 ab — die Art kam dort nur nie an.
//
// **Das ist §3 in Reinform**: zwei grüne Seiten einer Naht sind kein Beleg für
// die Naht. `Uebersetze` hatte seinen Lauf, `antwort.go` hatte seinen, und
// dazwischen fiel die Art heraus.
package app

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Code, den [Uebersetze] setzen kann, kommt als seine Art zurück** —
// über **beide** Wege zurück: die Methode [Fehler.AlsFehler] und die Funktion
// [AlsFehler], die Kommandozeile, MCP und Dienst an vier Stellen rufen.
//
// # Warum die Codes aus dem Quelltext kommen
//
// **Weil die Tabelle hier selbst die Falle war.** Sie führte sieben Codes von
// zehn, und `CodeVoll` kam mit #1812 dazu, ohne dass jemand sie anfasste —
// die Datei hieß „vollständig" und war es nicht (Zweitblick über Phase 6.1,
// Befund 5). Jetzt liest der Lauf jede Konstante vom Typ [Fehlercode] aus
// `fehlercodes.go`, und eine ohne erwartete Art macht ihn rot.
func TestJederFehlercodeFindetSeineArtZurueck(t *testing.T) {
	t.Parallel()

	erwartet := map[Fehlercode]error{
		CodeNichtGefunden:   domain.ErrNotFound,
		CodeKonflikt:        domain.ErrConflict,
		CodeValidierung:     domain.ErrValidation,
		CodeBelegt:          domain.ErrBelegt,
		CodeErlaubnis:       domain.ErrErlaubnis,
		CodeSchluesselFehlt: domain.ErrSchluesselFehlt,
		CodeNurLesend:       domain.ErrNurLesend,
		CodeVoll:            domain.ErrVoll,
		CodeAbgebrochen:     context.Canceled,
		CodeAbgeschaltet:    domain.ErrAbgeschaltet,
		CodeNetz:            domain.ErrGegenstelle,
		CodeDrossel:         domain.ErrKontingent,
	}

	for _, code := range fehlercodesImQuelltext(t) {
		if code == CodeUnbekannt {
			continue
		}

		art, bekannt := erwartet[code]
		if !bekannt {
			t.Errorf("%q hat hier keine erwartete Art — trag sie ein, und in `arten`", code)

			continue
		}

		t.Run(string(code), func(t *testing.T) {
			t.Parallel()

			for weg, zurueck := range map[string]error{
				"Methode":  (&Fehler{Code: code, Meldung: "der Satz"}).AlsFehler(),
				"Funktion": AlsFehler(&Fehler{Code: code, Meldung: "der Satz"}),
			} {
				if !errors.Is(zurueck, art) {
					t.Errorf("%s: %q kam als %v zurück, erwartet war %v", weg, code, zurueck, art)
				}

				// **Und der Satz bleibt der Satz.** Wer ihn neu baute, hätte
				// das Klassenpräfix zweimal darin.
				if zurueck.Error() != "der Satz" {
					t.Errorf("%s: der Satz lautet %q", weg, zurueck.Error())
				}
			}
		})
	}
}

// fehlercodesImQuelltext liest jede Konstante vom Typ [Fehlercode].
//
// **Aus `fehlercodes.go` und nicht aus einer Liste**, und er bricht laut, wenn er
// weniger als die zehn findet, die es am 2026-09-23 gab — ein Ausdruck, der
// ins Leere griffe, gäbe sonst einen grünen Lauf über nichts.
func fehlercodesImQuelltext(t *testing.T) []Fehlercode {
	t.Helper()

	baum, fehler := parser.ParseFile(token.NewFileSet(), "fehlercodes.go", nil, 0)
	if fehler != nil {
		t.Fatalf("fehlercodes.go lesen: %v", fehler)
	}

	var codes []Fehlercode

	for _, erklaerung := range baum.Decls {
		gruppe, istGruppe := erklaerung.(*ast.GenDecl)
		if !istGruppe || gruppe.Tok != token.CONST {
			continue
		}

		for _, angabe := range gruppe.Specs {
			wert, istWert := angabe.(*ast.ValueSpec)
			if !istWert || wert.Type == nil || fmt.Sprint(wert.Type) != "Fehlercode" {
				continue
			}

			for _, zahl := range wert.Values {
				text, istText := zahl.(*ast.BasicLit)
				if istText && text.Kind == token.STRING {
					codes = append(codes, Fehlercode(strings.Trim(text.Value, `"`)))
				}
			}
		}
	}

	if len(codes) < 10 {
		t.Fatalf("nur %d Fehlercodes in fehlercodes.go gefunden: %v", len(codes), codes)
	}

	return codes
}

// **Und [CodeUnbekannt] bleibt ohne Art.**
//
// **Er heißt genau das**: wir wissen nicht, was es war. Ihn auf
// `ErrValidation` abzubilden hieße zu behaupten, der Aufrufer sei schuld.
func TestEinUnbekannterCodeBekommtKeineArt(t *testing.T) {
	t.Parallel()

	zurueck := (&Fehler{Code: CodeUnbekannt, Meldung: "etwas ging schief"}).AlsFehler()

	for _, art := range []error{
		domain.ErrNotFound, domain.ErrConflict, domain.ErrValidation,
		domain.ErrBelegt, domain.ErrErlaubnis, domain.ErrNurLesend,
	} {
		if errors.Is(zurueck, art) {
			t.Errorf("der unbekannte Code kam als %v zurück", art)
		}
	}
}
