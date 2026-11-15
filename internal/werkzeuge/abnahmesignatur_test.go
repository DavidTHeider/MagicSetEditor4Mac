package werkzeuge

import (
	"strings"
	"testing"
)

// **Das Abnahmebündel ist zuletzt versiegelt** — gefunden am 2026-10-11 bei
// I03 (#2261).
//
// `scripts/bauen.sh abnahme` legte `mse`, `mse-mcp` und `mse-rest` ins Bündel,
// nachdem es versiegelt war, und signierte nie neu: `codesign --verify` sagte
// „a sealed resource is missing or invalid", und Gatekeeper prüfte im Gast ein
// beschädigtes Bündel statt eines ad hoc signierten. Das Freigabebündel
// (`scripts/release.sh`) war gültig — die Proben maßen also ein anderes.
//
// **Was er nicht prüft**: dass die Signatur gelingt. Er liest das Skript wie
// `abnahmeoberflaeche_test.go`; geprüft wird sie am gebauten Bündel, das jede
// Hüllenprobe baut, mit `codesign --verify --deep --strict`.
func TestDasAbnahmebuendelIstZuletztVersiegelt(t *testing.T) {
	t.Parallel()

	skript := string(projektdatei(t, "scripts/bauen.sh"))

	werkzeuge := strings.Index(skript, `"$WURZEL/scripts/werkzeuge.sh" "$APP"`)
	if werkzeuge < 0 {
		t.Fatal("scripts/bauen.sh legt die Werkzeuge nicht mehr ins Bündel — dieser Wächter liest an der falschen Stelle")
	}

	// **Mit Hardened Runtime** seit SEC03 (#2356, Vermerk 0195) — dass sie
	// dasteht, hält `laufzeithaertung_test.go`; hier zählt die Stelle.
	siegel := strings.LastIndex(skript, "codesign --force --deep --options runtime --sign -")
	if siegel < werkzeuge {
		t.Errorf("scripts/bauen.sh versiegelt das Bündel nicht nach den Werkzeugen (Siegel an %d, Werkzeuge an %d)", siegel, werkzeuge)
	}
}
