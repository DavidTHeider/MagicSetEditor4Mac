package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Einordnung reicht von der Paketdatei bis in den Katalog** (#244).
//
// Die Naht ist der Grund für diesen Test: `msepaket` liest die beiden Angaben
// seit eben, `storage` hält sie, `domain` ordnet die Arten — und dazwischen
// liegt `lies`, das jedes Feld einzeln überträgt. Ein vergessenes Feld dort
// ist grün auf beiden Seiten und leer in der Mitte.
func TestDerVerwalterUebernimmtDieEinordnung(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketpfad := filepath.Join(wurzel, "magic-m15.mse-style")

	if err := os.MkdirAll(paketpfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: M15\ninstaller group: magic/m15 style\nposition hint: 095\n"
	if err := os.WriteFile(filepath.Join(paketpfad, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), wurzel)

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(katalog.pakete) != 1 {
		t.Fatalf("%d Pakete", len(katalog.pakete))
	}

	if katalog.pakete[0].Gruppe != "magic/m15 style" {
		t.Errorf("Gruppe = %q", katalog.pakete[0].Gruppe)
	}

	if katalog.pakete[0].Rang != 95 {
		t.Errorf("Rang = %d", katalog.pakete[0].Rang)
	}
}

// **Ein Paket ohne Hinweis trägt den Vorgabewert, nicht die Null.**
//
// Der Weg dorthin ist der eigentliche Fall: `msepaket` setzt
// [msepaket.RangOhneHinweis], und wer beim Übertragen ein `int` vergisst,
// bekommt eine 0 — und damit ein Paket, das **vor** allen anderen steht statt
// hinter ihnen. Still falsch, in genau der Richtung, die auffällt.
func TestOhneHinweisTraegtDasPaketDenVorgabewert(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketpfad := filepath.Join(wurzel, "schlicht.mse-style")

	if err := os.MkdirAll(paketpfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(paketpfad, "style"),
		[]byte("mse version: 2.0.0\nshort name: Schlicht\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), wurzel)

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if katalog.pakete[0].Rang != msepaket.RangOhneHinweis {
		t.Errorf("Rang = %d, erwartet %d", katalog.pakete[0].Rang, msepaket.RangOhneHinweis)
	}
}
