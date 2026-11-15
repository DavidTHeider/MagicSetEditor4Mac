package storage_test

// **Was die Einlösung meldet, wenn der Bestand nicht mitspielt** — NET-01b
// (#1954).
//
// **Nach jedem Abbruch steht nichts**: kein Kontakt, und die Nonce ist noch
// einlösbar. Die Abbrüche sind Auslöser in einer echten Datenbank — dieselbe
// Vorrichtung wie bei der Annahme.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinloesungOhneProfil(t *testing.T) {
	ablage := storage.NeueKontaktablage(frischeDatenbank(t))

	if fehler := ablage.Ausgestellt(t.Context(), "n1", domain.ZweckHandel,
		einloesezeit, einloesezeit); fehler == nil {
		t.Error("ausgestellt ohne Profil")
	}

	if fehler := ablage.Einloesen(t.Context(), "n1", einEinloesender("bbert"), einloesezeit); fehler == nil {
		t.Error("eingelöst ohne Profil")
	}
}

func TestEinloesungOhneKennung(t *testing.T) {
	ablage, ausstellen := eineAusgestellte(t)

	ausstellen("n1")

	if fehler := ablage.Einloesen(probekontext(t), "n1", einEinloesender(""), einloesezeit); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("ohne Kennung: %v", fehler)
	}
}

func TestEinloesungOhneTafel(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	if _, fehler := db.ExecContext(t.Context(), `DROP TABLE ausgestellte_einladungen`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ablage.Ausgestellt(probekontext(t), "n1", domain.ZweckHandel,
		einloesezeit, einloesezeit); fehler == nil {
		t.Error("ausgestellt ohne Tafel")
	}

	if fehler := ablage.Einloesen(probekontext(t), "n1", einEinloesender("bbert"), einloesezeit); fehler == nil ||
		errors.Is(fehler, domain.ErrEinladungFremd) {
		t.Errorf("ohne Tafel: %v", fehler)
	}
}

func TestEinloesungBrichtAbOhneSpur(t *testing.T) {
	for _, fall := range []struct{ name, ausloeser string }{
		{"der Kontakt klemmt", `CREATE TRIGGER klemmt BEFORE INSERT ON peerkontakte
			BEGIN SELECT RAISE(ABORT, 'klemmt'); END`},
		{"der Vermerk klemmt", `CREATE TRIGGER klemmt BEFORE UPDATE ON ausgestellte_einladungen
			BEGIN SELECT RAISE(ABORT, 'klemmt'); END`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db := frischeDatenbank(t)
			ablage := storage.NeueKontaktablage(db)
			ctx := probekontext(t)

			if fehler := ablage.Ausgestellt(ctx, "n1", domain.ZweckHandel,
				einloesezeit.Add(time.Hour), einloesezeit); fehler != nil {
				t.Fatal(fehler)
			}

			if _, fehler := db.ExecContext(ctx, fall.ausloeser); fehler != nil {
				t.Fatal(fehler)
			}

			if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); fehler == nil {
				t.Fatal("eingelöst trotz Abbruch")
			}

			if _, fehler := ablage.Lesen(ctx, "bbert"); !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("der Kontakt steht: %v", fehler)
			}

			if _, fehler := db.ExecContext(ctx, `DROP TRIGGER klemmt`); fehler != nil {
				t.Fatal(fehler)
			}

			// **Und die Nonce ist noch da** — der Abbruch hat sie nicht
			// verbraucht.
			if fehler := ablage.Einloesen(ctx, "n1", einEinloesender("bbert"), einloesezeit); fehler != nil {
				t.Errorf("danach: %v", fehler)
			}
		})
	}
}

// **Das Räumen alter Einladungen klemmt** — dann wird auch nichts
// ausgestellt.
func TestKlemmendesRaeumenStelltNichtsAus(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	if fehler := ablage.Ausgestellt(ctx, "alt", domain.ZweckHandel, einloesezeit, einloesezeit); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := db.ExecContext(ctx, `CREATE TRIGGER klemmt BEFORE DELETE ON ausgestellte_einladungen
		BEGIN SELECT RAISE(ABORT, 'klemmt'); END`); fehler != nil {
		t.Fatal(fehler)
	}

	spaeter := einloesezeit.Add(30 * 24 * time.Hour)
	if fehler := ablage.Ausgestellt(ctx, "neu", domain.ZweckHandel, spaeter, spaeter); fehler == nil {
		t.Fatal("ausgestellt trotz klemmendem Räumen")
	}

	if fehler := ablage.Einloesen(ctx, "neu", einEinloesender("bbert"), spaeter); !errors.Is(
		fehler, domain.ErrEinladungFremd) {
		t.Errorf("die neue steht da: %v", fehler)
	}
}
