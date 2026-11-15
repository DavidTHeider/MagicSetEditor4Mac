package app

// **Der Konflikt, der kein überholter Stand ist** — ENG-05.AK-03
// (#1730),
// nativer Befund F08.
//
// # Warum er einen eigenen Schlüssel braucht
//
// **Weil `konflikt` „lade neu" sagt**, und das stimmt nur für einen Stand,
// der sich seit dem Lesen geändert hat. Gezählt am 2026-09-23: 37 Stellen in
// `storage` und 82 in `service` melden einen Konflikt, und nur sieben davon
// sind ein überholter Stand. Die übrigen sagen etwas anderes — „der letzte
// Nutzer bleibt", „es liegen nur 2 Stück da", „der Nachweis ist schon
// eingelöst" —, und das Fenster zeigte über jedem denselben falschen Satz.
//
// **Dieselbe Gestalt wie [validierungsgrund]**: der Satz aus dem Fachkern
// reist unter `grund` mit, der Katalog setzt ihn in seinen Rahmen.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// konfliktgrund gibt den Konflikt, der seinen Grund nennt.
func konfliktgrund(fehler error) *Fehler {
	return &Fehler{
		Code:       CodeKonflikt,
		Schluessel: domain.SchluesselKonfliktgrund,
		Meldung:    fehler.Error(),
		Werte: map[string]string{
			"grund": strings.ReplaceAll(fehler.Error(), domain.ErrConflict.Error()+": ", ""),
		},
	}
}
