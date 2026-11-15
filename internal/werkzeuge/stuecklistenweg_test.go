package werkzeuge

// **Eine Stückliste entsteht auf einem Weg** — Zweitblick über Phase 5,
// Befund 5 (SEC04, #2357). `make sbom` rief `cyclonedx-gomod` ohne `-json`
// und schrieb XML unter `stueckliste-go.json`, die Fehlerklasse, die
// `scripts/releasestueckliste.sh` für die Freigabe behoben hatte.
//
// # Was er prüft
//
// Das `Makefile` ruft weder `cyclonedx-gomod` noch `npm sbom` selbst, sondern
// den Weg der Freigabe (`scripts/stueckliste.sh` über
// `stuecklisten_schreiben`). Was dieser Weg prüft, hält
// `scripts/releasestueckliste.test.sh`.

import (
	"strings"
	"testing"
)

func TestDieStuecklisteEntstehtAufEinemWeg(t *testing.T) {
	t.Parallel()

	makefile := string(projektdatei(t, "Makefile"))

	for _, verboten := range []string{"cyclonedx-gomod", "npm sbom"} {
		if strings.Contains(makefile, verboten) {
			t.Errorf("das Makefile ruft %s selbst — die Stückliste entsteht über scripts/stueckliste.sh", verboten)
		}
	}

	if !strings.Contains(makefile, "scripts/stueckliste.sh") {
		t.Error("das Makefile nimmt den Weg der Freigabe nicht (scripts/stueckliste.sh)")
	}

	if !strings.Contains(string(projektdatei(t, "scripts/stueckliste.sh")), "stuecklisten_schreiben") {
		t.Error("scripts/stueckliste.sh ruft stuecklisten_schreiben nicht")
	}
}
