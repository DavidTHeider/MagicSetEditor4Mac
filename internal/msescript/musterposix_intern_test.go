package msescript

import "testing"

// **Was der Übersetzer stehen lässt** — ein `[` in einer Klasse ohne `:`,
// ein Name ohne Schluss und ein unbekannter Name gehen unverändert an
// `regexp2`. Dazu die Fälle, die er ausschreibt.
func TestPosixKlassenAusgeschrieben(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ muster, erwartet string }{
		{`[[:space:]]`, `[\t\n\v\f\r ]`},
		{`[^[:lower:]x]`, `[^a-zx]`},
		{`[[:^alnum:]]`, `[^0-9A-Za-z]`},
		// Hinter `\[` öffnet `[` eine gewöhnliche Klasse aus `:`, `s` …
		{`\[[:space:]`, `\[[:space:]`},
		{`[a[b]`, `[a[b]`},
		{`[[:space]`, `[[:space]`},
		{`[[:fremd:]]`, `[[:fremd:]]`},
		{`[]a]`, `[]a]`},
		{`(?x) a []] b`, `(?x) a []] b`},
	} {
		if gesehen, _ := posixKlassenAusgeschrieben(fall.muster); gesehen != fall.erwartet { //nolint:errcheck // hier nur gültige
			t.Errorf("%q wird %q, erwartet %q", fall.muster, gesehen, fall.erwartet)
		}
	}
}
