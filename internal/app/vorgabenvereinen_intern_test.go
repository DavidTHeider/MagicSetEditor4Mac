package app

import (
	"strings"
	"testing"
)

// **Was sich nicht als oberstes Wörterbuch lesen lässt, wird nicht
// importiert** — #2424: ein halber Umzug ist schlimmer als keiner.
func TestUnlesbareVorgabenWerdenNichtVereint(t *testing.T) {
	t.Parallel()

	gut := []byte(leereVorgaben)

	for name, roh := range map[string]string{
		"kein XML":                      "<plist><dict><key>a</key>",
		"ein Wert ohne Schlüssel":       "<plist><dict><string>a</string></dict></plist>",
		"ein ungeschlossener Schlüssel": "<plist><dict><key>a</plist>",
		"ein ungeschlossener Wert":      "<plist><dict><key>a</key><dict><key>b</key></plist>",
	} {
		if _, err := vorgabenVereinen([]byte(roh), gut, "alt"); err == nil {
			t.Errorf("%s als alte Domäne wurde vereint", name)
		}

		if _, err := vorgabenVereinen(gut, []byte(roh), "alt"); err == nil {
			t.Errorf("%s als neue Domäne wurde vereint", name)
		}
	}
}

// **Die neue gewinnt, die Marke kommt einmal, und Sonderzeichen bleiben
// Zeichen.**
func TestDieNeueDomaeneGewinnt(t *testing.T) {
	t.Parallel()

	alt := `<plist><dict><key>a&amp;b</key><string>alt</string><key>x</key><integer>1</integer><key>` + umzugsmarke + `</key><string>alt</string></dict></plist>`
	neu := `<plist><dict><key>x</key><integer>2</integer></dict></plist>`

	vereint, err := vorgabenVereinen([]byte(alt), []byte(neu), "com.<alt>")
	if err != nil {
		t.Fatal(err)
	}

	text := string(vereint)
	if !strings.Contains(text, "<key>x</key>\n\t<integer>2</integer>") || strings.Contains(text, "<integer>1</integer>") ||
		!strings.Contains(text, "<key>a&amp;b</key>") || strings.Count(text, "<key>"+umzugsmarke+"</key>") != 1 ||
		!strings.Contains(text, "<string>com.&lt;alt&gt;</string>") {
		t.Errorf("vereint:\n%s", text)
	}
}
