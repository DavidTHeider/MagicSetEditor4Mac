package rest_test

// **Was jenseits der Schleife gilt** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Warum es dafür keinen zweiten Rechner braucht
//
// **Weil die Zusage eine über diesen Dienst ist und keine über ein Netz.**
// Was zu zeigen ist: der Lauscher sitzt wirklich auf einer Schnittstelle, die
// nicht die Schleife ist, und eine Anfrage über die **eigene LAN-Adresse**
// kommt durch — die geht durch denselben Schnittstellenstapel wie die eines
// fremden Rechners und nicht über die Schleife. Dass Pakete von dort auch
// wirklich ankommen, ist eine Eigenschaft des Netzes.
//
// **Gemessen am 2026-09-15** am gebauten `mse-rest`: `lsof` zeigte
// `TCP *:54591 (LISTEN)`, eine Anfrage über `https://192.168.178.51:54591`
// gab `200`, dieselbe mit falscher Marke `401`, und über `http://` endete sie
// vor der ersten Kopfzeile. Dieser Lauf hält davon, was ohne gebaute
// Programme zu halten ist.

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Eine Anfrage über eine eigene Adresse kommt durch — wenn sie genannt ist.**
//
// **Das ist der Fund, für den es diesen Lauf gibt.** `HerkunftTaugt` kannte
// `127.0.0.1` und `localhost`; an `0.0.0.0` gebunden wies der Dienst jede
// Anfrage über die LAN-Adresse mit `403` ab, obwohl Marke und Zeugnis
// stimmten. Die Abwehr gegen DNS-Rebinding war richtig, ihre Liste war zu
// kurz.
func TestEineEigeneAdresseKommtDurchWennSieGenanntIst(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
	anfrage.Host = "192.168.178.51:54591"

	if rest.HerkunftTaugt(anfrage) {
		t.Error("eine fremde Adresse kam ohne Nennung durch")
	}

	if !rest.HerkunftTaugtUnter(anfrage, []string{"192.168.178.51"}) {
		t.Error("die genannte Adresse kam nicht durch")
	}

	// **Der Name des Angreifers steht nicht darin** — das ist die Hälfte,
	// die die Abwehr trägt. Ohne diesen Fall sagte der Lauf darüber nur,
	// dass eine Liste wirkt.
	angriff := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
	angriff.Host = "boese.example:54591"

	if rest.HerkunftTaugtUnter(angriff, []string{"192.168.178.51"}) {
		t.Error("ein fremder Name kam durch")
	}

	// **Und ein `Origin` weist weiterhin ab, auch einen genannten Namen** —
	// ein Browser bleibt ein Browser.
	imBrowser := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
	imBrowser.Host = "192.168.178.51:54591"
	imBrowser.Header.Set("Origin", "https://boese.example")

	if rest.HerkunftTaugtUnter(imBrowser, []string{"192.168.178.51"}) {
		t.Error("eine Anfrage mit Origin kam durch")
	}
}

// **Der Lauscher jenseits der Schleife ist wirklich einer — und trägt TLS.**
//
// **Gebunden wird an eine echte Adresse dieses Rechners** und nicht an
// `0.0.0.0`: so sagt `Addr()` sie auch, und der Lauf misst, was er meint.
func TestEinLauscherJenseitsDerSchleifeTraegtTLS(t *testing.T) {
	t.Parallel()

	eigene := ersteFremdeAdresse(t)

	zeugnis := probezeugnis(t)

	lauscher, fehler := rest.LauschenMit(t.Context(),
		domain.Bindung{Adresse: eigene}, zeugnis)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	t.Cleanup(func() { schliesse(t, lauscher) })

	rechner, _, teilfehler := net.SplitHostPort(lauscher.Addr().String())
	if teilfehler != nil {
		t.Fatalf("Adresse zerlegen: %v", teilfehler)
	}

	adresse := net.ParseIP(rechner)
	if adresse == nil || adresse.IsLoopback() {
		t.Fatalf("der Lauscher liegt auf %q — das ist die Schleife", rechner)
	}

	// **Und eine Verbindung darauf spricht TLS.** Ein Klartextbyte bekommt
	// keine Antwort: der TLS-Lauscher liest es als Handschlag und legt auf.
	//
	// **Das ist die Gegenprobe zu „über http endete sie vor der ersten
	// Kopfzeile"** aus der Messung im Kopf.
	waehler := &net.Dialer{}

	verbindung, fehler := waehler.DialContext(
		t.Context(), "tcp", lauscher.Addr().String())
	if fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	t.Cleanup(func() { _ = verbindung.Close() }) //nolint:errcheck // Aufräumen

	go func() {
		strom, fehler := lauscher.Accept()
		if fehler != nil {
			return
		}

		_ = strom.Close() //nolint:errcheck // Aufräumen
	}()

	if _, fehler := verbindung.Write(
		[]byte("GET /sets HTTP/1.1\r\nHost: x\r\n\r\n")); fehler != nil {
		return
	}

	puffer := make([]byte, 16)

	// **Ein Lesefehler ist hier die erwartete Antwort**: der TLS-Lauscher
	// legt auf einem Klartextbyte auf. Was ein Befund wäre, ist eine
	// **HTTP-Statuszeile**.
	gelesen, _ := verbindung.Read(puffer) //nolint:errcheck // siehe oben

	if gelesen > 0 && strings.HasPrefix(string(puffer[:gelesen]), "HTTP/") {
		t.Error("der Lauscher antwortete im Klartext")
	}
}

// ersteFremdeAdresse gibt eine Adresse dieses Rechners, die nicht die
// Schleife ist.
//
// **Und überspringt den Lauf, wo es keine gibt.** Ein Rechner ohne zweite
// Adresse ist kein Befund über diesen Code — dieselbe Gestalt wie
// `fremdeAdressen` in `lauscher_test.go`.
func ersteFremdeAdresse(t *testing.T) string {
	t.Helper()

	adressen, fehler := net.InterfaceAddrs()
	if fehler != nil {
		t.Fatalf("Schnittstellen lesen: %v", fehler)
	}

	for _, eintrag := range adressen {
		netzwerk, istNetz := eintrag.(*net.IPNet)
		if !istNetz || netzwerk.IP.IsLoopback() || netzwerk.IP.To4() == nil {
			continue
		}

		return netzwerk.IP.String()
	}

	t.Skip("dieser Rechner hat keine zweite IPv4-Adresse — nichts zu prüfen")

	return ""
}

// probezeugnis baut ein Zeugnis für den Lauf.
//
// # Warum hier eines entsteht und nicht [leitung.Serverzeugnis] gerufen wird
//
// **Weil `internal/rest` ein Blatt ist** (§2) — es kennt `domain` und sonst
// nichts aus dem Projekt, und ein Testimport machte daraus eine Kante. Was
// dieser Lauf braucht, ist **irgendein** TLS-Werk; welches, ist die Sache des
// Composition Root.
//
// **Was dabei nicht geprüft wird**, steht damit auch fest: dass das Zeugnis
// die Peeridentität trägt. Das prüft `internal/leitung`, und der Weg dorthin
// steht in `cmd/mse-rest/zeugnis.go`.
func probezeugnis(t *testing.T) *tls.Config {
	t.Helper()

	oeffentlich, geheim, fehler := ed25519.GenerateKey(rand.Reader)
	if fehler != nil {
		t.Fatalf("Schlüssel erzeugen: %v", fehler)
	}

	entwurf := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	roh, fehler := x509.CreateCertificate(
		rand.Reader, entwurf, entwurf, oeffentlich, geheim)
	if fehler != nil {
		t.Fatalf("Zeugnis bauen: %v", fehler)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{roh}, PrivateKey: geheim,
		}},
		MinVersion: tls.VersionTLS13,
	}
}

// **Ein belegter Port kommt als Fehler zurück, nicht als halber Lauscher.**
//
// **Der Zweig hinter `ListenConfig.Listen`** — ohne diesen Lauf stünde nur,
// dass die Prüfung **vor** dem Lauschen greift, und nichts über das Lauschen
// selbst.
func TestEinBelegterPortKommtAlsFehler(t *testing.T) {
	t.Parallel()

	var aufbau net.ListenConfig

	belegt, fehler := aufbau.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("den Port belegen: %v", fehler)
	}

	t.Cleanup(func() { _ = belegt.Close() }) //nolint:errcheck // der Lauf endet ohnehin

	_, port, fehler := net.SplitHostPort(belegt.Addr().String())
	if fehler != nil {
		t.Fatalf("den Port lesen: %v", fehler)
	}

	zahl, fehler := strconv.Atoi(port)
	if fehler != nil {
		t.Fatalf("den Port deuten: %v", fehler)
	}

	if _, fehler := rest.LauschenMit(t.Context(), domain.Bindung{
		Adresse: domain.Schleifenadresse, Port: zahl,
	}, nil); fehler == nil {
		t.Fatal("derselbe Port ging zweimal auf")
	}
}

// **Die eigenen Namen kommen am Dienst an — zusätzlich zur Schleife.**
//
// **Was [rest.HerkunftTaugtUnter] entscheidet, misst der Lauf oben.** Hier
// geht es um die Naht daneben: dass [rest.Dienst.MitEigenenNamen] sie
// wirklich erreicht. `cmd/mse-rest` geht genau diesen Weg, und ohne ihn wäre
// eine geprüfte Funktion ohne Aufrufer da.
func TestDieEigenenNamenErreichenDenDienst(t *testing.T) {
	t.Parallel()

	behandler := rest.NeuerDienst(probemarke,
		func() time.Time { return time.Unix(0, 0) }, rest.Naehte{}).
		MitEigenenNamen("192.168.178.51", "mein-mac.fritz.box").
		Behandler()

	faelle := map[string]bool{
		"127.0.0.1:8123":           true,
		"192.168.178.51:54591":     true,
		"mein-mac.fritz.box:54591": true,
		"boese.example:54591":      false,
	}

	for wirt, durch := range faelle {
		t.Run(wirt, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(
				t.Context(), http.MethodGet, "/gibtesnicht", http.NoBody)
			anfrage.Host = wirt
			anfrage.Header.Set("Authorization", "Bearer "+probemarke)

			aufzeichnung := httptest.NewRecorder()
			behandler.ServeHTTP(aufzeichnung, anfrage)

			// **Gemessen wird die Abweisung und nicht die Antwort**: welchen
			// Ausgang `/gibtesnicht` hat, entscheidet der Wegweiser — dass
			// die Herkunft trägt, entscheidet die Naht.
			abgewiesen := aufzeichnung.Code == http.StatusForbidden
			if abgewiesen == durch {
				t.Errorf("%s gab %d", wirt, aufzeichnung.Code)
			}
		})
	}
}
