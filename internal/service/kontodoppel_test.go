package service_test

// **Das Doppel der Kennungsablage** — Werkstück 6 von Phase 4.12
// (#1351).
//
// **Es steht eigens hier und nicht neben den Fällen**: die Testdatei hält die
// Verhaltensklassen zusammen, und ein Doppel darunter machte sie zu der einen
// Stelle, an der jede Änderung stattfindet (§2).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kennungsdoppel ist eine Ablage mit **einem** bekannten Nutzer.
//
// **Einer reicht**, weil die Suche nach der Kennung geht und nicht nach dem
// Bestand: was hier geprüft wird, ist die Entscheidung des Dienstes, nicht
// eine Abfrage — die steht in `internal/storage` und wird dort geprüft.
type kennungsdoppel struct {
	// bekannt ist der Nutzer, den [kennungsdoppel.NutzerZuKennung] findet —
	// ein Nullwert heißt: es gibt keinen.
	bekannt domain.Nutzer

	// fehlt lässt die Ablage antworten wie eine leere.
	fehlt bool

	// angelegt ist, was [kennungsdoppel.NutzerAnlegen] zu sehen bekam.
	angelegt domain.Nutzer

	// kaputt ist der Fehler, den beide Wege stattdessen geben.
	//
	// **Ein Fehler, der **kein** [domain.ErrNotFound] ist**: der Dienst
	// unterscheidet „gibt es nicht" von „ging schief", und ohne diesen Weg
	// bliebe die zweite Hälfte jeder Unterscheidung unbetreten.
	kaputt error
}

func (k *kennungsdoppel) NutzerZuKennung(_ context.Context, kennung string) (
	domain.Nutzer, error,
) {
	if k.kaputt != nil {
		return domain.Nutzer{}, k.kaputt
	}

	if !k.fehlt && (k.bekannt.Email == kennung || k.bekannt.Telefon == kennung) {
		return k.bekannt, nil
	}

	return domain.Nutzer{}, fmt.Errorf(
		"%w: zu dieser Kennung gibt es keinen Nutzer", domain.ErrNotFound)
}

// NutzerAnlegen **merkt sich, was es anlegt**.
//
// **Weil die Abnahme des Werkstücks über zwei Rufe geht**: eintragen, dann
// anmelden. Ein Doppel, das das Angelegte wegwirft, ließe den zweiten Ruf
// gegen einen leeren Bestand laufen — und der Test prüfte dann, dass eine
// unbekannte Kennung abgewiesen wird, mit dem Namen der Abnahme darüber.
func (k *kennungsdoppel) NutzerAnlegen(_ context.Context, neu domain.Nutzer) (
	domain.Nutzer, error,
) {
	if k.kaputt != nil {
		return domain.Nutzer{}, k.kaputt
	}

	k.angelegt = neu
	neu.ID = 42
	k.bekannt = neu

	return neu, nil
}
