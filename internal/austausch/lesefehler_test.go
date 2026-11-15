package austausch_test

// **Die Wege, auf denen ein Umschlag nicht gelesen wird** — Werkstück N4 von
// Phase 5.4 (#1457).
//
// **Jeder einzeln**, und die Reihenfolge ist die Zusage: wer beim Manifest
// scheitert, hat noch kein Objekt auf der Platte.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

func TestDieWegeAufDenenNichtGelesenWird(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	echt := austausch.Manifest{
		Fassung: 1, Urheber: string(ich.Kennung()),
		Objekte: []austausch.Objekt{{
			Art: austausch.ArtKarte, Name: "a.json",
			Summe: eineSumme, Groesse: 1,
		}},
	}

	roh, fehler := json.Marshal(echt)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	for name, eintraege := range map[string]map[string][]byte{
		// **Ohne Manifest ist es kein Umschlag.**
		"ohne Manifest": {
			"austausch.sig": ich.Signiere(roh),
			"a.json":        []byte("x"),
		},

		// **Ohne Unterschrift ebenso** — ein Umschlag ohne sie sagt nicht,
		// von wem er kommt.
		"ohne Unterschrift": {
			"austausch.json": roh,
			"a.json":         []byte("x"),
		},

		// **Ein Manifest, das kein JSON ist.**
		"krummes Manifest": {
			"austausch.json": []byte("{ das ist kein JSON"),
			"austausch.sig":  ich.Signiere([]byte("{ das ist kein JSON")),
			"a.json":         []byte("x"),
		},

		// **Ein Manifest, das ein Objekt nennt, das nicht im Umschlag
		// liegt.** Der Fall tritt auf, wenn jemand einen Eintrag
		// herausnimmt und das Manifest stehen lässt.
		"Objekt fehlt": {
			"austausch.json": roh,
			"austausch.sig":  ich.Signiere(roh),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pfad := baueUmschlag(t, eintraege)

			ziel := t.TempDir()

			if _, lesefehler := austausch.Lies(pfad, ziel); lesefehler == nil {
				t.Fatal("der Umschlag kam durch")
			}

			// **Und nichts liegt im Inhaltsverzeichnis** — die Zusage der
			// Reihenfolge: wer vorher scheitert, packt nichts aus.
			// **Ein Fehler hier heißt: das Verzeichnis gibt es nicht** —
			// und genau das ist die Zusage. Deshalb wird er nicht geprüft,
			// sondern die Zahl.
			liegend, _ := os.ReadDir( //nolint:errcheck // siehe darüber
				filepath.Join(ziel, "inhalt"))
			if len(liegend) != 0 {
				t.Errorf("es liegen %d Objekte da", len(eintraege))
			}
		})
	}
}

// **Eine Datei, die gar kein Archiv ist.**
func TestEineFremdeDateiIstKeinUmschlag(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "fremd"+austausch.Endung)
	if fehler := os.WriteFile(pfad, []byte("kein ZIP"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if _, fehler := austausch.Lies(pfad, t.TempDir()); fehler == nil {
		t.Error("eine fremde Datei kam durch")
	}
}

// **Ein Objekt, das beide Manifeste nennen und das im Umschlag fehlt.**
//
// **Der Unterschied zum Fall „Objekt fehlt" darüber**: dort kennt auch das
// äußere Manifest es nicht, und `sicherung` reicht es durch. Hier steht es in
// beiden — und dann muss das Auspacken scheitern, nicht erst die Prüfsumme.
func TestEinObjektDasBeideManifesteNennenUndDasFehlt(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	manifest := austausch.Manifest{
		Fassung: 1, Urheber: string(ich.Kennung()),
		Objekte: []austausch.Objekt{{
			Art: austausch.ArtKarte, Name: "fehlt.json",
			Summe: eineSumme, Groesse: 1,
		}},
	}

	roh, fehler := json.Marshal(manifest)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	pfad := baueUmschlagMitLuecke(t, map[string][]byte{
		"austausch.json": roh,
		"austausch.sig":  ich.Signiere(roh),
	}, "fehlt.json")

	if _, lesefehler := austausch.Lies(pfad, t.TempDir()); lesefehler == nil {
		t.Error("ein Umschlag mit fehlendem Objekt kam durch")
	}
}
