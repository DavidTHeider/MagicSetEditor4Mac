package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Der ganze Weg: mitgeliefertes Set anklicken, Karten haben** —
// #702.
//
// # Was er prüft, das die anderen nicht prüfen
//
// `beispielsets_test.go` prüft die Liste und die Absage; hier geht der Weg
// **durch** die Bindung in den Bestand. Ohne ihn bliebe der Erfolgszweig von
// [Bindungen.BeispielsetOeffnen] ungeprüft — und das ist der Zweig, den jemand
// wirklich geht.
func TestEinBeispielsetLaesstSichAufschlagen(t *testing.T) {
	ort := t.TempDir()

	// **Ein echtes Set und keine Zeichenkette**: `SetOeffnen` liest eine
	// Paketdatei, und eine Vorrichtung, die ärmer ist als der Betrieb, prüft
	// den falschen Weg (§3).
	satz := filepath.Join(ort, "Probe.mse-set")
	if err := os.MkdirAll(satz, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: probe-schlicht\n" +
		"set info:\n\ttitle: Mitgeliefert\ncard:\n\tname: Feuerball\n"

	if err := os.WriteFile(filepath.Join(satz, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := mitSaat(t)
	anwendung.MitBeispielsetort(func() string { return ort })

	bindungen := NeueBindungen(anwendung)

	// **Zwei, nicht eines**: neben dem mitgelieferten steht seit U7 (#1418)
	// das Testkarten-Set der Saat, und `Probe` sortiert davor.
	liste := bindungen.Beispielsets(t.Context())
	if liste.Fehler != nil || len(liste.Sets) != 2 || liste.Sets[0].Name != "Probe" {
		t.Fatalf("Liste = %+v", liste)
	}

	geoeffnet := bindungen.BeispielsetOeffnen(t.Context(), liste.Sets[0].Pfad)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Karten) != 1 {
		t.Fatalf("Karten = %+v", karten)
	}
}
