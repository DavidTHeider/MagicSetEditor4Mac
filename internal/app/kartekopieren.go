package app

// Eine Karte kopieren — #625.
//
// **MSE2 hat dafür kein „Duplizieren"**, sondern Kopieren und Einfügen über die
// Zwischenablage. Was daraus hier gilt — die ganze Karte reist mit, die Kopie
// kommt ans Ende, es ist ein Schritt —, steht bei
// `storage.KarteKopieren`; diese Schicht übersetzt nur (§2).

import (
	"context"
	"strconv"
)

// Kartenkopierer legt eine Kopie an — die Naht zum Speicher.
type Kartenkopierer interface {
	KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error)
}

// KarteKopieren hängt eine Kopie der Karte an ihr Set.
//
// **Die Kennung kommt zurück**, aus demselben Grund wie beim Anlegen: wer
// kopiert, will die Kopie aufschlagen, und die Oberfläche soll dafür nicht die
// ganze Liste durchsuchen müssen.
func (a *App) KarteKopieren(ctx context.Context, setID, karteID string) AnlegeAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return AnlegeAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return AnlegeAntwort{Fehler: befund}
	}

	neue, err := a.kopierer.KarteKopieren(ctx, setkennung, kennung)
	if err != nil {
		return AnlegeAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return AnlegeAntwort{KarteID: strconv.FormatInt(neue, 10)}
}

// MitKartenkopierer setzt die Naht — für Tests und den Composition Root.
func (a *App) MitKartenkopierer(kopierer Kartenkopierer) { a.kopierer = kopierer }
