package app_test

// **Der ganze Weg zum Marktplatz** — Werkstück A4 von Phase 5.3
// (#1443).
//
// # Warum das hier steht
//
// **Weil der Weg durch vier Schichten geht**: Bestand → Auflösung über
// Erweiterungen und Blueprints → Abruf → Ablage der Zuordnung. Keine Fläche
// ohne Fenster legt einen Druck an, und `internal/cardtrader` kennt keinen
// Bestand — die Naht dazwischen ist prüfbar nur hier (§10, Nähte).

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Druck wird aufgelöst, die Angebote kommen, und die Zuordnung bleibt.**
func TestDerMarktabrufLoestAufUndMerktSich(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	server := marktgegenstelle(t)
	t.Setenv(app.CardTraderUmgebung, server.URL)

	anwendung := app.MarktAnwendungUeber(db).MitMarktbund(marktbund{marke: "geheim"})

	antwort := anwendung.Marktangebote(probekontext(t), "cardtrader",
		"11111111-1111-4111-8111-111111111111", "", "", nil)

	if antwort.Fehler != nil {
		t.Fatalf("Marktangebote: %+v", antwort.Fehler)
	}

	if antwort.Fassung != "89" || !antwort.Zugeordnet {
		t.Errorf("die Antwort ist %+v", antwort.MarktantwortDTO)
	}

	if len(antwort.Angebote) != 1 {
		t.Fatalf("die Angebote sind %+v", antwort.Angebote)
	}

	erstes := antwort.Angebote[0]
	if erstes.Wert != "2.17" || erstes.Waehrung != "EUR" ||
		erstes.Zustand != "Moderately Played" || erstes.Verkaeufer != "einer" {
		t.Errorf("das erste Angebot ist %+v", erstes)
	}

	// **Und der Befund der Quelle reist mit** (§8).
	if len(antwort.Befunde) != 1 {
		t.Errorf("die Befunde sind %v", antwort.Befunde)
	}

	// **Der zweite Abruf löst nicht noch einmal auf** — das ist die Zusage,
	// für die die Zuordnung überhaupt abgelegt wird.
	zweite := anwendung.Marktangebote(probekontext(t), "cardtrader",
		"11111111-1111-4111-8111-111111111111", "", "", nil)

	if zweite.Fehler != nil {
		t.Fatalf("der zweite Abruf: %+v", zweite.Fehler)
	}

	if zweite.Zugeordnet {
		t.Error("der zweite Abruf hat noch einmal aufgelöst")
	}
}

// **Ein Set, das der Marktplatz nicht führt, ist »nicht gefunden«.**
func TestEinUnbekanntesSetIstNichtGefunden(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	server := marktgegenstelle(t)
	t.Setenv(app.CardTraderUmgebung, server.URL)

	// **Ein Druck aus einem Set, das die Gegenstelle nicht kennt.**
	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Fremd", Setcode: "xyz", Nummer: "2", Sprache: "en",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: "33333333-3333-4333-8333-333333333333",
		}},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader",
			"33333333-3333-4333-8333-333333333333", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein unbekanntes Set kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "xyz") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Ein abgeschalteter Dienst hält den Abruf auf** (§16, Regel 6).
func TestEinAbgeschalteterMarktplatzWirdNichtGefragt(t *testing.T) {
	db := frischerBestand(t)
	marktbestand(t, db)

	// **Der Schalter wird ausdrücklich zurückgenommen** — die Vorgabe ist
	// ohnehin „aus", und ein Lauf, der sich darauf verlässt, prüfte die
	// Vorgabe statt den Schalter.
	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstCardTrader), domain.DienstAus,
	); fehler != nil {
		t.Fatalf("den Dienst abschalten: %v", fehler)
	}

	server := marktgegenstelle(t)
	t.Setenv(app.CardTraderUmgebung, server.URL)

	antwort := app.MarktAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Marktangebote(probekontext(t), "cardtrader",
			"11111111-1111-4111-8111-111111111111", "", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("der abgeschaltete Dienst wurde gefragt")
	}
}

// marktbestand legt einen Druck an und schaltet CardTrader ein.
func marktbestand(t *testing.T, db *sql.DB) {
	t.Helper()

	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Probe", Setcode: "ptkdf", Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: "11111111-1111-4111-8111-111111111111",
		}},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.DienstEinstellung(domain.DienstCardTrader), domain.DienstAn,
	); fehler != nil {
		t.Fatalf("den Dienst anschalten: %v", fehler)
	}
}

// marktgegenstelle antwortet wie CardTrader — Erweiterungen, Blueprints, Markt.
func marktgegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	antworten := map[string]string{
		"/expansions": `[{"id": 92, "game_id": 1, "code": "ptkdf", "name": "Tarkir"},
		                 {"id": 7, "game_id": 4, "code": "ptkdf", "name": "Fremdes Spiel"}]`,
		"/blueprints/export": `[
		  {"id": 89, "name": "Probe", "game_id": 1, "expansion_id": 92,
		   "scryfall_id": "11111111-1111-4111-8111-111111111111",
		   "card_market_ids": [], "tcg_player_id": null},
		  {"id": 0, "name": "ohne Kennung", "game_id": 1, "expansion_id": 92,
		   "scryfall_id": null, "card_market_ids": [], "tcg_player_id": null}]`,
		"/marketplace/products": `{"89": [{
		  "id": 101862104, "blueprint_id": 89, "quantity": 3,
		  "price": {"cents": 217, "currency": "EUR"},
		  "properties_hash": {"condition": "Moderately Played",
		                      "mtg_foil": false, "mtg_language": "en"},
		  "user": {"username": "einer", "country_code": "FI"},
		  "bundle_size": 1}]}`,
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			inhalt, bekannt := antworten[gestellt.URL.Path]
			if !bekannt {
				schreiben.WriteHeader(http.StatusNotFound)

				return
			}

			_, _ = schreiben.Write([]byte(inhalt)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// marktbund gibt eine feste Marke — ohne den echten Schlüsselbund (§3).
type marktbund struct{ marke string }

func (m marktbund) Lies(_ context.Context, _, _ string) (string, error) {
	return m.marke, nil
}
