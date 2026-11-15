package app

// **Die Außerbetriebnahme nennt jeden Ort, den diese Anwendung anlegt** —
// OPS03 (#2349),
// Befund A-10: `ADMINISTRATION.md` nannte Programm, Anwendungsordner und
// Schlüsselbund, und nichts von dem, was *Werkzeuge ohne Fenster* anlegt.
//
// # Was er hält
//
// Jedes Ziel der Schnittstellen mit Datei, Gedächtnis und Begleitordner —
// **aus denselben Zielen, die einrichten**, nicht aus einer Abschrift —, die
// Webordner und die Vorgaben unter der neuen und jeder alten Kennung und den
// Zwischenspeicher.
//
// # Was er nicht hält
//
// Dass die Befehle daneben so wirken, wie sie dastehen; und den
// Schlüsselbund — den hält `internal/werkzeuge/schluesselbunddienste_test.go`.

import (
	"context"
	"os"
	"strings"
	"testing"
)

// ausserbetriebnahme gibt den Abschnitt »Außerbetriebnahme« der
// Administration.
func ausserbetriebnahme(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile("../../docs/ADMINISTRATION.md")
	if err != nil {
		t.Fatal(err)
	}

	_, abschnitt, gefunden := strings.Cut(string(roh), "\n## Außerbetriebnahme\n")
	if !gefunden {
		t.Fatal("ADMINISTRATION.md hat keinen Abschnitt »Außerbetriebnahme«")
	}

	abschnitt, _, _ = strings.Cut(abschnitt, "\n## ")

	return abschnitt
}

// orteDerAnwendung sind die Orte, die die Außerbetriebnahme nennen muss —
// unter `~` geschrieben.
func orteDerAnwendung(t *testing.T) []string {
	t.Helper()

	const heim = "/HEIM"

	anwendung := NeueApp("").MitSchnittstellenorten(
		func() (string, error) { return heim, nil },
		func() (string, error) { return "/Programme/MagicSetEditor4Mac.app/Contents/MacOS", nil },
		func(context.Context) string { return "/usr/bin:/bin" },
	)

	ziele, err := anwendung.schnittstellenziele(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	unterHeim := func(pfad string) string { return "~" + strings.TrimPrefix(pfad, heim) }

	var orte []string

	for _, ziel := range ziele {
		for _, ort := range []string{ziel.Datei, ziel.Gedaechtnis, ziel.Begleitordner} {
			if ort != "" {
				orte = append(orte, unterHeim(ort))
			}
		}

		// **Und die Dienste unter einem früheren Etikett** (Zweitblick über
		// Phase 4, Befund 7).
		for _, alt := range ziel.Altlasten {
			orte = append(orte, unterHeim(alt.Datei))
		}
	}

	for _, kennung := range append([]string{Kennung}, altkennungen[0]) {
		orte = append(orte, "~/Library/Preferences/"+kennung+".plist")

		for _, ordner := range webordner {
			orte = append(orte, "~/Library/"+ordner+"/"+kennung+"/")
		}
	}

	// **Was der Umzug der Kennung selbst ablegt** (Befund 13) — aus seinen
	// Definitionen gelesen, nicht abgeschrieben: der Zwischenspeicher der
	// Vorgaben und die beiseitegelegten Webordner.
	orte = append(orte, unterHeim(vorgabenablage(heim))+"/")

	for _, ordner := range webordner {
		orte = append(orte, "~/Library/"+ordner+"/"+Kennung+beiseitevorsatz,
			// **Und der Zwischenort eines abgebrochenen Umzugs** (Nachlauf
			// über Phase 4, Befund 11).
			"~/Library/"+ordner+"/"+Kennung+zwischenendung+"/")
	}

	return append(orte, "~/Library/Caches/MagicSetEditor4Mac/", "~/Library/Application Support/MagicSetEditor4Mac/",
		"~/Library/LaunchAgents/"+dienstlabel+".plist")
}

func TestDieAusserbetriebnahmeNenntJedenOrt(t *testing.T) {
	t.Parallel()

	abschnitt := ausserbetriebnahme(t)

	for _, ort := range orteDerAnwendung(t) {
		// **Ein Ordner darf ohne Schrägstrich am Ende stehen** — in der Tafel
		// steht `~/.local/bin/mse`, nicht der Ordner allein.
		if !strings.Contains(abschnitt, strings.TrimSuffix(ort, "/")) {
			t.Errorf("die Außerbetriebnahme nennt %s nicht", ort)
		}
	}
}
