package service

// **Ein Baum, der nach dem Start bezogen wurde, wird gelesen** — gefunden
// vom Flächenlauf über die Bezugsstände (G04,
// #2251).
//
// Die Wurzeln standen fest, seit der Verwalter gebaut war. Ein Bezug legt
// seinen Baum **unter** den Paketordner, eine Ebene tiefer, als dort gelesen
// wird — und der Vermerk nach dem Bezug fand die neuen Pakete nicht: ihre
// Herkunft kam nie in den Katalog, an der Kommandozeile so wenig wie im
// Fenster bis zum nächsten Start.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEinSpaeterBezogenerBaumWirdVermerkt(t *testing.T) {
	verwaltet := filepath.Join(t.TempDir(), "pakete")
	if err := os.MkdirAll(verwaltet, 0o750); err != nil {
		t.Fatal(err)
	}

	katalog := &innerkatalog{}
	verwalter := NeuerPaketverwalter(katalog, verwaltet, verwaltet)

	// **Erst nach dem Bau des Verwalters** — so, wie ein Bezug ihn anlegt.
	paket := filepath.Join(verwaltet, "Pack", "data", "spaet.mse-game")
	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(paket, "game"), []byte("mse version: 2.0.0\nshort name: Spät\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := verwalter.Vermerken(t.Context(), Vermerk{
		Ort: filepath.Join(verwaltet, "Pack"), Herkunft: "https://github.com/x/y@abc", Pruefsumme: "sha256:1",
	}); err != nil {
		t.Fatalf("Vermerken: %v", err)
	}

	gefunden := false

	for _, gelesen := range katalog.pakete {
		if gelesen.Name == "spaet" {
			gefunden = gelesen.Herkunft == "https://github.com/x/y@abc" && gelesen.Pruefsumme == "sha256:1"
		}
	}

	if !gefunden {
		t.Errorf("der später bezogene Baum steht nicht mit Herkunft im Katalog: %+v", katalog.pakete)
	}
}

// **Ohne Paketordner gibt es nichts neu zu suchen** — dann gelten die
// Wurzeln, mit denen der Verwalter gebaut wurde, und sonst keine.
func TestOhnePaketordnerGeltenDieFestenWurzeln(t *testing.T) {
	wurzel := t.TempDir()
	verwalter := NeuerPaketverwalter(&innerkatalog{}, "", wurzel)

	if wurzeln := verwalter.abzugehende(); len(wurzeln) != 1 || wurzeln[0] != wurzel {
		t.Errorf("Wurzeln %v, erwartet nur %s", wurzeln, wurzel)
	}
}
