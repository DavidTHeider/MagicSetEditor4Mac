package werkzeuge_test

// **Wer einen HTTP-Klienten baut** — Werkstück 4 von A18
// (#1533).
//
// # Warum es diesen Wächter gibt
//
// **Weil die Eigenschaften eines Transports dort entstehen, wo er gebaut
// wird.** Proxy, HTTP/2 und die drei Fristen sitzen in
// `netz.NeuerKlient`; wer daneben einen eigenen `http.Client` baut, bekommt
// sie nicht — und niemand sieht es, weil beide Wege funktionieren.
//
// **Gemessen am 2026-09-12**: `internal/netz` las `HTTPS_PROXY` nicht und
// sprach nur HTTP/1.1, weil sein handgebauter Transport `Proxy` und
// `ForceAttemptHTTP2` nicht erbt. Dieselbe Falle steht jedem offen, der einen
// zweiten baut.
//
// # Was er hält
//
// **Jede Stelle, die einen `http.Client`, einen `http.Transport` oder Gos
// Vorgaben benutzt, steht in [httpausnahmen] — mit einem Grund.** Eine neue
// erzwingt damit die Entscheidung, statt still einen zweiten Transport mit
// anderen Eigenschaften aufzumachen.
//
// # Was er nicht hält
//
// **Ob die Eigenschaften der Ausnahme die richtigen sind.** Er sieht, dass
// jemand eine Entscheidung getroffen hat, nicht dass sie gut war. Dass Gos
// Vorgabetransport Proxy und h2 mitbringt, steht im Grund der Ausnahme und
// ist nachgemessen, nicht von hier.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// httpausnahmen nennt jede Datei, die einen eigenen Klienten oder Transport
// baut — samt Grund.
//
// **Eine Aufzählung und keine Regel über Verzeichnisse**, aus demselben Grund
// wie bei `namensweg` in `einstellungstexte_test.go`: wer einen zweiten Weg
// ins Netz aufmacht, entscheidet hier — der Lauf wird sonst rot und nennt die
// Datei.
var httpausnahmen = map[string]string{
	"internal/netz/klient.go": "**die eine Stelle**, an der Fristen, Proxy, " +
		"Protokoll und Drossel entstehen — der Grund für die Aufzählung und " +
		"nicht ihre Ausnahme",

	"internal/anmeldung/rufer.go": "**der Anmeldefluss ist ein eigener Weg** " +
		"und steht als solcher schon in `internal/schichten/netz_test.go`: " +
		"er horcht zusätzlich auf einem eigenen Port, und ein Horcher ist " +
		"kein Klient. **Er baut auf Gos Vorgabetransport**, und der bringt " +
		"`ProxyFromEnvironment` und `ForceAttemptHTTP2` mit — nachgemessen " +
		"am 2026-09-12. Was ihm fehlt, ist die Dreiteilung der Fristen; er " +
		"hat eine Gesamtfrist, und das steht in seinem eigenen Kopf",
	"internal/netz/socketklient.go": "**Apples Gerätemodell über einen " +
		"Unix-Socket** (AI-05, #2042) — ohne Proxy, ohne Weiterleitung und " +
		"ohne Zielsperre, denn er geht nicht ins Netz: ein Proxy schickte " +
		"eine Anfrage an diesen Rechner hinaus. Die Fristen sind dieselben " +
		"`Fristen` wie bei `NeuerKlient`; das steht in seinem Kopf und in " +
		"`docs/NETZEIGENSCHAFTEN.md`",
}

// httpmuster findet jede Stelle, die einen Klienten oder Transport baut.
var httpmuster = regexp.MustCompile(
	`http\.Client\{|http\.Transport\{|http\.DefaultClient|http\.DefaultTransport`)

// **Jede Stelle, die einen HTTP-Klienten baut, ist eingetragen.**
func TestJederHTTPKlientStehtInDerAufzaehlung(t *testing.T) {
	t.Parallel()

	gefunden := httpstellen(t)

	for _, datei := range gefunden {
		if _, eingetragen := httpausnahmen[datei]; !eingetragen {
			t.Errorf("%s baut einen eigenen HTTP-Klienten oder Transport "+
				"und steht nicht in `httpausnahmen` — dann hat er andere "+
				"Eigenschaften als `netz.NeuerKlient`, und niemand hat das "+
				"entschieden", datei)
		}
	}

	// **Die Reichweite der Messung** (§10): ein Ausdruck, der nichts findet,
	// wäre grün. Die eine Stelle, die es geben **muss**, ist `internal/netz`.
	if !stehtInDerListe(gefunden, "internal/netz/klient.go") {
		t.Fatal("der Ausdruck findet nicht einmal `internal/netz/klient.go` " +
			"— dann sagt dieser Lauf nichts")
	}
}

// **Und jeder Eintrag nennt eine Datei, die es gibt — mit einem Grund.**
//
// **Die andere Richtung** (§10): eine Aufzählung, deren Einträge ins Leere
// zeigen, entschärft den Lauf darüber, ohne dass jemand es sieht.
func TestJedeHTTPAusnahmeIstEineWirklicheStelle(t *testing.T) {
	t.Parallel()

	gefunden := httpstellen(t)

	for datei, grund := range httpausnahmen {
		if !stehtInDerListe(gefunden, datei) {
			t.Errorf("`httpausnahmen` nennt %s — dort steht kein eigener "+
				"Klient mehr; der Eintrag entschärft nichts und sieht aus, "+
				"als täte er es", datei)
		}

		if strings.TrimSpace(grund) == "" {
			t.Errorf("%s steht ohne Grund in der Aufzählung — ein Eintrag "+
				"ohne ihn ist eine Ausnahme, die niemand nachprüft", datei)
		}
	}
}

// httpstellen sucht jede Quelldatei, die einen Klienten oder Transport baut.
//
// **Ohne Testdateien.** Ein Lauf darf einen eigenen Klienten bauen — er geht
// gegen `httptest` und nicht ins Netz.
func httpstellen(t *testing.T) []string {
	t.Helper()

	gefunden := make([]string, 0, 8)

	for _, wurzel := range []string{"internal", "cmd"} {
		gefunden = append(gefunden, httpstellenUnter(t,
			filepath.Join("..", "..", wurzel), wurzel)...)
	}

	return gefunden
}

// httpstellenUnter durchsucht einen Baum.
func httpstellenUnter(t *testing.T, ort, gezeigt string) []string {
	t.Helper()

	gefunden := make([]string, 0, 8)

	fehler := filepath.Walk(ort, func(pfad string, eintrag os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
			strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Baum im Repositorium
		if lesefehler != nil {
			return lesefehler
		}

		// **Der Quelltext ohne Kommentare.** Mehrere Köpfe erklären, warum
		// `http.DefaultClient` **nicht** benutzt wird — wer den Text roh
		// läse, zählte die Erklärung als Verstoß.
		if httpmuster.MatchString(ohneKommentare(string(roh))) {
			gefunden = append(gefunden, gezeigt+
				strings.TrimPrefix(filepath.ToSlash(pfad),
					filepath.ToSlash(ort)))
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchsuchen: %v", ort, fehler)
	}

	return gefunden
}

// stehtInDerListe sagt, ob eine Datei gefunden wurde.
//
// **Ein eigener Name**, weil `enthaelt` im Paket schon vergeben ist
// (`readmewerkzeuge_test.go`) — und `ohneKommentare` ebenso
// (`datenschutz_test.go`), das hier deshalb mitbenutzt wird.
func stehtInDerListe(liste []string, datei string) bool {
	for _, eintrag := range liste {
		if eintrag == datei {
			return true
		}
	}

	return false
}
