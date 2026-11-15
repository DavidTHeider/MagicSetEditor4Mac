package storage_test

// **Die Fehlerwege der Leihablage** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Eine eigene Datei, weil es eine eigene Frage ist**: `deckleihen_test.go`
// fragt, ob eine Leihe zurückkommt, wie sie hineinging; hier steht, was
// geschieht, wenn der Weg dorthin nicht trägt. Jeder Fehlerweg wird betreten
// (§3).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Profil im Kontext geht gar nichts** — die Klemme aus #1351, auf
// allen drei Wegen.
func TestOhneProfilGehtKeineLeihe(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueLeihablage(frischeDatenbank(t))
	ohne := t.Context()

	wege := map[string]func() error{
		"schreiben": func() error {
			return ablage.LeiheSchreiben(ohne, einLeihstand())
		},
		"lesen": func() error {
			_, fehler := ablage.Leihe(ohne, "l1")

			return fehler
		},
		"listen": func() error {
			_, fehler := ablage.Leihen(ohne)

			return fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); fehler == nil {
				t.Fatal("ohne Profil kam kein Fehler")
			}
		})
	}
}

// **Eine geschlossene Datenbank wird gemeldet und nicht verschluckt.**
func TestEineGeschlosseneDatenbankMeldetSichBeiDerLeihe(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeueLeihablage(db), probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	wege := map[string]func() error{
		"schreiben": func() error {
			return ablage.LeiheSchreiben(ctx, einLeihstand())
		},
		"lesen": func() error {
			_, fehler := ablage.Leihe(ctx, "l1")

			return fehler
		},
		"listen": func() error {
			_, fehler := ablage.Leihen(ctx)

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

// **Eine unlesbare Zeile wird gemeldet** — beim Lesen und beim Listen.
//
// **Sie kann entstehen**: eine Sicherung von Hand, eine Wanderung, ein fremdes
// Werkzeug. Was dann kommt, ist eine Meldung und keine halbe Leihe (§8).
func TestEineUnlesbareLeiheWirdGemeldet(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ dokument, abzug string }{
		"kaputtes Dokument": {dokument: "{", abzug: "{}"},
		"kaputter Abzug":    {dokument: "{}", abzug: "{"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := frischeDatenbank(t)
			ablage, ctx := storage.NeueLeihablage(db), probekontext(t)

			if _, fehler := db.ExecContext(ctx,
				`INSERT INTO deckleihen
				   (kennung, profil_id, dokument, abzug, verliehen)
				 VALUES ('kaputt', 1, ?, ?, 1)`,
				fall.dokument, fall.abzug); fehler != nil {
				t.Fatalf("legen: %v", fehler)
			}

			if _, fehler := ablage.Leihe(ctx, "kaputt"); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("lesen: erwartet ErrValidation, bekam %v", fehler)
			}

			if _, fehler := ablage.Leihen(ctx); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("listen: erwartet ErrValidation, bekam %v", fehler)
			}
		})
	}
}

// **Und die Liste gibt zurück, was darin steht.**
func TestDieListeGibtJedeLeihe(t *testing.T) {
	t.Parallel()

	ablage, ctx := storage.NeueLeihablage(frischeDatenbank(t)), probekontext(t)

	for _, kennung := range []string{"l1", "l2"} {
		if fehler := ablage.LeiheSchreiben(ctx, einLeihstand(
			func(s *domain.Leihstand) { s.Leihe.Kennung = kennung },
		)); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	leihen, fehler := ablage.Leihen(ctx)
	if fehler != nil {
		t.Fatalf("listen: %v", fehler)
	}

	if len(leihen) != 2 {
		t.Fatalf("erwartet zwei Leihen, bekam %d", len(leihen))
	}

	if leihen[0].Abzug.Abdruck == "" {
		t.Error("der Abzug fehlt in der Liste")
	}
}

// **Ohne Profil wird nichts vergessen.**
func TestOhneProfilWirdNichtsVergessen(t *testing.T) {
	t.Parallel()

	if fehler := storage.NeuePartieablage(frischeDatenbank(t)).Vergessen(
		t.Context(), "p1"); fehler == nil {
		t.Fatal("ohne Profil kam kein Fehler")
	}
}

// **Eine geschlossene Datenbank meldet sich auf jedem Freigabeweg.**
func TestEineGeschlosseneDatenbankMeldetSichBeiDerFreigabe(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeuePartieablage(db), probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	wege := map[string]func() error{
		"lesen": func() error {
			_, fehler := ablage.Freigaben(ctx, "p1")

			return fehler
		},
		"freigeben": func() error {
			return ablage.Freigeben(ctx, "p1",
				domain.Offenlegung{Von: "anna", An: []string{"ben"}})
		},
		"löschen": func() error {
			return ablage.Freigabeloeschen(ctx, "p1", "anna")
		},
		"vergessen": func() error { return ablage.Vergessen(ctx, "p1") },
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
