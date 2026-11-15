package storage_test

// **Die Angaben eines Profils im Bestand** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 10 bis 13.
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher, und nebenläufig meldet SQLite „database
// table is locked".

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die vier Angaben gehen hinein und kommen zurück.**
//
// **Über das Lesen und nicht über die Antwort des Schreibens** — die gibt
// zurück, was hineinging, und prüfte damit sich selbst.
func TestDieProfilangabenGehenInDenBestandUndZurueck(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := t.Context()

	profile, fehler := ablage.Profile(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(profile) != 1 {
		t.Fatalf("es sind %d Profile", len(profile))
	}

	geaendert := profile[0]
	geaendert.Subslogan = "Nur für Turniere"
	geaendert.Kurzebeschreibung = "Was ich mitnehme."
	geaendert.Langebeschreibung = "Und warum."
	geaendert.Badge = "profil-1.png"

	if _, err := ablage.ProfilAendern(ctx, geaendert); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	gelesen, lesefehler := ablage.Profile(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	eines := gelesen[0]
	if eines.Subslogan != "Nur für Turniere" || eines.Badge != "profil-1.png" ||
		eines.Kurzebeschreibung != "Was ich mitnehme." ||
		eines.Langebeschreibung != "Und warum." {
		t.Errorf("zurück kam %+v", eines)
	}

	// **Und die Fassung ist gewandert** — sonst schriebe der nächste Ruf
	// gegen einen Stand, den es nicht mehr gibt.
	if eines.Fassung != geaendert.Fassung+1 {
		t.Errorf("die Fassung ist %d statt %d", eines.Fassung, geaendert.Fassung+1)
	}
}

// **Ein Badge, das ein Pfad ist, kommt nicht in den Bestand.**
//
// # Warum das hier steht und nicht nur in der Domäne
//
// **Weil zwei grüne Seiten kein Beleg über die Naht sind.** Dass
// `domain.Profil.Validate` den Wert abweist, hält der Lauf dort; dass die
// Ablage **fragt**, bevor sie schreibt, hält nur dieser.
func TestEinPfadAlsBadgeKommtNichtInDenBestand(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := t.Context()

	profile, fehler := ablage.Profile(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	geaendert := profile[0]
	geaendert.Badge = "../heimlich.png"

	if _, err := ablage.ProfilAendern(ctx, geaendert); err == nil {
		t.Fatal("der Pfad kam durch")
	}

	// **Und es blieb, wie es war** — eine Ablehnung, die trotzdem geschrieben
	// hätte, wäre ein halber Stand (§8).
	gelesen, lesefehler := ablage.Profile(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen[0].Badge != "" {
		t.Errorf("der Bestand führt %q", gelesen[0].Badge)
	}
}
