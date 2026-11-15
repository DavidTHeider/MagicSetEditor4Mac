package storage_test

// **Auflisten, Abmelden, Löschen** — Werkstück 4 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turniere_test.go` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Turnier
// **entsteht und läuft**, hier, was daneben mit ihm geschieht.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Turniere lassen sich auflisten — je Spiel und alle.**
func TestTurniereLassenSichAuflistenUndFiltern(t *testing.T) {
	ablage := turnierablage(t)

	for _, eines := range []domain.Turnier{
		{Name: "Küchentisch", Spiel: "magic"},
		{Name: "Anderswo", Spiel: "vs"},
	} {
		if _, err := ablage.Anlegen(probekontext(t), eines); err != nil {
			t.Fatalf("%s anlegen: %v", eines.Name, err)
		}
	}

	alle, err := ablage.Turniere(probekontext(t), "")
	if err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	// **Nach Namen geordnet** — „Anderswo" vor „Küchentisch".
	if len(alle) != 2 || alle[0].Name != "Anderswo" {
		t.Errorf("%+v", alle)
	}

	magic, err := ablage.Turniere(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	if len(magic) != 1 || magic[0].Name != "Küchentisch" {
		t.Errorf("%+v", magic)
	}
}

// **Ein Teilnehmer steigt aus — und wieder ein.**
func TestEinTeilnehmerSteigtAusUndWiederEin(t *testing.T) {
	ablage := turnierablage(t)
	turnier := turnierMitTeilnehmern(t, ablage, "Ada")

	wer := turnier.Teilnehmer[0].ID

	if err := ablage.Aussteigen(probekontext(t), wer, true); err != nil {
		t.Fatalf("aussteigen: %v", err)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !gelesen.Teilnehmer[0].Ausgestiegen {
		t.Error("er ist nicht ausgestiegen")
	}

	if wieder := ablage.Aussteigen(probekontext(t), wer, false); wieder != nil {
		t.Fatalf("wieder einsteigen: %v", wieder)
	}

	gelesen, err = ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Teilnehmer[0].Ausgestiegen {
		t.Error("er ist immer noch ausgestiegen")
	}
}

// **Ein Teilnehmer, den es nicht gibt, steigt nicht aus.**
func TestEinFehlenderTeilnehmerSteigtNichtAus(t *testing.T) {
	if err := turnierablage(t).Aussteigen(probekontext(t), 999, true); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Ein Turnier lässt sich löschen — samt seinem Verlauf.**
//
// Ein Turnier ohne seine Runden wäre kein Turnier, sondern ein Name.
func TestEinTurnierLaesstSichSamtVerlaufLoeschen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueTurnierablage(db)

	turnier := turnierMitTeilnehmernAn(t, ablage, "Ada", "Bert")

	if _, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	if err := ablage.Loeschen(probekontext(t), turnier.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	for _, tabelle := range []string{
		"turnierteilnehmer", "turnierrunden", "turnierpaarungen",
	} {
		var zeilen int

		if err := db.QueryRowContext(probekontext(t),
			"SELECT count(*) FROM "+tabelle).Scan(&zeilen); err != nil { //nolint:gosec // fester Name
			t.Fatalf("%s zählen: %v", tabelle, err)
		}

		if zeilen != 0 {
			t.Errorf("%s hat noch %d Zeilen", tabelle, zeilen)
		}
	}
}

// **Ein Turnier, das es nicht gibt, lässt sich nicht löschen.**
func TestEinFehlendesTurnierLaesstSichNichtLoeschen(t *testing.T) {
	if err := turnierablage(t).Loeschen(probekontext(t), 999); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Der Name einer eigenen Karte lässt sich lesen.**
//
// Die Turnierleitung braucht ihn: der Abzug einer Deckliste ist Text und kein
// Verweis. Ohne den Namen bliebe von einer eigenen Karte nichts übrig, sobald
// jemand sie löscht.
func TestDerNameEinerKarteLaesstSichLesen(t *testing.T) {
	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	name, err := leser.NameZurKarte(probekontext(t), karteMitNamen(t, db, "Lightning Bolt"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if name != "Lightning Bolt" {
		t.Errorf("der Name ist %q", name)
	}
}

// **Eine Karte, die es nicht gibt, hat keinen Namen — und das ist ein Nein.**
func TestEineFehlendeKarteHatKeinenNamen(t *testing.T) {
	if _, err := storage.NeuerSetLeser(frischeDatenbank(t)).NameZurKarte(
		probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Eine namenlose Karte gibt eine leere Zeichenkette, keinen Fehler.**
//
// Eine frisch angelegte Karte hat keinen Namen, und ein Turnier abzuweisen,
// weil eine Karte namenlos ist, wäre die falsche Strenge.
func TestEineNamenloseKarteGibtEineLeereZeichenkette(t *testing.T) {
	db := frischeDatenbank(t)

	name, err := storage.NeuerSetLeser(db).NameZurKarte(
		probekontext(t), karteMitNamen(t, db, ""))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if name != "" {
		t.Errorf("der Name ist %q, erwartet leer", name)
	}
}

// turnierMitTeilnehmernAn legt ein Turnier an und meldet Namen an.
//
// **Wie `turnierMitTeilnehmern`, nur mit einer eigenen Ablage** — dieser Lauf
// braucht die Datenbank selbst, um die Kaskade nachzuzählen.
func turnierMitTeilnehmernAn(
	t *testing.T, ablage *storage.Turnierablage, namen ...string,
) domain.Turnier {
	t.Helper()

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for _, name := range namen {
		angemeldet, anmeldefehler := ablage.Anmelden(probekontext(t), turnier.ID,
			domain.Teilnehmer{Name: name})
		if anmeldefehler != nil {
			t.Fatalf("%s anmelden: %v", name, anmeldefehler)
		}

		turnier.Teilnehmer = append(turnier.Teilnehmer, angemeldet)
	}

	return turnier
}

// karteMitNamen legt ein Set mit einer Karte an — **mit oder ohne Namen**.
//
// **Ein leerer Name legt keinen Feldwert an.** So sieht der Leser den Fall,
// den eine frisch angelegte Karte wirklich hat: keine Zeile in `card_values`,
// nicht eine leere.
func karteMitNamen(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	ergebnis, err := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, stil, geoeffnet, profil_id)
		 VALUES ('Probe', 'magic', 'm15', datetime('now'), 1)`)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	setID, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		t.Fatalf("die Setkennung: %v", kennungsfehler)
	}

	ergebnis, err = db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (?, 0)`, setID)
	if err != nil {
		t.Fatalf("die Karte anlegen: %v", err)
	}

	karteID, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		t.Fatalf("die Kartenkennung: %v", kennungsfehler)
	}

	if name != "" {
		if _, err := db.ExecContext(probekontext(t),
			`INSERT INTO card_values (card_id, feld, wert, reihenfolge)
			 VALUES (?, 'name', ?, 0)`, karteID, name); err != nil {
			t.Fatalf("den Namen setzen: %v", err)
		}
	}

	return karteID
}
