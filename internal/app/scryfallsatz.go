// **Ein ganzer Satz von Scryfall**
// (#1033).
//
// Eine eigene Datei, weil es ein eigener Vorgang ist: `scryfallanlage.go` legt
// **eine** Karte an — hier entsteht ein Set.

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Satzquelle ist, was die Fassade zum Einlesen eines Satzes braucht.
type Satzquelle interface {
	Einlesen(ctx context.Context, code, spiel string) (*service.Ergebnis, error)
}

// MitScryfallsaetzen setzt die Naht — für Tests und den Composition Root.
func (a *App) MitScryfallsaetzen(quelle Satzquelle) { a.scryfallsaetze = quelle }

// ScryfallSatzEinlesen legt ein Set aus einem Satzcode an.
//
// # Was zurückkommt, ist [OeffnenAntwort]
//
// **Dieselbe Gestalt wie beim MTG-JSON-Import**, und aus demselben Grund: das
// Fenster zeigt beides an derselben Stelle, und zwei Gestalten für „ein Set ist
// entstanden" wären zwei Wege durch dieselbe Oberfläche.
//
// # Was er nicht tut
//
// **Er holt keine Bilder.** Ein Satz mit 300 Karten wären 300 Abrufe und 300
// Nennungen; der Bildweg je Karte steht daneben.
func (a *App) ScryfallSatzEinlesen(ctx context.Context, code, spiel string) OeffnenAntwort {
	ctx = a.imProfil(ctx)

	leer := OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}}

	if a.scryfallsaetze == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	if code == "" {
		leer.Abgebrochen = true

		return leer
	}

	ergebnis, err := a.scryfallsaetze.Einlesen(ctx, code, spiel)

	// **Ein Abbruch gibt das Ergebnis mit** (#155) — dieselbe Haltung wie beim
	// MTG-JSON-Import: was vor dem Fehler durchkam, steht im Bestand und
	// gehört benannt, nicht verschwiegen.
	if ergebnis != nil {
		leer.SetID = strconv.FormatInt(ergebnis.SetID, 10)
		leer.Karten = ergebnis.Karten
		leer.SchonOffen = ergebnis.SchonOffen

		for _, kennung := range ergebnis.Angelegt {
			leer.Angelegt = append(leer.Angelegt, strconv.FormatInt(kennung, 10))
		}

		for _, befund := range ergebnis.Befunde {
			leer.Befunde = append(leer.Befunde, befund.String())
		}
	}

	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)
	}

	return leer
}
