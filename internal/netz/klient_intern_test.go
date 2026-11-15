package netz

import (
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

// **Eine Verbindung, an der keine Frist mehr zu setzen ist, liest nicht**
// (#1002).
//
// # Warum das ein interner Test ist
//
// `fristverbindung` ist nicht exportiert, und sie soll es nicht sein — sie ist
// eine Eigenschaft des Transports, keine Fähigkeit für Aufrufer. Der Zweig
// darin lässt sich von außen nicht betreten: der HTTP-Transport liest nicht
// nach dem Schließen.
//
// # Warum der Zweig trotzdem dasteht
//
// Weil `SetReadDeadline` scheitern **kann**, und ein Lesen mit einer Frist,
// die nicht gesetzt wurde, hätte gar keine — es wäre genau der unbeobachtete
// Rumpf, gegen den es die Frist gibt (§3: eine unbequeme Verzweigung ist kein
// Grund für eine Ausnahme).
func TestEineGeschlosseneVerbindungLiestNicht(t *testing.T) {
	t.Parallel()

	links, rechts := net.Pipe()

	if err := rechts.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if err := links.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	mitFrist := &fristverbindung{Conn: links, frist: time.Second}

	_, fehler := mitFrist.Read(make([]byte, 8))
	if fehler == nil {
		t.Fatal("das Lesen an einer geschlossenen Verbindung ging durch")
	}

	// **Welcher der beiden Wege scheitert, ist offen** — `SetReadDeadline` an
	// einer geschlossenen Röhre oder das `Read` danach. Beide sind der Fall,
	// gegen den dieser Test steht: es wird nicht ohne Frist gelesen.
	if !errors.Is(fehler, os.ErrClosed) &&
		!errors.Is(fehler, net.ErrClosed) &&
		!errors.Is(fehler, io.ErrClosedPipe) {
		t.Errorf("Fehler %v ist kein Schließfehler", fehler)
	}
}
