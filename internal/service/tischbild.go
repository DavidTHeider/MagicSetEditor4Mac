package service

// **Ein eigenes Tischbild** — Werkstück G3 von Phase 5.5
// (#1460):
// „neutrale und rechtlich verwendbare Hintergründe **einschließlich Import**".
//
// # Warum es beim Menschen liegt
//
// **Dieselbe Überlegung wie bei der Wahl aus den sechs Sätzen**: wer seinen
// Tisch so sehen will, sieht ihn so in jedem seiner Profile. Es liegt deshalb
// neben dem Profilbild unter `nutzer/` — und nicht bei einem Set, das man
// löschen kann.
//
// # Und warum es hier steht und nicht in `nutzerbild.go`
//
// **Weil jene Datei an der Zeilengrenze steht** (§2) und weil der Schnitt der
// Frage folgt: dort geht es um das Bild **eines Menschen**, hier um das Bild
// **eines Tisches**. Dass beide unter demselben Verzeichnis liegen, ist eine
// Frage der Ablage und nicht der Zuständigkeit.

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
)

// Tischbildkante ist die größte Kante eines abgelegten Tischbilds.
//
// **1920 und nicht 256 wie beim Profilbild**: ein Hintergrund füllt die
// Fläche, ein Abzeichen steht neben einem Namen. Und nicht mehr: eine
// Aufnahme aus einer Kamera wiegt zwanzig Megabyte und sieht als Tischbild
// aus wie eine von 1920 Punkten — sie kostet nur das Zwanzigfache, jedes Mal,
// wenn der Tisch aufgeht.
const Tischbildkante = 1920

// TischbildAblegen schreibt das Tischbild eines Nutzers und gibt seinen Namen.
//
// **Es wird überschrieben**, wie das Profilbild: ein Mensch hat **einen**
// Tisch, und ein laufender Zähler hinterließe bei jedem Wechsel eine Datei,
// die niemand mehr findet.
func (b *Bildablage) TischbildAblegen(
	nutzerID int64, bild image.Image,
) (string, error) {
	return b.ablegenUnter(tischbildname(nutzerID), bild, Tischbildkante)
}

// TischbildWegnehmen nimmt das Tischbild eines Nutzers wieder weg.
//
// **Danach steht wieder einer der sechs Sätze da** — die Wahl bleibt, nur das
// Bild ist fort.
func (b *Bildablage) TischbildWegnehmen(nutzerID int64) error {
	return b.wegnehmen(tischbildname(nutzerID))
}

// TischbildLesen gibt das abgelegte Bild — `nil` ohne Fehler heißt: keines.
//
// **Dieselbe Zusage wie [Bildablage.NutzerbildLesen]**, und aus demselben
// Grund: die meisten Menschen haben keines, und ein Fehler wäre eine Meldung,
// die niemand braucht.
func (b *Bildablage) TischbildLesen(name string) (image.Image, error) {
	return b.NutzerbildLesen(name)
}

// TischbildBytes gibt die abgelegte Datei, wie sie ist.
//
// **Und dekodiert sie nicht.** Was dort liegt, hat diese Anwendung selbst als
// PNG geschrieben; es zu lesen und neu zu kodieren kostete zweimal Rechenzeit
// und brächte einen Fehlerweg mit, den kein Lauf je beträte (§3). Wer das
// **Bild** braucht, nimmt [Bildablage.TischbildLesen].
func (b *Bildablage) TischbildBytes(nutzerID int64) ([]byte, bool) {
	return b.nutzerdateiBytes(tischbildname(nutzerID))
}

// nutzerdateiBytes gibt eine Datei unter `nutzer/`, wie sie ist — `false`,
// wenn es sie nicht gibt.
//
// **Für Tisch- und Fensterbild dieselbe**: beide sind PNG, die diese
// Anwendung selbst geschrieben hat, und beide heißen nach einer Zahl.
func (b *Bildablage) nutzerdateiBytes(name string) ([]byte, bool) {
	ort := b.Nutzerverzeichnis()
	if ort == "" {
		return nil, false
	}

	// #nosec G304 -- ein Dateiname aus der eigenen Ablage, aus einer Zahl
	// gebildet
	inhalt, fehler := os.ReadFile(filepath.Join(ort, name))
	if fehler != nil {
		return nil, false
	}

	return inhalt, true
}

// tischbildname ist der Dateiname unter `nutzer/`.
func tischbildname(nutzerID int64) string {
	return fmt.Sprintf("tischbild-%d.png", nutzerID)
}
