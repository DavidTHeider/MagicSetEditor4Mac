package cardtrader

// **Der Zugang zu CardTrader** — Werkstück A4 von Phase 5.3
// (#1443).

import "strings"

// Zugang ist, was CardTrader zum Anmelden verlangt.
//
// **Eine Angabe, nicht zwei** — nachgelesen am 2026-09-11 (0095):
// `Authorization: Bearer <Marke>`, und die Marke steht in den
// Profileinstellungen des Kontos. Scrydex verlangt zwei Kopfzeilen, MTGJSON
// eine; die Unterschiede stehen je Dienst und nicht in einer gemeinsamen
// Vermutung.
//
// **Sie kommt aus dem Schlüsselbund und nicht aus dem Bestand** (0056): einen
// Bestand gibt man weiter.
type Zugang struct{ Marke string }

// Fehlt nennt, was fehlt — leer, wenn nichts fehlt.
//
// **Als Satz und nicht als `bool`**: eine Meldung „der Zugang ist
// unvollständig" sagt niemandem, was einzutragen ist.
func (z Zugang) Fehlt() string {
	if strings.TrimSpace(z.Marke) == "" {
		return "Für CardTrader fehlt die Zugangsmarke. " +
			"Sie steht in den Profileinstellungen des CardTrader-Kontos " +
			"und gehört unter Einstellungen → Dienste."
	}

	return ""
}
