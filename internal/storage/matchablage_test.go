package storage_test

// **Bindung und Meldung im Bestand** — Werkstück G7 von Phase 5.5
// (#1464).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einMatchbestand gibt eine Ablage samt Turnier, Runde und Paarung.
func einMatchbestand(t *testing.T) (*storage.Matchablage, context.Context) {
	t.Helper()

	db, ctx := frischeDatenbank(t), probekontext(t)

	anweisungen := []string{
		`INSERT INTO turniere (id, profil_id, name, spiel, fassung)
		 VALUES (1, 1, 'Freitag', 'magic', 3)`,
		`INSERT INTO turnierteilnehmer (id, turnier_id, name, stelle)
		 VALUES (10, 1, 'Anna', 1), (11, 1, 'Bert', 2)`,
		`INSERT INTO turnierrunden (id, turnier_id, nummer) VALUES (100, 1, 1)`,
		`INSERT INTO turnierpaarungen (id, runde_id, a, b, stelle)
		 VALUES (7, 100, 10, 11, 1)`,
		// **Und die Partie, an die gebunden wird** — seit #1815 fragt die
		// Ablage, ob sie dem Profil gehört; der Dienst holte sie ohnehin
		// vorher (`Matchleitung.Binden`).
		`INSERT INTO partien (id, profil_id, spiel, stand, zustand)
		 VALUES ('p1', 1, 'magic', 'laeuft', '{}'), ('p2', 1, 'magic', 'laeuft', '{}')`,
	}

	for _, anweisung := range anweisungen {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}

	return storage.NeueMatchablage(db), ctx
}

// eineTestbindung gibt die gewöhnliche Bindung.
func eineTestbindung() domain.Matchbindung {
	return domain.Matchbindung{
		TurnierID: 1, PaarungID: 7, Fassung: 3, PartieID: "p1",
		Hoechstens: 3,
		Sitze: []domain.Matchsitz{
			{TeilnehmerID: 10, Sitz: "anna", Deckabdruck: "abdruckA"},
			{
				TeilnehmerID: 11, Sitz: "bert", Peer: "12D3KooWBert",
				Deckabdruck: "abdruckB",
			},
		},
	}
}

func TestEineBindungKommtZurueckWieSieHineinging(t *testing.T) {
	t.Parallel()

	ablage, ctx := einMatchbestand(t)

	if fehler := ablage.BindungSchreiben(
		ctx, eineTestbindung()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	zurueck, fehler := ablage.Bindung(ctx, 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	switch {
	case zurueck.PartieID != "p1" || zurueck.Fassung != 3:
		t.Errorf("die Bindung ist %+v", zurueck)
	case len(zurueck.Sitze) != 2:
		t.Fatalf("es sind %d Sitze", len(zurueck.Sitze))
	case zurueck.Sitze[1].Peer != "12D3KooWBert":
		t.Errorf("die Peerkennung ist %q", zurueck.Sitze[1].Peer)
	case zurueck.Hoechstens != 3:
		t.Errorf("die Spielgrenze ist %d", zurueck.Hoechstens)
	}
}

// **Eine zweite Bindung ersetzt die erste — samt ihren Meldungen.**
//
// **Die Meldungen gehören zur alten Partie.** Wer sie stehen ließe, hätte
// nach einem Neuanfang Aussagen über einen Strom, den es nicht mehr gibt.
func TestEineZweiteBindungRaeumtDieMeldungenWeg(t *testing.T) {
	t.Parallel()

	ablage, ctx := einMatchbestand(t)

	if fehler := ablage.BindungSchreiben(
		ctx, eineTestbindung()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := ablage.MeldungSchreiben(ctx, eineTestmeldung("anna")); fehler != nil {
		t.Fatalf("melden: %v", fehler)
	}

	zweite := eineTestbindung()
	zweite.PartieID = "p2"

	if fehler := ablage.BindungSchreiben(ctx, zweite); fehler != nil {
		t.Fatalf("zweite Bindung: %v", fehler)
	}

	meldungen, fehler := ablage.Meldungen(ctx, 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(meldungen) != 0 {
		t.Fatalf("die alten Meldungen stehen noch da: %v", meldungen)
	}
}

// eineTestmeldung gibt die Meldung eines Sitzes.
func eineTestmeldung(von string) domain.Matchmeldung {
	return domain.Matchmeldung{
		PaarungID: 7, Fassung: 3, PartieID: "p1", Von: von,
		Ergebnis: domain.Ergebnis{SpieleA: 2, SpieleB: 1},
		Kette:    "letztesglied",
	}
}

// **Eine je Sitz, und die letzte gilt.**
func TestDieLetzteMeldungEinesSitzesGilt(t *testing.T) {
	t.Parallel()

	ablage, ctx := einMatchbestand(t)

	if fehler := ablage.BindungSchreiben(
		ctx, eineTestbindung()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	for _, spiele := range []int{1, 2} {
		meldung := eineTestmeldung("anna")
		meldung.Ergebnis.SpieleA = spiele

		if fehler := ablage.MeldungSchreiben(ctx, meldung); fehler != nil {
			t.Fatalf("melden: %v", fehler)
		}
	}

	meldungen, fehler := ablage.Meldungen(ctx, 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(meldungen) != 1 || meldungen[0].Ergebnis.SpieleA != 2 {
		t.Fatalf("die Meldungen sind %+v", meldungen)
	}
}

// **Ohne Bindung gibt es nichts zu lesen — und das wird gesagt.**
func TestOhneBindungWirdEsGesagt(t *testing.T) {
	t.Parallel()

	ablage, ctx := einMatchbestand(t)

	if _, fehler := ablage.Bindung(ctx, 7); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}

	meldungen, fehler := ablage.Meldungen(ctx, 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if meldungen == nil || len(meldungen) != 0 {
		t.Fatalf("die Liste ist %v", meldungen)
	}
}

// **Eine tote Datenbank meldet sich auf jedem Weg.**
func TestEineGeschlosseneDatenbankMeldetSichBeimMatch(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeueMatchablage(db), probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	wege := map[string]func() error{
		"binden": func() error {
			return ablage.BindungSchreiben(ctx, eineTestbindung())
		},
		"lesen": func() error {
			_, fehler := ablage.Bindung(ctx, 7)

			return fehler
		},
		"melden": func() error {
			return ablage.MeldungSchreiben(ctx, eineTestmeldung("anna"))
		},
		"Meldungen lesen": func() error {
			_, fehler := ablage.Meldungen(ctx, 7)

			return fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); fehler == nil {
				t.Fatal("die geschlossene Datenbank kam nicht an")
			}
		})
	}
}

// **Eine unlesbare Sitzspalte wird gemeldet.**
func TestUnlesbareSitzeWerdenGemeldet(t *testing.T) {
	t.Parallel()

	db, ctx := frischeDatenbank(t), probekontext(t)
	ablage := storage.NeueMatchablage(db)

	anweisungen := []string{
		`INSERT INTO turniere (id, profil_id, name, spiel, fassung)
		 VALUES (1, 1, 'Freitag', 'magic', 3)`,
		`INSERT INTO turnierteilnehmer (id, turnier_id, name, stelle)
		 VALUES (10, 1, 'Anna', 1), (11, 1, 'Bert', 2)`,
		`INSERT INTO turnierrunden (id, turnier_id, nummer) VALUES (100, 1, 1)`,
		`INSERT INTO turnierpaarungen (id, runde_id, a, b, stelle)
		 VALUES (7, 100, 10, 11, 1)`,
		`INSERT INTO matchbindungen
		   (paarung_id, turnier_id, fassung, partie_id, sitze)
		 VALUES (7, 1, 3, 'p1', '{')`,
	}

	for _, anweisung := range anweisungen {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}

	if _, fehler := ablage.Bindung(ctx, 7); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}
