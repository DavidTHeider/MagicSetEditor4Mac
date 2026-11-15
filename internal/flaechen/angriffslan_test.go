// **Der Angriffslauf über die eigene LAN-Adresse, mit TLS** — SEC01
// (#2354),
// Befund A-15: die Matrix aus QA-09 (`docs/PENTEST.md`) am heutigen Code und
// an der Stelle, an der ein fremder Rechner anklopft.
//
// # Was er fährt
//
// Ein `mse-rest` an `0.0.0.0` mit seinem Zeugnis, gefragt über die **eigene
// LAN-Adresse** — derselbe Schnittstellenstapel wie von einem fremden Rechner
// (`dienstbindung_test.go`) —, und jede Anfrage mit einem **festgenagelten**
// Klienten: er nimmt nur das Zeugnis mit dem angesagten Fingerabdruck, wie
// 0115 es verlangt.
//
//  1. **Zugang**: ohne und mit falscher Marke 401; ist das Budget der
//     Fehlversuche erschöpft, 429 — und der Berechtigte kommt trotzdem durch.
//  2. **Herkunft**: mit gültiger Marke, aber mit `Origin`, mit `Referer` oder
//     mit einem fremden Namen im `Host` (DNS-Rebinding) 403 — und **nichts
//     geschrieben**. Daneben dieselbe Anfrage ohne sie: sie schreibt.
//  3. **Zeugniswechsel**: nach `mse peer-rotieren` zeigt ein neu gestarteter
//     Dienst ein anderes Zeugnis; der Klient mit dem alten Abdruck kommt dort
//     nicht durch, der mit dem neuen schon.
//  4. **Die Ansage**: mit `-ansage` 0600 im 0700-Ordner (SEC02).
//
// # Was er nicht fährt
//
//   - zwei Geräte und fremde Netze (`docs/LABOR.md`),
//   - den Tausch des Bestands unter einem laufenden Dienst — das halten M01 und
//     M02, und `docs/PENTEST.md` nennt die Läufe,
//   - Rollen und Tokens: `rollenwechsel_test.go`, `rollenmatrix_test.go` und
//     `autorisierungsmatrix_test.go`.
package flaechen_test

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDerAngriffslaufUeberDieEigeneLANAdresse(t *testing.T) {
	eigene := eigeneAdresse(t)
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	mussKommandozeile(t, umgebung, "peer-anlegen")

	ansagedatei := filepath.Join(t.TempDir(), "schnittstellen", "dienst.txt")
	erst := starteLANDienst(t, umgebung, "-ansage", ansagedatei)
	wirt := net.JoinHostPort(eigene, portAus(t, erst.adresse))
	basis := "https://" + wirt
	klient := festgenagelt(erst.fingerabdruck)

	t.Run("Zugang", func(t *testing.T) {
		for name, marke := range map[string]string{"ohne Marke": "", "falsche Marke": strings.Repeat("0", 64)} {
			if status, _ := rufMitKopf(t, klient, http.MethodGet, basis+"/sets", marke, "", nil, ""); status != http.StatusUnauthorized {
				t.Errorf("%s: %d", name, status)
			}
		}

		// **Das Budget der Fehlversuche** — 60 Anfragen (`drosseln_test.go`).
		gedrosselt := false
		for range 70 {
			if status, _ := rufMitKopf(t, klient, http.MethodGet, basis+"/sets", "falsch", "", nil, ""); status == http.StatusTooManyRequests {
				gedrosselt = true

				break
			}
		}

		if !gedrosselt {
			t.Error("die Fehlversuche wurden nicht gedrosselt")
		}

		if status, rumpf := rufMitKopf(t, klient, http.MethodGet, basis+"/sets", erst.marke, "", nil, ""); status != http.StatusOK {
			t.Errorf("der Berechtigte wurde ausgesperrt: %d — %s", status, rumpf)
		}
	})

	t.Run("Herkunft", func(t *testing.T) {
		for name, fall := range map[string]struct {
			kopf map[string]string
			host string
		}{
			"Origin":       {map[string]string{"Origin": "https://angreifer.example"}, ""},
			"Referer":      {map[string]string{"Referer": "https://angreifer.example/seite"}, ""},
			"fremder Host": {nil, "angreifer.example:" + portAus(t, erst.adresse)},
		} {
			status, rumpf := rufMitKopf(t, klient, http.MethodPost, basis+"/decks", erst.marke,
				`{"name":"Angriff `+name+`","spiel":"magic"}`, fall.kopf, fall.host)
			if status != http.StatusForbidden {
				t.Errorf("%s: %d — %s", name, status, rumpf)
			}
		}

		// **Daneben dieselbe Anfrage ohne sie: sie schreibt** — sonst hieße
		// das Nein nur, dass der Weg nicht ging (§10).
		if status, rumpf := rufMitKopf(t, klient, http.MethodPost, basis+"/decks", erst.marke,
			`{"name":"Berechtigt","spiel":"magic"}`, nil, ""); status != http.StatusCreated && status != http.StatusOK {
			t.Fatalf("die berechtigte Anfrage: %d — %s", status, rumpf)
		}

		_, decks := rufMitKopf(t, klient, http.MethodGet, basis+"/decks", erst.marke, "", nil, "")
		if strings.Contains(decks, "Angriff") || !strings.Contains(decks, "Berechtigt") {
			t.Errorf("geschrieben wurde, was nicht durfte: %s", decks)
		}
	})

	t.Run("Ansage", func(t *testing.T) {
		for pfad, rechte := range map[string]os.FileMode{ansagedatei: 0o600, filepath.Dir(ansagedatei): 0o700} {
			if stand, fehler := os.Stat(pfad); fehler != nil || stand.Mode().Perm() != rechte {
				t.Errorf("%s: %v, %v", pfad, stand, fehler)
			}
		}
	})

	t.Run("Zeugniswechsel", func(t *testing.T) {
		mussKommandozeile(t, umgebung, "peer-rotieren")

		dann := starteLANDienst(t, umgebung)
		neuerWirt := "https://" + net.JoinHostPort(eigene, portAus(t, dann.adresse))

		if dann.fingerabdruck == erst.fingerabdruck || dann.fingerabdruck == "" {
			t.Fatalf("nach der Rotation derselbe Abdruck %q", dann.fingerabdruck)
		}

		// **Der alte Abdruck kommt beim neuen Zeugnis nicht durch** — die
		// Verbindung endet im Handschlag, vor jeder Kopfzeile.
		if status, rumpf := rufMitKopf(t, klient, http.MethodGet, neuerWirt+"/sets", dann.marke, "", nil, ""); status != 0 {
			t.Errorf("mit dem alten Abdruck: %d — %s", status, rumpf)
		}

		if status, rumpf := rufMitKopf(t, festgenagelt(dann.fingerabdruck), http.MethodGet, neuerWirt+"/sets", dann.marke, "", nil, ""); status != http.StatusOK {
			t.Errorf("mit dem neuen Abdruck: %d — %s", status, rumpf)
		}
	})
}

// starteLANDienst startet `mse-rest --adresse 0.0.0.0` mit weiteren
// Argumenten und gibt seine Ansage.
func starteLANDienst(t *testing.T, umgebung []string, weitere ...string) ansage {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.rest, append([]string{"--adresse", "0.0.0.0", "--port", "0"}, weitere...)...)
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if fehler := lauf.Start(); fehler != nil {
		beenden()
		t.Fatalf("den Dienst starten: %v", fehler)
	}

	geendet := make(chan struct{})

	go func() {
		_ = lauf.Wait() //nolint:errcheck,gosec // beendet wird über den Kontext
		close(geendet)
	}()

	t.Cleanup(func() {
		beenden()
		<-geendet
	})

	adresse, marke, fingerabdruck := ansageLesen(t, angesagt, geendet)
	if !strings.HasPrefix(adresse, "https://") || fingerabdruck == "" {
		t.Fatalf("der Dienst sagt %q mit Abdruck %q an — jenseits der Schleife ist TLS Pflicht", adresse, fingerabdruck)
	}

	return ansage{adresse: adresse, marke: marke, fingerabdruck: fingerabdruck}
}

// festgenagelt ist ein Klient, der nur das Zeugnis mit `abdruck` nimmt.
func festgenagelt(abdruck string) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //#nosec G402 -- geprüft wird der Abdruck, in VerifyConnection
			MinVersion:         tls.VersionTLS13,
			VerifyConnection: func(stand tls.ConnectionState) error {
				if len(stand.PeerCertificates) == 0 {
					return errors.New("kein Zeugnis")
				}

				oeffentlich, taugt := stand.PeerCertificates[0].PublicKey.(ed25519.PublicKey)
				if !taugt {
					return errors.New("kein Ed25519")
				}

				kennung, fehler := peer.KennungFuer(oeffentlich)
				if fehler != nil || peer.Fingerabdruck(kennung) != abdruck {
					return errors.New("ein fremdes Zeugnis")
				}

				return nil
			},
		}},
	}
}

// rufMitKopf fragt mit Kopfzeilen und, wenn nicht leer, einem anderen `Host`.
// **0 heißt: keine Antwort** — die Verbindung kam nicht zustande.
func rufMitKopf(
	t *testing.T, klient *http.Client, methode, adresse, marke, koerper string, kopf map[string]string, host string,
) (status int, rumpf string) {
	t.Helper()

	var leser io.Reader
	if koerper != "" {
		leser = strings.NewReader(koerper)
	}

	anfrage, fehler := http.NewRequestWithContext(t.Context(), methode, adresse, leser)
	if fehler != nil {
		t.Fatalf("die Anfrage bauen: %v", fehler)
	}

	if marke != "" {
		anfrage.Header.Set("Authorization", "Bearer "+marke)
	}

	for name, wert := range kopf {
		anfrage.Header.Set(name, wert)
	}

	if host != "" {
		anfrage.Host = host
	}

	antwort, fehler := klient.Do(anfrage)
	if fehler != nil {
		return 0, fehler.Error()
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	roh, _ := io.ReadAll(io.LimitReader(antwort.Body, 1<<20)) //nolint:errcheck // ein halber Rumpf ist hier kein Befund

	return antwort.StatusCode, string(roh)
}
