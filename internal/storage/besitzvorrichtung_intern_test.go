package storage

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorrichtung für die Besitzläufe** —
// #1815.
//
// Profil 2 besitzt von jeder Art ein Objekt samt seinen Abkömmlingen; Profil
// 1 besitzt nichts davon. Die Läufe rufen aus Profil 1 mit den Kennungen von
// Profil 2 und verlangen „nicht gefunden" — und dass danach in Profil 2
// alles steht wie vorher.

// fremdeKennungen sind die Kennungen der Objekte, die Profil 2 gehören.
type fremdeKennungen struct {
	set, karte, wort, deck, platz, exemplar, zuordnung int64
	turnier, teilnehmer, zweiter, paarung              int64
	partie                                             string
}

// besitzbestand gibt einen migrierten Bestand, in dem Profil 2 von jeder Art
// ein Objekt besitzt.
func besitzbestand(t *testing.T) (*sql.DB, fremdeKennungen) {
	t.Helper()

	db := migrierteDatenbank(t)

	for _, anweisung := range []string{
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Fremd')`,
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (20, 'Fremdes Set', 'magic', '2026-09-23', 2)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (200, 20, 0), (201, 20, 1)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES
		   (200, 'name', 'Fremde Karte', 0), (201, 'name', 'Zweite fremde Karte', 0)`,
		`INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (20, 'title', 'Fremd', 0)`,
		`INSERT INTO set_quellen (set_id, datei, inhalt) VALUES (20, '/fremd.mse-set', x'00')`,
		`INSERT INTO styling_values (set_id, feld, wert, reihenfolge) VALUES (20, 'rahmen', 'rot', 0)`,
		`INSERT INTO set_keywords (id, set_id, reihenfolge, wort) VALUES (21, 20, 0, 'fremdwort')`,
		`INSERT INTO karten_marken (card_id, marke) VALUES (200, 'fremd')`,
		`INSERT INTO karten_verweis (card_id, printing, name, gesetzt) VALUES (200, 'druck-fremd', 'Fremde Karte', 1)`,
		`INSERT INTO history (set_id, karte_id, zeitpunkt, handlung, rueckweg, hinweg) VALUES
		   (20, 200, '2026-09-23T00:00:00.000Z', 'feld',
		    '{"feld":"name","wert":"Alt","gabEs":true}', '{"feld":"name","wert":"Fremde Karte","gabEs":true}')`,
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (30, 'Fremdes Deck', 'magic', 2)`,
		`INSERT INTO deckplaetze (id, deck_id, karte_id, menge, bereich) VALUES (31, 30, 200, 1, 'hauptdeck')`,
		`INSERT INTO sammlung (id, printing, menge, zustand, sprache, finish, erwerbspreis, profil_id)
		   VALUES (40, 'druck-fremd', 2, 'nearMint', 'en', 'nonfoil', 0, 2)`,
		`INSERT INTO deckzuordnungen (id, deck_id, platz_id, exemplar_id, menge) VALUES (41, 30, 31, 40, 1)`,
		`INSERT INTO deck_herkunft (deck_id, art, quelle, uebernommen) VALUES (30, 'einfuhr', 'x', '2026-09-23')`,
		`INSERT INTO sammlung_herkunft (exemplar_id, art, quelle, uebernommen) VALUES (40, 'einfuhr', 'x', '2026-09-23')`,
		`INSERT INTO turniere (id, name, spiel, profil_id) VALUES (50, 'Fremdes Turnier', 'magic', 2)`,
		`INSERT INTO turnierteilnehmer (id, turnier_id, name, stelle) VALUES (51, 50, 'Anna', 1), (52, 50, 'Bert', 2)`,
		`INSERT INTO turnierrunden (id, turnier_id, nummer) VALUES (53, 50, 1)`,
		`INSERT INTO turnierpaarungen (id, runde_id, a, b, stelle) VALUES (54, 53, 51, 52, 1)`,
		`INSERT INTO partien (id, profil_id, spiel, stand, zustand) VALUES ('partie-fremd', 2, 'magic', 'laeuft', '{}')`,
		`INSERT INTO partiefreigaben (partie_id, von, an) VALUES ('partie-fremd', 'sitz1', 'sitz2')`,
		`INSERT INTO partieereignisse (partie_id, folge, art, oeffentlich, kette)
		   VALUES ('partie-fremd', 1, 'zug', '{}', 'kette-fremd')`,
		`INSERT INTO partiebefehle (partie_id, kennung, absender, art, von_folge, bis_folge)
		   VALUES ('partie-fremd', 'befehl-fremd', 'sitz1', 'zug', 1, 1)`,
		`INSERT INTO matchbindungen (paarung_id, turnier_id, fassung, partie_id, sitze)
		   VALUES (54, 50, 1, 'partie-fremd', '[]')`,
		`INSERT INTO matchmeldungen (paarung_id, von, fassung, partie_id, spiele_a, spiele_b, unentschieden, kette)
		   VALUES (54, 'sitz1', 1, 'partie-fremd', 2, 0, 0, 'kette-fremd')`,
	} {
		if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
			t.Fatalf("%s: %v", strings.Fields(anweisung)[2], fehler)
		}
	}

	return db, fremdeKennungen{
		set: 20, karte: 200, wort: 21, deck: 30, platz: 31, exemplar: 40, zuordnung: 41,
		turnier: 50, teilnehmer: 51, zweiter: 52, paarung: 54, partie: "partie-fremd",
	}
}

// besitztafeln sind die Tafeln, deren Inhalt ein Lauf aus einem fremden Profil
// nicht ändern darf.
var besitztafeln = []string{
	"sets", "set_quellen", "cards", "card_values", "set_values", "styling_values", "set_keywords",
	"karten_marken", "karten_verweis", "history", "decks", "deckplaetze",
	"deckzuordnungen", "deck_herkunft", "sammlung", "sammlung_herkunft",
	"turniere", "turnierteilnehmer", "turnierrunden", "turnierpaarungen",
	"matchbindungen", "matchmeldungen", "partien", "partiefreigaben",
	"partieereignisse", "partiebefehle",
}

// besitzschnappschuss gibt den Inhalt der Tafeln als Text.
func besitzschnappschuss(t *testing.T, db *sql.DB) string {
	t.Helper()

	var teile []string

	for _, tafel := range besitztafeln {
		for _, zeile := range zeilenAus(t, t.Context(), db, `SELECT * FROM `+tafel) {
			teile = append(teile, tafel+": "+strings.Join(zeile, "|"))
		}
	}

	// **Sortiert und nicht nach `rowid`**: manche Tafel hat keine.
	slices.Sort(teile)

	return strings.Join(teile, "\n")
}

// ausFremdemProfil prüft, dass jeder Ruf aus Profil 1 „nicht gefunden" sagt
// und nichts ändert.
func ausFremdemProfil(t *testing.T, db *sql.DB, rufe map[string]func() error) {
	t.Helper()

	vorher := besitzschnappschuss(t, db)

	for name, ruf := range rufe {
		if fehler := ruf(); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("%s aus einem fremden Profil: %v, erwartet nicht gefunden", name, fehler)
		}
	}

	if nachher := besitzschnappschuss(t, db); nachher != vorher {
		t.Errorf("Profil 2 hat sich geändert:\n%s", unterschied(vorher, nachher))
	}
}

// unterschied nennt die Zeilen, die nur in einem der beiden Stände stehen.
func unterschied(vorher, nachher string) string {
	alt := map[string]bool{}
	for _, zeile := range strings.Split(vorher, "\n") {
		alt[zeile] = true
	}

	var neu []string

	for _, zeile := range strings.Split(nachher, "\n") {
		if !alt[zeile] {
			neu = append(neu, "+ "+zeile)
		}

		delete(alt, zeile)
	}

	for zeile := range alt {
		neu = append(neu, "- "+zeile)
	}

	return strings.Join(neu, "\n")
}
