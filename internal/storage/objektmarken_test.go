package storage_test

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.

import (
	"database/sql"
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// objektstand legt je Art ein Objekt im ersten Profil an — und ein Deck im
// zweiten.
func objektstand(t *testing.T) (db *sql.DB, kennungen map[domain.Markenart]string, fremdesDeck string, zweites int64) {
	t.Helper()

	db, _, setID, karteID := markenstand(t)
	zweites = zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier")

	for _, anweisung := range []string{
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (11, 'Rot', 'magic', 1)`,
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (12, 'Fremd', 'magic', ?)`,
		`INSERT INTO turniere (id, name, spiel, profil_id) VALUES (21, 'Freitag', 'magic', 1)`,
		`INSERT INTO sammlung (id, printing, menge, zustand, sprache, profil_id) VALUES (31, 'p', 1, 'nearMint', 'en', 1)`,
		`INSERT INTO partien (id, profil_id, spiel, stand, zustand) VALUES ('partie-1', 1, 'magic', 'laufend', '{}')`,
		`INSERT INTO handelsfolge (profil_id, vorgang) VALUES (1, 'tausch-mit-ada')`,
		`INSERT INTO packages (art, name, pfad) VALUES ('game', 'magic', '/x')`,
	} {
		argumente := []any{}
		if strings.Contains(anweisung, "?") {
			argumente = append(argumente, zweites)
		}

		if _, fehler := db.ExecContext(t.Context(), anweisung, argumente...); fehler != nil {
			t.Fatalf("%s: %v", anweisung, fehler)
		}
	}

	return db, map[domain.Markenart]string{
		domain.MarkenartKarte: strconv.FormatInt(karteID, 10), domain.MarkenartSet: strconv.FormatInt(setID, 10),
		domain.MarkenartExemplar: "31", domain.MarkenartDeck: "11", domain.MarkenartPaket: "game/magic",
		domain.MarkenartTurnier: "21", domain.MarkenartPartie: "partie-1", domain.MarkenartHandel: "tausch-mit-ada",
	}, "12", zweites
}

func TestJedeArtTraegtEineMarke(t *testing.T) {
	db, kennungen, _, _ := objektstand(t)
	ablage := storage.NeueObjektmarkenablage(db)

	for art, kennung := range kennungen {
		if vorhanden, fehler := ablage.Vorhanden(probekontext(t), art, kennung); fehler != nil || !vorhanden {
			t.Errorf("%s %s: vorhanden %v, %v", art, kennung, vorhanden, fehler)
		}

		// **Zweimal dieselbe Marke ist nichts** — wie an der Karte (0053).
		for range 2 {
			if fehler := ablage.Setzen(probekontext(t), art, kennung, "tausch"); fehler != nil {
				t.Fatalf("%s: %v", art, fehler)
			}
		}

		if marken, fehler := ablage.An(probekontext(t), art, kennung); fehler != nil || !slices.Equal(marken, []string{"tausch"}) {
			t.Errorf("%s: %v %v", art, marken, fehler)
		}

		je, fehler := ablage.JeObjekt(probekontext(t), art)
		if fehler != nil || !slices.Equal(je[kennung], []string{"tausch"}) {
			t.Errorf("%s je Objekt: %v %v", art, je, fehler)
		}
	}

	katalog, fehler := ablage.Katalog(probekontext(t))
	if fehler != nil || len(katalog) != 1 || katalog[0].Marke != "tausch" || katalog[0].Anzahl != 8 {
		t.Errorf("Katalog: %+v %v", katalog, fehler)
	}

	for art, kennung := range kennungen {
		if fehler := ablage.Entfernen(probekontext(t), art, kennung, "tausch"); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if katalog, _ := ablage.Katalog(probekontext(t)); len(katalog) != 0 { //nolint:errcheck // der Vergleich sagt es
		t.Errorf("nach dem Entfernen: %+v", katalog)
	}
}

// **Keine Lecks zwischen Profilen** (UX-12.AK-04).
func TestEinFremdesObjektGibtEsNicht(t *testing.T) {
	db, kennungen, fremdesDeck, zweites := objektstand(t)
	ablage := storage.NeueObjektmarkenablage(db)

	for _, fall := range []struct {
		art     domain.Markenart
		kennung string
	}{
		{domain.MarkenartDeck, fremdesDeck},
		{domain.MarkenartDeck, "999"},
		{domain.MarkenartPartie, "gibt-es-nicht"},
		{domain.MarkenartPaket, "game/fremd"},
	} {
		if vorhanden, fehler := ablage.Vorhanden(probekontext(t), fall.art, fall.kennung); fehler != nil || vorhanden {
			t.Errorf("%s %s: %v %v", fall.art, fall.kennung, vorhanden, fehler)
		}
	}

	// **Aus dem zweiten Profil ist das eigene Deck fremd** — und seine
	// Marken sind dort nicht zu sehen.
	if fehler := ablage.Setzen(probekontext(t), domain.MarkenartDeck, kennungen[domain.MarkenartDeck], "rot"); fehler != nil {
		t.Fatal(fehler)
	}

	dort := storage.MitProfil(t.Context(), zweites)
	if vorhanden, _ := ablage.Vorhanden(dort, domain.MarkenartDeck, kennungen[domain.MarkenartDeck]); vorhanden { //nolint:errcheck // der Vergleich sagt es
		t.Error("das Deck des ersten Profils ist im zweiten vorhanden")
	}

	if katalog, _ := ablage.Katalog(dort); len(katalog) != 0 { //nolint:errcheck // dito
		t.Errorf("Marken des ersten Profils im zweiten: %+v", katalog)
	}

	// **Ein Paket gehört dem Rechner, seine Marke dem Profil** (0135, Punkt 3).
	if fehler := ablage.Setzen(probekontext(t), domain.MarkenartPaket, "game/magic", "lieblingsspiel"); fehler != nil {
		t.Fatal(fehler)
	}

	if marken, _ := ablage.An(dort, domain.MarkenartPaket, "game/magic"); len(marken) != 0 { //nolint:errcheck // dito
		t.Errorf("die Marke am Paket reist ins zweite Profil: %v", marken)
	}
}

// **Wenn ein Objekt geht, gehen seine Marken** — je Art ein Auslöser.
func TestMitDemObjektGehenSeineMarken(t *testing.T) {
	db, kennungen, _, _ := objektstand(t)
	ablage := storage.NeueObjektmarkenablage(db)

	for art, kennung := range kennungen {
		if fehler := ablage.Setzen(probekontext(t), art, kennung, "weg"); fehler != nil {
			t.Fatal(fehler)
		}
	}

	for _, anweisung := range []string{
		`DELETE FROM decks WHERE id = 11`, `DELETE FROM turniere WHERE id = 21`,
		`DELETE FROM sammlung WHERE id = 31`, `DELETE FROM partien WHERE id = 'partie-1'`,
		`DELETE FROM handelsfolge WHERE vorgang = 'tausch-mit-ada'`, `DELETE FROM packages WHERE name = 'magic'`,
		`DELETE FROM sets`,
	} {
		if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
			t.Fatalf("%s: %v", anweisung, fehler)
		}
	}

	var uebrig int
	if fehler := db.QueryRowContext(t.Context(),
		`SELECT (SELECT count(*) FROM objektmarken) + (SELECT count(*) FROM karten_marken)`).Scan(&uebrig); fehler != nil || uebrig != 0 {
		t.Errorf("übrig: %d %v", uebrig, fehler)
	}
}

func TestUmbenennenUndZusammenfuehren(t *testing.T) {
	db, kennungen, _, zweites := objektstand(t)
	ablage := storage.NeueObjektmarkenablage(db)

	karte, deck := kennungen[domain.MarkenartKarte], kennungen[domain.MarkenartDeck]
	for _, gesetzt := range []struct {
		art     domain.Markenart
		kennung string
		marke   string
	}{
		{domain.MarkenartKarte, karte, "tausch"}, {domain.MarkenartDeck, deck, "tausch"}, {domain.MarkenartDeck, deck, "handel"},
	} {
		if fehler := ablage.Setzen(probekontext(t), gesetzt.art, gesetzt.kennung, gesetzt.marke); fehler != nil {
			t.Fatal(fehler)
		}
	}

	// **Das zweite Profil hat eine eigene „tausch"** — sie bleibt, was sie ist.
	if _, fehler := db.ExecContext(t.Context(),
		`INSERT INTO objektmarken (profil_id, art, kennung, marke) VALUES (?, 'paket', 'game/magic', 'tausch')`, zweites); fehler != nil {
		t.Fatal(fehler)
	}

	// **Auf einen vorhandenen Namen nur ausdrücklich** (UX-12.AK-02).
	if _, fehler := ablage.Umbenennen(probekontext(t), "tausch", "handel", false); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Kollision: %v", fehler)
	}

	betroffen, fehler := ablage.Umbenennen(probekontext(t), "tausch", "handel", true)
	if fehler != nil || betroffen != 2 {
		t.Fatalf("zusammenführen: %d %v", betroffen, fehler)
	}

	if marken, _ := ablage.An(probekontext(t), domain.MarkenartDeck, deck); !slices.Equal(marken, []string{"handel"}) { //nolint:errcheck // der Vergleich sagt es
		t.Errorf("Deck: %v", marken)
	}

	if marken, _ := ablage.An(probekontext(t), domain.MarkenartKarte, karte); !slices.Equal(marken, []string{"handel"}) { //nolint:errcheck // dito
		t.Errorf("Karte: %v", marken)
	}

	// **Ohne Kollision einfach umbenannt**, und dann überall entfernt.
	if betroffen, fehler := ablage.Umbenennen(probekontext(t), "handel", "abgabe", false); fehler != nil || betroffen != 2 {
		t.Errorf("umbenennen: %d %v", betroffen, fehler)
	}

	if betroffen, fehler := ablage.UeberallEntfernen(probekontext(t), "abgabe"); fehler != nil || betroffen != 2 {
		t.Errorf("überall entfernen: %d %v", betroffen, fehler)
	}

	if katalog, _ := ablage.Katalog(storage.MitProfil(t.Context(), zweites)); len(katalog) != 1 || katalog[0].Marke != "tausch" { //nolint:errcheck // der Vergleich sagt es
		t.Errorf("das zweite Profil: %+v", katalog)
	}
}

// **Die Massenvergabe ist eins oder nichts** (UX-12.AK-02).
func TestDieMassenvergabeGehtGanzOderGarNicht(t *testing.T) {
	db, kennungen, fremdesDeck, _ := objektstand(t)
	ablage := storage.NeueObjektmarkenablage(db)

	if _, fehler := db.ExecContext(t.Context(),
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (13, 'Blau', 'magic', 1)`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.SetzenAlle(probekontext(t), domain.MarkenartDeck,
		[]string{kennungen[domain.MarkenartDeck], fremdesDeck}, "turnier"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("mit fremdem Deck: %v", fehler)
	}

	if je, _ := ablage.JeObjekt(probekontext(t), domain.MarkenartDeck); len(je) != 0 { //nolint:errcheck // der Vergleich sagt es
		t.Errorf("halb vergeben: %v", je)
	}

	if fehler := ablage.SetzenAlle(probekontext(t), domain.MarkenartDeck,
		[]string{kennungen[domain.MarkenartDeck], "13"}, "turnier"); fehler != nil {
		t.Fatal(fehler)
	}

	if je, _ := ablage.JeObjekt(probekontext(t), domain.MarkenartDeck); len(je) != 2 { //nolint:errcheck // dito
		t.Errorf("vergeben: %v", je)
	}
}
