// **Karten kopieren, ausschneiden und einfügen** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # Die Zwischenablage ist eine Naht wie der Dateidialog
//
// Sie gehört dem Fenster, nicht dem Fachkern — und sie wird **eingesetzt**,
// damit ein Test ohne laufendes Fenster auskommt. Dieselbe Gestalt wie
// [App.MitDateiwahl].

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Ablagequelle ist, was das Fenster vom Kopieren und Einfügen braucht.
type Ablagequelle interface {
	Kopieren(ctx context.Context, setID int64, karteIDs []int64) (string, error)
	// Ausschneiden legt die Karten ab und entfernt sie **danach** — Befund 2
	// des Nachlaufs über Phase 3.7.
	//
	// **`ablegen` ist der Weg nach draußen**, und er muss glücken, bevor eine
	// Karte fällt. Vorher schrieb diese Fassade das Ergebnis erst hinterher
	// ins Klemmbrett; scheiterte das, waren die Karten fort und in der Ablage
	// stand der Inhalt von vorhin.
	Ausschneiden(ctx context.Context, setID int64, karteIDs []int64,
		ablegen func(string) error) (string, error)
	Einfuegen(ctx context.Context, setID int64, inhalt string) (*service.Einfuegung, error)
}

// MitKartenablage setzt den Anwendungsfall — für Tests und den Composition Root.
func (a *App) MitKartenablage(quelle Ablagequelle) { a.ablage = quelle }

// MitZwischenablage setzt die Naht zur Zwischenablage.
//
// **Zwei Funktionen und keine Schnittstelle**: es sind zwei Aufrufe, und Wails
// gibt sie als zwei (`Clipboard.Text`, `Clipboard.SetText`).
func (a *App) MitZwischenablage(lesen func(context.Context) (string, error),
	schreiben func(context.Context, string) error,
) {
	a.ablageLesen, a.ablageSchreiben = lesen, schreiben
}

// EinfuegeAntwort trägt die Kennungen der eingefügten Karten.
type EinfuegeAntwort struct {
	// Karten sind die neuen Karten, in der Reihenfolge des Inhalts.
	//
	// **Auch wenn das Füllen scheitert.** Eine Karte, die entstanden ist,
	// verschwindet nicht; wer sie nicht nennt, lässt sie im Bestand zurück,
	// ohne dass jemand davon weiß.
	Karten []string `json:"karten"`

	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}

// AblageKopieren legt die genannten Karten in die Zwischenablage.
func (a *App) AblageKopieren(
	ctx context.Context, setID string, karteIDs []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDieAblage(ctx, setID, karteIDs,
		func(quelle Ablagequelle, kennung int64, karten []int64) (string, error) {
			return quelle.Kopieren(ctx, kennung, karten)
		}, false)
}

// AblageAusschneiden legt sie hinein und entfernt sie danach aus dem Set.
func (a *App) AblageAusschneiden(
	ctx context.Context, setID string, karteIDs []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.inDieAblage(ctx, setID, karteIDs,
		func(quelle Ablagequelle, kennung int64, karten []int64) (string, error) {
			// **Das Ablegen wird hineingereicht** und nicht danach getan: es
			// muss glücken, bevor eine Karte fällt (Befund 2 des Nachlaufs
			// über Phase 3.7).
			return quelle.Ausschneiden(ctx, kennung, karten, func(inhalt string) error {
				return a.ablageSchreiben(ctx, inhalt)
			})
		}, true)
}

// inDieAblage ist der gemeinsame Rumpf von Kopieren und Ausschneiden.
//
// **Ein Rumpf für beide Wege**: zwei gleichlautende wären zwei, von denen nur
// einer je einen Test sähe (§3).
//
// **Und geschrieben wird auch bei einem Fehler, wenn Inhalt da ist.** Ein
// gescheitertes Löschen beim Ausschneiden gibt den Inhalt trotzdem heraus —
// er in der Ablage und die Karte im Set ist wiederholbar; beides zu verlieren
// nicht.
func (a *App) inDieAblage(
	ctx context.Context, setID string, karteIDs []string,
	hole func(Ablagequelle, int64, []int64) (string, error),
	schonAbgelegt bool,
) SchreibAntwort {
	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	karten, befund := kennungen(karteIDs)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.ablage == nil || a.ablageSchreiben == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Zwischenablage ist nicht eingerichtet", domain.ErrValidation))}
	}

	inhalt, fehler := hole(a.ablage, kennung, karten)
	if inhalt == "" {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Das Kopieren legt hier ab, das Ausschneiden hat es schon getan.**
	//
	// Beim Ausschneiden reicht [App.AblageAusschneiden] das Ablegen **hinein**
	// — es muss glücken, bevor eine Karte fällt (Befund 2 des Nachlaufs über
	// Phase 3.7). Ein zweites Schreiben hier wäre dasselbe noch einmal.
	if !schonAbgelegt {
		if schreibfehler := a.ablageSchreiben(ctx, inhalt); schreibfehler != nil {
			return SchreibAntwort{Fehler: Uebersetze(ctx, schreibfehler)}
		}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
}

// AblageEinfuegen legt die Karten aus der Zwischenablage im Set an.
func (a *App) AblageEinfuegen(ctx context.Context, setID string) EinfuegeAntwort {
	ctx = a.imProfil(ctx)

	leer := EinfuegeAntwort{Karten: []string{}, Befunde: []string{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.ablage == nil || a.ablageLesen == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: die Zwischenablage ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	inhalt, lesefehler := a.ablageLesen(ctx)
	if lesefehler != nil {
		leer.Fehler = Uebersetze(ctx, lesefehler)

		return leer
	}

	ergebnis, fehler := a.ablage.Einfuegen(ctx, kennung, inhalt)
	if ergebnis == nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	return EinfuegeAntwort{
		Karten:  kennungstexte(ergebnis.Karten),
		Befunde: befundtexte(ergebnis.Befunde),
		Fehler:  Uebersetze(ctx, fehler),
	}
}

// kennungen liest eine Liste von Kartenkennungen.
//
// **Eine unbrauchbare wird abgewiesen und nicht übergangen** — dieselbe Regel
// wie bei der Kartenauswahl: wer sie wegließe, kopierte etwas anderes, als
// jemand gemeint hat.
func kennungen(roh []string) ([]int64, *Fehler) {
	gelesen := make([]int64, 0, len(roh))

	for _, eine := range roh {
		kennung, befund := benannteKennungAus("karteID", eine)
		if befund != nil {
			return nil, befund
		}

		gelesen = append(gelesen, kennung)
	}

	return gelesen, nil
}

// kennungstexte macht aus Kennungen die Zeichenketten der Naht.
func kennungstexte(kennungen []int64) []string {
	texte := make([]string, 0, len(kennungen))
	for _, kennung := range kennungen {
		texte = append(texte, fmt.Sprint(kennung))
	}

	return texte
}
