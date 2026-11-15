package app_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// satzquelleprobe ist die Naht zum Einlesen.
type satzquelleprobe struct {
	code     string
	spiel    string
	ergebnis *service.Ergebnis
	fehler   error
}

func (s *satzquelleprobe) Einlesen(_ context.Context, code, spiel string) (
	*service.Ergebnis, error,
) {
	s.code, s.spiel = code, spiel

	return s.ergebnis, s.fehler
}

// **Ein ganzer Satz über die Fassade** —
// #1033.
func TestScryfallSatzEinlesen(t *testing.T) {
	naht := &satzquelleprobe{ergebnis: &service.Ergebnis{
		SetID: 9, Karten: 103, Angelegt: []int64{9},
		Befunde: []domain.Befund{{Ort: "cards/search", Meldung: "gekürzt"}},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitScryfallsaetzen(naht)

	antwort := anwendung.ScryfallSatzEinlesen(t.Context(), "lea", "magic")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if naht.code != "lea" || naht.spiel != "magic" {
		t.Errorf("die Naht bekam %q/%q", naht.code, naht.spiel)
	}

	if antwort.SetID != "9" || antwort.Karten != 103 {
		t.Errorf("Antwort = %+v", antwort)
	}

	if len(antwort.Angelegt) != 1 || len(antwort.Befunde) != 1 {
		t.Errorf("Angelegt = %v, Befunde = %v", antwort.Angelegt, antwort.Befunde)
	}
}

// **Ein leerer Satzcode ist ein Abbruch und kein Fehler** — dieselbe
// Übereinkunft wie bei einem abgebrochenen Dateidialog.
func TestEinLeererSatzcodeIstEinAbbruch(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitScryfallsaetzen(&satzquelleprobe{})

	antwort := anwendung.ScryfallSatzEinlesen(t.Context(), "", "magic")
	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("Antwort = %+v", antwort)
	}
}

// **Ohne eingerichteten Zugang gibt es eine Meldung.**
func TestScryfallSatzOhneZugang(t *testing.T) {
	if antwort := app.NeueApp("").ScryfallSatzEinlesen(
		probekontext(t), "lea", "magic"); antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein Abbruch gibt das Ergebnis mit** (#155): was durchkam, steht im
// Bestand und gehört benannt.
func TestEinAbbruchGibtDasSatzergebnisMit(t *testing.T) {
	naht := &satzquelleprobe{
		ergebnis: &service.Ergebnis{SetID: 9, Karten: 3, Angelegt: []int64{9}},
		fehler:   errors.New("der Bestand ist nicht erreichbar"),
	}

	anwendung := app.NeueApp("")
	anwendung.MitScryfallsaetzen(naht)

	antwort := anwendung.ScryfallSatzEinlesen(t.Context(), "lea", "magic")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.SetID != "9" {
		t.Errorf("SetID = %q — was angelegt wurde, gehört benannt", antwort.SetID)
	}
}

// **Und der ganze Weg vom Satzcode bis ins Set.**
//
// **`HOME` und die Adresse, beides** (§3) — ohne das erste sieht der Lauf den
// Zwischenspeicher dieses Rechners, ohne das zweite geht er wirklich ins Netz.
func TestDerGanzeWegVomSatzcodeBisInsSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/cards/search" {
				if _, err := schreiben.Write([]byte(`{
					"object": "list", "has_more": false, "total_cards": 1,
					"data": [{
						"object": "card",
						"id": "00000000-0000-4000-8000-000000000009",
						"name": "Prüfstein", "layout": "normal",
						"mana_cost": "{2}{W}", "type_line": "Kreatur — Prüfling",
						"rarity": "rare",
						"scryfall_uri": "https://beispiel.invalid/karte/9"
					}]
				}`)); err != nil {
					t.Errorf("schreiben: %v", err)
				}

				return
			}

			if _, err := schreiben.Write([]byte(
				`{"object":"set","code":"prb","name":"Der Probesatz"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)

	naht, fehler := app.ScryfallsatznahtUeber(db)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ergebnis, einlesefehler := naht.Einlesen(probekontext(t), "prb", "magic")
	if einlesefehler != nil {
		t.Fatalf("Einlesen: %v", einlesefehler)
	}

	if ergebnis.SetID == 0 || ergebnis.Karten != 1 {
		t.Fatalf("Ergebnis = %+v", ergebnis)
	}

	// **Nachgesehen im Bestand** — das ist der Punkt.
	danach, _, standfehler := app.Kartenstand(probekontext(t), db, ergebnis.SetID,
		ersteKarte(t, db, ergebnis.SetID))
	if standfehler != nil {
		t.Fatalf("Kartenstand: %v", standfehler)
	}

	if danach["name"] != "Prüfstein" || danach["casting cost"] != "2W" {
		t.Errorf("im Bestand steht %v", danach)
	}

	// **Und ein zweiter Aufruf legt nichts Neues an** (#886).
	nochmal, nochmalfehler := naht.Einlesen(probekontext(t), "prb", "magic")
	if nochmalfehler != nil {
		t.Fatalf("zweites Einlesen: %v", nochmalfehler)
	}

	if !nochmal.SchonOffen || nochmal.SetID != ergebnis.SetID {
		t.Errorf("zweites Ergebnis = %+v", nochmal)
	}
}

// ersteKarte gibt die Kennung der ersten Karte eines Sets.
func ersteKarte(t *testing.T, db *sql.DB, setID int64) int64 {
	t.Helper()

	var kennung int64
	if err := db.QueryRowContext(probekontext(t),
		`SELECT id FROM cards WHERE set_id = ? ORDER BY reihenfolge LIMIT 1`,
		setID).Scan(&kennung); err != nil {
		t.Fatalf("die erste Karte lesen: %v", err)
	}

	return kennung
}

// **Ohne Zuhause kommt die Satznaht nicht zusammen.**
func TestOhneZuhauseKommtDieSatznahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.ScryfallsatznahtUeber(frischerBestand(t)); fehler == nil {
		t.Fatal("kein Fehler ohne Zuhause")
	}
}

// **Ein Fehler vor dem Anlegen gibt kein Ergebnis** — und die Naht macht
// daraus keinen halben Stand.
func TestEinFehlerVorDemAnlegenGibtKeinErgebnis(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, fehler := app.ScryfallsatznahtUeber(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ergebnis, einlesefehler := naht.Einlesen(probekontext(t), "gibbesnicht", "magic")
	if einlesefehler == nil {
		t.Fatal("kein Fehler")
	}

	if ergebnis.SetID != 0 || len(ergebnis.Befunde) != 0 {
		t.Errorf("Ergebnis = %+v, erwartet ein leeres", ergebnis)
	}
}
