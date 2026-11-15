// Der Weg zum Dienst auf der Rückschleife
// (#997).
//
// # Warum er einen eigenen Helfer braucht
//
// Kommandozeile und MCP sind Prozesse, die etwas tun und enden. Der Dienst
// **läuft**, und ein Test muss ihn starten, seine Marke abwarten, ihn
// befragen und danach beenden. Das ist genug für eine eigene Datei.
package flaechen_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// dienst ist ein laufender `mse-rest` samt seiner Adresse und Marke.
type dienst struct {
	adresse string
	marke   string
	klient  *http.Client
}

// starteDienst baut einen laufenden Dienst über einem Bestand.
//
// **Er endet mit dem Test.** `t.Cleanup` schickt ihm ein Signal und wartet;
// ein Dienst, der einen Test überlebte, hielte den Bestand und ließe den
// nächsten scheitern.
func starteDienst(t *testing.T, umgebung []string) dienst {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.rest, "-port", "0")
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if err := lauf.Start(); err != nil {
		beenden()
		t.Fatalf("den Dienst starten: %v", err)
	}

	// **Ein früher Tod wird gemeldet, nicht abgewartet.**
	//
	// Der erste Anlauf wartete stets bis zur 30-Sekunden-Frist, auch wenn der
	// Prozess sofort endete — und meldete dann „hat nichts angesagt" statt
	// „ist gestorben". Der Nachlauf zum Zweitblick über Phase U hat es
	// angestrichen.
	// **Geschlossen statt gesendet.** Der erste Anlauf schickte den Fehler in
	// einen Kanal — und `ansageLesen` nahm ihn heraus, wenn der Dienst starb.
	// Danach wartete `t.Cleanup` auf einen Wert, der schon weg war, und der
	// Lauf hing bis zur Zeitgrenze. Gefunden von der Gegenprobe zu genau
	// diesem Fall.
	//
	// Ein **geschlossener** Kanal gibt jedem Leser sofort etwas, beliebig oft.
	geendet := make(chan struct{})

	go func() {
		// **Der Ausgang wird nicht gelesen.** Beendet wird über ein Signal,
		// und `Wait` meldete dann genau das; was zählt, ist **dass** der
		// Prozess weg ist — und das sagt der geschlossene Kanal.
		_ = lauf.Wait() //nolint:errcheck,gosec // siehe oben
		close(geendet)
	}()

	t.Cleanup(func() {
		beenden()
		<-geendet
	})

	adresse, marke, _ := ansageLesen(t, angesagt, geendet)

	return dienst{
		adresse: adresse, marke: marke,
		klient: &http.Client{Timeout: 2 * time.Minute},
	}
}

// ruf fragt den Dienst und gibt Status und Rumpf.
//
// **Die Marke geht mit, wenn es eine gibt.** Ein leerer Wert heißt: ohne — so
// lässt sich dieselbe Hilfsfunktion für die Abweisung benutzen.
func (d dienst) ruf(
	t *testing.T, methode, weg, rumpf string,
) (status int, antwort string) {
	t.Helper()

	var koerper io.Reader = http.NoBody
	if rumpf != "" {
		koerper = strings.NewReader(rumpf)
	}

	anfrage, err := http.NewRequestWithContext(t.Context(), methode,
		d.adresse+weg, koerper)
	if err != nil {
		t.Fatalf("Anfrage bauen: %v", err)
	}

	if d.marke != "" {
		anfrage.Header.Set("Authorization", "Bearer "+d.marke)
	}

	kam, fehler := d.klient.Do(anfrage)
	if fehler != nil {
		t.Fatalf("%s %s: %v", methode, weg, fehler)
	}

	defer func() {
		if err := kam.Body.Close(); err != nil {
			t.Errorf("den Rumpf schließen: %v", err)
		}
	}()

	roh, lesefehler := io.ReadAll(kam.Body)
	if lesefehler != nil {
		t.Fatalf("den Rumpf lesen: %v", lesefehler)
	}

	return kam.StatusCode, string(roh)
}

// kopfruf ist [dienst.ruf] samt der Kopfzeilen der Antwort.
//
// **Für die Vorgangskennung** — Werkstück O9 von Phase 5.2 (#1433): der
// Dienst trägt sie im Kopf und nicht im Rumpf, und ob sie wirklich an
// **jeder** Antwort steht, sagt nur ein Lauf gegen den gebauten Dienst.
func (d dienst) kopfruf(
	t *testing.T, methode, weg, rumpf string,
) (status int, koepfe http.Header) {
	t.Helper()

	var koerper io.Reader = http.NoBody
	if rumpf != "" {
		koerper = strings.NewReader(rumpf)
	}

	anfrage, err := http.NewRequestWithContext(t.Context(), methode,
		d.adresse+weg, koerper)
	if err != nil {
		t.Fatalf("Anfrage bauen: %v", err)
	}

	if d.marke != "" {
		anfrage.Header.Set("Authorization", "Bearer "+d.marke)
	}

	kam, fehler := d.klient.Do(anfrage)
	if fehler != nil {
		t.Fatalf("%s %s: %v", methode, weg, fehler)
	}

	if schliessfehler := kam.Body.Close(); schliessfehler != nil {
		t.Errorf("den Rumpf schließen: %v", schliessfehler)
	}

	return kam.StatusCode, kam.Header
}

// ohneMarke gibt denselben Dienst ohne Marke — für die Abweisung.
func (d dienst) ohneMarke() dienst {
	d.marke = ""

	return d
}

// mitNachweis gibt denselben Dienst mit einem anderen Nachweis.
//
// **Für die Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351):
// ein Token steht in derselben Kopfzeile wie die Prozessmarke, und ob es
// wirklich an ihrer Stelle gilt, sagt nur ein Lauf gegen den **gebauten**
// Dienst.
func (d dienst) mitNachweis(wert string) dienst {
	d.marke = wert

	return d
}

// ansageLesen wartet, bis der Dienst Adresse und Marke gesagt hat.
//
// **Oder bis er stirbt.** Ein Prozess, der sofort endet, sagt nie etwas an;
// dreißig Sekunden darauf zu warten und dann „hat nichts angesagt" zu melden,
// verschweigt den eigentlichen Grund.
func ansageLesen(
	t *testing.T, angesagt *mitschrift, geendet <-chan struct{},
) (adresse, marke, fingerabdruck string) {
	t.Helper()

	frist := time.After(30 * time.Second)

	for {
		select {
		case <-geendet:
			t.Fatalf("der Dienst ist gestorben: %q", angesagt.text())
		case <-frist:
			t.Fatalf("der Dienst hat nichts angesagt: %q", angesagt.text())
		default:
		}

		// **Gesucht wird je Vorsatz und nicht je Zeilennummer.**
		//
		// Seit N6 (#1572) steht jenseits der Schleife eine dritte Zeile
		// dazwischen — der Fingerabdruck. Ein Leser, der `zeilen[1]` nimmt,
		// hielte ihn dann für die Marke und wäre grün dabei.
		//
		// **Fertig ist die Ansage mit „Marke"**, denn sie steht zuletzt;
		// das ist die Zusage von [app.Dienstansage].
		adresse, marke, fingerabdruck = "", "", ""

		for _, zeile := range strings.Split(
			strings.TrimSpace(angesagt.text()), "\n") {
			// **Das Schema bleibt dran** — der Dienst sagt `https://` an, wo
			// er ein Zeugnis trägt, und ein Leser, der `http://`
			// abschneidet, ließe bei TLS ein `s://` stehen.
			//
			// **Wer nur Wirt und Port braucht**, nimmt [ohneSchemaUndWeg].
			if wert, ja := strings.CutPrefix(zeile, "Dienst auf "); ja {
				adresse = wert
			}

			if wert, ja := strings.CutPrefix(zeile, "Fingerabdruck "); ja {
				fingerabdruck = wert
			}

			if wert, ja := strings.CutPrefix(zeile, "Marke "); ja {
				marke = wert
			}
		}

		if adresse != "" && marke != "" {
			return adresse, marke, fingerabdruck
		}
	}
}

// mitschrift sammelt, was ein Prozess sagt — gegen gleichzeitiges Lesen gesichert.
type mitschrift struct {
	schranke sync.Mutex
	inhalt   bytes.Buffer
}

func (m *mitschrift) Write(roh []byte) (int, error) {
	m.schranke.Lock()
	defer m.schranke.Unlock()

	return m.inhalt.Write(roh) //nolint:wrapcheck // ein Puffer scheitert nicht
}

func (m *mitschrift) text() string {
	m.schranke.Lock()
	defer m.schranke.Unlock()

	return m.inhalt.String()
}

// ausRest liest ein Feld aus einer JSON-Antwort.
func ausRest(t *testing.T, rumpf, feld string) string {
	t.Helper()

	var gelesen map[string]any
	if err := json.Unmarshal([]byte(rumpf), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, rumpf)
	}

	wert, gefunden := gelesen[feld]
	if !gefunden {
		t.Fatalf("die Antwort nennt %q nicht: %s", feld, rumpf)
	}

	if text, istText := wert.(string); istText {
		return text
	}

	return ""
}
