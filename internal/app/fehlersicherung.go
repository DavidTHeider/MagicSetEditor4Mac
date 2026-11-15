package app

// **Wie eine abgewiesene Sicherung übersetzt wird** — die drei Fälle, die
// mehr als einen Schlüssel mitbringen.
//
// Eigene Datei, weil `fehler.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **welcher Ausgang** ein Fehler ist,
// hier, **was eine Sicherung mitbringt**, die diese Fassung nicht öffnen kann.
//
// **Beide Fälle stehen in `Uebersetze` vor `ErrConflict` und `ErrValidation`**
// — der Grund steht bei [fremdeArchivfassung], und er ist gemessen: der erste
// Anlauf landete dort und zeigte den falschen Satz.
import (
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// istNeuereFassung sagt, ob eine Sicherung aus einer neueren Fassung stammt.
func istFremdeArchivfassung(fehler error) bool {
	var fassung *domain.FremdeArchivfassung

	return errors.As(fehler, &fassung)
}

// fremdeArchivfassung übersetzt die Abweisung samt ihren Teilen.
//
// **Sie steht vor `ErrConflict` und `ErrValidation`** — und zwar, weil sie
// sonst dort landete. Genau das ist am 2026-08-26 passiert: der erste Anlauf
// umschloss den Fehler mit `domain.ErrConflict`, und im Fenster stand „Der
// Eintrag wurde zwischenzeitlich geändert. Bitte neu laden." Der Go-Test war
// grün, denn er las die **Meldung** und nicht den **Schlüssel**; das Fenster
// bevorzugt den Katalog (§3, „zwischen den Prüfebenen").
func fremdeArchivfassung(fehler error) *Fehler {
	var fassung *domain.FremdeArchivfassung

	// Die Wache steht in `istFremdeArchivfassung`; hier wird nur ausgepackt.
	_ = errors.As(fehler, &fassung)

	return &Fehler{
		Code:       CodeValidierung,
		Schluessel: domain.SchluesselSicherungFremdeFassung,
		Meldung:    fassung.Error(),
		Werte: map[string]string{
			"dort": fassung.Dort,
			"hier": fassung.Hier,
		},
	}
}

func istNeuereFassung(fehler error) bool {
	var fassung *domain.NeuereFassung

	return errors.As(fehler, &fassung)
}

// neuereFassung übersetzt die Abweisung samt ihren Teilen.
//
// **Die Namen der Migrationen reisen als Werte mit**, nicht im Satz: so kann
// eine englische Sitzung den Rahmen übersetzen und die Bezeichner stehen
// lassen (§9).
func neuereFassung(fehler error) *Fehler {
	var fassung *domain.NeuereFassung

	// Die Wache steht in `istNeuereFassung`; hier wird nur ausgepackt.
	_ = errors.As(fehler, &fassung)

	return &Fehler{
		Code:       CodeValidierung,
		Schluessel: domain.SchluesselSicherungNeuereFassung,
		Meldung:    fassung.Error(),
		Werte: map[string]string{
			"unbekannt": strings.Join(fassung.Unbekannt, ", "),
			"stand":     fassung.Stand,
		},
	}
}

func istAeltereFassung(fehler error) bool {
	var fassung *domain.AeltereFassung

	return errors.As(fehler, &fassung)
}

// aeltereFassung übersetzt die Abweisung einer Sicherung von vor der
// Grundlinie (Vermerk 0189). **Mit der Grundlinie als Wert** und ohne die
// fehlenden Schritte: einer Sicherung aus 0.1.0 fehlen 96, und der Rat — über
// 0.7.0 — hängt an keinem davon.
func aeltereFassung(fehler error) *Fehler {
	var fassung *domain.AeltereFassung

	// Die Wache steht in `istAeltereFassung`; hier wird nur ausgepackt.
	_ = errors.As(fehler, &fassung)

	return &Fehler{
		Code:       CodeValidierung,
		Schluessel: domain.SchluesselSicherungAeltereFassung,
		Meldung:    fassung.Error(),
		Werte:      map[string]string{"grundlinie": fassung.Grundlinie},
	}
}
