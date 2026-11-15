//go:build darwin

// **Die Ränder der Verträglichkeitsprüfung** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Von innen**, weil die zwei Nähte unexportiert sind: sie zählen die
// laufende Arbeit und machen die Sicherungsprobe. Beide gehören zum Fenster,
// und beide werden hier gegen echte Dinge geprüft.
package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Ein Fenster ohne Auftrag hat nichts laufen** — und sagt das, statt zu
// schweigen.
func TestOhneAuftragLaeuftNichts(t *testing.T) {
	t.Parallel()

	laufen, fehler := vorgangsnaht{anwendung: NeueApp("")}.LaufendeVorgaenge(t.Context())

	switch {
	case fehler != nil:
		t.Fatalf("zählen: %v", fehler)
	case laufen != 0:
		t.Errorf("es liefen %d Vorgänge", laufen)
	}
}

// **Ohne Einstellungsspeicher wird nicht gefragt.** Eine Anwendung vor dem
// Start geht nicht ins Netz — und „aus" ist hier kein erfundener Wert,
// sondern die Vorgabe.
func TestOhneEinstellungsspeicherIstDieAktualisierungAus(t *testing.T) {
	t.Parallel()

	an, fehler := NeueApp("").aktualisierungAn(t.Context())

	switch {
	case fehler != nil:
		t.Fatalf("fragen: %v", fehler)
	case an:
		t.Error("ohne Einstellungsspeicher war die Aktualisierung an")
	}
}

// **Die Sicherungsprobe macht eine Sicherung und liest sie zurück** — gegen
// eine gestartete Anwendung, denn genau dort entsteht sie im Betrieb.
//
// **Von innen**, weil `bestandsweg` unexportiert ist: eine Naht, die nur der
// Composition Root baut, muss nicht hinaus.
//
// **Und die Probe lässt nichts liegen**: sie ist der Beweis, nicht die
// Sicherung, die jemand aufbewahrt.
func TestDieSicherungsprobeSichertUndLiestZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bestand, fehler := Datenbankpfad()
	if fehler != nil {
		t.Fatalf("Datenbankpfad: %v", fehler)
	}

	if anlegefehler := os.MkdirAll(filepath.Dir(bestand), 0o700); anlegefehler != nil {
		t.Fatalf("anlegen: %v", anlegefehler)
	}

	anwendung := NeueApp(bestand)
	if startfehler := anwendung.Starten(t.Context()); startfehler != nil {
		t.Fatalf("starten: %v", startfehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	if fehler := SicherungsprobeUeber(
		anwendung.bestandsweg()).SichereUndPruefe(t.Context()); fehler != nil {
		t.Fatalf("die Probe scheiterte: %v", fehler)
	}
}

// **Eine Probe an einem geschlossenen Bestand scheitert — und sagt, woran.**
// Das ist der Fall, für den es das Hindernis gibt: eine Sicherung, die
// niemand schreiben kann, ist keine.
func TestEineSicherungsprobeAmGeschlossenenBestandScheitert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bestand, fehler := Datenbankpfad()
	if fehler != nil {
		t.Fatalf("Datenbankpfad: %v", fehler)
	}

	if anlegefehler := os.MkdirAll(filepath.Dir(bestand), 0o700); anlegefehler != nil {
		t.Fatalf("anlegen: %v", anlegefehler)
	}

	anwendung := NeueApp(bestand)
	if startfehler := anwendung.Starten(t.Context()); startfehler != nil {
		t.Fatalf("starten: %v", startfehler)
	}

	probe := SicherungsprobeUeber(anwendung.bestandsweg())

	if beendefehler := anwendung.Beenden(); beendefehler != nil {
		t.Fatalf("beenden: %v", beendefehler)
	}

	if probefehler := probe.SichereUndPruefe(t.Context()); probefehler == nil {
		t.Error("die Probe am geschlossenen Bestand gelang")
	}
}
