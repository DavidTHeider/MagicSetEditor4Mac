package app_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Wortlistennaht bedient die Flächen ohne Fenster** —
// #1074.
//
// §16: was das Fenster kann, können Kommandozeile, MCP und REST. Für eine
// Wortliste ist das nicht bloß Formsache — sie sagt einem Sprachmodell,
// welche Untertypen ein Spiel **kennt**, statt es raten zu lassen (§15).
func TestDieWortlistennahtGibtDieBereiche(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	_, karteID := setMitKarte(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	// **Ohne Spielpaket gibt es keine Bereiche, und das ist kein Fehler**:
	// die allermeisten Felder haben keine Auswahl (§8).
	bereiche, lesefehler := naht.Wortlisten(probekontext(t), karteID, "type")
	if lesefehler != nil {
		t.Fatalf("Wortlisten: %v", lesefehler)
	}

	if len(bereiche) != 0 {
		t.Errorf("ohne Spielpaket kamen %+v", bereiche)
	}
}

// **Eine Wahl ohne Bereich wird abgewiesen und schreibt nichts.**
func TestEineWahlOhneBereichWirdAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := setMitKarte(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if err := naht.Waehlen(
		probekontext(t), setID, karteID, "type", 0, []int{0}); err == nil {
		t.Error("die Wahl ins Leere kam ohne Fehler durch")
	}
}

// **Ein leerer Feldname wird abgewiesen, bevor irgendetwas gelesen wird.**
//
// Dieselbe Regel wie bei [app.App.FeldSetzen]: eine Zeile ohne Feldnamen
// findet niemand wieder.
func TestEineWahlOhneFeldnamenWirdAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := setMitKarte(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if err := naht.Waehlen(probekontext(t), setID, karteID, "", 0, []int{0}); err == nil {
		t.Error("der leere Feldname kam durch")
	}
}

// **Eine Karte, die es nicht gibt, kommt als Fehler zurück.**
func TestDieWortlistennahtMeldetEineFehlendeKarte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, _ := setMitKarte(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if _, err := naht.Wortlisten(probekontext(t), 9999, "type"); err == nil {
		t.Error("die fehlende Karte kam beim Lesen durch")
	}

	if err := naht.Waehlen(
		probekontext(t), setID, 9999, "type", 0, []int{0}); err == nil {
		t.Error("die fehlende Karte kam beim Schreiben durch")
	}
}

// **Ohne Zuhause kommt die Naht nicht zusammen** — beide Hälften.
func TestOhneZuhauseKommtDieWortlistennahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.WortlistennahtUeber(
		frischerBestand(t), "kommandozeile"); fehler == nil {
		t.Fatal("kein Fehler ohne Zuhause")
	}
}

// **Der ganze Weg über die Naht — mit einem echten Spielpaket.**
//
// Die Läufe darüber prüfen die Abweisungen; dieser prüft, dass eine Wahl
// wirklich schreibt. Ohne ihn stünde die Naht mit voller Abdeckung da und
// hätte nie einen Wert bewegt (§10, erste Dimension).
func TestDieWortlistennahtSchreibtEineWahl(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	bereiche, lesefehler := naht.Wortlisten(probekontext(t), karteID, "text")
	if lesefehler != nil {
		t.Fatalf("Wortlisten: %v", lesefehler)
	}

	if len(bereiche) != 1 || bereiche[0].Liste != "probe-art" {
		t.Fatalf("die Bereiche sind %+v", bereiche)
	}

	if err := naht.Waehlen(probekontext(t), setID, karteID, "text", 0, []int{2}); err != nil {
		t.Fatalf("Waehlen: %v", err)
	}

	karte, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("KarteLesen: %v", err)
	}

	erwartet := "<word-list-probe-art>Große Spruch</word-list-probe-art>"
	for _, wert := range karte.Values {
		if wert.Field == "text" && wert.Text != erwartet {
			t.Errorf("im Feld steht\n  %q\nerwartet\n  %q", wert.Text, erwartet)
		}
	}
}

// **Ein abgebrochener Kontext kommt durch die Naht.**
func TestDieWortlistennahtGibtDenAbbruchWeiter(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	_, karteID := saatsetImBestand(t, db)

	naht, fehler := app.WortlistennahtUeber(db, "kommandozeile")
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ctx, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	if _, err := naht.Wortlisten(ctx, karteID, "text"); !errors.Is(err, context.Canceled) {
		t.Errorf("der Abbruch kam als %v durch", err)
	}
}

// saatInsZuhause kopiert `saat/` in das Paketverzeichnis eines Zuhauses.
func saatInsZuhause(t *testing.T, zuhause string) {
	t.Helper()

	quelle, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	ziel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete")

	if kopierfehler := os.CopyFS(ziel, os.DirFS(quelle)); kopierfehler != nil {
		t.Fatalf("Saat kopieren: %v", kopierfehler)
	}
}

// saatsetImBestand legt ein Set der Saat mit einer ausgezeichneten Karte an.
func saatsetImBestand(t *testing.T, db *sql.DB) (setID, karteID int64) {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe", Style: "probe-schlicht"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	schreiber := storage.NeuerKartenschreiber(db)

	karteID, kartenfehler := schreiber.KarteAnlegen(probekontext(t), setID)
	if kartenfehler != nil {
		t.Fatalf("Karte anlegen: %v", kartenfehler)
	}

	if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, "text",
		"<word-list-probe-art>Große Kreatur</word-list-probe-art>"); err != nil {
		t.Fatalf("Feld setzen: %v", err)
	}

	return setID, karteID
}
