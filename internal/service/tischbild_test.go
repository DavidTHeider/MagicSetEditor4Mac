package service_test

// **Ein eigenes Tischbild** — Werkstück G3 von Phase 5.5
// (#1460):
// „neutrale und rechtlich verwendbare Hintergründe **einschließlich Import**".
//
// # Warum es beim Menschen liegt und nicht beim Profil
//
// **Dieselbe Überlegung wie bei der Wahl aus den sechs Sätzen**: wer seinen
// Tisch so sehen will, sieht ihn so in jedem seiner Profile. Es liegt deshalb
// neben dem Profilbild unter `nutzer/` — und nicht bei einem Set, das man
// löschen kann.

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tischprobebild gibt ein Bild der gewünschten Größe.
func tischprobebild(breite, hoehe int) image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	return bild
}

// abgelegt legt ein Bild ab und gibt seinen Namen — oder bricht ab.
func abgelegt(
	t *testing.T, ablage *service.Bildablage, nutzerID int64, kante int,
) string {
	t.Helper()

	name, fehler := ablage.TischbildAblegen(nutzerID, tischprobebild(kante, kante))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	return name
}

func TestEinTischbildLandetUnterNutzer(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	name, fehler := ablage.TischbildAblegen(7, tischprobebild(400, 300))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.TischbildLesen(name)
	switch {
	case lesefehler != nil:
		t.Fatalf("lesen: %v", lesefehler)
	case gelesen == nil:
		t.Fatal("kein Bild zurück")
	case gelesen.Bounds().Dx() != 400:
		t.Errorf("Breite %d, erwartet 400", gelesen.Bounds().Dx())
	}
}

// **Ein Hintergrund darf groß sein, aber nicht beliebig groß.**
//
// Eine Aufnahme aus einer Kamera wiegt zwanzig Megabyte; als Tischbild ist sie
// so gut wie eine von 1920 Punkten und kostet das Zwanzigfache — jedes Mal,
// wenn der Tisch aufgeht.
func TestEinGrossesTischbildWirdKleiner(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	name, fehler := ablage.TischbildAblegen(7, tischprobebild(6000, 4000))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.TischbildLesen(name)
	if lesefehler != nil || gelesen == nil {
		t.Fatalf("lesen: %v, Bild da: %t", lesefehler, gelesen != nil)
	}

	if breite := gelesen.Bounds().Dx(); breite != service.Tischbildkante {
		t.Errorf("Breite %d, erwartet %d", breite, service.Tischbildkante)
	}
}

// **Ein kleines bleibt, wie es ist** — hochgerechnet wäre es nur unschärfer.
func TestEinKleinesTischbildBleibt(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	name, ablegefehler := ablage.TischbildAblegen(7, tischprobebild(320, 200))
	if ablegefehler != nil {
		t.Fatalf("ablegen: %v", ablegefehler)
	}

	gelesen, lesefehler := ablage.TischbildLesen(name)
	if lesefehler != nil || gelesen == nil {
		t.Fatalf("lesen: %v, Bild da: %t", lesefehler, gelesen != nil)
	}

	if breite := gelesen.Bounds().Dx(); breite != 320 {
		t.Errorf("Breite %d, erwartet 320", breite)
	}
}

// **Jeder Mensch hat eines**, und es wird überschrieben — wie sein Profilbild.
func TestJederNutzerHatEinTischbild(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	einer := abgelegt(t, ablage, 7, 100)
	nochmal := abgelegt(t, ablage, 7, 200)
	anderer := abgelegt(t, ablage, 8, 100)

	switch {
	case einer != nochmal:
		t.Errorf("zweimal derselbe Nutzer, zwei Namen: %q und %q", einer, nochmal)
	case einer == anderer:
		t.Errorf("zwei Nutzer, ein Name: %q", einer)
	}
}

// **Weggenommen ist weggenommen** — und danach steht wieder ein Satz aus den
// sechsen da.
func TestTischbildWegnehmen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())
	name := abgelegt(t, ablage, 7, 100)

	if fehler := ablage.TischbildWegnehmen(7); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	gelesen, lesefehler := ablage.TischbildLesen(name)
	switch {
	case lesefehler != nil:
		t.Errorf("lesen: %v", lesefehler)
	case gelesen != nil:
		t.Error("das Bild liegt noch da")
	}
}

// **Wo keines liegt, ist das keine Klage** (§8) — dieselbe Zusage wie beim
// Profilbild: die meisten Menschen haben keines.
func TestKeinTischbildIstKeinFehler(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	gelesen, fehler := ablage.TischbildLesen("tischbild-99.png")
	switch {
	case gelesen != nil:
		t.Error("ein Bild, wo keines liegt")
	case fehler != nil:
		t.Errorf("Fehler %v, erwartet keinen", fehler)
	}
}

// **Die rohen Bytes kommen, wie sie liegen.**
//
// **Und werden nicht dekodiert**: was dort liegt, hat diese Anwendung selbst
// als PNG geschrieben; es zu lesen und neu zu kodieren kostete zweimal
// Rechenzeit und brächte einen Fehlerweg mit, den kein Lauf je beträte (§3).
func TestTischbildBytesGibtDieDatei(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())
	abgelegt(t, ablage, 7, 100)

	bytes, da := ablage.TischbildBytes(7)
	switch {
	case !da:
		t.Fatal("keine Bytes zurück")
	case len(bytes) < 8:
		t.Fatalf("%d Bytes — zu wenig für ein PNG", len(bytes))
	case string(bytes[1:4]) != "PNG":
		t.Errorf("kein PNG: % x", bytes[:8])
	}
}

// **Wo nichts liegt, kommt nichts** — und das ist keine Klage.
func TestTischbildBytesSchweigtOhneDatei(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	if _, da := ablage.TischbildBytes(99); da {
		t.Error("Bytes, wo keine Datei liegt")
	}
}

// **Ohne Verzeichnis kommt auch nichts.**
func TestTischbildBytesSchweigtOhneVerzeichnis(t *testing.T) {
	t.Parallel()

	if _, da := service.NeueBildablage("").TischbildBytes(7); da {
		t.Error("Bytes ohne eingerichtetes Verzeichnis")
	}
}

// **Ein Bild ohne Fläche wird auch als Tischbild abgewiesen** — derselbe Weg
// wie beim Profilbild, seit beide Namensarten getrennt stehen (Zweitblick
// über Phase 6.1, Befund 3): `png.Encode` weist es ab, und es entsteht keine
// Datei.
func TestEinTischbildOhneFlaecheWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	if _, fehler := ablage.TischbildAblegen(7, image.NewNRGBA(image.Rect(0, 0, 0, 0))); fehler == nil {
		t.Fatal("ein Bild ohne Fläche kam an")
	}

	if _, da := ablage.TischbildBytes(7); da {
		t.Error("es liegt trotzdem eine Datei da")
	}
}
