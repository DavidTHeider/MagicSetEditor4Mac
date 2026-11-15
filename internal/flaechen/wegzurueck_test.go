// **Der Weg zurück wird benannt** — Werkstück 5 von Phase 4.13
// (#1352).
//
// # Warum am gebauten Programm
//
// **Weil die Auskunft dort entsteht, wo es kein Fenster gibt.** Das Fenster
// zeigt den Weg zurück auf der Wiederherstellungsseite; die Kommandozeile
// bekommt ihn an den Fehler gehängt — und ob das wirklich ankommt, sagt nur
// ein Lauf gegen `mse` selbst.
//
// # Und warum die Gegenprobe im selben Lauf steht
//
// **Weil „nennt eine Datei" nichts wert ist, wenn es die falsche sein darf.**
// Neben dem Bestand liegt eine zweite, jüngere Datei mit demselben
// Namensmuster, die keine Sicherung ist. Wer sie nennte, schickte jemanden
// in eine Abweisung.
package flaechen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDieKommandozeileNenntDenWegZurueck(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	// **Erst ein Bestand** — er wird gleich die Sicherung.
	mussKommandozeile(t, umgebung, "sets")

	ort := bestandsortAus(t, umgebung)
	neben := filepath.Dir(ort)

	gut, err := os.ReadFile(ort) //nolint:gosec // Pfad aus der Testumgebung
	if err != nil {
		t.Fatalf("den Bestand lesen: %v", err)
	}

	rueckweg := filepath.Join(neben, "bestand-vor-0.3.0.db")

	//nolint:gosec // der Pfad kommt aus der Testumgebung, nicht von außen
	if err := os.WriteFile(rueckweg, gut, 0o600); err != nil {
		t.Fatalf("die Sicherung anlegen: %v", err)
	}

	// **Und eine jüngere, die nicht aufgeht.** Sie trägt einen Namen, den der
	// Erzeuger wirklich schreibt — nur ist ihr Inhalt keine Datenbank.
	//
	// **Nicht `bestand-vor-muell.db`**, wie hier zuerst stand: seit dem
	// Nachlauf zum Zweitblick erkennt das Muster nur noch Namen mit einer
	// Fassung, und ein erfundener wäre gar keine Sicherung mehr — der Lauf
	// hätte dann etwas anderes geprüft als gemeint.
	falsch := filepath.Join(neben, "bestand-vor-0.4.0.db")
	if err := os.WriteFile(falsch, make([]byte, 4096), 0o600); err != nil {
		t.Fatalf("die Attrappe anlegen: %v", err)
	}

	// **Jetzt geht der Bestand nicht mehr auf.**
	if err := os.WriteFile(ort, []byte("kein Bestand"), 0o600); err != nil {
		t.Fatalf("den Bestand verderben: %v", err)
	}

	for _, begleiter := range []string{ort + "-wal", ort + "-shm"} {
		_ = os.Remove(begleiter) //nolint:errcheck // was nicht da ist, muss nicht weg
	}

	aus, fehler, code := kommandozeile(t, umgebung, "sets")
	if code == 0 {
		t.Fatalf("mse öffnete einen verdorbenen Bestand:\n%s", aus)
	}

	gesagt := aus + fehler

	if !strings.Contains(gesagt, "bestand-vor-0.3.0.db") {
		t.Errorf("die Meldung nennt den Weg zurück nicht:\n%s", gesagt)
	}

	if strings.Contains(gesagt, "bestand-vor-0.4.0.db") {
		t.Errorf("die Meldung schickt in eine Datei, die nicht aufgeht:\n%s",
			gesagt)
	}
}
