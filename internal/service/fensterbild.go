package service

// **Ein eigenes Fensterbild** — UX-09.AK-01
// (#1681):
// *„Nutzerhintergründe besitzen Vorschau, Skalierung und eine sofort
// erreichbare Rücksetzung."*
//
// # Warum es neben dem Tischbild liegt und nicht dasselbe ist
//
// **Weil es zwei Flächen sind.** Ein Tisch trägt Filz oder Holz, ein Fenster
// eher ein ruhiges Motiv; wer beides dasselbe sein lassen will, wählt
// zweimal dieselbe Datei. Es liegt aus demselben Grund beim **Menschen**:
// wer sein Fenster so sehen will, sieht es in jedem seiner Profile so — ob
// es erscheint, sagt dagegen die Einstellung `hintergrund` des Profils.

import (
	"fmt"
	"image"
)

// Fensterbildkante ist die größte Kante eines abgelegten Fensterbilds.
//
// **2.560 und nicht 1.920 wie beim Tischbild**: ein Fenster auf einem
// Bildschirm mit 5K ist so breit, und ein Hintergrund, der hochgerechnet
// wird, ist unscharf. Mehr kostet nur — bei jedem Start.
const Fensterbildkante = 2560

// FensterbildAblegen schreibt das Fensterbild eines Nutzers und gibt seinen
// Namen. **Es wird überschrieben**, wie das Tischbild.
func (b *Bildablage) FensterbildAblegen(
	nutzerID int64, bild image.Image,
) (string, error) {
	return b.ablegenUnter(fensterbildname(nutzerID), bild, Fensterbildkante)
}

// FensterbildWegnehmen nimmt das Fensterbild eines Nutzers wieder weg.
func (b *Bildablage) FensterbildWegnehmen(nutzerID int64) error {
	return b.wegnehmen(fensterbildname(nutzerID))
}

// FensterbildBytes gibt die abgelegte Datei, wie sie ist — `false`, wenn es
// keine gibt.
func (b *Bildablage) FensterbildBytes(nutzerID int64) ([]byte, bool) {
	return b.nutzerdateiBytes(fensterbildname(nutzerID))
}

// fensterbildname ist der Dateiname unter `nutzer/`.
func fensterbildname(nutzerID int64) string {
	return fmt.Sprintf("fensterbild-%d.png", nutzerID)
}
