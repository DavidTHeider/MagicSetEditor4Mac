// **Ein Schreibfehler lässt eine vorhandene Bilddatei unverändert** —
// #853.
//
// # Was vorher galt
//
// `os.WriteFile` öffnet mit `O_TRUNC`: eine vorhandene Datei ist ab diesem
// Augenblick leer, und scheitert das Schreiben danach, bleibt sie leer oder
// halb zurück. Ein halbes PNG sieht im Dateibrowser aus wie ein ganzes.
//
// Der Kommentar daneben stimmte für den **Kodierer** und nur für ihn: „Erst
// packen, dann schreiben" fängt einen Kodiererfehler, nicht einen des
// Dateisystems.
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Scheitert das Schreiben, bleibt die alte Datei stehen.**
//
// Erzwungen über ein Ziel, das ein **Verzeichnis** ist: `os.Rename` kann eine
// Datei nicht über ein nichtleeres Verzeichnis legen. Vorher hätte
// `os.WriteFile` schon am Öffnen gescheitert — der Unterschied ist, dass es
// bei einem Ziel, das eine *Datei* ist, erst gekürzt und dann gescheitert
// wäre.
func TestEinSchreibfehlerLaesstDieAlteDateiStehen(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "Feuerball.png")

	if fehler := os.MkdirAll(filepath.Join(ziel, "imweg"), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	ausfuhr := bildausfuhr(t, planDouble{})

	if _, fehler := ausfuhr.Schreibe(t.Context(), 1, ordner, "Feuerball.png",
		service.FormatPNG, 0, domain.BildkonfliktUeberschreiben,
		render.GueteMittel); fehler == nil {
		t.Fatal("kein Fehler")
	}

	// **Und die Nachbardatei bleibt nicht liegen.** Ein Verzeichnis, das sich
	// bei jedem gescheiterten Lauf mit Bruchstücken füllt, ist eine zweite
	// Sorte Schaden.
	pruefeKeineReste(t, ordner, "Feuerball.png")
}

// pruefeKeineReste sucht liegengebliebene Nachbardateien.
func pruefeKeineReste(t *testing.T, ordner, ziel string) {
	t.Helper()

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if eintrag.Name() != ziel && strings.Contains(eintrag.Name(), ziel) {
			t.Errorf("die Nachbardatei %q ist liegengeblieben", eintrag.Name())
		}
	}
}

// **Dieselbe Zusage für SVG**, denn es geht denselben Weg.
func TestEinSchreibfehlerBeiSVGLaesstDieAlteDateiStehen(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "Feuerball.svg")

	if fehler := os.MkdirAll(filepath.Join(ziel, "imweg"), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := vektorausfuhr(t).Schreibe(t.Context(), 1, ordner, "Feuerball.svg",
		service.FormatSVG, 0, domain.BildkonfliktUeberschreiben,
		render.GueteMittel); fehler == nil {
		t.Fatal("kein Fehler")
	}

	pruefeKeineReste(t, ordner, "Feuerball.svg")
}

// **Und ein geglückter Lauf überschreibt weiterhin.**
//
// Das Umbenennen ist der Punkt: die alte Datei steht bis zum letzten
// Augenblick, und danach steht die neue — nie etwas dazwischen.
func TestEinGeglueckterLaufUeberschreibt(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "Feuerball.png")

	if fehler := os.WriteFile(ziel, []byte("alt"), 0o600); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := bildausfuhr(t, planDouble{}).Schreibe(t.Context(), 1, ordner,
		"Feuerball.png", service.FormatPNG, 0,
		domain.BildkonfliktUeberschreiben, render.GueteMittel); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	roh, fehler := os.ReadFile(ziel) //#nosec G304 -- ein Pfad aus t.TempDir
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if string(roh) == "alt" {
		t.Error("die alte Datei steht noch da")
	}

	pruefeKeineReste(t, ordner, "Feuerball.png")
}

// **Der Fall, an dem sich die beiden Wege unterscheiden.**
//
// In einem nur lesbaren Verzeichnis mit einer vorhandenen Zieldatei hätte
// `os.WriteFile` sie **überschrieben**: die vorhandene Datei zu öffnen braucht
// Schreibrecht an der **Datei**, nicht am Verzeichnis. Eine Nachbardatei
// anzulegen braucht eines am Verzeichnis — und deshalb scheitert der Lauf
// jetzt, statt das Ziel anzufassen.
//
// **Gegenprobe**: mit `os.WriteFile` an dieser Stelle ist der Lauf grün und
// die alte Datei weg.
func TestInEinemNurLesbarenVerzeichnisBleibtDasZielUnberuehrt(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "Feuerball.png")

	if fehler := os.WriteFile(ziel, []byte("alt"), 0o600); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Ein Verzeichnis braucht das Ausführungsrecht**, sonst ließe sich
	// nicht einmal hineinsehen; gosec misst Dateirechte.
	if fehler := os.Chmod(ordner, 0o500); fehler != nil { //#nosec G302 -- ein Verzeichnis, kein Dateiinhalt
		t.Fatalf("Rechte setzen: %v", fehler)
	}

	// **Zurücksetzen, sonst räumt `t.TempDir` nicht auf.**
	t.Cleanup(func() {
		if fehler := os.Chmod(ordner, 0o750); fehler != nil { //#nosec G302 -- dito
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	if _, fehler := bildausfuhr(t, planDouble{}).Schreibe(t.Context(), 1, ordner,
		"Feuerball.png", service.FormatPNG, 0,
		domain.BildkonfliktUeberschreiben, render.GueteMittel); fehler == nil {
		t.Fatal("kein Fehler")
	}

	roh, fehler := os.ReadFile(ziel) //#nosec G304 -- ein Pfad aus t.TempDir
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if string(roh) != "alt" {
		t.Errorf("die alte Datei ist %q — sie sollte unberührt bleiben", roh)
	}
}

// **Die Rechte einer vorhandenen Datei überleben das Ersetzen.**
//
// # Der Befund
//
// **Befund 6 des Zweitblicks über Phase 4.13** (2026-09-07): `os.WriteFile`
// schrieb in die bestehende Datei und ließ ihren Modus; die Nachbardatei
// entsteht mit 0600 und trug ihn nicht. Wer sein `.mse-set` auf 0644 gestellt
// hat, fand es nach dem Sichern auf 0600.
//
// **Erst nur aufgeschrieben, dann behoben.** Der Einwand traf: der Normalfall
// ist ohne konstruierten Fehlerweg prüfbar, und ein Verhalten, das
// deterministisch da ist, wird abgestellt statt genannt.
func TestDerModusEinerVorhandenenDateiBleibt(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "Feuerball.png")

	//nolint:gosec // 0644 **ist** der Fall: geprüft wird, dass er überlebt
	if fehler := os.WriteFile(ziel, []byte("alt"), 0o644); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := service.SchreibeAtomar(ziel, []byte("neu")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	angabe, fehler := os.Stat(ziel)
	if fehler != nil {
		t.Fatalf("nachsehen: %v", fehler)
	}

	if modus := angabe.Mode().Perm(); modus != 0o644 {
		t.Errorf("der Modus ist %04o, erwartet war 0644", modus)
	}

	// **Und der Inhalt ist der neue** — sonst prüfte der Lauf einen
	// Schreibvorgang, der gar nicht stattgefunden hat.
	roh, lesefehler := os.ReadFile(ziel) //nolint:gosec // Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(roh) != "neu" {
		t.Errorf("der Inhalt ist %q", roh)
	}
}

// **Und wo nichts lag, bleibt es bei 0600.**
func TestEineNeueDateiBekommt0600(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "neu.png")

	if fehler := service.SchreibeAtomar(ziel, []byte("x")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	angabe, fehler := os.Stat(ziel)
	if fehler != nil {
		t.Fatalf("nachsehen: %v", fehler)
	}

	if modus := angabe.Mode().Perm(); modus != 0o600 {
		t.Errorf("der Modus ist %04o, erwartet war 0600", modus)
	}
}
