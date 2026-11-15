//go:build fremdpakete

package app

import "testing"

// **MSE reinigt den ganzen Dateinamen, nicht den Kartennamen** — QA-03
// (#1656).
//
//	String filename = clean_filename(untag(ctx.eval(*filename_script)->toString()));
//
// `export_images`, `src/data/format/image.cpp`, Zeile 82, Fassung f9d9356 vom
// 2021-07-06 — und das Skript ist `{card.name}.png`. Der Schlusspunkt eines
// Namens steht damit **vor** `.png` und bleibt stehen. Hier wurde der Name für
// sich gereinigt und `.png` danach angehängt: `Phoebe, Head of S.N.E.A.K.`
// fand sein Referenzbild nie, in jedem Kanonlauf, als »kein Referenzbild«.
func TestDieReferenzdateiIstMSEsDateiname(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, erwartet string }{
		{"Phoebe, Head of S.N.E.A.K.", "Phoebe Head of S.N.E.A.K..png"},
		{"Ajani, Caller of the Pride", "Ajani Caller of the Pride.png"},
		{" .Vorn gereinigt", "Vorn gereinigt.png"},
		{"Æther Vial", "Æther Vial.png"},
		// Bleibt vom Namen nichts, bleibt `.png` — und dessen Punkt fällt als
		// führendes Zeichen weg.
		{"???", "png"},
	} {
		if got := referenzdatei(fall.name); got != fall.erwartet {
			t.Errorf("referenzdatei(%q) = %q, erwartet %q", fall.name, got, fall.erwartet)
		}
	}
}
