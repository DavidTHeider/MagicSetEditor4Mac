package app

import (
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kontingent übersetzt das eigene Budget eines Aufrufers — oder gibt nichts.
// Der Wert `sekunden` sagt, wann der nächste Ruf geht.
func kontingent(fehler error) *Fehler {
	var grenze *domain.Kontingentgrenze
	if !errors.As(fehler, &grenze) {
		return nil
	}

	uebersetzt := mitSatz(CodeDrossel, domain.SchluesselKontingent, fehler)
	uebersetzt.Werte = map[string]string{"sekunden": strconv.Itoa(grenze.Sekunden())}

	return uebersetzt
}
