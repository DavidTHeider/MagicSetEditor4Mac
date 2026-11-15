package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Meldung zum Archiv kommt an** —
// #704.
//
// # Der Befund
//
// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)
// entschied, ⌘S auf ein Set aus einem Archiv abzuweisen — **„mit einem Satz,
// der das Archiv nennt und auf ‚Sichern unter' zeigt"**. Der Satz steht in
// `service.keinArchiv` und ist sorgfältig geschrieben.
//
// Angekommen ist er nie. `app.Uebersetze` bildet ein blankes `ErrValidation`
// auf den Rückfallsatz ab: **„Es ist ein unerwarteter Fehler aufgetreten."**
// Der Wächter hielt die halbe Zusage und machte die andere unsichtbar (§10) —
// gefunden beim Rundlauf über ein gepacktes Set
// (#702).
func TestDieArchivmeldungKommtAn(t *testing.T) {
	ort := t.TempDir()

	// **Ein echtes Set als Verzeichnis, gesichert auf eine Datei.** So sieht
	// der Fall aus: die Herkunft ist ein Archiv, und ⌘S nimmt sie als Ziel.
	quelle := filepath.Join(ort, "Probe.mse-set")
	if err := os.MkdirAll(quelle, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: probe-schlicht\n" +
		"set info:\n\ttitle: Probe\ncard:\n\tname: Feuerball\n"

	if err := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	archiv := filepath.Join(ort, "Ziel.mse-set")
	if err := os.WriteFile(archiv, []byte("PK\x03\x04"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	antwort := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, archiv)
	if antwort.Fehler == nil {
		t.Fatal("das Archiv wurde angenommen")
	}

	// **Der Satz nennt das Archiv und den Ausweg.** Ohne beides ist die
	// Absage eine Sackgasse.
	if !strings.Contains(antwort.Fehler.Meldung, "Ziel.mse-set") {
		t.Errorf("die Meldung nennt das Archiv nicht: %q", antwort.Fehler.Meldung)
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Sichern unter") {
		t.Errorf("die Meldung nennt den Ausweg nicht: %q", antwort.Fehler.Meldung)
	}
}
