package app_test

// **Wer antwortet hier** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"runtime"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestMetaNenntFassungPlattformUndLaufzeit(t *testing.T) {
	t.Parallel()

	angaben := app.NeueApp("").Meta(t.Context())

	switch {
	case angaben.Fehler != nil:
		t.Fatalf("Meta gab einen Fehler: %+v", angaben.Fehler)
	case angaben.Fassung != app.Fassung:
		t.Errorf("Fassung = %q, erwartet %q", angaben.Fassung, app.Fassung)
	case angaben.Plattform != runtime.GOOS+"/"+runtime.GOARCH:
		t.Errorf("Plattform = %q", angaben.Plattform)
	case !strings.HasPrefix(angaben.Go, "go1."):
		t.Errorf("Go = %q", angaben.Go)
	}
}

// **Ohne Angabe gilt das Fenster** — dieselbe Regel wie beim Akteur, und sie
// steht dort begründet: Wails reicht einen nackten Kontext herein.
func TestMetaNenntDieFlaecheDesRufs(t *testing.T) {
	t.Parallel()

	if flaeche := app.NeueApp("").Meta(t.Context()).Flaeche; flaeche != "fenster" {
		t.Errorf("ohne Angabe = %q, erwartet fenster", flaeche)
	}

	ctx := app.MitDienstmodus(t.Context(), domain.ModusDienst)
	if flaeche := app.NeueApp("").Meta(ctx).Flaeche; flaeche != "dienst" {
		t.Errorf("am Dienst = %q", flaeche)
	}
}

// **Bau und Bindungsvertrag stehen an jeder Stelle gleich** — ENG-06.AK-04
// (#1731):
// in der Meta-Auskunft der drei Flächen und im Diagnosetext des Fensters.
// Eine Stelle, die eine eigene Fassung davon schriebe, liefe auseinander.
func TestMetaUndDiagnoseNennenDenselbenBau(t *testing.T) {
	anwendung := app.NeueApp("")
	angaben := anwendung.Meta(t.Context())

	if angaben.Bau != app.DiesenBau() || angaben.Bindungsvertrag != app.Vertragsabdruck() {
		t.Errorf("Meta nennt %+v und %q", angaben.Bau, angaben.Bindungsvertrag)
	}

	text := anwendung.Diagnose(t.Context()).Text

	for _, teil := range []string{angaben.Bau.Stand, angaben.Bindungsvertrag} {
		if !strings.Contains(text, teil) {
			t.Errorf("der Diagnosetext nennt %q nicht:\n%s", teil, text)
		}
	}
}
