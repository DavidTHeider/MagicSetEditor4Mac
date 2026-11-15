package storage

// **Jeder Fehlerweg der Objektmarken meldet sich** — UX-12
// (#1684).
//
// Dieselbe Bauform wie `markenfehler_intern_test.go`: ein Kontext ohne
// Profil, eine fehlende Tafel, ein Auslöser, der jedes Schreiben abweist, und
// ein `Scan` gegen die falsche Spaltenzahl.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// jederWeg ruft jeden Weg der Ablage und zählt, wer ohne Fehler zurückkam.
func jederWeg(t *testing.T, ablage *Objektmarkenablage, ohneProfil bool) []string {
	t.Helper()

	ctx := probekontext(t)
	if ohneProfil {
		ctx = t.Context()
	}

	ohneFehler := []string{}
	pruefe := func(name string, fehler error) {
		if fehler == nil {
			ohneFehler = append(ohneFehler, name)
		}
	}

	_, fehler := ablage.Vorhanden(ctx, domain.MarkenartDeck, "1")
	pruefe("Vorhanden", fehler)
	pruefe("Setzen", ablage.Setzen(ctx, domain.MarkenartDeck, "1", "x"))
	pruefe("Setzen Karte", ablage.Setzen(ctx, domain.MarkenartKarte, "1", "x"))
	pruefe("SetzenAlle", ablage.SetzenAlle(ctx, domain.MarkenartDeck, []string{"1"}, "x"))
	pruefe("Entfernen", ablage.Entfernen(ctx, domain.MarkenartDeck, "1", "x"))
	pruefe("Entfernen Karte", ablage.Entfernen(ctx, domain.MarkenartKarte, "1", "x"))
	_, fehler = ablage.JeObjekt(ctx, domain.MarkenartDeck)
	pruefe("JeObjekt", fehler)
	_, fehler = ablage.Katalog(ctx)
	pruefe("Katalog", fehler)
	_, fehler = ablage.Umbenennen(ctx, "x", "y", true)
	pruefe("Umbenennen", fehler)
	_, fehler = ablage.UeberallEntfernen(ctx, "x")
	pruefe("UeberallEntfernen", fehler)

	return ohneFehler
}

func TestOhneProfilGehtKeinWeg(t *testing.T) {
	ohneFehler := jederWeg(t, NeueObjektmarkenablage(migrierteDatenbank(t)), true)
	if len(ohneFehler) != 0 {
		t.Errorf("ohne Profil kamen durch: %v", ohneFehler)
	}
}

func TestOhneTafelnGehtKeinWeg(t *testing.T) {
	db := migrierteDatenbank(t)

	for _, anweisung := range []string{`DROP TABLE objektmarken`, `DROP TABLE karten_marken`, `DROP TABLE decks`} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if ohneFehler := jederWeg(t, NeueObjektmarkenablage(db), false); len(ohneFehler) != 0 {
		t.Errorf("ohne Tafeln kamen durch: %v", ohneFehler)
	}
}

// **Ein Schreiben, das der Bestand abweist**, kommt als Fehler an — auch
// mitten im Umbenennen, wenn das Zählen noch ging.
func TestEinAbgewiesenesSchreibenMeldetSich(t *testing.T) {
	db := migrierteDatenbank(t)
	ablage := NeueObjektmarkenablage(db)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO decks (id, name, spiel, profil_id) VALUES (1, 'Rot', 'magic', 1)`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Setzen(probekontext(t), domain.MarkenartDeck, "1", "x"); fehler != nil {
		t.Fatal(fehler)
	}

	for _, ausloeser := range []string{
		`CREATE TRIGGER stolper_neu BEFORE INSERT ON objektmarken BEGIN SELECT RAISE(ABORT, 'nein'); END`,
		`CREATE TRIGGER stolper_weg BEFORE DELETE ON objektmarken BEGIN SELECT RAISE(ABORT, 'nein'); END`,
		`CREATE TRIGGER stolper_karte BEFORE INSERT ON karten_marken BEGIN SELECT RAISE(ABORT, 'nein'); END`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), ausloeser); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := ablage.SetzenAlle(probekontext(t), domain.MarkenartDeck, []string{"1"}, "y"); fehler == nil {
		t.Error("die Massenvergabe kam durch")
	}

	if _, fehler := ablage.Umbenennen(probekontext(t), "x", "y", false); fehler == nil {
		t.Error("das Umbenennen kam durch")
	}

	if _, fehler := db.ExecContext(probekontext(t), `DROP TRIGGER stolper_neu`); fehler != nil {
		t.Fatal(fehler)
	}

	// **Jetzt geht das Einfügen, und das Entfernen scheitert** — der
	// zweite Schritt des Umbenennens.
	if _, fehler := ablage.Umbenennen(probekontext(t), "x", "y", false); fehler == nil {
		t.Error("das Umbenennen kam ohne Entfernen durch")
	}

	if _, fehler := ablage.UeberallEntfernen(probekontext(t), "x"); fehler == nil {
		t.Error("das Entfernen kam durch")
	}
}

func TestEinMarkenscanGegenDieFalscheSpaltenzahlMeldetSichAuchAmObjekt(t *testing.T) {
	ablage := NeueObjektmarkenablage(offeneDatenbank(t))

	if _, fehler := ablage.markenzeilen(probekontext(t), `SELECT 1`); fehler == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}

	if _, fehler := ablage.posten(probekontext(t), `SELECT 1`); fehler == nil {
		t.Error("ein Zählscan gegen eine Spalte kam durch")
	}

	if _, fehler := ablage.posten(probekontext(t), `SELECT * FROM gibt_es_nicht`); fehler == nil {
		t.Error("eine fehlende Tafel kam durch")
	}
}
