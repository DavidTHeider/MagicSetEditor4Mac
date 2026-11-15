package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der ganze Weg vom Namen bis zur Karte im Bestand** —
// #1032.
//
// # Warum das über einen echten Bestand geht
//
// Weil die Naht **drei** Dinge verbindet, die je für sich geprüft sind: das
// Spiel des Sets, den Abgleich und den Lebenszyklus. Was zwischen ihnen
// liegt, sieht kein Test der einzelnen Teile.
//
// **`HOME` und die Adresse, beides** (§3): ohne das erste sieht der Lauf den
// Zwischenspeicher dieses Rechners, ohne das zweite geht er wirklich ins Netz.
// Beides einzeln reicht nicht — in Phase V hat ein Bindungstest zweimal
// 25 MB von Scryfall geladen (#888).
func TestDerGanzeWegVomNamenBisZurKarte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(`{
				"object": "card",
				"id": "00000000-0000-4000-8000-000000000002",
				"name": "Prüfstein",
				"layout": "normal",
				"mana_cost": "{2}{W}",
				"type_line": "Kreatur — Prüfling",
				"oracle_text": "Fliegend.",
				"rarity": "rare",
				"scryfall_uri": "https://beispiel.invalid/karte/2"
			}`)); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)
	setID, _ := setMitKarte(t, db)

	naht, fehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ergebnis, anlegefehler := naht.Anlegen(probekontext(t), setID, "Prüfstein")
	if anlegefehler != nil {
		t.Fatalf("Anlegen: %v", anlegefehler)
	}

	if ergebnis.KarteID == 0 {
		t.Fatal("keine Kartenkennung — ohne sie ist der Aufruf nicht umkehrbar")
	}

	if ergebnis.Herkunft != "https://beispiel.invalid/karte/2" {
		t.Errorf("Herkunft = %q", ergebnis.Herkunft)
	}

	// **Nachgesehen im Bestand** — das ist der Punkt: was hier steht, sehen
	// Vorschau, Ausfuhr, MCP und die Kommandozeile.
	danach, _, standfehler := app.Kartenstand(probekontext(t), db, setID, ergebnis.KarteID)
	if standfehler != nil {
		t.Fatalf("Kartenstand: %v", standfehler)
	}

	if danach["name"] != "Prüfstein" || danach["rule text"] != "Fliegend." {
		t.Errorf("im Bestand steht %v", danach)
	}
}

// **Ein Set, das es nicht gibt, legt nichts an.**
func TestDasAnlegenWeistEinUnbekanntesSetAb(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	vorher := app.MitScryfalladresse("http://127.0.0.1:1")
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)

	naht, fehler := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if _, anlegefehler := naht.Anlegen(probekontext(t), 999, "Prüfstein"); anlegefehler == nil {
		t.Fatal("kein Fehler bei einem Set, das es nicht gibt")
	}
}

// **Und `Setspiel` über einem echten Bestand.**
func TestSetspielUeberEinemBestand(t *testing.T) {
	db := frischerBestand(t)
	setID, _ := setMitKarte(t, db)

	spiel, fehler := app.Setspiel(probekontext(t), db, setID)
	if fehler != nil {
		t.Fatalf("Setspiel: %v", fehler)
	}

	if spiel == "" {
		t.Error("kein Spiel — ohne es gibt es keine Zuordnung auf Scryfall-Felder")
	}
}
