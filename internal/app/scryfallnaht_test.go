package app_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Naht für die Flächen ohne Fenster kommt über einem echten Bestand
// zusammen** (#1005).
//
// # Was hier geprüft wird
//
// Die **Verdrahtung**: dass aus einer `*sql.DB` etwas Benutzbares wird und
// dass die Kartenprüfung greift. Was ein Vorschlag fachlich ist, hält
// `internal/service`; wie er über ein Protokoll geht, das jeweilige
// Flächenpaket.
//
// **Ohne Netz** — der Abgleich wird gebaut, aber nur mit Kennungen gefüttert,
// die schon vorher scheitern.
func TestDieNahtKommtUeberEinemBestandZusammen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	naht, fehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if naht == nil {
		t.Fatal("die Naht ist leer")
	}

	// **Eine Karte, die es nicht gibt, scheitert vor dem Netz.**
	if _, err := naht.Vorschlag(probekontext(t), 1, 4711, "Prüfstein"); err == nil {
		t.Error("die fehlende Karte kam durch")
	}

	if _, err := naht.Uebernehmen(probekontext(t), 1, 4711, "Prüfstein", nil); err == nil {
		t.Error("die fehlende Karte kam bei der Übernahme durch")
	}
}

// **Ohne Zuhause kommt sie nicht zusammen.**
func TestOhneZuhauseKommtDieAbgleichsnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Scryfallabgleichnaht(nil, domain.UrheberKommandozeile); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Und der Kartenstand weist eine Karte aus einem anderen Set ab.**
//
// # Warum das die eigentliche Wache ist
//
// Ohne sie käme die Zuordnung vom **falschen Spiel** — und ein Vorschlag
// schriebe Werte in eine Karte, die zu einem anderen Set gehört.
func TestDerKartenstandWeistEineFremdeKarteAb(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	if _, _, fehler := app.Kartenstand(probekontext(t), db, 4711, 4712); fehler == nil {
		t.Error("die fremde Karte kam durch")
	}
}

// frischerBestand gibt eine migrierte Datenbank.
func frischerBestand(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	return db
}

// **Der ganze Weg, einmal gefahren: Bestand → Dienst → Klient → Gegenstelle**
// (#1005).
//
// # Warum dieser Test sein muss
//
// Alles darunter ist mit Doubles geprüft: die Fassade gegen eine Quelle, der
// Dienst gegen eine Fremdquelle, der Klient gegen eine Gegenstelle. **Keiner
// davon zeigt, dass die vier zusammenpassen** — und genau dort entstand in
// Phase 5 der Fehler, den kein Test sah (#130).
//
// Gefahren wird gegen eine eigene Gegenstelle, nicht gegen Scryfall: §11, und
// ein Testlauf, der fremde Rechner befragt, ist keiner.
func TestDerGanzeWegVomBestandBisZurGegenstelle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(`{
				"object": "card",
				"id": "00000000-0000-4000-8000-000000000001",
				"name": "Prüfstein",
				"layout": "normal",
				"mana_cost": "{2}{W}",
				"type_line": "Kreatur — Prüfling",
				"oracle_text": "Fliegend.",
				"rarity": "rare",
				"scryfall_uri": "https://beispiel.invalid/karte/1"
			}`)); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)
	setID, karteID := setMitKarte(t, db)

	naht, fehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	vorschlag, holfehler := naht.Vorschlag(probekontext(t), setID, karteID, "Prüfstein")
	if holfehler != nil {
		t.Fatalf("Vorschlag: %v", holfehler)
	}

	if len(vorschlag.Aenderungen) == 0 {
		t.Fatal("der Vorschlag ist leer")
	}

	if vorschlag.Herkunft != "https://beispiel.invalid/karte/1" {
		t.Errorf("Herkunft %q", vorschlag.Herkunft)
	}

	// **Und die Übernahme landet wirklich im Bestand.**
	uebernommen, schreibfehler := naht.Uebernehmen(probekontext(t), setID, karteID,
		"Prüfstein", []string{"rule text"})
	if schreibfehler != nil {
		t.Fatalf("Übernehmen: %v", schreibfehler)
	}

	if len(uebernommen) != 1 || uebernommen[0].Feld != "rule text" {
		t.Fatalf("übernommen wurde %+v", uebernommen)
	}

	// **Nachgesehen im Bestand** — das ist der Punkt: was hier steht, sehen
	// Vorschau, Ausfuhr, MCP und die Kommandozeile.
	danach, _, standfehler := app.Kartenstand(probekontext(t), db, setID, karteID)
	if standfehler != nil {
		t.Fatalf("Kartenstand: %v", standfehler)
	}

	if danach["rule text"] != "Fliegend." {
		t.Errorf("im Bestand steht %q", danach["rule text"])
	}
}

// **Und ein Feld außerhalb des Vorschlags wird auch hier abgewiesen.**
func TestEinFremdesFeldWirdAuchUeberDieNahtAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"object":"card","id":"x","name":"Prüfstein","layout":"normal"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)
	setID, karteID := setMitKarte(t, db)

	naht, nahtfehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, fehler := naht.Uebernehmen(probekontext(t), setID, karteID,
		"Prüfstein", []string{"power"}); fehler == nil {
		t.Error("das fremde Feld wurde übernommen")
	}
}

// setMitKarte legt ein Set mit einer Karte an.
func setMitKarte(t *testing.T, db *sql.DB) (setID, karteID int64) {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "magic"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karteID, kartenfehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID)
	if kartenfehler != nil {
		t.Fatalf("Karte anlegen: %v", kartenfehler)
	}

	return setID, karteID
}

// **Ein Ausfall der Gegenstelle kommt durch — bei beiden Schritten.**
func TestEinAusfallDerGegenstelleKommtUeberDieNahtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)
	setID, karteID := setMitKarte(t, db)

	naht, fehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if _, err := naht.Vorschlag(probekontext(t), setID, karteID, "x"); err == nil {
		t.Error("der Ausfall kam beim Vorschlag nicht durch")
	}

	if _, err := naht.Uebernehmen(probekontext(t), setID, karteID, "x", nil); err == nil {
		t.Error("der Ausfall kam bei der Übernahme nicht durch")
	}
}

// **Und ein Speicherort, der keiner sein kann, ebenso.**
//
// # Wie er hergestellt wird
//
// Eine **Datei** dort, wo das Verzeichnis hinsoll. `os.MkdirAll` scheitert
// daran, und die Naht kommt nicht zusammen — ohne dass irgendetwas am Bestand
// falsch wäre.
func TestEinSpeicherortDerKeinerSeinKannLaesstDieNahtScheitern(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	// Erst den Ort bestimmen lassen, dann eine Datei dorthin legen.
	ort, fehler := app.Scryfallverzeichnis()
	if fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	if err := os.RemoveAll(ort); err != nil {
		t.Fatalf("aufräumen: %v", err)
	}

	if err := os.WriteFile(ort, []byte("keine Verzeichnisdatei"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := app.Scryfallabgleichnaht(frischerBestand(t),
		domain.UrheberKommandozeile); err == nil {
		t.Error("die Naht kam über einer Datei zusammen")
	}
}
