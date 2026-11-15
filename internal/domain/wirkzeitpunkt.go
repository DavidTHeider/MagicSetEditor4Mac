// **Wann eine Einstellung wirkt** — UX-04.AK-02
// (#1676).
//
// # Warum das hier steht und nicht im Fenster
//
// **Weil es eine Eigenschaft der Einstellung ist und keine der Fläche.** Ob
// der Proxy sofort wirkt, hängt daran, wie `internal/netz` ihn liest — nicht
// daran, ob jemand ihn im Fenster, an der Kommandozeile, über REST oder über
// MCP umlegt. Eine Antwort je Fläche wären vier Antworten auf dieselbe Frage
// (§16).
//
// # Zwei Werte und nicht drei
//
// **Es geht um die Frage, ob jemand etwas tun muss.** „Wirkt beim nächsten
// Aufruf" heißt für den Menschen davor dasselbe wie „sofort"; „beim nächsten
// Start" heißt: neu öffnen.

package domain

// Die zwei Zeitpunkte.
const (
	// WirkungSofort heißt: das Umlegen genügt.
	WirkungSofort = "sofort"

	// WirkungNeustart heißt: es gilt erst, wenn die Anwendung neu öffnet.
	WirkungNeustart = "neustart"
)

// EinstellungsWirkung sagt, wann ein Schlüssel wirkt — leer für einen
// unbekannten.
//
// **Leer und nicht „sofort"**: ein Bestand von morgen kann einen Schlüssel
// tragen, den dieses Programm nicht kennt, und ihm einen Zeitpunkt
// anzudichten wäre eine Behauptung über etwas, das es hier nicht gibt (§10).
func EinstellungsWirkung(schluessel string) string {
	eine, _ := finde(schluessel)

	return eine.Wirkung
}
