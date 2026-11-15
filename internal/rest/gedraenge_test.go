package rest_test

import (
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Mehr als eine benannte Zahl von Verbindungen wird nicht angenommen**
// (Zweitblick über Phase U).
//
// # Der Befund
//
// Eine halb geschriebene Anfrage erreicht die Drossel **nie** — die zählt erst,
// wenn ein Behandler läuft. Ein lokaler Prozess konnte damit beliebig viele
// Verbindungen aufmachen und je zehn Sekunden halten; auf macOS ist die
// Vorgabe für offene Dateien 256 je Prozess, und dann nimmt der Dienst
// überhaupt nichts mehr an.
//
// **Die Frist allein reicht nicht.** Sie begrenzt, wie lange eine Verbindung
// hängt, nicht wie viele.
func TestMehrAlsDieGrenzeWirdNichtAngenommen(t *testing.T) {
	t.Parallel()

	roh, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	const hoechstens = 3

	lauscher := rest.HoechstensGleichzeitig(roh, hoechstens)
	t.Cleanup(func() { schliesse(t, lauscher) })

	// **Es wird wirklich angenommen, nicht nur verbunden.** Ein `Accept` im
	// Hintergrund holt jede Verbindung ab; ohne das läge sie im Rückstau des
	// Betriebssystems und der Test prüfte diesen statt die Grenze.
	angenommen := make(chan net.Conn, hoechstens+2)

	go func() {
		for {
			verbindung, err := lauscher.Accept()
			if err != nil {
				return
			}

			angenommen <- verbindung
		}
	}()

	adresse := lauscher.Addr().String()
	offen := make([]net.Conn, 0, hoechstens)

	for range hoechstens {
		offen = append(offen, waehle(t, adresse))
	}

	// **Die Serverseiten werden festgehalten**, denn an ihnen hängt der Platz.
	//
	// Der erste Anlauf schloss die **Klientenseite** und wartete darauf, dass
	// ein Platz frei wird — er wurde es nicht. Das ist kein Fehler des
	// Lauschers: solange der Server seine Seite offen hält, ist die Verbindung
	// offen, und der Platz gehört ihr zu Recht.
	serverseiten := make([]net.Conn, 0, hoechstens)

	for range hoechstens {
		select {
		case verbindung := <-angenommen:
			serverseiten = append(serverseiten, verbindung)
		case <-time.After(2 * time.Second):
			t.Fatal("eine der erlaubten Verbindungen wurde nicht angenommen")
		}
	}

	// **Die vierte darf nicht angenommen werden.**
	ueberzaehlig := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, ueberzaehlig) })

	select {
	case <-angenommen:
		t.Error("die vierte Verbindung wurde angenommen")
	case <-time.After(300 * time.Millisecond):
	}

	// **Und wenn eine geht, geht die nächste durch.** Ohne diesen Teil wäre
	// auch ein Lauscher grün, der nach der Grenze für immer zumacht.
	if err := serverseiten[0].Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	select {
	case <-angenommen:
	case <-time.After(2 * time.Second):
		t.Error("nach dem Schließen kam keine neue Verbindung durch")
	}

	// **Aufgeräumt wird ausdrücklich**, und ein Fehlschlag dabei ist einer:
	// eine Verbindung, die sich nicht schließen lässt, hielte ihren Platz.
	for _, verbindung := range append(offen, serverseiten[1:]...) {
		schliesse(t, verbindung)
	}
}

// waehle macht eine Verbindung auf.
func waehle(t *testing.T, adresse string) net.Conn {
	t.Helper()

	waehler := net.Dialer{Timeout: 2 * time.Second}

	verbindung, err := waehler.DialContext(t.Context(), "tcp", adresse)
	if err != nil {
		t.Fatalf("verbinden: %v", err)
	}

	return verbindung
}
