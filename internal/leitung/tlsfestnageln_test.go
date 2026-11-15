package leitung_test

// **Wogegen festgenagelt wird** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine zweite Datei**, weil `tls_test.go` mit ihr über die Zeilengrenze aus
// §2 ginge. Der Schnitt folgt der Frage: dort, **dass** der Weg trägt — hier,
// **mit wem**.
//
// # Warum das eine eigene Frage ist
//
// **Weil eine Adresse nichts darüber sagt, wer an ihr sitzt.** Sie kann
// umgeleitet, übernommen oder schlicht falsch abgetippt sein. Ohne das
// Festnageln redete „Adresse von Hand" mit irgendwem — und der Handschlag aus
// N1 fiele erst danach darüber, wenn schon eine Verbindung steht.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Ein anderes Gegenüber als erwartet wird abgewiesen.**
//
// **Das ist der ganze Grund für das Festnageln.** Eine Adresse sagt nichts
// darüber, wer an ihr sitzt — sie kann umgeleitet, übernommen oder schlicht
// falsch abgetippt sein. Wer das Zeugnis nicht gegen die erwartete Kennung
// hielte, redete mit irgendwem.
func TestEinAnderesGegenueberAlsErwartetWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	fremde, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	go func() {
		genommen, fehler := horcher.Nimm(t.Context())
		if fehler == nil {
			_ = genommen.Schliesse() //nolint:errcheck // nur ein Lauf
		}
	}()

	if _, fehler := leitung.Waehle(
		t.Context(), horcher.Adresse(), hiesige, fremde.Kennung(),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ohne erwartete Kennung wird gar nicht erst gewählt.**
//
// **Eine leere Erwartung wäre „mit jedem, der da sitzt"** — und damit genau
// das, wogegen das Festnageln steht. Sie wird vor dem ersten Paket abgewiesen.
func TestOhneErwarteteKennungWirdNichtGewaehlt(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	if _, fehler := leitung.Waehle(
		t.Context(), horcher.Adresse(), hiesige, "",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ein Zeugnis mit einem anderen Schlüsseltyp kommt nicht durch.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Kennung ein Ed25519-Schlüssel *ist*.** Ein Gegenüber mit einem
// ECDSA-Zeugnis hätte keine — und käme es durch, bände der Kanal an etwas, das
// kein Mensch je bestätigen könnte. Der Handschlag fiele erst danach darüber,
// und dann stünde schon eine Verbindung.
//
// # Die Gegenprobe, und der erste Versuch, der keine war
//
// **Der erste Anlauf hängte die Typprüfung aus und blieb grün** — und das war
// nicht der Test, sondern der Eingriff: ohne sie steht in `oeffentlich` ein
// `nil`, und die Zeile darunter weist es ab, weil `peer.KennungFuer` eine
// Länge prüft. Zwei Wächter für eine Sache, und der Schaden traf den, der
// nicht zählte (§10: eine Gegenprobe, die nichts ändert, ist keine).
//
// **Der zweite traf.** Mit einem `if erwartet == "" { return nil }` am Anfang
// der Prüfung — dem Fehler, den jemand wirklich baute, weil die horchende
// Seite ja „ohnehin niemanden erwartet" — meldet dieser Lauf `<nil>` statt
// einer Ablehnung. Nachgestellt am 2026-09-14.
func TestEinZeugnisMitAnderemSchluesseltypKommtNichtDurch(t *testing.T) {
	t.Parallel()

	_, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		genommen, fehler := horcher.Nimm(t.Context())

		dortFehler = fehler

		if fehler == nil {
			_ = genommen.Schliesse() //nolint:errcheck // nur ein Lauf
		}
	}()

	waehler := &tls.Dialer{Config: &tls.Config{
		Certificates:       []tls.Certificate{einEcdsaZeugnis(t)},
		InsecureSkipVerify: true, //#nosec G402 -- der Lauf prüft die Gegenseite
		MinVersion:         tls.VersionTLS13,
	}}

	strom, fehler := waehler.DialContext(
		t.Context(), "tcp", horcher.Adresse())
	if fehler == nil {
		// **Erst ein Schreiben zeigt es dem Klienten** — unter TLS 1.3
		// scheitert der Handschlag der anbietenden Seite später als der der
		// annehmenden.
		_, _ = strom.Write([]byte("x")) //nolint:errcheck // hier zählt die andere Seite
		_ = strom.Close()               //nolint:errcheck // nur ein Lauf
	}

	warten.Wait()

	if !errors.Is(dortFehler, domain.ErrValidation) {
		t.Fatalf("die horchende Seite sagte: %v", dortFehler)
	}
}

// einEcdsaZeugnis baut ein selbst unterschriebenes Zeugnis, das kein
// Ed25519-Zeugnis ist.
func einEcdsaZeugnis(t *testing.T) tls.Certificate {
	t.Helper()

	schluessel, fehler := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if fehler != nil {
		t.Fatalf("GenerateKey: %v", fehler)
	}

	entwurf := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	roh, fehler := x509.CreateCertificate(
		rand.Reader, entwurf, entwurf, &schluessel.PublicKey, schluessel)
	if fehler != nil {
		t.Fatalf("CreateCertificate: %v", fehler)
	}

	return tls.Certificate{Certificate: [][]byte{roh}, PrivateKey: schluessel}
}
