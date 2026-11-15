//go:build darwin

// **Schreiben, nicht überschreiben, löschen** — Werkstück O10 von Phase 5.2
// (#1434).
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein bestehendes Paket wird nicht weggenommen**, und der Fehlschlag lässt
// es unversehrt — das ist die Abnahme „Abbruch hinterlässt kein scheinbar
// fertiges Paket", von der anderen Seite: hier bleibt das **echte** stehen.
func TestEinBestehendesPaketWirdNichtUeberschrieben(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if fehler := os.WriteFile(ziel, []byte("das gehört jemandem"), 0o600); fehler != nil {
		t.Fatalf("vorbelegen: %v", fehler)
	}

	antwort := anwendung.DiagnosepaketNach(t.Context(), ziel)
	if antwort.Fehler == nil {
		t.Fatal("das Paket hat das bestehende weggenommen")
	}

	gelesen, lesefehler := os.ReadFile(ziel) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(gelesen) != "das gehört jemandem" {
		t.Errorf("die bestehende Datei ist verändert: %q", gelesen)
	}

	// **Und keine halbe Nachbardatei bleibt liegen.**
	eintraege, verzeichnisfehler := os.ReadDir(filepath.Dir(ziel))
	if verzeichnisfehler != nil {
		t.Fatalf("Verzeichnis lesen: %v", verzeichnisfehler)
	}

	if len(eintraege) != 1 {
		t.Errorf("%d Einträge, erwartet 1 — es blieb etwas liegen", len(eintraege))
	}
}

// **Das geschriebene Paket lässt sich wieder wegnehmen.**
func TestDasGeschriebenePaketLaesstSichLoeschen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	if geloescht := anwendung.DiagnosepaketLoeschen(t.Context()); geloescht.Fehler != nil {
		t.Fatalf("löschen: %+v", geloescht.Fehler)
	}

	if _, fehler := os.Stat(ziel); !os.IsNotExist(fehler) {
		t.Errorf("das Paket liegt noch da: %v", fehler)
	}

	// **Und ein zweites Löschen geht ins Leere und sagt es.**
	if nochmal := anwendung.DiagnosepaketLoeschen(t.Context()); nochmal.Fehler == nil {
		t.Error("das zweite Löschen meldete Erfolg")
	}
}

// **Ohne geschriebenes Paket gibt es nichts zu löschen** — und keinen Weg,
// mit dem jemand eine beliebige Datei benennt.
func TestOhneGeschriebenesPaketGibtEsNichtsZuLoeschen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if antwort := anwendung.DiagnosepaketLoeschen(t.Context()); antwort.Fehler == nil {
		t.Fatal("das Löschen ohne Paket meldete Erfolg")
	}
}

// **Ein Paket, das jemand von Hand weggenommen hat**, meldet sich beim
// Löschen — statt still Erfolg zu behaupten.
func TestEinVerschwundenesPaketMeldetSichBeimLoeschen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	if fehler := os.Remove(ziel); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	if antwort := anwendung.DiagnosepaketLoeschen(t.Context()); antwort.Fehler == nil {
		t.Error("das Löschen einer verschwundenen Datei meldete Erfolg")
	}
}
