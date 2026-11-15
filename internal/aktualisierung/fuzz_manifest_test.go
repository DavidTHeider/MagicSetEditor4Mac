package aktualisierung_test

// **Das Manifest einer Freigabe aus fremder Hand** — Werkstück O11 von
// Phase 5.2 (#1435).
//
// # Warum
//
// Das Manifest kommt **aus dem Netz** und entscheidet, ob dieser Bestand mit
// einer anderen Fassung geöffnet werden kann (O6,
// [0084](../../docs/entscheidungen/0084-das-kompatibilitaetsmanifest.md)). Es
// hängt an einer signierten Prüfsummenliste — aber die Signatur sagt nur, dass
// niemand es unterwegs geändert hat; **was darin steht**, bleibt fremd.
//
// Gefuzzt wird der ganze Weg: lesen **und** urteilen. Getrennt geprüft hieße,
// dass ein Manifest, das sich lesen lässt, ein Urteil erzwingen darf, das
// niemand geprüft hat.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse — auch auf dem Weg der
//     Ablehnung
//   - **ein Manifest und ein Fehler**, oder keins von beidem
//   - **ein `bereit` ohne Vertraeglichkeit**: nur ein vollständiges,
//     plattformgleiches Manifest mit bekannten Wanderungen darf vorbereiten —
//     die Zusage aus O6, und die einzige, deren Bruch etwas kaputtmacht
//   - **ein Hindernis ohne Art** — es sagte sonst nicht, was im Weg steht (§8)
//   - **ein Urteil außerhalb der drei** — `ungeklaert` ist kein Nein, und ein
//     viertes Wort verstünde keine Fläche
//
//	go test ./internal/aktualisierung -run Fuzz -fuzz FuzzManifestAus -fuzztime 2m

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// manifestsaat sind die Gestalten, mit denen der Fuzzer beginnt.
var manifestsaat = []string{
	`{}`,
	`{"fassung":"0.5.0"}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64"}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","artefakt":"abc","einstellungen":1,"protokoll":1}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","wanderungen":[]}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","wanderungen":[{"name":"0001_start.sql","summe":"aa"}]}`,
	`{"fassung":"0.5.0","plattform":"linux/amd64","wanderungen":[{"name":"0001_start.sql","summe":"aa"}]}`,
	`{"fassung":"","plattform":"darwin/arm64"}`,
	`{"fassung":"0.5.0","plattform":""}`,
	`{"fassung":123,"plattform":"darwin/arm64"}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","wanderungen":null}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","wanderungen":[{"name":"","summe":""}]}`,
	`{"fassung":"0.5.0","plattform":"darwin/arm64","einstellungen":1e309}`,
	`[]`,
	`null`,
	``,
}

// **Der Leser bricht nicht ab, und kein fremdes Manifest macht bereit.**
func FuzzManifestAus(f *testing.F) {
	for _, saat := range manifestsaat {
		f.Add([]byte(saat))
	}

	// verbucht ist der Bestand dieses Rechners: eine Wanderung, die er kennt.
	verbucht := map[string]string{"0001_start.sql": "aa"}

	f.Fuzz(func(t *testing.T, roh []byte) {
		manifest, fehler := aktualisierung.ManifestAus(roh)

		nochmal, zweiterFehler := aktualisierung.ManifestAus(roh)

		if (fehler == nil) != (zweiterFehler == nil) {
			t.Fatalf("dasselbe Manifest einmal so und einmal anders: %v gegen %v",
				fehler, zweiterFehler)
		}

		if fehler != nil {
			if !reflect.DeepEqual(manifest, domain.Fassungsmanifest{}) {
				t.Errorf("abgelehnt (%v) und trotzdem ein Manifest: %+v",
					fehler, manifest)
			}

			if fehler.Error() != zweiterFehler.Error() {
				t.Errorf("dasselbe Manifest gab zwei Meldungen: %q gegen %q",
					fehler, zweiterFehler)
			}

			return
		}

		if !reflect.DeepEqual(manifest, nochmal) {
			t.Errorf("dasselbe Manifest gab zwei verschiedene Ergebnisse: %+v gegen %+v",
				manifest, nochmal)
		}

		if manifest.Fassung == "" || manifest.Plattform == "" {
			t.Errorf("angenommen ohne Fassung oder Plattform: %+v", manifest)
		}

		urteil, hindernisse := domain.Vertraeglichkeit(manifest, verbucht, "darwin/arm64")

		switch urteil {
		case domain.UrteilVertraeglich, domain.UrteilUnvertraeglich, domain.UrteilUngeklaert:
		default:
			t.Errorf("ein viertes Urteil: %q", urteil)
		}

		for _, hindernis := range hindernisse {
			if hindernis.Art == "" {
				t.Errorf("ein Hindernis ohne Art: %+v", hindernis)
			}
		}

		befund := domain.Bereitschaft(urteil, hindernisse)

		// **Bereit heißt: verträglich und ohne ein einziges Hindernis.**
		if befund.Bereit &&
			(urteil != domain.UrteilVertraeglich || len(hindernisse) != 0) {
			t.Errorf("bereit trotz %q und %d Hindernissen: %+v",
				urteil, len(hindernisse), manifest)
		}

		// **Und ein bereites Manifest gehört zu dieser Plattform.**
		if befund.Bereit && manifest.Plattform != "darwin/arm64" {
			t.Errorf("bereit für eine fremde Plattform %q", manifest.Plattform)
		}
	})
}
