// **Die Absprache über das Hostbudget** — Befund Z7-03 von Phase 3.2
// (#1163).
//
// Wie sie arbeitet, steht in `internal/netz`. Hier steht, dass dieser
// Composition Root sie einrichtet — und was geschieht, wenn er es nicht kann.
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne Benutzerverzeichnis gibt es keine Absprache — und keinen Fehler.**
//
// Der Fall entsteht auf einem Rechner ohne `HOME`. Dann drosselt jeder Prozess
// für sich, wie vor Phase 3.2; eine Anwendung, die deswegen keine Karte mehr
// sucht, wäre schlechter als eine, die ihr Budget nur je Prozess einhält (§8).
func TestOhneBenutzerverzeichnisGibtEsKeineAbsprache(t *testing.T) {
	t.Setenv("HOME", "")

	if ort := app.ScryfallAbspracheort(); ort != "" {
		t.Errorf("ohne Zuhause entstand der Ort %q", ort)
	}
}

// **Und mit einem gibt es sie.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn der Ort
// **immer** leer bliebe.
func TestMitBenutzerverzeichnisGibtEsEineAbsprache(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if ort := app.ScryfallAbspracheort(); ort == "" {
		t.Error("mit Zuhause blieb der Ort leer")
	}
}

// **Lässt sich der Ordner nicht anlegen, gibt es keine Absprache.**
//
// Dann drosselt jeder Prozess für sich — dieselbe Antwort wie ohne
// Benutzerverzeichnis, und aus demselben Grund (§8).
func TestOhneAnlegbarenOrdnerGibtEsKeineAbsprache(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	// **Eine Datei am Ort des Ordners**: `os.MkdirAll` scheitert.
	anwendungsordner := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac")
	if err := os.MkdirAll(anwendungsordner, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(anwendungsordner, app.Drosselordner),
		[]byte("kein Ordner"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if ort := app.ScryfallAbspracheort(); ort != "" {
		t.Errorf("es entstand der Ort %q", ort)
	}
}
