package app

// **Ein fremder Dienst, der nicht antwortet** — D02
// (#2233).
//
// Je Fall aus `domain/gegenstelle.go` ein Schlüssel, dessen Katalogsatz den
// nächsten Schritt nennt; dazu der Wirt und, wo eine Frist griff, ihre
// Sekunden. **Der Grund — die Meldung des Transports — geht ins Protokoll**
// (über [domain.MitGrund]), nicht in den Satz.

import (
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gegenstellenfaelle steht in der Reihenfolge, in der gefragt wird — die
// Proxyanmeldung zuerst, denn sie ist eine Gegenstelle **und** sagt mehr.
var gegenstellenfaelle = []struct {
	klasse     error
	schluessel domain.Schluessel
}{
	{domain.ErrProxyAnmeldung, domain.SchluesselNetzProxyAnmeldung},
	{domain.ErrNameUnaufloesbar, domain.SchluesselNetzName},
	{domain.ErrNichtErreichbar, domain.SchluesselNetzNichtErreichbar},
	{domain.ErrKeineAntwort, domain.SchluesselNetzKeineAntwort},
	{domain.ErrAntwortStockt, domain.SchluesselNetzAntwortStockt},
	{domain.ErrVerbindungAbgerissen, domain.SchluesselNetzAbgerissen},

	// **Ein 429 ist kein Ausfall, aber derselbe Rat von außen**: nichts am
	// Bestand ist falsch, und später geht es wieder.
	{domain.ErrGedrosselt, domain.SchluesselGedrosselt},
}

// gegenstelle übersetzt einen der Fälle — oder gibt nichts.
func gegenstelle(fehler error) *Fehler {
	for _, fall := range gegenstellenfaelle {
		if !errors.Is(fehler, fall.klasse) {
			continue
		}

		uebersetzt := mitSatz(CodeNetz, fall.schluessel, fehler)
		uebersetzt.Werte = map[string]string{}

		var genauer *domain.Gegenstellenfehler
		if errors.As(fehler, &genauer) {
			uebersetzt.Werte["wirt"] = genauer.Wirt

			if genauer.Frist > 0 {
				uebersetzt.Werte["sekunden"] = strconv.FormatFloat(genauer.Frist.Seconds(), 'f', -1, 64)
			}

			// **Mit Ende die Uhrzeit**, in der Zeit dieses Rechners (R01,
			// #2333): der Satz ohne sie hieße nur „warte einen Moment".
			if !genauer.Bis.IsZero() {
				uebersetzt.Schluessel = domain.SchluesselGedrosseltBis
				uebersetzt.Werte["uhrzeit"] = genauer.Bis.Local().Format("15:04")
			}
		}

		return uebersetzt
	}

	return nil
}
