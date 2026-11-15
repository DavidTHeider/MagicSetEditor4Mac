// **Die Übersetzung zwischen Preisstand und DTO**
// (#1165).
//
// # Drei Angaben reisen mit jedem Preis
//
// **Quelle, Zeitpunkt, Alter.** Die Quelle sagt über ihre Preise ausdrücklich:
// „Absolutely no guarantee is made for any price information" (nachgelesen am
// 2026-08-27). Eine Zahl ohne diese drei wäre hier eine Behauptung, für die
// niemand einsteht — und §16, Satz 6 verlangt, dass nichts still übernommen
// wird.

package app

import (
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// jetzt gibt die Uhr dieser Anwendung.
//
// **[NeueApp] setzt sie, und trotzdem gibt es den Rückfall.** Ein `App{}` als
// Nullwert kommt in Läufen vor, und ein `nil`-Aufruf darüber wäre ein Absturz
// statt einer Antwort — genau die Klasse, an der in Phase 3.3 eine
// Kartenvorschau hing. Wer eine feste Uhr braucht, setzt sie über [App.MitUhr].
func (a *App) jetzt() time.Time {
	if a.uhr == nil {
		return time.Now()
	}

	return a.uhr()
}

// MitUhr setzt die Uhr — für Läufe, die ein Alter prüfen.
func (a *App) MitUhr(uhr func() time.Time) { a.uhr = uhr }

// zuPreisstandDTOs macht aus Preisständen ihre DTOs.
//
// **Das Alter kommt herein und wird nicht je Stand gerechnet.** Bei einer
// Antwort aus dem Netz ist es das der **Antwort** — alle Stände darin sind
// gleich alt, und der Zwischenspeicher weiß als Einziger, wie alt.
func zuPreisstandDTOs(staende []domain.Preisstand, alter time.Duration) []PreisstandDTO {
	dtos := make([]PreisstandDTO, 0, len(staende))

	for _, stand := range staende {
		dtos = append(dtos, PreisstandDTO{
			Printing:     stand.Printing,
			DruckID:      kennungtext(stand.DruckID),
			Quelle:       string(stand.Quelle),
			Markt:        string(stand.Markt),
			Art:          string(stand.Art),
			Waehrung:     stand.Waehrung,
			Finish:       string(stand.Finish),
			Zustand:      stand.Zustand,
			Sprache:      stand.Sprache,
			Lage:         string(stand.Lage),
			Wert:         wertOderNichts(stand),
			Beobachtet:   zeitpunkttext(stand.Beobachtet),
			Erhoben:      stand.Erhoben.UTC().Format(time.RFC3339),
			AlterStunden: inStunden(alter),
		})
	}

	return dtos
}

// zuPortfoliozeilen macht aus den Summen ihre DTOs.
func zuPortfoliozeilen(zeilen []service.Portfoliozeile, jetzt time.Time) []PortfoliozeileDTO {
	dtos := make([]PortfoliozeileDTO, 0, len(zeilen))

	for _, zeile := range zeilen {
		dtos = append(dtos, PortfoliozeileDTO{
			Waehrung:       zeile.Waehrung,
			Wert:           preistext(zeile.Wert, zeile.Waehrung),
			Exemplare:      zeile.Exemplare,
			AeltesterPreis: zeile.AeltesterPreis.UTC().Format(time.RFC3339),
			AlterStunden:   inStunden(jetzt.Sub(zeile.AeltesterPreis)),
			Quelle:         string(zeile.Quelle),
		})
	}

	return dtos
}

// alterSeit gibt das Alter des **ältesten** Standes einer Liste.
//
// **Des ältesten**, aus demselben Grund wie bei einer Portfoliosumme: eine
// Auskunft ist so frisch wie ihr ältester Bestandteil.
func alterSeit(jetzt time.Time, staende []domain.Preisstand) time.Duration {
	var aeltestes time.Duration

	for _, stand := range staende {
		if alter := stand.Alter(jetzt); alter > aeltestes {
			aeltestes = alter
		}
	}

	return aeltestes
}

// inStunden macht aus einer Dauer volle Stunden — nie negativ.
func inStunden(dauer time.Duration) int {
	if dauer < 0 {
		return 0
	}

	return int(dauer / time.Hour)
}

// wertOderNichts gibt den Preis — und nichts, wo keiner genannt ist.
//
// **Eine `"0.00"` sähe aus wie „umsonst"** (A3, #1442): bei einer Lage, die
// nicht `genannt` ist, steht null in der Zeile und bedeutet nichts.
func wertOderNichts(stand domain.Preisstand) string {
	if stand.Lage != domain.PreisGenannt {
		return ""
	}

	return preistext(stand.Wert, stand.Waehrung)
}

// zeitpunkttext gibt einen Zeitpunkt in RFC 3339 — leer für keinen.
func zeitpunkttext(zeitpunkt time.Time) string {
	if zeitpunkt.IsZero() {
		return ""
	}

	return zeitpunkt.UTC().Format(time.RFC3339)
}
