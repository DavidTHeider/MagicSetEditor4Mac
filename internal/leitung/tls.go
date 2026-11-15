package leitung

// **Die Leitung über TCP** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum hier TLS steht und nicht blankes TCP
//
// **Weil unterschrieben nicht verschlossen heißt.** `internal/peer`
// unterschreibt jede Hülle und verschlüsselt keine — ein Mitleser auf der
// Strecke kann nichts fälschen, aber alles lesen, und was hier hinausgeht,
// sind Kartenbestände, Namen und Lagerorte. libp2p hätte Noise oder TLS
// mitgebracht; seit [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md)
// kommt libp2p nicht, und damit ist das hier eigene Arbeit.
//
// # Wie hier festgenagelt wird
//
// **Ohne Zertifizierungsstelle, und das ist der Punkt.** Das Zeugnis ist
// selbst unterschrieben, und sein Schlüssel **ist** die Peeridentität. Wer
// wählt, nennt die erwartete Kennung; geprüft wird, ob der Schlüssel im
// Zeugnis dieselbe ergibt. Eine Zertifizierungsstelle beantwortete eine andere
// Frage — „gehört diese Adresse jemandem" statt „ist das der, den ein Mensch
// bestätigt hat".
//
// **Dass die Gegenseite den Schlüssel wirklich hat, sagt TLS selbst.** Ein
// Zeugnis vorzuzeigen genügt nicht: die Gegenseite unterschreibt im
// Handschlag, und ohne den privaten Schlüssel kommt sie nicht durch. Das ist
// der Grund, warum ein Vergleich des Schlüssels hier ausreicht.
//
// # Und warum der Kanal von hier kommt
//
// **Weil TLS ihn besser kann als eine Adresse.** Der Handschlag aus N1 bindet
// sich an eine Zeichenkette je Verbindung; aus dem Vierertupel abgeleitet wäre
// sie eindeutig und für jeden auf der Strecke lesbar. `ExportKeyingMaterial`
// (RFC 5705) gibt stattdessen Schlüsselmaterial dieser einen Verbindung:
// beide Seiten rechnen dasselbe, niemand sonst kann es.

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// zufallsquelle ist, woher ein Zeugnis seine Zufallszahlen nimmt.
//
// # Warum eine Veränderliche und kein Argument
//
// **Weil der Fehlerweg sonst unbetretbar wäre.** [Waehle] und [Horche] bauen
// das Zeugnis selbst; ein Argument dafür müsste durch ihre Signatur, und der
// einzige Aufrufer im Betrieb — der Composition Root — hätte dann einen
// Zweig, den **er** nicht betreten kann. Der Fehler wanderte nur eine Schicht
// weiter (§3).
//
// **Sie wird ausschließlich von Läufen getauscht**, und zwar ohne
// `t.Parallel()` und mit `t.Cleanup`.
var zufallsquelle io.Reader = rand.Reader

// Kanalbereich trennt diese Ausfuhr von jeder anderen.
//
// **RFC 5705 verlangt eine Marke**, und sie gehört diesem Programm: ein
// anderes, das dieselbe Verbindung anders benutzte, bekäme so anderes
// Material.
const Kanalbereich = "mse4mac/kanal/1"

// Kanallaenge ist, wie viele Bytes ausgeführt werden.
const Kanallaenge = 32

// Handschlagfrist ist, wie lange ein hereinkommender Handschlag dauern darf.
//
// **Ohne sie hielte ein Gegenüber, das schweigt, den Horcher fest** — es
// genügte, die Verbindung zu öffnen und nichts zu schicken.
const Handschlagfrist = 20 * time.Second

// Wahlfrist ist, wie lange eine Wahl höchstens dauert.
//
// **Sie steht hier, weil die Abnahme sie verlangt**: „eine unerreichbare
// Adresse endet mit einem benannten Fehler und nicht mit Warten." Ein
// Aufrufer mit eigener Frist bekommt die kürzere von beiden.
const Wahlfrist = 20 * time.Second

// Waehle baut eine gesicherte Leitung zu einer Adresse.
//
// **Die erwartete Kennung ist keine Zierde.** Ohne sie redete dieser Weg mit
// jedem, der an der Adresse sitzt — und eine Adresse sagt nichts darüber, wer
// das ist.
func Waehle(
	ctx context.Context, adresse string,
	ich peer.Identitaet, erwartet peer.Kennung,
) (*Leitung, error) {
	if erwartet == "" {
		return nil, fmt.Errorf(
			"%w: ohne erwartete Kennung wird nicht gewählt",
			domain.ErrValidation)
	}

	zeugnis, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), ich.Signierer())
	if fehler != nil {
		return nil, fehler
	}

	ctx, abbrechen := context.WithTimeout(ctx, Wahlfrist)
	defer abbrechen()

	strom, fehler := (&net.Dialer{}).DialContext(ctx, "tcp", adresse)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrPeerNichtErreichbar, fehler,
			"die Gegenstelle ist nicht erreichbar")
	}

	var gegenstelle peer.Kennung

	gesichert := tls.Client(strom, &tls.Config{
		Certificates: []tls.Certificate{zeugnis},

		// **Nicht „unsicher", sondern „anders geprüft".** Gos eigene Prüfung
		// fragte nach einer Zertifizierungsstelle und nach einem Rechnernamen;
		// beides gibt es hier nicht. Geprüft wird darunter, und zwar gegen
		// die Kennung, die ein Mensch bestätigt hat.
		InsecureSkipVerify: true, //#nosec G402 -- siehe VerifyConnection

		VerifyConnection: pruefeGegenueber(erwartet, &gegenstelle),
		MinVersion:       tls.VersionTLS13,
	})

	return schliesseHandschlagAb(ctx, gesichert, strom, &gegenstelle)
}

// Horcher nimmt gesicherte Verbindungen entgegen.
//
// **Er entsteht nur auf Aufruf.** Das Issue verlangt es: „nur nach bewusster
// Auswahl … sichtbar werden; keine unbemerkte Dauerwerbung aller Profile."
type Horcher struct {
	lauscher net.Listener
	zeugnis  tls.Certificate
}

// Horche öffnet einen Horcher an einer Adresse.
func Horche(
	ctx context.Context, adresse string, ich peer.Identitaet,
) (*Horcher, error) {
	zeugnis, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), ich.Signierer())
	if fehler != nil {
		return nil, fehler
	}

	lauscher, fehler := (&net.ListenConfig{}).Listen(ctx, "tcp", adresse)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"an dieser Adresse lässt sich nicht horchen")
	}

	return &Horcher{lauscher: lauscher, zeugnis: zeugnis}, nil
}

// Adresse ist, wo dieser Horcher steht.
//
// **Sie wird gelesen und nicht wiederholt**: wer „127.0.0.1:0" angibt, bekommt
// vom Betriebssystem einen Port, und nur der Lauscher weiß welchen.
func (h *Horcher) Adresse() string { return h.lauscher.Addr().String() }

// Nimm wartet auf die nächste Verbindung.
//
// # Warum hier keine erwartete Kennung steht
//
// **Weil niemand weiß, wer anklopft.** Wer horcht, erfährt es aus dem
// Handschlag — und ob er eintreten darf, entscheidet [service.Peersitzungen]
// anhand der Kontakte. Ein Horcher, der eine Kennung erwartete, könnte immer
// nur auf einen warten.
//
// # Was dieser Ruf zusagt und was nicht
//
// **Der Kontext begrenzt den Handschlag, nicht das Warten.** `Accept` kennt
// keinen; wer aufhören will, ruft [Horcher.Schliesse] — dann endet dieser Ruf
// mit einem Fehler. Das ist die übliche Gestalt in Go, und sie steht hier,
// damit niemand vom Kontext etwas erwartet, das er nicht leistet.
//
// **Und er ist der Reihe nach.** Ein Gegenüber, das die Verbindung öffnet und
// dann schweigt, hält den nächsten auf — höchstens aber [Handschlagfrist]
// lang. Wer das nicht will, ruft `Nimm` in einer Schleife und führt jeden
// Handschlag nebenher; diese Entscheidung gehört dem Aufrufer und nicht dem
// Blatt.
func (h *Horcher) Nimm(ctx context.Context) (*Leitung, error) {
	strom, fehler := h.lauscher.Accept()
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrNotFound, fehler,
			"es kam keine Verbindung herein")
	}

	// **Derselbe Weg wie über ein fremdes Netz** — die Einstellung steht in
	// [serverwerk], und wer sie ändert, ändert sie für beide. Eine zweite
	// Fassung hier wäre die Stelle, an der ein Weg geprüft ist und der
	// andere nicht.
	ctx, abbrechen := context.WithTimeout(ctx, Handschlagfrist)
	defer abbrechen()

	var gegenstelle peer.Kennung

	return schliesseHandschlagAb(
		ctx, tls.Server(strom, serverwerk(h.zeugnis, &gegenstelle)), strom, &gegenstelle)
}

// Schliesse legt den Horcher nieder.
func (h *Horcher) Schliesse() error {
	if fehler := h.lauscher.Close(); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"den Horcher niederlegen")
	}

	return nil
}

// schliesseHandschlagAb führt den TLS-Handschlag und baut die Leitung.
//
// **Der Kanal entsteht erst danach** — vor dem Handschlag gibt es kein
// Schlüsselmaterial, und ein Kanal aus der Adresse wäre der schwächere.
func schliesseHandschlagAb(
	ctx context.Context, gesichert *tls.Conn, roh net.Conn, gegenstelle *peer.Kennung,
) (*Leitung, error) {
	if fehler := gesichert.HandshakeContext(ctx); fehler != nil {
		_ = roh.Close() //nolint:errcheck // der Weg endet hier ohnehin

		return nil, tlsfehler(fehler)
	}

	return leitungAus(gesichert, *gegenstelle)
}

// leitungAus rechnet die Kanalbindung aus und baut die Leitung darüber.
//
// # Warum beides eine Frage ist und nicht zwei
//
// **Weil der Aufrufer sonst einen Zweig hätte, den er nicht betreten kann.**
// Nach einem geglückten TLS-1.3-Handschlag gibt `ExportKeyingMaterial` immer
// etwas her; ein `if fehler != nil` eine Ebene höher wäre eine Zusage, die
// kein Lauf einlöst (§3). Hier steht er dort, wo er entsteht — und dort ist er
// zu betreten: vor dem Handschlag gibt es kein Material.
func leitungAus(gesichert *tls.Conn, gegenstelle peer.Kennung) (*Leitung, error) {
	// **Der Stand wird festgehalten** — `ExportKeyingMaterial` hängt an
	// einem Zeiger, und ein Ausdruck als Ruf ließe sich nicht adressieren.
	stand := gesichert.ConnectionState()

	roh, fehler := stand.ExportKeyingMaterial(Kanalbereich, nil, Kanallaenge)
	if fehler != nil {
		_ = gesichert.Close() //nolint:errcheck // der Weg endet hier ohnehin

		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Bindung dieser Verbindung ausrechnen")
	}

	leitung := Ueber(gesichert, hex.EncodeToString(roh))
	leitung.gegenstelle = gegenstelle

	return leitung, nil
}

// tlsfehler gibt einem gescheiterten Handschlag seine Klasse.
//
// # Warum ein gescheiterter Handschlag immer `ErrValidation` ist
//
// **Weil ein Horcher die zwei Klassen unterscheiden können muss.** Er läuft in
// einer Schleife, und die Frage bei jedem Fehler lautet: *diese eine
// Verbindung* ist nichts geworden — oder *der Lauscher* ist zu? Beim ersten
// wird weitergehorcht, beim zweiten aufgehört; wer beides gleich meldete,
// hätte entweder eine Schleife, die nach dem ersten krummen Gegenüber
// aufhört, oder eine, die an einem geschlossenen Lauscher endlos dreht.
//
// **Deshalb: `ErrNotFound` heißt „da ist niemand"** — der Lauscher gibt nichts
// mehr her, oder eine Adresse antwortet nicht (dann genauer
// [domain.ErrPeerNichtErreichbar], das `ErrNotFound` umschließt, #2491). **`ErrValidation` heißt „diese
// Verbindung ist nichts geworden"**, gleich ob am Festnageln, an der Fassung
// oder am Abriss.
//
// **Der Aufrufer liest diese Klasse** ([service.Horchposten]), und sie ist
// damit eine Zusage und keine Kosmetik.
func tlsfehler(fehler error) error {
	if errors.Is(fehler, domain.ErrValidation) {
		return fehler
	}

	return domain.MitGrund(domain.ErrValidation, fehler,
		"die gesicherte Verbindung kam nicht zustande")
}
