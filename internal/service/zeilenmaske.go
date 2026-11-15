package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// zeilenmaskeAus gibt den Pfad der Maske, an der ein Textfeld umbrochen wird
// ([render.Satz.Zeilenmaske], #2299).
//
// **Nur eine Datei.** MSE lädt die Maske über `ScriptableImage::generate` und
// könnte auch eine Verknüpfung nehmen; im Pack ist jede `mask:`-Angabe eines
// Textfelds eine Datei, 52 in 43 Stilen. Eine Verknüpfung bliebe hier ohne
// Wirkung auf den Umbruch — gezeichnet und beschnitten wird sie trotzdem.
func zeilenmaskeAus(maske render.Bildbaum) string {
	if !maske.IstBlatt() {
		return ""
	}

	return maske.Pfad
}

// messmaske rechnet dieselbe Maske für den Messsatz aus. Ihre Befunde meldet
// die Zeichnung, die dieselbe Angabe auswertet.
func messmaske(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
	vorlage Vorlage,
) string {
	maske, _ := bildangabe(ctx, feld.Feld, "mask", feld.Maske, umgebung)
	if maske.Knotenart != "datei" {
		return ""
	}

	return paketpfad(vorlage.Stil.Paket, maske.Pfad)
}
