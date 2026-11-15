package service

// **Ein Paketordner wird nach der Grenze nicht weiter durchlaufen** —
// Befund 1 des Nachlaufs über UX-13
// (#1685):
// die Liste war begrenzt, der Gang durch den Baum nicht.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDerOrdnergangEndetAnDerGrenze(t *testing.T) {
	wurzel := t.TempDir()

	for ordner := range 3 {
		unter := filepath.Join(wurzel, "ordner"+strconv.Itoa(ordner))
		if fehler := os.Mkdir(unter, 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		for datei := range 20 {
			if fehler := os.WriteFile(filepath.Join(unter, strconv.Itoa(datei)), nil, 0o600); fehler != nil {
				t.Fatal(fehler)
			}
		}
	}

	besucht := 0
	vorher := ordnergang
	ordnergang = func(wurzel string, schritt fs.WalkDirFunc) error {
		return vorher(wurzel, func(pfad string, eintrag fs.DirEntry, unterwegs error) error {
			besucht++

			return schritt(pfad, eintrag, unterwegs)
		})
	}

	t.Cleanup(func() { ordnergang = vorher })

	eintraege, _ := ordnerEintraege(wurzel, 5)
	if len(eintraege) != 5 {
		t.Errorf("Einträge: %d", len(eintraege))
	}

	// Die Wurzel, fünf Einträge und der eine, an dem der Gang endet.
	if besucht > 7 {
		t.Errorf("nach der Grenze weitergelaufen: %d Schritte", besucht)
	}
}
