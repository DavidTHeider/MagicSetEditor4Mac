package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege des Kartenlöschens**
// (#621).
//
// Paketintern und über den Stolpertreiber, aus demselben Grund wie überall
// hier: mit einer funktionierenden Datenbank sind sie nicht erreichbar, weil
// sie funktioniert.
//
// **Und sie gehören geprüft, weil ein halber Stand hier Datenverlust wäre.**
// Scheitert das Schreiben des Verlaufs und die Karte verschwindet trotzdem,
// ist sie ohne ⌘Z weg — genau der Fall, vor dem
// #562
// gewarnt hat.
func TestDasLoeschenMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht":  "",
		"das Lesen der Karte scheitert":  "stolper-fragen",
		"das Lesen der Felder scheitert": "stolper-lesen",
		"das Zeilenende scheitert":       "stolper-umwandeln",
		"das Schreiben scheitert":        "stolper-schreiben",
		"das zweite Schreiben scheitert": "stolper-zweites-schreiben",
		"der Verlauf scheitert":          "stolper-karte-verlauf",
		"die Feldabfrage scheitert":      "stolper-karte-feldabfrage",
		"ein Feld liest sich nicht":      "stolper-karte-feldscan",
		"das Feldende stolpert":          "stolper-karte-feldende",
		"das Wegschreiben scheitert":     "stolper-karte-schreiben",
		"der Commit scheitert":           "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if err := NeuerKartenschreiber(db).KarteLoeschen(probekontext(t), 1); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Ein unlesbarer Kartenweg ist ein Fehler und keine leere Karte.**
//
// # Warum das zählt
//
// `hinweg` und `rueckweg` sind JSON in einer Textspalte. Eine Zeile aus einer
// Sitzung, die abgebrochen ist, oder aus einer Fassung, die etwas anderes
// schrieb, kommt hier an — und eine leere Karte zurückzugeben hieße, eine
// Rücknahme auszuführen, die nichts wiederherstellt.
func TestEinUnlesbarerKartenwegIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, err := kartenstandAus("{kein JSON"); err == nil {
		t.Error("kein Fehler")
	}

	// **Ein leerer Weg ist keiner** — er heißt „auf dieser Seite gab es die
	// Karte nicht", und das ist der Hinweg jedes Löschens.
	stand, err := kartenstandAus("")
	if err != nil {
		t.Fatalf("der leere Weg: %v", err)
	}

	if stand.GabEs {
		t.Error("der leere Weg sagt, es habe die Karte gegeben")
	}
}

// **Und ein unlesbarer Schritt wird gemeldet, nicht übersprungen.**
func TestEinUnlesbarerKartenschrittWirdGemeldet(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, hin, zurueck string }{
		{name: "der Hinweg", hin: "{", zurueck: "{}"},
		{name: "der Rückweg", hin: "{}", zurueck: "{"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := kartenschrittAus(domain.Verlaufseintrag{
				Handlung: domain.HandlungKarte,
			}, fall.hin, fall.zurueck, 0, string(domain.UrheberFenster))
			if err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Die Wege der ganzen Karte melden ihre Fehler.**
//
// Über eine geschlossene Datenbank: jede Anweisung scheitert, und damit ist
// jeder der drei Zweige erreichbar — entfernen, anlegen, ein Feld anlegen.
func TestDieWegeDerGanzenKarteMeldenIhreFehler(t *testing.T) {
	// **Eine zurückgerollte Transaktion** gibt `ErrTxDone` auf jede Anweisung
	// — damit ist jeder der Zweige erreichbar.
	tx := abgebrocheneTransaktion(t)

	faelle := map[string]*domain.Kartenstand{
		"entfernen":  {GabEs: false},
		"anlegen":    {GabEs: true, Reihenfolge: 0},
		"ein Feld":   {GabEs: true, Felder: []domain.Feldstand{{Feld: "name"}}},
		"kein Stand": nil,
	}

	for name, stand := range faelle {
		t.Run(name, func(t *testing.T) {
			if err := ganzeKarteGehen(probekontext(t), tx, 1, 1, stand); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Ein doppeltes Feld im Rückweg scheitert am Primärschlüssel.**
//
// # Warum das kein erfundener Fall ist
//
// `hinweg` und `rueckweg` sind JSON in einer Textspalte. Eine Zeile aus einer
// abgebrochenen Sitzung oder aus einer Fassung, die anders schrieb, kann
// zweimal dasselbe Feld führen — und `card_values` trägt
// `PRIMARY KEY (card_id, feld)`.
//
// **Gemeldet, nicht übergangen**: eine halb zurückgebrachte Karte wäre
// schlimmer als eine, die nicht zurückkommt.
func TestEinDoppeltesFeldImRueckwegWirdGemeldet(t *testing.T) {
	db, err := Oeffnen(probekontext(t), ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // geschlossen ist geschlossen

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		 VALUES (1, 'P', 'magic', '2026-01-01T00:00:00Z', 1)`,
	); fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	tx, err := db.BeginTx(probekontext(t), nil)
	if err != nil {
		t.Fatalf("Transaktion: %v", err)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // der Test rollt zurück

	stand := &domain.Kartenstand{GabEs: true, Felder: []domain.Feldstand{
		{Feld: "name", Wert: "Erst"},
		{Feld: "name", Wert: "Nochmal"},
	}}

	if fehler := ganzeKarteGehen(probekontext(t), tx, 7, 1, stand); fehler == nil {
		t.Error("das doppelte Feld wurde nicht gemeldet")
	}
}

// abgebrocheneTransaktion gibt eine Transaktion, die nichts mehr ausführt.
func abgebrocheneTransaktion(t *testing.T) *sql.Tx {
	t.Helper()

	db, err := Oeffnen(probekontext(t), ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // geschlossen ist geschlossen

	tx, err := db.BeginTx(probekontext(t), nil)
	if err != nil {
		t.Fatalf("Transaktion: %v", err)
	}

	// **Zurückgerollt, nicht geschlossen**: eine zurückgerollte Transaktion
	// gibt `ErrTxDone` auf jede Anweisung, und genau das braucht der Test.
	if fehler := tx.Rollback(); fehler != nil {
		t.Fatalf("zurückrollen: %v", fehler)
	}

	return tx
}
