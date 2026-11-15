package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die leseintensiven Listen lesen nicht den ganzen Bestand** — ENG-05.AK-02
// (#1730).
//
// # Gemessen, nicht erklärt
//
// **SQLite zählt selbst mit**: `sqlite3_stmt_status` mit
// `SQLITE_STMTSTATUS_FULLSCAN_STEP` sagt, wie viele Schritte eine Anweisung
// in einem vollständigen Durchlauf einer Tafel gegangen ist. Gezählt wird an
// den **echten** Abfragen — die Listenfunktionen laufen gegen einen Bestand,
// und ein Rückruf am Treiber liest nach jeder Anweisung ihre Zähler. Ein
// `EXPLAIN QUERY PLAN` daneben wäre eine zweite Abschrift derselben Abfrage,
// und die liefe auseinander.
//
// # Die Probe
//
// **Zwei Profile, das fremde hundertmal größer.** Liest eine Liste des kleinen
// Profils über einen Index, bleibt ihr Zähler unter dem, was das große
// Profil in derselben Tafel hält; liest sie die ganze Tafel, liegt er darüber.
// Die Schwelle ist damit keine Schätzung, sondern die Größe des Fremden.
//
// # Was er nicht misst
//
//   - **Die Zeit.** Sie hängt an der Maschine; die Zahl der Schritte nicht.
//   - **Schreibwege.** Beim Löschen sucht SQLite die Kinder eines
//     Fremdschlüssels — das hält `TestJederFremdschluesselHatSeinenIndex`.
//   - **Was nur ein Profil sieht**, das die ganze Tafel ist. Das ist die
//     Größe seiner Liste und kein Befund.
func TestDieListenLesenNichtDenGanzenBestand(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := MitProfil(t.Context(), 1)

	fremd := befuelleZweiProfile(t, db)
	zaehler := zaehlerAnbringen(t, db)

	// **Die Gegenprobe zuerst** (§10): eine Abfrage, die die Tafel ganz lesen
	// muss, zählt ihre Zeilen. Bliebe der Zähler hier bei null, sagten die
	// Nullen darunter nichts.
	var keine int
	if fehler := db.QueryRowContext(ctx,
		`SELECT count(*) FROM sammlung WHERE notiz = ?`, "nirgends").Scan(&keine); fehler != nil {
		t.Fatalf("die Gegenprobe: %v", fehler)
	}

	if probe := zaehler.gelesen(); len(probe) != 1 || probe[0].vollscan < fremd {
		t.Fatalf("die Gegenprobe zählte %+v — erwartet mindestens %d Schritte", probe, fremd)
	}

	for _, liste := range []struct {
		name string
		lies func() error
	}{
		{"Sets", func() error { _, f := NeuerSetLeser(db).Setkoepfe(ctx); return f }},
		{"Kartenzeilen", func() error { _, f := NeuerSetLeser(db).Kartenzeilen(ctx, 1); return f }},
		{"Suche", func() error {
			_, f := NeuerSetLeser(db).Suche(ctx, domain.SearchQuery{Text: "Karte"})
			return f
		}},
		{"Sammlung", func() error {
			_, f := NeuerSammlungsspeicher(db).Fenster(ctx, func(ordnung []*domain.Exemplar) []int64 {
				kennungen := make([]int64, 0, 50)
				for _, eines := range ordnung[:min(50, len(ordnung))] {
					kennungen = append(kennungen, eines.ID)
				}

				return kennungen
			})

			return f
		}},
		{"Decks", func() error { _, f := NeueDeckablage(db).Decks(ctx, "magic"); return f }},
		{"Turniere", func() error { _, f := NeueTurnierablage(db).Turniere(ctx, "magic"); return f }},
		{"Formate", func() error { _, f := NeueFormatablage(db).Formate(ctx, "magic"); return f }},
		{"Verlauf", func() error { _, f := NeuerVerlaufsleser(db).Verlauf(ctx, 1); return f }},
		{"Handel", func() error { _, f := NeueHandelsablage(db).Uebersicht(ctx); return f }},
		{"Preisverlauf", func() error {
			_, f := NeuerPreisspeicher(db).Verlauf(ctx, "druck-1", domain.Preisstand{
				Quelle: "scryfall", Markt: "cardmarket", Art: "trend",
				Waehrung: "EUR", Finish: domain.FinishNonfoil, Zustand: "", Sprache: "",
			})

			return f
		}},
	} {
		zaehler.leeren()

		if fehler := liste.lies(); fehler != nil {
			t.Fatalf("%s: %v", liste.name, fehler)
		}

		for _, anweisung := range zaehler.gelesen() {
			plan := planVon(t, db, anweisung.sql)
			t.Logf("%-12s %6d Schritte im Vollscan  %s", liste.name, anweisung.vollscan, knapp(anweisung.sql))
			t.Logf("%-12s        Plan: %s", liste.name, plan)

			// **Ein Bereich so groß wie die Sammlung ist kein Vollscan** und
			// entgeht dem Zähler: die Handelsliste suchte je Vorgang alle
			// Scryfall-Kennungen über `system=?` allein (TRADE-01, #1715).
			if liste.name == "Handel" && !strings.Contains(plan, "system=? AND wert=?") {
				t.Errorf("die Handelsliste sucht die Kennung nicht über ihren Wert: %s", plan)
			}

			if anweisung.vollscan >= fremd {
				t.Errorf("%s liest %d Zeilen am Stück — mehr, als das fremde Profil hält (%d): %s",
					liste.name, anweisung.vollscan, fremd, knapp(anweisung.sql))
			}
		}
	}
}

// messpunkt ist eine gelaufene Anweisung mit ihrem Zähler.
type messpunkt struct {
	sql      string
	vollscan int
}

// zaehlwerk sammelt die Zähler der Anweisungen, die gelaufen sind.
type zaehlwerk struct {
	schutz sync.Mutex
	punkte []messpunkt
}

func (z *zaehlwerk) leeren() {
	z.schutz.Lock()
	defer z.schutz.Unlock()

	z.punkte = nil
}

func (z *zaehlwerk) gelesen() []messpunkt {
	z.schutz.Lock()
	defer z.schutz.Unlock()

	return append([]messpunkt(nil), z.punkte...)
}

// zaehlerAnbringen hängt einen Rückruf an die eine Verbindung des Bestands
// (`SetMaxOpenConns(1)`), der nach jeder Anweisung ihren Zähler liest.
func zaehlerAnbringen(t *testing.T, db *sql.DB) *zaehlwerk {
	t.Helper()

	werk := &zaehlwerk{}

	verbindung, fehler := db.Conn(t.Context())
	if fehler != nil {
		t.Fatalf("die Verbindung: %v", fehler)
	}

	defer func() { _ = verbindung.Close() }() //nolint:errcheck // gibt sie nur dem Vorrat zurück

	fehler = verbindung.Raw(func(roh any) error {
		treiber, ist := roh.(interface{ Raw() *sqlite.Conn })
		if !ist {
			return fmt.Errorf("%T gibt keine SQLite-Verbindung her", roh)
		}

		return treiber.Raw().Trace(sqlite.TRACE_PROFILE, func(_ sqlite.TraceEvent, arg1, _ any) error {
			anweisung, ist := arg1.(*sqlite.Stmt)
			if !ist {
				return nil
			}

			werk.schutz.Lock()
			defer werk.schutz.Unlock()

			werk.punkte = append(werk.punkte, messpunkt{
				sql:      anweisung.SQL(),
				vollscan: anweisung.Status(sqlite.STMTSTATUS_FULLSCAN_STEP, true),
			})

			return nil
		})
	})
	if fehler != nil {
		t.Fatalf("den Zähler anbringen: %v", fehler)
	}

	return werk
}

// befuelleZweiProfile legt ein kleines Profil 1 und ein hundertmal größeres
// Profil 2 an und gibt die Zeilenzahl, die das große in jeder Liste hält.
func befuelleZweiProfile(t *testing.T, db *sql.DB) int {
	t.Helper()

	const klein, gross = 20, 2000

	ctx := t.Context()

	for _, anweisung := range []string{
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Groß')`,
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		  SELECT i, 'Set ' || i, 'magic', '2026-09-23', CASE WHEN i <= %d THEN 1 ELSE 2 END FROM n`,
			klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO cards (id, set_id, reihenfolge) SELECT i, 1 + (i %% %d), i FROM n`,
			(klein+gross)*5, klein+gross),
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge)
		  SELECT id, 'name', 'Karte ' || id, 0 FROM cards`,
		`INSERT INTO karten_marken (card_id, marke) SELECT id, 'Karte' FROM cards`,
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, rueckweg)
		  SELECT set_id, id, '2026-09-23T00:00:00.000Z', 'feld',
		         '{"feld":"name","wert":"alt","gabEs":true}' FROM cards`,
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO sammlung (printing, menge, zustand, sprache, finish, name, erwerbspreis, profil_id)
		  SELECT 'druck-' || i, 1, 'nearMint', 'en', 'nonfoil', 'Karte ' || i, 0,
		         CASE WHEN i <= %d THEN 1 ELSE 2 END FROM n`, klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO decks (name, spiel, profil_id)
		  SELECT 'Deck ' || i, 'magic', CASE WHEN i <= %d THEN 1 ELSE 2 END FROM n`, klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO turniere (name, spiel, profil_id)
		  SELECT 'Turnier ' || i, 'magic', CASE WHEN i <= %d THEN 1 ELSE 2 END FROM n`, klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO formate (name, spiel, profil_id)
		  SELECT 'Format ' || i, 'magic', CASE WHEN i <= %d THEN 1 ELSE 2 END FROM n`, klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO handelsrevisionen (profil_id, vorgang, revision, abdruck, fassung, zustand, angebot)
		  SELECT CASE WHEN i <= %d THEN 1 ELSE 2 END, 'h-' || i, 1, 'ab-' || i, 1, 'entwurf',
		         '{"Vorgang":"h-' || i || '","Revision":1,"Positionen":[{"Printing":"druck-' || i || '"}]}'
		    FROM n`, klein+gross, klein),
		fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < %d)
		  INSERT INTO preisstaende (printing, quelle, markt, art, waehrung, finish, zustand, sprache, wert, erhoben)
		  SELECT 'druck-' || (i %% %d), 'scryfall', 'cardmarket', 'trend', 'EUR', 'nonfoil', '', '',
		         100 + i, 1700000000 + i FROM n`, (klein+gross)*5, klein+gross),
	} {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
			t.Fatalf("befüllen: %v\n%s", fehler, anweisung)
		}
	}

	return gross
}

// knapp kürzt eine Anweisung auf eine Zeile.
func knapp(anweisung string) string {
	eine := strings.Join(strings.Fields(anweisung), " ")
	if len(eine) > 140 {
		return eine[:140] + " …"
	}

	return eine
}

// planVon gibt den Plan einer Anweisung auf einer Zeile — für das Protokoll
// des Laufs und damit für `docs/ABFRAGEPLAENE.md`.
func planVon(t *testing.T, db *sql.DB, anweisung string) string {
	t.Helper()

	// **Leere Werte für die Platzhalter**: der Plan hängt nicht an ihnen,
	// `database/sql` verlangt aber so viele, wie die Anweisung nennt.
	platzhalter := make([]any, strings.Count(anweisung, "?"))
	zeilen := zeilenAus(t, t.Context(), db, "EXPLAIN QUERY PLAN "+anweisung, platzhalter...)

	schritte := make([]string, 0, len(zeilen))
	for _, zeile := range zeilen {
		schritte = append(schritte, zeile[len(zeile)-1])
	}

	return strings.Join(schritte, " · ")
}
