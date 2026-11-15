package app_test

// **Kontostand und Fremdverlauf an der Naht** — Werkstück A5 von Phase 5.3
// (#1444).

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Dienst, der kein Kontingent nennt, wird abgewiesen.**
func TestEinDienstOhneKontingentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		Anbieterstand(probekontext(t), "cardmarket")

	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "cardmarket") {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Ohne Bestand gibt es keinen Stand.**
func TestOhneBestandGibtEsKeinenAnbieterstand(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").
		Anbieterstand(probekontext(t), "scrydex"); antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Stand")
	}
}

// **Ohne Zugang wird nicht gefragt.**
func TestOhneZugangGibtEsKeinenAnbieterstand(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{}).
		Anbieterstand(probekontext(t), "scrydex")

	if antwort.Fehler == nil {
		t.Fatal("ohne Zugang kam ein Stand")
	}
}

// **Und mit Zugang kommt er an — samt der Frist.**
func TestDerAnbieterstandKommtAnDerNahtAn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"total_credits":100000,"remaining_credits":84213,
				  "used_credits":15787,"overage_credit_rate":0.0004}`))
		}))
	t.Cleanup(server.Close)

	db := scrydexbestand(t, server.URL)

	antwort := app.AnbieterAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Anbieterstand(probekontext(t), "scrydex")

	if antwort.Fehler != nil {
		t.Fatalf("Anbieterstand: %+v", antwort.Fehler)
	}

	if antwort.Rest != 84213 || antwort.Ueberzugssatz != "0.0004" {
		t.Errorf("der Stand ist %+v", antwort.AnbieterstandDTO)
	}

	// **Die Frist steht in jeder Antwort** — ohne sie sähe der Rest aus wie
	// eine Sperre.
	if antwort.HoechstensAltMinuten != 30 {
		t.Errorf("die Frist ist %d Minuten", antwort.HoechstensAltMinuten)
	}
}

// **Eine Quelle, die es nicht gibt, wird abgewiesen.**
func TestEineUnbekannteVerlaufsquelleWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		Fremdpreisverlauf(probekontext(t), "cardmarket", "abc", "", nil)

	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "cardmarket") {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Eine Ausführung, die es nicht gibt, ebenso.**
func TestEineUnbekannteVerlaufsausfuehrungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		Fremdpreisverlauf(probekontext(t), "scrydex", "abc", "glaenzend", nil)

	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "glaenzend") {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Ohne Bestand gibt es keinen Verlauf.**
func TestOhneBestandGibtEsKeinenFremdverlauf(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").
		Fremdpreisverlauf(probekontext(t), "scrydex", "abc", "", nil); antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Verlauf")
	}
}

// **Ohne Zugang wird nicht gefragt.**
func TestOhneZugangGibtEsKeinenFremdverlauf(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		MitMarktbund(marktbund{}).
		Fremdpreisverlauf(probekontext(t), "scrydex", "abc", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ohne Zugang kam ein Verlauf")
	}
}

// **Und der ganze Weg trägt** — suchen, zuordnen, holen, ablegen.
func TestDerFremdverlaufGehtDenGanzenWeg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"data":[{"date":"2026-09-10","prices":[
					   {"variant":"nonfoil","condition":"NM","type":"raw",
					    "low":4.17,"market":5.02,"currency":"USD"}]}]}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	db := scrydexbestand(t, server.URL)

	anwendung := app.AnbieterAnwendungUeber(db).MitMarktbund(marktbund{marke: "geheim"})

	antwort := anwendung.Fremdpreisverlauf(probekontext(t), "scrydex",
		"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	// **Zwei Stände aus einer Zeile** — `low` und `market` sind zwei Reihen.
	if antwort.Staende != 2 || antwort.Fassung != "mtg-m11-149" || !antwort.Zugeordnet {
		t.Errorf("die Bilanz ist %+v", antwort.FremdpreisbilanzDTO)
	}

	if len(antwort.Befunde) != 0 {
		t.Errorf("der erste Abruf meldete %v", antwort.Befunde)
	}

	// **Der zweite Abruf sucht nicht noch einmal.**
	if zweite := anwendung.Fremdpreisverlauf(probekontext(t), "scrydex",
		"11111111-1111-4111-8111-111111111111", "", nil); zweite.Zugeordnet {
		t.Errorf("der zweite Abruf hat noch einmal zugeordnet: %+v", zweite.Befunde)
	}
}

// scrydexbestand legt einen Druck an, schaltet Scrydex ein und lenkt um.
func scrydexbestand(t *testing.T, adresse string) *sql.DB {
	t.Helper()

	zurueck := app.MitDienstadressen(adresse, adresse, adresse)
	t.Cleanup(zurueck)

	db := frischerBestand(t)

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
		domain.DienstEinstellung(domain.DienstScrydex), domain.DienstAn,
	); fehler != nil {
		t.Fatalf("den Dienst anschalten: %v", fehler)
	}

	return db
}

// **Ein Fehler beim Kontostand reist weiter.**
func TestEinFehlerBeimKontostandReistWeiter(t *testing.T) {
	antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, kaputteGegenstelle(t))).
		MitMarktbund(marktbund{marke: "geheim"}).
		Anbieterstand(probekontext(t), "scrydex")

	if antwort.Fehler == nil {
		t.Fatal("ein 500 kam als Stand zurück")
	}
}

// **Ein Befund des Bezugs reist bis ans Fenster.**
//
// Er entsteht, wo die Quelle etwas nennt, das niemand deuten kann — hier ein
// Tag, der keiner ist. **Er hält den Bezug nicht auf** (§8): die übrigen
// Stände kommen an, und der Befund steht daneben.
func TestEinBefundDesBezugsReistBisAnsFenster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"data":[{"date":"vorgestern","prices":[
					   {"variant":"nonfoil","condition":"NM","type":"raw",
					    "low":4.17,"market":5.02,"currency":"USD"}]}]}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, server.URL)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	if antwort.Staende != 0 {
		t.Errorf("%d Stände abgelegt, erwartet keiner", antwort.Staende)
	}

	if len(antwort.Befunde) != 1 || !strings.Contains(antwort.Befunde[0], "vorgestern") {
		t.Errorf("die Befunde sind %v", antwort.Befunde)
	}
}

// **Ein klemmender Bund hält auch den Verlauf an.**
func TestEinKlemmenderBundHaeltDenFremdverlaufAn(t *testing.T) {
	t.Parallel()

	antwort := app.AnbieterAnwendungUeber(frischerBestand(t)).
		MitMarktbund(klemmenderBund{}).
		Fremdpreisverlauf(probekontext(t), "scrydex", "abc", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein klemmender Bund gab einen Verlauf")
	}
}
