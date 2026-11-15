// Das Doppel der Recherche — herausgezogen aus `recherche_test.go`, weil es
// dort über die Zeilengrenze aus §2 ging (A14, #1453).
//
// **Der Schnitt folgt der Frage**: dort stehen die Läufe über die
// Übersetzung, hier die Vorrichtung — und vor allem, **was sie prüft**.
package app_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// recherchedoppel hält, was die Bindungen bekommen sollen.
type recherchedoppel struct {
	ergebnis domain.Rechercheergebnis
	paket    service.Kontextpaket

	suchfehler    error
	paketfehler   error
	gesehen       domain.Rechercheanfrage
	fortsetzung   string
	gesehenWunsch service.Kontextwunsch
}

func (r *recherchedoppel) Recherchieren(
	_ context.Context, anfrage domain.Rechercheanfrage, fortsetzung string,
) (domain.Rechercheergebnis, error) {
	r.gesehen, r.fortsetzung = anfrage, fortsetzung

	return r.ergebnis, r.suchfehler
}

// Zusammenstellen gibt das eingestellte Paket — **und weist einen Wunsch ab,
// den der echte Dienst abwiese**.
//
// # Warum das Doppel prüft
//
// **Weil es sonst großzügiger ist als der Betrieb** (§3). Es tat es nicht,
// und A13.2a fiel deshalb erst am Flächenlauf auf: die Einschätzung schickte
// **kein** Kontextbudget und **keine** Trefferzahl los, und der echte
// [service.Kontextdienst] weist beides ab. Am Doppel war es grün.
//
// **Geprüft wird, was `pruefeKontextwunsch` prüft** — dieselben drei Fragen,
// und mehr nicht: ein Doppel, das strenger wäre als der Dienst, verböte
// Aufrufe, die im Betrieb durchgehen.
func (r *recherchedoppel) Zusammenstellen(
	_ context.Context, wunsch service.Kontextwunsch,
) (service.Kontextpaket, error) {
	r.gesehenWunsch = wunsch

	if fehler := brauchbarerWunsch(wunsch); fehler != nil {
		return service.Kontextpaket{}, fehler
	}

	return r.paket, r.paketfehler
}

// brauchbarerWunsch stellt die Fragen des Kontextdienstes.
func brauchbarerWunsch(wunsch service.Kontextwunsch) error {
	if strings.TrimSpace(wunsch.Aufgabe) == "" {
		return fmt.Errorf("%w: ein Kontextpaket braucht seine Aufgabe",
			domain.ErrValidation)
	}

	if strings.TrimSpace(wunsch.Zielmodell) == "" ||
		strings.TrimSpace(wunsch.Zielanbieter) == "" {
		return fmt.Errorf("%w: ein Kontextpaket wird für ein Modell budgetiert",
			domain.ErrValidation)
	}

	if wunsch.Kontextgrenze <= 0 ||
		wunsch.Ausgabereserve >= wunsch.Kontextgrenze {
		return fmt.Errorf(
			"%w: das Kontextfenster ist %d und die Ausgabereserve %d",
			domain.ErrValidation, wunsch.Kontextgrenze, wunsch.Ausgabereserve)
	}

	return wunsch.Anfrage.Pruefe()
}
