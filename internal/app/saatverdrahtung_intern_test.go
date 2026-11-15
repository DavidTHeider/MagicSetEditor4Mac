// Der Wächter über die Naht, die #58 gefehlt hat.
//
// **Er prüft nicht, dass die Saat gebündelt *werden kann* — sondern dass die
// laufende Anwendung sie findet.** Das ist der Unterschied, an dem in Phase 6
// dreimal ein fertiges Werkstück unsichtbar blieb: beide Seiten geprüft,
// dazwischen nichts. `scripts/saat.test.sh` hält die eine Seite (die Saat kommt
// ins Bündel), `werkzeuge.TestWailsJsonLegtDieSaatInsBuendel` die Erklärung —
// hier steht die Naht.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der Composition Root macht die Saat zur Paketwurzel.**
//
// # Die Gegenprobe, die dahintersteht
//
// Wer in `app.go` das `a.saatort()` aus dem Aufruf von `msepaket.Paketwurzeln`
// entfernt, macht diesen Lauf rot: die Vorschau meldet dann, das Stilpaket
// liege in keinem Paketverzeichnis, und zeigt den Behelfsplan. Genau so war der
// Stand vor #58 — mit dem Unterschied, dass es damals niemand sah.
func TestDerStartFindetDieMitgelieferteSaat(t *testing.T) {
	anwendung := mitSaat(t)

	antwort := anwendung.SetOeffnen(t.Context(), saatset(t))
	if antwort.Fehler != nil {
		t.Fatalf("öffnen: %+v", antwort.Fehler)
	}

	karten := anwendung.Karten(t.Context(), antwort.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if len(karten.Karten) == 0 {
		t.Fatal("keine Karte im geöffneten Set")
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	// **Über die Befunde und nicht über das Bild.** Ein Bild entsteht auch aus
	// dem Behelfsplan; was den Unterschied macht, ist die Meldung, dass der Stil
	// nicht zu finden war. Ein Befund ist hier kein Nebenton, sondern der
	// ganze Befund: mit gefundener Saat gibt es keinen.
	for _, meldung := range vorschau.Vorschau.Befunde {
		if strings.Contains(meldung, "Paketverzeichnis") || strings.Contains(meldung, "Stilpaket") {
			t.Errorf("die Anwendung findet die Saat nicht: %s", meldung)
		}
	}
}

// **Ohne Saat im Bündel sagt die Anwendung, woran es liegt.**
//
// Die andere Hälfte derselben Naht: der Wächter oben wäre auch dann grün, wenn
// er gar nichts prüfte. Hier steht, dass die Meldung kommt, wenn sie kommen
// soll — und damit, dass die obige Abwesenheit etwas heißt.
func TestOhneSaatBleibtEsBeimBehelfsplan(t *testing.T) {
	anwendung := ohneSaat(t)

	antwort := anwendung.SetOeffnen(t.Context(), saatset(t))
	if antwort.Fehler != nil {
		t.Fatalf("öffnen: %+v", antwort.Fehler)
	}

	karten := anwendung.Karten(t.Context(), antwort.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	if !nennt(vorschau.Vorschau.Befunde, "Paketverzeichnis") {
		t.Errorf("kein Hinweis auf das fehlende Paketverzeichnis: %v", vorschau.Vorschau.Befunde)
	}
}

// nennt sagt, ob eine der Meldungen ein Wort führt.
func nennt(meldungen []string, wort string) bool {
	for _, meldung := range meldungen {
		if strings.Contains(meldung, wort) {
			return true
		}
	}

	return false
}

// mitSaat baut eine Anwendung, die die Saat des Repositoriums als Bündelsaat
// führt — und **kein** bezogenes Paket.
//
// Der leere Paketort ist Absicht: auf einem Rechner, auf dem jemand `M15` bezogen
// hat, sagte ein Lauf gegen das echte Verzeichnis nichts über die Saat.
func mitSaat(t *testing.T) *App {
	t.Helper()

	return gestartetMit(t, func() string {
		wurzel, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			t.Fatalf("Wurzel: %v", err)
		}

		return filepath.Join(wurzel, "saat")
	})
}

// ohneSaat baut eine Anwendung ohne jede Vorlage — der frische Rechner.
func ohneSaat(t *testing.T) *App { t.Helper(); return gestartetMit(t, func() string { return "" }) }

// gestartetMit startet eine Anwendung mit gesetzter Saatnaht und leerem Paketort.
// schriftwurzelDesLaufs ist die Schriftwurzel, die ein Fremdpaketlauf
// mitbekommt — leer, wenn er ohne läuft. **Die eine Stelle, die sie liest**
// ([gestartetMit] sagt, warum es eine bleibt); wer den Pfad selbst braucht,
// etwa um Dateien zu vergleichen, fragt hier.
func schriftwurzelDesLaufs() string { return os.Getenv("MSE_SCHRIFTWURZEL") }

func gestartetMit(t *testing.T, saatort func() string) *App {
	t.Helper()

	// **Die Schriftwurzel kommt hier herein und nirgendwo sonst.**
	//
	// Die Schriften eines Packs liegen **neben** seinen Paketen — bei
	// `M15-Magic-Pack` unter `Magic - Fonts/` —, und `TestMain` gibt dem
	// ganzen Lauf seit dem 2026-08-26 ein eigenes Zuhause unter `/tmp`. Ohne
	// die Angabe findet die Suche keine einzige.
	//
	// **Sie stand bis zum 2026-08-28 im Abstandslauf**, und nur dort. Der
	// Fremdpaketlauf zur Freigabe 0.3.0 hat gezeigt, was das kostet:
	// `TestDieKanonischenSetsZeichnen` liest dieselben Pakete und wusste
	// nichts davon — 8.325 Befunde, davon 2.432 „die Schrift Beleren Bold ist
	// nicht zu finden". Er war **rot aus einem Grund, der nichts mit dem Code
	// zu tun hat**.
	//
	// Es ist der **fünfte** Weg dieser Art an einem Tag; die vier davor stehen
	// in `docs/CHRONIK.md`. Deshalb steht sie jetzt an der Stelle, an der
	// jeder dieser Läufe die Anwendung startet — und ein Wächter hält, dass es
	// bei einer bleibt (`internal/werkzeuge/schriftwurzel_test.go`).
	if ort := schriftwurzelDesLaufs(); ort != "" {
		t.Cleanup(MitSchriftwurzel(func() string { return ort }))
	}

	leer := t.TempDir()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.paketort = func() (string, error) { return leer, nil }
	anwendung.saatort = saatort

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	return anwendung
}

// saatset schreibt ein Set, das auf Spiel und Stil der Saat zeigt.
//
// **Von Hand geschrieben und nicht aus `testdaten/`**: die Probesets dort
// nennen `probe-stil`, und der ganze Punkt dieses Wächters ist der Stil, der im
// Bündel liegt.
func saatset(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "saatprobe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Mit einem mehrzeiligen Wert** —
	// #704.
	//
	// Er stand hier nicht, und deshalb ging der Rundlauf über die Saat grün
	// durch, während er über ein echtes Set 16 von 18 Feldwerten verlor: ein
	// mehrzeiliger `notes` wurde einzeilig geschrieben, und die zweite Zeile
	// beendete den `card:`-Block. Eine Vorrichtung, die ärmer ist als der
	// Betrieb, prüft den falschen Weg (§3).
	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: probe-schlicht\n" +
		"set info:\n\ttitle: Aus der Saat\ncard:\n\tname: Feuerball\n" +
		"\ttext: Füge drei Schaden zu.\n\tfarbe: rot\n" +
		"\tnotes:\n\t\tErste Zeile\n\t\t\n\t\tDritte Zeile\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}
