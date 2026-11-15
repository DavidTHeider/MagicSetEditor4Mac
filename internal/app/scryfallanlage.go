// **Eine Karte aus einem Namen** —
// #1032.
//
// Eine eigene Datei, weil es eine eigene Frage ist: `scryfall.go` gleicht eine
// **vorhandene** Karte ab — hier entsteht eine.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Scryfallanleger ist, was die Fassade zum Anlegen braucht.
type Scryfallanleger interface {
	Anlegen(ctx context.Context, setID int64, spiel, name string) (
		service.Anlageergebnis, error)
}

// spielVonSet gibt das Spiel eines Sets aus dem Bestand.
//
// **Das Fenster gibt es nicht mit**, und das ist Absicht: es steht am Set, und
// eine Fläche, die es mitgeben müsste, könnte es falsch mitgeben. Dieselbe
// Auskunft holt sich die Naht ohne Fenster über [Setspiel].
func (a *App) spielVonSet(ctx context.Context, setID int64) (string, *Fehler) {
	if a.bestand == nil {
		return "", Uebersetze(ctx, fmt.Errorf(
			"%w: der Bestand ist nicht eingerichtet", domain.ErrValidation))
	}

	koepfe, err := a.bestand.Setkoepfe(ctx)
	if err != nil {
		return "", Uebersetze(ctx, err)
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf.Game, nil
		}
	}

	return "", Uebersetze(ctx, fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID))
}

// MitScryfallanlage setzt die Naht — für Tests und den Composition Root.
func (a *App) MitScryfallanlage(anlage Scryfallanleger) { a.scryfallanlage = anlage }

// ScryfallKarteAntwort ist eine angelegte Karte samt dem, was in sie kam.
type ScryfallKarteAntwort struct {
	// KarteID ist die angelegte Karte — **auch wenn das Füllen scheitert**.
	//
	// Eine Karte, die entstanden ist, verschwindet nicht (§7, Regel 3). Wer
	// sie nicht nennt, lässt sie im Bestand zurück, ohne dass jemand davon
	// weiß (#155).
	KarteID string `json:"karteID"`

	// Felder sind die gefüllten Feldnamen, in fester Reihenfolge.
	Felder []string `json:"felder"`

	// Befunde sind die Auffälligkeiten der Antwort — etwa ein unbekannter
	// Aufzählungswert.
	Befunde []string `json:"befunde"`

	Herkunft string `json:"herkunft"`

	// AlterStunden sagt, wie alt der Stand ist (§16, Regel 5).
	AlterStunden float64 `json:"alterStunden"`

	// Nennung steht dabei, wenn ein Bild mitgeholt wurde.
	Nennung string `json:"nennung"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// ScryfallKarteAnlegen legt eine Karte an und füllt sie aus Scryfall.
//
// # Das Bild ist wahlweise dabei
//
// **Ohne `art` wird keines geholt.** Ein Bild ist ein eigener Weg mit einer
// eigenen Bedingung — Künstlername und Copyright in derselben Oberfläche — und
// mit einem eigenen Abruf. Wer nur die Textfelder will, soll ihn nicht zahlen.
//
// **Mit `art` muss `bildfeld` dastehen.** Welches Feld das Bild aufnimmt,
// weiß nur die Vorlage; es zu raten hieße, bei jeder fremden Vorlage falsch zu
// raten.
func (a *App) ScryfallKarteAnlegen(
	ctx context.Context, setID, name, bildfeld, art string,
) ScryfallKarteAntwort {
	ctx = a.imProfil(ctx)

	if a.scryfallanlage == nil {
		return ScryfallKarteAntwort{
			Felder: []string{}, Befunde: []string{},
			Fehler: Uebersetze(ctx, fmt.Errorf(
				"%w: der Zugang zu Scryfall ist nicht eingerichtet",
				domain.ErrValidation)),
		}
	}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return ScryfallKarteAntwort{
			Felder: []string{}, Befunde: []string{}, Fehler: befund,
		}
	}

	if art != "" && bildfeld == "" {
		return ScryfallKarteAntwort{
			Felder: []string{}, Befunde: []string{},
			Fehler: Uebersetze(ctx, fmt.Errorf(
				"%w: zu einer Bildart gehört das Feld, das das Bild aufnimmt",
				domain.ErrValidation)),
		}
	}

	spiel, spielbefund := a.spielVonSet(ctx, kennung)
	if spielbefund != nil {
		return ScryfallKarteAntwort{
			Felder: []string{}, Befunde: []string{}, Fehler: spielbefund,
		}
	}

	ergebnis, err := a.scryfallanlage.Anlegen(ctx, kennung, spiel, name)

	antwort := ScryfallKarteAntwort{
		KarteID: kennungText(ergebnis.KarteID),
		Felder:  nichtNil(felderAus(ergebnis.Uebernommen)),
		Befunde: nichtNil(befundtexte(ergebnis.Befunde)), Herkunft: ergebnis.Herkunft,
		AlterStunden: ergebnis.Alter.Hours(),
	}

	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	if art == "" {
		return antwort
	}

	return a.bildDazu(ctx, setID, antwort, bildfeld, name, art)
}

// bildDazu holt das Bild und setzt es in die eben angelegte Karte.
//
// **Über denselben Weg wie sonst** ([App.ScryfallBild] und
// [App.ScryfallBildEinsetzen]): eine zweite Fassung liefe bei der ersten
// Änderung an der Nennungspflicht auseinander.
func (a *App) bildDazu(
	ctx context.Context, setID string, antwort ScryfallKarteAntwort,
	bildfeld, name, art string,
) ScryfallKarteAntwort {
	geholt := a.ScryfallBild(ctx, antwort.KarteID, bildfeld, name, art)
	if geholt.Fehler != nil {
		antwort.Fehler = geholt.Fehler

		return antwort
	}

	gesetzt := a.ScryfallBildEinsetzen(ctx, setID, antwort.KarteID, bildfeld,
		geholt.Pfad, geholt.Kuenstler, geholt.Adresse, geholt.Auswahl)
	if gesetzt.Fehler != nil {
		antwort.Fehler = gesetzt.Fehler

		return antwort
	}

	antwort.Felder = append(antwort.Felder, bildfeld)
	antwort.Nennung = geholt.Nennung

	return antwort
}

// felderAus zieht die Feldnamen aus den übernommenen Änderungen.
func felderAus(uebernommen []service.Feldvorschlag) []string {
	felder := make([]string, 0, len(uebernommen))
	for _, eine := range uebernommen {
		felder = append(felder, eine.Feld)
	}

	return felder
}

// befundtexte macht aus Befunden die Sätze, die eine Fläche zeigt.
func befundtexte(befunde []domain.Befund) []string {
	texte := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		texte = append(texte, befund.String())
	}

	return texte
}
