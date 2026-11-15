package service_test

// **Wie ein Bildvorrat entsteht und was er auf der Platte erwartet** —
// Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Die „Bereinigung nach Abbruch/Neustart" aus dem Satz des Issues — und
// die drei Arten, auf die sie scheitern kann.** Der Vorrat liegt im
// Zwischenspeicher des Benutzers, und dort bestimmt nicht diese Anwendung
// allein, was geht: das Verzeichnis kann unräumbar sein, unanlegbar, oder
// mitten im Betrieb verschwinden.
//
// **Jeder dieser Fälle sagt es**, statt still weiterzulaufen — ein Vorrat,
// der den alten Stand stehen ließe, behielte Dateien, deren Besitzer niemand
// mehr kennt.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Neustart räumt auf.**
//
// Das Issue verlangt „Bereinigung nach Abbruch/Neustart", und ein Abbruch
// hinterlässt genau das: Dateien ohne Eintrag. Der Vorrat weiß beim Start
// nichts von ihnen — deshalb wird das Verzeichnis geleert und nicht gelesen.
func TestEinNeustartRaeumtAuf(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	liegengeblieben := filepath.Join(ort, "abcdef0123456789abcdef0123456789.png")
	if fehler := os.WriteFile(liegengeblieben, einVorratsbild(t, 8, 8), 0o600); fehler != nil {
		t.Fatalf("die Datei ließ sich nicht anlegen: %v", fehler)
	}

	if _, fehler := service.NeuerBildvorrat(ort, service.UeblicheBildgrenzen(),
		time.Now); fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	if _, fehler := os.Stat(liegengeblieben); !os.IsNotExist(fehler) {
		t.Fatalf("die liegengebliebene Datei ist noch da: %v", fehler)
	}
}

// **Ohne Verzeichnis gibt es keinen Vorrat.**
func TestOhneVerzeichnisKeinVorrat(t *testing.T) {
	t.Parallel()

	if _, fehler := service.NeuerBildvorrat("", service.UeblicheBildgrenzen(),
		time.Now); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der leere Ort kam durch: %v", fehler)
	}
}

// **Das Verzeichnis gibt es noch nicht** — deshalb kommt `RemoveAll` durch
// und `MkdirAll` scheitert. Der Lauf darunter ist der umgekehrte Fall.
func TestOhneAnlegenKeinVorrat(t *testing.T) {
	t.Parallel()

	eltern := t.TempDir()

	// **0500 und nicht 0400**: ein Verzeichnis ohne Ausführrecht lässt sich
	// nicht einmal betreten, und dann scheitert schon `RemoveAll`.
	if fehler := os.Chmod(eltern, 0o500); fehler != nil { //nolint:gosec // siehe oben
		t.Fatalf("die Rechte ließen sich nicht setzen: %v", fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(eltern, 0o700) }) //nolint:errcheck,gosec // damit t.TempDir aufräumen kann

	if _, fehler := service.NeuerBildvorrat(filepath.Join(eltern, "neu"),
		service.UeblicheBildgrenzen(), time.Now); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der unanlegbare Ort kam durch: %v", fehler)
	}
}

// **Was der Vorrat nicht mehr auf der Platte findet, ist ein Fehler und
// keine leere Datei.**
//
// Der Fall entsteht, wenn jemand von außen aufräumt — ein Reinigungsdienst
// im Zwischenspeicherverzeichnis tut genau das.
func TestEinBildOhneDateiIstEinFehler(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	eintrag, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 8, 8))
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	if fehler := os.RemoveAll(vorrat.Ort()); fehler != nil {
		t.Fatalf("das Verzeichnis ließ sich nicht räumen: %v", fehler)
	}

	if _, _, fehler := vorrat.Lesen(eintrag.Kennung, 7); fehler == nil {
		t.Fatal("das verschwundene Bild kam als Erfolg zurück")
	}
}

// **Sonst belegte es die Quote für nichts**: der Eintrag zählte mit, die
// Datei gäbe es nicht, und niemand könnte sie verwerfen.
func TestEinNichtGeschriebenesBildBelegtNichts(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	if fehler := os.RemoveAll(vorrat.Ort()); fehler != nil {
		t.Fatalf("das Verzeichnis ließ sich nicht räumen: %v", fehler)
	}

	if _, fehler := vorrat.Legen(einBildbesitz(),
		einVorratsbild(t, 8, 8)); fehler == nil {
		t.Fatal("das Ablegen ohne Verzeichnis gelang")
	}

	if weg := vorrat.Aufraeumen(); weg != 0 {
		t.Fatalf("es blieben %d Einträge zurück", weg)
	}
}

// **Weil der Vorrat im Zwischenspeicher des Benutzers liegt**, und dort
// bestimmt nicht diese Anwendung allein die Rechte. Lässt sich nicht räumen,
// wäre die stille Alternative, den alten Stand stehen zu lassen — also
// Dateien zu behalten, deren Besitzer niemand mehr kennt.
func TestOhneRaeumenKeinVorrat(t *testing.T) {
	t.Parallel()

	eltern := t.TempDir()
	ort := filepath.Join(eltern, "vorrat")

	if fehler := os.MkdirAll(filepath.Join(ort, "darin"), 0o700); fehler != nil {
		t.Fatalf("das Verzeichnis ließ sich nicht anlegen: %v", fehler)
	}

	// **Ohne Schreibrecht am Elternteil geht kein Eintrag weg.**
	if fehler := os.Chmod(ort, 0o500); fehler != nil { //nolint:gosec // ein Verzeichnis braucht das Ausführrecht
		t.Fatalf("die Rechte ließen sich nicht setzen: %v", fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck,gosec // damit t.TempDir aufräumen kann

	if _, fehler := service.NeuerBildvorrat(ort, service.UeblicheBildgrenzen(),
		time.Now); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der unräumbare Ort kam durch: %v", fehler)
	}
}
