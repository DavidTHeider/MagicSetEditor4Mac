package anmeldung_test

// **Nur der passende Rückruf zählt, und er zählt einmal** — SEC-18
// (#2152).
//
// Der Horcher nahm den **ersten** Ruf, gleich welchen: ein `POST`, das
// `favicon.ico` des Browsers oder ein Rückruf mit fremdem `state` verbrauchten
// das eine Mal, und der richtige kam danach nicht mehr an. Und der Browser
// las „Angemeldet", bevor der Code eingelöst war.

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

func rufe(t *testing.T, methode, adresse string) string {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(), methode, adresse, http.NoBody)
	if fehler != nil {
		t.Fatal(fehler)
	}

	antwort, fehler := http.DefaultClient.Do(anfrage)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	roh, _ := io.ReadAll(io.LimitReader(antwort.Body, 4096)) //nolint:errcheck // der Rumpf ist eine Auskunft

	return string(roh)
}

func TestNurDerPassendeRueckrufZaehlt(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // nur ein Lauf

	horcher.Erwarte("s")

	basis := strings.TrimSuffix(horcher.Adresse, "/")

	for _, falsch := range []struct{ methode, weg string }{
		{http.MethodPost, "/?code=c&state=s"},
		{http.MethodGet, "/favicon.ico"},
		{http.MethodGet, "/?code=c&state=fremd"},
		{http.MethodGet, "/?code=c"},
		{http.MethodGet, "/?state=s"},
		{http.MethodGet, "/anders?code=c&state=s"},
	} {
		rufe(t, falsch.methode, basis+falsch.weg)
	}

	kurz, ende := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer ende()

	if werte, wartefehler := horcher.Warte(kurz); wartefehler == nil {
		t.Fatalf("ein falscher Rückruf wurde verbraucht: %v", werte)
	}

	seite := rufe(t, http.MethodGet, basis+"/?code=c&state=s")

	// **Vor dem Tausch nur „Antwort erhalten"** — angemeldet ist erst, wer
	// den Code eingelöst hat.
	if strings.Contains(seite, "Angemeldet") || !strings.Contains(seite, "Antwort") {
		t.Errorf("der Browser liest: %q", seite)
	}

	werte, wartefehler := horcher.Warte(t.Context())
	if wartefehler != nil || werte.Get("code") != "c" {
		t.Fatalf("der passende Rückruf: %v (%v)", werte, wartefehler)
	}

	// **Ein zweiter passender blockiert nichts.**
	rufe(t, http.MethodGet, basis+"/?code=c&state=s")
}

func TestEineAbsageDesAnbietersZaehlt(t *testing.T) {
	t.Parallel()

	horcher, fehler := anmeldung.Horche()
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // nur ein Lauf

	horcher.Erwarte("s")
	rufe(t, http.MethodGet, strings.TrimSuffix(horcher.Adresse, "/")+"/?error=access_denied&state=s")

	if werte, wartefehler := horcher.Warte(t.Context()); wartefehler != nil || werte.Get("error") != "access_denied" {
		t.Fatalf("die Absage: %v (%v)", werte, wartefehler)
	}
}
