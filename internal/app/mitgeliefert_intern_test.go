package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Was das Bündel mitbringt** —
// #702.
//
// # Wofür
//
// Seit die Vorlagen mitgeliefert werden dürfen
// ([0025](../../docs/entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md)),
// legt der Freigabeweg zwei weitere Verzeichnisse neben die Saat: die
// **Vorlagenpakete**, damit eine Karte ohne Bezug gezeichnet wird, und ein
// **Beispielset**, damit der erste Klick nicht durch einen Dateidialog führt.
//
// **Derselbe Weg wie bei der Saat** und aus demselben Grund: sie sind Teil der
// Anwendung, kein Bestand der Nutzenden.
func TestMitgelieferteVerzeichnisseLiegenNebenDemProgramm(t *testing.T) {
	buendel := t.TempDir()

	programm := filepath.Join(buendel, "X.app", "Contents", "MacOS", "X")
	if err := os.MkdirAll(filepath.Dir(programm), 0o750); err != nil {
		t.Fatalf("Bündel: %v", err)
	}

	faelle := []struct {
		name   string
		ordner string
		holen  func(func() (string, error)) string
	}{
		{name: "Vorlagen", ordner: "vorlagen", holen: vorlagenNeben},
		{name: "Beispielsets", ordner: "beispielsets", holen: beispielsetsNeben},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			naht := func() (string, error) { return programm, nil }

			// **Erst wenn es das Verzeichnis gibt.** Ein Bündel ohne
			// mitgelieferte Vorlagen ist der Normalfall beim Entwickeln; ein
			// Pfad ins Leere wäre eine Wurzel, unter der nichts liegt.
			if gefunden := fall.holen(naht); gefunden != "" {
				t.Fatalf("ohne Verzeichnis kam %q", gefunden)
			}

			ort := filepath.Join(buendel, "X.app", "Contents", "Resources", fall.ordner)
			if err := os.MkdirAll(ort, 0o750); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			if gefunden := fall.holen(naht); gefunden != ort {
				t.Errorf("gefunden %q, erwartet %q", gefunden, ort)
			}
		})
	}
}

// **Die beiden öffentlichen Wege gehen durch dieselbe Rechnung** — sie stehen
// hier, weil ein Test ohne `.app` sonst keinen von beiden beträte und der
// Composition Root sie ungeprüft einsetzte.
func TestMitgelieferteVerzeichnisseOhneBuendel(t *testing.T) {
	// **Der Testlauf liegt in keinem `.app`.** Damit gibt jede der beiden
	// Auskünfte den leeren Pfad — genau der Fall „ohne Bündel gestartet".
	if ort := Vorlagenverzeichnis(); ort != "" {
		t.Errorf("Vorlagenverzeichnis = %q", ort)
	}

	if ort := Beispielsetverzeichnis(); ort != "" {
		t.Errorf("Beispielsetverzeichnis = %q", ort)
	}
}
