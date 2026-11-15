package scryfall_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Druck aus dem Vorrat, ohne zu fragen** — P10
// (#2378).
//
// Eine Handelszeile zeigt Name, Set und Bild ihres Drucks. Was dieser Rechner
// schon einmal über `/cards/:id` gelesen hat — beim Preisholen, beim Abgleich
// —, liegt unter demselben Schlüssel; **eine Anfrage geht dafür nicht
// hinaus** (§16, Regel 5).
func TestKarteImVorratFragtNicht(t *testing.T) {
	t.Parallel()

	const kennung = "00000000-0000-4000-8000-000000000001"

	var anfragen atomic.Int32

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			// **Für jede Kennung dieselbe Karte** — damit fällt auf, wer
			// eine Antwort einem anderen Druck zuschreibt.
			if _, err := schreiben.Write(beispiel(t, "einfach.json")); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	// **Ohne Kennung ist nichts gemeint** — auch kein Vorrat.
	if _, gabEs := klient.KarteImVorrat("  "); gabEs {
		t.Fatal("eine leere Kennung hat einen Vorrat")
	}

	if _, gabEs := klient.KarteImVorrat(kennung); gabEs || anfragen.Load() != 0 {
		t.Fatalf("ohne Vorrat: gab es %v nach %d Anfragen", gabEs, anfragen.Load())
	}

	if _, _, _, fehler := klient.KarteNachKennung(t.Context(), kennung); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	karte, gabEs := klient.KarteImVorrat("  " + kennung + "  ")
	if !gabEs || karte.Name != "Prüfstein" || karte.Set != "prb" || karte.Nummer != "1" {
		t.Fatalf("aus dem Vorrat: %v, %+v", gabEs, karte)
	}

	if anfragen.Load() != 1 {
		t.Errorf("%d Anfragen — der Vorrat fragte selbst", anfragen.Load())
	}

	// **Eine Antwort, die einen anderen Druck nennt, ist kein Vorrat für
	// diesen**: sonst stünde an der Zeile das Bild einer anderen Karte.
	const fremd = "00000000-0000-4000-8000-0000000000ff"
	if _, _, _, fehler := klient.KarteNachKennung(t.Context(), fremd); fehler == nil {
		t.Fatal("die vertauschte Antwort ging durch")
	}

	if karte, gabEs := klient.KarteImVorrat(fremd); gabEs {
		t.Errorf("unter %s liegt %q und kam als Vorrat", fremd, karte.Name)
	}
}

// **Was eine Suche gezeigt hat, ist Vorrat** — gefunden von der Hüllenprobe
// `p10handel09` (#2378): ein Druck, den jemand über „Karte" gewählt hatte,
// stand danach als „Unbekannter Druck" mit seiner Kennung an der Zeile. Die
// Suchantwort lag nur unter dem Schlüssel der Suche.
func TestEineSucheLegtIhreDruckeInDenVorrat(t *testing.T) {
	t.Parallel()

	const kennung = "00000000-0000-4000-8000-000000000001"

	var anfragen atomic.Int32

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			if _, err := schreiben.Write([]byte(`{"object":"list","has_more":false,"data":[` +
				string(beispiel(t, "einfach.json")) + `]}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.Druckesuche(t.Context(), `!"Prüfstein"`); fehler != nil {
		t.Fatalf("suchen: %v", fehler)
	}

	karte, gabEs := klient.KarteImVorrat(kennung)
	if !gabEs || karte.Name != "Prüfstein" || karte.Bilder.Klein == "" {
		t.Fatalf("aus dem Vorrat: %v, %+v", gabEs, karte)
	}

	if anfragen.Load() != 1 {
		t.Errorf("%d Anfragen — der Vorrat fragte selbst", anfragen.Load())
	}

	// **Der Preisweg bleibt, wie er war**: er fragt `/cards/:id`, auch wenn
	// die Suche den Druck schon kannte.
	if _, _, _, fehler := klient.PreiseFuerDruck(t.Context(), kennung); fehler == nil || anfragen.Load() != 2 {
		t.Errorf("der Preisweg nach der Suche: %d Anfragen, Fehler %v", anfragen.Load(), fehler)
	}
}
