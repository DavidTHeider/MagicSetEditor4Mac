package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Zusatzwertpflege ist, was `app` zum Lesen und Setzen der Zusatzfelder einer
// Karte braucht
// (#2298).
type Zusatzwertpflege interface {
	Felder(ctx context.Context, setID, karteID int64) ([]service.Zusatzfeldwert, error)
	Setzen(ctx context.Context, setID, karteID int64, feld, wert string) (service.Zusatzvorwert, error)
}

// ZusatzfeldDTO ist ein Zusatzfeld des Stils einer Karte samt ihrem Wert.
//
// **Die Form ist die eines Kartenfelds** ([FeldbeschreibungDTO]) — das
// Formular zeichnet sie mit denselben Bauteilen, und MSE2 bearbeitet sie im
// selben Editor wie die Felder der Karte.
type ZusatzfeldDTO struct {
	Form        FeldbeschreibungDTO `json:"form"`
	Wert        string              `json:"wert"`
	Gespeichert bool                `json:"gespeichert"`
}

// ZusatzfelderAntwort trägt die Zusatzfelder einer Karte — oder einen Befund.
type ZusatzfelderAntwort struct {
	Felder []ZusatzfeldDTO `json:"felder"`
	Fehler *Fehler         `json:"fehler"`
}

// MitZusatzwertpflege setzt die Naht zu den Zusatzfeldern.
func (a *App) MitZusatzwertpflege(pflege Zusatzwertpflege) { a.zusatzwerte = pflege }

// Kartenzusatzfelder gibt die Zusatzfelder des Stils einer Karte, die sich von
// Hand setzen lassen.
func (a *App) Kartenzusatzfelder(ctx context.Context, setID, karteID string) ZusatzfelderAntwort {
	ctx = a.imProfil(ctx)

	setkennung, karte, befund := zwischenKennungen(setID, karteID)
	if befund != nil {
		return ZusatzfelderAntwort{Felder: []ZusatzfeldDTO{}, Fehler: befund}
	}

	gelesen, err := a.zusatzwerte.Felder(ctx, setkennung, karte)
	if err != nil {
		return ZusatzfelderAntwort{Felder: []ZusatzfeldDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return ZusatzfelderAntwort{Felder: zusatzfelderAus(gelesen, a.sprache)}
}

// zusatzfelderAus macht aus den Zusatzfeldern ihre DTOs — **eine Abbildung
// für alle Flächen**, das Fenster wie [Zusatzweg].
func zusatzfelderAus(gelesen []service.Zusatzfeldwert, sprache string) []ZusatzfeldDTO {
	felder := make([]domain.Field, 0, len(gelesen))
	for i := range gelesen {
		felder = append(felder, gelesen[i].Feld)
	}

	formen := beschreibungenAus(felder, sprache)
	antwort := make([]ZusatzfeldDTO, 0, len(gelesen))

	for i := range gelesen {
		antwort = append(antwort, ZusatzfeldDTO{
			Form: formen[i], Wert: gelesen[i].Wert, Gespeichert: gelesen[i].Gespeichert,
		})
	}

	return antwort
}

// ZusatzwertSetzen setzt den Wert eines Zusatzfelds an einer Karte.
func (a *App) ZusatzwertSetzen(ctx context.Context, setID, karteID, feld, wert string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, karte, befund := zwischenKennungen(setID, karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	// Den Vorwert braucht das Fenster nicht: es lädt nach, und die Rücknahme
	// geht über den Verlauf.
	_, err := a.zusatzwerte.Setzen(ctx, setkennung, karte, feld, wert)

	return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
}

// zwischenKennungen prüft Set und Karte — **das Set zuerst**, wie bei
// [App.FeldSetzen] (#808).
func zwischenKennungen(setID, karteID string) (setkennung, karte int64, befund *Fehler) {
	setkennung, befund = benannteKennungAus("setID", setID)
	if befund != nil {
		return 0, 0, befund
	}

	karte, befund = kennungAus(karteID)

	return setkennung, karte, befund
}
