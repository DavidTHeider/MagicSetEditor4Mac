package service_test

// **Die zeitweiligen Dateien des Teilens** — NET-04.AK-04
// (#1708):
// „Temporäre Exportdateien werden sicher aufbewahrt und anschließend
// bereinigt."
//
// **Sicher heißt**: je Teilung ein eigener Ordner, nur für den Eigentümer
// (`0700`, Dateien `0600`), unter einer Kennung, die kein Weg sein kann.
// **Bereinigt heißt**: auf Zuruf verworfen, und was ein Absturz liegen ließ,
// beim nächsten Aufräumen.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineTeilungLiegtSicherUndGehtWieder(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join(t.TempDir(), "teilen")
	stelle := service.NeueTeilstelle(wurzel, time.Now)

	kennung, ordner, fehler := stelle.Ordner()
	if fehler != nil {
		t.Fatalf("Ordner: %v", fehler)
	}

	if domain.PruefeTeilkennung(kennung) != nil || filepath.Dir(ordner) != wurzel {
		t.Fatalf("Kennung %q, Ordner %q", kennung, ordner)
	}

	if rechte := rechteVon(t, ordner); rechte != 0o700 {
		t.Errorf("der Ordner hat die Rechte %v", rechte)
	}

	pfad, groesse, fehler := stelle.Ablegen(kennung, "Deck.txt", []byte("Deck\n4 Bolt\n"))
	if fehler != nil || groesse != 12 || filepath.Dir(pfad) != ordner {
		t.Fatalf("Ablegen: %q, %d, %v", pfad, groesse, fehler)
	}

	if rechte := rechteVon(t, pfad); rechte != 0o600 {
		t.Errorf("die Datei hat die Rechte %v", rechte)
	}

	if fehler := stelle.Verwerfen(kennung); fehler != nil {
		t.Fatalf("Verwerfen: %v", fehler)
	}

	if _, fehler := os.Stat(ordner); !os.IsNotExist(fehler) {
		t.Errorf("der Ordner steht noch: %v", fehler)
	}
}

// **Eine Kennung, die ein Weg wäre, erreicht die Platte nicht.**
func TestEineFremdeKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	stelle := service.NeueTeilstelle(t.TempDir(), time.Now)

	if fehler := stelle.Verwerfen("../../etwas"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Verwerfen: %v", fehler)
	}

	if _, _, fehler := stelle.Ablegen("../x", "a.txt", nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Ablegen: %v", fehler)
	}

	// **Ein Dateiname mit Weg wird kein Weg** — er wird zum Namen.
	kennung, ordner := einOrdner(t, stelle)

	pfad, _, fehler := stelle.Ablegen(kennung, "../../ausbruch.txt", []byte("x"))
	if fehler != nil || filepath.Dir(pfad) != ordner {
		t.Errorf("der Name wurde zum Weg: %q, %v", pfad, fehler)
	}
}

// **Was liegen blieb, räumt das Aufräumen weg** — und nur das.
func TestAufraeumenNimmtNurAlteTeilungen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	jetzt := time.Now()
	stelle := service.NeueTeilstelle(wurzel, func() time.Time { return jetzt })

	alt, altordner := einOrdner(t, stelle)
	_, neuordner := einOrdner(t, stelle)

	vorher := jetzt.Add(-2 * 24 * time.Hour)
	if fehler := os.Chtimes(altordner, vorher, vorher); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ein Fremdkörper im Teilordner bleibt liegen** — geräumt wird nur, was
	// eine Kennung trägt.
	fremd := filepath.Join(wurzel, "nicht-von-uns")
	if fehler := os.WriteFile(fremd, []byte("x"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Chtimes(fremd, vorher, vorher); fehler != nil {
		t.Fatal(fehler)
	}

	geraeumt, fehler := stelle.Aufraeumen(24 * time.Hour)
	if fehler != nil || geraeumt != 1 {
		t.Fatalf("geräumt %d: %v", geraeumt, fehler)
	}

	for pfad, soll := range map[string]bool{altordner: false, neuordner: true, fremd: true} {
		if _, fehler := os.Stat(pfad); (fehler == nil) != soll {
			t.Errorf("%s steht %v, sollte %v", pfad, fehler == nil, soll)
		}
	}

	_ = alt

	// **Ohne Teilordner gibt es nichts zu räumen** — und keinen Fehler.
	if geraeumt, fehler := service.NeueTeilstelle(filepath.Join(wurzel, "fehlt"), time.Now).Aufraeumen(time.Hour); fehler != nil || geraeumt != 0 {
		t.Errorf("ohne Ordner: %d, %v", geraeumt, fehler)
	}
}

// **Wo kein Ordner entstehen kann, entsteht keine Teilung.**
func TestOhneBeschreibbareWurzelKeineTeilung(t *testing.T) {
	t.Parallel()

	datei := filepath.Join(t.TempDir(), "eine-datei")
	if fehler := os.WriteFile(datei, []byte("x"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if _, _, fehler := service.NeueTeilstelle(filepath.Join(datei, "teilen"), time.Now).Ordner(); fehler == nil {
		t.Error("unter einer Datei entstand ein Ordner")
	}

	// **Und eine Kennung ohne Ordner nimmt keine Datei an.**
	stelle := service.NeueTeilstelle(t.TempDir(), time.Now)
	if _, _, fehler := stelle.Ablegen("0123456789abcdef0123456789abcdef", "a.txt", []byte("x")); fehler == nil {
		t.Error("eine Datei landete in einem Ordner, den es nicht gibt")
	}

	if _, fehler := service.NeueTeilstelle(datei, time.Now).Aufraeumen(time.Hour); fehler == nil {
		t.Error("eine Datei als Teilordner ließ sich aufräumen")
	}
}

// **Was sich nicht wegräumen lässt, meldet sich** — und das Aufräumen zählt
// es nicht als geräumt.
func TestEinSchreibgeschuetzterTeilordnerBleibt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	jetzt := time.Now()
	stelle := service.NeueTeilstelle(wurzel, func() time.Time { return jetzt.Add(48 * time.Hour) })

	kennung, _ := einOrdner(t, stelle)
	if _, _, fehler := stelle.Ablegen(kennung, "a.txt", []byte("x")); fehler != nil {
		t.Fatal(fehler)
	}

	//nolint:gosec // ein Ordner ohne Schreibrecht ist der Fall, den der Lauf prüft
	if fehler := os.Chmod(filepath.Join(wurzel, kennung), 0o500); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(wurzel, kennung), 0o700) }) //nolint:errcheck,gosec // nur ein Lauf

	if fehler := stelle.Verwerfen(kennung); fehler == nil {
		t.Error("ein schreibgeschützter Ordner ließ sich verwerfen")
	}

	if geraeumt, fehler := stelle.Aufraeumen(time.Hour); fehler != nil || geraeumt != 0 {
		t.Errorf("geräumt %d: %v", geraeumt, fehler)
	}
}

// einOrdner legt einen Teilordner an — ein Fehler bricht den Lauf ab.
func einOrdner(t *testing.T, stelle *service.Teilstelle) (kennung, ordner string) {
	t.Helper()

	kennung, ordner, fehler := stelle.Ordner()
	if fehler != nil {
		t.Fatalf("Ordner: %v", fehler)
	}

	return kennung, ordner
}

// rechteVon liest die Rechte eines Pfads.
func rechteVon(t *testing.T, pfad string) os.FileMode {
	t.Helper()

	info, fehler := os.Stat(pfad)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	return info.Mode().Perm()
}
