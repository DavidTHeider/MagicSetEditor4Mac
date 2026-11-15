package werkzeuge_test

// **Die WebKit-Probe richtet ihre Ansicht ein wie das Fenster** —
// #1828.
//
// Sie ist nur so viel wert, wie ihr WKWebView dem des Fensters gleicht. Eine
// Vorliebe, die das Fenster setzt und die Probe nicht, misst ein anderes
// WebKit — der Fund, der sie nötig machte, hing an genau einer: Tab erreichte
// ohne `TabFocusesLinks` keinen Knopf.
//
// **Was er nicht hält**: dass jede Vorliebe hier steht. Er hält die Paare,
// die er kennt; eine neue Vorliebe im Fenster bekommt ihr Paar mit ihr.

import (
	"strings"
	"testing"
)

func TestDieWebkitprobeRichtetIhreAnsichtEinWieDasFenster(t *testing.T) {
	t.Parallel()

	fenster := lies(t, "internal/fenster/ereignisse.go")
	probe := lies(t, "scripts/webkitprobe/Motoren.swift")

	paare := []struct{ imFenster, inDerProbe string }{
		{"TabFocusesLinks: application.Enabled", "preferences.tabFocusesLinks = true"},
	}

	for _, paar := range paare {
		if strings.Contains(fenster, paar.imFenster) != strings.Contains(probe, paar.inDerProbe) {
			t.Errorf("das Fenster setzt %q, die Probe %q — nur eines von beiden steht da",
				paar.imFenster, paar.inDerProbe)
		}
	}
}
