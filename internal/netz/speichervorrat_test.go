package netz_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Was schon da ist, ohne zu fragen** — P10
// (#2378).
//
// Eine Handelszeile nennt Name, Set und Bild ihres Drucks, und das Öffnen
// eines Tauschs darf dafür nicht ins Netz gehen (§16, Regel 5). `Vorrat`
// gibt deshalb, was liegt — **auch nach der Frist**: ein Name veraltet
// nicht, und die Frist ist eine Zusage an die Gegenstelle, keine über den
// Inhalt.
func TestVorratGibtWasLiegtOhneZuFragen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.October, 19, 8, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), 24*time.Hour, func() time.Time { return jetzt })

	if _, _, gabEs := speicher.Vorrat("cards/id:x"); gabEs {
		t.Fatal("ein leerer Speicher hat einen Vorrat")
	}

	if _, _, fehler := speicher.Hole(t.Context(), "cards/id:x", func() ([]byte, error) {
		return []byte(`{"id":"x"}`), nil
	}); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	jetzt = jetzt.Add(30 * 24 * time.Hour)

	inhalt, alter, gabEs := speicher.Vorrat("cards/id:x")
	if !gabEs || string(inhalt) != `{"id":"x"}` {
		t.Fatalf("Vorrat nach dreißig Tagen: %q, %v", inhalt, gabEs)
	}

	if alter != 30*24*time.Hour {
		t.Errorf("Alter %v, erwartet waren dreißig Tage", alter)
	}

	// **Gegenprobe, dass Hole an derselben Stelle wirklich fragen würde**:
	// sonst bewiese das Ergebnis darüber nur einen Speicher ohne Frist.
	gefragt := false
	if _, _, fehler := speicher.Hole(t.Context(), "cards/id:x", func() ([]byte, error) {
		gefragt = true

		return nil, errors.New("kein Netz")
	}); fehler != nil || !gefragt {
		t.Errorf("Hole nach der Frist: gefragt %v, Fehler %v", gefragt, fehler)
	}
}

// **Abgelegt ist Vorrat** — ohne dass jemand geholt hat.
func TestAbgelegtesStehtImVorrat(t *testing.T) {
	t.Parallel()

	speicher := netz.NeuerSpeicher(t.TempDir(), 24*time.Hour, time.Now)

	if fehler := speicher.Ablegen("druck/vorrat:x", []byte(`{"id":"x"}`)); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if inhalt, _, gabEs := speicher.Vorrat("druck/vorrat:x"); !gabEs || string(inhalt) != `{"id":"x"}` {
		t.Errorf("Vorrat: %q, %v", inhalt, gabEs)
	}
}
