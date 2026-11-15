package storage_test

// **Was die Annahme meldet, wenn der Bestand nicht mitspielt** — NET-01
// (#1705).
//
// # Was hier gemessen wird
//
// **Jeder Abbruch mitten in der Annahme, und dass danach nichts steht** —
// weder Kontakt noch Recht noch verbrauchte Nonce (AK-04). Die Abbrüche sind
// Auslöser in einer echten Datenbank; ein nachgebauter Treiber prüfte sich
// selbst.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestAnnahmeOhneProfil(t *testing.T) {
	ablage := storage.NeueKontaktablage(frischeDatenbank(t))

	if fehler := ablage.Aufnehmen(t.Context(),
		eineAufnahme("n1", domain.KontaktUnbestaetigt), derPeerlauf); fehler == nil {
		t.Fatal("angenommen ohne Profil")
	}

	if _, fehler := ablage.Angenommen(t.Context(), "n1"); fehler == nil {
		t.Fatal("nachgeschlagen ohne Profil")
	}
}

func TestAnnahmeOhneTafel(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	if _, fehler := db.ExecContext(t.Context(), `DROP TABLE kontakteinladungen`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Aufnehmen(probekontext(t),
		eineAufnahme("n1", domain.KontaktUnbestaetigt), derPeerlauf); fehler == nil {
		t.Fatal("angenommen ohne Tafel")
	}

	if _, fehler := ablage.Angenommen(probekontext(t), "n1"); fehler == nil {
		t.Fatal("nachgeschlagen ohne Tafel")
	}
}

func TestAnnahmeBrichtAbOhneSpur(t *testing.T) {
	faelle := []struct {
		name      string
		ausloeser string
		ziel      error
	}{
		{
			"der Kontakt klemmt",
			`CREATE TRIGGER klemmt BEFORE INSERT ON peerkontakte
			 BEGIN SELECT RAISE(ABORT, 'klemmt'); END`,
			nil,
		},
		{
			"der Stand fällt zurück, bevor das Recht kommt",
			`CREATE TRIGGER zurueck AFTER INSERT ON peerkontakte
			 BEGIN UPDATE peerkontakte SET stand = 'unbestaetigt' WHERE id = NEW.id; END`,
			domain.ErrValidation,
		},
		{
			"die Zeile verschwindet, bevor das Recht kommt",
			`CREATE TRIGGER weg AFTER INSERT ON peerkontakte
			 BEGIN DELETE FROM peerkontakte WHERE id = NEW.id; END`,
			domain.ErrNotFound,
		},
		{
			"das Recht klemmt",
			`CREATE TRIGGER rechtklemmt BEFORE INSERT ON peerrechte
			 BEGIN SELECT RAISE(ABORT, 'klemmt'); END`,
			nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db := frischeDatenbank(t)
			ablage := storage.NeueKontaktablage(db)
			ctx := probekontext(t)

			if _, fehler := db.ExecContext(ctx, fall.ausloeser); fehler != nil {
				t.Fatal(fehler)
			}

			fehler := ablage.Aufnehmen(ctx, eineAufnahme(
				"n1", domain.KontaktBestaetigt, domain.ZweckHandel), derPeerlauf)
			if fehler == nil || fall.ziel != nil && !errors.Is(fehler, fall.ziel) {
				t.Fatalf("Aufnehmen: %v", fehler)
			}

			// **Nichts steht da**: keine Nonce, kein Kontakt.
			if angenommen, befund := ablage.Angenommen(ctx, "n1"); befund != nil || angenommen {
				t.Fatal("die Nonce ist verbraucht")
			}

			if _, fehler := ablage.Lesen(ctx, "banna"); !errors.Is(fehler, domain.ErrNotFound) {
				t.Fatalf("ein halber Kontakt steht da: %v", fehler)
			}
		})
	}
}
