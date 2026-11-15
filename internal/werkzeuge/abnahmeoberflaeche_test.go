package werkzeuge

import (
	"regexp"
	"testing"
)

// **Ein Abnahmebündel trägt die Oberfläche, die im Baum steht** — gefunden
// am 2026-10-02 bei der Messung nachher zu UXR-01
// (#1909).
//
// `scripts/bauen.sh abnahme` baute die Oberfläche nur, wenn es noch kein
// `frontend/dist` gab. Die Hüllenprobe maß danach Zahl für Zahl die
// Oberfläche von vor dem Umbau — und sah aus wie ein Umbau ohne Wirkung.
//
// **Was er nicht prüft**: dass der Bau gelingt. Er liest das Skript, wie
// `lokalnetzzweck_test.go` es tut; gebaut wird in der Hüllenprobe.
func TestDasAbnahmebuendelBautDieOberflaecheNeu(t *testing.T) {
	t.Parallel()

	skript := string(projektdatei(t, "scripts/bauen.sh"))
	zweig := regexp.MustCompile(`(?s)elif \[ "\$ART" = abnahme \]; then\n(.*?)\nelif`)

	treffer := zweig.FindStringSubmatch(skript)
	if treffer == nil {
		t.Fatal("scripts/bauen.sh hat keinen eigenen Zweig für die Abnahme vor dem Bau der Oberfläche")
	}

	if !regexp.MustCompile(`npm run build`).MatchString(treffer[1]) {
		t.Errorf("der Abnahmezweig baut die Oberfläche nicht:\n%s", treffer[1])
	}
}
