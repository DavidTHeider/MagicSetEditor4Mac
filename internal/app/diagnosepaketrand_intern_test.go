//go:build darwin

// **Die zwei Ränder des Diagnosepakets** — Werkstück O10 von Phase 5.2
// (#1434).
//
// **Von innen**, weil beide Wege von außen nicht erreichbar sind: ein Teil,
// den die Aufzählung nicht kennt, und ein Bestand, dessen Ort sich nicht
// bestimmen lässt. Beide sind keine erfundenen Fälle — der erste entsteht,
// sobald jemand `domain.Diagnoseteile()` erweitert und den Zweig vergisst,
// der zweite auf einem Rechner ohne Zuhause.
package app

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinUnbekannterTeilFaelltAufStattLeerZuBleiben(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	texte := DialogtexteFuer("de")

	inhalt, grund := anwendung.paketteil(t.Context(), domain.Diagnoseteil("gibtsnicht"), texte)

	switch {
	case inhalt != "":
		t.Errorf("der unbekannte Teil trug Inhalt: %q", inhalt)
	case grund != texte.DiagnoseTeilUnbekannt:
		t.Errorf("Grund = %q", grund)
	}
}

// **Ohne Zuhause gibt es keinen Bestandsort — und die Diagnose sagt das.**
// Ein Weg, der hier abbräche, ließe den Fall unberichtet, für den es ihn gibt
// (§8, wie bei [App.Diagnose]).
func TestOhneZuhauseNenntDerBestandsteilDenGrund(t *testing.T) {
	t.Setenv("HOME", "")

	anwendung := NeueApp("")
	texte := DialogtexteFuer("de")

	if teil := anwendung.bestandsteil(t.Context(), texte); !strings.Contains(
		teil, texte.DiagnoseNichtLesbar) {
		t.Errorf("der Teil nennt den Grund nicht:\n%s", teil)
	}
}
