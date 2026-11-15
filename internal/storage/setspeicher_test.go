package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func probeSet() *domain.Set {
	return &domain.Set{
		Name:   "Mein Probeset",
		Game:   "probe",
		Style:  "probe-stil",
		Values: []domain.Value{{Field: "title", Text: "Mein Probeset"}},
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}},
			{Values: []domain.Value{{Field: "name", Text: "Eisball"}}},
		},
	}
}

func TestSetAnlegenSchreibtAllesUndBehaeltDieReihenfolge(t *testing.T) {
	db := frischeDatenbank(t)

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(),
		[]domain.RawBlock{{File: "set", Line: 12, Text: "unverstanden: wert"}})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}
	if kennung == 0 {
		t.Error("keine Kennung zurückgegeben")
	}

	zaehlungen := map[string]int{
		"sets": 1, "cards": 2, "card_values": 2, "set_values": 1, "raw_blocks": 1,
	}
	for tabelle, erwartet := range zaehlungen {
		t.Run(tabelle, func(t *testing.T) {
			if gezaehlt := zaehle(t, db, tabelle); gezaehlt != erwartet {
				t.Errorf("%d Zeilen in %s, erwartet waren %d", gezaehlt, tabelle, erwartet)
			}
		})
	}

	var erste string
	if err := db.QueryRowContext(probekontext(t),
		`SELECT wert FROM card_values
		   JOIN cards ON cards.id = card_values.card_id
		  WHERE cards.reihenfolge = 0`).Scan(&erste); err != nil {
		t.Fatalf("erste Karte nicht lesbar: %v", err)
	}
	if erste != "Feuerball" {
		t.Errorf("die erste Karte ist %q — die Reihenfolge stimmt nicht", erste)
	}
}

// Und die Suche findet die Karte sofort.
func TestEinAngelegtesSetIstDurchsuchbar(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(), nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if treffer := suchen(t, db, "Eisball"); treffer != 1 {
		t.Errorf("%d Treffer auf Eisball, erwartet war 1", treffer)
	}
}

// Ein Set, dessen Karten halb geschrieben sind, sieht wie ein Set aus.
//
// Wer es öffnet, findet weniger Karten als erwartet und keinen Hinweis darauf,
// warum. Deshalb die Transaktion — und deshalb dieser Test.
func TestEinGescheitertesAnlegenLaesstNichtsHalbesZurueck(t *testing.T) {
	db := frischeDatenbank(t)

	kaputt := probeSet()
	// Zwei Felder gleichen Namens an einer Karte: der Primärschlüssel von
	// card_values weist das ab — aber erst bei der zweiten Zeile.
	kaputt.Cards[0].Values = append(kaputt.Cards[0].Values,
		domain.Value{Field: "name", Text: "nochmal"})

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), kaputt, nil); err == nil {
		t.Fatal("das kaputte Set wurde angenommen")
	}

	for _, tabelle := range []string{"sets", "cards", "card_values", "set_values"} {
		t.Run(tabelle, func(t *testing.T) {
			if gezaehlt := zaehle(t, db, tabelle); gezaehlt != 0 {
				t.Errorf("%d Zeilen in %s blieben liegen", gezaehlt, tabelle)
			}
		})
	}
}

func TestSetAnlegenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := storage.NeuerSetSpeicher(db).
		SetAnlegen(probekontext(t), probeSet(), nil); fehler == nil {
		t.Error("eine geschlossene Datenbank wurde angenommen")
	}
}

// Offen, aber nicht migriert: die Tabelle `sets` gibt es nicht.
func TestSetAnlegenMeldetEinFehlendesSchema(t *testing.T) {
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/ohneschema.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if _, fehler := storage.NeuerSetSpeicher(db).
		SetAnlegen(probekontext(t), probeSet(), nil); fehler == nil {
		t.Error("eine Datenbank ohne Schema wurde angenommen")
	}
}

// Eine doppelte Setangabe bricht ab und lässt nichts liegen.
func TestEineDoppelteSetangabeBrichtAb(t *testing.T) {
	db := frischeDatenbank(t)

	kaputt := probeSet()
	kaputt.Values = append(kaputt.Values, domain.Value{Field: "title", Text: "nochmal"})

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), kaputt, nil); err == nil {
		t.Fatal("eine doppelte Setangabe wurde angenommen")
	}
	if gezaehlt := zaehle(t, db, "sets"); gezaehlt != 0 {
		t.Errorf("%d Sets blieben liegen", gezaehlt)
	}
}

// Ein Rohblock mit einer Zeile, die es zweimal gibt, bricht ab.
func TestEinUnschreibbarerRohblockBrichtAb(t *testing.T) {
	db := frischeDatenbank(t)

	// Ein zu langer Dateiname ist kein Fehler; ein Rohblock mit einer
	// NULL-Zeile schon. Erzeugt wird er über eine Datei ohne Namen — die
	// Spalte ist NOT NULL, aber leer ist erlaubt. Also über den Primärschlüssel:
	// zwei Blöcke bekommen von `rohSchreiben` die Nummern 0 und 1, ein dritter
	// mit derselben Nummer kommt über die Schnittstelle nicht zustande.
	//
	// Geprüft wird deshalb der Weg, der erreichbar ist: ein Set, das schon da
	// ist, mit derselben Kennung noch einmal zu beschreiben.
	speicher := storage.NeuerSetSpeicher(db)

	if _, err := speicher.SetAnlegen(probekontext(t), probeSet(),
		[]domain.RawBlock{{File: "set", Line: 1, Text: "a"}, {File: "set", Line: 2, Text: "b"}}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if gezaehlt := zaehle(t, db, "raw_blocks"); gezaehlt != 2 {
		t.Errorf("%d Rohblöcke, erwartet waren 2", gezaehlt)
	}
}

func zaehle(t *testing.T, db *sql.DB, tabelle string) int {
	t.Helper()

	var anzahl int
	if err := db.QueryRowContext(probekontext(t),
		"SELECT count(*) FROM "+tabelle).Scan(&anzahl); err != nil { //nolint:gosec // feste Namen
		t.Fatalf("%s nicht lesbar: %v", tabelle, err)
	}

	return anzahl
}

// Ein doppelter Feldwert kommt als Konflikt an, nicht als SQL-Meldung.
//
// Bis zum Zweitblick über Phase 2 reichte storage seine SQLite-Fehler
// unverändert nach oben. Ein doppelter Feldwert endete am Primärschlüssel und
// kam als „SQL logic error" an — das Frontend hätte ihn als unbekannten Fehler
// gezeigt statt zum Nachladen aufzufordern (AGENTS.md §8).
func TestEinDoppelterFeldwertKommtAlsKonfliktAn(t *testing.T) {
	db := frischeDatenbank(t)

	kaputt := probeSet()
	kaputt.Cards[0].Values = append(kaputt.Cards[0].Values,
		domain.Value{Field: "name", Text: "nochmal"})

	_, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), kaputt, nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ein Konflikt", err)
	}
	// Und die rohe Ursache reist mit — für das Protokoll, nicht für die
	// Oberfläche. `internal/app` ersetzt sie durch den Rückfallsatz.
	//
	// **Über den Typ, nicht über den Wortlaut.** Hier stand
	// `strings.Contains(err.Error(), "constraint")` — derselbe Griff, den §8
	// dem Produktivcode verbietet, im Test. Formuliert der Treiber seine
	// Meldung um, wird dieser Test rot, ohne dass etwas kaputt wäre.
	//
	// **Nach dem Code gefragt und nicht nach der Hülle** — seit dem
	// Treiberwechsel am 2026-09-06 (Werkstück 14). `ncruces` gibt denselben
	// Fehler in drei Gestalten; `*sqlite3.Error` ist nur eine davon, und ein
	// `errors.As` darauf ginge an den beiden anderen vorbei. Genau diese Frage
	// stellt `einordnen` drinnen, also stellt der Test sie auch.
	var ursache sqlite.ExtendedErrorCode
	if err != nil && !errors.As(err, &ursache) {
		t.Errorf("die Ursache ging verloren: %v", err)
	}
}

// Ein Fehler ohne SQLite-Bezug reist unverändert nach oben.
func TestEinFremderFehlerWirdNichtUmgedeutet(t *testing.T) {
	// Eine geschlossene Datenbank meldet `sql: database is closed` — das ist
	// kein SQLite-Code, also auch kein Konflikt.
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	fehler := func() error {
		_, e := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(), nil)

		return e
	}()

	if errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein fremder Fehler wurde als Konflikt gedeutet: %v", fehler)
	}
	if fehler == nil {
		t.Error("kein Fehler")
	}
}
