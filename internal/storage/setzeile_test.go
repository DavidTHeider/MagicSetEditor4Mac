package storage_test

// **Die Bestandszeile eines Sets** — Befund **F02**
// (#1746).
//
// Sie sagt, welches Spiel und welcher Stil gemeint sind, wenn **keine Datei**
// es sagt: ein Set, das in dieser Anwendung entstanden ist, hat keine Quelle,
// und ohne `game:` ist kein Paket zu schreiben.
//
// **Die Läufe stehen neben denen zu `Setstil`**, weil sie dieselbe Frage
// stellen — und getrennt von ihnen, weil `setleser_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieBestandszeileEinesSetsKommtZurueck(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser, ctx := storage.NeuerSetLeser(db), probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO sets (id, profil_id, name, spiel, stil, herkunft, geoeffnet)
		 VALUES (9, 1, 'Sternengarten', 'magic', 'm15-mainframe',
		         '/tmp/stern.mse-set', datetime('now'))`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	zeile, fehler := leser.Setzeile(ctx, 9)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	erwartet := domain.SetHead{
		ID: 9, Name: "Sternengarten", Game: "magic",
		Style: "m15-mainframe", Origin: "/tmp/stern.mse-set",
	}

	if zeile != erwartet {
		t.Errorf("die Zeile ist %+v, erwartet %+v", zeile, erwartet)
	}
}

// **Die Kartenzahl bleibt null, und das ist kein Versehen.**
//
// Sie ist genau das, was diese Abfrage einspart — der Zähler über `cards` ist
// der Grund, warum `Setkoepfe` hier nicht taugt. Der Lauf hält es fest, damit
// niemand später eine Zahl erwartet, die nie erhoben wird.
func TestDieBestandszeileZaehltKeineKarten(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser, ctx := storage.NeuerSetLeser(db), probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (9, 1, 'Probe', 'magic', 'm15', datetime('now'));
		 INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 9, 0)`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	zeile, fehler := leser.Setzeile(ctx, 9)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if zeile.Cards != 0 {
		t.Errorf("die Zeile zählt %d Karten — sie soll keine zählen", zeile.Cards)
	}
}

// **Ein Set, das es nicht gibt, wird gemeldet** — und nicht als leere Zeile.
func TestEinFehlendesSetHatKeineBestandszeile(t *testing.T) {
	t.Parallel()

	_, fehler := storage.NeuerSetLeser(frischeDatenbank(t)).Setzeile(
		probekontext(t), 99)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Und ein Set eines fremden Profils ist hier keines.**
func TestEinFremdesSetHatHierKeineBestandszeile(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (9, 1, 'Probe', 'magic', 'm15', datetime('now'))`); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	fremd := storage.MitProfil(t.Context(),
		zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier"))

	if _, fehler := leser.Setzeile(fremd, 9); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Ohne Profil geht gar nichts, und eine tote Datenbank wird gemeldet.**
func TestDieBestandszeileMeldetIhreFehlerwege(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	leser := storage.NeuerSetLeser(db)

	if _, fehler := leser.Setzeile(t.Context(), 9); fehler == nil {
		t.Fatal("ohne Profil kam kein Fehler")
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := leser.Setzeile(probekontext(t), 9); fehler == nil {
		t.Fatal("die geschlossene Datenbank kam nicht an")
	}
}
