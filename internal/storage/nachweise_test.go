package storage_test

// **Nachweise im Bestand** — Werkstück O1 von Phase 5.2
// (#1425).
//
// # Was hier geprüft wird
//
// Ein Nachweis überlebt den Prozess — ein Handel in Phase 5.6 wartet auf
// einen Partner —, und das Einlösen ist **atomar**: von zwei, die zugleich
// kommen, bekommt genau einer den Nachweis, der andere den Konflikt.

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func probenachweis(kennung string) domain.Nachweis {
	jetzt := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

	return domain.Nachweis{
		Kennung: kennung, NutzerID: 1, ProfilID: 1,
		Zweck: "handel.angebot", Ziel: "angebot-42", Revision: 5,
		Grenze:      domain.Kostengrenze{Betrag: 1500, Waehrung: "EUR"},
		Ausgestellt: jetzt, Ablauf: jetzt.Add(10 * time.Minute),
	}
}

func TestEinNachweisUeberlebtDenWegDurchDenBestand(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueNachweisablage(db)

	hin := probenachweis("n-1")
	if fehler := ablage.NachweisAnlegen(t.Context(), hin); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	zurueck, fehler := ablage.Nachweis(t.Context(), "n-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Zeiten in UTC verglichen** — SQLite kennt keine Zone, und die Ablage
	// gibt zurück, was sie bekam.
	if !zurueck.Ausgestellt.Equal(hin.Ausgestellt) || !zurueck.Ablauf.Equal(hin.Ablauf) {
		t.Fatalf("Zeiten: %v/%v, erwartet %v/%v", zurueck.Ausgestellt, zurueck.Ablauf, hin.Ausgestellt, hin.Ablauf)
	}

	zurueck.Ausgestellt, zurueck.Ablauf = hin.Ausgestellt, hin.Ablauf
	if zurueck != hin {
		t.Fatalf("Nachweis %+v, erwartet %+v", zurueck, hin)
	}

	// **Dieselbe Kennung zweimal ist ein Konflikt** — eine Kennung ist, was
	// der Aufruf vorzeigt, und zwei Nachweise dahinter wären zwei Antworten.
	if fehler := ablage.NachweisAnlegen(t.Context(), hin); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("zweimal anlegen: %v, erwartet ErrConflict", fehler)
	}

	if _, fehler := ablage.Nachweis(t.Context(), "gibts-nicht"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("unbekannt: %v, erwartet ErrNotFound", fehler)
	}
}

func TestEinNachweisWirdGenauEinmalEingeloest(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueNachweisablage(db)

	if fehler := ablage.NachweisAnlegen(t.Context(), probenachweis("n-1")); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Zwei Einlöser zugleich** — genau einer gewinnt.
	var (
		gruppe    sync.WaitGroup
		schloss   sync.Mutex
		gelungen  int
		konflikte int
	)

	for i := 0; i < 8; i++ {
		gruppe.Add(1)

		go func() {
			defer gruppe.Done()

			fehler := ablage.NachweisEinloesen(t.Context(), "n-1")

			schloss.Lock()
			defer schloss.Unlock()

			switch {
			case fehler == nil:
				gelungen++
			case errors.Is(fehler, domain.ErrConflict):
				konflikte++
			default:
				t.Errorf("einlösen: %v", fehler)
			}
		}()
	}

	gruppe.Wait()

	if gelungen != 1 || konflikte != 7 {
		t.Fatalf("%d gelungen, %d Konflikte — erwartet 1 und 7", gelungen, konflikte)
	}

	zurueck, fehler := ablage.Nachweis(t.Context(), "n-1")
	if fehler != nil || !zurueck.Eingeloest {
		t.Fatalf("nach dem Einlösen: %+v, %v", zurueck, fehler)
	}

	if fehler := ablage.NachweisEinloesen(t.Context(), "gibts-nicht"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("unbekannt einlösen: %v, erwartet ErrNotFound", fehler)
	}
}

// **Ein Nachweis ohne Grenze steht ohne Betrag im Bestand** — und kommt so
// zurück.
func TestEinNachweisOhneGrenze(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueNachweisablage(db)

	hin := probenachweis("n-2")
	hin.Grenze = domain.Kostengrenze{}

	if fehler := ablage.NachweisAnlegen(t.Context(), hin); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	zurueck, fehler := ablage.Nachweis(t.Context(), "n-2")
	if fehler != nil || zurueck.Grenze != (domain.Kostengrenze{}) {
		t.Fatalf("Grenze %+v, %v", zurueck.Grenze, fehler)
	}
}

// **Ein Nachweis hängt an seinem Profil** — geht das Profil, geht er mit.
func TestEinNachweisGehtMitSeinemProfil(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueNachweisablage(db)

	if fehler := ablage.NachweisAnlegen(t.Context(), probenachweis("n-3")); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), "DELETE FROM profile WHERE id = 1"); fehler != nil {
		t.Fatalf("Profil löschen: %v", fehler)
	}

	if _, fehler := ablage.Nachweis(t.Context(), "n-3"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("nach dem Löschen des Profils: %v, erwartet ErrNotFound", fehler)
	}
}

// **Jeder Weg meldet einen kaputten Bestand.**
func TestDieNachweisablageMeldetEinenKaputtenBestand(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueNachweisablage(kaputterBestand(t))

	if fehler := ablage.NachweisAnlegen(t.Context(), probenachweis("n-1")); fehler == nil {
		t.Error("Anlegen im kaputten Bestand ging durch")
	}

	if _, fehler := ablage.Nachweis(t.Context(), "n-1"); fehler == nil {
		t.Error("Lesen im kaputten Bestand ging durch")
	}

	if fehler := ablage.NachweisEinloesen(t.Context(), "n-1"); fehler == nil {
		t.Error("Einlösen im kaputten Bestand ging durch")
	}
}

// **Eine Zeile mit einer Zeit, die keine ist, kommt nicht durch** — die
// Tafel ist STRICT, aber TEXT hält jede Zeichenkette.
func TestEinNachweisMitUnlesbarerZeitIstEinFehler(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueNachweisablage(db)

	for kennung, zeiten := range map[string][2]string{
		"ausgestellt-kaputt": {"keine Zeit", "2026-09-08T12:10:00Z"},
		"ablauf-kaputt":      {"2026-09-08T12:00:00Z", "keine Zeit"},
	} {
		if _, fehler := db.ExecContext(t.Context(), `
			INSERT INTO freigabenachweise
			  (kennung, nutzer_id, profil_id, zweck, ziel, revision, ausgestellt, ablauf)
			VALUES (?, 1, 1, 'zweck', 'ziel', 1, ?, ?)`, kennung, zeiten[0], zeiten[1]); fehler != nil {
			t.Fatalf("%s: %v", kennung, fehler)
		}

		if _, fehler := ablage.Nachweis(t.Context(), kennung); fehler == nil {
			t.Errorf("%s: die unlesbare Zeile kam durch", kennung)
		}
	}
}

// **Ein eingelöster Nachweis bleibt als eingelöst lesbar.**
func TestEinEingeloesterNachweisBleibtEingeloest(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueNachweisablage(frischeDatenbank(t))

	verbraucht := probenachweis("n-1")
	verbraucht.Eingeloest = true

	if fehler := ablage.NachweisAnlegen(t.Context(), verbraucht); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gelesen, fehler := ablage.Nachweis(t.Context(), "n-1")
	if fehler != nil || !gelesen.Eingeloest {
		t.Errorf("gelesen %+v, %v — erwartet eingelöst", gelesen, fehler)
	}

	if fehler := ablage.NachweisEinloesen(t.Context(), "n-1"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("noch einmal einlösen: %v, erwartet ErrConflict", fehler)
	}
}
