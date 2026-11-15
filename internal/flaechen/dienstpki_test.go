package flaechen_test

// **REST und MCP mit einer eigenen PKI, gemessen mit den Klienten, die es
// geben wird** — NET-03.AK-04
// (#1707):
// „Client-Kompatibilität wird mit den tatsächlich vorgesehenen curl-, Go-
// und MCP-Clients gemessen, nicht nur mit dem eigenen Client."
//
// # Was hier gemessen wird
//
//   - **`/usr/bin/curl` von macOS mit `--cacert`**, ohne `-k`. Am
//     Zeugnis der Identität scheiterte es am Handschlag (Ed25519 kennt
//     LibreSSL 3.3.6 nicht, `docs/NETZEIGENSCHAFTEN.md`); an einem ECDSA-Zeugnis
//     aus einer eigenen Stelle muss es durchkommen.
//   - **Ein Go-Klient, der prüft** — gegen den Anker des Hauses und den
//     Namen im Zeugnis, ohne `InsecureSkipVerify`.
//   - **Ein MCP-Aufruf** (`initialize`) über denselben Weg.
//   - **Die Gegenprobe**: ein Klient, der nur die Anker des Systems kennt,
//     wird abgewiesen. Ohne sie sagte ein grünes „200" nicht, ob überhaupt
//     geprüft wurde.
//   - **Der Weg über die Einstellungen**: dasselbe Zeugnis, gesetzt mit
//     `mse einstellungen setzen`, und der Dienst ohne Flaggen gestartet.

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// eigenePKI legt Wurzel und Blatt für eine Adresse ab.
func eigenePKI(t *testing.T, adresse string) (leitung.Zeugnisdateien, *x509.CertPool) {
	t.Helper()

	ordner := t.TempDir()
	dateien := leitung.Zeugnisdateien{
		Zertifikat: filepath.Join(ordner, "dienst.pem"),
		Schluessel: filepath.Join(ordner, "schluessel.pem"),
		Anker:      filepath.Join(ordner, "anker.pem"),
	}

	wurzelschluessel := muss[*ecdsa.PrivateKey](t)(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
	wurzelvorlage := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Probe-Haus"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}

	wurzelroh, fehler := x509.CreateCertificate(rand.Reader, wurzelvorlage, wurzelvorlage,
		&wurzelschluessel.PublicKey, wurzelschluessel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	wurzel := muss[*x509.Certificate](t)(x509.ParseCertificate(wurzelroh))

	blattschluessel := muss[*ecdsa.PrivateKey](t)(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
	blattroh, fehler := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "mse"},
		DNSNames: []string{"dienst.haus.test"}, IPAddresses: []net.IP{net.ParseIP(adresse)},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, wurzel, &blattschluessel.PublicKey, wurzelschluessel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	schluesselroh := muss[[]byte](t)(x509.MarshalPKCS8PrivateKey(blattschluessel))

	for pfad, block := range map[string]*pem.Block{
		dateien.Zertifikat: {Type: "CERTIFICATE", Bytes: blattroh},
		dateien.Schluessel: {Type: "PRIVATE KEY", Bytes: schluesselroh},
		dateien.Anker:      {Type: "CERTIFICATE", Bytes: wurzelroh},
	} {
		if fehler := os.WriteFile(pfad, pem.EncodeToMemory(block), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	anker := x509.NewCertPool()
	anker.AddCert(wurzel)

	return dateien, anker
}

// muss bricht den Lauf ab, wenn ein Schritt der Vorrichtung scheitert.
func muss[W any](t *testing.T) func(W, error) W {
	t.Helper()

	return func(wert W, fehler error) W {
		t.Helper()

		if fehler != nil {
			t.Fatal(fehler)
		}

		return wert
	}
}

// pruefenderKlient kennt nur die gegebenen Anker — nie `InsecureSkipVerify`.
func pruefenderKlient(anker *x509.CertPool) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs: anker, MinVersion: tls.VersionTLS13,
		}},
	}
}

func zeugnisflaggen(dateien leitung.Zeugnisdateien) []string {
	return []string{"--zeugnis", dateien.Zertifikat, "--schluessel", dateien.Schluessel, "--anker", dateien.Anker}
}

func TestEigenePKIMitCurlGoUndMCP(t *testing.T) {
	eigene := eigeneAdresse(t)
	dateien, anker := eigenePKI(t, eigene)
	umgebung := mitEigenerSaat(t)

	rest := starteMitFlaggen(t, gebaut.rest, umgebung, zeugnisflaggen(dateien)...)
	ziel := "https://" + net.JoinHostPort(eigene, portAus(t, rest.adresse)) + "/sets"

	// **Go, prüfend.**
	if status, rumpf := rufMitKlient(t, pruefenderKlient(anker),
		http.MethodGet, ziel, rest.marke, ""); status != http.StatusOK {
		t.Errorf("Go mit Anker: %d — %s", status, rumpf)
	}

	// **Die Gegenprobe: nur die Anker des Systems.** Status 0 heißt: der
	// Handschlag scheiterte an der Prüfung.
	if status, _ := rufMitKlient(t, pruefenderKlient(nil),
		http.MethodGet, ziel, rest.marke, ""); status != 0 {
		t.Errorf("ohne den Anker des Hauses kam %d zurück", status)
	}

	// **Und der Zugang gilt weiter** (AK-04): ohne Marke keine Antwort.
	if status, _ := rufMitKlient(t, pruefenderKlient(anker),
		http.MethodGet, ziel, "", ""); status != http.StatusUnauthorized {
		t.Errorf("ohne Marke kam %d zurück", status)
	}

	// **`/usr/bin/curl` von macOS**, ohne `-k`.
	if _, fehler := os.Stat("/usr/bin/curl"); fehler == nil {
		// #nosec G204 -- fester Pfad, eigene Argumente
		ausgabe, fehler := exec.CommandContext(t.Context(), "/usr/bin/curl", "-sS",
			"--cacert", dateien.Anker, "-H", "Authorization: Bearer "+rest.marke,
			"-o", "/dev/null", "-w", "%{http_code}", ziel).CombinedOutput()
		if fehler != nil || strings.TrimSpace(string(ausgabe)) != "200" {
			t.Errorf("/usr/bin/curl: %v — %s", fehler, ausgabe)
		}
	}

	// **MCP über denselben Weg** — `initialize`, wie ein Klient beginnt.
	mcp := starteMitFlaggen(t, gebaut.mcp, umgebung, zeugnisflaggen(dateien)...)
	mcpziel := "https://" + net.JoinHostPort(eigene, portAus(t, mcp.adresse)) + "/mcp"

	status, rumpf := rufMitKlient(t, pruefenderKlient(anker), http.MethodPost, mcpziel, mcp.marke,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18",`+
			`"capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`)
	if status != http.StatusOK || !strings.Contains(rumpf, "protocolVersion") {
		t.Errorf("MCP initialize: %d — %s", status, rumpf)
	}
}

// **Der Weg über die Einstellungen** — die Konfiguration dieses Rechners.
func TestDieEinstellungenGebenDemDienstSeinZeugnis(t *testing.T) {
	eigene := eigeneAdresse(t)
	dateien, anker := eigenePKI(t, eigene)
	umgebung := mitEigenerSaat(t)

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "dienst.tls.zertifikat", dateien.Zertifikat)
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "dienst.tls.schluessel", dateien.Schluessel)
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "dienst.tls.anker", dateien.Anker)

	rest := starteMitFlaggen(t, gebaut.rest, umgebung)
	ziel := "https://" + net.JoinHostPort(eigene, portAus(t, rest.adresse)) + "/sets"

	if status, rumpf := rufMitKlient(t, pruefenderKlient(anker),
		http.MethodGet, ziel, rest.marke, ""); status != http.StatusOK {
		t.Errorf("mit dem Zeugnis aus den Einstellungen: %d — %s", status, rumpf)
	}
}

// starteMitFlaggen startet ein Programm an `0.0.0.0` mit weiteren Flaggen.
//
// **`--adresse` bleibt die erste Flagge** — daran findet
// `internal/werkzeuge/flaechenlauf_test.go` die Fähigkeit, und der Vergleich
// in [starteProgrammAn] hält es.
func starteMitFlaggen(t *testing.T, programm string, umgebung []string, flaggen ...string) ansage {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, programm,
		append([]string{"--adresse", "0.0.0.0", "--port", "0"}, flaggen...)...)
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if fehler := lauf.Start(); fehler != nil {
		beenden()
		t.Fatalf("%s starten: %v", programm, fehler)
	}

	geendet := make(chan struct{})

	go func() {
		_ = lauf.Wait() //nolint:errcheck,gosec // beendet wird über ein Signal

		close(geendet)
	}()

	t.Cleanup(func() {
		beenden()
		<-geendet
	})

	gefunden, marke, abdruck := ansageLesen(t, angesagt, geendet)

	// **Angesagt wird der Abdruck des vorgezeigten Zeugnisses**, in der Form
	// von `openssl x509 -fingerprint -sha256`.
	if strings.Count(abdruck, ":") != 31 {
		t.Errorf("der angesagte Abdruck %q hat nicht die Form von openssl", abdruck)
	}

	if !strings.Contains(angesagt.text(), "Zeugnis von Probe-Haus bis ") {
		t.Errorf("die Ansage nennt die Stelle nicht: %q", angesagt.text())
	}

	return ansage{adresse: gefunden, marke: marke, fingerabdruck: abdruck}
}
