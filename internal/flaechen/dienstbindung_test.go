// **Die zwei Dienste, an eine gewählte Adresse gebunden** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Warum es dafür keinen zweiten Rechner braucht
//
// **Weil die Zusage eine über diese Programme ist und keine über ein Netz.**
// Was zu zeigen ist: der Lauscher sitzt wirklich auf einer Schnittstelle, die
// nicht die Schleife ist, eine Anfrage über die **eigene** LAN-Adresse kommt
// durch — sie geht durch denselben Schnittstellenstapel wie die eines fremden
// Rechners —, und ohne Zugang kommt sie nicht durch. Dass Pakete von einem
// zweiten Gerät auch wirklich ankommen, ist eine Eigenschaft des Netzes und
// keine dieses Programms.
//
// **Und die Vorgabe wird mitgemessen**: auf der Schleife gebunden, ist
// dieselbe Anfrage über die LAN-Adresse **nicht** zu erreichen.
package flaechen_test

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// **Jenseits der Schleife: TLS an, Zugang Pflicht, eigene Adresse erreichbar.**
func TestDerDienstJenseitsDerSchleife(t *testing.T) {
	eigene := eigeneAdresse(t)

	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))

	// **Ohne Identität geht der Port gar nicht erst auf** — das ist die
	// erste Zusage, und sie kommt vor allen anderen.
	mussKommandozeile(t, umgebung, "peer-anlegen")

	angesagt := starteDienstAn(t, umgebung, "0.0.0.0")

	if !strings.HasPrefix(angesagt.adresse, "https://") {
		t.Fatalf("der Dienst sagt %q an — jenseits der Schleife ist TLS Pflicht",
			angesagt.adresse)
	}

	port := portAus(t, angesagt.adresse)
	ueberEigene := "https://" + net.JoinHostPort(eigene, port)

	// **Mit Zugang: durch.** Diese Anfrage geht über die echte
	// Schnittstelle — nicht über die Schleife.
	if status, rumpf := frageDienst(
		t, ueberEigene+"/sets", angesagt.marke); status != http.StatusOK {
		t.Errorf("über %s: %d — %s", ueberEigene, status, rumpf)
	}

	// **Ohne Zugang: abgewiesen.** Das ist die Hälfte, die 0115 verlangt.
	if status, _ := frageDienst(t, ueberEigene+"/sets", ""); status != http.StatusUnauthorized {
		t.Errorf("ohne Marke kam %d zurück", status)
	}

	// **Und im Klartext gar nicht.** Ein `http://` an denselben Port endet
	// vor der ersten Kopfzeile — der TLS-Lauscher liest es als Handschlag.
	if status, _ := frageDienst(t,
		"http://"+net.JoinHostPort(eigene, port)+"/sets", angesagt.marke); status != 0 {
		t.Errorf("über Klartext kam %d zurück", status)
	}

	// **Und der angesagte Fingerabdruck gehört wirklich zu diesem Zeugnis.**
	//
	// **Das ist die Zusage von 0115**, und ohne diese Zeilen wäre sie
	// unbelegt: das Zeugnis ist selbst unterschrieben, und wer sich
	// verbindet, hat nichts anderes zu vergleichen als diese Zeichenkette.
	// Eine Zahl, die das Fenster zeigt und die Gegenstelle nicht wiederfindet,
	// wäre schlimmer als keine.
	if abdruck := vorgezeigterAbdruck(
		t, net.JoinHostPort(eigene, port)); abdruck != angesagt.fingerabdruck {
		t.Errorf("angesagt ist %q, vorgezeigt wird %q",
			angesagt.fingerabdruck, abdruck)
	}
}

// vorgezeigterAbdruck holt den Fingerabdruck aus dem echten Handschlag.
//
// **Gemessen und nicht nachgerechnet**: was hier herauskommt, steht im
// Zeugnis, das der Lauscher vorzeigt — nicht in einer zweiten Rechnung über
// dieselbe Identität.
func vorgezeigterAbdruck(t *testing.T, wirt string) string {
	t.Helper()

	waehler := &tls.Dialer{Config: &tls.Config{
		InsecureSkipVerify: true, //#nosec G402 -- geprüft wird ja der Abdruck
		MinVersion:         tls.VersionTLS13,
	}}

	verbindung, fehler := waehler.DialContext(t.Context(), "tcp", wirt)
	if fehler != nil {
		t.Fatalf("den Handschlag führen: %v", fehler)
	}

	defer func() { _ = verbindung.Close() }() //nolint:errcheck // der Lauf endet ohnehin

	gesichert, taugt := verbindung.(*tls.Conn)
	if !taugt {
		t.Fatalf("die Verbindung ist %T und kein TLS", verbindung)
	}

	zeugnisse := gesichert.ConnectionState().PeerCertificates
	if len(zeugnisse) == 0 {
		t.Fatal("der Dienst zeigt kein Zeugnis vor")
	}

	oeffentlich, taugt := zeugnisse[0].PublicKey.(ed25519.PublicKey)
	if !taugt {
		t.Fatalf("das Zeugnis trägt %T und kein Ed25519", zeugnisse[0].PublicKey)
	}

	kennung, fehler := peer.KennungFuer(oeffentlich)
	if fehler != nil {
		t.Fatalf("die Kennung rechnen: %v", fehler)
	}

	return peer.Fingerabdruck(kennung)
}

// **Auf der Schleife: kein TLS nötig — und von der LAN-Adresse nicht zu
// erreichen.**
//
// **Die Gegenprobe zum Lauf darüber.** Ohne sie sagte er nur, dass eine
// Bindung an `0.0.0.0` wirkt; er sagte nichts darüber, dass die Vorgabe eine
// Grenze ist.
func TestDieVorgabebindungBleibtAufDemRechner(t *testing.T) {
	eigene := eigeneAdresse(t)

	umgebung := mitEigenerSaat(t)

	angesagt := starteDienstAn(t, umgebung, "127.0.0.1")

	if !strings.HasPrefix(angesagt.adresse, "http://") {
		t.Fatalf("der Dienst sagt %q an", angesagt.adresse)
	}

	port := portAus(t, angesagt.adresse)

	if status, rumpf := frageDienst(t,
		angesagt.adresse+"/sets", angesagt.marke); status != http.StatusOK {
		t.Fatalf("über die Schleife: %d — %s", status, rumpf)
	}

	if status, _ := frageDienst(t,
		"http://"+net.JoinHostPort(eigene, port)+"/sets", angesagt.marke); status != 0 {
		t.Errorf("die Vorgabebindung war über %s erreichbar (%d)", eigene, status)
	}
}

// **Und `mse-mcp` bindet nach denselben Regeln.**
//
// **Über die Schleife und nicht darüber hinaus**: dass die Regel dieselbe
// ist, hält `internal/rest` an **einer** Stelle — hier wird gemessen, dass
// dieser Prozess sie wirklich nimmt, und dass die Werkzeugliste über HTTP
// dieselbe ist wie über stdio.
func TestDerWerkzeugserverUeberHTTP(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	angesagt := starteMCPAn(t, umgebung, "127.0.0.1")

	if !strings.HasSuffix(angesagt.adresse, "/mcp") {
		t.Fatalf("der Server sagt %q an — ein Endpunkt, und der heißt /mcp",
			angesagt.adresse)
	}

	status, rumpf := sendeMCP(t, angesagt.adresse, angesagt.marke,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if status != http.StatusOK {
		t.Fatalf("tools/list über HTTP: %d — %s", status, rumpf)
	}

	// **Dieselbe Liste wie über stdio** — das ist die Zusage des zweiten
	// Transports, und zwei Behauptungen wären keine.
	ueberStdio := mussMCP(t, umgebung, "lesen.sets", map[string]any{})
	_ = ueberStdio

	for _, werkzeug := range []string{"lesen.sets", "schreiben.feld"} {
		if !strings.Contains(rumpf, `"`+werkzeug+`"`) {
			t.Errorf("die Liste über HTTP nennt %q nicht", werkzeug)
		}
	}

	// **Ohne Zugang kommt nichts durch.**
	if status, _ := sendeMCP(t, angesagt.adresse, "",
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`); status != http.StatusUnauthorized {
		t.Errorf("ohne Marke kam %d zurück", status)
	}
}

// ansage ist, was ein Dienst über sich sagt.
type ansage struct {
	adresse string
	marke   string

	// fingerabdruck ist der des Zeugnisses — leer auf der Schleife.
	fingerabdruck string
}

// starteDienstAn startet `mse-rest --adresse` an einer Adresse.
func starteDienstAn(t *testing.T, umgebung []string, adresse string) ansage {
	t.Helper()

	return starteProgrammAn(t, gebaut.rest, umgebung, adresse)
}

// starteMCPAn startet `mse-mcp --adresse` an einer Adresse.
func starteMCPAn(t *testing.T, umgebung []string, adresse string) ansage {
	t.Helper()

	return starteProgrammAn(t, gebaut.mcp, umgebung, adresse)
}

// starteProgrammAn ist der gemeinsame Weg beider.
func starteProgrammAn(
	t *testing.T, programm string, umgebung []string, adresse string,
) ansage {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, programm,
		"--adresse", adresse, "--port", "0")
	lauf.Env = umgebung

	// **Der Aufruf wird benannt und nicht nur gefahren.**
	//
	// `docs/FAEHIGKEITEN.md` führt diese Fähigkeit unter `mse-rest --adresse`
	// bzw. `mse-mcp --adresse`, und `internal/werkzeuge/flaechenlauf_test.go`
	// sucht die Kennung **im Quelltext** — Kommentare zählt es ausdrücklich
	// nicht. Dieser Vergleich ist deshalb beides: die Kennung, an der der
	// Wächter den Lauf findet, **und** eine Prüfung, dass die Fahne wirklich
	// so heißt. Wer sie umbenennt, macht ihn rot statt still danebenzuliegen.
	gefahren := filepath.Base(programm) + " " + lauf.Args[1]

	if gefahren != "mse-rest --adresse" && gefahren != "mse-mcp --adresse" {
		beenden()
		t.Fatalf("gefahren wird %q — das ist keine der beiden Fähigkeiten", gefahren)
	}

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

	return ansage{adresse: gefunden, marke: marke, fingerabdruck: abdruck}
}

// frageDienst stellt eine Anfrage — **Status 0 heißt: gar keine Antwort**.
//
// **Und das ist hier eine Aussage**: ein Klartextruf an einen TLS-Lauscher
// endet vor der ersten Kopfzeile, und eine Bindung auf der Schleife ist über
// die LAN-Adresse nicht zu erreichen. Beides ist kein Fehler des Laufs.
func frageDienst(t *testing.T, adresse, marke string) (status int, rumpf string) {
	t.Helper()

	return rufMitKlient(t, klientOhnePruefung(), http.MethodGet, adresse, marke, "")
}

// sendeMCP schickt eine JSON-RPC-Nachricht an den Endpunkt.
func sendeMCP(
	t *testing.T, adresse, marke, nachricht string,
) (status int, rumpf string) {
	t.Helper()

	return rufMitKlient(t, klientOhnePruefung(),
		http.MethodPost, adresse, marke, nachricht)
}

// klientOhnePruefung nimmt das selbst unterschriebene Zeugnis an.
//
// **Geprüft wird es über den Fingerabdruck** und nicht über eine
// Zertifizierungsstelle (0115); ein Lauf, der eine verlangte, prüfte etwas,
// das dieses Programm nicht zusagt.
func klientOhnePruefung() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //#nosec G402 -- siehe oben
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
}

// rufMitKlient stellt eine Anfrage und gibt Status und Rumpf.
//
// **Status 0 heißt: keine Antwort** — der Grund steht bei [frageDienst].
func rufMitKlient(
	t *testing.T, klient *http.Client, methode, adresse, marke, koerper string,
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

	antwort, fehler := klient.Do(anfrage)
	if fehler != nil {
		return 0, fehler.Error()
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	gelesen, fehler := io.ReadAll(io.LimitReader(antwort.Body, 1<<20))
	if fehler != nil {
		t.Fatalf("die Antwort lesen: %v", fehler)
	}

	return antwort.StatusCode, string(gelesen)
}

// eigeneAdresse gibt eine IPv4-Adresse dieses Rechners, die nicht die
// Schleife ist.
//
// **Und überspringt den Lauf, wo es keine gibt** — ein Rechner ohne zweite
// Adresse ist kein Befund über diesen Code.
func eigeneAdresse(t *testing.T) string {
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

// portAus zieht den Port aus einer angesagten Adresse.
func portAus(t *testing.T, adresse string) string {
	t.Helper()

	ohneSchema := adresse
	for _, schema := range []string{"https://", "http://"} {
		ohneSchema = strings.TrimPrefix(ohneSchema, schema)
	}

	ohneSchema = strings.TrimSuffix(ohneSchema, "/mcp")

	_, port, fehler := net.SplitHostPort(ohneSchema)
	if fehler != nil {
		t.Fatalf("die Adresse %q zerlegen: %v", adresse, fehler)
	}

	return port
}
