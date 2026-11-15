package msescript

import (
	"fmt"
	"strings"
)

// posixKlassen sind die Klassen, die RE2 in eckigen Klammern kennt — **in
// seiner Fassung**, also ASCII (`regexp/syntax`, `posixGroup`).
//
// **Warum ausgeschrieben** (#2315): `regexp2`, die Rückfalllinie für
// Muster mit Rückblick oder `(?x)`, kennt keine POSIX-Klassen und liest
// `[[:space:]]` als Klasse aus `[`, `:`, `s` … und ein `]` dahinter. MSE2
// übersetzt mit `boost::regex`, das sie kennt (`src/util/regex.cpp`, Zeile
// 20, Fassung f9d9356 vom 2021-07-06). Ausgeschrieben wie bei RE2, damit
// dasselbe Muster auf beiden Linien dasselbe trifft.
//
// **Was das nicht hält**: dass `boost` dieselben Zeichen meint — dort richtet
// sich `[:alpha:]` nach der Zeichentabelle, hier ist es ASCII. Über das ganze
// Pack stehen nur `[:space:]` und `[:lower:]` (gezählt am 2026-10-14), und
// `[:space:]` ist in beiden gleich bis auf Leerzeichen jenseits von ASCII.
var posixKlassen = map[string]string{
	"alnum":  `0-9A-Za-z`,
	"alpha":  `A-Za-z`,
	"ascii":  `\x00-\x7F`,
	"blank":  `\t `,
	"cntrl":  `\x00-\x1F\x7F`,
	"digit":  `0-9`,
	"graph":  `!-~`,
	"lower":  `a-z`,
	"print":  ` -~`,
	"punct":  `!-/:-@\[-` + "`" + `{-~`,
	"space":  `\t\n\v\f\r `,
	"upper":  `A-Z`,
	"word":   `0-9A-Za-z_`,
	"xdigit": `0-9A-Fa-f`,
}

// posixKlassenAusgeschrieben ersetzt `[:name:]` in eckigen Klammern durch
// die Zeichen, die es meint. Eine verneinte Klasse (`[:^name:]`) geht nur
// als ganze Klammer — `[[:^alnum:]]` wird `[^0-9A-Za-z]`.
//
// **Mitten in einer Klasse wird sie abgewiesen** (Zweitblick über Phase 1
// von 0.9.0, Befund 12): .NET kennt keine verneinte Teilmenge, die in ASCII
// dasselbe meint — `\D` meint dort Unicode —, und stehen gelassen las
// `regexp2` die Schreibweise als Zeichen. Im Pack steht sie nicht.
func posixKlassenAusgeschrieben(muster string) (string, error) {
	for name, zeichen := range posixKlassen {
		muster = strings.ReplaceAll(muster, "[[:^"+name+":]]", "[^"+zeichen+"]")
	}

	var gebaut strings.Builder

	zeichen := []rune(muster)
	inKlasse := false

	for stelle := 0; stelle < len(zeichen); stelle++ {
		switch {
		case zeichen[stelle] == '\\' && stelle+1 < len(zeichen):
			gebaut.WriteRune(zeichen[stelle])
			stelle++
		case !inKlasse && zeichen[stelle] == '[':
			inKlasse = true
			stelle = klassenanfang(zeichen, stelle, &gebaut)

			continue
		case inKlasse && zeichen[stelle] == ']':
			inKlasse = false
		case inKlasse && zeichen[stelle] == '[':
			if name, ende, ist := posixname(zeichen, stelle); ist {
				gebaut.WriteString(posixKlassen[name])
				stelle = ende

				continue
			}

			if strings.HasPrefix(string(zeichen[stelle:]), "[:^") {
				return "", fmt.Errorf("%w: eine verneinte POSIX-Klasse mitten in einer Klasse "+
					"lässt sich auf der Rückfalllinie nicht übersetzen", ErrMuster)
			}
		}

		gebaut.WriteRune(zeichen[stelle])
	}

	return gebaut.String(), nil
}

// klassenanfang schreibt `[`, ein `^` und ein `]` gleich danach — das ist
// dort ein Zeichen und kein Ende — und gibt die letzte geschriebene Stelle.
func klassenanfang(zeichen []rune, stelle int, gebaut *strings.Builder) int {
	gebaut.WriteByte('[')

	if stelle+1 < len(zeichen) && zeichen[stelle+1] == '^' {
		stelle++
		gebaut.WriteByte('^')
	}

	if stelle+1 < len(zeichen) && zeichen[stelle+1] == ']' {
		stelle++
		gebaut.WriteByte(']')
	}

	return stelle
}

// posixname liest `[:name:]` ab dieser Stelle — mit dem Namen und der Stelle
// des schließenden `]`.
func posixname(zeichen []rune, stelle int) (name string, ende int, ist bool) {
	if stelle+1 >= len(zeichen) || zeichen[stelle+1] != ':' {
		return "", 0, false
	}

	rest := string(zeichen[stelle+2:])

	schluss := strings.Index(rest, ":]")
	if schluss < 0 {
		return "", 0, false
	}

	name = rest[:schluss]
	if _, bekannt := posixKlassen[name]; !bekannt {
		return "", 0, false
	}

	return name, stelle + 2 + len([]rune(name)) + 1, true
}
