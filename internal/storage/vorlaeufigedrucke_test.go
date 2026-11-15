package storage_test

// **Zwei vorläufige Drucke sind zwei Zeilen** — Zweitblick über DOM-03
// (#1698),
// Befund 2, Migration 0086.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestZweiVorlaeufigeDruckeSindZweiZeilen(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	drucke := make([]int64, 0, 2)

	for _, name := range []string{"Bolt", "Counterspell"} {
		var id int64
		if err := db.QueryRowContext(ctx,
			`INSERT INTO drucke (name, setcode, nummer, sprache) VALUES (?, 'm11', '1', '') RETURNING id`,
			name).Scan(&id); err != nil {
			t.Fatal(err)
		}

		drucke = append(drucke, id)
	}

	ablage := storage.NeuerSammlungsspeicher(db)
	exemplar := func(druck int64) *domain.Exemplar {
		return &domain.Exemplar{DruckID: druck, Menge: 1, Zustand: domain.ZustandNearMint, Sprache: "en"}
	}

	for _, druck := range drucke {
		if _, fehler := ablage.Anlegen(ctx, exemplar(druck)); fehler != nil {
			t.Fatalf("der vorläufige Druck %d: %v", druck, fehler)
		}
	}

	// **Derselbe vorläufige Druck noch einmal bleibt eine Dublette.**
	_, fehler := ablage.Anlegen(ctx, exemplar(drucke[0]))

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Key != domain.SchluesselDoppeltesExemplar {
		t.Errorf("derselbe vorläufige Druck zweimal: %v", fehler)
	}
}

// **Auch eine Zeile ohne Druckzeile ist einmalig** — Nachlauf über ID-04
// und die Migration 0086: `druck_id` steht in der Bedingung, und SQLite
// hält zwei `NULL` für verschieden. Eine solche Zeile kann nur aus einem
// Altbestand stammen (0051 ordnete jedem Alias eine Zeile zu); dort gilt
// die alte Regel weiter (Migration 0087).
func TestZweiZeilenOhneDruckzeileSindEineDublette(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	anweisung := `INSERT INTO sammlung (printing, menge, zustand, sprache, finish, profil_id)
	              VALUES ('scryfall:alt', 1, 'nearMint', 'en', 'nonfoil', 1)`

	if _, err := db.ExecContext(ctx, anweisung); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, anweisung); err == nil {
		t.Error("zwei gleiche Zeilen ohne Druckzeile gingen hinein")
	}
}
