package main

// **Zurück geht es auch über einem neueren Bestand** — M02
// (#2346),
// der zweite Befund aus dem Abnahmelauf von M01 (`aufstieg09`): die
// Kommandozeile von 0.8.0 öffnete und wanderte den Bestand, bevor sie
// zurückspielte, und scheiterte an der Wanderung, die sie nicht kannte. Die
// heutige war genauso gebaut — für den Rückschritt von einer Fassung nach
// dieser wäre es derselbe Befund. Gewandert wird jetzt erst der
// zurückgespielte Bestand.

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestZurueckGehtAuchUeberEinemNeuerenBestand(t *testing.T) {
	ort := frischerBestand(t)

	if code, _, fehler := fahreAn(t, ort, "sets"); code != CodeGut {
		t.Fatalf("anlegen: %d %s", code, fehler)
	}

	sicherung := filepath.Join(t.TempDir(), "vorher.msesicherung")
	if code, _, fehler := fahreAn(t, ort, "bestand-sichern", sicherung); code != CodeGut {
		t.Fatalf("sichern: %d %s", code, fehler)
	}

	// **Eine neuere Fassung wandert** — mit der nächsten freien Nummer.
	db, err := storage.Oeffnen(t.Context(), ort)
	if err != nil {
		t.Fatal(err)
	}

	verbucht, _ := storage.AngewendeteMigrationen(t.Context(), db) //nolint:errcheck // gezählt
	name := fmt.Sprintf("%04d_kuenftig.sql", len(verbucht)+1)

	if fehler := storage.MigriereMit(t.Context(), db,
		fstest.MapFS{"migrationen/" + name: {Data: []byte("CREATE TABLE kuenftig (wert INTEGER) STRICT;")}}, nil); fehler != nil {
		t.Fatal(fehler)
	}

	_ = db.Close() //nolint:errcheck // gewandert

	if code, _, fehler := fahreAn(t, ort, "sets"); code == CodeGut || !strings.Contains(fehler, "neueren Fassung") {
		t.Fatalf("die heutige Fassung öffnete den neueren Bestand: %d %s", code, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "--json", "bestand-schauen", sicherung)
	if code != CodeGut {
		t.Fatalf("über dem neueren Bestand lässt sich die Sicherung nicht ansehen: %d %s", code, fehler)
	}

	var schau struct {
		AngelegtAm string `json:"angelegtAm"`
	}

	if err := json.Unmarshal([]byte(aus), &schau); err != nil {
		t.Fatalf("%v: %s", err, aus)
	}

	if code, _, fehler := fahreAn(t, ort, "bestand-zurueckspielen", sicherung, schau.AngelegtAm); code != CodeGut {
		t.Fatalf("über dem neueren Bestand spielt sie nicht zurück: %d %s", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "sets"); code != CodeGut {
		t.Errorf("danach öffnet die heutige Fassung ihn nicht: %d %s", code, fehler)
	}
}
