//go:build netzlauf

// **Zwei Anwendungen finden einander über echtes Multicast** —
// NET-02.NAT-01
// (#1706).
//
// # Wofür es diesen Lauf gibt
//
// **Weil `docs/HAERTUNG.md` bisher ausdrücklich sagt, dass er fehlt**: der
// mDNS-Lauf geht über `lo0`, „aber beide Enden stehen auf **derselben**
// Maschine, und Multicast über die Rückschleife ist nicht Multicast über ein
// WLAN".
//
// **Hier stehen beide Enden auch auf derselben Maschine** — aber der Weg
// geht über die **gewählten** Schnittstellen, also über echtes Multicast im
// Netz. Das ist der Zwei-Maschinen-Fall bis auf die zweite Maschine: die
// Pakete verlassen den Rechner und kommen über die Multicast-Rückführung
// wieder an.
//
// # Warum hinter einem Tag
//
// **Weil er hinausgeht.** Er braucht ein aktives Netz und auf macOS die
// Erlaubnis *Lokales Netzwerk*; ohne sie gibt jeder Ruf `EHOSTUNREACH`. In
// der Pipeline hätte er nichts zu suchen (§3, §16) — was er belegt, steht in
// `docs/HAERTUNG.md`.
//
// # Was er **nicht** belegt
//
// **Dass ein zweiter Rechner sie findet.** Dafür gibt es keinen Ersatz; es
// bleibt eine Abnahme von Hand.
//
//	go test -tags netzlauf ./internal/entdeckung/ -run ZweiInstanzen -v
package entdeckung_test

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func TestZweiInstanzenFindenEinander(t *testing.T) {
	taugliche, fehler := entdeckung.Taugliche()
	if fehler != nil {
		t.Fatalf("die Schnittstellen lesen: %v", fehler)
	}

	if len(taugliche) == 0 {
		t.Skip("dieser Rechner hat keine taugliche Schnittstelle — " +
			"ohne Netz gibt es nichts zu finden")
	}

	for _, schnitt := range taugliche {
		t.Logf("taugliche Schnittstelle: %s", schnitt.Name)
	}

	// **Die erste Anwendung kündigt sich an** — über jede taugliche, wie im
	// Betrieb.
	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("eine Peeridentität erzeugen: %v", fehler)
	}

	anmeldung, fehler := entdeckung.Melde(entdeckung.Angebot{
		Kennung: ich.Kennung(), Name: "Erste Anwendung", Port: 47110,
	})
	if fehler != nil {
		t.Fatalf("ankündigen: %v — fehlt die Erlaubnis Lokales Netzwerk?",
			fehler)
	}

	t.Cleanup(func() {
		if beim := anmeldung.Beende(); beim != nil {
			t.Errorf("die Ankündigung zurücknehmen: %v", beim)
		}
	})

	// **Die zweite sucht** — ebenfalls über jede taugliche.
	//
	// **Eine kurze Frist reicht nicht.** mDNS antwortet, wenn es antwortet;
	// `Hoechstfrist` samt Nachlauf ist die Zusage, die dieses Paket gibt.
	ergebnis, fehler := entdeckung.Suche(context.Background(), entdeckung.Frage{
		Frist: 3 * time.Second,
	})
	if fehler != nil {
		t.Fatalf("suchen: %v — fehlt die Erlaubnis Lokales Netzwerk?", fehler)
	}

	for _, fund := range ergebnis.Funde {
		if fund.Kennung == ich.Kennung() {
			t.Logf("gefunden unter %s (%s)", fund.Adresse(), fund.Name)

			return
		}
	}

	t.Errorf("die eigene Ankündigung wurde nicht gefunden — %d Funde, "+
		"%d verworfen", len(ergebnis.Funde), ergebnis.Verworfen)
}
