package leitung_test

// **Eine Leitung über TCP, und was darauf zu sehen ist** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
// # Warum hier TLS steht und nicht blankes TCP
//
// **Weil unterschrieben nicht verschlossen heißt.** `internal/peer`
// unterschreibt jede Hülle und verschlüsselt keine; libp2p hätte Noise oder
// TLS mitgebracht, und ohne libp2p
// ([0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md)) gingen
// Kartenbestände, Namen und Lagerorte im Klartext über das Netz.
//
// **Und der Lauf misst es am Schaden**: ein Mitleser sitzt auf der Strecke und
// schreibt mit, und was er mitschreibt, darf die Nachricht nicht sein.

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// zweiIdentitaeten gibt zwei frische Schlüsselpaare.
func zweiIdentitaeten(t *testing.T) (hier, dort peer.Identitaet) {
	t.Helper()

	for _, ziel := range []*peer.Identitaet{&hier, &dort} {
		erzeugt, fehler := peer.Erzeugen(nil)
		if fehler != nil {
			t.Fatalf("Erzeugen: %v", fehler)
		}

		*ziel = erzeugt
	}

	return hier, dort
}

// einHorcher öffnet einen Horcher auf der Rückschleife und räumt ihn wieder.
func einHorcher(t *testing.T, ich peer.Identitaet) *leitung.Horcher {
	t.Helper()

	horcher, fehler := leitung.Horche(t.Context(), "127.0.0.1:0", ich)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() {
		_ = horcher.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	return horcher
}

// **Zwei Rechner über TCP kommen auf dieselbe Leitung.**
func TestUeberTCPKommtEineNachrichtAn(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	var (
		warten     sync.WaitGroup
		dortige2   *leitung.Leitung
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		dortige2, dortFehler = horcher.Nimm(t.Context())
	}()

	hiesige2, fehler := leitung.Waehle(
		t.Context(), horcher.Adresse(), hiesige, dortige.Kennung())

	warten.Wait()

	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	if dortFehler != nil {
		t.Fatalf("Nimm: %v", dortFehler)
	}

	t.Cleanup(func() {
		_ = hiesige2.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dortige2.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	// **Beide Enden nennen denselben Kanal** — darauf ruht die Bindung des
	// Handschlags.
	if hiesige2.Kanal() != dortige2.Kanal() || hiesige2.Kanal() == "" {
		t.Fatalf("die Kanäle sind %q und %q",
			hiesige2.Kanal(), dortige2.Kanal())
	}

	// **Und er steht nicht in den Adressen.** Ein Kanal aus dem Vierertupel
	// wäre auf beiden Seiten gleich und für jeden auf der Strecke lesbar;
	// dieser kommt aus dem Schlüsselmaterial der Verbindung.
	for _, teil := range []string{
		horcher.Adresse(), "127.0.0.1", "tcp",
	} {
		if strings.Contains(hiesige2.Kanal(), teil) {
			t.Errorf("der Kanal %q trägt %q", hiesige2.Kanal(), teil)
		}
	}

	if befund := hiesige2.Sende(
		t.Context(), []byte("hallo")); befund != nil {
		t.Fatalf("Sende: %v", befund)
	}

	gelesen, befund := dortige2.Empfange(t.Context())
	if befund != nil {
		t.Fatalf("Empfange: %v", befund)
	}

	if string(gelesen) != "hallo" {
		t.Fatalf("angekommen ist %q", gelesen)
	}
}

// **Ein Mitleser auf der Strecke sieht die Nachricht nicht.**
//
// # Der Lauf, für den TLS hier steht
//
// **Er misst am Schaden** (§10): zwischen den beiden sitzt ein Weiterleiter,
// der jedes Byte mitschreibt. Ohne TLS stünde „Mausbraten" in seinem
// Mitschnitt — mit TLS steht es nicht darin.
func TestEinMitleserSiehtDieNachrichtNicht(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	mitschnitt := &mitleser{ziel: horcher.Adresse()}
	adresse := mitschnitt.starte(t)

	var (
		warten     sync.WaitGroup
		dortFehler error
		dortLeit   *leitung.Leitung
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		dortLeit, dortFehler = horcher.Nimm(t.Context())
	}()

	hierLeit, fehler := leitung.Waehle(
		t.Context(), adresse, hiesige, dortige.Kennung())

	warten.Wait()

	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	if dortFehler != nil {
		t.Fatalf("Nimm: %v", dortFehler)
	}

	t.Cleanup(func() {
		_ = hierLeit.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dortLeit.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	if fehler := hierLeit.Sende(
		t.Context(), []byte("Mausbraten")); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	if _, fehler := dortLeit.Empfange(t.Context()); fehler != nil {
		t.Fatalf("Empfange: %v", fehler)
	}

	if gehoert := mitschnitt.gehoert(); strings.Contains(
		gehoert, "Mausbraten") {
		t.Fatal("die Nachricht stand im Klartext auf der Leitung")
	}
}

// mitleser leitet weiter und schreibt mit.
type mitleser struct {
	ziel string

	schloss sync.Mutex
	bytes   strings.Builder
}

// starte öffnet den Weiterleiter und gibt seine Adresse.
func (m *mitleser) starte(t *testing.T) string {
	t.Helper()

	lauscher, fehler := (&net.ListenConfig{}).Listen(
		t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("Listen: %v", fehler)
	}

	t.Cleanup(func() {
		_ = lauscher.Close() //nolint:errcheck // nur ein Lauf
	})

	ctx := t.Context()

	go func() {
		herein, fehler := lauscher.Accept()
		if fehler != nil {
			return
		}

		hinaus, fehler := (&net.Dialer{}).DialContext(ctx, "tcp", m.ziel)
		if fehler != nil {
			_ = herein.Close() //nolint:errcheck // nur ein Lauf

			return
		}

		//nolint:errcheck // beide Richtungen enden mit der Verbindung
		go func() { _, _ = io.Copy(herein, hinaus) }()

		//nolint:errcheck // dito
		_, _ = io.Copy(io.MultiWriter(hinaus, m), herein)
	}()

	return lauscher.Addr().String()
}

func (m *mitleser) Write(teil []byte) (int, error) {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	m.bytes.Write(teil)

	return len(teil), nil
}

func (m *mitleser) gehoert() string {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	return m.bytes.String()
}
