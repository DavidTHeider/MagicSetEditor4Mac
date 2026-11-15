// **Die Löschnaht für die Flächen ohne Fenster** —
// #1218.
//
// # Warum das ein Lauf gegen echtes Storage ist
//
// Die Naht ist der **Composition Root** dieser Fähigkeit: sie baut den
// Speicher, die Löschung und — wenn es ein Zuhause gibt — die Bildablage.
// Gegen Doubles geprüft bliebe genau das ungeprüft, was sie tut.
//
// **`HOME` wird umgebogen** (§3): sonst räumte dieser Lauf im echten Zuhause
// dieses Rechners auf, und er wäre grün, weil er wirklich funktioniert hat.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der ganze Weg: anlegen, löschen, weg.**
func TestDieSetloeschnahtNimmtDasSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	kennung, err := app.Einfuhrnaht(db).Neu(probekontext(t), "Zum Wegwerfen", "probe", "probe-x")
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	naht := app.SetloeschnahtUeber(db)

	befunde, loeschfehler := naht.SetLoeschen(probekontext(t), kennung, "Zum Wegwerfen")
	if loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde, wo keine sein sollten: %v", befunde)
	}

	// **Und es ist wirklich weg** — über dieselbe Naht gefragt, die es gelöscht
	// hat: ein zweites Löschen findet nichts mehr.
	if _, nochmal := naht.SetLoeschen(probekontext(t), kennung, "Zum Wegwerfen"); !errors.Is(
		nochmal, domain.ErrNotFound) {
		t.Errorf("das zweite Löschen gab %v, erwartet war ErrNotFound", nochmal)
	}
}

// **Ein falscher Name löscht nichts, und die Naht sagt warum.**
func TestDieSetloeschnahtVerlangtDenNamen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	kennung, err := app.Einfuhrnaht(db).Neu(probekontext(t), "Bleibt stehen", "probe", "probe-x")
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	naht := app.SetloeschnahtUeber(db)

	if _, fehler := naht.SetLoeschen(probekontext(t), kennung, "Heißt anders"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("Fehler war %v, erwartet war ErrConflict", fehler)
	}

	// **Und der Beleg, dass es noch da ist**: mit dem richtigen Namen geht es.
	// Ohne diesen zweiten Teil wäre der Fehler oben auch dann erklärbar, wenn
	// die Naht gar nichts fände.
	if _, richtig := naht.SetLoeschen(probekontext(t), kennung, "Bleibt stehen"); richtig != nil {
		t.Errorf("mit dem richtigen Namen: %v", richtig)
	}
}

// **Die Gestalt fürs Fenster gibt dieselbe Antwort in Befunden.**
//
// Sie steht neben `SetLoeschen`, weil `app.SetLoeschen` sie ruft — und ohne
// diesen Lauf wäre sie eine Methode, die niemand betritt.
func TestDieSetloeschnahtGibtDieBefundeAlsDomaenentyp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	kennung, err := app.Einfuhrnaht(db).Neu(probekontext(t), "Über das Fenster", "probe", "probe-x")
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	befunde, loeschfehler := app.SetloeschnahtUeber(db).
		Loeschen(probekontext(t), kennung, "Über das Fenster")
	if loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde, wo keine sein sollten: %v", befunde)
	}
}
