package service_test

// **Was ein Transport nicht sagen darf** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Der Anlass
//
// **Ein Kopf, der mehr versprach als sein Rumpf hielt** (§10). [Verbindung]
// hatte eine Methode `Gegenueber`, und daneben stand: „die Kennung am anderen
// Ende — **geprüft** durch den Handschlag, nicht behauptet durch den
// Adapter."
//
// **Beides war falsch.** Sie war behauptet: der Adapter bekam sie beim Bauen
// hereingereicht und gab sie zurück. Und sie war ungeprüft: **kein einziger
// produktiver Leser** las sie je — die Kennung, mit der `service` arbeitet,
// kommt aus [peer.Entgegennahme], also aus dem Handschlag.
//
// **Sichtbar wurde es am Horchen.** Eine angenommene Verbindung hat kein
// Gegenüber, bis der Handschlag eines nennt; der Adapter hätte eine leere
// Kennung erfinden müssen, um eine Schnittstelle zu erfüllen, die niemand
// liest.

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Verbindung trägt Bytes und keine Identität.**
//
// **Strukturell und nicht am Verhalten**, aus demselben Grund wie beim
// Netzwächter: „der Transport behauptet keine Identität" ist eine Aussage über
// **jeden** Adapter, auch den, den jemand morgen schreibt. Wer die Methode
// wieder einführte, müsste diesen Lauf ändern und sähe dabei, was er tut.
func TestEineVerbindungBehauptetKeinGegenueber(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf((*service.Verbindung)(nil)).Elem()

	for _, verboten := range []string{
		"Gegenueber", "Kennung", "Identitaet", "Peer",
	} {
		if _, hat := typ.MethodByName(verboten); hat {
			t.Errorf(
				"Verbindung hat %q — wer da ist, sagt der Handschlag",
				verboten)
		}
	}

	// **Und die Gegenprobe, dass hier überhaupt etwas gemessen wird**: die
	// Schnittstelle steht, und sie hat die Methoden, die sie haben soll.
	for _, gebraucht := range []string{
		"Kanal", "Sende", "Empfange", "Schliesse",
	} {
		if _, hat := typ.MethodByName(gebraucht); !hat {
			t.Errorf("Verbindung hat kein %q mehr", gebraucht)
		}
	}
}
