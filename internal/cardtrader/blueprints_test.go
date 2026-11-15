package cardtrader_test

// **Blueprints und ihre Fremdkennungen** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Blueprint bringt seine Kennungen mit — und das Array wird zu mehreren.**
//
// `card_market_ids` ist ein Array, und das ist kein Zufall: ein Blueprint kann
// mehreren Cardmarket-Produkten entsprechen (0095). Wer daraus **eine**
// Kennung machte, verlöre die zweite still.
func TestEinBlueprintTraegtSeineKennungen(t *testing.T) {
	t.Parallel()

	server := gegenstelle(t, map[string]string{"/blueprints/export": "blueprints.json"})

	blueprints, befunde, fehler := klientFuer(server).Blueprints(t.Context(), 8)
	if fehler != nil {
		t.Fatalf("Blueprints: %v", fehler)
	}

	if len(blueprints) != 2 {
		t.Fatalf("%d Blueprints, erwartet 2 — der dritte hat keine Kennung", len(blueprints))
	}

	erster := blueprints[0]
	if erster.ID != 89 || erster.Name != "Beast // Plant" || erster.Fassung != "" {
		t.Errorf("der erste Blueprint ist %+v", erster)
	}

	// **Die Scryfall-Kennung kommt klein an** — dieselbe Kennung soll dieselbe
	// Zeichenkette sein ([domain.Druckkennung.Normalisieren]).
	erwartet := []domain.Druckkennung{
		{System: domain.SystemCardTrader, Semantik: domain.SemantikDruck, Wert: "89"},
		{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: "381854e2-7369-473e-a604-4dd7c010fc89",
		},
		{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "362677"},
		{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "362678"},
	}

	if len(erster.Kennungen) != len(erwartet) {
		t.Fatalf("%d Kennungen, erwartet %d: %+v",
			len(erster.Kennungen), len(erwartet), erster.Kennungen)
	}

	for i, soll := range erwartet {
		if erster.Kennungen[i] != soll {
			t.Errorf("Kennung %d ist %+v, erwartet %+v", i, erster.Kennungen[i], soll)
		}
	}

	// **Und ein `null` ist keine Kennung.** Der zweite hat keine Scryfall-ID
	// und kein Cardmarket-Array, aber eine TCGplayer-Nummer.
	zweiter := blueprints[1]
	if zweiter.Fassung != "Retro" {
		t.Errorf("die Fassung des zweiten ist %q", zweiter.Fassung)
	}

	if len(zweiter.Kennungen) != 2 ||
		zweiter.Kennungen[1] != (domain.Druckkennung{
			System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt, Wert: "77712",
		}) {
		t.Errorf("die Kennungen des zweiten sind %+v", zweiter.Kennungen)
	}

	// **Ein Blueprint ohne jede Kennung ist ein Befund und kein Schweigen** (§8).
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Ohne Erweiterung wird nicht gefragt.**
//
// **Nachgelesen** (0095): `expansion_id` ist Pflicht, und ohne sie antwortet
// CardTrader mit 404. Eine Anfrage, von der feststeht, dass sie scheitert,
// wird nicht gestellt — das ist dieselbe Sparsamkeit wie §16, Satz 4.
func TestOhneErweiterungWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Blueprints(t.Context(), 0); fehler == nil {
		t.Fatal("ohne Erweiterung kam eine Antwort")
	}

	if gefragt {
		t.Error("der Server wurde ohne Erweiterung gefragt")
	}
}

// **Die Erweiterung steht in der Abfrage.**
func TestDieErweiterungStehtInDerAbfrage(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			gesehen <- gestellt.URL.RawQuery
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Blueprints(t.Context(), 92); fehler != nil {
		t.Fatalf("Blueprints: %v", fehler)
	}

	if abfrage := <-gesehen; abfrage != "expansion_id=92" {
		t.Errorf("die Abfrage ist %q", abfrage)
	}
}

// **Eine kaputte Antwort ist ein Fehler und keine leere Liste.**
func TestEineKaputteAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte("{nicht json")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := cardtrader.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), cardtrader.Zugang{Marke: "geheim"})

	if _, _, fehler := klient.Blueprints(t.Context(), 8); fehler == nil {
		t.Fatal("eine kaputte Antwort kam als Erfolg zurück")
	}
}
