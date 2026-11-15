// Der Durchlauf, den #206 verlangt: von der Setdatei bis zum Bild.
//
// **Er geht durch die echte Anwendung**, nicht an ihr vorbei — `Starten`,
// `SetOeffnen`, `Karten`, `Vorschau`. Genau dazwischen lag der Befund: jede
// Schicht war geprüft, und die Naht zwischen ihnen führte zum Behelfsplan.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Eine Karte aus einem echten Stilpaket, durch die ganze Anwendung.**
//
// # Was er hält
//
// Dass `Vorschau` den **Stilplan** nimmt, sobald das Stilpaket zu finden ist.
// Vor #206 kam hier der Behelfsplan, und kein Test wurde davon rot.
func TestDieVorschauZeigtDenStilAusEinemEchtenPaket(t *testing.T) {
	pakete := t.TempDir()

	// **Zwei Ebenen tief, so wie ein bezogenes Repositorium aussieht.**
	// `M15-Magic-Pack` legt seine 190 Stile unter `M15-Magic-Pack/data/` ab;
	// wer die Saat hier flach hinlegte, prüfte den Fall nicht, der #206 zur
	// Hälfte ausmachte — `Bildquelle` und `Paketleser` fanden **nichts**.
	tief := filepath.Join(pakete, "Probe-Pack", "data")
	if err := os.MkdirAll(tief, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	kopiere(t, filepath.Join("..", "..", "saat"), tief)

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	// **Das Paketverzeichnis wird untergeschoben.** Sonst liefe der Test gegen
	// das Verzeichnis dieses Rechners — und wäre damit auf einer zweiten
	// Maschine grün oder rot, je nachdem, wer dort was bezogen hat (§4).
	anwendung.paketort = func() (string, error) { return pakete, nil }

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	setdatei := gelegtesSet(t, "probe-schlicht")

	if antwort := anwendung.SetOeffnen(t.Context(), setdatei); antwort.Fehler != nil {
		t.Fatalf("öffnen: %v", antwort.Fehler)
	}

	karten := anwendung.Karten(t.Context(), "1")
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten)
	}

	vorschau := anwendung.Vorschau(t.Context(), "1", 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %v", vorschau.Fehler)
	}

	// **Über die Vorbehalte und das Bild.** Der Satz hält die Paketnaht; die
	// linke obere Fläche unterscheidet den weißen Stilplan vom beigen Behelf.
	// Ohne die zweite Gegenprobe bliebe dieser Lauf grün, wenn `Kartenplan` das
	// Paket zwar fände, danach aber trotzdem `Behelfsplan` nähme.
	for _, befund := range vorschau.Vorschau.Befunde {
		if strings.Contains(befund, "Stilverzeichnis") ||
			strings.Contains(befund, "liegt in keinem Paketverzeichnis") {
			t.Errorf("Befund %q — die Anwendung findet das Stilpaket nicht", befund)
		}
	}

	if vorschau.Vorschau.Breite == 0 {
		t.Error("die Vorschau ist leer")
	}

	bild := bildAus(t, anwendung, vorschau.Vorschau.Bild)

	rot, gruen, blau, alpha := bild.At(0, 0).RGBA()
	if rot != 0xffff || gruen != 0xffff || blau != 0xffff || alpha != 0xffff {
		t.Errorf("linke obere Farbe = %04x/%04x/%04x/%04x — das ist nicht der Stilplan",
			rot, gruen, blau, alpha)
	}
}

// **Und ohne Stilpaket kommt der Behelfsplan — mit dem Grund.**
//
// Die andere Hälfte derselben Zusage (#206): ein fehlendes Paket macht die
// Vorschau nicht leer und nicht stillschweigend behilflich.
func TestOhneStilpaketNenntDieVorschauDenGrund(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	leer := t.TempDir()
	anwendung.paketort = func() (string, error) { return leer, nil }

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	if antwort := anwendung.SetOeffnen(t.Context(),
		gelegtesSet(t, "gibtesnicht")); antwort.Fehler != nil {
		t.Fatalf("öffnen: %v", antwort.Fehler)
	}

	vorschau := anwendung.Vorschau(t.Context(), "1", 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %v", vorschau.Fehler)
	}

	if !hatBefundMit(vorschau.Vorschau.Befunde, "Paketverzeichnis") {
		t.Errorf("%v — kein Wort davon, dass das Stilpaket fehlt",
			vorschau.Vorschau.Befunde)
	}
}

// hatBefundMit sagt, ob ein Befund den Teiltext führt.
func hatBefundMit(befunde []string, teil string) bool {
	for _, befund := range befunde {
		if strings.Contains(befund, teil) {
			return true
		}
	}

	return false
}

// gelegtesSet schreibt ein `.mse-set` mit einem gewählten Stilnamen.
func gelegtesSet(t *testing.T, stil string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "durchlauf.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: " + stil +
		"\nset info:\n\ttitle: Durchlauf\ncard:\n\tname: Feuerball\n" +
		"\ttext: Füge drei Schaden zu.\n\tfarbe: rot\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// kopiere legt ein Verzeichnis samt Inhalt an einem zweiten Ort ab.
func kopiere(t *testing.T, von, nach string) {
	t.Helper()

	eintraege, err := os.ReadDir(von)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		quelle, ziel := filepath.Join(von, eintrag.Name()), filepath.Join(nach, eintrag.Name())

		if eintrag.IsDir() {
			if err := os.MkdirAll(ziel, 0o750); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			kopiere(t, quelle, ziel)

			continue
		}

		inhalt, err := os.ReadFile(quelle) // #nosec G304 -- Testdaten aus dem Repository
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		if err := os.WriteFile(ziel, inhalt, 0o600); err != nil { // #nosec G703 -- Ziel ist t.TempDir()
			t.Fatalf("schreiben: %v", err)
		}
	}
}
