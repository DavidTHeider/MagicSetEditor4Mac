package flaechen_test

// **Kein Lauf über die Flächen sieht das echte Zuhause** — Zweitblick über
// Phase C, Befund 5 (C06, #2231). `eigenerBestand` gab `os.Environ()` weiter,
// samt dem echten `HOME`, und `fensterAus` übernahm es nach seinem eigenen
// wieder: die Vorrichtung des Fensters lief mit dem Zuhause der Nutzenden.
// Der Riegel der Anwendung gegen ein echtes Zuhause begrenzte den Schaden;
// die Zusage aus AGENTS.md §3 hielt trotzdem nicht.

import (
	"os"
	"os/user"
	"strings"
	"testing"
)

// echtesZuhause ist das Zuhause aus dem Benutzereintrag — nicht aus `HOME`,
// das ein Test schon umgesetzt haben kann.
func echtesZuhause(t *testing.T) string {
	t.Helper()

	ich, fehler := user.Current()
	if fehler != nil {
		t.Fatalf("der Benutzereintrag: %v", fehler)
	}

	return ich.HomeDir
}

// zuletzt gibt den Wert, der für einen Namen in einer Umgebung gilt — bei
// doppeltem Schlüssel der letzte, wie `os/exec` ihn nimmt.
func zuletzt(umgebung []string, name string) string {
	wert := ""

	for _, zeile := range umgebung {
		if schluessel, w, gefunden := strings.Cut(zeile, "="); gefunden && schluessel == name {
			wert = w
		}
	}

	return wert
}

func TestDieProzesseBekommenEinEigenesZuhause(t *testing.T) {
	if zuhause := zuletzt(eigenerBestand(t), "HOME"); zuhause == echtesZuhause(t) || zuhause == "" {
		t.Errorf("die Kindprozesse bekämen HOME=%q — erwartet ein eigenes", zuhause)
	}
}

func TestDasFensterNimmtNieDasEchteZuhause(t *testing.T) {
	echt := echtesZuhause(t)

	// Ein Aufrufer, der das echte ausdrücklich nennt — das hat die
	// Vorrichtung zu übergehen, nicht zu übernehmen.
	fensterAus(t, append(eigenerBestand(t), "HOME="+echt))

	if os.Getenv("HOME") == echt {
		t.Errorf("das Fenster läuft mit dem echten Zuhause %q", echt)
	}
}
