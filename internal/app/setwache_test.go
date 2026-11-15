// **Die Setwache für die Flächen ohne Fenster** —
// #1336.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Set, das es gibt, kommt durch — eines, das es nicht gibt, nicht.**
//
// **Gegen echtes Storage**, weil diese Naht der Composition Root der Frage
// ist: sie baut den Setleser. Gegen ein Doppel geprüft bliebe genau das
// ungeprüft, was sie tut.
func TestDieSetwacheAmBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	// **Über den Speicher und nicht über eine Naht**: gebraucht wird ein Set,
	// nicht der Weg dorthin.
	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if err := app.SetDaSein(probekontext(t), db, setID); err != nil {
		t.Errorf("ein vorhandenes Set gab %v", err)
	}

	if err := app.SetDaSein(probekontext(t), db,
		999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Die Vorbehalte weisen eine fehlende Kennung ab.**
func TestDieVorbehalteWeisenEineFehlendeKennungAb(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, err := app.SetangabennahtUeber(frischerBestand(t), domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("SetangabennahtUeber: %v", err)
	}

	if _, fehler := weg.Vorbehalte(probekontext(t),
		999); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}
}

// **Und ein Set, das es gibt, kommt durch.**
//
// Die Gegenprobe: ohne sie wäre eine Wache, die **jede** Kennung abweist,
// grün — und die Vorbehalte wären kaputt statt streng.
func TestDieVorbehalteEinesVorhandenenSets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	weg, err := app.SetangabennahtUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("SetangabennahtUeber: %v", err)
	}

	vorbehalte, lesefehler := weg.Vorbehalte(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Vorbehalte: %v", lesefehler)
	}

	if len(vorbehalte) != 0 {
		t.Errorf("%d Vorbehalte an einem frischen Set", len(vorbehalte))
	}
}

// **Und die Suche ebenso** — `--set=999` ist eine Kennung, kein leerer Filter.
func TestDieSucheWeistEineFehlendeSetkennungAb(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSucher(&suchprobe{})

	antwort := anwendung.Suchen(t.Context(), "feuer", "", "999")

	if antwort.Fehler == nil {
		t.Fatal("eine Kennung, die auf nichts zeigt, kam ohne Fehler durch")
	}

	if antwort.Fehler.Code != "nichtGefunden" {
		t.Errorf("Code %q, erwartet nichtGefunden", antwort.Fehler.Code)
	}
}

// **Ein Lesefehler der Rohblöcke kommt durch.**
//
// # Warum dieser Lauf jetzt nötig ist
//
// **Weil die Setwache davorsteht.** Vorher erreichte eine erfundene Kennung
// diesen Zweig; jetzt wird sie vorher abgewiesen, und der Fehlerweg dahinter
// braucht einen echten Fehler — einen geschlossenen Bestand
// (#1336).
//
// Das ist die Kehrseite einer Wache: sie macht einen Zweig unerreichbar, den
// vorher ein billiger Fall betreten hat.
func TestEinLesefehlerDerVorbehalte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	weg, err := app.SetangabennahtUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("SetangabennahtUeber: %v", err)
	}

	// **Die Wache liest die Setköpfe vor den Rohblöcken.** Ein geschlossener
	// Bestand ließe sie zuerst scheitern; deshalb wird nur die Tabelle der
	// Rohblöcke weggenommen.
	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE raw_blocks`); fehler != nil {
		t.Fatalf("Tabelle wegnehmen: %v", fehler)
	}

	if _, lesefehler := weg.Vorbehalte(probekontext(t), setID); lesefehler == nil {
		t.Error("ohne die Tabelle kam kein Fehler")
	}
}
