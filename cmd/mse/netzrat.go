package main

// **Der Rat zu einem fremden Dienst, der nicht antwortet** — D02
// (#2233).
//
// Die Sätze entsprechen denen des deutschen Katalogs; eine Kommandozeile hat
// keinen Katalog und spricht Deutsch wie jede Meldung dieses Programms (siehe
// [schreiber.klageUeber]).

import (
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// netzraete steht in der Reihenfolge, in der gefragt wird — die
// Proxyanmeldung zuerst, denn sie ist auch eine Gegenstelle.
var netzraete = []struct {
	klasse error
	rat    string
}{
	{domain.ErrProxyAnmeldung, "Gib die Zugangsdaten des Proxys über HTTPS_PROXY in der Umgebung an — " +
		"die Einstellung netz.proxy.https nimmt bewusst kein Kennwort."},
	{domain.ErrNameUnaufloesbar, "Prüfe die Internetverbindung und, falls einer eingerichtet ist, den Proxy."},
	{domain.ErrNichtErreichbar, "Läuft der Dienst, und stimmt seine Adresse? Sonst versuche es später noch einmal."},
	{domain.ErrKeineAntwort, "Versuche es später noch einmal."},
	{domain.ErrAntwortStockt, "Prüfe die Verbindung und versuche es noch einmal."},
	{domain.ErrVerbindungAbgerissen, "Versuche es noch einmal; bleibt es dabei, liegt es am Dienst oder an einem Proxy dazwischen."},
	{domain.ErrGedrosselt, "Warte einen Moment und versuche es dann noch einmal."},

	// **Und ein Gegenüber, das nicht horcht** (#2491) — kein fremder Dienst,
	// aber derselbe Rat von außen: nichts ist gelöscht, und eine Adresse
	// lässt sich prüfen.
	{domain.ErrPeerNichtErreichbar, "Horcht das Gegenüber, und stimmt seine Adresse (mse kontakt-adresse)? " +
		"Sonst versuche es später noch einmal."},
	{domain.ErrPeerOhneAdresse, "Trage die Adresse mit mse kontakt-adresse ein."},
}

// netzratFuer gibt den Rat zu einem der Fälle — oder nichts.
func netzratFuer(fehler error) string {
	for _, eintrag := range netzraete {
		if errors.Is(fehler, eintrag.klasse) {
			return eintrag.rat
		}
	}

	return ""
}
