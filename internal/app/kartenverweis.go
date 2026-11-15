// **Der Verweis einer Karte auf einen echten Druck an der Fassade** —
// Werkstück 2 von Phase 3.6
// (#1167).
//
// **Eine eigene Datei**, weil `scryfall.go` mit ihr über 400 Zeilen ginge (§2)
// — und weil der Schnitt einer Frage folgt: dort steht der Abgleich, hier der
// Bezug, dem er folgt.

package app

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verweisquelle ist, was das Fenster von den Verweisen braucht.
type Verweisquelle interface {
	Setzen(ctx context.Context, karteID int64, printing string) (
		*domain.Kartenverweis, []domain.Befund, error)
	Loesen(ctx context.Context, karteID int64) error
	Fuer(ctx context.Context, karteID int64) (*domain.Kartenverweis, error)
	JeSet(ctx context.Context, setID int64) (map[int64]domain.Kartenverweis, error)
}

// MitVerweisen setzt die Naht — für Tests und für den Composition Root.
func (a *App) MitVerweisen(quelle Verweisquelle) { a.verweise = quelle }

// KartenverweisDTO ist ein Verweis, wie das Fenster ihn sieht.
type KartenverweisDTO struct {
	// KarteID ist die Karte, die verweist — als Zeichenkette, wie jede
	// Kennung an dieser Naht.
	KarteID string `json:"karteID"`

	// Printing ist die Scryfall-Kennung des Drucks.
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441), als
	// Zeichenkette; Vorlaeufig sagt, dass der Druck keinen Alias hat.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Adresse ist die Kartenseite bei Scryfall — anklickbar.
	Adresse string `json:"adresse"`

	// Gesetzt ist der Zeitpunkt in RFC 3339, wie bei den Preisständen.
	//
	// **Er sagt, wie alt die Anzeige ist** (§16, Regel 5): Name, Setcode und
	// Nummer stammen aus der Antwort von damals.
	Gesetzt string `json:"gesetzt"`
}

// VerweisAntwort trägt einen Verweis oder keinen.
type VerweisAntwort struct {
	// Verweis ist `null`, wenn die Karte keinen hat — das ist der Normalfall
	// und kein Fehler.
	Verweis *KartenverweisDTO `json:"verweis"`

	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}

// VerweiseAntwort trägt die Verweise eines ganzen Sets.
//
// **Eine Liste und keine Abbildung.** Eine Map mit Zahlen als Schlüssel wird
// über JSON zu einer Map mit Zeichenketten, und die Oberfläche müsste sie
// zurückrechnen; die Kennung steht deshalb **im** Eintrag.
type VerweiseAntwort struct {
	Verweise []KartenverweisDTO `json:"verweise"`
	Fehler   *Fehler            `json:"fehler"`
}

// alsVerweisDTO übersetzt einen Verweis.
func alsVerweisDTO(karteID int64, verweis domain.Kartenverweis) KartenverweisDTO {
	return KartenverweisDTO{
		KarteID: fmt.Sprint(karteID), Printing: verweis.Printing,
		DruckID: kennungtext(verweis.DruckID), Vorlaeufig: verweis.Vorlaeufig(),
		Name: verweis.Name, Setcode: verweis.Setcode, Nummer: verweis.Nummer,
		Adresse: verweis.Adresse,
		Gesetzt: verweis.Gesetzt.UTC().Format(time.RFC3339),
	}
}

// verweisquelle gibt die Quelle oder den Fehler, wenn sie fehlt.
func (a *App) verweisquelle() (Verweisquelle, *Fehler) {
	if a.verweise == nil {
		return nil, Nutzerfehler(fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))
	}

	return a.verweise, nil
}

// KarteVerweis gibt den Verweis einer Karte — oder keinen.
func (a *App) KarteVerweis(ctx context.Context, karteID string) VerweisAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return VerweisAntwort{Befunde: []string{}, Fehler: befund}
	}

	quelle, fehlt := a.verweisquelle()
	if fehlt != nil {
		return VerweisAntwort{Befunde: []string{}, Fehler: fehlt}
	}

	verweis, fehler := quelle.Fuer(ctx, kennung)
	if fehler != nil {
		return VerweisAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	if verweis == nil {
		return VerweisAntwort{Befunde: []string{}}
	}

	gestalt := alsVerweisDTO(kennung, *verweis)

	return VerweisAntwort{Verweis: &gestalt, Befunde: []string{}}
}

// KarteVerweisSetzen bezieht eine Karte auf einen Druck.
//
// **Der Druck wird geholt, bevor der Verweis entsteht** — die Regel steht in
// `service`, weil sie an allen vier Flächen gilt (§16). Hier steht nur die
// Übersetzung.
func (a *App) KarteVerweisSetzen(
	ctx context.Context, karteID, printing string,
) VerweisAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return VerweisAntwort{Befunde: []string{}, Fehler: befund}
	}

	quelle, fehlt := a.verweisquelle()
	if fehlt != nil {
		return VerweisAntwort{Befunde: []string{}, Fehler: fehlt}
	}

	verweis, befunde, fehler := quelle.Setzen(ctx, kennung, printing)
	if fehler != nil {
		return VerweisAntwort{Befunde: befundtexte(befunde), Fehler: Uebersetze(ctx, fehler)}
	}

	gestalt := alsVerweisDTO(kennung, *verweis)

	return VerweisAntwort{Verweis: &gestalt, Befunde: befundtexte(befunde)}
}

// KarteVerweisLoesen nimmt den Verweis einer Karte zurück.
func (a *App) KarteVerweisLoesen(ctx context.Context, karteID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	quelle, fehlt := a.verweisquelle()
	if fehlt != nil {
		return SchreibAntwort{Fehler: fehlt}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, quelle.Loesen(ctx, kennung))}
}

// SetVerweise gibt die Verweise aller Karten eines Sets.
//
// **In einem Zug**: die Kartenliste zeigt sie an jeder Zeile, und ein Set hat
// leicht dreihundert Karten.
func (a *App) SetVerweise(ctx context.Context, setID string) VerweiseAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return VerweiseAntwort{Verweise: []KartenverweisDTO{}, Fehler: befund}
	}

	quelle, fehlt := a.verweisquelle()
	if fehlt != nil {
		return VerweiseAntwort{Verweise: []KartenverweisDTO{}, Fehler: fehlt}
	}

	verweise, fehler := quelle.JeSet(ctx, kennung)
	if fehler != nil {
		return VerweiseAntwort{Verweise: []KartenverweisDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Nach Kartenkennung sortiert.** Eine Map gibt ihre Schlüssel in
	// wechselnder Reihenfolge; eine Liste, die bei jedem Aufruf anders
	// aussieht, lässt die Oberfläche neu zeichnen, wo sich nichts geändert
	// hat — und ist in einem Test nicht prüfbar.
	gestalt := make([]KartenverweisDTO, 0, len(verweise))
	for karteID, verweis := range verweise {
		gestalt = append(gestalt, alsVerweisDTO(karteID, verweis))
	}

	sortiereVerweise(gestalt)

	return VerweiseAntwort{Verweise: gestalt}
}

// sortiereVerweise ordnet nach Kartenkennung — numerisch, nicht als Text.
//
// **Numerisch, weil die Kennung eine Zahl ist**: als Text käme 10 vor 2, und
// die Kartenliste stünde in einer Reihenfolge, die niemand erwartet.
func sortiereVerweise(gestalt []KartenverweisDTO) {
	sort.Slice(gestalt, func(links, rechts int) bool {
		return kennungZahl(gestalt[links].KarteID) < kennungZahl(gestalt[rechts].KarteID)
	})
}

// kennungZahl liest eine Kennung als Zahl zurück.
//
// **Ohne Fehlerweg**: die Zeichenkette entstand ein paar Zeilen darüber aus
// einer `int64`. Ein Fehlerzweig hier wäre einer, den kein Test je beträte.
func kennungZahl(kennung string) int64 {
	zahl, _ := strconv.ParseInt(kennung, 10, 64) //nolint:errcheck // selbst erzeugt, siehe oben

	return zahl
}
