package service_test

// **Ein eigenes Fensterbild** — UX-09.AK-01
// (#1681):
// „Nutzerhintergründe besitzen Vorschau, Skalierung und eine sofort
// erreichbare Rücksetzung". Es liegt wie das Tischbild beim Menschen, und es
// ist ein anderes Bild als das des Tisches.

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinFensterbildLandetUnterNutzerUndLiegtAlsPNGBereit(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	if _, fehler := ablage.FensterbildAblegen(7, tischprobebild(400, 300)); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	inhalt, da := ablage.FensterbildBytes(7)
	if !da {
		t.Fatal("kein Bild da")
	}

	bild, fehler := png.Decode(bytes.NewReader(inhalt))
	if fehler != nil {
		t.Fatalf("kein PNG: %v", fehler)
	}

	if breite := bild.Bounds().Dx(); breite != 400 {
		t.Errorf("Breite %d, erwartet 400", breite)
	}
}

// **Größer als das Tischbild und doch begrenzt**: ein Fenster auf einem
// Bildschirm mit 5K ist 2.560 Punkte breit; mehr kostet nur.
func TestEinGrossesFensterbildWirdKleiner(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	if _, fehler := ablage.FensterbildAblegen(7, tischprobebild(6000, 4000)); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	inhalt, _ := ablage.FensterbildBytes(7)

	bild, fehler := png.Decode(bytes.NewReader(inhalt))
	if fehler != nil {
		t.Fatalf("kein PNG: %v", fehler)
	}

	if breite := bild.Bounds().Dx(); breite != service.Fensterbildkante {
		t.Errorf("Breite %d, erwartet %d", breite, service.Fensterbildkante)
	}
}

// **Tisch und Fenster sind zwei Bilder** — wer das eine wechselt, behält das
// andere.
func TestFensterbildUndTischbildSindZweiBilder(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())
	abgelegt(t, ablage, 7, 100)

	if _, da := ablage.FensterbildBytes(7); da {
		t.Fatal("das Tischbild gilt als Fensterbild")
	}

	if _, fehler := ablage.FensterbildAblegen(7, tischprobebild(50, 50)); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if fehler := ablage.FensterbildWegnehmen(7); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	if _, da := ablage.FensterbildBytes(7); da {
		t.Fatal("das Fensterbild ist noch da")
	}

	if _, da := ablage.TischbildBytes(7); !da {
		t.Fatal("das Tischbild ging mit")
	}
}

// **Ohne Ablage kein Bild** — und kein Fehler: die meisten haben keines.
func TestOhneAblageGibtEsKeinFensterbild(t *testing.T) {
	t.Parallel()

	if _, da := service.NeueBildablage("").FensterbildBytes(7); da {
		t.Fatal("ein Bild aus dem Nichts")
	}
}
