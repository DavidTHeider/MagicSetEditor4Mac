// **Die Antworten der Symbolwege** — Werkstück
// #1391.
//
// # Was `endpunktliste_test.go` nicht prüft
//
// Dort stehen die Wege mit ihren vier Klassen — Marke, Herkunft, gültiger
// Aufruf, Drossel. Was der gültige Aufruf **antwortet**, prüft er nicht: er
// zählt, dass die Naht gerufen wurde. Hier steht die andere Hälfte, samt den
// Fehlerwegen, die kein Endpunktlauf betritt.
package rest_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// symboldienst gibt einen Dienst mit einer willigen Symbolwerkstatt.
func symboldienst(t *testing.T) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Symbol: williger{}, Schreiber: williger{}})
}

// **Jeder Weg gibt, was die Naht liefert.**
func TestDieSymbolwegeReichenIhreAntwortDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		methode  string
		weg      string
		rumpf    string
		erwartet string
	}{
		"lesen":   {http.MethodGet, "/sets/1/symbol", "", `"teile"`},
		"ablegen": {http.MethodPut, "/sets/1/symbol", `{"teile":[],"unter":"x"}`, `probe.mse-symbol`},
		"rechnen": {http.MethodPost, "/symbol/grundform", `{"art":"Rechteck"}`, `gerechnet`},
		"palette": {http.MethodGet, "/karten/2/symbolpalette/name", "", `"W"`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, symboldienst(t), endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf,
				erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code != http.StatusOK {
				t.Fatalf("Status %d (%s)", antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), fall.erwartet) {
				t.Errorf("die Antwort enthält %s nicht: %s",
					fall.erwartet, antwort.Body.String())
			}
		})
	}
}

// **Und jeder weist eine unbrauchbare Kennung ab.**
//
// Die Gegenprobe zum Lauf darüber: ohne sie wäre er auch dann grün, wenn ein
// Weg jede Kennung annähme und der Naht eine Null reichte.
func TestDieSymbolwegePruefenIhreKennungen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		methode string
		weg     string
		rumpf   string
	}{
		"lesen":    {http.MethodGet, "/sets/keine-zahl/symbol", ""},
		"ablegen":  {http.MethodPut, "/sets/keine-zahl/symbol", `{"teile":[]}`},
		"palette":  {http.MethodGet, "/karten/keine-zahl/symbolpalette/name", ""},
		"felder":   {http.MethodPost, "/sets/keine-zahl/felder", `{"karteIDs":[1],"feld":"a","wert":"b"}`},
		"vorschau": {http.MethodPost, "/sets/keine-zahl/felder/vorschau", `{"karteIDs":[1],"feld":"a","wert":"b"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, symboldienst(t), endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf,
				erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code == http.StatusOK {
				t.Errorf("die Kennung kam durch: %s", antwort.Body.String())
			}
		})
	}
}

// **Ein Rumpf, der kein JSON ist, wird abgewiesen.**
func TestEinUnlesbarerSymbolrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"ablegen":  "/sets/1/symbol",
		"felder":   "/sets/1/felder",
		"vorschau": "/sets/1/felder/vorschau",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			methode := http.MethodPut
			if name == "felder" || name == "vorschau" {
				methode = http.MethodPost
			}

			antwort := ruf(t, symboldienst(t), endpunkt{
				methode: methode, weg: weg, rumpf: `{kein json`,
				erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status %d, erwartet 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und ein Fehler der Naht kommt durch.**
//
// `symbolRechnen` liest seinen Rumpf roh — ohne `gelesen`, weil die acht
// Rechnungen acht Argumentmengen haben. Der Fehlerweg liegt damit **in der
// Naht**, und dieser Lauf ist der einzige, der ihn betritt.
func TestEinFehlerDerSymbolnahtKommtDurch(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	boese := unwilliger{fehler: fmt.Errorf("das Symbol: %w", domain.ErrNotFound)}

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Symbol: boese})

	for name, fall := range map[string]struct {
		methode string
		weg     string
		rumpf   string
	}{
		"lesen":   {http.MethodGet, "/sets/1/symbol", ""},
		"ablegen": {http.MethodPut, "/sets/1/symbol", `{"teile":[]}`},
		"rechnen": {http.MethodPost, "/symbol/grundform", `{"art":"Rechteck"}`},
		"palette": {http.MethodGet, "/karten/2/symbolpalette/name", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf,
				erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code == http.StatusOK {
				t.Errorf("der Fehler kam nicht durch: %s", antwort.Body.String())
			}
		})
	}
}

// **Ein Rumpf über einem Mebibyte wird abgewiesen, bevor er gelesen ist.**
//
// `symbolRechnen` liest roh und ohne feste Gestalt — die Grenze ist damit das
// Einzige, was den Speicher schützt. Ein Symbol ist viel kleiner; wer ein
// Mebibyte schickt, meint etwas anderes.
func TestEinZuGrosserSymbolrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	const einMebibyte = 1 << 20

	antwort := ruf(t, symboldienst(t), endpunkt{
		methode: http.MethodPost, weg: "/symbol/grundform",
		rumpf:  `{"art":"` + strings.Repeat("x", einMebibyte+1) + `"}`,
		erfolg: http.StatusOK,
	}, gueltig)

	if antwort.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Status %d, erwartet 413: %s", antwort.Code, antwort.Body.String())
	}
}
