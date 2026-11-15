package main

// **Die Partieverwaltung an der Kommandozeile** — PLAY-03.AK-05
// (#1711).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der ganze Weg: ausführen, einlesen, kopieren, löschen.**
func TestDerGanzeVerwaltungswegAnDerZeile(t *testing.T) {
	ort, _ := eingefahren(t)

	partieID := einePartieAnDerZeile(t, ort)
	datei := filepath.Join(t.TempDir(), "partie.msepartie")

	if aus := mussLaufen(t, ort, "partie-ausfuhr", partieID, datei); !strings.Contains(
		aus, "ausgefuehrt") {
		t.Errorf("die Ausfuhr meldete %q", aus)
	}

	roh, fehler := os.ReadFile(datei) //nolint:gosec // der Pfad kommt aus dem Lauf
	if fehler != nil {
		t.Fatalf("die Datei lesen: %v", fehler)
	}

	if !strings.Contains(string(roh), `"fassung"`) {
		t.Errorf("in der Datei steht %.80q", roh)
	}

	if aus := mussLaufen(t, ort, "partie-einfuhr", datei); !strings.Contains(
		aus, "eingelesen") {
		t.Errorf("die Einfuhr meldete %q", aus)
	}

	if aus := mussLaufen(t, ort, "partie-kopieren", partieID); !strings.Contains(
		aus, "kopiert") {
		t.Errorf("das Kopieren meldete %q", aus)
	}

	if aus := mussLaufen(t, ort, "partie-loeschen", partieID); !strings.Contains(
		aus, "geloescht") {
		t.Errorf("das Löschen meldete %q", aus)
	}

	// **Und danach ist sie weg** — ein zweites Löschen ist ein Befund.
	if code, _, _ := fahreAn(t, ort, "partie-loeschen", partieID); code == CodeGut {
		t.Error("die Partie ließ sich zweimal löschen")
	}
}

// **Jeder der vier Befehle sagt, was ihm fehlt** — und keiner tut dabei
// etwas (§8).
func TestDieVerwaltungsbefehleNennenIhreBedienung(t *testing.T) {
	ort, _ := eingefahren(t)

	for _, befehl := range [][]string{
		{"partie-loeschen"},
		{"partie-kopieren", "eine", "zu viel"},
		{"partie-ausfuhr", "nur-die-partie"},
		{"partie-einfuhr"},
	} {
		t.Run(befehl[0], func(t *testing.T) {
			code, _, fehler := fahreAn(t, ort, befehl...)
			if code != CodeBedienung {
				t.Fatalf("Code %d statt %d", code, CodeBedienung)
			}

			if !strings.Contains(fehler, befehl[0]) {
				t.Errorf("die Klage nennt den Befehl nicht: %q", fehler)
			}
		})
	}
}

// einePartieAnDerZeile legt zwei Decks und eine Partie an.
func einePartieAnDerZeile(t *testing.T, ort string) string {
	t.Helper()

	sitze := make([]string, 0, 2)

	for _, name := range []string{"Anna", "Bert"} {
		deckID := strings.TrimSpace(
			mussLaufen(t, ort, "deck", "neu", name, "magic"))

		mussLaufen(t, ort, "deck", "legen",
			deckID, "a1b2c3d4-0000-4000-8000-00000000000a", "4")

		sitze = append(sitze,
			strings.ToLower(name)+":"+name+":"+deckID)
	}

	aus := mussLaufen(t, ort, append(
		[]string{"partie-anlegen", "magic"}, sitze...)...)

	_, nachher, gefunden := strings.Cut(aus, "Partie ")
	if !gefunden {
		t.Fatalf("die Antwort nennt keine Partie: %q", aus)
	}

	kennung, _, _ := strings.Cut(nachher, " ")

	return kennung
}
