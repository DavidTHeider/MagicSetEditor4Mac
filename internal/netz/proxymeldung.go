package netz

// **Ein Proxy, der scheitert, sagt, dass es ein Proxy war** — Werkstück N5
// von Phase 5.4
// (#1571).
//
// # Der Anlass
//
// **Er steht seit A18 als offener Punkt**
// (#1533)
// und ist an der Oberfläche nie angekommen: hinter einem Firmenproxy bekam
// jemand „connection refused" und keinen Hinweis darauf, dass die Verbindung
// gar nicht zur Gegenstelle ging. §8 nennt das die stille Form — der Satz
// schickt jemanden in die falsche Richtung.
//
// # Was hier **nicht** geschieht
//
// **Es wird nicht geraten, ob der Proxy schuld war.** Ein Transportfehler
// nennt die Adresse, mit der Go gesprochen hat, und das ist bei gesetztem
// Proxy dessen Adresse — mehr als „ein Proxy stand dazwischen" lässt sich
// daraus nicht ableiten, und mehr steht deshalb auch nicht da.

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitProxyhinweis hängt an einen Transportfehler, welcher Proxy im Spiel war.
//
// **Nur an einen Transportfehler.** Ein 404 ist eine Antwort und kein
// Verbindungsproblem; ein Hinweis auf den Proxy stünde dort im Weg.
// **Der Aufrufer hat schon geprüft, dass es einen Fehler gibt.** Ein
// `if fehler == nil` hier wäre ein Zweig, den kein Ruf betritt (§3) — beide
// Stellen rufen aus ihrem eigenen Fehlerzweig heraus.
func (k *Klient) mitProxyhinweis(anfrage *http.Request, fehler error) error {
	// **Der Stand wird hier noch einmal gerechnet** — derselbe Weg wie beim
	// Transport, und er muss es sein: die Auswahl entsteht je Anfrage, und
	// ein festgehaltener Stand nennte womöglich einen Proxy, der beim
	// Fehlschlag gar nicht mehr galt.
	ctx := context.Background()
	if anfrage != nil {
		ctx = anfrage.Context()
	}

	stand := StandMit(ctx, k.proxy)
	if !stand.Benutzt() {
		return fehler
	}

	durch, herkunft := stand.HTTPS.Wert, stand.HTTPS.Herkunft

	if anfrage != nil && anfrage.URL != nil && anfrage.URL.Scheme == "http" {
		durch, herkunft = stand.HTTP.Wert, stand.HTTP.Herkunft
	}

	// **Der leere Fall ist echt**: wer nur `netz.proxy.http` setzt, hat für
	// `https://` keinen — und dann stand keiner dazwischen.
	if strings.TrimSpace(durch) == "" {
		return fehler
	}

	// **Ohne Kennwort** (Nachlauf über Phase 6.7, N3): die Umgebung darf
	// Zugangsdaten tragen, und diese Meldung geht ins Protokoll und in die
	// Diagnose — dieselbe Form wie die Anzeige der Einstellung.
	return fmt.Errorf("%w (über den Proxy %s aus %s)", fehler,
		domain.ProxyadresseOhneKennwort(durch), herkunft)
}
