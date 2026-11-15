package leitung

// **Die Wege, die von außen nicht zu betreten sind** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
// # Warum diese Läufe **im** Paket stehen
//
// **Weil sie Fehlerwege der Kryptographie prüfen**, und die treten im Betrieb
// nicht auf: `crypto/rand` versagt auf diesem Betriebssystem nicht, ein
// Zeugnis aus einer echten [peer.Identitaet] ist nie ein krummes Paar, und
// `ExportKeyingMaterial` gibt nach einem geglückten TLS-1.3-Handschlag immer
// etwas her.
//
// **Und trotzdem stehen die Zweige da**, weil ihre Rufe einen Fehler
// zurückgeben. Ein `_ =` daneben wäre die stille Absenkung, die §3 verbietet;
// eine Ausnahme in `.covignore` wäre sie auch — dort steht nur, was ohne
// echtes Fenster, echten Prozess oder echte Plattform nicht läuft.
//
// **Also werden sie betreten.** Das kostet zwei Nähte: eine Zufallsquelle, die
// sich tauschen lässt, und ein `zeugnisFuer`, das seine drei Teile einzeln
// entgegennimmt.

import (
	"crypto"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// versiegt gibt nie ein Byte her.
type versiegt struct{}

func (versiegt) Read([]byte) (int, error) {
	return 0, errors.New("die Quelle ist zu")
}

// ohneZufall tauscht die Quelle für die Dauer eines Laufs.
//
// **Ohne `t.Parallel()` und mit `t.Cleanup`** — sie ist eine Veränderliche des
// Pakets, und zwei Läufe zugleich sähen einander dabei zu.
func ohneZufall(t *testing.T) {
	t.Helper()

	vorher := zufallsquelle
	zufallsquelle = versiegt{}

	t.Cleanup(func() { zufallsquelle = vorher })
}

// eineIdentitaet gibt ein frisches Paar.
func eineIdentitaet(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich
}

// **Ohne Zufall entsteht kein Zeugnis — und damit keine Leitung.**
func TestOhneZufallEntstehtKeinZeugnis(t *testing.T) {
	ohneZufall(t)

	ich := eineIdentitaet(t)

	if _, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), ich.Signierer(),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Und beide Wege darüber melden es weiter** — ein Zeugnis, das nicht
	// entsteht, ist das Ende, und nicht eine Verbindung ohne Ausweis.
	if _, fehler := Waehle(
		t.Context(), "127.0.0.1:1", ich, ich.Kennung(),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Waehle gab %v", fehler)
	}

	if _, fehler := Horche(
		t.Context(), "127.0.0.1:0", ich,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Horche gab %v", fehler)
	}

	// **Und der dritte Weg ebenso** — [Serverzeugnis] seit N6
	// (#1572).
	// Ohne diese Zeile stünde der Dienst ohne Zeugnis da und meldete es
	// nicht; 0115 verlangt das Gegenteil: jenseits der Schleife geht der
	// Port ohne Ausweis gar nicht erst auf.
	if _, fehler := Serverzeugnis(ich); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Serverzeugnis gab %v", fehler)
	}
}

// **Ein Signierer ohne brauchbaren Schlüssel bekommt kein Zeugnis.**
//
// # Was der erste Anlauf gemessen hat, und warum er es nicht tat
//
// **Er gab den öffentlichen Schlüssel des einen und den Signierer des
// anderen** — und `x509.CreateCertificate` kam **durch**. Es prüft nicht, ob
// die beiden zusammengehören; es schreibt den einen hinein und unterschreibt
// mit dem anderen. Das fällt erst beim Prüfen auf, und dort tut es das auch:
// `pruefeGegenueber` vergleicht den Schlüssel im Zeugnis mit der erwarteten
// Kennung.
//
// **Was wirklich scheitert**, ist ein Signierer, aus dessen `Public()` sich
// kein Unterschriftsverfahren ableiten lässt. Im Betrieb kommt er aus einer
// [peer.Identitaet] und ist immer Ed25519 — deshalb ist das der Weg, den nur
// ein Lauf betreten kann, der einen eigenen mitbringt.
func TestEinSigniererOhneBrauchbarenSchluesselBekommtKeinZeugnis(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	if _, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), krummerSignierer{},
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// krummerSignierer zeigt etwas vor, das kein Schlüssel ist.
type krummerSignierer struct{}

func (krummerSignierer) Public() crypto.PublicKey { return struct{}{} }

func (krummerSignierer) Sign(
	io.Reader, []byte, crypto.SignerOpts,
) ([]byte, error) {
	return nil, errors.New("hier unterschreibt niemand")
}

// **Vor dem Handschlag gibt es keinen Kanal.**
//
// **Und das ist ein Befund und keine leere Zeichenkette**: ein Kanal, den
// niemand ausgerechnet hat, bände den Handschlag an nichts.
func TestVorDemHandschlagGibtEsKeinenKanal(t *testing.T) {
	t.Parallel()

	einer, _ := net.Pipe()

	t.Cleanup(func() {
		_ = einer.Close() //nolint:errcheck // nur ein Lauf
	})

	//#nosec G402 -- hier wird nichts verbunden, nur der Stand gelesen
	stumm := tls.Client(einer, &tls.Config{
		InsecureSkipVerify: true, MinVersion: tls.VersionTLS13,
	})

	if _, fehler := leitungAus(stumm, ""); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine Gegenstelle ohne Zeugnis wird abgewiesen.**
//
// **Direkt und nicht über eine Verbindung**: TLS selbst weist einen Klienten
// ohne Zeugnis schon ab, bevor diese Prüfung gerufen wird — der Zweig steht
// trotzdem da, denn eine Prüfung, die bei null Zeugnissen ins Leere griffe,
// wäre ein Absturz an der Stelle, an der eine Ablehnung hingehört.
func TestEineGegenstelleOhneZeugnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := pruefeGegenueber("", new(peer.Kennung))(tls.ConnectionState{}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und ein Ed25519-Schlüssel falscher Länge ebenso.**
//
// **Ein zweiter Wächter für dieselbe Sache**, und beide sind gewollt: der
// erste fragt den Typ, der zweite die Länge. `crypto/x509` gibt nie einen
// kurzen heraus — aber die Prüfung steht hier und nicht dort, und eine, die
// sich auf den Nachbarn verließe, wäre eine Zusage über fremden Code.
func TestEinEd25519SchluesselFalscherLaengeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	zeugnis := &x509.Certificate{
		PublicKey: ed25519.PublicKey(make([]byte, 16)),
	}

	fehler := pruefeGegenueber("", new(peer.Kennung))(tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{zeugnis},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}
