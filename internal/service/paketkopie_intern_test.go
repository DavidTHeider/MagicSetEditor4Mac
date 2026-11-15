package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Die beiden Fehlerwege von [kopierePaket], die kein Aufrufer erreicht.
//
// **Ein interner Test, weil die Aufrufer vorher prüfen.** `Installieren` und
// `uebernimm` schlagen den Platz nach, bevor sie kopieren; von dort aus ist
// weder ein fehlendes Verwaltungsverzeichnis noch ein belegter Zielname zu
// erreichen. Beide Wege gibt es trotzdem — ein zweiter Vorgang kann zwischen
// Nachschlagen und Umbenennen dasselbe tun.
func TestKopierePaketMeldetSeineFehlerwege(t *testing.T) {
	t.Parallel()

	t.Run("ohne Verwaltungsverzeichnis", func(t *testing.T) {
		t.Parallel()

		quelle := t.TempDir()

		// **Kein Verzeichnis, in das das Vorläufige passte.** In der Anwendung
		// legt `app.Paketverzeichnis()` es an; hier fehlt es.
		fehler := kopierePaket(quelle, filepath.Join(t.TempDir(), "fehlt", "paket.mse-style"))
		if fehler == nil {
			t.Fatal("ein fehlendes Verwaltungsverzeichnis blieb unbemerkt")
		}

		if !strings.Contains(fehler.Error(), "kopieren") {
			t.Errorf("Meldung: %v", fehler)
		}
	})

	t.Run("auf einen belegten Namen", func(t *testing.T) {
		t.Parallel()

		quelle := t.TempDir()
		if err := os.WriteFile(filepath.Join(quelle, "style"), []byte("x"), 0o600); err != nil {
			t.Fatalf("Quelle anlegen: %v", err)
		}

		verwaltet := t.TempDir()
		ziel := filepath.Join(verwaltet, "paket.mse-style")

		// **Belegt und nicht leer**: ein Umbenennen auf ein leeres Verzeichnis
		// ginge durch, auf ein volles nicht.
		if err := os.MkdirAll(filepath.Join(ziel, "darin"), 0o750); err != nil {
			t.Fatalf("Platz belegen: %v", err)
		}

		fehler := kopierePaket(quelle, ziel)
		if fehler == nil {
			t.Fatal("ein belegter Zielname blieb unbemerkt")
		}

		if !strings.Contains(fehler.Error(), "an seinen Platz") {
			t.Errorf("Meldung: %v", fehler)
		}

		// **Und das Vorläufige ist fort.** Ein gescheitertes Umbenennen ließe
		// sonst einen halben Bestand in der Wurzel liegen — genau das, wogegen
		// dieser Weg gebaut ist.
		eintraege, err := os.ReadDir(verwaltet)
		if err != nil {
			t.Fatalf("Wurzel lesen: %v", err)
		}

		if len(eintraege) != 1 {
			t.Errorf("in der Wurzel steht mehr als das Paket: %v", eintraege)
		}
	})
}
