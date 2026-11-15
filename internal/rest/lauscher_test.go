package rest_test

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// zeitgrenze ist die Frist für einen Verbindungsversuch.
//
// **Kurz, weil ein Erfolg der Befund ist.** Wer hier abnimmt, tut es sofort;
// wer nicht da ist, lässt die Zeit verstreichen. Länger zu warten machte den
// Lauf langsamer, ohne etwas zu zeigen.
const zeitgrenze = 200 * time.Millisecond

// **Der Dienst bindet an die Rückschleife und nirgends sonst**
// (#997).
//
// # Warum das der erste Test dieses Pakets ist
//
// Es ist die eine Zusage, um derentwillen
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// überhaupt gegeben werden konnte. Fällt sie, ist der Dienst nicht mehr das,
// wofür die Entscheidung gefallen ist — er ist ein anderer.
//
// # Was er prüft, und was nicht
//
// Geprüft wird die **wirklich geöffnete** Adresse, nicht die gewünschte: der
// Lauscher wird aufgemacht und gefragt, wo er liegt. Eine Prüfung der
// Zeichenkette, die hineingeht, sagte nichts über den Fall, dass jemand sie
// unterwegs ersetzt.
//
// **Nicht geprüft** ist, dass kein anderer Weg im Prozess einen zweiten
// Lauscher aufmacht. Dafür steht der Wächter in `internal/schichten`: außer
// diesem Paket und `cmd/mse-rest` erreicht niemand `net/http`.
func TestDerLauscherLiegtAufDerRueckschleife(t *testing.T) {
	t.Parallel()

	lauscher, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	t.Cleanup(func() { schliesse(t, lauscher) })

	rechner, _, teilfehler := net.SplitHostPort(lauscher.Addr().String())
	if teilfehler != nil {
		t.Fatalf("Adresse zerlegen: %v", teilfehler)
	}

	adresse := net.ParseIP(rechner)
	if adresse == nil {
		t.Fatalf("keine Adresse: %q", rechner)
	}

	if !adresse.IsLoopback() {
		t.Errorf("der Lauscher liegt auf %s — das ist nicht die Rückschleife", adresse)
	}
}

// **Und er ist von außen nicht zu erreichen.**
//
// Die Zusicherung oben sagt, wohin gebunden wurde. Diese sagt, was daraus
// folgt: auf keiner anderen Adresse dieses Rechners nimmt jemand ab.
//
// **Was sie nicht prüft**: einen Zugriff von einem **anderen** Rechner. Dafür
// bräuchte es einen zweiten, und den gibt es hier nicht
// ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
func TestAufDenAnderenAdressenNimmtNiemandAb(t *testing.T) {
	t.Parallel()

	lauscher, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	t.Cleanup(func() { schliesse(t, lauscher) })

	_, port, teilfehler := net.SplitHostPort(lauscher.Addr().String())
	if teilfehler != nil {
		t.Fatalf("Adresse zerlegen: %v", teilfehler)
	}

	for _, adresse := range andereAdressen(t) {
		waehler := net.Dialer{Timeout: zeitgrenze}

		verbindung, err := waehler.DialContext(t.Context(), "tcp",
			net.JoinHostPort(adresse, port))
		if err == nil {
			schliesse(t, verbindung)
			t.Errorf("auf %s nimmt jemand ab", adresse)
		}
	}
}

// schliesse macht zu und meldet, wenn es nicht ging.
func schliesse(t *testing.T, was io.Closer) {
	t.Helper()

	if err := was.Close(); err != nil {
		t.Errorf("schließen: %v", err)
	}
}

// andereAdressen gibt die Adressen dieses Rechners, die nicht Rückschleife sind.
func andereAdressen(t *testing.T) []string {
	t.Helper()

	adressen, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("Schnittstellen: %v", err)
	}

	var fremde []string

	for _, eintrag := range adressen {
		netz, istNetz := eintrag.(*net.IPNet)
		if !istNetz || netz.IP.IsLoopback() || netz.IP.To4() == nil {
			continue
		}

		fremde = append(fremde, netz.IP.String())
	}

	if len(fremde) == 0 {
		t.Skip("dieser Rechner hat keine zweite IPv4-Adresse — nichts zu prüfen")
	}

	return fremde
}

// **Jenseits der Schleife gibt es keinen Lauscher ohne TLS.**
//
// # Was dieser Wächter vorher hielt
//
// **„Kein Lauscher außerhalb der Schleife."** Das war die Zusage, um
// derentwillen [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// gegeben werden konnte, und sie hat bis N6 gehalten.
//
// **[0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)
// hat sie fortgeschrieben**, und dieser Lauf mit ihr: die Bindung wird
// einstellbar, und was an ihre Stelle tritt, ist „kein Lauscher außerhalb der
// Schleife **ohne TLS**". Das Issue verlangt genau das — umschreiben statt
// entfernen.
//
// # Warum der Fehler **vor** dem Lauscher kommt
//
// **Weil es sonst einen Augenblick gäbe, in dem der Port offen und
// ungesichert ist.** Ein Lauscher, der aufgeht und danach abgewiesen wird,
// hat für die Dauer dieses Augenblicks genau das getan, was er nicht darf.
func TestJenseitsDerSchleifeGibtEsKeinenLauscherOhneTLS(t *testing.T) {
	t.Parallel()

	for _, wunsch := range []string{"0.0.0.0", "::", "192.168.1.1"} {
		t.Run(wunsch, func(t *testing.T) {
			t.Parallel()

			lauscher, fehler := rest.LauschenMit(t.Context(),
				domain.Bindung{Adresse: wunsch}, nil)
			if fehler == nil {
				schliesse(t, lauscher)
				t.Fatalf("%q wurde ohne TLS angenommen", wunsch)
			}

			if !strings.Contains(fehler.Error(), wunsch) {
				t.Errorf("die Meldung nennt die Adresse nicht: %v", fehler)
			}

			// **Und die Meldung sagt, warum** (§8): „abgewiesen" ohne den
			// Grund schickt jemanden auf die Suche nach einem Tippfehler.
			if !strings.Contains(fehler.Error(), "TLS") {
				t.Errorf("die Meldung nennt TLS nicht: %v", fehler)
			}
		})
	}
}

// **Ein Name wird abgewiesen — auch mit Zeugnis.**
//
// **Weil er zur Laufzeit aufgelöst würde**, und wohin, weiß dieses Programm
// nicht. `localhost` sähe aus wie die Schleife und ist es auf einem fremd
// eingerichteten Rechner nicht.
func TestEinNameWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, wunsch := range []string{"localhost", "mein.mac"} {
		t.Run(wunsch, func(t *testing.T) {
			t.Parallel()

			lauscher, fehler := rest.LauschenMit(t.Context(),
				domain.Bindung{Adresse: wunsch}, &tls.Config{MinVersion: tls.VersionTLS13})
			if fehler == nil {
				schliesse(t, lauscher)
				t.Fatalf("%q wurde angenommen", wunsch)
			}
		})
	}
}

// **Die Rückschleifenverwandten dürfen — ohne Zeugnis.**
//
// # Warum sie jetzt dürfen und vorher nicht
//
// **`::1` und `127.0.0.2` sind Rückschleife.** Der alte Wächter wies sie ab,
// weil der Produktivcode auf **Gleichheit** verglich — und das war richtig,
// solange es nur einen Weg geben sollte: zwei Wege wären zwei zu prüfen
// gewesen.
//
// **Seit die Frage lautet „verlässt das den Rechner?"**, ist die Antwort für
// beide dieselbe wie für `127.0.0.1`: nein. Sie über eine Zeichenkette
// abzuweisen wäre jetzt eine Strenge ohne Grund — und eine, die jemanden mit
// einem IPv6-Rechner ohne Not aussperrt.
func TestDieSchleifenverwandtenDuerfenOhneZeugnis(t *testing.T) {
	t.Parallel()

	// **`127.0.0.2` steht hier nicht, und das ist kein Nachlassen der
	// Zusage.** Die **Regel** nimmt sie — `domain.Bindung` sagt es, und der
	// Lauf dort prüft es. Was sie hier abweist, ist macOS: die Adresse ist
	// auf dem Schleifengerät nicht eingerichtet, und `bind` endet mit
	// „can't assign requested address". Gemessen am 2026-09-15. Eine Regel
	// über eine Maschineneigenschaft zu prüfen, prüfte die Maschine (§10).
	for _, wunsch := range []string{"127.0.0.1", "::1", ""} {
		t.Run(wunsch, func(t *testing.T) {
			t.Parallel()

			lauscher, fehler := rest.LauschenMit(t.Context(),
				domain.Bindung{Adresse: wunsch}, nil)
			if fehler != nil {
				t.Fatalf("%q wurde abgewiesen: %v", wunsch, fehler)
			}

			t.Cleanup(func() { schliesse(t, lauscher) })

			rechner, _, teilfehler := net.SplitHostPort(lauscher.Addr().String())
			if teilfehler != nil {
				t.Fatalf("Adresse zerlegen: %v", teilfehler)
			}

			if adresse := net.ParseIP(rechner); adresse == nil || !adresse.IsLoopback() {
				t.Errorf("der Lauscher liegt auf %q", rechner)
			}
		})
	}
}
