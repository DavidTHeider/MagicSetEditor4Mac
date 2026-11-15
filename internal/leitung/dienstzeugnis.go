package leitung

// **Ein Zeugnis für einen gewöhnlichen HTTPS-Dienst** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Warum es hier steht und nicht bei `internal/rest`
//
// **Weil es dasselbe Zeugnis ist wie das der Leitung** — selbst
// unterschrieben, auf der Peeridentität, mit demselben Bauweg
// ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md):
// „kein zweiter Krypto-Baukasten"). Ein zweiter Bauweg in `rest` wäre eine
// zweite Fassung derselben Frage, und die eine liefe bei der ersten Änderung
// der anderen davon.
//
// # Warum es **nicht** [serverwerk] ist
//
// **Weil auf der anderen Seite ein Browser oder ein Skript sitzt und kein
// Peer.** `serverwerk` verlangt von jedem Gegenüber ein Zeugnis
// (`RequireAnyClientCert`) — das ist richtig für einen Handschlag zwischen
// zwei Programmen dieser Art und unmöglich für `curl`.
//
// **Was an seine Stelle tritt, ist der Zugang**: jede Anfrage trägt eine
// Marke oder ein Token, und ohne eines kommt keine durch. Die Gegenseite
// prüft **uns** über den Fingerabdruck, den das Fenster zeigt; wir prüfen
// **sie** über den Zugang. Das ist die Aufteilung, die 0115 beschreibt.

import (
	"crypto/tls"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Serverzeugnis baut die TLS-Einstellung eines HTTPS-Dienstes.
//
// **Ohne Zeugnispflicht für die Gegenseite** — der Grund steht im Kopf.
func Serverzeugnis(ich peer.Identitaet) (*tls.Config, error) {
	zeugnis, fehler := zeugnisFuer(ich.Kennung(), ich.Oeffentlich(), ich.Signierer())
	if fehler != nil {
		return nil, fehler
	}

	return &tls.Config{
		Certificates: []tls.Certificate{zeugnis},

		// **TLS 1.3 und nichts darunter** — dieselbe Untergrenze wie bei der
		// Leitung. Wer ein älteres Gegenüber hat, hat ein Gegenüber, das
		// dieses Programm nicht bedient.
		MinVersion: tls.VersionTLS13,
	}, nil
}
