// **Welcher Baum zu einem Paket gehört** —
// #1190.
//
// `internal/bezug` holt ein ganzes Repositorium und legt es als
// `<wurzel>/<name>/` ab; die Pakete darin stehen tiefer. Gesperrt wird gegen
// den **Tausch des Baums**, und der betrifft das erste Glied.
package service

import (
	"path/filepath"
	"testing"
)

func TestDerBezugsbaumIstDasErsteGliedUnterDerWurzel(t *testing.T) {
	t.Parallel()

	wurzel := filepath.FromSlash("/p/pakete")

	for _, fall := range []struct {
		name     string
		ort      string
		erwartet string
	}{
		{
			name:     "ein Paket im Bezugsbaum",
			ort:      filepath.FromSlash("/p/pakete/M15-Magic-Pack/data/magic-m15.mse-style"),
			erwartet: filepath.FromSlash("/p/pakete/M15-Magic-Pack"),
		},
		{
			// **Es kam nicht über einen Bezug** — und die Sperre trifft dann
			// genau es. Das ist richtig: `Beziehen` könnte es unter demselben
			// Namen ersetzen.
			name:     "ein Paket unmittelbar unter der Wurzel",
			ort:      filepath.FromSlash("/p/pakete/probe.mse-style"),
			erwartet: filepath.FromSlash("/p/pakete/probe.mse-style"),
		},
		{
			// **Außerhalb der Wurzel gibt es keinen Baum** — gesperrt wird der
			// Ort selbst. Der Weg dorthin ist von `Paketwerkstatt.ort`
			// verstellt; hier steht, was dann gälte.
			name:     "ein Ort außerhalb der Wurzel",
			ort:      filepath.FromSlash("/anderswo/probe.mse-style"),
			erwartet: filepath.FromSlash("/anderswo/probe.mse-style"),
		},
		{
			name:     "die Wurzel selbst",
			ort:      wurzel,
			erwartet: wurzel,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if baum := bezugsbaum(wurzel, fall.ort); baum != fall.erwartet {
				t.Errorf("Baum %q, erwartet %q", baum, fall.erwartet)
			}
		})
	}
}

// **Ein Ort, der sich nicht auf die Wurzel beziehen lässt, ist sein eigener
// Baum.**
//
// Auf allen Plattformen dieses Projekts gibt `filepath.Rel` genau dann einen
// Fehler, wenn eine Seite absolut und die andere relativ ist — der Fall ist
// von `Paketwerkstatt.ort` verstellt, und dieser Lauf sagt, was dann gälte.
func TestEinUnbeziehbarerOrtIstSeinEigenerBaum(t *testing.T) {
	t.Parallel()

	if baum := bezugsbaum("/p/pakete", "relativ/probe.mse-style"); baum != "relativ/probe.mse-style" {
		t.Errorf("Baum %q, erwartet den Ort selbst", baum)
	}
}

// **Und eine Sperrdatei, die sich nicht anlegen lässt, wird gemeldet.**
func TestEineUnanlegbareSperreWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, fehler := alleinAmBaum(filepath.Join(t.TempDir(), "gibtsnicht", "P")); fehler == nil {
		t.Error("die Sperre wurde genommen, obwohl das Verzeichnis fehlt")
	}
}
