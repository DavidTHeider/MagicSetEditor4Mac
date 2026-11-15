package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Fehlerwege der Stilablage.
//
// Mit einer echten SQLite-Datenbank nicht herzustellen: was die Prüfung
// durchlässt, schreibt sich auch. Im Betrieb entstehen sie an einer vollen
// Platte oder einer abgebrochenen Verbindung — und ein Fehlerpfad, den kein
// Test betritt, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).

func TestStilwerteMeldetDieFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
	}{
		{"die Abfrage scheitert", "stolper-fragen"},
		{"eine Zeile ist unlesbar", "stolper-umwandeln"},
		{"ein Lesefehler nach der Schleife", "stolper-lesen"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeueStilablage(db).Stilwerte(probekontext(t), 1); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

func TestStilSetzenMeldetDieFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		werte   []domain.Value
	}{
		{"das Setzen des Stils scheitert", "stolper-schreiben", nil},
		{"die Zählung der betroffenen Zeilen scheitert", "stolper-zaehlen", nil},
		{"das Räumen der Optionen scheitert", "stolper-zweites-schreiben", nil},
		{
			"das Schreiben einer Option scheitert", "stolper-stiloption",
			[]domain.Value{{Field: "hintergrund", Text: "dunkel"}},
		},
		{"das Commit scheitert", "stolper-commit", nil},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			fehler := NeueStilablage(db).StilSetzen(probekontext(t), 1, "probe", "probe-bild", fall.werte)
			if !errors.Is(fehler, errGestolpert) {
				t.Errorf("Fehler %v, erwartet war der Stolperfehler", fehler)
			}
		})
	}
}

// **Eine geschlossene Datenbank hat keine Transaktion.**
func TestStilSetzenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := Oeffnen(probekontext(t), ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := NeueStilablage(db).StilSetzen(probekontext(t), 1, "probe", "probe-bild", nil); fehler == nil {
		t.Error("kein Fehler")
	}
}

// Die Fehlerwege der beiden Einzeloperationen.
//
// **Sie sind nicht dieselben wie die von [Stilablage.StilSetzen]**: dort läuft
// eine Transaktion über drei Anweisungen, hier ist es je eine. Was bleibt, ist
// das Scheitern der Anweisung selbst.
//
// **Die Zählung der betroffenen Zeilen ist seit #1815 weg**: ein fehlendes Set
// erkennt die Profilklemme davor, und sie erkennt es auch dann, wenn es einem
// anderen Profil gehört — was die Zählung nie konnte.
func TestDieEinzeloperationenMeldenIhreFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		ruf     func(*Stilablage) error
	}{
		{
			"das Setzen scheitert", "stolper-schreiben",
			func(a *Stilablage) error { return a.StiloptionSetzen(probekontext(t), 1, "f", "w") },
		},
		{
			"das Entfernen scheitert", "stolper-schreiben",
			func(a *Stilablage) error { return a.StiloptionEntfernen(probekontext(t), 1, "f") },
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.ruf(NeueStilablage(db)))
		})
	}
}
