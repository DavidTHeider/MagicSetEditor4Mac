package service

// **Der Abzug eines Tisches wird geschrieben** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Warum ein eigener Dienst und kein Zweig im Bildausfuhrdienst
//
// **Weil es eine andere Frage ist** (§2). `Bildausfuhr` rastert eine **Karte**
// aus einem Zeichenplan — sie kennt Güte, Zielbreite und Namensmuster. Hier
// kommt ein fertiges Bild aus dem Fenster, und zu entscheiden bleibt nur, ob
// es eines ist und wohin es geht.
//
// # Und warum das Fenster es rastert
//
// **Weil nur dort ein Tisch steht.** Der Abzug ist das, was die Leinwand
// gerade zeigt — mit Kamera, Drehung und gewähltem Tischbild. Ihn auf der
// Go-Seite nachzubauen hieße, den ganzen Zeichenweg ein zweites Mal zu haben,
// und die zweite Kopie wäre die veraltete (dieselbe Überlegung wie in
// [Vermerk 0126](../../docs/entscheidungen/0126-der-tisch-zeichnet-nicht-optimistisch.md)).
//
// **Was daraus folgt, steht in `FAEHIGKEITEN.md`**: diese Fähigkeit gibt es
// nur am Fenster, und der Grund ist derselbe wie beim Dateiwahldialog — ohne
// Fenster hat sie keinen Gegenstand.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tischbildausfuhr schreibt einen Abzug des Tisches als PNG.
type Tischbildausfuhr struct{}

// NeueTischbildausfuhr gibt den Dienst.
func NeueTischbildausfuhr() Tischbildausfuhr { return Tischbildausfuhr{} }

// Schreibe legt den Abzug an den Pfad.
//
// **Der Pfad kommt vom Dialog samt Endung** — macOS setzt sie aus dem Filter.
// Dieser Dienst benennt nichts um: eine Datei, die anders heißt, als jemand
// getippt hat, ist eine Überraschung.
func (Tischbildausfuhr) Schreibe(datenURL, pfad string) error {
	if pfad == "" {
		return fmt.Errorf("%w: es fehlt der Ort", domain.ErrValidation)
	}

	rohes, fehler := domain.DeuteAbzug(datenURL)
	if fehler != nil {
		return fehler
	}

	// **Atomar, wie jedes andere Schreiben hier** (#853): ein halbes PNG
	// sieht im Dateibrowser aus wie ein ganzes.
	if fehler := schreibeAtomar(pfad, rohes); fehler != nil {
		return fmt.Errorf("den Abzug schreiben: %w", fehler)
	}

	return nil
}
