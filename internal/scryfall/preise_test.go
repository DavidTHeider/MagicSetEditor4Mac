// **Die Preise einer Karte** — Werkstück 3 von Phase 3.4
// (#1165).
//
// **Geprüft wird gegen eingecheckte, selbst geschriebene Beispiele** (§16):
// die Gestalt ist die der Dokumentation, der Inhalt ist eigener. Ein Lauf
// gegen den echten Dienst steht hinter `netzlauf`.
package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die sieben Felder werden ihrer Währung und Ausführung zugeordnet.**
//
// # Diese Antwort ist erfunden, und das gehört dazugesagt
//
// Sie trägt alle sieben Felder aus der Dokumentation. **Kein heute abgerufener
// Druck trägt `eur_etched`** — nachgemessen an acht Antworten am 2026-08-27,
// die Zahlen stehen bei [scryfall.LiesPreise]. Dieser Lauf prüft deshalb, was
// die **Zuordnung** tut, wenn ein Feld dasteht; dass es dasteht, prüft er
// nicht und behauptet es auch nicht.
//
// **Die Gegenseite steht darunter**: eine Antwort in der Gestalt, die der
// Dienst heute wirklich schickt.
func TestDieSiebenPreisfelderWerdenZugeordnet(t *testing.T) {
	t.Parallel()

	_, preise, befunde, err := scryfall.LiesPreise([]byte(`{
		"id": "probe-druck",
		"prices": {
			"usd": "4.17", "usd_foil": "12.00", "usd_etched": "20.50",
			"eur": "3.99", "eur_foil": "11.00", "eur_etched": "18.25",
			"tix": "0.03"
		}
	}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}

	// **Der Markt steht in derselben Zeile wie die Währung** (A3, #1442):
	// `tix` ist Magic Online, alles andere Papier.
	erwartet := []scryfall.Preis{
		{Waehrung: "USD", Finish: domain.FinishNonfoil, Markt: domain.MarktPapier, Betrag: 417},
		{Waehrung: "USD", Finish: domain.FinishFoil, Markt: domain.MarktPapier, Betrag: 1200},
		{Waehrung: "USD", Finish: domain.FinishEtched, Markt: domain.MarktPapier, Betrag: 2050},
		{Waehrung: "EUR", Finish: domain.FinishNonfoil, Markt: domain.MarktPapier, Betrag: 399},
		{Waehrung: "EUR", Finish: domain.FinishFoil, Markt: domain.MarktPapier, Betrag: 1100},
		{Waehrung: "EUR", Finish: domain.FinishEtched, Markt: domain.MarktPapier, Betrag: 1825},
		{Waehrung: "TIX", Finish: domain.FinishNonfoil, Markt: domain.MarktOnline, Betrag: 3},
	}

	if len(preise) != len(erwartet) {
		t.Fatalf("es sind %d Preise: %+v", len(preise), preise)
	}

	for i, soll := range erwartet {
		if preise[i] != soll {
			t.Errorf("an Stelle %d steht %+v statt %+v", i, preise[i], soll)
		}
	}
}

// **Die Kennung des Drucks kommt mit.**
//
// Ohne sie ist nicht zu sagen, zu welchem Druck die Preise gehören.
func TestDiePreiseTragenIhrenDruck(t *testing.T) {
	t.Parallel()

	kennung, _, _, err := scryfall.LiesPreise(
		[]byte(`{"id": "probe-druck", "prices": {"eur": "1.00"}}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if kennung != "probe-druck" {
		t.Errorf("die Kennung ist %q", kennung)
	}
}

// **`null` heißt „kein Preis bekannt" und ist kein Fehler.**
//
// Bei den meisten Karten sind `usd_etched` und `tix` genau das.
func TestEinFehlenderPreisIstKeinFehler(t *testing.T) {
	t.Parallel()

	_, preise, befunde, err := scryfall.LiesPreise([]byte(`{
		"id": "probe-druck",
		"prices": {"usd": "4.17", "usd_foil": null, "tix": null}
	}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if len(preise) != 1 || len(befunde) != 0 {
		t.Errorf("%d Preise, %d Befunde: %+v", len(preise), len(befunde), preise)
	}
}

// **Und `"0.00"` ist ein Preis, kein fehlender.**
//
// „Kostet nichts" und „kein Preis bekannt" sind zwei Aussagen; wer sie
// zusammenwirft, verliert eine.
func TestEinNullpreisIstEinPreis(t *testing.T) {
	t.Parallel()

	_, preise, _, err := scryfall.LiesPreise(
		[]byte(`{"id": "probe-druck", "prices": {"usd": "0.00"}}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if len(preise) != 1 || preise[0].Betrag != 0 {
		t.Errorf("es kam %+v", preise)
	}
}

// **Ein unlesbarer Preis ist ein Befund und bricht nichts** (§8, §16).
//
// # Warum das die Drifterkennung ist
//
// Ändert die Quelle ihre Schreibweise — ein Komma statt eines Punktes, ein
// Währungszeichen davor —, sagt es dieser Weg **laut**, statt still eine Null
// einzutragen. Und die übrigen Preise kommen trotzdem an.
func TestEinUnlesbarerPreisIstEinBefund(t *testing.T) {
	t.Parallel()

	_, preise, befunde, err := scryfall.LiesPreise([]byte(`{
		"id": "probe-druck",
		"prices": {"usd": "4,17", "eur": "3.99"}
	}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("es gab %d Befunde: %+v", len(befunde), befunde)
	}

	if befunde[0].Ort != "usd" || befunde[0].Datei != "probe-druck" {
		t.Errorf("der Befund zeigt auf %q in %q", befunde[0].Ort, befunde[0].Datei)
	}

	// **Und der lesbare ist trotzdem da.** 300 lesbare Karten und zwei
	// Befunde, nicht null Karten.
	if len(preise) != 1 || preise[0].Waehrung != "EUR" {
		t.Errorf("es kam %+v", preise)
	}
}

// **Ohne `prices` gibt es keine Preise und keinen Fehler.**
//
// Ein Token hat keine Marktdaten; das ist eine Auskunft und kein Fehlschlag.
func TestOhnePreiseGibtEsKeinenFehler(t *testing.T) {
	t.Parallel()

	_, preise, befunde, err := scryfall.LiesPreise([]byte(`{"id": "probe-druck"}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	// **Eine leere Liste und kein `nil`** (§3).
	if preise == nil || len(preise) != 0 || len(befunde) != 0 {
		t.Errorf("%v, %v", preise, befunde)
	}
}

// **Ohne `id` bricht es — und die Meldung sagt, welches Feld fehlt** (§16).
func TestOhneKennungBrichtEs(t *testing.T) {
	t.Parallel()

	_, _, _, err := scryfall.LiesPreise([]byte(`{"prices": {"usd": "1.00"}}`))

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Und kaputtes JSON ebenso.**
func TestKaputtesJSONBrichtBeimPreislesen(t *testing.T) {
	t.Parallel()

	_, _, _, err := scryfall.LiesPreise([]byte(`{"id":`))

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Die Preise eines Drucks kommen über seine Kennung.**
//
// # Warum über die Kennung und nicht über den Namen
//
// Ein Name führt auf irgendeinen Druck; ein Preis gehört zu **einem**.
func TestDiePreiseEinesDrucksKommenUeberSeineKennung(t *testing.T) {
	t.Parallel()

	var gefragt string

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt = anfrage.URL.Path

			if _, err := schreiben.Write([]byte(
				`{"id":"abc","prices":{"eur":"3.99"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	preise, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "abc")
	if err != nil {
		t.Fatalf("PreiseFuerDruck: %v", err)
	}

	// **Der Weg ist `/cards/:id`**, nachgelesen am 2026-08-27. Ohne diese
	// Zusicherung ginge auch ein Suchweg durch, der zufällig dieselbe Karte
	// findet.
	if gefragt != "/cards/abc" {
		t.Errorf("gefragt wurde %q", gefragt)
	}

	if len(preise) != 1 || preise[0].Wert != 399 {
		t.Errorf("es kam %+v", preise)
	}
}

// **Ein unlesbarer Preis ist eine ungeprüfte Reihe und kein Wegfall** —
// F04 (#2246).
//
// Die Quelle hat für `eur` etwas gesagt, nur nicht in einer Form, die sich
// deuten lässt. Fehlte die Reihe in der Antwort, widerriefe die Preispflege
// sie als „weggefallen" — ein Lesefehler machte aus einem gültigen Stand eine
// Aussage, die Scryfall nie getroffen hat. Als [domain.PreisUngeprueft]
// bleibt der letzte bekannte Wert gültig.
func TestEinUnlesbarerPreisKommtAlsUngepruefteReihe(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"id":"abc","prices":{"usd":"4.17","eur":"3,99"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	staende, _, befunde, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "abc")
	if err != nil {
		t.Fatalf("PreiseFuerDruck: %v", err)
	}

	if len(befunde) != 1 {
		t.Errorf("es gab %d Befunde: %+v", len(befunde), befunde)
	}

	lagen := map[string]domain.Preislage{}
	for _, stand := range staende {
		lagen[stand.Waehrung] = stand.Lage
	}

	if lagen["USD"] != domain.PreisGenannt || lagen["EUR"] != domain.PreisUngeprueft {
		t.Errorf("die Lagen sind %v — erwartet USD genannt und EUR ungeprüft", lagen)
	}
}

// **Ohne Kennung wird gar nicht erst gefragt.**
//
// Ein leerer Weg `/cards/` ist eine andere Abfrage, und die Antwort darauf
// hätte mit dem gemeinten Druck nichts zu tun.
func TestOhneKennungWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	gefragt := false

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) { gefragt = true }))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "   ")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}

	if gefragt {
		t.Error("es ging trotzdem eine Anfrage hinaus")
	}
}

// **Eine Antwort, die kein JSON ist, kommt als Fehler durch.**
func TestEineKaputteAntwortBeimPreisabrufKommtDurch(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(`{"id":`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "abc")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Und eine Gegenstelle, die 404 sagt, ebenso.**
func TestEinUnbekannterDruckMeldetSich(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "gibtesnicht")

	if err == nil {
		t.Error("der unbekannte Druck kam durch")
	}
}

// **Eine Antwort über einen anderen Druck wird abgewiesen.**
//
// # Warum das eine eigene Zusage ist
//
// Die Kennung wird gelesen und dann verglichen. Ohne den Vergleich landete
// eine vertauschte oder beschädigte Antwort aus dem Zwischenspeicher als Preis
// des **angefragten** Drucks im Bestand — still, und niemand fände es je.
func TestEineAntwortUeberEinenAnderenDruckWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"id":"ein-anderer","prices":{"eur":"1.00"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "gefragt")

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Der Klient sagt, wer er ist** — Werkstück A3 von Phase 5.3 (#1442).
//
// **Damit ein leerer Abruf trotzdem eine Aussage ist**: „Scryfall nennt für
// diesen Druck nichts mehr" lässt sich aus null Ständen nicht ablesen, und der
// Dienst braucht den Namen, um **nur** die eigenen Reihen zu widerrufen.
func TestDerKlientNenntSeineQuelle(t *testing.T) {
	t.Parallel()

	klient := scryfall.NeuerKlient("https://example.invalid",
		netz.NeuerKlient(netz.UeblicheFristen()), nil, time.Now, "prüfung")

	if name := klient.Name(); name != domain.PreisquelleScryfall {
		t.Errorf("der Klient nennt sich %q", string(name))
	}
}
