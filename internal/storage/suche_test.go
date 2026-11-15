package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// suchbestand legt zwei Sets mit je zwei Karten an.
func suchbestand(t *testing.T) (*storage.SetLeser, *sql.DB) {
	t.Helper()

	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	erst := &domain.Set{
		Name: "Erstes Set", Game: "magic",
		Cards: []*domain.Card{
			{Values: []domain.Value{
				{Field: "name", Text: "Feuerball"},
				{Field: "typ", Text: "Zauber"},
				{Field: "text", Text: "Feuerball verursacht Schaden."},
			}},
			{Values: []domain.Value{
				{Field: "name", Text: "Eisblitz"},
				{Field: "typ", Text: "Zauber"},
			}},
		},
	}

	zweit := &domain.Set{
		Name: "Zweites Set", Game: "yugioh",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Feuerdrache"}}},
		},
	}

	for _, set := range []*domain.Set{erst, zweit} {
		if _, err := speicher.SetAnlegen(probekontext(t), set, nil); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	return storage.NeuerSetLeser(db), db
}

func TestDieSucheFindetUeberAlleFelder(t *testing.T) {
	leser, _ := suchbestand(t)

	faelle := []struct {
		name    string
		frage   domain.SearchQuery
		treffer int
	}{
		{"ein Wort im Namen", domain.SearchQuery{Text: "Eisblitz"}, 1},
		// **Feuerball steht im Namen und im Text** — zwei Stellen, zwei Treffer.
		{"ein Wort in zwei Feldern", domain.SearchQuery{Text: "Feuerball"}, 2},
		// Ein Wortanfang genügt: wer „feuer" tippt, meint auch „Feuerdrache".
		{"ein Wortanfang", domain.SearchQuery{Text: "feuer"}, 3},
		{"in einem Feld", domain.SearchQuery{Text: "feuer", Field: "name"}, 2},
		{"in einem Set", domain.SearchQuery{Text: "feuer", SetID: 1}, 2},
		{
			"in einem Feld und einem Set",
			domain.SearchQuery{Text: "feuer", Field: "name", SetID: 1},
			1,
		},
		{
			"mehrere Wörter gelten zusammen",
			domain.SearchQuery{Text: "Feuerball Schaden"},
			1,
		},
		{"nichts gefunden", domain.SearchQuery{Text: "gibtsnicht"}, 0},
		{"leere Eingabe", domain.SearchQuery{}, 0},
		{"nur Leerraum", domain.SearchQuery{Text: "   "}, 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ergebnis, err := leser.Suche(probekontext(t), fall.frage)
			if err != nil {
				t.Fatalf("suchen: %v", err)
			}

			if len(ergebnis.Hits) != fall.treffer {
				t.Errorf("%d Treffer, erwartet waren %d: %+v",
					len(ergebnis.Hits), fall.treffer, ergebnis.Hits)
			}
		})
	}
}

// **Ein Treffer sagt, welches Feld getroffen hat — und mit welchem Wert.**
//
// Das konnte die Tabelle aus 0003 nicht: `content=”` heißt inhaltslos, und ein
// `SELECT feld, wert` gab NULL. Gefunden hat es niemand, weil bis Phase 5
// niemand gesucht hat.
func TestEinTrefferNenntFeldUndWert(t *testing.T) {
	leser, _ := suchbestand(t)

	ergebnis, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: "Eisblitz"})
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if len(ergebnis.Hits) != 1 {
		t.Fatalf("%d Treffer", len(ergebnis.Hits))
	}

	fund := ergebnis.Hits[0]
	if fund.Field != "name" || fund.Value != "Eisblitz" {
		t.Errorf("%+v", fund)
	}

	if fund.SetID != 1 || fund.Number != 1 {
		t.Errorf("%+v", fund)
	}
}

// Ein geänderter Feldwert ist sofort anders zu finden.
func TestEineAenderungIstSofortZuFinden(t *testing.T) {
	leser, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	if err := schreiber.FeldSetzen(probekontext(t), 2, "name", "Kältewelle"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	alt, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: "Eisblitz"})
	if err != nil || len(alt.Hits) != 0 {
		t.Errorf("der alte Wert ist noch zu finden: %+v / %v", alt, err)
	}

	neu, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: "Kältewelle"})
	if err != nil || len(neu.Hits) != 1 {
		t.Errorf("der neue Wert ist nicht zu finden: %+v / %v", neu, err)
	}
}

// **Ein Anführungszeichen ist ein Zeichen, kein Befehl.**
//
// FTS5 hat eine eigene Sprache; wer die Eingabe durchreicht, bekommt bei einem
// Kartennamen mit `"` einen Syntaxfehler statt eines Treffers.
func TestSonderzeichenSindZeichen(t *testing.T) {
	db := frischeDatenbank(t)

	set := &domain.Set{Name: "s", Game: "g", Cards: []*domain.Card{
		{Values: []domain.Value{{Field: "name", Text: `Der "Große" Drache`}}},
	}}

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	for _, eingabe := range []string{`"Große"`, `"`, `Große* AND`, `NEAR(a b)`, `-Drache`} {
		t.Run(eingabe, func(t *testing.T) {
			if _, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: eingabe}); err != nil {
				t.Errorf("die Eingabe %q ergab einen Fehler: %v", eingabe, err)
			}
		})
	}

	ergebnis, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: `Große`})
	if err != nil || len(ergebnis.Hits) != 1 {
		t.Errorf("%+v / %v", ergebnis, err)
	}
}

// Die Schranke ist zu sehen, nicht still.
func TestDieTrefferschrankeIstZuSehen(t *testing.T) {
	db := frischeDatenbank(t)

	karten := make([]*domain.Card, 0, domain.SearchLimit+5)
	for range domain.SearchLimit + 5 {
		karten = append(karten, &domain.Card{
			Values: []domain.Value{{Field: "name", Text: "Gleichname"}},
		})
	}

	if _, err := storage.NeuerSetSpeicher(db).
		SetAnlegen(probekontext(t), &domain.Set{Name: "s", Game: "g", Cards: karten}, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ergebnis, err := storage.NeuerSetLeser(db).
		Suche(probekontext(t), domain.SearchQuery{Text: "Gleichname"})
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if !ergebnis.Truncated {
		t.Errorf("%d Treffer — die Kürzung wurde nicht gemeldet", len(ergebnis.Hits))
	}

	if len(ergebnis.Hits) != domain.SearchLimit {
		t.Errorf("%d nach der Kürzung", len(ergebnis.Hits))
	}

	// Und unterhalb der Schranke wird nichts gemeldet.
	wenige, err := storage.NeuerSetLeser(db).
		Suche(probekontext(t), domain.SearchQuery{Text: "gibtsnicht"})
	if err != nil || wenige.Truncated {
		t.Errorf("%+v / %v", wenige, err)
	}

	if len(wenige.Hits) != 0 {
		t.Errorf("%d", len(wenige.Hits))
	}
}

func TestKartenwerteGebenMehrereKartenAufEinmal(t *testing.T) {
	leser, _ := suchbestand(t)

	werte, err := leser.Kartenwerte(probekontext(t), []int64{1, 2})
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(werte[1]) != 3 || len(werte[2]) != 2 {
		t.Errorf("%+v", werte)
	}

	// Die Reihenfolge ist die aus der Datei — sie zählt für den Rundlauf.
	if werte[1][0].Field != "name" || werte[1][2].Field != "text" {
		t.Errorf("%+v", werte[1])
	}

	leer, err := leser.Kartenwerte(probekontext(t), nil)
	if err != nil || len(leer) != 0 {
		t.Errorf("%v / %v", leer, err)
	}
}

func TestDieSucheMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	leser := storage.NeuerSetLeser(db)

	if _, err := leser.Suche(probekontext(t), domain.SearchQuery{Text: "a"}); err == nil {
		t.Error("die Suche kam ohne Fehler zurück")
	}

	if _, err := leser.Kartenwerte(probekontext(t), []int64{1}); err == nil {
		t.Error("Kartenwerte kam ohne Fehler zurück")
	}

	if _, err := leser.Felder(probekontext(t), 0); err == nil {
		t.Error("Felder kam ohne Fehler zurück")
	}
}

// **Die Feldnamen kommen aus dem Bestand, nicht aus einer Liste bei uns.**
//
// Ein fremdes Spiel führt Felder, die wir nicht kennen; ein Filter über eine
// selbst geschriebene Aufzählung ließe genau die nicht filtern.
func TestFelderKommenAusDemBestand(t *testing.T) {
	leser, _ := suchbestand(t)

	alle, err := leser.Felder(probekontext(t), 0)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// Sortiert und ohne Wiederholung — „name" steht an drei Karten.
	if len(alle) != 3 || alle[0] != "name" || alle[1] != "text" || alle[2] != "typ" {
		t.Errorf("%q", alle)
	}

	// Mit Set gefiltert: das zweite Set führt nur „name".
	eines, err := leser.Felder(probekontext(t), 2)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eines) != 1 || eines[0] != "name" {
		t.Errorf("%q", eines)
	}
}
