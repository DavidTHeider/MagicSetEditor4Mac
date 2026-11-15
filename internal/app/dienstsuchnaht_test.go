package app_test

// **Die Abnahme von Phase 4.8** — Werkstück 8
// (#1321),
// Abnahmekriterium 1: „Eine Kartensuche über magicthegathering.io liefert
// eine Karte."
//
// **Gemessen an der Naht, die alle vier Flächen benutzen** — nicht an einer
// Vorrichtung daneben. Wer hier durchkommt, kommt im Fenster, auf der
// Kommandozeile, über MCP und über den Dienst durch.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bundprobe gibt feste Geheimnisse.
type bundprobe struct{ werte map[string]string }

func (b bundprobe) Lies(_ context.Context, _, konto string) (string, error) {
	wert, gibts := b.werte[konto]
	if !gibts {
		return "", domain.ErrNotFound
	}

	return wert, nil
}

// fragewahlprobe sagt, welcher Dienst die Suche beantwortet.
type fragewahlprobe struct{ dienst domain.Dienst }

func (f fragewahlprobe) FuerFrage(_ context.Context, _ domain.Frage) (
	domain.Dienst, error,
) {
	return f.dienst, nil
}

// dienstwahlprobe schaltet genannte Dienste an.
type dienstwahlprobe struct{ an map[domain.Dienst]bool }

func (d dienstwahlprobe) IstAn(_ context.Context, dienst domain.Dienst) (
	bool, error,
) {
	return d.an[dienst], nil
}

// **Eine Suche über magicthegathering.io liefert eine Karte.**
//
// **Der Testserver antwortet in der Gestalt des echten Dienstes** — dieselbe,
// gegen die `internal/mtgapi` seinen Vertragstest fährt und die der Lauf
// hinter `netzlauf` am Original geprüft hat.
func TestEineSucheUeberDieMTGAPILiefertEineKarte(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, gestellt *http.Request) {
			anfragen.Add(1)

			if !strings.Contains(gestellt.URL.RawQuery, "name=Blitz") {
				t.Errorf("die Abfrage lautet %q", gestellt.URL.RawQuery)
			}

			if _, err := schreiber.Write([]byte(`{"cards":[{
				"id":"aaaa","name":"Blitzschlag","set":"PRB","setName":"Probeset",
				"manaCost":"{R}","type":"Instant","rarity":"Common","number":"141"
			}]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstMTGAPI: true}},
		bundprobe{},
		fragewahlprobe{dienst: domain.DienstMTGAPI})

	auskunft, suchfehler := quelle.Suche(probekontext(t), "Blitz")
	if suchfehler != nil {
		t.Fatalf("suchen: %v", suchfehler)
	}

	if len(auskunft.Treffer) != 1 {
		t.Fatalf("%d Treffer, erwartet 1", len(auskunft.Treffer))
	}

	eine := auskunft.Treffer[0]

	for _, fall := range []struct{ feld, gelesen, erwartet string }{
		{"Name", eine.Name, "Blitzschlag"},
		{"Kennung", eine.Kennung, "aaaa"},
		{"Setcode", eine.Setcode, "PRB"},
		{"Manakosten", eine.Manakosten, "{R}"},
	} {
		if fall.gelesen != fall.erwartet {
			t.Errorf("%s ist %q, erwartet %q", fall.feld, fall.gelesen, fall.erwartet)
		}
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der Dienst sah %d Anfragen, erwartet genau eine", gezaehlt)
	}
}

// **Ein Dienst ohne Zugang steht nicht bereit — und das wird gesagt.**
//
// Scrydex braucht zwei Angaben; fehlt eine, kommt er nicht in die Karte. Die
// Meldung nennt den gewählten Dienst, statt „nichts gefunden" zu sagen (§8).
func TestEinDienstOhneZugangStehtNichtBereit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{
			domain.DienstScryfall: true,
			domain.DienstScrydex:  true,
		}},
		// **Nur der Schlüssel, keine Team-Kennung.**
		bundprobe{werte: map[string]string{"scrydex.schluessel": "geheim"}},
		fragewahlprobe{dienst: domain.DienstScrydex})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if suchfehler == nil {
		t.Fatal("die Suche glückte, erwartet ein Fehler")
	}

	if !strings.Contains(suchfehler.Error(), "scrydex") {
		t.Errorf("die Meldung %q nennt den gewählten Dienst nicht", suchfehler)
	}
}

// **Steht gar kein Dienst bereit, sagt es die Suche.**
//
// **Nicht mehr die Naht** — Befund 1 des Zweitblicks über Phase 4.8. Sie
// entschied das beim Verdrahten, und das Fenster verdrahtet einmal, beim
// Start: wer danach seinen einzigen Dienst einschaltete, hatte bis zum
// Neustart gar keine Suche, weil die Naht gescheitert war und `a.suche` leer
// blieb. Jetzt kommt die Auskunft je Anfrage — und sie nennt den **gewählten**
// Dienst dazu.
func TestOhneJedenDienstSagtEsDieSuche(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{}},
		bundprobe{},
		fragewahlprobe{dienst: domain.DienstScryfall})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")
	if suchfehler == nil {
		t.Fatal("die Suche glückte, erwartet ein Fehler")
	}

	if !strings.Contains(suchfehler.Error(), "Einstellungen") {
		t.Errorf("die Meldung %q sagt nicht, wo man einen einschaltet",
			suchfehler)
	}
}
